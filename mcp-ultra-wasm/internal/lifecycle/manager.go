package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/logger"
)

// State represents the current state of the application
type State int32

const (
	StateInitializing State = iota
	StateStarting
	StateReady
	StateHealthy
	StateDegraded
	StateStopping
	StateStopped
	StateError
)

func (s State) String() string {
	switch s {
	case StateInitializing:
		return "initializing"
	case StateStarting:
		return "starting"
	case StateReady:
		return "ready"
	case StateHealthy:
		return string(HealthStatusHealthy)
	case StateDegraded:
		return "degraded"
	case StateStopping:
		return "stopping"
	case StateStopped:
		return "stopped"
	case StateError:
		return "error"
	default:
		return "unknown"
	}
}

// Component represents a lifecycle-managed component
type Component interface {
	Name() string
	Priority() int // Lower number = higher priority for startup order
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	HealthCheck(ctx context.Context) error
	IsReady() bool
	IsHealthy() bool
}

// Event represents events during lifecycle transitions
type Event struct {
	Type      string                 `json:"type"`
	Component string                 `json:"component,omitempty"`
	State     State                  `json:"state"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Error     error                  `json:"error,omitempty"`
}

// Manager manages application lifecycle and component orchestration
type Manager struct {
	state         int32 // atomic State
	components    []Component
	eventHandlers []func(Event)

	// Context and cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Synchronization
	mu sync.RWMutex
	wg sync.WaitGroup

	// Configuration
	config Config

	// Dependencies
	logger    *logger.Logger
	telemetry *observability.TelemetryService

	// State tracking
	startTime       time.Time
	readyTime       time.Time
	componentStates map[string]ComponentState
	eventHistory    []Event
	maxEventHistory int

	// Health monitoring
	healthTicker     *time.Ticker
	gracefulShutdown chan struct{}
	forceShutdown    chan struct{}
}

// ComponentState tracks individual component state
type ComponentState struct {
	Name        string            `json:"name"`
	State       string            `json:"state"`
	LastHealthy time.Time         `json:"last_healthy"`
	ErrorCount  int               `json:"error_count"`
	StartTime   time.Time         `json:"start_time"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Config configures the lifecycle manager
type Config struct {
	StartupTimeout          time.Duration `json:"startup_timeout"`
	ShutdownTimeout         time.Duration `json:"shutdown_timeout"`
	HealthCheckInterval     time.Duration `json:"health_check_interval"`
	MaxRetries              int           `json:"max_retries"`
	RetryDelay              time.Duration `json:"retry_delay"`
	GracefulShutdownTimeout time.Duration `json:"graceful_shutdown_timeout"`
	MaxEventHistory         int           `json:"max_event_history"`
	EnableMetrics           bool          `json:"enable_metrics"`
	EnableTracing           bool          `json:"enable_tracing"`
}

// DefaultConfig returns default lifecycle manager configuration
func DefaultConfig() Config {
	return Config{
		StartupTimeout:          2 * time.Minute,
		ShutdownTimeout:         30 * time.Second,
		HealthCheckInterval:     30 * time.Second,
		MaxRetries:              3,
		RetryDelay:              5 * time.Second,
		GracefulShutdownTimeout: 15 * time.Second,
		MaxEventHistory:         1000,
		EnableMetrics:           true,
		EnableTracing:           true,
	}
}

// NewManager creates a new lifecycle manager
func NewManager(
	config Config,
	logger *logger.Logger,
	telemetry *observability.TelemetryService,
) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	lm := &Manager{
		state:            int32(StateInitializing),
		components:       make([]Component, 0),
		eventHandlers:    make([]func(Event), 0),
		ctx:              ctx,
		cancel:           cancel,
		config:           config,
		logger:           logger,
		telemetry:        telemetry,
		componentStates:  make(map[string]ComponentState),
		eventHistory:     make([]Event, 0, config.MaxEventHistory),
		maxEventHistory:  config.MaxEventHistory,
		gracefulShutdown: make(chan struct{}),
		forceShutdown:    make(chan struct{}),
	}

	// Initialize health monitoring
	if config.HealthCheckInterval > 0 {
		lm.healthTicker = time.NewTicker(config.HealthCheckInterval)
	}

	lm.emitEvent("lifecycle_manager_created", "", StateInitializing, "Lifecycle manager initialized", nil, nil)

	return lm
}

// RegisterComponent registers a component for lifecycle management
func (lm *Manager) RegisterComponent(component Component) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.components = append(lm.components, component)
	lm.componentStates[component.Name()] = ComponentState{
		Name:        component.Name(),
		State:       "registered",
		LastHealthy: time.Time{},
		ErrorCount:  0,
		StartTime:   time.Time{},
		Metadata:    make(map[string]string),
	}

	lm.logger.Info("Component registered",
		"component", component.Name(),
		"priority", component.Priority(),
	)

	lm.emitEvent("component_registered", component.Name(), lm.GetState(),
		fmt.Sprintf("Component %s registered with priority %d", component.Name(), component.Priority()),
		map[string]interface{}{"priority": component.Priority()}, nil)
}

// RegisterEventHandler registers an event handler for lifecycle events
func (lm *Manager) RegisterEventHandler(handler func(Event)) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.eventHandlers = append(lm.eventHandlers, handler)
}

// Start starts all registered components in priority order
func (lm *Manager) Start(ctx context.Context) error {
	lm.setState(StateStarting)
	lm.startTime = time.Now()

	lm.logger.Info("Starting application lifecycle")
	lm.emitEvent("startup_initiated", "", StateStarting, "Application startup initiated", nil, nil)

	// Start components in priority order (lower number = higher priority)
	components := lm.getSortedComponents()

	startupCtx, cancel := context.WithTimeout(ctx, lm.config.StartupTimeout)
	defer cancel()

	for _, component := range components {
		if err := lm.startComponent(startupCtx, component); err != nil {
			lm.setState(StateError)
			lm.emitEvent("startup_failed", component.Name(), StateError,
				fmt.Sprintf("Failed to start component %s", component.Name()),
				nil, err)
			return fmt.Errorf("failed to start component %s: %w", component.Name(), err)
		}
	}

	lm.setState(StateReady)
	lm.readyTime = time.Now()

	lm.logger.Info("Application started successfully",
		"startup_duration", time.Since(lm.startTime),
		"components_count", len(lm.components),
	)

	lm.emitEvent("startup_completed", "", StateReady, "Application startup completed successfully",
		map[string]interface{}{
			"startup_duration": time.Since(lm.startTime).String(),
			"components_count": len(lm.components),
		}, nil)

	// Start health monitoring
	lm.startHealthMonitoring()

	// Start background tasks
	lm.startBackgroundTasks()

	return nil
}

// Stop stops all components in reverse priority order
func (lm *Manager) Stop(ctx context.Context) error {
	lm.setState(StateStopping)

	lm.logger.Info("Stopping application lifecycle")
	lm.emitEvent("shutdown_initiated", "", StateStopping, "Application shutdown initiated", nil, nil)

	// Signal graceful shutdown
	close(lm.gracefulShutdown)

	// Stop health monitoring
	if lm.healthTicker != nil {
		lm.healthTicker.Stop()
	}

	// Stop components in reverse priority order
	components := lm.getSortedComponents()

	shutdownCtx, cancel := context.WithTimeout(ctx, lm.config.ShutdownTimeout)
	defer cancel()

	// Reverse order for shutdown
	for i := len(components) - 1; i >= 0; i-- {
		component := components[i]
		if err := lm.stopComponent(shutdownCtx, component); err != nil {
			lm.logger.Error("Failed to stop component gracefully",
				"component", component.Name(),
				"error", err,
			)
			// Continue stopping other components even if one fails
		}
	}

	// Cancel context and wait for background tasks
	lm.cancel()
	lm.wg.Wait()

	lm.setState(StateStopped)

	lm.logger.Info("Application stopped successfully")
	lm.emitEvent("shutdown_completed", "", StateStopped, "Application shutdown completed", nil, nil)

	return nil
}

// GetState returns the current lifecycle state
func (lm *Manager) GetState() State {
	return State(atomic.LoadInt32(&lm.state))
}

// IsReady returns true if the application is ready to serve requests
func (lm *Manager) IsReady() bool {
	state := lm.GetState()
	return state == StateReady || state == StateHealthy || state == StateDegraded
}

// IsHealthy returns true if the application is healthy
func (lm *Manager) IsHealthy() bool {
	return lm.GetState() == StateHealthy
}

// GetComponentStates returns the current state of all components
func (lm *Manager) GetComponentStates() map[string]ComponentState {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	states := make(map[string]ComponentState)
	for k, v := range lm.componentStates {
		states[k] = v
	}
	return states
}

// GetEventHistory returns recent lifecycle events
func (lm *Manager) GetEventHistory(limit int) []Event {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if limit <= 0 || limit > len(lm.eventHistory) {
		limit = len(lm.eventHistory)
	}

	// Return most recent events
	start := len(lm.eventHistory) - limit
	events := make([]Event, limit)
	copy(events, lm.eventHistory[start:])

	return events
}

// GetMetrics returns lifecycle metrics
func (lm *Manager) GetMetrics() Metrics {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	metrics := Metrics{
		State:             lm.GetState().String(),
		ComponentCount:    len(lm.components),
		ReadyComponents:   0,
		HealthyComponents: 0,
		ErrorComponents:   0,
		EventCount:        len(lm.eventHistory),
	}

	if !lm.startTime.IsZero() {
		metrics.Uptime = time.Since(lm.startTime)
	}

	if !lm.readyTime.IsZero() {
		metrics.StartupDuration = lm.readyTime.Sub(lm.startTime)
	}

	// Count component states
	for _, state := range lm.componentStates {
		switch state.State {
		case "ready":
			metrics.ReadyComponents++
		case string(HealthStatusHealthy):
			metrics.HealthyComponents++
		case "error":
			metrics.ErrorComponents++
		}
	}

	return metrics
}

// Metrics contains lifecycle metrics
type Metrics struct {
	State             string        `json:"state"`
	ComponentCount    int           `json:"component_count"`
	ReadyComponents   int           `json:"ready_components"`
	HealthyComponents int           `json:"healthy_components"`
	ErrorComponents   int           `json:"error_components"`
	Uptime            time.Duration `json:"uptime"`
	StartupDuration   time.Duration `json:"startup_duration"`
	EventCount        int           `json:"event_count"`
}

// Private methods

func (lm *Manager) setState(state State) {
	atomic.StoreInt32(&lm.state, int32(state))
}

func (lm *Manager) getSortedComponents() []Component {
	lm.mu.RLock()
	components := make([]Component, len(lm.components))
	copy(components, lm.components)
	lm.mu.RUnlock()

	// Sort by priority (lower number = higher priority)
	for i := 0; i < len(components)-1; i++ {
		for j := i + 1; j < len(components); j++ {
			if components[i].Priority() > components[j].Priority() {
				components[i], components[j] = components[j], components[i]
			}
		}
	}

	return components
}

func (lm *Manager) startComponent(ctx context.Context, component Component) error {
	name := component.Name()

	lm.logger.Info("Starting component", "component", name)
	lm.updateComponentState(name, "starting", nil)

	lm.emitEvent("component_starting", name, lm.GetState(),
		fmt.Sprintf("Starting component %s", name), nil, nil)

	// Start with retries
	var lastErr error
	for attempt := 1; attempt <= lm.config.MaxRetries; attempt++ {
		if err := component.Start(ctx); err != nil {
			lastErr = err
			lm.logger.Warn("Component start failed",
				"component", name,
				"attempt", attempt,
				"max_attempts", lm.config.MaxRetries,
				"error", err,
			)

			if attempt < lm.config.MaxRetries {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(lm.config.RetryDelay):
					continue
				}
			}
		} else {
			lastErr = nil
			break
		}
	}

	if lastErr != nil {
		lm.updateComponentState(name, "error", lastErr)
		lm.emitEvent("component_start_failed", name, lm.GetState(),
			fmt.Sprintf("Failed to start component %s after %d attempts", name, lm.config.MaxRetries),
			map[string]interface{}{"attempts": lm.config.MaxRetries}, lastErr)
		return lastErr
	}

	lm.updateComponentState(name, "ready", nil)
	lm.emitEvent("component_started", name, lm.GetState(),
		fmt.Sprintf("Component %s started successfully", name), nil, nil)

	lm.logger.Info("Component started successfully", "component", name)
	return nil
}

func (lm *Manager) stopComponent(ctx context.Context, component Component) error {
	name := component.Name()

	lm.logger.Info("Stopping component", "component", name)
	lm.updateComponentState(name, "stopping", nil)

	lm.emitEvent("component_stopping", name, lm.GetState(),
		fmt.Sprintf("Stopping component %s", name), nil, nil)

	if err := component.Stop(ctx); err != nil {
		lm.updateComponentState(name, "error", err)
		lm.emitEvent("component_stop_failed", name, lm.GetState(),
			fmt.Sprintf("Failed to stop component %s", name), nil, err)
		return err
	}

	lm.updateComponentState(name, "stopped", nil)
	lm.emitEvent("component_stopped", name, lm.GetState(),
		fmt.Sprintf("Component %s stopped successfully", name), nil, nil)

	lm.logger.Info("Component stopped successfully", "component", name)
	return nil
}

func (lm *Manager) updateComponentState(name, state string, err error) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	compState := lm.componentStates[name]
	compState.State = state

	if err != nil {
		compState.ErrorCount++
	} else if state == string(HealthStatusHealthy) {
		compState.LastHealthy = time.Now()
		compState.ErrorCount = 0
	}

	if state == "starting" {
		compState.StartTime = time.Now()
	}

	lm.componentStates[name] = compState
}

func (lm *Manager) emitEvent(eventType, component string, state State, message string, metadata map[string]interface{}, err error) {
	event := Event{
		Type:      eventType,
		Component: component,
		State:     state,
		Message:   message,
		Timestamp: time.Now(),
		Metadata:  metadata,
		Error:     err,
	}

	// Add to history
	lm.mu.Lock()
	lm.eventHistory = append(lm.eventHistory, event)
	if len(lm.eventHistory) > lm.maxEventHistory {
		// Remove oldest events
		copy(lm.eventHistory, lm.eventHistory[len(lm.eventHistory)-lm.maxEventHistory:])
		lm.eventHistory = lm.eventHistory[:lm.maxEventHistory]
	}
	lm.mu.Unlock()

	// Notify handlers
	lm.mu.RLock()
	handlers := lm.eventHandlers
	lm.mu.RUnlock()

	for _, handler := range handlers {
		go func(h func(Event)) {
			defer func() {
				if r := recover(); r != nil {
					lm.logger.Error("Event handler panicked",
						"event_type", eventType,
						"panic", r,
					)
				}
			}()
			h(event)
		}(handler)
	}
}

func (lm *Manager) startHealthMonitoring() {
	if lm.healthTicker == nil {
		return
	}

	lm.wg.Add(1)
	go func() {
		defer lm.wg.Done()
		lm.runHealthChecks()
	}()
}

func (lm *Manager) startBackgroundTasks() {
	// Start metrics collection if enabled
	if lm.config.EnableMetrics && lm.telemetry != nil {
		lm.wg.Add(1)
		go func() {
			defer lm.wg.Done()
			lm.collectMetrics()
		}()
	}
}

func (lm *Manager) runHealthChecks() {
	for {
		select {
		case <-lm.ctx.Done():
			return
		case <-lm.gracefulShutdown:
			return
		case <-lm.healthTicker.C:
			lm.performHealthChecks()
		}
	}
}

func (lm *Manager) performHealthChecks() {
	components := lm.getSortedComponents()
	healthyCount := 0
	errorCount := 0

	for _, component := range components {
		ctx, cancel := context.WithTimeout(lm.ctx, 10*time.Second)
		err := component.HealthCheck(ctx)
		cancel()

		name := component.Name()
		if err != nil {
			lm.updateComponentState(name, "error", err)
			errorCount++
			lm.logger.Warn("Component health check failed",
				"component", name,
				"error", err,
			)
		} else {
			lm.updateComponentState(name, string(HealthStatusHealthy), nil)
			healthyCount++
		}
	}

	// Update overall health state
	totalComponents := len(components)
	if errorCount == 0 && healthyCount == totalComponents {
		lm.setState(StateHealthy)
	} else if errorCount > 0 && healthyCount > 0 {
		lm.setState(StateDegraded)
	} else if errorCount == totalComponents {
		lm.setState(StateError)
	}
}

func (lm *Manager) collectMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-lm.ctx.Done():
			return
		case <-lm.gracefulShutdown:
			return
		case <-ticker.C:
			metrics := lm.GetMetrics()

			// Emit telemetry metrics
			if lm.telemetry != nil {
				lm.telemetry.RecordCounter("lifecycle_component_count", float64(metrics.ComponentCount), map[string]string{
					"state": metrics.State,
				})

				lm.telemetry.RecordGauge("lifecycle_ready_components", float64(metrics.ReadyComponents), nil)
				lm.telemetry.RecordGauge("lifecycle_healthy_components", float64(metrics.HealthyComponents), nil)
				lm.telemetry.RecordGauge("lifecycle_error_components", float64(metrics.ErrorComponents), nil)

				if metrics.Uptime > 0 {
					lm.telemetry.RecordGauge("lifecycle_uptime_seconds", metrics.Uptime.Seconds(), nil)
				}
			}
		}
	}
}
