package app

import (
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"runtime"
	"time"
)

const (
	LogKeyRequestId   = "request_id"
	LogKeyComponent   = "component"
	LogKeyError       = "error"
	LogKeyErrorReason = "reason"
)

var (
	_ slog.Handler = (*LogHandlerMiddleware)(nil)
	_ slog.Handler = (*NoopLogger)(nil)
)

// Logger is a slog.Logger wrapper with some shortcut methods
type Logger struct {
	*slog.Logger
	lvl    *slog.LevelVar
	addSrc bool
}

// NewLogger returns ready to use instance of Logger.
//
// If w is nil it creates 'noop' logger useful for tests
//
// There are options:
//   - WithLogLevel(slog.LogLevel) - sets level, slog.LevelError is default
//   - WithAddSource - enables/disables logging of source code position
func NewLogger(w io.Writer, options ...LoggerOption) *Logger {
	log := &Logger{lvl: new(slog.LevelVar)}
	log.lvl.Set(slog.LevelError)

	for _, opt := range options {
		opt(log)
	}

	var h slog.Handler = &NoopLogger{}
	if w != nil {
		h = slog.NewJSONHandler(w, &slog.HandlerOptions{Level: log.lvl, AddSource: log.addSrc})
		h = NewLogHandlerMiddleware(h)
	}

	log.Logger = slog.New(h)
	return log
}

// Err writes ERROR message
func (l *Logger) Err(err error, msg string, args ...any) {
	args = append(args, LogKeyError, err.Error(), LogKeyErrorReason, errors.Unwrap(err))
	l.logCtx(context.Background(), slog.LevelError, msg, args...)
}

// ErrContext writes ERROR message with context
func (l *Logger) ErrContext(ctx context.Context, err error, msg string, args ...any) {
	args = append(args, LogKeyError, err.Error(), LogKeyErrorReason, errors.Unwrap(err))
	l.logCtx(ctx, slog.LevelError, msg, args...)
}

// SetLevel updates current logger level
func (l *Logger) SetLevel(level slog.Level) {
	l.lvl.Set(level)
}

// LogLogger returns std log.Logger that acts as bridge to structured handler.
//
// For compatibility with libs that uses the older log API (http.Server.ErrorLog for example)
func (l *Logger) LogLogger() *log.Logger {
	return slog.NewLogLogger(l.Handler(), l.lvl.Level())
}

// With returns a Logger that includes the given attributes in each output operation.
// Arguments are converted to attributes as if by Logger.Log
func (l *Logger) With(args ...any) *Logger {
	return &Logger{Logger: l.Logger.With(args...), lvl: l.lvl, addSrc: l.addSrc}
}

// WithGroup returns a Logger that starts a group, if name is non-empty.
// The keys of all attributes added to the Logger will be qualified by the given name.
// (How that qualification happens depends on the Handler.WithGroup method of the Logger's Handler.)
// If name is empty, WithGroup returns the receiver.
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{Logger: l.Logger.WithGroup(name), lvl: l.lvl, addSrc: l.addSrc}
}

// WithComponent returns a Logger that appends 'component' attribute.
// If component is empty, WithGroup returns the receiver.
func (l *Logger) WithComponent(component string) *Logger {
	if component == "" {
		return l
	}
	return &Logger{Logger: l.Logger.With(LogKeyComponent, component), lvl: l.lvl, addSrc: l.addSrc}
}

func (l *Logger) logCtx(ctx context.Context, lvl slog.Level, msg string, args ...any) {
	if !l.Logger.Enabled(ctx, lvl) {
		return
	}
	var pc uintptr
	if l.addSrc {
		var pcs [1]uintptr
		runtime.Callers(3, pcs[:]) // skip current wrapper, it's caller and runtime.Callers
		pc = pcs[0]
	}
	r := slog.NewRecord(time.Now(), lvl, msg, pc)
	r.Add(args...)
	_ = l.Logger.Handler().Handle(ctx, r)
}

type LoggerOption func(*Logger)

// WithLogLevel sets the log level
func WithLogLevel(level slog.Level) func(*Logger) {
	return func(log *Logger) {
		log.lvl.Set(level)
	}
}

// WithAddSource enables/disables logging of source code position
func WithAddSource(addSource bool) func(logger *Logger) {
	return func(log *Logger) {
		log.addSrc = addSource
	}
}

// LogHandlerMiddleware is a slog.Handler wrapper that
// enriches log record with context data
type LogHandlerMiddleware struct {
	next slog.Handler
}

func NewLogHandlerMiddleware(next slog.Handler) *LogHandlerMiddleware {
	return &LogHandlerMiddleware{next: next}
}

func (l LogHandlerMiddleware) Enabled(ctx context.Context, level slog.Level) bool {
	return l.next.Enabled(ctx, level)
}

func (l LogHandlerMiddleware) Handle(ctx context.Context, record slog.Record) error {
	if reqID := RequestID(ctx); reqID != "" {
		record.Add(LogKeyRequestId, reqID)
	}
	for key, val := range LogAttrs(ctx) {
		record.Add(key, val)
	}
	return l.next.Handle(ctx, record)
}

func (l LogHandlerMiddleware) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LogHandlerMiddleware{next: l.next.WithAttrs(attrs)}
}

func (l LogHandlerMiddleware) WithGroup(name string) slog.Handler {
	return &LogHandlerMiddleware{next: l.next.WithGroup(name)}
}

// NoopLogger is a slog.Handler implementation that do nothing, like >/dev/null
//
// Useful for tests
type NoopLogger struct {
}

func (n NoopLogger) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}

func (n NoopLogger) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (n NoopLogger) WithAttrs(_ []slog.Attr) slog.Handler {
	return n
}

func (n NoopLogger) WithGroup(_ string) slog.Handler {
	return n
}
