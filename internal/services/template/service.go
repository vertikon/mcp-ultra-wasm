package template

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"github.com/vertikon/mcp-ultra-templates/internal/config"
	"github.com/vertikon/mcp-ultra-templates/internal/models"
	repo "github.com/vertikon/mcp-ultra-templates/internal/repository/fs"
	pkgtemplate "github.com/vertikon/mcp-ultra-templates/pkg/template"
)

// Repository define o comportamento esperado para storage de templates.
type Repository interface {
	ListTemplates(ctx context.Context) ([]models.TemplateMetadata, error)
	LoadTemplate(ctx context.Context, name string) (*models.TemplateMetadata, string, error)
}

// Service orquestra a renderização de templates utilizando repositório e métricas.
type Service struct {
	cfg     config.RenderingConfig
	logger  zerolog.Logger
	repo    Repository
	metrics renderMetrics
}

type renderMetrics struct {
	duration *prometheus.HistogramVec
	errors   *prometheus.CounterVec
	success  *prometheus.CounterVec
}

// New cria um novo Template Service.
func New(cfg config.RenderingConfig, logger zerolog.Logger, registry *prometheus.Registry, repository Repository) *Service {
	m := renderMetrics{
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "template_render_duration_seconds",
			Help:    "Duração da renderização de templates",
			Buckets: prometheus.DefBuckets,
		}, []string{"template"}),
		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "template_render_errors_total",
			Help: "Total de erros durante renderização",
		}, []string{"template", "reason"}),
		success: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "template_render_success_total",
			Help: "Total de renderizações bem sucedidas",
		}, []string{"template"}),
	}

	if registry != nil {
		registry.MustRegister(m.duration, m.errors, m.success)
	}

	return &Service{
		cfg:     cfg,
		logger:  logger,
		repo:    repository,
		metrics: m,
	}
}

// NewDefaultRepository é um helper para criar repositório filesystem com configuração padrão.
func NewDefaultRepository(root string) Repository {
	return repo.New(root)
}

// RenderRequest contém os parâmetros necessários.
type RenderRequest struct {
	TemplateName string
	OutputDir    string
	Values       map[string]string
	Overwrite    bool
}

// RenderResponse retorna metadados pós-renderização.
type RenderResponse struct {
	Template models.TemplateMetadata
	Output   string
}

// List retorna os templates disponíveis.
func (s *Service) List(ctx context.Context) ([]models.TemplateMetadata, error) {
	return s.repo.ListTemplates(ctx)
}

// Render aplica o template específico e gera o projeto.
func (s *Service) Render(ctx context.Context, req RenderRequest) (*RenderResponse, error) {
	if req.TemplateName == "" {
		return nil, errors.New("template name is required")
	}
	if req.OutputDir == "" {
		return nil, errors.New("output directory is required")
	}

	ctx, cancel := context.WithTimeout(ctx, s.cfg.OperationTimeout)
	defer cancel()

	start := time.Now()
	meta, templatePath, err := s.repo.LoadTemplate(ctx, req.TemplateName)
	if err != nil {
		s.metrics.errors.WithLabelValues(req.TemplateName, "load").Inc()
		return nil, err
	}

	values := mergeValues(meta, req.Values)
	if err := validateVariables(meta, values); err != nil {
		s.metrics.errors.WithLabelValues(req.TemplateName, "validation").Inc()
		return nil, err
	}

	if err := s.prepareOutput(req.OutputDir, req.Overwrite); err != nil {
		s.metrics.errors.WithLabelValues(req.TemplateName, "output").Inc()
		return nil, err
	}

	operation := func() error {
		opts := pkgtemplate.RenderOptions{
			IgnoredPaths: map[string]struct{}{
				"template.yaml": {},
			},
		}
		return pkgtemplate.RenderDirectory(ctx, templatePath, req.OutputDir, values, opts)
	}

	notify := func(err error, d time.Duration) {
		s.logger.Warn().
			Err(err).
			Dur("retry_in", d).
			Str("template", req.TemplateName).
			Msg("falha ao renderizar template, tentando novamente")
	}

	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = 500 * time.Millisecond
	expBackoff.MaxElapsedTime = s.cfg.OperationTimeout

	if err := backoff.RetryNotify(operation, backoff.WithContext(backoff.WithMaxRetries(expBackoff, uint64(s.cfg.MaxRetryAttempts)), ctx), notify); err != nil {
		s.metrics.errors.WithLabelValues(req.TemplateName, "render").Inc()
		return nil, fmt.Errorf("render template: %w", err)
	}

	elapsed := time.Since(start).Seconds()
	s.metrics.duration.WithLabelValues(req.TemplateName).Observe(elapsed)
	s.metrics.success.WithLabelValues(req.TemplateName).Inc()

	s.logger.Info().
		Str("template", req.TemplateName).
		Str("output", req.OutputDir).
		Dur("duration", time.Since(start)).
		Msg("template renderizado com sucesso")

	return &RenderResponse{
		Template: *meta,
		Output:   req.OutputDir,
	}, nil
}

func mergeValues(meta *models.TemplateMetadata, values map[string]string) map[string]string {
	result := make(map[string]string, len(meta.Defaults)+len(values))
	for k, v := range meta.Defaults {
		result[k] = v
	}
	for k, v := range values {
		result[k] = v
	}
	return result
}

func validateVariables(meta *models.TemplateMetadata, values map[string]string) error {
	for _, variable := range meta.Variables {
		if !variable.Required {
			continue
		}
		if values[variable.Key] == "" {
			return pkgtemplate.ErrMissingVariable{Key: variable.Key}
		}
	}
	return nil
}

func (s *Service) prepareOutput(path string, overwrite bool) error {
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("output path is not a directory: %s", path)
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return fmt.Errorf("read output dir: %w", err)
		}
		if len(entries) > 0 && !overwrite {
			return fmt.Errorf("output directory is not empty: %s", path)
		}
		if overwrite {
			for _, entry := range entries {
				if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
					return fmt.Errorf("clear output dir: %w", err)
				}
			}
		}
		return nil
	}
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0o755)
	}
	return fmt.Errorf("stat output dir: %w", err)
}

