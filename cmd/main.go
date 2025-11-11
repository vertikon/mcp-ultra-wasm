package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vertikon/mcp-ultra-templates/internal/handlers/cli"
)

var exitFunc = os.Exit

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "erro: %v\n", err)
		exitFunc(1)
	}
}

func run(ctx context.Context, args []string) error {
	return cli.ExecuteWithArgs(ctx, args)
}

