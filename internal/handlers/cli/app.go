package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"

	"github.com/vertikon/mcp-ultra-templates/internal/config"
	"github.com/vertikon/mcp-ultra-templates/internal/services/observability"
	templateservice "github.com/vertikon/mcp-ultra-templates/internal/services/template"
	"github.com/vertikon/mcp-ultra-templates/pkg/log"
)

// App orquestra os componentes principais da CLI.
type App struct {
	cfg             *config.Config
	loggerProvider  *log.LoggerProvider
	logger          zerolog.Logger
	obs             *observability.Service
	templateService *templateservice.Service
}

// NewApp carrega a configuração e prepara dependências centrais.
func NewApp(cfg *config.Config) *App {
	loggerProvider := log.New(cfg.Logging.Level, cfg.Logging.Pretty)
	logger := loggerProvider.Logger()

	obsSvc := observability.New(cfg.Observability, logger)
	repository := templateservice.NewDefaultRepository(cfg.TemplatesPath)
	templateSvc := templateservice.New(cfg.Rendering, logger, obsSvc.Registry(), repository)

	return &App{
		cfg:             cfg,
		loggerProvider:  loggerProvider,
		logger:          logger,
		obs:             obsSvc,
		templateService: templateSvc,
	}
}

// Context retorna um contexto preparado com tratamento de sinais.
func (a *App) Context() (context.Context, context.CancelFunc) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	return ctx, cancel
}

// StartObservability inicializa métricas e tracing.
func (a *App) StartObservability(ctx context.Context) error {
	return a.obs.Start(ctx)
}

// Shutdown executa finalização ordenada.
func (a *App) Shutdown(ctx context.Context) {
	a.obs.Shutdown(ctx)
}

// TemplateService expõe o serviço principal para as commands.
func (a *App) TemplateService() *templateservice.Service {
	return a.templateService
}

// Logger expõe o logger configurado.
func (a *App) Logger() zerolog.Logger {
	return a.logger
}

