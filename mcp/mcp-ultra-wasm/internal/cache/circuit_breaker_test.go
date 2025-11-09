package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second, 2)

	// Circuit breaker should start in closed state
	assert.Equal(t, CircuitBreakerClosed, cb.State())

	// Successful calls should keep circuit closed
	for i := 0; i < 5; i++ {
		allowed := cb.Allow()
		assert.True(t, allowed, "Request should be allowed in closed state")
		cb.RecordSuccess()
		assert.Equal(t, CircuitBreakerClosed, cb.State())
	}
}

func TestCircuitBreaker_OpenState(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond, 2)

	// Generate failures to open the circuit
	for i := 0; i < 3; i++ {
		allowed := cb.Allow()
		assert.True(t, allowed, "Request should be allowed before circuit opens")
		cb.RecordFailure()
	}

	// Circuit should now be open
	assert.Equal(t, CircuitBreakerOpen, cb.State())

	// Calls should fail immediately without executing the function
	allowed := cb.Allow()
	assert.False(t, allowed, "Request should be denied when circuit is open")
}

func TestCircuitBreaker_HalfOpenState(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond, 2)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Allow()
		cb.RecordFailure()
	}

	assert.Equal(t, CircuitBreakerOpen, cb.State())

	// Wait for timeout to transition to half-open
	time.Sleep(60 * time.Millisecond)

	// Next call should transition to half-open and be allowed
	allowed := cb.Allow()
	assert.True(t, allowed, "First request after timeout should be allowed")
	assert.Equal(t, CircuitBreakerHalfOpen, cb.State())

	// Record success to transition back to closed
	cb.RecordSuccess()
	cb.Allow()
	cb.RecordSuccess()
	cb.Allow()
	cb.RecordSuccess()

	// Should transition back to closed after success threshold
	assert.Equal(t, CircuitBreakerClosed, cb.State())
}

func TestCircuitBreaker_HalfOpenToOpen(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond, 2)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Allow()
		cb.RecordFailure()
	}

	assert.Equal(t, CircuitBreakerOpen, cb.State())

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Fail again in half-open state
	allowed := cb.Allow()
	assert.True(t, allowed, "First request after timeout should be allowed")
	assert.Equal(t, CircuitBreakerHalfOpen, cb.State())

	cb.RecordFailure()

	// Should transition back to open
	assert.Equal(t, CircuitBreakerOpen, cb.State())
}

func TestCircuitBreaker_Stats(t *testing.T) {
	cb := NewCircuitBreaker(5, time.Second, 2)

	// Execute some successful calls
	for i := 0; i < 3; i++ {
		cb.Allow()
		cb.RecordSuccess()
	}

	// Execute some failed calls
	for i := 0; i < 2; i++ {
		cb.Allow()
		cb.RecordFailure()
	}

	stats := cb.Stats()
	assert.Equal(t, CircuitBreakerClosed, stats.State)
	assert.Equal(t, 2, stats.FailureCount)
	assert.Equal(t, 5, stats.FailureThreshold)
	assert.Equal(t, time.Second, stats.RecoveryTimeout)
}

func TestCircuitBreaker_ConcurrentExecution(t *testing.T) {
	cb := NewCircuitBreaker(10, time.Second, 5)

	numGoroutines := 50
	results := make(chan bool, numGoroutines)

	// Execute concurrent calls
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			allowed := cb.Allow()
			if allowed {
				time.Sleep(10 * time.Millisecond) // Simulate some work
				if i%10 == 0 {
					cb.RecordFailure()
				} else {
					cb.RecordSuccess()
				}
			}
			results <- allowed
		}(i)
	}

	// Collect results
	allowedCount := 0
	for i := 0; i < numGoroutines; i++ {
		if <-results {
			allowedCount++
		}
	}

	// Most requests should be allowed
	assert.Greater(t, allowedCount, 40, "Most requests should be allowed")
	// Circuit should still be closed (not enough failures)
	assert.Equal(t, CircuitBreakerClosed, cb.State())
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(2, time.Second, 2)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Allow()
		cb.RecordFailure()
	}

	assert.Equal(t, CircuitBreakerOpen, cb.State())

	// Reset the circuit breaker
	cb.Reset()

	// Should be back to closed state
	assert.Equal(t, CircuitBreakerClosed, cb.State())

	// Should accept calls normally
	allowed := cb.Allow()
	assert.True(t, allowed, "Request should be allowed after reset")
	cb.RecordSuccess()
	assert.Equal(t, CircuitBreakerClosed, cb.State())
}

func TestCircuitBreaker_ForceOpen(t *testing.T) {
	cb := NewCircuitBreaker(5, time.Second, 2)

	// Circuit starts closed
	assert.Equal(t, CircuitBreakerClosed, cb.State())

	// Force open
	cb.ForceOpen()

	// Should be open now
	assert.Equal(t, CircuitBreakerOpen, cb.State())

	// Requests should be denied
	allowed := cb.Allow()
	assert.False(t, allowed, "Request should be denied when circuit is forced open")
}

func TestCircuitBreaker_OnStateChange(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond, 2)

	stateChanges := make(chan string, 10)

	cb.OnStateChange(func(from, to CircuitBreakerState) {
		stateChanges <- from.String() + " -> " + to.String()
	})

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Allow()
		cb.RecordFailure()
	}

	// Wait for callback
	time.Sleep(10 * time.Millisecond)

	select {
	case change := <-stateChanges:
		assert.Equal(t, "closed -> open", change)
	default:
		t.Error("Expected state change callback")
	}

	// Wait for timeout and transition to half-open
	time.Sleep(60 * time.Millisecond)

	cb.Allow() // This triggers the transition

	// Wait for callback
	time.Sleep(10 * time.Millisecond)

	select {
	case change := <-stateChanges:
		assert.Equal(t, "open -> half_open", change)
	default:
		t.Error("Expected state change callback for half-open")
	}
}

func TestCircuitBreaker_HalfOpenMaxRequests(t *testing.T) {
	cb := NewCircuitBreaker(2, 50*time.Millisecond, 3)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Allow()
		cb.RecordFailure()
	}

	assert.Equal(t, CircuitBreakerOpen, cb.State())

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Should allow up to HalfOpenMaxRequests (3)
	for i := 0; i < 3; i++ {
		allowed := cb.Allow()
		assert.True(t, allowed, "Request %d should be allowed in half-open", i)
	}

	// Next request should be denied
	allowed := cb.Allow()
	assert.False(t, allowed, "Request should be denied after max half-open requests")
}

func TestAdaptiveCircuitBreaker_Creation(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold:    5,
		RecoveryTimeout:     time.Second,
		HalfOpenMaxRequests: 3,
		SuccessThreshold:    3,
	}

	acb := NewAdaptiveCircuitBreaker(config)
	assert.NotNil(t, acb)
	assert.Equal(t, CircuitBreakerClosed, acb.State())

	stats := acb.Stats()
	assert.Equal(t, 5, stats.FailureThreshold)
	assert.Equal(t, time.Second, stats.RecoveryTimeout)
	assert.Equal(t, 3, stats.HalfOpenMaxRequests)
}

func TestAdaptiveCircuitBreaker_RecordRequest(t *testing.T) {
	config := CircuitBreakerConfig{
		FailureThreshold:    5,
		RecoveryTimeout:     time.Second,
		HalfOpenMaxRequests: 3,
		SuccessThreshold:    3,
	}

	acb := NewAdaptiveCircuitBreaker(config)

	// Record some requests
	for i := 0; i < 10; i++ {
		acb.RecordRequest()
	}

	// Record some failures
	for i := 0; i < 2; i++ {
		acb.RecordFailure()
	}

	// Check failure rate
	rate := acb.GetFailureRate()
	assert.Greater(t, rate, 0.0, "Failure rate should be greater than 0")
	assert.Less(t, rate, 1.0, "Failure rate should be less than 1")
}
