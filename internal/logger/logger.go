package logger

import (
	"context"
	"log/slog"
	"net/http"
	"os"
)

type Logger struct {
	*slog.Logger
}

func New() *Logger {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	return &Logger{Logger: log}
}

func NewMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := NewContext(r.Context(), New())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type loggerContextKey string

const contextKeyValue loggerContextKey = "context-logger"

func NewContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, contextKeyValue, l)
}

func FromContext(ctx context.Context) *Logger {
	if l := ctx.Value(contextKeyValue); l != nil {
		return l.(*Logger)
	}

	return New()
}
