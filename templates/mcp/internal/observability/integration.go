package observability

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
)

// Service aggregates all observability functionality
type Service struct {
	telemetry *TelemetryService
	logger    *zap.Logger
	config    config.TelemetryConfig
}

// NewService creates a new observability service
func NewService(cfg config.TelemetryConfig, logger *zap.Logger) (*Service, error) {
	if !cfg.Enabled {
		return &Service{
			logger: logger,
			config: cfg,
		}, nil
	}

	// Convert config to telemetry config
	telemetryConfig := TelemetryConfig{
		Enabled:        cfg.Enabled,
		ServiceName:    cfg.ServiceName,
		ServiceVersion: cfg.ServiceVersion,
		Environment:    cfg.Environment,
		Debug:          cfg.Debug,

		// Tracing configuration
		TracingEnabled:    cfg.Tracing.Enabled,
		TracingSampleRate: cfg.Tracing.SampleRate,
		TracingMaxSpans:   cfg.Tracing.MaxSpans,
		TracingBatchSize:  cfg.Tracing.BatchSize,
		TracingTimeout:    cfg.Tracing.Timeout,

		// Metrics configuration
		MetricsEnabled:   cfg.Metrics.Enabled,
		MetricsPort:      cfg.Metrics.Port,
		MetricsPath:      cfg.Metrics.Path,
		MetricsInterval:  cfg.Metrics.CollectInterval,
		HistogramBuckets: cfg.Metrics.HistogramBuckets,

		// Exporters configuration
		JaegerEnabled:  cfg.Exporters.Jaeger.Enabled,
		JaegerEndpoint: cfg.Exporters.Jaeger.Endpoint,
		JaegerUser:     cfg.Exporters.Jaeger.User,
		JaegerPassword: cfg.Exporters.Jaeger.Password,

		OTLPEnabled:  cfg.Exporters.OTLP.Enabled,
		OTLPEndpoint: cfg.Exporters.OTLP.Endpoint,
		OTLPInsecure: cfg.Exporters.OTLP.Insecure,
		OTLPHeaders:  cfg.Exporters.OTLP.Headers,

		ConsoleEnabled: cfg.Exporters.Console.Enabled,
	}

	// Create telemetry service
	telemetryService, err := NewTelemetryService(telemetryConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("creating telemetry service: %w", err)
	}

	return &Service{
		telemetry: telemetryService,
		logger:    logger,
		config:    cfg,
	}, nil
}

// Start initializes the observability service
func (s *Service) Start(ctx context.Context) error {
	if s.telemetry == nil {
		s.logger.Info("Observability disabled, skipping telemetry initialization")
		return nil
	}

	if err := s.telemetry.Start(ctx); err != nil {
		return fmt.Errorf("starting telemetry service: %w", err)
	}

	s.logger.Info("Observability service started",
		zap.String("service", s.config.ServiceName),
		zap.String("version", s.config.ServiceVersion),
		zap.String("environment", s.config.Environment),
		zap.Bool("tracing_enabled", s.config.Tracing.Enabled),
		zap.Bool("metrics_enabled", s.config.Metrics.Enabled),
	)

	return nil
}

// Stop gracefully shuts down the observability service
func (s *Service) Stop(ctx context.Context) error {
	if s.telemetry == nil {
		return nil
	}

	if err := s.telemetry.Stop(ctx); err != nil {
		s.logger.Error("Error stopping telemetry service", zap.Error(err))
		return err
	}

	s.logger.Info("Observability service stopped")
	return nil
}

// HTTPMiddleware returns the HTTP telemetry middleware
func (s *Service) HTTPMiddleware() func(http.Handler) http.Handler {
	if s.telemetry == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return s.telemetry.HTTPTelemetryMiddleware
}

// GetTelemetryService returns the underlying telemetry service
func (s *Service) GetTelemetryService() *TelemetryService {
	return s.telemetry
}

// HealthCheck returns the observability service health status
func (s *Service) HealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"observability": map[string]interface{}{
			"enabled": s.config.Enabled,
			"status":  "healthy",
		},
	}

	if s.telemetry != nil {
		health["telemetry"] = map[string]interface{}{
			"tracing_enabled": s.config.Tracing.Enabled,
			"metrics_enabled": s.config.Metrics.Enabled,
			"service_name":    s.config.ServiceName,
			"environment":     s.config.Environment,
		}
	}

	return health
}

// RecordTaskOperation records a task-related operation for telemetry
func (s *Service) RecordTaskOperation(ctx context.Context, operation, taskID string, fn func(context.Context) error) error {
	if s.telemetry == nil {
		return fn(ctx)
	}

	// Use BusinessOperation with task-specific attributes
	attrs := []attribute.KeyValue{
		attribute.String("task_id", taskID),
		attribute.String("operation", operation),
	}
	return s.telemetry.BusinessOperation(ctx, "task."+operation, attrs, fn)
}

// RecordDatabaseOperation records a database operation for telemetry
func (s *Service) RecordDatabaseOperation(ctx context.Context, operation, table, query string, fn func(context.Context) error) error {
	if s.telemetry == nil {
		return fn(ctx)
	}

	wrapper := NewDatabaseTelemetryWrapper(s.telemetry)
	return wrapper.WrapDatabaseOperation(ctx, operation, table, query, fn)
}

// RecordCacheOperation records a cache operation for telemetry
func (s *Service) RecordCacheOperation(ctx context.Context, operation, key string, fn func(context.Context) error) error {
	if s.telemetry == nil {
		return fn(ctx)
	}

	return s.telemetry.CacheOperation(ctx, operation, key, fn)
}

// RecordMessageQueueOperation records a message queue operation for telemetry
func (s *Service) RecordMessageQueueOperation(ctx context.Context, operation, subject string, fn func(context.Context) error) error {
	if s.telemetry == nil {
		return fn(ctx)
	}

	return s.telemetry.MessageQueueOperation(ctx, operation, subject, fn)
}

// LogWithTrace logs a message with tracing context
func (s *Service) LogWithTrace(ctx context.Context, level, message string, fields ...zap.Field) {
	if s.telemetry != nil {
		s.telemetry.LogEvent(ctx, level, message, fields...)
	} else {
		// Fallback to regular logging without trace context
		switch level {
		case "debug":
			s.logger.Debug(message, fields...)
		case "info":
			s.logger.Info(message, fields...)
		case "warn":
			s.logger.Warn(message, fields...)
		case "error":
			s.logger.Error(message, fields...)
		}
	}
}
