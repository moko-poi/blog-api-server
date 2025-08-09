package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger to provide a consistent logging interface
// Following Mat Ryer's pattern of simple, focused interfaces
type Logger struct {
	*slog.Logger
}

// New creates a new Logger with the specified output and level
func New(output io.Writer, level slog.Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(output, opts)
	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewDefault creates a new Logger with sensible defaults
func NewDefault() *Logger {
	return New(os.Stdout, slog.LevelInfo)
}

// Info logs an info message with key-value pairs
func (l *Logger) Info(ctx context.Context, msg string, keysAndValues ...any) {
	l.Logger.InfoContext(ctx, msg, keysAndValues...)
}

// Error logs an error message with key-value pairs
func (l *Logger) Error(ctx context.Context, msg string, keysAndValues ...any) {
	l.Logger.ErrorContext(ctx, msg, keysAndValues...)
}

// Debug logs a debug message with key-value pairs
func (l *Logger) Debug(ctx context.Context, msg string, keysAndValues ...any) {
	l.Logger.DebugContext(ctx, msg, keysAndValues...)
}

// Warn logs a warning message with key-value pairs
func (l *Logger) Warn(ctx context.Context, msg string, keysAndValues ...any) {
	l.Logger.WarnContext(ctx, msg, keysAndValues...)
}

// WithError adds an error to the logger context
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err),
	}
}

// WithFields adds fields to the logger context
func (l *Logger) WithFields(keysAndValues ...any) *Logger {
	return &Logger{
		Logger: l.Logger.With(keysAndValues...),
	}
}

// ParseLevel converts a string level to slog.Level
func ParseLevel(level string) (slog.Level, error) {
	switch level {
	case "debug", "DEBUG":
		return slog.LevelDebug, nil
	case "info", "INFO":
		return slog.LevelInfo, nil
	case "warn", "WARN", "warning", "WARNING":
		return slog.LevelWarn, nil
	case "error", "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown level: %s", level)
	}
}