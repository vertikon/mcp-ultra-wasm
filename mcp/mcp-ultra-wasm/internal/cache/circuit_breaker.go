package cache

import (
	"sync"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// String returns string representation of circuit breaker state
func (s CircuitBreakerState) String() string {
	switch s {
	case CircuitBreakerClosed:
		return "closed"
	case CircuitBreakerOpen:
		return "open"
	case CircuitBreakerHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold    int           `yaml:"failure_threshold"`
	RecoveryTimeout     time.Duration `yaml:"recovery_timeout"`
	HalfOpenMaxRequests int           `yaml:"half_open_max_requests"`
	SuccessThreshold    int           `yaml:"success_threshold"`
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu                  sync.RWMutex
	state               CircuitBreakerState
	failureCount        int
	successCount        int
	halfOpenRequests    int
	lastFailureTime     time.Time
	failureThreshold    int
	successThreshold    int
	recoveryTimeout     time.Duration
	halfOpenMaxRequests int

	// Callbacks
	onStateChange func(from, to CircuitBreakerState)
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failureThreshold int, recoveryTimeout time.Duration, halfOpenMaxRequests int) *CircuitBreaker {
	return &CircuitBreaker{
		state:               CircuitBreakerClosed,
		failureThreshold:    failureThreshold,
		successThreshold:    3, // Default success threshold for half-open -> closed
		recoveryTimeout:     recoveryTimeout,
		halfOpenMaxRequests: halfOpenMaxRequests,
	}
}

// Allow checks if the request should be allowed through
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true

	case CircuitBreakerOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailureTime) >= cb.recoveryTimeout {
			cb.setState(CircuitBreakerHalfOpen)
			cb.halfOpenRequests = 0
			cb.successCount = 0
			return true
		}
		return false

	case CircuitBreakerHalfOpen:
		// Allow limited requests in half-open state
		if cb.halfOpenRequests < cb.halfOpenMaxRequests {
			cb.halfOpenRequests++
			return true
		}
		return false

	default:
		return false
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitBreakerClosed:
		cb.failureCount = 0 // Reset failure count on success

	case CircuitBreakerHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.setState(CircuitBreakerClosed)
			cb.failureCount = 0
			cb.successCount = 0
			cb.halfOpenRequests = 0
		}
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.lastFailureTime = time.Now()

	switch cb.state {
	case CircuitBreakerClosed:
		cb.failureCount++
		if cb.failureCount >= cb.failureThreshold {
			cb.setState(CircuitBreakerOpen)
		}

	case CircuitBreakerHalfOpen:
		cb.setState(CircuitBreakerOpen)
		cb.halfOpenRequests = 0
		cb.successCount = 0
	}
}

// State returns the current state of the circuit breaker
func (cb *CircuitBreaker) State() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}

// Stats returns circuit breaker statistics
func (cb *CircuitBreaker) Stats() CircuitBreakerStats {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return CircuitBreakerStats{
		State:               cb.state,
		FailureCount:        cb.failureCount,
		SuccessCount:        cb.successCount,
		HalfOpenRequests:    cb.halfOpenRequests,
		LastFailureTime:     cb.lastFailureTime,
		FailureThreshold:    cb.failureThreshold,
		SuccessThreshold:    cb.successThreshold,
		RecoveryTimeout:     cb.recoveryTimeout,
		HalfOpenMaxRequests: cb.halfOpenMaxRequests,
	}
}

// CircuitBreakerStats contains circuit breaker statistics
type CircuitBreakerStats struct {
	State               CircuitBreakerState `json:"state"`
	FailureCount        int                 `json:"failure_count"`
	SuccessCount        int                 `json:"success_count"`
	HalfOpenRequests    int                 `json:"half_open_requests"`
	LastFailureTime     time.Time           `json:"last_failure_time"`
	FailureThreshold    int                 `json:"failure_threshold"`
	SuccessThreshold    int                 `json:"success_threshold"`
	RecoveryTimeout     time.Duration       `json:"recovery_timeout"`
	HalfOpenMaxRequests int                 `json:"half_open_max_requests"`
}

// OnStateChange sets a callback for state changes
func (cb *CircuitBreaker) OnStateChange(callback func(from, to CircuitBreakerState)) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.onStateChange = callback
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := cb.state
	cb.state = CircuitBreakerClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.halfOpenRequests = 0
	cb.lastFailureTime = time.Time{}

	if cb.onStateChange != nil && oldState != CircuitBreakerClosed {
		cb.onStateChange(oldState, CircuitBreakerClosed)
	}
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := cb.state
	cb.setState(CircuitBreakerOpen)
	cb.lastFailureTime = time.Now()

	if cb.onStateChange != nil && oldState != CircuitBreakerOpen {
		cb.onStateChange(oldState, CircuitBreakerOpen)
	}
}

// setState sets the state and triggers callback if registered
func (cb *CircuitBreaker) setState(newState CircuitBreakerState) {
	oldState := cb.state
	cb.state = newState

	if cb.onStateChange != nil && oldState != newState {
		go cb.onStateChange(oldState, newState) // Non-blocking callback
	}
}

// AdaptiveCircuitBreaker extends CircuitBreaker with adaptive behavior
type AdaptiveCircuitBreaker struct {
	*CircuitBreaker

	mu                   sync.RWMutex
	recentRequests       []time.Time
	recentFailures       []time.Time
	adaptiveWindow       time.Duration
	minFailureRate       float64
	maxFailureRate       float64
	adaptiveThreshold    bool
	baseFailureThreshold int
	maxFailureThreshold  int
}

// NewAdaptiveCircuitBreaker creates an adaptive circuit breaker
func NewAdaptiveCircuitBreaker(config CircuitBreakerConfig) *AdaptiveCircuitBreaker {
	cb := NewCircuitBreaker(
		config.FailureThreshold,
		config.RecoveryTimeout,
		config.HalfOpenMaxRequests,
	)

	acb := &AdaptiveCircuitBreaker{
		CircuitBreaker:       cb,
		adaptiveWindow:       time.Minute,
		minFailureRate:       0.1, // 10%
		maxFailureRate:       0.5, // 50%
		adaptiveThreshold:    true,
		baseFailureThreshold: config.FailureThreshold,
		maxFailureThreshold:  config.FailureThreshold * 3,
		recentRequests:       make([]time.Time, 0),
		recentFailures:       make([]time.Time, 0),
	}

	// Set up adaptive behavior
	go acb.adaptiveAdjustment()

	return acb
}

// RecordRequest records a request (for adaptive behavior)
func (acb *AdaptiveCircuitBreaker) RecordRequest() {
	acb.mu.Lock()
	defer acb.mu.Unlock()

	now := time.Now()
	acb.recentRequests = append(acb.recentRequests, now)

	// Clean old requests
	cutoff := now.Add(-acb.adaptiveWindow)
	i := 0
	for i < len(acb.recentRequests) && acb.recentRequests[i].Before(cutoff) {
		i++
	}
	acb.recentRequests = acb.recentRequests[i:]
}

// RecordFailure records a failure with adaptive behavior
func (acb *AdaptiveCircuitBreaker) RecordFailure() {
	acb.mu.Lock()
	now := time.Now()
	acb.recentFailures = append(acb.recentFailures, now)

	// Clean old failures
	cutoff := now.Add(-acb.adaptiveWindow)
	i := 0
	for i < len(acb.recentFailures) && acb.recentFailures[i].Before(cutoff) {
		i++
	}
	acb.recentFailures = acb.recentFailures[i:]
	acb.mu.Unlock()

	// Call parent implementation
	acb.CircuitBreaker.RecordFailure()
}

// GetFailureRate returns the current failure rate
func (acb *AdaptiveCircuitBreaker) GetFailureRate() float64 {
	acb.mu.RLock()
	defer acb.mu.RUnlock()

	if len(acb.recentRequests) == 0 {
		return 0
	}

	return float64(len(acb.recentFailures)) / float64(len(acb.recentRequests))
}

// adaptiveAdjustment runs in background to adjust thresholds
func (acb *AdaptiveCircuitBreaker) adaptiveAdjustment() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !acb.adaptiveThreshold {
			continue
		}

		failureRate := acb.GetFailureRate()

		acb.CircuitBreaker.mu.Lock()
		currentThreshold := acb.CircuitBreaker.failureThreshold

		// Adjust threshold based on failure rate
		newThreshold := currentThreshold

		if failureRate > acb.maxFailureRate {
			// High failure rate - be more sensitive
			newThreshold = max(acb.baseFailureThreshold/2, 1)
		} else if failureRate < acb.minFailureRate {
			// Low failure rate - be less sensitive
			newThreshold = min(currentThreshold*2, acb.maxFailureThreshold)
		}

		if newThreshold != currentThreshold {
			acb.CircuitBreaker.failureThreshold = newThreshold
		}

		acb.CircuitBreaker.mu.Unlock()
	}
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
