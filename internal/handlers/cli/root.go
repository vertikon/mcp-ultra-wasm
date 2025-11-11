package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/vertikon/mcp-ultra-templates/internal/config"
)

type contextKey string

var appInstance *App

// Execute inicializa a CLI raiz com as subcommands configuradas.
func Execute(ctx context.Context) error {
	return ExecuteWithArgs(ctx, os.Args[1:])
}

// ExecuteWithArgs permite injetar argumentos (útil em testes).
func ExecuteWithArgs(ctx context.Context, args []string) error {
	cfgPath := ""
	templatesDir := ""
	appInstance = nil

	rootCmd := &cobra.Command{
		Use:   "mcp-templates",
		Short: "Gerador de projetos a partir dos templates MCP Ultra",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if appInstance != nil {
				return nil
			}

			cfg, err := config.Load(cfgPath)
			if err != nil {
				return fmt.Errorf("carregar configuração: %w", err)
			}

			if templatesDir != "" {
				cfg.TemplatesPath = templatesDir
			}

			appInstance = NewApp(cfg)

			if err := appInstance.StartObservability(cmd.Context()); err != nil {
				return fmt.Errorf("iniciar observabilidade: %w", err)
			}

			cmd.SetContext(context.WithValue(cmd.Context(), contextKey("app"), appInstance))
			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgPath, "config", "", "Arquivo de configuração YAML")
	rootCmd.PersistentFlags().StringVar(&templatesDir, "templates-path", "", "Caminho raiz dos templates")

	rootCmd.AddCommand(listCommand(), renderCommand())

	if args != nil {
		rootCmd.SetArgs(args)
	}

	defer func() {
		if appInstance != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			appInstance.Shutdown(shutdownCtx)
		}
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return err
	}

	return nil
}

// MustApp obtém a instância atual da aplicação a partir do contexto do comando.
func MustApp(cmd *cobra.Command) *App {
	value := cmd.Context().Value(contextKey("app"))
	if value == nil {
		panic("app não inicializado")
	}
	return value.(*App)
}

