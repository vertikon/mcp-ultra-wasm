package observability

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// TelemetryConfig holds telemetry configuration
type TelemetryConfig struct {
	ServiceName    string        `yaml:"service_name"`
	ServiceVersion string        `yaml:"service_version"`
	Environment    string        `yaml:"environment"`
	JaegerEndpoint string        `yaml:"jaeger_endpoint"`
	OTLPEndpoint   string        `yaml:"otlp_endpoint"`
	MetricsPort    int           `yaml:"metrics_port"`
	SamplingRate   float64       `yaml:"sampling_rate"`
	SampleRate     float64       `yaml:"sample_rate"`
	BatchTimeout   time.Duration `yaml:"batch_timeout"`
	BatchSize      int           `yaml:"batch_size"`
	Enabled        bool          `yaml:"enabled"`
	Debug          bool          `yaml:"debug"`
	Exporter       string        `yaml:"exporter"`

	// Tracing specific
	TracingEnabled    bool          `yaml:"tracing_enabled"`
	TracingSampleRate float64       `yaml:"tracing_sample_rate"`
	TracingMaxSpans   int           `yaml:"tracing_max_spans"`
	TracingBatchSize  int           `yaml:"tracing_batch_size"`
	TracingTimeout    time.Duration `yaml:"tracing_timeout"`

	// Metrics specific
	MetricsEnabled   bool          `yaml:"metrics_enabled"`
	MetricsPath      string        `yaml:"metrics_path"`
	MetricsInterval  time.Duration `yaml:"metrics_interval"`
	HistogramBuckets []float64     `yaml:"histogram_buckets"`

	// Exporters configuration
	JaegerEnabled  bool              `yaml:"jaeger_enabled"`
	JaegerUser     string            `yaml:"jaeger_user"`
	JaegerPassword string            `yaml:"jaeger_password"`
	OTLPEnabled    bool              `yaml:"otlp_enabled"`
	OTLPInsecure   bool              `yaml:"otlp_insecure"`
	OTLPHeaders    map[string]string `yaml:"otlp_headers"`
	ConsoleEnabled bool              `yaml:"console_enabled"`
}

// TelemetryService manages OpenTelemetry instrumentation
type TelemetryService struct {
	config         TelemetryConfig
	logger         *zap.Logger
	tracerProvider trace.TracerProvider
	meterProvider  metric.MeterProvider
	tracer         trace.Tracer
	meter          metric.Meter

	// Business metrics
	requestCounter    metric.Int64Counter
	requestDuration   metric.Float64Histogram
	activeConnections metric.Int64UpDownCounter
	errorCounter      metric.Int64Counter
	taskMetrics       *TaskMetrics

	// System metrics
	cpuUsage    metric.Float64ObservableGauge
	memoryUsage metric.Float64ObservableGauge
	goroutines  metric.Int64ObservableGauge
}

// TaskMetrics holds task-specific metrics
type TaskMetrics struct {
	taskCreated   metric.Int64Counter
	taskCompleted metric.Int64Counter
	taskFailed    metric.Int64Counter
	taskDuration  metric.Float64Histogram

	tasksByStatus   metric.Int64ObservableGauge
	tasksByPriority metric.Int64ObservableGauge
}

// NewTelemetryService creates a new telemetry service
func NewTelemetryService(config TelemetryConfig, logger *zap.Logger) (*TelemetryService, error) {
	if !config.Enabled {
		logger.Info("Telemetry disabled")
		return &TelemetryService{
			config: config,
			logger: logger,
		}, nil
	}

	ts := &TelemetryService{
		config: config,
		logger: logger,
	}

	// Initialize resource
	res, err := ts.initResource()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize resource: %w", err)
	}

	// Initialize tracing
	if err := ts.initTracing(res); err != nil {
		return nil, fmt.Errorf("failed to initialize tracing: %w", err)
	}

	// Initialize metrics
	if err := ts.initMetrics(res); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize business metrics
	if err := ts.initBusinessMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize business metrics: %w", err)
	}

	// Initialize system metrics
	if err := ts.initSystemMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize system metrics: %w", err)
	}

	logger.Info("Telemetry initialized successfully",
		zap.String("service", config.ServiceName),
		zap.String("version", config.ServiceVersion),
		zap.String("environment", config.Environment))

	return ts, nil
}

// initResource creates the OpenTelemetry resource
func (ts *TelemetryService) initResource() (*resource.Resource, error) {
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(ts.config.ServiceName),
		semconv.ServiceVersion(ts.config.ServiceVersion),
		semconv.DeploymentEnvironment(ts.config.Environment),
		attribute.String("service.instance.id", generateInstanceID()),
		attribute.String("telemetry.sdk.name", "opentelemetry"),
		attribute.String("telemetry.sdk.language", "go"),
		attribute.String("telemetry.sdk.version", otel.Version()),
	), nil
}

// initTracing sets up distributed tracing
func (ts *TelemetryService) initTracing(res *resource.Resource) error {
	var exporter sdktrace.SpanExporter
	var err error

	// Choose exporter based on configuration
	// Note: Jaeger exporter removed as it was deprecated in July 2023
	// Use OTLP instead (Jaeger supports OTLP natively)
	if ts.config.OTLPEndpoint != "" {
		exporter, err = otlptracehttp.New(context.Background(),
			otlptracehttp.WithEndpoint(ts.config.OTLPEndpoint),
			otlptracehttp.WithInsecure(), // Use HTTPS in production
		)
		if err != nil {
			return fmt.Errorf("failed to create OTLP exporter: %w", err)
		}
		ts.logger.Info("Using OTLP exporter", zap.String("endpoint", ts.config.OTLPEndpoint))
	} else {
		// No exporter configured - use no-op for tests/disabled telemetry
		ts.logger.Debug("No tracing exporter configured, using no-op tracer")
		ts.tracerProvider = otel.GetTracerProvider()
		ts.tracer = ts.tracerProvider.Tracer(
			ts.config.ServiceName,
			trace.WithInstrumentationVersion(ts.config.ServiceVersion),
			trace.WithSchemaURL(semconv.SchemaURL),
		)
		return nil
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(ts.config.SamplingRate)),
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(ts.config.BatchTimeout),
			sdktrace.WithMaxExportBatchSize(ts.config.BatchSize),
		),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	ts.tracerProvider = tp
	ts.tracer = tp.Tracer(
		ts.config.ServiceName,
		trace.WithInstrumentationVersion(ts.config.ServiceVersion),
		trace.WithSchemaURL(semconv.SchemaURL),
	)

	return nil
}

// initMetrics sets up metrics collection
func (ts *TelemetryService) initMetrics(res *resource.Resource) error {
	// Create Prometheus exporter
	promExporter, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("failed to create Prometheus exporter: %w", err)
	}

	// Create meter provider
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(promExporter),
	)

	otel.SetMeterProvider(mp)

	ts.meterProvider = mp
	ts.meter = mp.Meter(
		ts.config.ServiceName,
		metric.WithInstrumentationVersion(ts.config.ServiceVersion),
		metric.WithSchemaURL(semconv.SchemaURL),
	)

	return nil
}

// initBusinessMetrics creates business-specific metrics
func (ts *TelemetryService) initBusinessMetrics() error {
	var err error

	// HTTP request metrics
	ts.requestCounter, err = ts.meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return fmt.Errorf("failed to create request counter: %w", err)
	}

	ts.requestDuration, err = ts.meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
	)
	if err != nil {
		return fmt.Errorf("failed to create request duration histogram: %w", err)
	}

	ts.activeConnections, err = ts.meter.Int64UpDownCounter(
		"http_active_connections",
		metric.WithDescription("Number of active HTTP connections"),
		metric.WithUnit("{connection}"),
	)
	if err != nil {
		return fmt.Errorf("failed to create active connections counter: %w", err)
	}

	ts.errorCounter, err = ts.meter.Int64Counter(
		"application_errors_total",
		metric.WithDescription("Total number of application errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		return fmt.Errorf("failed to create error counter: %w", err)
	}

	// Initialize task metrics
	ts.taskMetrics, err = ts.initTaskMetrics()
	if err != nil {
		return fmt.Errorf("failed to initialize task metrics: %w", err)
	}

	return nil
}

// initTaskMetrics creates task-specific metrics
func (ts *TelemetryService) initTaskMetrics() (*TaskMetrics, error) {
	taskMetrics := &TaskMetrics{}
	var err error

	taskMetrics.taskCreated, err = ts.meter.Int64Counter(
		"tasks_created_total",
		metric.WithDescription("Total number of tasks created"),
		metric.WithUnit("{task}"),
	)
	if err != nil {
		return nil, err
	}

	taskMetrics.taskCompleted, err = ts.meter.Int64Counter(
		"tasks_completed_total",
		metric.WithDescription("Total number of tasks completed"),
		metric.WithUnit("{task}"),
	)
	if err != nil {
		return nil, err
	}

	taskMetrics.taskFailed, err = ts.meter.Int64Counter(
		"tasks_failed_total",
		metric.WithDescription("Total number of tasks failed"),
		metric.WithUnit("{task}"),
	)
	if err != nil {
		return nil, err
	}

	taskMetrics.taskDuration, err = ts.meter.Float64Histogram(
		"task_duration_seconds",
		metric.WithDescription("Task processing duration in seconds"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.1, 0.5, 1, 2, 5, 10, 30, 60, 300),
	)
	if err != nil {
		return nil, err
	}

	// Observable gauges for current state
	taskMetrics.tasksByStatus, err = ts.meter.Int64ObservableGauge(
		"tasks_by_status",
		metric.WithDescription("Number of tasks by status"),
		metric.WithUnit("{task}"),
	)
	if err != nil {
		return nil, err
	}

	taskMetrics.tasksByPriority, err = ts.meter.Int64ObservableGauge(
		"tasks_by_priority",
		metric.WithDescription("Number of tasks by priority"),
		metric.WithUnit("{task}"),
	)
	if err != nil {
		return nil, err
	}

	return taskMetrics, nil
}

// initSystemMetrics creates system-level metrics
func (ts *TelemetryService) initSystemMetrics() error {
	var err error

	ts.cpuUsage, err = ts.meter.Float64ObservableGauge(
		"system_cpu_usage_percent",
		metric.WithDescription("CPU usage percentage"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return err
	}

	ts.memoryUsage, err = ts.meter.Float64ObservableGauge(
		"system_memory_usage_bytes",
		metric.WithDescription("Memory usage in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return err
	}

	ts.goroutines, err = ts.meter.Int64ObservableGauge(
		"go_goroutines",
		metric.WithDescription("Number of goroutines"),
		metric.WithUnit("{goroutine}"),
	)
	if err != nil {
		return err
	}

	// Register callbacks for system metrics
	_, err = ts.meter.RegisterCallback(
		ts.collectSystemMetrics,
		ts.cpuUsage,
		ts.memoryUsage,
		ts.goroutines,
	)
	if err != nil {
		return fmt.Errorf("failed to register system metrics callback: %w", err)
	}

	return nil
}

// Start is a no-op method for compatibility with Service.Start()
// Initialization happens in NewTelemetryService
func (ts *TelemetryService) Start(ctx context.Context) error {
	if ts.config.Debug {
		ts.logger.Debug("TelemetryService.Start called (initialization already complete)")
	}
	return nil
}

// Stop gracefully shuts down telemetry
func (ts *TelemetryService) Stop(ctx context.Context) error {
	return ts.Shutdown(ctx)
}

// Tracer returns the configured tracer
func (ts *TelemetryService) Tracer() trace.Tracer {
	if ts.tracer == nil {
		return otel.Tracer("noop")
	}
	return ts.tracer
}

// GetTracer returns a named tracer from the tracer provider
func (ts *TelemetryService) GetTracer(name string) trace.Tracer {
	if ts.tracerProvider == nil {
		return otel.Tracer(name)
	}
	return ts.tracerProvider.Tracer(
		name,
		trace.WithInstrumentationVersion(ts.config.ServiceVersion),
		trace.WithSchemaURL(semconv.SchemaURL),
	)
}

// Meter returns the configured meter
func (ts *TelemetryService) Meter() metric.Meter {
	if ts.meter == nil {
		return otel.Meter("noop")
	}
	return ts.meter
}

// GetMeter returns a named meter from the meter provider
func (ts *TelemetryService) GetMeter(name string) metric.Meter {
	if ts.meterProvider == nil {
		return otel.Meter(name)
	}
	return ts.meterProvider.Meter(
		name,
		metric.WithInstrumentationVersion(ts.config.ServiceVersion),
		metric.WithSchemaURL(semconv.SchemaURL),
	)
}

// StartSpan starts a new trace span
func (ts *TelemetryService) StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if !ts.config.Enabled {
		return ctx, trace.SpanFromContext(ctx)
	}
	return ts.tracer.Start(ctx, spanName, opts...)
}

// RecordHTTPRequest records HTTP request metrics
func (ts *TelemetryService) RecordHTTPRequest(method, path, status string, duration time.Duration) {
	if !ts.config.Enabled {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.String("http.status", status),
	}

	ts.requestCounter.Add(context.Background(), 1, metric.WithAttributes(attrs...))
	ts.requestDuration.Record(context.Background(), duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordError records application errors
func (ts *TelemetryService) RecordError(errorType, component string) {
	if !ts.config.Enabled {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("error.type", errorType),
		attribute.String("component", component),
	}

	ts.errorCounter.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

// RecordTaskCreated records task creation metrics
func (ts *TelemetryService) RecordTaskCreated(priority string) {
	if !ts.config.Enabled || ts.taskMetrics == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("task.priority", priority),
	}

	ts.taskMetrics.taskCreated.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

// RecordTaskCompleted records task completion metrics
func (ts *TelemetryService) RecordTaskCompleted(priority string, duration time.Duration) {
	if !ts.config.Enabled || ts.taskMetrics == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("task.priority", priority),
	}

	ts.taskMetrics.taskCompleted.Add(context.Background(), 1, metric.WithAttributes(attrs...))
	ts.taskMetrics.taskDuration.Record(context.Background(), duration.Seconds(), metric.WithAttributes(attrs...))
}

// RecordTaskFailed records task failure metrics
func (ts *TelemetryService) RecordTaskFailed(priority string, reason string) {
	if !ts.config.Enabled || ts.taskMetrics == nil {
		return
	}

	attrs := []attribute.KeyValue{
		attribute.String("task.priority", priority),
		attribute.String("failure.reason", reason),
	}

	ts.taskMetrics.taskFailed.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

// IncrementActiveConnections increments active connections counter
func (ts *TelemetryService) IncrementActiveConnections() {
	if !ts.config.Enabled {
		return
	}
	ts.activeConnections.Add(context.Background(), 1)
}

// DecrementActiveConnections decrements active connections counter
func (ts *TelemetryService) DecrementActiveConnections() {
	if !ts.config.Enabled {
		return
	}
	ts.activeConnections.Add(context.Background(), -1)
}

// IncrementRequestCounter increments the HTTP request counter
func (ts *TelemetryService) IncrementRequestCounter(ctx context.Context, method, path, statusCode string) error {
	if !ts.config.Enabled || ts.requestCounter == nil {
		return nil
	}

	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.String("http.status", statusCode),
	}

	ts.requestCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	return nil
}

// RecordRequestDuration records HTTP request duration
func (ts *TelemetryService) RecordRequestDuration(ctx context.Context, method, path string, duration time.Duration) error {
	if !ts.config.Enabled || ts.requestDuration == nil {
		return nil
	}

	attrs := []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.path", path),
	}

	ts.requestDuration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	return nil
}

// IncrementErrorCounter increments the error counter
func (ts *TelemetryService) IncrementErrorCounter(ctx context.Context, errorType, errorCode string) error {
	if !ts.config.Enabled || ts.errorCounter == nil {
		return nil
	}

	attrs := []attribute.KeyValue{
		attribute.String("error.type", errorType),
		attribute.String("error.code", errorCode),
	}

	ts.errorCounter.Add(ctx, 1, metric.WithAttributes(attrs...))
	return nil
}

// RecordProcessingTime records processing time for an operation
func (ts *TelemetryService) RecordProcessingTime(ctx context.Context, operationType string, duration time.Duration) error {
	if !ts.config.Enabled || ts.meter == nil {
		return nil
	}

	// Create or get a histogram for processing time
	histogram, err := ts.meter.Float64Histogram(
		"processing_time_seconds",
		metric.WithDescription("Processing time in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	attrs := []attribute.KeyValue{
		attribute.String("operation.type", operationType),
	}

	histogram.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
	return nil
}

// HTTPMiddleware returns an HTTP middleware that instruments requests with tracing and metrics
func (ts *TelemetryService) HTTPMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !ts.config.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// Start span for tracing
			ctx, span := ts.tracer.Start(r.Context(), r.Method+" "+r.URL.Path,
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.host", r.Host),
					attribute.String("http.scheme", r.URL.Scheme),
					attribute.String("http.user_agent", r.UserAgent()),
				),
			)
			defer span.End()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Serve the request
			next.ServeHTTP(wrapped, r.WithContext(ctx))

			// Record metrics
			duration := time.Since(start)
			statusCode := fmt.Sprintf("%d", wrapped.statusCode)

			if err := ts.IncrementRequestCounter(ctx, r.Method, r.URL.Path, statusCode); err != nil {
				ts.logger.Warn("Failed to increment request counter", zap.Error(err))
			}
			if err := ts.RecordRequestDuration(ctx, r.Method, r.URL.Path, duration); err != nil {
				ts.logger.Warn("Failed to record request duration", zap.Error(err))
			}

			// Add status to span
			span.SetAttributes(attribute.Int("http.status_code", wrapped.statusCode))
			if wrapped.statusCode >= 400 {
				span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrapped.statusCode))
			}
		})
	}
}

// HealthCheck returns health status of the telemetry service
func (ts *TelemetryService) HealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
		"components": map[string]interface{}{
			"telemetry": map[string]interface{}{
				"enabled": ts.config.Enabled,
				"service": ts.config.ServiceName,
			},
		},
	}

	if !ts.config.Enabled {
		health["status"] = "disabled"
		return health
	}

	components := health["components"].(map[string]interface{})

	// Check tracer provider
	if ts.tracerProvider != nil {
		components["tracing"] = map[string]interface{}{
			"status":   "active",
			"exporter": ts.config.Exporter,
		}
	} else {
		components["tracing"] = map[string]interface{}{
			"status": "inactive",
		}
	}

	// Check meter provider
	if ts.meterProvider != nil {
		components["metrics"] = map[string]interface{}{
			"status": "active",
		}
	} else {
		components["metrics"] = map[string]interface{}{
			"status": "inactive",
		}
	}

	return health
}

// collectSystemMetrics collects system-level metrics
func (ts *TelemetryService) collectSystemMetrics(ctx context.Context, observer metric.Observer) error {
	// Collect system metrics (simplified implementation)
	// In production, use proper system metric collection libraries

	// CPU usage (mock implementation)
	observer.ObserveFloat64(ts.cpuUsage, 0.0) // Would collect actual CPU usage

	// Memory usage (mock implementation)
	observer.ObserveFloat64(ts.memoryUsage, 0.0) // Would collect actual memory usage

	// Goroutines
	observer.ObserveInt64(ts.goroutines, int64(runtime.NumGoroutine()))

	return nil
}

// Shutdown gracefully shuts down the telemetry service
func (ts *TelemetryService) Shutdown(ctx context.Context) error {
	if !ts.config.Enabled {
		return nil
	}

	var err error

	// Shutdown tracer provider
	if tp, ok := ts.tracerProvider.(*sdktrace.TracerProvider); ok {
		if shutdownErr := tp.Shutdown(ctx); shutdownErr != nil {
			err = fmt.Errorf("failed to shutdown tracer provider: %w", shutdownErr)
		}
	}

	// Shutdown meter provider
	if mp, ok := ts.meterProvider.(*sdkmetric.MeterProvider); ok {
		if shutdownErr := mp.Shutdown(ctx); shutdownErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to shutdown meter provider: %w", err, shutdownErr)
			} else {
				err = fmt.Errorf("failed to shutdown meter provider: %w", shutdownErr)
			}
		}
	}

	ts.logger.Info("Telemetry service shutdown complete")
	return err
}

// generateInstanceID generates a unique instance identifier
func generateInstanceID() string {
	// In production, this could be based on hostname, pod name, etc.
	return fmt.Sprintf("instance-%d", time.Now().UnixNano())
}
