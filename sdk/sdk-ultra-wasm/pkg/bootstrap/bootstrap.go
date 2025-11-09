// pkg/bootstrap/bootstrap.go
package bootstrap

import (
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/logger"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/registry"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/router"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/router/middleware"
)

// Config configura bootstrap
type Config struct {
	CORSOrigins    []string
	EnableRecovery bool
	EnableLogger   bool
}

// Bootstrap inicializa SDK
// - Aplica middlewares globais
// - Registra rotas dos plugins
// - Retorna mux pronto para uso
func Bootstrap(cfg Config) *router.Mux {
	mux := router.NewMux()

	// Health endpoints
	mux.Handle("GET", "/healthz", healthz)
	mux.Handle("GET", "/readyz", readiness)
	mux.Handle("GET", "/health", healthz) // Alias
	mux.Handle("GET", "/ping", healthz)   // Alias

	// Middlewares globais
	if cfg.EnableRecovery {
		mux.Use(middleware.Recovery())
	}
	if cfg.EnableLogger {
		mux.Use(middleware.Logger())
	}
	if len(cfg.CORSOrigins) > 0 {
		mux.Use(middleware.CORS(cfg.CORSOrigins))
	}

	// Middlewares customizados (de plugins)
	for _, mi := range registry.MiddlewareInjectors() {
		logger.Info("registering middleware",
			"name", mi.Name(),
			"priority", mi.Priority())
		mux.Use(mi.Middleware())
	}

	// Rotas de plugins
	for _, ri := range registry.RouteInjectors() {
		logger.Info("registering plugin",
			"name", ri.Name(),
			"version", ri.Version())

		for _, route := range ri.Routes() {
			logger.Info("registering route",
				"method", route.Method,
				"path", route.Path,
				"plugin", ri.Name())
			mux.Handle(route.Method, route.Path, route.Handler)
		}
	}

	// Marca como pronto
	MarkReady()

	return mux
}
