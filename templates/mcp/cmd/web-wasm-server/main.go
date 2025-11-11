package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/wasm/handlers"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/wasm/nats"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/wasm/observability"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/wasm/security"
	"go.uber.org/zap"
)

var (
	port       = flag.String("port", "8080", "Porta do servidor web")
	natsURL    = flag.String("nats-url", "nats://localhost:4222", "URL do servidor NATS")
	logLevel   = flag.String("log-level", "info", "Nível de log (debug, info, warn, error)")
	staticPath = flag.String("static-path", "./wasm/static", "Caminho dos arquivos estáticos")
	wasmPath   = flag.String("wasm-path", "./wasm/wasm", "Caminho dos arquivos WASM")
)

func main() {
	flag.Parse()

	// Inicializar logger
	logger, err := observability.NewLogger(*logLevel)
	if err != nil {
		log.Fatalf("Erro ao inicializar logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Iniciando Web WASM Server",
		zap.String("port", *port),
		zap.String("nats_url", *natsURL),
		zap.String("static_path", *staticPath),
		zap.String("wasm_path", *wasmPath),
	)

	// Configurar modo Gin
	if *logLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Inicializar cliente NATS
	natsClient, err := nats.NewClient(*natsURL, logger)
	if err != nil {
		logger.Fatal("Erro ao conectar ao NATS", zap.Error(err))
	}
	defer natsClient.Close()

	// Inicializar publisher NATS
	publisher := nats.NewPublisher(natsClient, logger)

	// Inicializar middleware de segurança
	securityMiddleware := security.NewMiddleware(logger)

	// Inicializar handlers
	uiHandler := handlers.NewUIHandler(*staticPath, *wasmPath, logger)
	apiHandler := handlers.NewAPIHandler(publisher, logger)
	wsHandler := handlers.NewWebSocketHandler(logger)

	// Configurar router
	router := gin.New()

	// Middleware globais
	router.Use(gin.Recovery())
	router.Use(observability.LoggingMiddleware(logger))
	router.Use(securityMiddleware.CORS())
	router.Use(securityMiddleware.RateLimit())

	// Rotas da UI
	router.Static("/static", *staticPath)
	router.GET("/wasm/*filepath", uiHandler.ServeWASM)
	router.GET("/", uiHandler.ServeIndex)

	// Rotas da API
	api := router.Group("/api/v1")
	{
		api.POST("/tasks", apiHandler.CreateTask)
		api.GET("/tasks/:id", apiHandler.GetTask)
		api.GET("/tasks", apiHandler.ListTasks)
		api.DELETE("/tasks/:id", apiHandler.CancelTask)
	}

	// WebSocket
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
			"service":   "wasm-server",
			"version":   "1.0.0",
		})
	})

	// Configurar servidor HTTP
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Iniciar servidor em goroutine
	go func() {
		logger.Info("Servidor web iniciado", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Erro ao iniciar servidor", zap.Error(err))
		}
	}()

	// Aguardar sinais de shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Desligando servidor...")

	// Shutdown graceful
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Erro durante shutdown", zap.Error(err))
	}

	logger.Info("Servidor desligado com sucesso")
}


