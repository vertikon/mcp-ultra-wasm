//go:build integration
// +build integration

package observability_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability"
)

// TestObservabilityIntegration tests the complete observability stack
func TestObservabilityIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping observability integration test in short mode")
	}

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Setup telemetry configuration for testing
	cfg := config.TelemetryConfig{
		Enabled:        true,
		ServiceName:    "mcp-ultra-wasm-test",
		ServiceVersion: "test-1.0.0",
		Environment:    "test",
		Debug:          true,
		Tracing: config.TracingConfig{
			Enabled:    true,
			SampleRate: 1.0, // Sample everything in test
			MaxSpans:   100,
			BatchSize:  10,
			Timeout:    time.Second * 2,
		},
		Metrics: config.MetricsConfig{
			Enabled:         true,
			Port:            9091, // Different port to avoid conflicts
			Path:            "/metrics",
			CollectInterval: time.Second * 5,
		},
		Exporters: config.ExportersConfig{
			OTLP: config.OTLPConfig{
				Enabled:  true,
				Endpoint: getOTLPEndpoint(),
				Insecure: true,
			},
			Console: config.ConsoleConfig{
				Enabled: false, // Disable console output in tests
			},
		},
	}

	// Create observability service
	obsService, err := observability.NewService(cfg, logger)
	require.NoError(t, err, "Failed to create observability service")

	ctx := context.Background()

	// Start the observability service
	err = obsService.Start(ctx)
	require.NoError(t, err, "Failed to start observability service")

	// Ensure cleanup
	defer func() {
		if err := obsService.Stop(ctx); err != nil {
			t.Logf("Failed to stop observability service: %v", err)
		}
	}()

	t.Run("HealthCheck", func(t *testing.T) {
		health := obsService.HealthCheck()
		assert.NotNil(t, health)

		observabilityHealth, exists := health["observability"]
		require.True(t, exists)

		healthMap := observabilityHealth.(map[string]interface{})
		assert.True(t, healthMap["enabled"].(bool))
		assert.Equal(t, "healthy", healthMap["status"])
	})

	t.Run("TracingIntegration", func(t *testing.T) {
		telemetryService := obsService.GetTelemetryService()
		require.NotNil(t, telemetryService)

		// Create a test span
		ctx, span := telemetryService.StartSpan(ctx, "test-operation")
		span.SetAttributes(
			telemetryService.CreateAttribute("test.key", "test-value"),
			telemetryService.CreateAttribute("test.number", 42),
		)

		// Simulate some work
		time.Sleep(10 * time.Millisecond)

		// Add an event to the span
		telemetryService.AddSpanEvent(ctx, "test-event",
			telemetryService.CreateAttribute("event.type", "test"))

		span.End()

		// Give time for span to be exported
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("MetricsIntegration", func(t *testing.T) {
		telemetryService := obsService.GetTelemetryService()
		require.NotNil(t, telemetryService)

		// Record some test metrics
		telemetryService.RecordHTTPRequest("GET", "/test", "200", 50*time.Millisecond)
		telemetryService.RecordHTTPRequest("POST", "/test", "201", 100*time.Millisecond)
		telemetryService.RecordHTTPRequest("GET", "/test", "404", 10*time.Millisecond)

		telemetryService.RecordError("test_error", "test")
		telemetryService.IncrementCounter("test_counter", "test-service")

		// Give metrics time to be collected
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("TaskOperationInstrumentation", func(t *testing.T) {
		err := obsService.RecordTaskOperation(ctx, "create", "test-task-id", func(ctx context.Context) error {
			// Simulate task operation
			time.Sleep(20 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)

		// Test error case
		err = obsService.RecordTaskOperation(ctx, "delete", "error-task-id", func(ctx context.Context) error {
			return fmt.Errorf("simulated task error")
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "simulated task error")
	})

	t.Run("DatabaseOperationInstrumentation", func(t *testing.T) {
		err := obsService.RecordDatabaseOperation(ctx, "select", "tasks", "SELECT * FROM tasks WHERE id = $1", func(ctx context.Context) error {
			time.Sleep(5 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("CacheOperationInstrumentation", func(t *testing.T) {
		err := obsService.RecordCacheOperation(ctx, "get", "test-key", func(ctx context.Context) error {
			time.Sleep(2 * time.Millisecond)
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("HTTPMiddlewareIntegration", func(t *testing.T) {
		// Create a test HTTP handler
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(10 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
				// Handle encoding error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		})

		// Apply observability middleware
		instrumentedHandler := obsService.HTTPMiddleware()(testHandler)

		// Create test server
		server := &http.Server{
			Addr:    ":0", // Use any available port
			Handler: instrumentedHandler,
		}

		// Start server in goroutine
		go func() {
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				t.Logf("Test server error: %v", err)
			}
		}()

		// Give server time to start
		time.Sleep(50 * time.Millisecond)

		// Cleanup
		defer func() {
			if err := server.Shutdown(context.Background()); err != nil {
				t.Logf("Failed to shutdown test server: %v", err)
			}
		}()
	})

	t.Run("LogWithTraceContext", func(t *testing.T) {
		// Create a traced context
		telemetryService := obsService.GetTelemetryService()
		ctx, span := telemetryService.StartSpan(ctx, "log-test-operation")
		defer span.End()

		// Log with trace context
		obsService.LogWithTrace(ctx, "info", "Test log with trace context",
			zap.String("test.field", "test-value"),
			zap.Int("test.number", 123),
		)
	})

	// Wait a bit for all telemetry data to be exported
	time.Sleep(500 * time.Millisecond)
}

// TestObservabilityDisabled tests that the system works correctly when observability is disabled
func TestObservabilityDisabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := config.TelemetryConfig{
		Enabled: false,
	}

	obsService, err := observability.NewService(cfg, logger)
	require.NoError(t, err)

	ctx := context.Background()
	err = obsService.Start(ctx)
	require.NoError(t, err)

	defer obsService.Stop(ctx)

	// Test that operations work without errors when observability is disabled
	t.Run("DisabledOperations", func(t *testing.T) {
		err := obsService.RecordTaskOperation(ctx, "test", "test-id", func(ctx context.Context) error {
			return nil
		})
		assert.NoError(t, err)

		err = obsService.RecordDatabaseOperation(ctx, "select", "table", "query", func(ctx context.Context) error {
			return nil
		})
		assert.NoError(t, err)

		// HTTP middleware should pass through without instrumentation
		handler := obsService.HTTPMiddleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		assert.NotNil(t, handler)

		// Health check should indicate disabled state
		health := obsService.HealthCheck()
		observabilityHealth := health["observability"].(map[string]interface{})
		assert.False(t, observabilityHealth["enabled"].(bool))
	})
}

// TestOTLPConnectivity tests connection to OTLP endpoint
func TestOTLPConnectivity(t *testing.T) {
	endpoint := getOTLPEndpoint()
	if endpoint == "" {
		t.Skip("No OTLP endpoint configured for testing")
	}

	// Simple connectivity test
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(endpoint + "/v1/traces")
	if err != nil {
		t.Logf("OTLP endpoint not reachable: %v", err)
		return
	}
	defer resp.Body.Close()

	// We expect either 200 OK or 404 Not Found (method not allowed is also acceptable)
	assert.True(t, resp.StatusCode == 200 || resp.StatusCode == 404 || resp.StatusCode == 405)
}

// BenchmarkObservabilityOverhead benchmarks the observability overhead
func BenchmarkObservabilityOverhead(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	cfg := config.TelemetryConfig{
		Enabled:        true,
		ServiceName:    "benchmark-test",
		ServiceVersion: "1.0.0",
		Environment:    "test",
		Debug:          false,
		Tracing: config.TracingConfig{
			Enabled:    true,
			SampleRate: 0.1, // Low sampling for benchmark
		},
		Exporters: config.ExportersConfig{
			Console: config.ConsoleConfig{Enabled: false},
		},
	}

	obsService, err := observability.NewService(cfg, logger)
	require.NoError(b, err)

	ctx := context.Background()
	err = obsService.Start(ctx)
	require.NoError(b, err)
	defer obsService.Stop(ctx)

	b.ResetTimer()

	b.Run("TaskOperation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			obsService.RecordTaskOperation(ctx, "benchmark", fmt.Sprintf("task-%d", i), func(ctx context.Context) error {
				return nil
			})
		}
	})

	b.Run("LogWithTrace", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			obsService.LogWithTrace(ctx, "info", "Benchmark log message",
				zap.Int("iteration", i))
		}
	})
}

// Helper function to get OTLP endpoint from environment
func getOTLPEndpoint() string {
	endpoint := os.Getenv("OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:4318" // Default HTTP endpoint
	}
	return endpoint
}
