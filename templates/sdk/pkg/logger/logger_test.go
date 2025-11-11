// pkg/logger/logger_test.go
package logger_test

import (
	"context"
	"testing"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/logger"
)

func TestLogger(t *testing.T) {
	// Teste básico para garantir que logger não causa panic
	logger.Info("test message",
		"key", "value",
		"number", 42,
	)

	logger.Error("error message",
		"error", "test error",
	)

	logger.Warn("warning message")
	logger.Debug("debug message")

	// Verificar que nenhum panic ocorreu
	t.Log("Logger test completed successfully")
}

func TestLoggerNotNil(t *testing.T) {
	if logger.Log == nil {
		t.Fatal("Logger global não deve ser nil")
	}
}

type contextKey string

const requestIDKey = contextKey("request_id")

func TestWithContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "nil context",
			ctx:  nil,
			want: "",
		},
		{
			name: "context without request_id",
			ctx:  context.Background(),
			want: "",
		},
		{
			name: "context with request_id",
			ctx:  context.WithValue(context.Background(), requestIDKey, "req-123"),
			want: "req-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextLogger := logger.WithContext(tt.ctx)
			if contextLogger == nil {
				t.Error("WithContext returned nil logger")
			}
			// Log to ensure no panic
			contextLogger.Info("test message")
		})
	}
}
