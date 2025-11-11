package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/internal/handlers"
)

const (
	defaultPort       = ":8080"
	readTimeout       = 10 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 60 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	mux := http.NewServeMux()

	// Health handlers
	healthHandler := handlers.NewHealthHandler()
	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/healthz", healthHandler.Livez)
	mux.HandleFunc("/readyz", healthHandler.Readyz)

	// Seed management handlers
	mux.HandleFunc("/seed/sync", handlers.SeedSyncHandler)
	mux.HandleFunc("/seed/status", handlers.SeedStatusHandler)

	// Metrics
	mux.Handle("/metrics", promhttp.Handler())

	// Configurar servidor com timeouts para segurança (G114)
	srv := &http.Server{
		Addr:              defaultPort,
		Handler:           logRequest(logger, mux),
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	logger.Info("server starting", "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}

// logging middleware simples (JSON)
func logRequest(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "ua", r.UserAgent())
		next.ServeHTTP(w, r)
	})
}
