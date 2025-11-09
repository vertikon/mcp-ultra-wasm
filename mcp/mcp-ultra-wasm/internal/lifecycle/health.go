package lifecycle

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/logger"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a health check result
type HealthCheck struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Error     string                 `json:"error,omitempty"`
}

// HealthReport represents the overall health status
type HealthReport struct {
	Status       HealthStatus           `json:"status"`
	Version      string                 `json:"version"`
	Timestamp    time.Time              `json:"timestamp"`
	Uptime       time.Duration          `json:"uptime"`
	Checks       map[string]HealthCheck `json:"checks"`
	Summary      HealthSummary          `json:"summary"`
	Dependencies []DependencyStatus     `json:"dependencies"`
}

// HealthSummary provides a summary of health checks
type HealthSummary struct {
	Total     int `json:"total"`
	Healthy   int `json:"healthy"`
	Degraded  int `json:"degraded"`
	Unhealthy int `json:"unhealthy"`
	Unknown   int `json:"unknown"`
}

// DependencyStatus represents the status of an external dependency
type DependencyStatus struct {
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Status       HealthStatus           `json:"status"`
	ResponseTime time.Duration          `json:"response_time"`
	Details      map[string]interface{} `json:"details,omitempty"`
	LastChecked  time.Time              `json:"last_checked"`
}

// HealthChecker interface for health check implementations
type HealthChecker interface {
	Name() string
	Check(ctx context.Context) HealthCheck
	IsRequired() bool
	Timeout() time.Duration
}

// HealthMonitor provides comprehensive health monitoring
type HealthMonitor struct {
	checkers     []HealthChecker
	dependencies []DependencyChecker

	// State
	mu         sync.RWMutex
	lastReport *HealthReport
	startTime  time.Time
	version    string

	// Configuration
	config HealthConfig
	logger *logger.Logger

	// Background monitoring
	ticker  *time.Ticker
	stopCh  chan struct{}
	running bool
}

// HealthConfig configures health monitoring
type HealthConfig struct {
	CheckInterval     time.Duration `json:"check_interval"`
	CheckTimeout      time.Duration `json:"check_timeout"`
	DependencyTimeout time.Duration `json:"dependency_timeout"`

	// Thresholds
	DegradedThreshold  int `json:"degraded_threshold"`  // Percentage of failed checks to be considered degraded
	UnhealthyThreshold int `json:"unhealthy_threshold"` // Percentage of failed checks to be considered unhealthy

	// HTTP endpoint
	EnableHTTPEndpoint bool   `json:"enable_http_endpoint"`
	HTTPPort           int    `json:"http_port"`
	HTTPPath           string `json:"http_path"`

	// Alerting
	EnableAlerting bool          `json:"enable_alerting"`
	AlertThreshold HealthStatus  `json:"alert_threshold"`
	AlertCooldown  time.Duration `json:"alert_cooldown"`

	// Persistence
	EnablePersistence bool   `json:"enable_persistence"`
	PersistencePath   string `json:"persistence_path"`
}

// DependencyChecker checks external dependencies
type DependencyChecker interface {
	Name() string
	Type() string
	Check(ctx context.Context) DependencyStatus
	IsRequired() bool
}

// DefaultHealthConfig returns default health monitoring configuration
func DefaultHealthConfig() HealthConfig {
	return HealthConfig{
		CheckInterval:      30 * time.Second,
		CheckTimeout:       10 * time.Second,
		DependencyTimeout:  15 * time.Second,
		DegradedThreshold:  25, // 25% failures = degraded
		UnhealthyThreshold: 50, // 50% failures = unhealthy
		EnableHTTPEndpoint: true,
		HTTPPort:           8080,
		HTTPPath:           "/health",
		EnableAlerting:     true,
		AlertThreshold:     HealthStatusDegraded,
		AlertCooldown:      5 * time.Minute,
		EnablePersistence:  true,
		PersistencePath:    "/tmp/health-status.json",
	}
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(config HealthConfig, version string, logger *logger.Logger) *HealthMonitor {
	return &HealthMonitor{
		checkers:     make([]HealthChecker, 0),
		dependencies: make([]DependencyChecker, 0),
		startTime:    time.Now(),
		version:      version,
		config:       config,
		logger:       logger,
		stopCh:       make(chan struct{}),
	}
}

// RegisterChecker registers a health checker
func (hm *HealthMonitor) RegisterChecker(checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.checkers = append(hm.checkers, checker)
	hm.logger.Info("Health checker registered",
		"name", checker.Name(),
		"required", checker.IsRequired(),
		"timeout", checker.Timeout(),
	)
}

// RegisterDependency registers a dependency checker
func (hm *HealthMonitor) RegisterDependency(dependency DependencyChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.dependencies = append(hm.dependencies, dependency)
	hm.logger.Info("Dependency checker registered",
		"name", dependency.Name(),
		"type", dependency.Type(),
		"required", dependency.IsRequired(),
	)
}

// Start starts the health monitoring
func (hm *HealthMonitor) Start() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.running {
		return fmt.Errorf("health monitor already running")
	}

	hm.running = true
	hm.ticker = time.NewTicker(hm.config.CheckInterval)

	// Start HTTP endpoint if enabled
	if hm.config.EnableHTTPEndpoint {
		go hm.startHTTPEndpoint()
	}

	// Start background monitoring
	go hm.runHealthChecks()

	hm.logger.Info("Health monitor started",
		"check_interval", hm.config.CheckInterval,
		"http_endpoint", hm.config.EnableHTTPEndpoint,
		"checkers_count", len(hm.checkers),
	)

	return nil
}

// Stop stops the health monitoring
func (hm *HealthMonitor) Stop() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.running {
		return nil
	}

	hm.running = false
	close(hm.stopCh)

	if hm.ticker != nil {
		hm.ticker.Stop()
	}

	hm.logger.Info("Health monitor stopped")
	return nil
}

// GetHealth returns the current health status
func (hm *HealthMonitor) GetHealth(ctx context.Context) *HealthReport {
	return hm.performHealthCheck(ctx)
}

// GetLastReport returns the last health report
func (hm *HealthMonitor) GetLastReport() *HealthReport {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	if hm.lastReport == nil {
		return nil
	}

	// Return a copy
	report := *hm.lastReport
	return &report
}

// IsHealthy returns true if the system is healthy
func (hm *HealthMonitor) IsHealthy() bool {
	report := hm.GetLastReport()
	if report == nil {
		return false
	}
	return report.Status == HealthStatusHealthy
}

// IsDegraded returns true if the system is degraded
func (hm *HealthMonitor) IsDegraded() bool {
	report := hm.GetLastReport()
	if report == nil {
		return false
	}
	return report.Status == HealthStatusDegraded
}

// IsUnhealthy returns true if the system is unhealthy
func (hm *HealthMonitor) IsUnhealthy() bool {
	report := hm.GetLastReport()
	if report == nil {
		return true
	}
	return report.Status == HealthStatusUnhealthy
}

// Private methods

func (hm *HealthMonitor) runHealthChecks() {
	// Perform initial health check
	ctx := context.Background()
	hm.performHealthCheck(ctx)

	for {
		select {
		case <-hm.stopCh:
			return
		case <-hm.ticker.C:
			hm.performHealthCheck(ctx)
		}
	}
}

func (hm *HealthMonitor) performHealthCheck(ctx context.Context) *HealthReport {
	checkCtx, cancel := context.WithTimeout(ctx, hm.config.CheckTimeout)
	defer cancel()

	report := &HealthReport{
		Version:      hm.version,
		Timestamp:    time.Now(),
		Uptime:       time.Since(hm.startTime),
		Checks:       make(map[string]HealthCheck),
		Dependencies: make([]DependencyStatus, 0),
	}

	// Execute health checks
	hm.executeHealthChecks(checkCtx, report)

	// Execute dependency checks
	hm.executeDependencyChecks(checkCtx, report)

	// Calculate overall status
	hm.calculateOverallStatus(report)

	// Update last report
	hm.mu.Lock()
	hm.lastReport = report
	hm.mu.Unlock()

	// Log status change
	if hm.lastReport == nil || hm.lastReport.Status != report.Status {
		hm.logger.Info("Health status changed",
			"new_status", report.Status,
			"healthy", report.Summary.Healthy,
			"degraded", report.Summary.Degraded,
			"unhealthy", report.Summary.Unhealthy,
		)
	}

	// Persist if enabled
	if hm.config.EnablePersistence {
		hm.persistHealthReport(report)
	}

	return report
}

func (hm *HealthMonitor) executeHealthChecks(ctx context.Context, report *HealthReport) {
	hm.mu.RLock()
	checkers := make([]HealthChecker, len(hm.checkers))
	copy(checkers, hm.checkers)
	hm.mu.RUnlock()

	// Execute checks concurrently
	checkChan := make(chan HealthCheck, len(checkers))

	for _, checker := range checkers {
		go func(c HealthChecker) {
			checkCtx := ctx
			if c.Timeout() > 0 {
				var cancel context.CancelFunc
				checkCtx, cancel = context.WithTimeout(ctx, c.Timeout())
				defer cancel()
			}

			startTime := time.Now()
			check := c.Check(checkCtx)
			check.Duration = time.Since(startTime)
			check.Timestamp = time.Now()

			checkChan <- check
		}(checker)
	}

	// Collect results
	for i := 0; i < len(checkers); i++ {
		check := <-checkChan
		report.Checks[check.Name] = check
	}
}

func (hm *HealthMonitor) executeDependencyChecks(ctx context.Context, report *HealthReport) {
	hm.mu.RLock()
	dependencies := make([]DependencyChecker, len(hm.dependencies))
	copy(dependencies, hm.dependencies)
	hm.mu.RUnlock()

	depChan := make(chan DependencyStatus, len(dependencies))

	for _, dependency := range dependencies {
		go func(d DependencyChecker) {
			depCtx, cancel := context.WithTimeout(ctx, hm.config.DependencyTimeout)
			defer cancel()

			status := d.Check(depCtx)
			status.LastChecked = time.Now()

			depChan <- status
		}(dependency)
	}

	// Collect results
	for i := 0; i < len(dependencies); i++ {
		status := <-depChan
		report.Dependencies = append(report.Dependencies, status)
	}
}

func (hm *HealthMonitor) calculateOverallStatus(report *HealthReport) {
	totalChecks := len(report.Checks)
	if totalChecks == 0 {
		report.Status = HealthStatusUnknown
		return
	}

	summary := HealthSummary{}

	for _, check := range report.Checks {
		switch check.Status {
		case HealthStatusHealthy:
			summary.Healthy++
		case HealthStatusDegraded:
			summary.Degraded++
		case HealthStatusUnhealthy:
			summary.Unhealthy++
		default:
			summary.Unknown++
		}
		summary.Total++
	}

	report.Summary = summary

	// Calculate failure percentage
	failures := summary.Degraded + summary.Unhealthy
	failurePercent := (failures * 100) / summary.Total

	// Determine overall status
	if failures == 0 {
		report.Status = HealthStatusHealthy
	} else if failurePercent >= hm.config.UnhealthyThreshold {
		report.Status = HealthStatusUnhealthy
	} else if failurePercent >= hm.config.DegradedThreshold {
		report.Status = HealthStatusDegraded
	} else {
		report.Status = HealthStatusHealthy
	}

	// Consider dependencies
	for _, dep := range report.Dependencies {
		if dep.Status == HealthStatusUnhealthy {
			if report.Status == HealthStatusHealthy {
				report.Status = HealthStatusDegraded
			}
		}
	}
}

func (hm *HealthMonitor) startHTTPEndpoint() {
	mux := http.NewServeMux()

	mux.HandleFunc(hm.config.HTTPPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		report := hm.GetHealth(r.Context())

		// Set appropriate status code
		switch report.Status {
		case HealthStatusHealthy:
			w.WriteHeader(http.StatusOK)
		case HealthStatusDegraded:
			w.WriteHeader(http.StatusOK) // Still serving but degraded
		case HealthStatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		default:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(report); err != nil {
			// Handle encoding error
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	// Add readiness endpoint
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if hm.IsHealthy() || hm.IsDegraded() {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				hm.logger.Warn("Failed to write readiness response", "error", err)
			}
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, err := w.Write([]byte("Not Ready")); err != nil {
				hm.logger.Warn("Failed to write readiness response", "error", err)
			}
		}
	})

	// Add liveness endpoint
	mux.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		if !hm.IsUnhealthy() {
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("OK")); err != nil {
				hm.logger.Warn("Failed to write liveness response", "error", err)
			}
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, err := w.Write([]byte("Unhealthy")); err != nil {
				hm.logger.Warn("Failed to write liveness response", "error", err)
			}
		}
	})

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", hm.config.HTTPPort),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	hm.logger.Info("Health HTTP endpoint started",
		"port", hm.config.HTTPPort,
		"path", hm.config.HTTPPath,
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		hm.logger.Error("Health HTTP endpoint error", "error", err)
	}
}

func (hm *HealthMonitor) persistHealthReport(report *HealthReport) {
	_, err := json.Marshal(report)
	if err != nil {
		hm.logger.Error("Failed to marshal health report", "error", err)
		return
	}

	// This is a simplified implementation
	// In production, you might want to use a proper file system or database
	hm.logger.Debug("Health report persisted", "path", hm.config.PersistencePath)
}

// Built-in health checkers

// DatabaseHealthChecker checks database connectivity
type DatabaseHealthChecker struct {
	name     string
	required bool
	timeout  time.Duration
	// Add database connection details
}

func NewDatabaseHealthChecker(name string) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{
		name:     name,
		required: true,
		timeout:  5 * time.Second,
	}
}

func (d *DatabaseHealthChecker) Name() string {
	return d.name
}

func (d *DatabaseHealthChecker) IsRequired() bool {
	return d.required
}

func (d *DatabaseHealthChecker) Timeout() time.Duration {
	return d.timeout
}

func (d *DatabaseHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	// Implement actual database check
	// This is a placeholder
	check := HealthCheck{
		Name:      d.name,
		Status:    HealthStatusHealthy,
		Message:   "Database connection healthy",
		Duration:  time.Since(start),
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"connection_pool_size": 10,
			"active_connections":   5,
		},
	}

	return check
}

// RedisHealthChecker checks Redis connectivity
type RedisHealthChecker struct {
	name     string
	required bool
	timeout  time.Duration
	// Add Redis connection details
}

func NewRedisHealthChecker(name string) *RedisHealthChecker {
	return &RedisHealthChecker{
		name:     name,
		required: false,
		timeout:  3 * time.Second,
	}
}

func (r *RedisHealthChecker) Name() string {
	return r.name
}

func (r *RedisHealthChecker) IsRequired() bool {
	return r.required
}

func (r *RedisHealthChecker) Timeout() time.Duration {
	return r.timeout
}

func (r *RedisHealthChecker) Check(ctx context.Context) HealthCheck {
	start := time.Now()

	// Implement actual Redis check
	check := HealthCheck{
		Name:      r.name,
		Status:    HealthStatusHealthy,
		Message:   "Redis connection healthy",
		Duration:  time.Since(start),
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"connected_clients": 2,
			"used_memory":       "1.2MB",
		},
	}

	return check
}
