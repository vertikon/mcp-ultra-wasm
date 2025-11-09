// pkg/router/middleware/logger_test.go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogger(t *testing.T) {
	middleware := Logger()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("ok")); err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})

	wrapped := middleware(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	w := httptest.NewRecorder()

	wrapped.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "ok" {
		t.Errorf("Expected body 'ok', got '%s'", w.Body.String())
	}
}

func TestLoggerMultipleRequests(t *testing.T) {
	middleware := Logger()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := middleware(handler)

	// Request 1
	req1 := httptest.NewRequest("GET", "/path1", nil)
	w1 := httptest.NewRecorder()
	wrapped.ServeHTTP(w1, req1)

	// Request 2
	req2 := httptest.NewRequest("POST", "/path2", nil)
	w2 := httptest.NewRecorder()
	wrapped.ServeHTTP(w2, req2)

	if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
		t.Error("Expected all requests to return 200")
	}
}
