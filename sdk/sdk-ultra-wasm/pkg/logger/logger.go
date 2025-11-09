// pkg/logger/logger.go
package logger

import (
	"context"
	"log/slog"
	"os"
)

var (
	// Log é o logger estruturado global
	Log *slog.Logger
)

func init() {
	// Configurar logger estruturado JSON
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	Log = slog.New(handler)
}

// WithContext retorna logger com contexto
func WithContext(ctx context.Context) *slog.Logger {
	return Log.With(
		slog.String("request_id", getRequestID(ctx)),
	)
}

// Info registra mensagem informativa
func Info(msg string, args ...any) {
	Log.Info(msg, args...)
}

// Error registra mensagem de erro
func Error(msg string, args ...any) {
	Log.Error(msg, args...)
}

// Warn registra mensagem de aviso
func Warn(msg string, args ...any) {
	Log.Warn(msg, args...)
}

// Debug registra mensagem de debug
func Debug(msg string, args ...any) {
	Log.Debug(msg, args...)
}

func getRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return ""
}
