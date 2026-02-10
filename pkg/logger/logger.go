package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type Logger struct {
	*slog.Logger
}

func New(output io.Writer, level slog.Level) (l *Logger) {
	if output == nil {
		output = os.Stdout
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(output, opts)
	l = &Logger{
		Logger: slog.New(handler),
	}
	return
}

func NewDefault() (l *Logger) {
	l = New(os.Stdout, slog.LevelInfo)
	return
}

func (l *Logger) WithContext(ctx context.Context) (logger *slog.Logger) {
	correlationID := ctx.Value("correlation_id")
	if correlationID != nil {
		logger = l.With("correlation_id", correlationID)
		return
	}
	logger = l.Logger
	return
}

func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	logger := l.WithContext(ctx)
	logger.Info(msg, args...)
}

func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	logger := l.WithContext(ctx)
	logger.Error(msg, args...)
}

func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	logger := l.WithContext(ctx)
	logger.Warn(msg, args...)
}

func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	logger := l.WithContext(ctx)
	logger.Debug(msg, args...)
}
