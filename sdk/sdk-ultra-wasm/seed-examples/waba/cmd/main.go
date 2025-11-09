// seed-examples/waba/cmd/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/bootstrap"

	// Import side-effect para auto-registro
	_ "github.com/vertikon/seed-examples/waba/internal/plugins/waba"
)

func main() {
	// Bootstrap SDK
	mux := bootstrap.Bootstrap(bootstrap.Config{
		EnableRecovery: true,
		EnableLogger:   true,
		CORSOrigins:    []string{"*"},
	})

	// Servidor HTTP
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Iniciar servidor em goroutine
	go func() {
		log.Println("🚀 Servidor WABA iniciando na porta 8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Erro ao iniciar servidor: %v", err)
		}
	}()

	// Aguardar sinal de shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("🛑 Desligando servidor...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Erro ao desligar servidor: %v", err)
	}

	log.Println("✅ Servidor desligado com sucesso")
}
