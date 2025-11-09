package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap/zaptest"
)

func createTestTelemetryConfig() TelemetryConfig {
	return TelemetryConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		JaegerEndpoint: "", // Disabled for tests
		OTLPEndpoint:   "", // Disabled for tests
		MetricsPort:    8080,
		SamplingRate:   1.0, // Sample all traces in tests
		BatchTimeout:   time.Second,
		BatchSize:      100,
		Enabled:        true,
		Debug:          true,
	}
}

func TestTelemetryService_Creation(t *testing.T) {
	config := createTestTelemetryConfig()
	logger := zaptest.NewLogger(t)

	service, err := NewTelemetryService(config, logger)
	require.NoError(t, err)
	require.NotNil(t, service)

	assert.Equal(t, config.ServiceName, service.config.ServiceName)
	assert.Equal(t, config.Environment, service.config.Environment)
	assert.Equal(t, config.Enabled, service.config.Enabled)
}

func TestTelemetryService_StartStop(t *testing.T) {
	config := createTestTelemetryConfig()
	logger := zaptest.NewLogger(t)

	service, err := NewTelemetryService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test Start
	err = service.Start(ctx)
	assert.NoError(t, err)

	// Verify tracer and meter are available
	tracer := service.GetTracer("test")
	assert.NotNil(t, tracer)

	meter := service.GetMeter("test")
	assert.NotNil(t, meter)

	// Test Stop
	err = service.Stop(ctx)
	assert.NoError(t, err)
}

func TestTelemetryService_Tracing(t *testing.T) {
	config := createTestTelemetryConfig()
	logger := zaptest.NewLogger(t)

	service, err := NewTelemetryService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = service.Start(ctx)
	require.NoError(t, err)
	defer func() {
		if stopErr := service.Stop(ctx); stopErr != nil {
			t.Logf("Warning: failed to stop service: %v", stopErr)
		}
	}()

	tracer := service.GetTracer("test-component")

	// Create a span
	ctx, span := tracer.Start(ctx, "test-operation")
	defer span.End()

	// Verify span was created
	assert.True(t, span.IsRecording())
	assert.NotEqual(t, trace.SpanID{}, span.SpanContext().SpanID())
	assert.NotEqual(t, trace.TraceID{}, span.SpanContext().TraceID())

	// Add attributes to span
	span.SetAttributes(
		attribute.String("test.attribute", "test-value"),
		attribute.Int("test.number", 42),
	)

	// Create child span
	_, childSpan := tracer.Start(ctx, "child-operation")
	childSpan.SetAttributes(attribute.String("child.attribute", "child-value"))
	childSpan.End()
}

func TestTelemetryService_Metrics(t *testing.T) {
	config := createTestTelemetryConfig()
	logger := zaptest.NewLogger(t)

	service, err := NewTelemetryService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = service.Start(ctx)
	require.NoError(t, err)
	defer func() {
		if stopErr := service.Stop(ctx); stopErr != nil {
			t.Logf("Warning: failed to stop service: %v", stopErr)
		}
	}()

	meter := service.GetMeter("test-metrics")

	// Create a counter
	counter, err := meter.Int64Counter(
		"test_counter",
		metric.WithDescription("A test counter"),
		metric.WithUnit("{count}"),
	)
	require.NoError(t, err)

	// Increment counter
	counter.Add(ctx, 1, metric.WithAttributes(
		attribute.String("test.label", "value1"),
	))
	counter.Add(ctx, 5, metric.WithAttributes(
		attribute.String("test.label", "value2"),
	))

	// Create a histogram
	histogram, err := meter.Float64Histogram(
		"test_histogram",
		metric.WithDescription("A test histogram"),
		metric.WithUnit("ms"),
	)
	require.NoError(t, err)

	// Record histogram values
	histogram.Record(ctx, 100.5, metric.WithAttributes(
		attribute.String("operation", "test"),
	))
	histogram.Record(ctx, 250.3, metric.WithAttributes(
		attribute.String("operation", "test"),
	))
}

func TestTelemetryService_BusinessMetrics(t *testing.T) {
	config := createTestTelemetryConfig()
	logger := zaptest.NewLogger(t)

	service, err := NewTelemetryService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = service.Start(ctx)
	require.NoError(t, err)
	defer func() {
		if stopErr := service.Stop(ctx); stopErr != nil {
			t.Logf("Warning: failed to stop service: %v", stopErr)
		}
	}()

	// Test increment request counter
	err = service.IncrementRequestCounter(ctx, "GET", "/api/test", "200")
	assert.NoError(t, err)

	err = service.IncrementRequestCounter(ctx, "POST", "/api/test", "201")
	assert.NoError(t, err)

	// Test record request duration
	err = service.RecordRequestDuration(ctx, "GET", "/api/test", time.Millisecond*150)
	assert.NoError(t, err)

	err = service.RecordRequestDuration(ctx, "POST", "/api/test", time.Millisecond*75)
	assert.NoError(t, err)

	// Test increment error counter
	err = service.IncrementErrorCounter(ctx, "database", "connection_timeout")
	assert.NoError(t, err)

	err = service.IncrementErrorCounter(ctx, "external_api", "rate_limit")
	assert.NoError(t, err)

	// Test record processing time
	err = service.RecordProcessingTime(ctx, "task_processing", time.Millisecond*500)
	assert.NoError(t, err)
}

func TestTelemetryService_HTTPMiddleware(t *testing.T) {
	config := createTestTelemetryConfig()
	logger := zaptest.NewLogger(t)

	service, err := NewTelemetryService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = service.Start(ctx)
	require.NoError(t, err)
	defer func() {
		if stopErr := service.Stop(ctx); stopErr != nil {
			t.Logf("Warning: failed to stop service: %v", stopErr)
		}
	}()

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write([]byte("OK")); writeErr != nil {
			t.Logf("Warning: failed to write response: %v", writeErr)
		}
	})

	// Wrap with telemetry middleware
	instrumentedHandler := service.HTTPMiddleware()(testHandler)

	// Create test request
	req := httptest.NewRequest("GET", "/test/endpoint", nil)
	w := httptest.NewRecorder()

	// Execute request
	instrumentedHandler.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Verify that tracing headers might be present
	// (This is a basic check - in a real test environment,
	// we'd verify actual trace propagation)
}

func TestTelemetryService_HealthCheck(t *testing.T) {
	config := createTestTelemetryConfig()
	logger := zaptest.NewLogger(t)

	service, err := NewTelemetryService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = service.Start(ctx)
	require.NoError(t, err)
	defer func() {
		if stopErr := service.Stop(ctx); stopErr != nil {
			t.Logf("Warning: failed to stop service: %v", stopErr)
		}
	}()

	health := service.HealthCheck()
	assert.NotNil(t, health)

	// Verify health structure (exact fields depend on implementation)
	assert.Contains(t, health, "status")
	assert.Contains(t, health, "components")
}

func TestTelemetryService_WithDisabledTelemetry(t *testing.T) {
	config := createTestTelemetryConfig()
	config.Enabled = false // Disable telemetry
	logger := zaptest.NewLogger(t)

	service, err := NewTelemetryService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = service.Start(ctx)
	assert.NoError(t, err) // Should not error even when disabled
	defer func() {
		if stopErr := service.Stop(ctx); stopErr != nil {
			t.Logf("Warning: failed to stop service: %v", stopErr)
		}
	}()

	// Service should still provide tracer/meter, but they might be no-ops
	tracer := service.GetTracer("test")
	assert.NotNil(t, tracer)

	meter := service.GetMeter("test")
	assert.NotNil(t, meter)

	// Operations should not fail
	ctx, span := tracer.Start(ctx, "test-span")
	span.End()

	counter, err := meter.Int64Counter("test_counter")
	require.NoError(t, err)
	counter.Add(ctx, 1)
}

func TestTelemetryService_ConcurrentMetrics(t *testing.T) {
	config := createTestTelemetryConfig()
	logger := zaptest.NewLogger(t)

	service, err := NewTelemetryService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = service.Start(ctx)
	require.NoError(t, err)
	defer func() {
		if stopErr := service.Stop(ctx); stopErr != nil {
			t.Logf("Warning: failed to stop service: %v", stopErr)
		}
	}()

	numGoroutines := 50
	done := make(chan bool, numGoroutines)

	// Run concurrent metric operations
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			// Record various metrics (ignoring errors in concurrent test as we're testing concurrency safety)
			_ = service.IncrementRequestCounter(ctx, "GET", "/test", "200")
			_ = service.RecordRequestDuration(ctx, "GET", "/test", time.Millisecond*100)
			_ = service.IncrementErrorCounter(ctx, "test", "concurrent")
			_ = service.RecordProcessingTime(ctx, "concurrent_task", time.Millisecond*50)
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestTelemetryService_SpanAttributes(t *testing.T) {
	config := createTestTelemetryConfig()
	logger := zaptest.NewLogger(t)

	service, err := NewTelemetryService(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = service.Start(ctx)
	require.NoError(t, err)
	defer func() {
		if stopErr := service.Stop(ctx); stopErr != nil {
			t.Logf("Warning: failed to stop service: %v", stopErr)
		}
	}()

	tracer := service.GetTracer("test")

	ctx, span := tracer.Start(ctx, "test-operation",
		trace.WithAttributes(
			attribute.String("service.name", "test-service"),
			attribute.String("operation.type", "test"),
			attribute.Int("operation.id", 12345),
		),
	)
	defer span.End()

	// Add more attributes during span execution
	span.SetAttributes(
		attribute.Bool("operation.success", true),
		attribute.Float64("operation.duration", 123.45),
		attribute.StringSlice("operation.tags", []string{"tag1", "tag2"}),
	)

	// Verify span is recording
	assert.True(t, span.IsRecording())

	// Test error recording
	span.RecordError(assert.AnError)
	span.SetStatus(codes.Error, "Test error")
}

func TestTelemetryConfig_Validation(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test with minimal valid config
	minimalConfig := TelemetryConfig{
		ServiceName: "test",
		Enabled:     true,
	}

	service, err := NewTelemetryService(minimalConfig, logger)
	assert.NoError(t, err)
	assert.NotNil(t, service)

	// Test with empty service name
	invalidConfig := TelemetryConfig{
		ServiceName: "", // Invalid empty name
		Enabled:     true,
	}

	service, err = NewTelemetryService(invalidConfig, logger)
	// Should either return error or handle gracefully
	if err == nil {
		assert.NotNil(t, service)
		// Service should have assigned a default name
		assert.NotEmpty(t, service.config.ServiceName)
	} else {
		assert.Contains(t, err.Error(), "service name")
	}
}
