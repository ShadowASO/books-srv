/*
---------------------------------------------------------------------------------------
File: mslogger.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 06-05-2026
---------------------------------------------------------------------------------------
Inicializa o serviço de logger  do sistema

	err = mslogger.InitGlobal(mslogger.Options{
		FilePath:   "./logs/app.log",
		Stdout:     true,
		Rotate:     true,
		MaxSizeMB:  20,
		MaxBackups: 10,
		MaxAgeDays: 30,
		Compress:   true,
		Level:      mslogger.DebugLevel,
		JSON:       true,
		Service:    "books-srv",
		AddSource:  true,
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		if mslogger.LoggerGlobal != nil {
			_ = mslogger.LoggerGlobal.Close()
		}
	}()
*/
package mslogger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Level int32

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	OffLevel
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case OffLevel:
		return "off"
	default:
		return "unknown"
	}
}

func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return DebugLevel
	case "info", "":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "off", "none":
		return OffLevel
	default:
		return InfoLevel
	}
}

type Options struct {
	FilePath   string
	Stdout     bool
	Rotate     bool
	MaxSizeMB  int
	MaxBackups int
	MaxAgeDays int
	Compress   bool
	Level      Level
	JSON       bool
	Service    string
	AddSource  bool
}

type Source struct {
	Function string `json:"function,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
}

type AppEntry struct {
	Time    string  `json:"time"`
	Level   string  `json:"level"`
	Message string  `json:"msg"`
	Service string  `json:"service,omitempty"`
	Source  *Source `json:"source,omitempty"`

	ID          string `json:"id,omitempty"`
	Context     string `json:"context,omitempty"`
	Error       string `json:"error,omitempty"`
	UserID      string `json:"user_id,omitempty"`
	Username    string `json:"username,omitempty"`
	Status      int    `json:"status,omitempty"`
	Code        int    `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

type AppLogData struct {
	ID          string
	Context     string
	Error       error
	UserID      string
	Username    string
	Status      int
	Code        int
	Description string
}

type HTTPEntry struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"msg"`
	Service string `json:"service,omitempty"`

	ID       string `json:"id,omitempty"`
	Status   int    `json:"status"`
	Method   string `json:"method"`
	Path     string `json:"path"`
	Route    string `json:"route,omitempty"`
	Handler  string `json:"handler,omitempty"`
	ClientIP string `json:"client_ip,omitempty"`

	Duration   string `json:"duration"`
	DurationMS int64  `json:"duration_ms"`
	DurationUS int64  `json:"duration_us"`

	ErrorCode   string `json:"error_code,omitempty"`
	ErrorDetail string `json:"error_detail,omitempty"`
}

type HTTPLogData struct {
	ID          string
	Status      int
	Method      string
	Path        string
	Route       string
	Handler     string
	ClientIP    string
	Duration    time.Duration
	ErrorCode   string
	ErrorDetail string
}

type Logger struct {
	closer  io.Closer
	level   atomic.Int32
	service string
	json    bool
	source  bool
	std     *log.Logger
	mu      sync.Mutex
}

var (
	LoggerGlobal     *Logger
	onceLoggerGlobal sync.Once
)

func InitGlobal(opts Options) error {
	var err error

	onceLoggerGlobal.Do(func() {
		if opts.Level == 0 && os.Getenv("LOG_LEVEL") != "" {
			opts.Level = ParseLevel(os.Getenv("LOG_LEVEL"))
		}

		LoggerGlobal, err = New(opts)
	})

	return err
}

func ensureDir(path string) error {
	if path == "" {
		return nil
	}

	return os.MkdirAll(filepath.Dir(path), 0o755)
}

func New(opts Options) (*Logger, error) {
	if opts.MaxSizeMB == 0 {
		opts.MaxSizeMB = 10
	}

	if opts.MaxBackups == 0 {
		opts.MaxBackups = 5
	}

	if opts.MaxAgeDays == 0 {
		opts.MaxAgeDays = 30
	}

	if opts.Service == "" {
		opts.Service = "microsrv"
	}

	var writers []io.Writer
	var closer io.Closer

	if opts.FilePath != "" {
		if err := ensureDir(opts.FilePath); err != nil {
			return nil, fmt.Errorf("create log dir: %w", err)
		}

		if opts.Rotate {
			rot := &lumberjack.Logger{
				Filename:   opts.FilePath,
				MaxSize:    opts.MaxSizeMB,
				MaxBackups: opts.MaxBackups,
				MaxAge:     opts.MaxAgeDays,
				Compress:   opts.Compress,
			}

			writers = append(writers, rot)
			closer = rot
		} else {
			f, err := os.OpenFile(opts.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return nil, fmt.Errorf("open log file: %w", err)
			}

			writers = append(writers, f)
			closer = f
		}
	}

	if opts.Stdout || len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	out := io.MultiWriter(writers...)

	l := &Logger{
		closer:  closer,
		service: opts.Service,
		json:    opts.JSON,
		source:  opts.AddSource,
		std:     log.New(out, "", 0),
	}

	l.level.Store(int32(opts.Level))

	return l, nil
}

func (l *Logger) SetLevel(level Level) {
	if l == nil {
		return
	}

	l.level.Store(int32(level))
}

func (l *Logger) Level() Level {
	if l == nil {
		return InfoLevel
	}

	return Level(l.level.Load())
}

func (l *Logger) enabled(target Level) bool {
	cur := l.Level()

	return cur != OffLevel && target >= cur
}

func caller(skip int) *Source {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	fn := runtime.FuncForPC(pc)

	fnName := ""
	if fn != nil {
		fnName = fn.Name()

		if idx := strings.LastIndex(fnName, "/"); idx >= 0 {
			fnName = fnName[idx+1:]
		}
	}

	return &Source{
		Function: fnName,
		File:     filepath.Base(file),
		Line:     line,
	}
}

func (l *Logger) app(level Level, skip int, msg string, data AppLogData) {
	if l == nil {
		return
	}

	if !l.enabled(level) {
		return
	}

	entry := AppEntry{
		Time:     time.Now().Format(time.RFC3339Nano),
		Level:    strings.ToUpper(level.String()),
		Message:  msg,
		Service:  l.service,
		ID:       data.ID,
		Context:  data.Context,
		UserID:   data.UserID,
		Username: data.Username,
	}

	if data.Error != nil {
		entry.Error = data.Error.Error()
	}

	if l.source {
		entry.Source = caller(skip)
	}

	if l.json {
		b, err := json.Marshal(entry)
		if err != nil {
			l.std.Printf(
				`{"time":"%s","level":"ERROR","msg":"erro ao serializar log","error":%q}`,
				time.Now().Format(time.RFC3339Nano),
				err.Error(),
			)
			return
		}

		l.std.Println(string(b))
		return
	}

	prefix := strings.ToUpper(level.String())

	if entry.Source != nil {
		l.std.Printf("[%s] %s:%d %s", prefix, entry.Source.File, entry.Source.Line, msg)
		return
	}

	l.std.Printf("[%s] %s", prefix, msg)
}

func (l *Logger) Debug(msg string) {
	l.app(DebugLevel, 4, msg, AppLogData{})
}

func (l *Logger) Debugf(format string, args ...any) {
	l.app(DebugLevel, 4, fmt.Sprintf(format, args...), AppLogData{})
}

func (l *Logger) Info(msg string) {
	l.app(InfoLevel, 4, msg, AppLogData{})
}

func (l *Logger) Infof(format string, args ...any) {
	l.app(InfoLevel, 3, fmt.Sprintf(format, args...), AppLogData{})
}

func (l *Logger) Warn(msg string) {
	l.app(WarnLevel, 4, msg, AppLogData{})
}

func (l *Logger) Warnf(format string, args ...any) {
	l.app(WarnLevel, 4, fmt.Sprintf(format, args...), AppLogData{})
}

func (l *Logger) Error(msg string) {
	l.app(ErrorLevel, 4, msg, AppLogData{})
}

func (l *Logger) Errorf(format string, args ...any) {
	l.app(ErrorLevel, 4, fmt.Sprintf(format, args...), AppLogData{})
}

func (l *Logger) ErrorErr(context string, err error) {
	if err == nil {
		return
	}

	l.app(ErrorLevel, 4, context, AppLogData{
		Error: err,
	})
}

func (l *Logger) DebugData(msg string, data AppLogData) {
	l.app(DebugLevel, 4, msg, data)
}

func (l *Logger) InfoData(msg string, data AppLogData) {
	l.app(InfoLevel, 4, msg, data)
}

func (l *Logger) WarnData(msg string, data AppLogData) {
	l.app(WarnLevel, 4, msg, data)
}

func (l *Logger) ErrorData(msg string, data AppLogData) {
	l.app(ErrorLevel, 4, msg, data)
}

func (l *Logger) HTTP(data HTTPLogData) {
	if l == nil {
		return
	}

	var level Level

	switch {
	case data.Status >= 500:
		level = ErrorLevel
	case data.Status >= 400:
		level = WarnLevel
	default:
		level = InfoLevel
	}

	if !l.enabled(level) {
		return
	}

	entry := HTTPEntry{
		Time:        time.Now().Format(time.RFC3339Nano),
		Level:       strings.ToUpper(level.String()),
		Message:     "http_request",
		Service:     l.service,
		ID:          data.ID,
		Status:      data.Status,
		Method:      data.Method,
		Path:        data.Path,
		Route:       data.Route,
		Handler:     data.Handler,
		ClientIP:    data.ClientIP,
		Duration:    data.Duration.String(),
		DurationMS:  data.Duration.Milliseconds(),
		DurationUS:  data.Duration.Microseconds(),
		ErrorCode:   data.ErrorCode,
		ErrorDetail: data.ErrorDetail,
	}

	if l.json {
		b, err := json.Marshal(entry)
		if err != nil {
			l.std.Printf(
				`{"time":"%s","level":"ERROR","msg":"erro ao serializar log http","error":%q}`,
				time.Now().Format(time.RFC3339Nano),
				err.Error(),
			)
			return
		}

		l.std.Println(string(b))
		return
	}

	l.std.Printf(
		"[%s] %d %s %s duration=%s id=%s",
		entry.Level,
		entry.Status,
		entry.Method,
		entry.Path,
		entry.Duration,
		entry.ID,
	)
}

func (l *Logger) Close() error {
	if l == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closer != nil {
		err := l.closer.Close()
		l.closer = nil
		return err
	}

	return nil
}
