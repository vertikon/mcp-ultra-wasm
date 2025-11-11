package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TracingManager gerencia tracing distribuído
type TracingManager struct {
	tp     trace.TracerProvider
	tracer trace.Tracer
	logger *zap.Logger
	config *TracingConfig
}

type TracingConfig struct {
	Enabled          bool          `json:"enabled"`
	ServiceName      string        `json:"service_name"`
	ServiceVersion   string        `json:"service_version"`
	Environment      string        `json:"environment"`
	JaegerEndpoint   string        `json:"jaeger_endpoint"`
	OTLPEndpoint     string        `json:"otlp_endpoint"`
	SamplingRatio    float64       `json:"sampling_ratio"`
	BatchTimeout     time.Duration `json:"batch_timeout"`
	ExportTimeout    time.Duration `json:"export_timeout"`
	MaxExportBatchSize int         `json:"max_export_batch_size"`
}

func NewTracingManager(config *TracingConfig, logger *zap.Logger) (*TracingManager, error) {
	if config == nil {
		config = &TracingConfig{
			Enabled:             false,
			ServiceName:         "wasm-server",
			ServiceVersion:      "1.0.0",
			Environment:         "development",
			SamplingRatio:       1.0,
			BatchTimeout:        5 * time.Second,
			ExportTimeout:       30 * time.Second,
			MaxExportBatchSize:  512,
		}
	}

	if !config.Enabled {
		// Criar tracer no-op se não estiver habilitado
		return &TracingManager{
			tp:     trace.NewNoopTracerProvider(),
			tracer: trace.NewNoopTracerProvider().Tracer("noop"),
			logger: logger.Named("tracing"),
			config: config,
		}, nil
	}

	// Criar tracer provider com exportadores
	tp, err := createTracerProvider(config, logger)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar tracer provider: %w", err)
	}

	// Configurar como tracer global
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tp.Tracer("wasm-tracer")

	manager := &TracingManager{
		tp:     tp,
		tracer: tracer,
		logger: logger.Named("tracing"),
		config: config,
	}

	logger.Info("Tracing inicializado",
		zap.String("service", config.ServiceName),
		zap.String("version", config.ServiceVersion),
		zap.Float64("sampling_ratio", config.SamplingRatio))

	return manager, nil
}

func createTracerProvider(config *TracingConfig, logger *zap.Logger) (trace.TracerProvider, error) {
	ctx := context.Background()

	// Criar resource com atributos do serviço
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", config.ServiceName),
			attribute.String("service.version", config.ServiceVersion),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar resource: %w", err)
	}

	// Configurar exportadores
	var exporters []sdktrace.SpanExporter

	// Exportador Jaeger
	if config.JaegerEndpoint != "" {
		jaegerExporter, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)))
		if err != nil {
			logger.Warn("Erro ao criar exportador Jaeger", zap.Error(err))
		} else {
			exporters = append(exporters, jaegerExporter)
			logger.Info("Exportador Jaeger configurado", zap.String("endpoint", config.JaegerEndpoint))
		}
	}

	// Exportador OTLP (para coletor OpenTelemetry)
	if config.OTLPEndpoint != "" {
		ctx, cancel := context.WithTimeout(ctx, config.ExportTimeout)
		defer cancel()

		otlpExporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(config.OTLPEndpoint),
		)
		if err != nil {
			logger.Warn("Erro ao criar exportador OTLP", zap.Error(err))
		} else {
			exporters = append(exporters, otlpExporter)
			logger.Info("Exportador OTLP configurado", zap.String("endpoint", config.OTLPEndpoint))
		}
	}

	// Se não houver exportadores, usar no-op
	if len(exporters) == 0 {
		logger.Warn("Nenhum exportador configurado, usando no-op tracer")
		return trace.NewNoopTracerProvider(), nil
	}

	// Criar tracer provider com sampling
	sampler := sdktrace.AlwaysSample()
	if config.SamplingRatio < 1.0 {
		sampler = sdktrace.TraceIDRatioBased(config.SamplingRatio)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporters...),
		sdktrace.WithSampler(sampler),
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(
			exporters...,
			sdktrace.WithBatchTimeout(config.BatchTimeout),
			sdktrace.WithMaxExportBatchSize(config.MaxExportBatchSize),
		),
	)

	return tp, nil
}

func (tm *TracingManager) Shutdown(ctx context.Context) error {
	if tp, ok := tm.tp.(*sdktrace.TracerProvider); ok {
		tm.logger.Info("Desligando tracing...")
		return tp.Shutdown(ctx)
	}
	return nil
}

// Métodos para criar spans

func (tm *TracingManager) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return tm.tracer.Start(ctx, name, opts...)
}

func (tm *TracingManager) StartHTTPSpan(ctx context.Context, method, path string) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.path", path),
			attribute.String("span.kind", "server"),
		),
	}
	return tm.tracer.Start(ctx, fmt.Sprintf("HTTP %s %s", method, path), opts...)
}

func (tm *TracingManager) StartWASMSpan(ctx context.Context, function string) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("wasm.function", function),
			attribute.String("span.kind", "internal"),
		),
	}
	return tm.tracer.Start(ctx, fmt.Sprintf("WASM %s", function), opts...)
}

func (tm *TracingManager) StartSDKSpan(ctx context.Context, plugin, method string) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("sdk.plugin", plugin),
			attribute.String("sdk.method", method),
			attribute.String("span.kind", "client"),
		),
	}
	return tm.tracer.Start(ctx, fmt.Sprintf("SDK %s.%s", plugin, method), opts...)
}

func (tm *TracingManager) StartNATSSpan(ctx context.Context, subject, operation string) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("nats.subject", subject),
			attribute.String("nats.operation", operation),
			attribute.String("span.kind", "producer"),
		),
	}
	return tm.tracer.Start(ctx, fmt.Sprintf("NATS %s %s", operation, subject), opts...)
}

func (tm *TracingManager) StartWebSocketSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
	opts := []trace.SpanStartOption{
		trace.WithAttributes(
			attribute.String("websocket.operation", operation),
			attribute.String("span.kind", "server"),
		),
	}
	return tm.tracer.Start(ctx, fmt.Sprintf("WebSocket %s", operation), opts...)
}

// SpanHelper utilitário para trabalhar com spans
type SpanHelper struct {
	span trace.Span
}

func NewSpanHelper(span trace.Span) *SpanHelper {
	return &SpanHelper{span: span}
}

func (sh *SpanHelper) SetAttributes(attrs ...attribute.KeyValue) {
	sh.span.SetAttributes(attrs...)
}

func (sh *SpanHelper) AddEvent(name string, attrs ...attribute.KeyValue) {
	sh.span.AddEvent(name, trace.WithAttributes(attrs...))
}

func (sh *SpanHelper) SetStatus(code trace.SpanStatusCode, message string) {
	sh.span.SetStatus(code, message)
}

func (sh *SpanHelper) RecordError(err error, opts ...trace.EventOption) {
	sh.span.RecordError(err, opts...)
	sh.span.SetStatus(trace.StatusCodeError, err.Error())
}

func (sh *SpanHelper) End() {
	sh.span.End()
}

// TracingMiddleware middleware para HTTP
func (tm *TracingManager) TracingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !tm.config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			ctx, span := tm.StartHTTPSpan(r.Context(), r.Method, r.URL.Path)
			defer span.End()

			// Adicionar headers de tracing à response
			w.Header().Set("X-Trace-ID", span.SpanContext().TraceID().String())
			w.Header().Set("X-Span-ID", span.SpanContext().SpanID().String())

			// Usar response writer wrapper para capturar status code
			rw := &responseWriter{ResponseWriter: w, statusCode: 200}

			// Continuar com a requisição
			next.ServeHTTP(rw, r.WithContext(ctx))

			// Atualizar span com status code
			span.SetAttributes(
				attribute.Int("http.status_code", rw.statusCode),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
			)

			// Atualizar status do span baseado no status code
			if rw.statusCode >= 400 {
				span.SetStatus(trace.StatusCodeError, fmt.Sprintf("HTTP %d", rw.statusCode))
			}
		})
	}
}

// responseWriter wrapper para capturar status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// ContextKey para passar informações de tracing no contexto
type ContextKey string

const (
	SpanContextKey = ContextKey("span")
	TraceIDKey     = ContextKey("trace_id")
)

// GetTraceID obtém o trace ID do contexto
func GetTraceID(ctx context.Context) string {
	if span := trace.SpanFromContext(ctx); span != nil {
		return span.SpanContext().TraceID().String()
	}
	return ""
}

// InjectHeaders injeta headers de tracing em uma requisição HTTP
func InjectHeaders(ctx context.Context, headers map[string]string) {
	propagator := propagation.TraceContext{}
	propagator.Inject(ctx, propagation.HeaderCarrier(headers))
}

// ExtractHeaders extrai contexto de tracing de headers HTTP
func ExtractHeaders(headers map[string]string) context.Context {
	propagator := propagation.TraceContext{}
	return propagator.Extract(context.Background(), propagation.HeaderCarrier(headers))
}

// GetTracer retorna o tracer configurado
func (tm *TracingManager) GetTracer() trace.Tracer {
	return tm.tracer
}

// GetTracerProvider retorna o tracer provider
func (tm *TracingManager) GetTracerProvider() trace.TracerProvider {
	return tm.tp
}

// IsEnabled retorna se o tracing está habilitado
func (tm *TracingManager) IsEnabled() bool {
	return tm.config.Enabled
}

// GetConfig retorna a configuração do tracing
func (tm *TracingManager) GetConfig() *TracingConfig {
	return tm.config
}