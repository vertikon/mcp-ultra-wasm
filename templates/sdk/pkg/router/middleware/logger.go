// pkg/router/middleware/logger.go
package middleware

import (
	"net/http"
	"time"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/logger"
)

// Logger registra requests
func Logger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			logger.Info("http request started",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent())
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			logger.Info("http request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"duration_ms", duration.Milliseconds())
		})
	}
}
