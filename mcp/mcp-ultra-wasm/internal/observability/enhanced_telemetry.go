package observability

import (
	"context"
	"fmt"
	"net/http"
	goruntime "runtime"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	metricSDK "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// EnhancedTelemetryService provides comprehensive observability
type EnhancedTelemetryService struct {
	config   TelemetryConfig
	logger   *zap.Logger
	tracer   oteltrace.Tracer
	meter    metric.Meter
	resource *resource.Resource

	// Prometheus metrics
	httpDuration      prometheus.HistogramVec
	httpRequests      prometheus.CounterVec
	activeConnections prometheus.Gauge
	dbConnections     prometheus.GaugeVec
	cacheHitRatio     prometheus.GaugeVec

	// OpenTelemetry metrics
	requestCounter  metric.Int64Counter
	requestDuration metric.Float64Histogram
	errorCounter    metric.Int64Counter
	cpuUsage        metric.Float64ObservableGauge
	memoryUsage     metric.Int64ObservableGauge
	goroutineCount  metric.Int64ObservableGauge

	// Custom business metrics
	taskCounter     metric.Int64Counter
	userActiveGauge metric.Int64ObservableGauge

	// Metric collection
	metricCollectors map[string]MetricCollector
	collectorMutex   sync.RWMutex

	// Service health
	healthCheckers  map[string]HealthChecker
	healthMutex     sync.RWMutex
	lastHealthCheck map[string]HealthResult

	// Tracing
	activeSpans map[string]oteltrace.Span

	// Alerting
	alertRules         map[string]AlertRule
	alertMutex         sync.RWMutex
	alertNotifications chan AlertNotification
}

// MetricCollector interface for custom metric collection
type MetricCollector interface {
	CollectMetrics(ctx context.Context) (map[string]float64, error)
	GetMetricNames() []string
}

// HealthChecker interface for service health checks
type HealthChecker interface {
	CheckHealth(ctx context.Context) HealthResult
	GetName() string
}

// HealthResult represents the result of a health check
type HealthResult struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	Name        string        `json:"name"`
	Metric      string        `json:"metric"`
	Condition   string        `json:"condition"` // "gt", "lt", "eq"
	Threshold   float64       `json:"threshold"`
	Duration    time.Duration `json:"duration"`
	Severity    string        `json:"severity"`
	Description string        `json:"description"`
	LastFired   time.Time     `json:"last_fired"`
}

// AlertNotification represents an alert notification
type AlertNotification struct {
	Rule      AlertRule         `json:"rule"`
	Value     float64           `json:"value"`
	Timestamp time.Time         `json:"timestamp"`
	Labels    map[string]string `json:"labels"`
}

// NewEnhancedTelemetryService creates a new enhanced telemetry service
func NewEnhancedTelemetryService(config TelemetryConfig, logger *zap.Logger) (*EnhancedTelemetryService, error) {
	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("mcp-ultra-wasm"),
			semconv.ServiceVersion("1.0.0"),
			semconv.ServiceInstanceID("instance-1"),
			attribute.String("environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating resource: %w", err)
	}

	service := &EnhancedTelemetryService{
		config:             config,
		logger:             logger,
		resource:           res,
		metricCollectors:   make(map[string]MetricCollector),
		healthCheckers:     make(map[string]HealthChecker),
		lastHealthCheck:    make(map[string]HealthResult),
		activeSpans:        make(map[string]oteltrace.Span),
		alertRules:         make(map[string]AlertRule),
		alertNotifications: make(chan AlertNotification, 1000),
	}

	// Initialize tracing
	if err := service.initTracing(); err != nil {
		return nil, fmt.Errorf("initializing tracing: %w", err)
	}

	// Initialize metrics
	if err := service.initMetrics(); err != nil {
		return nil, fmt.Errorf("initializing metrics: %w", err)
	}

	// Initialize Prometheus metrics
	service.initPrometheusMetrics()

	// Start background workers
	go service.metricsCollectionWorker()
	go service.healthCheckWorker()
	go service.alertingWorker()

	// Start runtime instrumentation
	if err := runtime.Start(); err != nil {
		logger.Warn("Failed to start runtime instrumentation", zap.Error(err))
	}

	return service, nil
}

// initTracing initializes OpenTelemetry tracing
func (ets *EnhancedTelemetryService) initTracing() error {
	// Note: This service uses a simplified trace setup
	// For production use, configure a proper OTLP exporter via the telemetry package
	// This method is deprecated and should be replaced with proper telemetry.TracingProvider
	ets.logger.Warn("EnhancedTelemetryService.initTracing is deprecated - use telemetry.TracingProvider instead")
	return nil

	/* Removed deprecated Jaeger support
	var exporter trace.SpanExporter
	var err error

	switch ets.config.Exporter {
	case "jaeger":
		exporter, err = jaeger.New(jaeger.WithCollectorEndpoint(
			jaeger.WithEndpoint(ets.config.JaegerEndpoint),
		))
	default:
		return fmt.Errorf("unsupported trace exporter: %s", ets.config.Exporter)
	}

	if err != nil {
		return fmt.Errorf("creating trace exporter: %w", err)
	}
	*/

	/* Deprecated code - commented out
	// Create trace provider
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(ets.resource),
		trace.WithSampler(trace.TraceIDRatioBased(ets.config.SampleRate)),
	)

	// Set global trace provider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Create tracer
	ets.tracer = tp.Tracer("mcp-ultra-wasm")

	return nil
	*/
}

// initMetrics initializes OpenTelemetry metrics
func (ets *EnhancedTelemetryService) initMetrics() error {
	// Create Prometheus exporter
	promExporter, err := promexporter.New()
	if err != nil {
		return fmt.Errorf("creating prometheus exporter: %w", err)
	}

	// Create metric provider
	mp := metricSDK.NewMeterProvider(
		metricSDK.WithResource(ets.resource),
		metricSDK.WithReader(promExporter),
	)

	// Set global metric provider
	otel.SetMeterProvider(mp)

	// Create meter
	ets.meter = mp.Meter("mcp-ultra-wasm")

	// Initialize metrics
	if err := ets.createMetrics(); err != nil {
		return fmt.Errorf("creating metrics: %w", err)
	}

	return nil
}

// createMetrics creates all OpenTelemetry metrics
func (ets *EnhancedTelemetryService) createMetrics() error {
	var err error

	// Request counter
	ets.requestCounter, err = ets.meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		return fmt.Errorf("creating request counter: %w", err)
	}

	// Request duration
	ets.requestDuration, err = ets.meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return fmt.Errorf("creating request duration histogram: %w", err)
	}

	// Error counter
	ets.errorCounter, err = ets.meter.Int64Counter(
		"errors_total",
		metric.WithDescription("Total number of errors"),
	)
	if err != nil {
		return fmt.Errorf("creating error counter: %w", err)
	}

	// CPU usage gauge
	ets.cpuUsage, err = ets.meter.Float64ObservableGauge(
		"cpu_usage_percent",
		metric.WithDescription("CPU usage percentage"),
		metric.WithUnit("%"),
	)
	if err != nil {
		return fmt.Errorf("creating CPU usage gauge: %w", err)
	}

	// Memory usage gauge
	ets.memoryUsage, err = ets.meter.Int64ObservableGauge(
		"memory_usage_bytes",
		metric.WithDescription("Memory usage in bytes"),
		metric.WithUnit("By"),
	)
	if err != nil {
		return fmt.Errorf("creating memory usage gauge: %w", err)
	}

	// Goroutine count gauge
	ets.goroutineCount, err = ets.meter.Int64ObservableGauge(
		"goroutines_total",
		metric.WithDescription("Number of goroutines"),
	)
	if err != nil {
		return fmt.Errorf("creating goroutine count gauge: %w", err)
	}

	// Task counter
	ets.taskCounter, err = ets.meter.Int64Counter(
		"tasks_total",
		metric.WithDescription("Total number of tasks processed"),
	)
	if err != nil {
		return fmt.Errorf("creating task counter: %w", err)
	}

	// User active gauge
	ets.userActiveGauge, err = ets.meter.Int64ObservableGauge(
		"users_active",
		metric.WithDescription("Number of active users"),
	)
	if err != nil {
		return fmt.Errorf("creating user active gauge: %w", err)
	}

	// Register callbacks for observable gauges
	_, err = ets.meter.RegisterCallback(
		ets.collectRuntimeMetrics,
		ets.cpuUsage,
		ets.memoryUsage,
		ets.goroutineCount,
	)
	if err != nil {
		return fmt.Errorf("registering runtime metrics callback: %w", err)
	}

	return nil
}

// initPrometheusMetrics initializes Prometheus metrics
func (ets *EnhancedTelemetryService) initPrometheusMetrics() {
	// HTTP request duration histogram
	ets.httpDuration = *promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "mcp_ultra_wasm_http_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP request counter
	ets.httpRequests = *promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "mcp_ultra_wasm_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Active connections gauge
	ets.activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "mcp_ultra_wasm_active_connections",
			Help: "Number of active connections",
		},
	)

	// Database connections gauge
	ets.dbConnections = *promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_ultra_wasm_db_connections",
			Help: "Number of database connections",
		},
		[]string{"database", "status"},
	)

	// Cache hit ratio gauge
	ets.cacheHitRatio = *promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "mcp_ultra_wasm_cache_hit_ratio",
			Help: "Cache hit ratio",
		},
		[]string{"cache_name"},
	)
}

// collectRuntimeMetrics collects runtime metrics
func (ets *EnhancedTelemetryService) collectRuntimeMetrics(ctx context.Context, observer metric.Observer) error {
	var m goruntime.MemStats
	goruntime.ReadMemStats(&m)

	// Memory metrics
	observer.ObserveInt64(ets.memoryUsage, int64(m.Alloc))

	// Goroutine count
	observer.ObserveInt64(ets.goroutineCount, int64(goruntime.NumGoroutine()))

	// CPU usage (simplified - in production use proper CPU monitoring)
	observer.ObserveFloat64(ets.cpuUsage, 50.0) // Placeholder

	return nil
}

// StartSpan starts a new tracing span
func (ets *EnhancedTelemetryService) StartSpan(ctx context.Context, name string, opts ...oteltrace.SpanStartOption) (context.Context, oteltrace.Span) {
	return ets.tracer.Start(ctx, name, opts...)
}

// RecordRequest records HTTP request metrics
func (ets *EnhancedTelemetryService) RecordRequest(method, endpoint, status string, duration time.Duration) {
	// OpenTelemetry metrics
	labels := []attribute.KeyValue{
		attribute.String("method", method),
		attribute.String("endpoint", endpoint),
		attribute.String("status", status),
	}

	ets.requestCounter.Add(context.Background(), 1, metric.WithAttributes(labels...))
	ets.requestDuration.Record(context.Background(), duration.Seconds(), metric.WithAttributes(labels...))

	// Prometheus metrics
	ets.httpRequests.WithLabelValues(method, endpoint, status).Inc()
	ets.httpDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
}

// RecordError records error metrics
func (ets *EnhancedTelemetryService) RecordError(ctx context.Context, errorType, component string) {
	labels := []attribute.KeyValue{
		attribute.String("error_type", errorType),
		attribute.String("component", component),
	}

	ets.errorCounter.Add(ctx, 1, metric.WithAttributes(labels...))
}

// RecordTask records task processing metrics
func (ets *EnhancedTelemetryService) RecordTask(ctx context.Context, taskType, status string) {
	labels := []attribute.KeyValue{
		attribute.String("task_type", taskType),
		attribute.String("status", status),
	}

	ets.taskCounter.Add(ctx, 1, metric.WithAttributes(labels...))
}

// UpdateConnectionCount updates active connection count
func (ets *EnhancedTelemetryService) UpdateConnectionCount(count float64) {
	ets.activeConnections.Set(count)
}

// UpdateDatabaseConnections updates database connection metrics
func (ets *EnhancedTelemetryService) UpdateDatabaseConnections(database, status string, count float64) {
	ets.dbConnections.WithLabelValues(database, status).Set(count)
}

// UpdateCacheHitRatio updates cache hit ratio metrics
func (ets *EnhancedTelemetryService) UpdateCacheHitRatio(cacheName string, ratio float64) {
	ets.cacheHitRatio.WithLabelValues(cacheName).Set(ratio)
}

// RegisterMetricCollector registers a custom metric collector
func (ets *EnhancedTelemetryService) RegisterMetricCollector(name string, collector MetricCollector) {
	ets.collectorMutex.Lock()
	defer ets.collectorMutex.Unlock()
	ets.metricCollectors[name] = collector
}

// RegisterHealthChecker registers a health checker
func (ets *EnhancedTelemetryService) RegisterHealthChecker(checker HealthChecker) {
	ets.healthMutex.Lock()
	defer ets.healthMutex.Unlock()
	ets.healthCheckers[checker.GetName()] = checker
}

// RegisterAlertRule registers an alert rule
func (ets *EnhancedTelemetryService) RegisterAlertRule(rule AlertRule) {
	ets.alertMutex.Lock()
	defer ets.alertMutex.Unlock()
	ets.alertRules[rule.Name] = rule
}

// metricsCollectionWorker collects custom metrics periodically
func (ets *EnhancedTelemetryService) metricsCollectionWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ets.collectorMutex.RLock()
		for name, collector := range ets.metricCollectors {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			metrics, err := collector.CollectMetrics(ctx)
			cancel()

			if err != nil {
				ets.logger.Error("Failed to collect custom metrics",
					zap.String("collector", name),
					zap.Error(err),
				)
				continue
			}

			for metricName, value := range metrics {
				ets.logger.Debug("Custom metric collected",
					zap.String("collector", name),
					zap.String("metric", metricName),
					zap.Float64("value", value),
				)
			}
		}
		ets.collectorMutex.RUnlock()
	}
}

// healthCheckWorker performs health checks periodically
func (ets *EnhancedTelemetryService) healthCheckWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ets.healthMutex.RLock()
		for name, checker := range ets.healthCheckers {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			result := checker.CheckHealth(ctx)
			cancel()

			ets.healthMutex.RUnlock()
			ets.healthMutex.Lock()
			ets.lastHealthCheck[name] = result
			ets.healthMutex.Unlock()
			ets.healthMutex.RLock()

			ets.logger.Debug("Health check completed",
				zap.String("checker", name),
				zap.String("status", result.Status),
				zap.Duration("duration", result.Duration),
			)
		}
		ets.healthMutex.RUnlock()
	}
}

// alertingWorker processes alert notifications
func (ets *EnhancedTelemetryService) alertingWorker() {
	for notification := range ets.alertNotifications {
		ets.logger.Warn("Alert triggered",
			zap.String("rule", notification.Rule.Name),
			zap.Float64("value", notification.Value),
			zap.Float64("threshold", notification.Rule.Threshold),
			zap.String("severity", notification.Rule.Severity),
		)

		// Here you would send notifications to external systems
		// like Slack, email, PagerDuty, etc.
	}
}

// GetHealthStatus returns the current health status
func (ets *EnhancedTelemetryService) GetHealthStatus() map[string]HealthResult {
	ets.healthMutex.RLock()
	defer ets.healthMutex.RUnlock()

	status := make(map[string]HealthResult)
	for name, result := range ets.lastHealthCheck {
		status[name] = result
	}
	return status
}

// CreateSpanWithError creates a span and records an error if present
func (ets *EnhancedTelemetryService) CreateSpanWithError(ctx context.Context, name string, err error, attrs ...attribute.KeyValue) {
	ctx, span := ets.StartSpan(ctx, name)
	defer span.End()

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ets.RecordError(ctx, "span_error", name)
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// HTTPMiddleware provides HTTP observability middleware
func (ets *EnhancedTelemetryService) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create span for the request
		ctx, span := ets.StartSpan(r.Context(), fmt.Sprintf("%s %s", r.Method, r.URL.Path))
		defer span.End()

		// Add request attributes
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.scheme", r.URL.Scheme),
			attribute.String("http.host", r.Host),
			attribute.String("user_agent", r.UserAgent()),
		)

		// Wrap response writer to capture status
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		// Execute request with tracing context
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Record metrics
		duration := time.Since(start)
		status := fmt.Sprintf("%d", wrapped.statusCode)

		ets.RecordRequest(r.Method, r.URL.Path, status, duration)

		// Update span with response info
		span.SetAttributes(
			attribute.Int("http.status_code", wrapped.statusCode),
			attribute.Int64("http.response_size", wrapped.bytesWritten),
		)

		if wrapped.statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", wrapped.statusCode))
			ets.RecordError(ctx, "http_error", r.URL.Path)
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture metrics
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// Shutdown gracefully shuts down the telemetry service
func (ets *EnhancedTelemetryService) Shutdown(ctx context.Context) error {
	close(ets.alertNotifications)
	ets.logger.Info("Enhanced telemetry service shutdown completed")
	return nil
}
