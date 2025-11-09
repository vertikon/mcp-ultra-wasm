package http

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestHealthService_RegisterRoutes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewHealthService("v1.0.0", "test", logger)

	r := chi.NewRouter()
	service.RegisterRoutes(r)

	// Test that all routes are registered
	routes := []string{"/health", "/healthz", "/ready", "/readyz", "/live", "/livez", "/status"}
	for _, route := range routes {
		req := httptest.NewRequest("GET", route, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// Should not return 404 (route exists)
		assert.NotEqual(t, http.StatusNotFound, w.Code, "Route %s should be registered", route)
	}
}

func TestHealthService_HealthHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewHealthService("v1.0.0", "test", logger)

	t.Run("should return healthy status with no checkers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		service.HealthHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
	})

	t.Run("should return unhealthy status with failing checker", func(t *testing.T) {
		// Register a failing health checker
		service.RegisterChecker("failing", &mockHealthChecker{
			name:   "failing",
			status: StatusUnhealthy,
			error:  "service unavailable",
		})

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		service.HealthHandler(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})

	t.Run("should return degraded status with degraded checker", func(t *testing.T) {
		service = NewHealthService("v1.0.0", "test", logger) // Fresh service
		service.RegisterChecker("degraded", &mockHealthChecker{
			name:   "degraded",
			status: StatusDegraded,
		})

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		service.HealthHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code) // Degraded still returns 200
	})
}

func TestHealthService_HealthzHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewHealthService("v1.0.0", "test", logger)

	t.Run("should return OK for healthy service", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()

		service.HealthzHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		assert.Equal(t, "OK", w.Body.String())
	})

	t.Run("should return 503 for unhealthy service", func(t *testing.T) {
		service.RegisterChecker("unhealthy", &mockHealthChecker{
			name:   "unhealthy",
			status: StatusUnhealthy,
		})

		req := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()

		service.HealthzHandler(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestHealthService_ReadinessHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewHealthService("v1.0.0", "test", logger)

	t.Run("should return Ready for ready service", func(t *testing.T) {
		// Register healthy critical services
		service.RegisterChecker("database", &mockHealthChecker{
			name:   "database",
			status: StatusHealthy,
		})
		service.RegisterChecker("redis", &mockHealthChecker{
			name:   "redis",
			status: StatusHealthy,
		})

		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		service.ReadinessHandler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "Ready", w.Body.String())
	})

	t.Run("should return 503 for not ready service", func(t *testing.T) {
		service = NewHealthService("v1.0.0", "test", logger) // Fresh service
		service.RegisterChecker("database", &mockHealthChecker{
			name:   "database",
			status: StatusUnhealthy,
			error:  "connection failed",
		})

		req := httptest.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		service.ReadinessHandler(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestHealthService_LivenessHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewHealthService("v1.0.0", "test", logger)

	req := httptest.NewRequest("GET", "/live", nil)
	w := httptest.NewRecorder()

	service.LivenessHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	assert.Equal(t, "Alive", w.Body.String())
}

func TestHealthService_StatusHandler(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewHealthService("v1.0.0", "test", logger)

	req := httptest.NewRequest("GET", "/status", nil)
	req.Header.Set("X-Request-ID", "test-request-123")
	w := httptest.NewRecorder()

	service.StatusHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
}

func TestDatabaseHealthChecker(t *testing.T) {
	t.Run("should return healthy when ping succeeds", func(t *testing.T) {
		checker := NewDatabaseHealthChecker("test-db", func(ctx context.Context) error {
			return nil // Ping succeeds
		})

		result := checker.Check(context.Background())

		assert.Equal(t, "test-db", result.Name)
		assert.Equal(t, StatusHealthy, result.Status)
		assert.Contains(t, result.Message, "healthy")
		assert.Empty(t, result.Error)
	})

	t.Run("should return unhealthy when ping fails", func(t *testing.T) {
		checker := NewDatabaseHealthChecker("test-db", func(ctx context.Context) error {
			return assert.AnError // Ping fails
		})

		result := checker.Check(context.Background())

		assert.Equal(t, "test-db", result.Name)
		assert.Equal(t, StatusUnhealthy, result.Status)
		assert.Contains(t, result.Message, "unreachable")
		assert.NotEmpty(t, result.Error)
	})
}

func TestRedisHealthChecker(t *testing.T) {
	t.Run("should return healthy when ping succeeds", func(t *testing.T) {
		checker := NewRedisHealthChecker(func(ctx context.Context) error {
			return nil // Ping succeeds
		})

		result := checker.Check(context.Background())

		assert.Equal(t, "redis", result.Name)
		assert.Equal(t, StatusHealthy, result.Status)
		assert.Contains(t, result.Message, "healthy")
		assert.Empty(t, result.Error)
	})

	t.Run("should return unhealthy when ping fails", func(t *testing.T) {
		checker := NewRedisHealthChecker(func(ctx context.Context) error {
			return assert.AnError // Ping fails
		})

		result := checker.Check(context.Background())

		assert.Equal(t, "redis", result.Name)
		assert.Equal(t, StatusUnhealthy, result.Status)
		assert.Contains(t, result.Message, "unreachable")
		assert.NotEmpty(t, result.Error)
	})
}

func TestNATSHealthChecker(t *testing.T) {
	t.Run("should return healthy when connected", func(t *testing.T) {
		checker := NewNATSHealthChecker(func() bool {
			return true // Connected
		})

		result := checker.Check(context.Background())

		assert.Equal(t, "nats", result.Name)
		assert.Equal(t, StatusHealthy, result.Status)
		assert.Contains(t, result.Message, "connected")
		assert.Empty(t, result.Error)
	})

	t.Run("should return unhealthy when disconnected", func(t *testing.T) {
		checker := NewNATSHealthChecker(func() bool {
			return false // Disconnected
		})

		result := checker.Check(context.Background())

		assert.Equal(t, "nats", result.Name)
		assert.Equal(t, StatusUnhealthy, result.Status)
		assert.Contains(t, result.Message, "down")
		assert.NotEmpty(t, result.Error)
	})
}

func TestHealthService_ConcurrentChecks(t *testing.T) {
	logger := zaptest.NewLogger(t)
	service := NewHealthService("v1.0.0", "test", logger)

	// Register multiple slow checkers
	for i := 0; i < 5; i++ {
		service.RegisterChecker(fmt.Sprintf("slow-%d", i), &mockHealthChecker{
			name:     fmt.Sprintf("slow-%d", i),
			status:   StatusHealthy,
			duration: 100 * time.Millisecond, // Simulate slow check
		})
	}

	start := time.Now()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	service.HealthHandler(w, req)

	duration := time.Since(start)

	// Should complete much faster than 5 * 100ms due to concurrent execution
	assert.Less(t, duration, 300*time.Millisecond, "Concurrent checks should be faster")
	assert.Equal(t, http.StatusOK, w.Code)
}

// Mock health checker for testing
type mockHealthChecker struct {
	name     string
	status   HealthStatus
	error    string
	duration time.Duration
}

func (m *mockHealthChecker) Check(ctx context.Context) HealthCheck {
	if m.duration > 0 {
		time.Sleep(m.duration)
	}

	start := time.Now()
	return HealthCheck{
		Name:      m.name,
		Status:    m.status,
		Error:     m.error,
		Timestamp: start,
		Duration:  time.Since(start),
	}
}
