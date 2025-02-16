// pkg/logger/logger.go
package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
)

// contextKey is a private type to prevent collisions in context.
type contextKey string

const (
	loggerKey contextKey = "logger"
)

// Logger представляет собой обертку вокруг slog.Logger и предоставляет дополнительные функции.
type Logger struct {
	*slog.Logger
}

// New создает новый экземпляр Logger.
func New(level slog.Level) *Logger {

	addSource := false
	if level == slog.LevelDebug {
		addSource = true
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: addSource,
	})
	logger := slog.New(handler)
	return &Logger{logger}
}

func NewTestLogger() *Logger {
	logger := slog.New(
		slog.NewTextHandler(
			io.Discard,
			&slog.HandlerOptions{Level: slog.LevelDebug},
		),
	)
	return &Logger{logger}
}

// WithLogger добавляет Logger в контекст.
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext извлекает Logger из контекста.
// Если логгер не найден, возвращается логгер по умолчанию.
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey).(*Logger); ok {
		return logger
	}
	// Возвращает дефолтный Logger, если он не найден в контексте.
	return New(slog.LevelInfo)
}

// With создает новый логгер с дополнительными атрибутами.
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{Logger: l.Logger.With(args...)}
}

// ParseLogLevel преобразует строковое представление уровня логирования в slog.Level.
func ParseLogLevel(levelStr string) (slog.Level, error) {
	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, errors.New("неверный уровень логгирования: " + levelStr)
	}
}
