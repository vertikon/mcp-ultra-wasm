// pkg/router/mux_test.go
package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewMux(t *testing.T) {
	mux := NewMux()
	if mux == nil {
		t.Fatal("NewMux() returned nil")
	}
	if mux.R == nil {
		t.Fatal("Mux.R is nil")
	}
}

func TestMuxHandle(t *testing.T) {
	mux := NewMux()

	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	mux.Handle("GET", "/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if !called {
		t.Error("Handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "ok" {
		t.Errorf("Expected body 'ok', got '%s'", w.Body.String())
	}
}

func TestMuxMiddleware(t *testing.T) {
	mux := NewMux()

	middlewareCalled := false
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			w.Header().Set("X-Test", "middleware")
			next.ServeHTTP(w, r)
		})
	}

	mux.Use(middleware)

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("GET", "/test", handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if !middlewareCalled {
		t.Error("Middleware was not called")
	}

	if w.Header().Get("X-Test") != "middleware" {
		t.Error("Middleware did not set header")
	}
}

func TestMuxMultipleRoutes(t *testing.T) {
	mux := NewMux()

	mux.Handle("GET", "/route1", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("route1")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	mux.Handle("POST", "/route2", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("route2")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	// Test route1
	req1 := httptest.NewRequest("GET", "/route1", nil)
	w1 := httptest.NewRecorder()
	mux.ServeHTTP(w1, req1)

	if w1.Body.String() != "route1" {
		t.Errorf("Expected 'route1', got '%s'", w1.Body.String())
	}

	// Test route2
	req2 := httptest.NewRequest("POST", "/route2", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)

	if w2.Body.String() != "route2" {
		t.Errorf("Expected 'route2', got '%s'", w2.Body.String())
	}
}
