package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type HealthStatus string

const (
	StatusHealthy   HealthStatus = "healthy"
	StatusDegraded  HealthStatus = "degraded"
	StatusUnhealthy HealthStatus = "unhealthy"
)

type HealthCheck struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
	LastSuccess *time.Time             `json:"last_success,omitempty"`
	LastFailure *time.Time             `json:"last_failure,omitempty"`
}

type HealthResponse struct {
	Status      HealthStatus           `json:"status"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Uptime      time.Duration          `json:"uptime"`
	Checks      map[string]HealthCheck `json:"checks"`
	System      SystemInfo             `json:"system"`
	Environment string                 `json:"environment,omitempty"`
}

type SystemInfo struct {
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"goroutines"`
	NumCPU       int    `json:"cpu_count"`
	MemStats     struct {
		Alloc        uint64 `json:"alloc_bytes"`
		TotalAlloc   uint64 `json:"total_alloc_bytes"`
		Sys          uint64 `json:"sys_bytes"`
		NumGC        uint32 `json:"gc_count"`
		LastGC       string `json:"last_gc,omitempty"`
		PauseTotalNs uint64 `json:"gc_pause_total_ns"`
	} `json:"memory"`
}

type HealthChecker interface {
	Check(ctx context.Context) HealthCheck
}

type HealthService struct {
	checkers    map[string]HealthChecker
	startTime   time.Time
	version     string
	environment string
	logger      *zap.Logger
	tracer      trace.Tracer
	mutex       sync.RWMutex
	lastResults map[string]HealthCheck
}

func NewHealthService(version, environment string, logger *zap.Logger) *HealthService {
	return &HealthService{
		checkers:    make(map[string]HealthChecker),
		startTime:   time.Now(),
		version:     version,
		environment: environment,
		logger:      logger,
		tracer:      otel.Tracer("mcp-ultra-wasm/health"),
		lastResults: make(map[string]HealthCheck),
	}
}

func (h *HealthService) RegisterChecker(name string, checker HealthChecker) {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	h.checkers[name] = checker
}

func (h *HealthService) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.HealthHandler)
	r.Get("/healthz", h.HealthzHandler)
	r.Get("/ready", h.ReadinessHandler)
	r.Get("/readyz", h.ReadinessHandler)
	r.Get("/live", h.LivenessHandler)
	r.Get("/livez", h.LivenessHandler)
	r.Get("/status", h.StatusHandler)
}

// HealthHandler provides detailed health information
func (h *HealthService) HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "health.detailed_check")
	defer span.End()

	response := h.performHealthChecks(ctx)

	// Determine HTTP status code
	var statusCode int
	switch response.Status {
	case StatusHealthy:
		statusCode = http.StatusOK
	case StatusDegraded:
		statusCode = http.StatusOK // Still operational
	case StatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	}

	span.SetAttributes(
		attribute.String("health.status", string(response.Status)),
		attribute.Int("health.checks_count", len(response.Checks)),
		attribute.Int("http.status_code", statusCode),
	)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode health response", zap.Error(err))
	}
}

// HealthzHandler provides simple health check (Kubernetes style)
func (h *HealthService) HealthzHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "health.simple_check")
	defer span.End()

	response := h.performHealthChecks(ctx)

	if response.Status == StatusUnhealthy {
		span.SetStatus(codes.Error, "service unhealthy")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		span.RecordError(err)
	}
}

// ReadinessHandler checks if service is ready to accept traffic
func (h *HealthService) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "health.readiness_check")
	defer span.End()

	// Check critical dependencies only
	critical := []string{"database", "redis", "nats"}
	isReady := true

	h.mutex.RLock()
	for _, name := range critical {
		if checker, exists := h.checkers[name]; exists {
			check := checker.Check(ctx)
			if check.Status == StatusUnhealthy {
				isReady = false
				span.SetAttributes(
					attribute.String("readiness.failed_check", name),
					attribute.String("readiness.error", check.Error),
				)
				break
			}
		}
	}
	h.mutex.RUnlock()

	if !isReady {
		span.SetStatus(codes.Error, "service not ready")
		http.Error(w, "Service Not Ready", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Ready")); err != nil {
		span.RecordError(err)
	}
}

// LivenessHandler checks if service is alive
func (h *HealthService) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "health.liveness_check")
	defer span.End()

	// Simple liveness check - if we can respond, we're alive
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("Alive")); err != nil {
		span.RecordError(err)
	}
}

// StatusHandler provides comprehensive status information
func (h *HealthService) StatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "health.status_check")
	defer span.End()

	response := h.performHealthChecks(ctx)

	// Add additional status information
	statusResponse := struct {
		HealthResponse
		RequestID string `json:"request_id,omitempty"`
		TraceID   string `json:"trace_id,omitempty"`
	}{
		HealthResponse: response,
		RequestID:      r.Header.Get("X-Request-ID"),
		TraceID:        span.SpanContext().TraceID().String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(statusResponse); err != nil {
		h.logger.Error("Failed to encode status response", zap.Error(err))
	}
}

func (h *HealthService) performHealthChecks(ctx context.Context) HealthResponse {
	start := time.Now()

	h.mutex.RLock()
	checkers := make(map[string]HealthChecker, len(h.checkers))
	for name, checker := range h.checkers {
		checkers[name] = checker
	}
	h.mutex.RUnlock()

	// Perform all health checks concurrently
	checkResults := make(map[string]HealthCheck)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for name, checker := range checkers {
		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()

			checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			result := checker.Check(checkCtx)

			mu.Lock()
			checkResults[name] = result
			h.lastResults[name] = result
			mu.Unlock()
		}(name, checker)
	}

	wg.Wait()

	// Determine overall status
	overallStatus := StatusHealthy
	for _, check := range checkResults {
		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
			break
		} else if check.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	// Get system information
	sysInfo := h.getSystemInfo()

	response := HealthResponse{
		Status:      overallStatus,
		Version:     h.version,
		Timestamp:   time.Now(),
		Uptime:      time.Since(h.startTime),
		Checks:      checkResults,
		System:      sysInfo,
		Environment: h.environment,
	}

	h.logger.Debug("Health check completed",
		zap.String("status", string(overallStatus)),
		zap.Duration("duration", time.Since(start)),
		zap.Int("checks", len(checkResults)))

	return response
}

func (h *HealthService) getSystemInfo() SystemInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	info := SystemInfo{
		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
	}

	info.MemStats.Alloc = memStats.Alloc
	info.MemStats.TotalAlloc = memStats.TotalAlloc
	info.MemStats.Sys = memStats.Sys
	info.MemStats.NumGC = memStats.NumGC
	info.MemStats.PauseTotalNs = memStats.PauseTotalNs

	if memStats.LastGC != 0 {
		lastGC := time.Unix(0, int64(memStats.LastGC))
		info.MemStats.LastGC = lastGC.Format(time.RFC3339)
	}

	return info
}

// Built-in health checkers

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	name string
	ping func(context.Context) error
}

func NewDatabaseHealthChecker(name string, pingFunc func(context.Context) error) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name: name,
		ping: pingFunc,
	}
}

func (d *DatabaseHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name:      d.name,
		Timestamp: start,
	}

	if err := d.ping(ctx); err != nil {
		check.Status = StatusUnhealthy
		check.Error = err.Error()
		check.Message = fmt.Sprintf("Database %s is unreachable", d.name)
	} else {
		check.Status = StatusHealthy
		check.Message = fmt.Sprintf("Database %s is healthy", d.name)
	}

	check.Duration = time.Since(start)
	return check
}

// RedisHealthChecker checks Redis connectivity
type RedisHealthChecker struct {
	ping func(context.Context) error
}

func NewRedisHealthChecker(pingFunc func(context.Context) error) *RedisHealthChecker {
	return &RedisHealthChecker{
		ping: pingFunc,
	}
}

func (r *RedisHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name:      "redis",
		Timestamp: start,
	}

	if err := r.ping(ctx); err != nil {
		check.Status = StatusUnhealthy
		check.Error = err.Error()
		check.Message = "Redis is unreachable"
	} else {
		check.Status = StatusHealthy
		check.Message = "Redis is healthy"
	}

	check.Duration = time.Since(start)
	return check
}

// NATSHealthChecker checks NATS connectivity
type NATSHealthChecker struct {
	isConnected func() bool
}

func NewNATSHealthChecker(isConnectedFunc func() bool) *NATSHealthChecker {
	return &NATSHealthChecker{
		isConnected: isConnectedFunc,
	}
}

func (n *NATSHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()
	check := HealthCheck{
		Name:      "nats",
		Timestamp: start,
	}

	if !n.isConnected() {
		check.Status = StatusUnhealthy
		check.Message = "NATS connection is down"
		check.Error = "connection lost"
	} else {
		check.Status = StatusHealthy
		check.Message = "NATS is connected"
	}

	check.Duration = time.Since(start)
	return check
}
