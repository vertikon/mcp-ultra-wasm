package observability

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/vertikon/mcp-ultra-templates/internal/config"
	pkgmetrics "github.com/vertikon/mcp-ultra-templates/pkg/metrics"
)

// Service inicializa métricas e tracing compartilhados.
type Service struct {
	cfg         config.ObservabilityConfig
	logger      zerolog.Logger
	metricsReg  *prometheus.Registry
	metricsHTTP *pkgmetrics.Registry
	tracer      *sdktrace.TracerProvider
}

// New cria um novo serviço de observabilidade baseado na configuração.
func New(cfg config.ObservabilityConfig, logger zerolog.Logger) *Service {
	return &Service{
		cfg:         cfg,
		logger:      logger,
		metricsReg:  pkgmetrics.NewRegistry(cfg.MetricsNS),
		metricsHTTP: &pkgmetrics.Registry{},
	}
}

// Start inicializa métricas e, opcionalmente, tracing OTLP.
func (s *Service) Start(ctx context.Context) error {
	if s.cfg.EnableMetrics {
		s.metricsHTTP.Serve(s.cfg.MetricsAddr, s.metricsReg)
		s.logger.Info().
			Str("address", s.cfg.MetricsAddr).
			Msg("servidor de métricas iniciado")
	}

	if s.cfg.EnableTracing {
		if err := s.initTracing(ctx); err != nil {
			return fmt.Errorf("init tracing: %w", err)
		}
	}

	return nil
}

// Shutdown encerra recursos de observabilidade.
func (s *Service) Shutdown(ctx context.Context) {
	if s.cfg.EnableMetrics {
		if err := s.metricsHTTP.Shutdown(); err != nil {
			s.logger.Error().Err(err).Msg("erro ao encerrar métricas")
		}
	}

	if s.tracer != nil {
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := s.tracer.Shutdown(ctx); err != nil {
			s.logger.Error().Err(err).Msg("erro ao encerrar tracer provider")
		}
	}
}

func (s *Service) initTracing(ctx context.Context) error {
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(s.cfg.OTLPEndpoint),
	)
	if err != nil {
		return fmt.Errorf("create otlp exporter: %w", err)
	}

	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(s.cfg.ServiceName),
		),
	)
	if err != nil {
		return fmt.Errorf("create otel resource: %w", err)
	}

	s.tracer = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(s.tracer)
	s.logger.Info().
		Str("endpoint", s.cfg.OTLPEndpoint).
		Msg("tracing OTLP configurado")

	return nil
}

// Registry retorna o registry Prometheus para registro de métricas customizadas.
func (s *Service) Registry() *prometheus.Registry {
	return s.metricsReg
}

