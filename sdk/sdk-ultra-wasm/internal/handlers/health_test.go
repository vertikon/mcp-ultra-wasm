package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	handler := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.Health(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}
	var p map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if p["status"] != "ok" {
		t.Fatalf(`status=%q; want "ok"`, p["status"])
	}
}

func TestHealthHandler_Live(t *testing.T) {
	handler := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rr := httptest.NewRecorder()
	handler.Live(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}
	var p map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if p["status"] != "alive" {
		t.Fatalf(`status=%q; want "alive"`, p["status"])
	}
}

func TestHealthHandler_Ready(t *testing.T) {
	handler := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	handler.Ready(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}
	var p map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if p["status"] != "ready" {
		t.Fatalf(`status=%q; want "ready"`, p["status"])
	}
}

func TestHealthHandler_Methods(t *testing.T) {
	handler := NewHealthHandler()

	tests := []struct {
		name           string
		method         func(http.ResponseWriter, *http.Request)
		expectedStatus string
	}{
		{"Health", handler.Health, "ok"},
		{"Live", handler.Live, "alive"},
		{"Ready", handler.Ready, "ready"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()
			tt.method(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("status = %d; want %d", rr.Code, http.StatusOK)
			}

			var p map[string]string
			if err := json.Unmarshal(rr.Body.Bytes(), &p); err != nil {
				t.Errorf("invalid json: %v", err)
			}

			if p["status"] != tt.expectedStatus {
				t.Errorf("status = %q; want %q", p["status"], tt.expectedStatus)
			}
		})
	}
}

func TestHealthHandler_ContentType(t *testing.T) {
	handler := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.Health(rr, req)

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q; want application/json", contentType)
	}
}

func TestHealthHandler_DifferentHTTPMethods(t *testing.T) {
	handler := NewHealthHandler()

	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			rr := httptest.NewRecorder()
			handler.Health(rr, req)

			// Should work regardless of HTTP method
			if rr.Code != http.StatusOK {
				t.Errorf("status = %d for method %s; want %d", rr.Code, method, http.StatusOK)
			}
		})
	}
}
