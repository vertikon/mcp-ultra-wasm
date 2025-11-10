// pkg/router/middleware/recovery.go
package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/httpx"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/logger"
)

// Recovery captura panics e retorna 500
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
						"method", r.Method,
						"path", r.URL.Path)
					http.Error(w, "Internal Server Error", httpx.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
