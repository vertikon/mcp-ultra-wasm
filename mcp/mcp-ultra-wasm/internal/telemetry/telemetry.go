package telemetry

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/httpx"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/logger"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/metrics"
)

var (
	// HTTP Metrics
	httpRequestsTotal = metrics.NewCounterVec(
		"http_requests_total",
		"Total number of HTTP requests",
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = metrics.NewHistogramVec(
		"http_request_duration_seconds",
		"Duration of HTTP requests in seconds",
		[]string{"method", "path", "status"},
	)

	// Business Metrics
	tasksTotal = metrics.NewCounterVec(
		"tasks_total",
		"Total number of tasks",
		[]string{"status", "priority"},
	)

	tasksProcessingTime = metrics.NewHistogramVec(
		"task_processing_seconds",
		"Time taken to process tasks",
		[]string{"operation"},
	)

	// System Metrics
	databaseConnections = metrics.NewGaugeVec(
		"database_connections",
		"Number of database connections",
		[]string{"database", "state"},
	)

	cacheOperations = metrics.NewCounterVec(
		"cache_operations_total",
		"Total number of cache operations",
		[]string{"operation", "result"},
	)
)

// Telemetry holds telemetry configuration and clients
type Telemetry struct {
	meter  metric.Meter
	logger *logger.Logger
}

// Init initializes telemetry system
func Init(_ config.TelemetryConfig) (*Telemetry, error) {
	log, err := logger.NewLogger()
	if err != nil {
		return nil, fmt.Errorf("creating logger: %w", err)
	}

	// Initialize Prometheus exporter
	exporter, err := promexporter.New()
	if err != nil {
		return nil, fmt.Errorf("creating prometheus exporter: %w", err)
	}

	// Create meter provider
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	// Create meter
	meter := provider.Meter("github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm")

	return &Telemetry{
		meter:  meter,
		logger: log,
	}, nil
}

// HTTPMetrics middleware for HTTP request metrics
func HTTPMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		ww := httpx.NewWrapResponseWriter(w, r.ProtoMajor)

		// Process request
		next.ServeHTTP(ww, r)

		// Record metrics
		duration := time.Since(start)
		status := strconv.Itoa(ww.Status())

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, status).Observe(duration.Seconds())
	})
}

// RecordTaskCreated records task creation metrics
func RecordTaskCreated(status, priority string) {
	tasksTotal.WithLabelValues(status, priority).Inc()
}

// RecordTaskProcessingTime records task processing time
func RecordTaskProcessingTime(operation string, duration time.Duration) {
	tasksProcessingTime.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordDatabaseConnections records database connection metrics
func RecordDatabaseConnections(database, state string, count float64) {
	databaseConnections.WithLabelValues(database, state).Set(count)
}

// RecordCacheOperation records cache operation metrics
func RecordCacheOperation(operation, result string) {
	cacheOperations.WithLabelValues(operation, result).Inc()
}

// TaskMetrics handles task-related metrics
type TaskMetrics struct {
	createdCounter   metric.Int64Counter
	completedCounter metric.Int64Counter
	processingTime   metric.Float64Histogram
	meter            metric.Meter
}

// NewTaskMetrics creates new task metrics
func NewTaskMetrics(meter metric.Meter) (*TaskMetrics, error) {
	createdCounter, err := meter.Int64Counter(
		"tasks_created_total",
		metric.WithDescription("Total number of tasks created"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating tasks created counter: %w", err)
	}

	completedCounter, err := meter.Int64Counter(
		"tasks_completed_total",
		metric.WithDescription("Total number of tasks completed"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating tasks completed counter: %w", err)
	}

	processingTime, err := meter.Float64Histogram(
		"task_processing_duration_seconds",
		metric.WithDescription("Time taken to process tasks"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating task processing time histogram: %w", err)
	}

	return &TaskMetrics{
		createdCounter:   createdCounter,
		completedCounter: completedCounter,
		processingTime:   processingTime,
		meter:            meter,
	}, nil
}

// RecordTaskCreated records a task creation
func (tm *TaskMetrics) RecordTaskCreated(ctx context.Context, priority, status string) {
	tm.createdCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("priority", priority),
			attribute.String("status", status),
		),
	)
}

// RecordTaskCompleted records a task completion
func (tm *TaskMetrics) RecordTaskCompleted(ctx context.Context, priority string, processingTime time.Duration) {
	tm.completedCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("priority", priority),
		),
	)

	tm.processingTime.Record(ctx, processingTime.Seconds(),
		metric.WithAttributes(
			attribute.String("priority", priority),
		),
	)
}

// FeatureFlagMetrics handles feature flag metrics
type FeatureFlagMetrics struct {
	evaluations metric.Int64Counter
	meter       metric.Meter
}

// NewFeatureFlagMetrics creates new feature flag metrics
func NewFeatureFlagMetrics(meter metric.Meter) (*FeatureFlagMetrics, error) {
	evaluations, err := meter.Int64Counter(
		"feature_flag_evaluations_total",
		metric.WithDescription("Total number of feature flag evaluations"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating feature flag evaluations counter: %w", err)
	}

	return &FeatureFlagMetrics{
		evaluations: evaluations,
		meter:       meter,
	}, nil
}

// RecordEvaluation records a feature flag evaluation
func (ffm *FeatureFlagMetrics) RecordEvaluation(ctx context.Context, key string, enabled bool) {
	ffm.evaluations.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("flag_key", key),
			attribute.Bool("enabled", enabled),
		),
	)
}
