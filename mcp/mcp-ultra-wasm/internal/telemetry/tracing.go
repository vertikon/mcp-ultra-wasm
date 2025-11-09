package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
)

type TracingConfig struct {
	Enabled        bool          `yaml:"enabled" envconfig:"TRACING_ENABLED" default:"true"`
	ServiceName    string        `yaml:"service_name" envconfig:"SERVICE_NAME" default:"mcp-ultra-wasm"`
	ServiceVersion string        `yaml:"service_version" envconfig:"SERVICE_VERSION" default:"v1.0.0"`
	Environment    string        `yaml:"environment" envconfig:"ENVIRONMENT" default:"development"`
	Exporter       string        `yaml:"exporter" envconfig:"TRACE_EXPORTER" default:"otlp"`
	SampleRate     float64       `yaml:"sample_rate" envconfig:"TRACING_SAMPLE_RATE" default:"0.1"`
	BatchTimeout   time.Duration `yaml:"batch_timeout" envconfig:"TRACE_BATCH_TIMEOUT" default:"5s"`

	// OTLP HTTP Configuration
	OTLPEndpoint string            `yaml:"otlp_endpoint" envconfig:"OTEL_EXPORTER_OTLP_ENDPOINT" default:"http://localhost:4318"`
	OTLPHeaders  map[string]string `yaml:"otlp_headers" envconfig:"OTEL_EXPORTER_OTLP_HEADERS"`

	// Jaeger Configuration
	JaegerEndpoint string `yaml:"jaeger_endpoint" envconfig:"JAEGER_ENDPOINT" default:"http://localhost:14268/api/traces"`
	JaegerUser     string `yaml:"jaeger_user" envconfig:"JAEGER_USER"`
	JaegerPassword string `yaml:"jaeger_password" envconfig:"JAEGER_PASSWORD"`

	// Additional attributes
	ResourceAttributes map[string]string `yaml:"resource_attributes"`
}

type TracingProvider struct {
	provider *sdktrace.TracerProvider
	config   *TracingConfig
	logger   *zap.Logger
	shutdown func(context.Context) error
}

func NewTracingProvider(config *TracingConfig, logger *zap.Logger) (*TracingProvider, error) {
	if !config.Enabled {
		logger.Info("Distributed tracing is disabled")
		return &TracingProvider{
			config:   config,
			logger:   logger,
			shutdown: func(context.Context) error { return nil },
		}, nil
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracing resource: %w", err)
	}

	// Add custom resource attributes
	if len(config.ResourceAttributes) > 0 {
		attrs := make([]attribute.KeyValue, 0, len(config.ResourceAttributes))
		for key, value := range config.ResourceAttributes {
			attrs = append(attrs, attribute.String(key, value))
		}

		customRes := resource.NewWithAttributes(semconv.SchemaURL, attrs...)
		res, err = resource.Merge(res, customRes)
		if err != nil {
			return nil, fmt.Errorf("failed to merge custom resource attributes: %w", err)
		}
	}

	// Create span exporter based on configuration
	exporter, err := createSpanExporter(config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create span exporter: %w", err)
	}

	// Create tracer provider
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter, sdktrace.WithBatchTimeout(config.BatchTimeout)),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(config.SampleRate)),
	)

	// Set global tracer provider
	otel.SetTracerProvider(provider)

	// Set global text map propagator
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("Distributed tracing initialized",
		zap.String("service_name", config.ServiceName),
		zap.String("service_version", config.ServiceVersion),
		zap.String("environment", config.Environment),
		zap.String("exporter", config.Exporter),
		zap.Float64("sample_rate", config.SampleRate))

	return &TracingProvider{
		provider: provider,
		config:   config,
		logger:   logger,
		shutdown: func(ctx context.Context) error {
			return provider.Shutdown(ctx)
		},
	}, nil
}

func createSpanExporter(config *TracingConfig, logger *zap.Logger) (sdktrace.SpanExporter, error) {
	switch config.Exporter {
	case "otlp", "otlp-http":
		return createOTLPHTTPExporter(config)
	case "stdout":
		return createStdoutExporter()
	case "noop":
		return &noopExporter{}, nil
	default:
		logger.Warn("Unknown trace exporter, falling back to stdout",
			zap.String("exporter", config.Exporter))
		return createStdoutExporter()
	}
}

func createOTLPHTTPExporter(config *TracingConfig) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.OTLPEndpoint),
		otlptracehttp.WithInsecure(), // Use WithTLSClientConfig for production
	}

	// Add custom headers if configured
	if len(config.OTLPHeaders) > 0 {
		headers := make(map[string]string)
		for k, v := range config.OTLPHeaders {
			headers[k] = v
		}
		opts = append(opts, otlptracehttp.WithHeaders(headers))
	}

	return otlptracehttp.New(context.Background(), opts...)
}

func createStdoutExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithoutTimestamps(),
	)
}

// GetTracer returns a tracer for the given name
func (tp *TracingProvider) GetTracer(name string) trace.Tracer {
	if tp.provider == nil {
		return noop.NewTracerProvider().Tracer(name)
	}
	return tp.provider.Tracer(name)
}

// Shutdown gracefully shuts down the tracing provider
func (tp *TracingProvider) Shutdown(ctx context.Context) error {
	if tp.shutdown != nil {
		tp.logger.Info("Shutting down tracing provider")
		return tp.shutdown(ctx)
	}
	return nil
}

// Utility functions for common tracing patterns

// TraceFunction wraps a function with tracing
func TraceFunction(ctx context.Context, tracer trace.Tracer, name string, fn func(context.Context) error) error {
	ctx, span := tracer.Start(ctx, name)
	defer span.End()

	if err := fn(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// TraceFunctionWithResult wraps a function with tracing and returns a result
func TraceFunctionWithResult[T any](ctx context.Context, tracer trace.Tracer, name string, fn func(context.Context) (T, error)) (T, error) {
	ctx, span := tracer.Start(ctx, name)
	defer span.End()

	result, err := fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		var zero T
		return zero, err
	}

	return result, nil
}

// AddSpanAttributes adds multiple attributes to the current span
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// SetSpanError sets error status on the current span
func SetSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
}

// GetTraceID returns the trace ID from the current context
func GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from the current context
func GetSpanID(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}
	return ""
}

// InjectTraceContext injects trace context into a map (for cross-service calls)
func InjectTraceContext(ctx context.Context, carrier map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, &mapCarrier{m: carrier})
}

// ExtractTraceContext extracts trace context from a map
func ExtractTraceContext(ctx context.Context, carrier map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, &mapCarrier{m: carrier})
}

// mapCarrier implements the TextMapCarrier interface
type mapCarrier struct {
	m map[string]string
}

func (c *mapCarrier) Get(key string) string {
	return c.m[key]
}

func (c *mapCarrier) Set(key, value string) {
	c.m[key] = value
}

func (c *mapCarrier) Keys() []string {
	keys := make([]string, 0, len(c.m))
	for k := range c.m {
		keys = append(keys, k)
	}
	return keys
}

// noopExporter is a no-op span exporter for disabled tracing
type noopExporter struct{}

func (e *noopExporter) ExportSpans(context.Context, []sdktrace.ReadOnlySpan) error {
	return nil
}

func (e *noopExporter) Shutdown(context.Context) error {
	return nil
}

// Span naming conventions
const (
	SpanNameHTTPRequest   = "http.request"
	SpanNameHTTPHandler   = "http.handler"
	SpanNameDBQuery       = "db.query"
	SpanNameDBTransaction = "db.transaction"
	SpanNameRedisOp       = "redis.operation"
	SpanNameNATSPublish   = "nats.publish"
	SpanNameNATSConsume   = "nats.consume"
	SpanNameServiceCall   = "service.call"
	SpanNameBusinessLogic = "business.logic"
)

// Common span attributes
var (
	AttrServiceName    = attribute.Key("service.name")
	AttrServiceVersion = attribute.Key("service.version")
	AttrEnvironment    = attribute.Key("deployment.environment")
	AttrUserID         = attribute.Key("user.id")
	AttrSessionID      = attribute.Key("session.id")
	AttrRequestID      = attribute.Key("request.id")
	AttrCorrelationID  = attribute.Key("correlation.id")
)
