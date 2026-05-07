/*
---------------------------------------------------------------------------------------
File: mslogger.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 07-05-2026
---------------------------------------------------------------------------------------
Inicializa o serviço de logger do sistema usando log/slog.

Exemplo de uso:

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
	"context"
	"fmt"
	"io"
	"log/slog"
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

func toSlogLevel(level Level) slog.Level {
	switch level {
	case DebugLevel:
		return slog.LevelDebug
	case InfoLevel:
		return slog.LevelInfo
	case WarnLevel:
		return slog.LevelWarn
	case ErrorLevel:
		return slog.LevelError
	case OffLevel:
		return slog.Level(1000)
	default:
		return slog.LevelInfo
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

type AppLogData struct {
	RequestID   string
	Context     string
	Error       error
	UserID      string
	Username    string
	Status      int
	Code        int
	Description string

	Mode string
	Env  string
}

type HTTPLogData struct {
	RequestID   string
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
	closer   io.Closer
	level    atomic.Int32
	levelVar *slog.LevelVar
	handler  slog.Handler
	service  string
	mu       sync.Mutex
}

var (
	LoggerGlobal *Logger
	//onceLoggerGlobal sync.Once
)

func InitGlobal(opts Options) error {
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		opts.Level = ParseLevel(envLevel)
	}

	logger, err := New(opts)
	if err != nil {
		return err
	}

	LoggerGlobal = logger
	return nil
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

	levelVar := new(slog.LevelVar)
	levelVar.Set(toSlogLevel(opts.Level))

	handlerOpts := &slog.HandlerOptions{
		Level:     levelVar,
		AddSource: opts.AddSource,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				src, ok := a.Value.Any().(*slog.Source)
				if !ok || src == nil {
					return slog.Attr{}
				}

				// Quando o record foi criado sem source válido,
				// o slog pode gerar file=".", line=0, function="".
				// Nesse caso, removemos o atributo source.
				if src.Function == "" || src.File == "" || src.File == "." || src.Line == 0 {
					return slog.Attr{}
				}

				return slog.Group(
					slog.SourceKey,
					slog.String("function", shortFunctionName(src.Function)),
					slog.String("file", filepath.Base(src.File)),
					slog.Int("line", src.Line),
				)
			}

			return a
		},
	}

	var handler slog.Handler

	if opts.JSON {
		handler = slog.NewJSONHandler(out, handlerOpts)
	} else {
		handler = slog.NewTextHandler(out, handlerOpts)
	}

	handler = handler.WithAttrs([]slog.Attr{
		slog.String("service", opts.Service),
	})

	l := &Logger{
		closer:   closer,
		levelVar: levelVar,
		handler:  handler,
		service:  opts.Service,
	}

	l.level.Store(int32(opts.Level))

	return l, nil
}

func shortFunctionName(fn string) string {
	if fn == "" {
		return ""
	}

	if idx := strings.LastIndex(fn, "/"); idx >= 0 {
		fn = fn[idx+1:]
	}

	return fn
}

func (l *Logger) SetLevel(level Level) {
	if l == nil {
		return
	}

	l.level.Store(int32(level))

	if l.levelVar != nil {
		l.levelVar.Set(toSlogLevel(level))
	}
}

func (l *Logger) Level() Level {
	if l == nil {
		return InfoLevel
	}

	return Level(l.level.Load())
}

func (l *Logger) enabled(ctx context.Context, level Level) bool {
	if l == nil {
		return false
	}

	if l.Level() == OffLevel {
		return false
	}

	if l.handler == nil {
		return false
	}

	return l.handler.Enabled(ctx, toSlogLevel(level))
}

func (l *Logger) log(ctx context.Context, level Level, skip int, msg string, attrs ...slog.Attr) {
	if l == nil {
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if !l.enabled(ctx, level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(skip, pcs[:])

	record := slog.NewRecord(
		time.Now(),
		toSlogLevel(level),
		msg,
		pcs[0],
	)

	record.AddAttrs(attrs...)

	_ = l.handler.Handle(ctx, record)
}

func (l *Logger) logNoSource(ctx context.Context, level Level, msg string, attrs ...slog.Attr) {
	if l == nil {
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if !l.enabled(ctx, level) {
		return
	}

	record := slog.NewRecord(
		time.Now(),
		toSlogLevel(level),
		msg,
		0,
	)

	record.AddAttrs(attrs...)

	_ = l.handler.Handle(ctx, record)
}

func appAttrs(data AppLogData) []slog.Attr {
	attrs := make([]slog.Attr, 0, 10)

	if data.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", data.RequestID))
	}

	if data.Context != "" {
		attrs = append(attrs, slog.String("context", data.Context))
	}

	if data.Error != nil {
		attrs = append(attrs, slog.String("error", data.Error.Error()))
	}

	if data.UserID != "" {
		attrs = append(attrs, slog.String("user_id", data.UserID))
	}

	if data.Username != "" {
		attrs = append(attrs, slog.String("username", data.Username))
	}

	if data.Status != 0 {
		attrs = append(attrs, slog.Int("status", data.Status))
	}

	if data.Code != 0 {
		attrs = append(attrs, slog.Int("code", data.Code))
	}

	if data.Description != "" {
		attrs = append(attrs, slog.String("description", data.Description))
	}
	if data.Mode != "" {
		attrs = append(attrs, slog.String("mode", data.Mode))
	}

	if data.Env != "" {
		attrs = append(attrs, slog.String("env", data.Env))
	}

	return attrs
}

func (l *Logger) app(level Level, skip int, msg string, data AppLogData) {
	l.log(context.Background(), level, skip, msg, appAttrs(data)...)
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
	l.app(InfoLevel, 4, fmt.Sprintf(format, args...), AppLogData{})
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

func (l *Logger) ErrorErr(msg string, err error) {
	if err == nil {
		return
	}

	l.app(ErrorLevel, 4, msg, AppLogData{
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

func httpLevel(status int) Level {
	switch {
	case status >= 500:
		return ErrorLevel
	case status >= 400:
		return WarnLevel
	default:
		return InfoLevel
	}
}

func (l *Logger) HTTP(data HTTPLogData) {
	if l == nil {
		return
	}

	level := httpLevel(data.Status)

	attrs := []slog.Attr{
		slog.Int("status", data.Status),
		slog.String("method", data.Method),
		slog.String("path", data.Path),
		slog.String("duration", data.Duration.String()),
		slog.Int64("duration_us", data.Duration.Microseconds()),
	}

	if data.RequestID != "" {
		attrs = append(attrs, slog.String("request_id", data.RequestID))
	}

	if data.Route != "" {
		attrs = append(attrs, slog.String("route", data.Route))
	}

	if data.Handler != "" {
		attrs = append(attrs, slog.String("handler", data.Handler))
	}

	if data.ClientIP != "" {
		attrs = append(attrs, slog.String("client_ip", data.ClientIP))
	}

	if data.ErrorCode != "" {
		attrs = append(attrs, slog.String("error_code", data.ErrorCode))
	}

	if data.ErrorDetail != "" {
		attrs = append(attrs, slog.String("error_detail", data.ErrorDetail))
	}

	l.logNoSource(context.Background(), level, "http_request", attrs...)
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
