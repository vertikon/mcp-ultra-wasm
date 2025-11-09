// pkg/bootstrap/health.go
package bootstrap

import (
	"net/http"
	"sync/atomic"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/httpx"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/logger"
)

var ready atomic.Bool

// MarkReady marca aplicação como pronta
func MarkReady() {
	ready.Store(true)
}

// MarkNotReady marca aplicação como não-pronta
func MarkNotReady() {
	ready.Store(false)
}

// healthz é o endpoint de liveness
func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(httpx.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		logger.Error("error writing healthz response", "error", err)
	}
}

// readiness é o endpoint de readiness
func readiness(w http.ResponseWriter, _ *http.Request) {
	if !ready.Load() {
		w.WriteHeader(httpx.StatusServiceUnavailable)
		if _, err := w.Write([]byte("not ready")); err != nil {
			logger.Error("error writing readiness not-ready response", "error", err)
		}
		return
	}
	w.WriteHeader(httpx.StatusOK)
	if _, err := w.Write([]byte("ready")); err != nil {
		logger.Error("error writing readiness ready response", "error", err)
	}
}
