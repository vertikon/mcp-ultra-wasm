package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap/zaptest"
)

func TestNewTracingProvider(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("should create provider with stdout exporter", func(t *testing.T) {
		config := &TracingConfig{
			Enabled:        true,
			ServiceName:    "test-service",
			ServiceVersion: "v1.0.0",
			Environment:    "test",
			Exporter:       "stdout",
			SampleRate:     1.0,
			BatchTimeout:   time.Second,
		}

		provider, err := NewTracingProvider(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.NotNil(t, provider.provider)

		// Clean up
		err = provider.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("should create provider with noop exporter", func(t *testing.T) {
		config := &TracingConfig{
			Enabled:        true,
			ServiceName:    "test-service",
			ServiceVersion: "v1.0.0",
			Environment:    "test",
			Exporter:       "noop",
			SampleRate:     0.5,
		}

		provider, err := NewTracingProvider(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("should handle disabled tracing", func(t *testing.T) {
		config := &TracingConfig{
			Enabled: false,
		}

		provider, err := NewTracingProvider(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Nil(t, provider.provider)

		// Shutdown should work even when disabled
		err = provider.Shutdown(context.Background())
		assert.NoError(t, err)
	})

	t.Run("should include custom resource attributes", func(t *testing.T) {
		config := &TracingConfig{
			Enabled:        true,
			ServiceName:    "test-service",
			ServiceVersion: "v1.0.0",
			Environment:    "test",
			Exporter:       "noop",
			ResourceAttributes: map[string]string{
				"team":       "platform",
				"datacenter": "us-west-1",
			},
		}

		provider, err := NewTracingProvider(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})
}

func TestTracingProvider_GetTracer(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("should return tracer when enabled", func(t *testing.T) {
		config := &TracingConfig{
			Enabled:     true,
			ServiceName: "test-service",
			Exporter:    "noop",
		}

		provider, err := NewTracingProvider(config, logger)
		require.NoError(t, err)

		tracer := provider.GetTracer("test-tracer")
		assert.NotNil(t, tracer)
	})

	t.Run("should return noop tracer when disabled", func(t *testing.T) {
		config := &TracingConfig{
			Enabled: false,
		}

		provider, err := NewTracingProvider(config, logger)
		require.NoError(t, err)

		tracer := provider.GetTracer("test-tracer")
		assert.NotNil(t, tracer)
		// Noop tracer should still work but not record spans
	})
}

func TestTraceFunction(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TracingConfig{
		Enabled:     true,
		ServiceName: "test-service",
		Exporter:    "noop",
		SampleRate:  1.0,
	}

	provider, err := NewTracingProvider(config, logger)
	require.NoError(t, err)

	tracer := provider.GetTracer("test")

	t.Run("should execute function successfully", func(t *testing.T) {
		executed := false
		err := TraceFunction(context.Background(), tracer, "test-operation", func(_ context.Context) error {
			executed = true
			return nil
		})

		assert.NoError(t, err)
		assert.True(t, executed)
	})

	t.Run("should handle function error", func(t *testing.T) {
		expectedError := assert.AnError
		err := TraceFunction(context.Background(), tracer, "failing-operation", func(_ context.Context) error {
			return expectedError
		})

		assert.Equal(t, expectedError, err)
	})
}

func TestTraceFunctionWithResult(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TracingConfig{
		Enabled:     true,
		ServiceName: "test-service",
		Exporter:    "noop",
		SampleRate:  1.0,
	}

	provider, err := NewTracingProvider(config, logger)
	require.NoError(t, err)

	tracer := provider.GetTracer("test")

	t.Run("should return result successfully", func(t *testing.T) {
		expectedResult := "test-result"
		result, err := TraceFunctionWithResult(context.Background(), tracer, "test-operation", func(_ context.Context) (string, error) {
			return expectedResult, nil
		})

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
	})

	t.Run("should handle function error and return zero value", func(t *testing.T) {
		expectedError := assert.AnError
		result, err := TraceFunctionWithResult(context.Background(), tracer, "failing-operation", func(_ context.Context) (string, error) {
			return "should-not-return", expectedError
		})

		assert.Equal(t, expectedError, err)
		assert.Equal(t, "", result) // Zero value for string
	})
}

func TestSpanUtilities(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TracingConfig{
		Enabled:     true,
		ServiceName: "test-service",
		Exporter:    "noop",
		SampleRate:  1.0,
	}

	provider, err := NewTracingProvider(config, logger)
	require.NoError(t, err)

	tracer := provider.GetTracer("test")

	t.Run("should add span attributes", func(_ *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		// This should not panic
		AddSpanAttributes(ctx,
			attribute.String("key1", "value1"),
			attribute.Int("key2", 42),
		)

		// Test with context without span (should not panic)
		AddSpanAttributes(context.Background(),
			attribute.String("key", "value"),
		)
	})

	t.Run("should add span events", func(_ *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		// This should not panic
		AddSpanEvent(ctx, "test-event",
			attribute.String("event.type", "test"),
		)

		// Test with context without span (should not panic)
		AddSpanEvent(context.Background(), "test-event")
	})

	t.Run("should set span error", func(_ *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		// This should not panic
		SetSpanError(ctx, assert.AnError)

		// Test with context without span (should not panic)
		SetSpanError(context.Background(), assert.AnError)
	})

	t.Run("should get trace and span IDs", func(t *testing.T) {
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		traceID := GetTraceID(ctx)
		spanID := GetSpanID(ctx)

		// IDs should be non-empty when span is active
		if span.SpanContext().HasTraceID() {
			assert.NotEmpty(t, traceID)
		}
		if span.SpanContext().HasSpanID() {
			assert.NotEmpty(t, spanID)
		}

		// Should return empty strings for context without span
		emptyTraceID := GetTraceID(context.Background())
		emptySpanID := GetSpanID(context.Background())
		assert.Empty(t, emptyTraceID)
		assert.Empty(t, emptySpanID)
	})
}

func TestTraceContextPropagation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TracingConfig{
		Enabled:     true,
		ServiceName: "test-service",
		Exporter:    "noop",
		SampleRate:  1.0,
	}

	provider, err := NewTracingProvider(config, logger)
	require.NoError(t, err)

	tracer := provider.GetTracer("test")

	t.Run("should inject and extract trace context", func(t *testing.T) {
		// Create a span with trace context
		ctx, span := tracer.Start(context.Background(), "parent-span")
		defer span.End()

		// Inject context into a map (simulating HTTP headers)
		carrier := make(map[string]string)
		InjectTraceContext(ctx, carrier)

		// Carrier should now contain trace context
		assert.NotEmpty(t, carrier)

		// Extract context from the map
		extractedCtx := ExtractTraceContext(context.Background(), carrier)
		assert.NotNil(t, extractedCtx)

		// The extracted context should have the same trace ID
		originalTraceID := GetTraceID(ctx)
		extractedTraceID := GetTraceID(extractedCtx)

		if originalTraceID != "" {
			assert.Equal(t, originalTraceID, extractedTraceID)
		}
	})

	t.Run("should handle empty carrier", func(t *testing.T) {
		emptyCarrier := make(map[string]string)
		extractedCtx := ExtractTraceContext(context.Background(), emptyCarrier)
		assert.NotNil(t, extractedCtx)

		// Should not have trace context
		traceID := GetTraceID(extractedCtx)
		assert.Empty(t, traceID)
	})
}

func TestMapCarrier(t *testing.T) {
	t.Run("should implement TextMapCarrier interface", func(t *testing.T) {
		carrier := &mapCarrier{m: make(map[string]string)}

		// Test Set and Get
		carrier.Set("test-key", "test-value")
		value := carrier.Get("test-key")
		assert.Equal(t, "test-value", value)

		// Test Get non-existent key
		emptyValue := carrier.Get("non-existent")
		assert.Empty(t, emptyValue)

		// Test Keys
		carrier.Set("key1", "value1")
		carrier.Set("key2", "value2")
		keys := carrier.Keys()
		assert.Contains(t, keys, "test-key")
		assert.Contains(t, keys, "key1")
		assert.Contains(t, keys, "key2")
		assert.Len(t, keys, 3)
	})
}

func TestNoopExporter(t *testing.T) {
	exporter := &noopExporter{}

	t.Run("should not return error on ExportSpans", func(t *testing.T) {
		err := exporter.ExportSpans(context.Background(), nil)
		assert.NoError(t, err)
	})

	t.Run("should not return error on Shutdown", func(t *testing.T) {
		err := exporter.Shutdown(context.Background())
		assert.NoError(t, err)
	})
}

func TestCreateSpanExporter(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name         string
		exporterType string
		shouldWork   bool
	}{
		{"stdout", "stdout", true},
		{"noop", "noop", true},
		{"unknown", "unknown", true}, // Falls back to stdout
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &TracingConfig{
				Exporter: tt.exporterType,
			}

			exporter, err := createSpanExporter(config, logger)

			if tt.shouldWork {
				assert.NoError(t, err)
				assert.NotNil(t, exporter)

				// Test that exporter can be shut down
				err = exporter.Shutdown(context.Background())
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
