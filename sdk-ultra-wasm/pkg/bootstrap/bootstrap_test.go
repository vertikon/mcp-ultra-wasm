package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBootstrap(t *testing.T) {
	config := Config{
		EnableRecovery: true,
		EnableLogger:   true,
		CORSOrigins:    []string{"*"},
	}

	mux := Bootstrap(config)

	if mux == nil {
		t.Fatal("Bootstrap returned nil mux")
	}
}

func TestBootstrapMinimal(t *testing.T) {
	config := Config{}

	mux := Bootstrap(config)

	if mux == nil {
		t.Fatal("Bootstrap with minimal config returned nil mux")
	}
}

func TestMarkReady(t *testing.T) {
	MarkNotReady()

	if ready.Load() {
		t.Error("Expected ready to be false after MarkNotReady")
	}

	MarkReady()

	if !ready.Load() {
		t.Error("Expected ready to be true after MarkReady")
	}
}

func TestHealthzEndpoint(t *testing.T) {
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()

	healthz(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "ok" {
		t.Errorf("Expected body 'ok', got '%s'", w.Body.String())
	}
}

func TestReadinessEndpoint(t *testing.T) {
	// Test not ready
	MarkNotReady()
	req := httptest.NewRequest("GET", "/readyz", nil)
	w := httptest.NewRecorder()

	readiness(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 when not ready, got %d", w.Code)
	}

	// Test ready
	MarkReady()
	req = httptest.NewRequest("GET", "/readyz", nil)
	w = httptest.NewRecorder()

	readiness(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 when ready, got %d", w.Code)
	}
}

func TestBootstrapWithAllMiddlewares(t *testing.T) {
	config := Config{
		EnableRecovery: true,
		EnableLogger:   true,
		CORSOrigins:    []string{"https://example.com", "https://app.com"},
	}

	mux := Bootstrap(config)

	if mux == nil {
		t.Fatal("Bootstrap returned nil mux")
	}

	// Test that health endpoints are accessible
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health endpoint failed: status %d", w.Code)
	}
}

func TestBootstrapPingEndpoint(t *testing.T) {
	config := Config{}
	mux := Bootstrap(config)

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /ping, got %d", w.Code)
	}

	if w.Body.String() != "ok" {
		t.Errorf("Expected body 'ok', got '%s'", w.Body.String())
	}
}

func TestBootstrapOnlyRecovery(t *testing.T) {
	config := Config{
		EnableRecovery: true,
		EnableLogger:   false,
	}

	mux := Bootstrap(config)

	if mux == nil {
		t.Fatal("Bootstrap returned nil mux")
	}
}

func TestBootstrapOnlyLogger(t *testing.T) {
	config := Config{
		EnableRecovery: false,
		EnableLogger:   true,
	}

	mux := Bootstrap(config)

	if mux == nil {
		t.Fatal("Bootstrap returned nil mux")
	}
}

func TestBootstrapOnlyCORS(t *testing.T) {
	config := Config{
		CORSOrigins: []string{"*"},
	}

	mux := Bootstrap(config)

	if mux == nil {
		t.Fatal("Bootstrap returned nil mux")
	}
}

func TestHealthzMultipleCalls(t *testing.T) {
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/healthz", nil)
		w := httptest.NewRecorder()
		healthz(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Call %d: expected status 200, got %d", i, w.Code)
		}
	}
}

func TestReadinessStateTransitions(t *testing.T) {
	// Start not ready
	MarkNotReady()
	if ready.Load() {
		t.Error("Should start not ready")
	}

	// Mark ready
	MarkReady()
	if !ready.Load() {
		t.Error("Should be ready after MarkReady")
	}

	// Mark not ready again
	MarkNotReady()
	if ready.Load() {
		t.Error("Should be not ready after second MarkNotReady")
	}

	// Final ready state
	MarkReady()
	if !ready.Load() {
		t.Error("Should be ready after final MarkReady")
	}
}
