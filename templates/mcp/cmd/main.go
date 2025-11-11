package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/handlers"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/httpx"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/metrics"
)

var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if syncErr := logger.Sync(); syncErr != nil {
			// Ignore sync errors on shutdown (common on Windows)
			log.Printf("Warning: failed to sync logger: %v", syncErr)
		}
	}()

	logger.Info("Starting MCP Ultra service",
		zap.String("version", Version),
		zap.String("build_date", BuildDate),
		zap.String("commit", GitCommit),
	)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize HTTP router
	router := httpx.NewRouter()

	// Add middleware
	router.Use(httpx.RequestID)
	router.Use(httpx.RealIP)
	router.Use(httpx.Logger)
	router.Use(httpx.Recoverer)
	router.Use(httpx.Timeout(60)) // Timeout in seconds

	// CORS configuration
	router.Use(httpx.DefaultCORS())

	// Initialize health handler
	healthHandler := handlers.NewHealthHandler()

	// Register routes
	router.Get("/livez", healthHandler.Livez)
	router.Get("/readyz", healthHandler.Readyz)
	router.Get("/health", healthHandler.Health)
	router.Method("GET", "/metrics", metrics.Handler())

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting HTTP server",
			zap.String("address", server.Addr),
			zap.Int("port", cfg.Server.Port),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
