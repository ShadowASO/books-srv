/*
---------------------------------------------------------------------------------------
File: mslogger.go
Autor: Aldenor
Data: 04-05-2026
Alteração: 04-05-2026
---------------------------------------------------------------------------------------
*/
package mslogger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

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
	FilePath      string // se vazio, não grava em arquivo
	Stdout        bool
	Rotate        bool // se FilePath != "" e Rotate=true usa lumberjack
	MaxSizeMB     int
	MaxBackups    int
	MaxAgeDays    int
	Compress      bool
	Level         Level
	Flags         int // log.Ldate|log.Ltime|...
	CallerDepth   int // depth para Output; default 3
	DisableCaller bool
}

type Logger struct {
	out    io.Writer
	closer io.Closer

	level atomic.Int32 // guarda Level

	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	err   *log.Logger

	depth int
	mu    sync.Mutex // protege Close em cenários concorrentes
}

// Conveniência global (opcional).
// Em Go, é aceitável desde que o "main" faça o Init e você evite nil.
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
	// defaults
	if opts.Flags == 0 {
		opts.Flags = log.Ldate | log.Ltime | log.Lmicroseconds
	}
	if opts.CallerDepth == 0 {
		opts.CallerDepth = 3 // wrapper -> method -> caller
	}
	if opts.MaxSizeMB == 0 {
		opts.MaxSizeMB = 10
	}
	if opts.MaxBackups == 0 {
		opts.MaxBackups = 5
	}
	if opts.MaxAgeDays == 0 {
		opts.MaxAgeDays = 30
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

	flags := opts.Flags
	if !opts.DisableCaller {
		flags |= log.Lshortfile
	}

	l := &Logger{
		out:    out,
		closer: closer,
		depth:  opts.CallerDepth,
		debug:  log.New(out, "[DEBUG] ", flags),
		info:   log.New(out, "[INFO]  ", flags),
		warn:   log.New(out, "[WARN]  ", flags),
		err:    log.New(out, "[ERROR] ", flags),
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

// logf centraliza tudo (menos duplicação, mais Go-like).
func (l *Logger) logf(target Level, lg *log.Logger, format string, args ...any) {
	if l == nil {
		// fallback simples e previsível
		log.Printf(format, args...)
		return
	}
	if lg == nil || !l.enabled(target) {
		return
	}

	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}

	// Output(depth, ...) controla o caller do Lshortfile
	_ = lg.Output(l.depth, msg)
}

func (l *Logger) Debug(msg string)                  { l.logf(DebugLevel, l.debug, "%s", msg) }
func (l *Logger) Debugf(format string, args ...any) { l.logf(DebugLevel, l.debug, format, args...) }

func (l *Logger) Info(msg string)                  { l.logf(InfoLevel, l.info, "%s", msg) }
func (l *Logger) Infof(format string, args ...any) { l.logf(InfoLevel, l.info, format, args...) }

func (l *Logger) Warn(msg string)                  { l.logf(WarnLevel, l.warn, "%s", msg) }
func (l *Logger) Warnf(format string, args ...any) { l.logf(WarnLevel, l.warn, format, args...) }

func (l *Logger) Error(msg string)                  { l.logf(ErrorLevel, l.err, "%s", msg) }
func (l *Logger) Errorf(format string, args ...any) { l.logf(ErrorLevel, l.err, format, args...) }

func (l *Logger) ErrorErr(context string, err error) {
	if err == nil {
		return
	}
	l.Errorf("%s: err=%v", context, err)
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
