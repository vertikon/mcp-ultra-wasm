package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const contentTypeJSON = "application/json"

func TestHealthHandler_Live(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rec := httptest.NewRecorder()

	h.Live(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != contentTypeJSON {
		t.Errorf("esperado application/json, obteve %s", ct)
	}
}

func TestHealthHandler_Ready(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", rec.Code)
	}
}

func TestHealthHandler_Health(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", rec.Code)
	}
}

func TestHealthHandler_Metrics(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	h.Metrics().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != "text/plain" {
		t.Errorf("esperado text/plain, obteve %s", ct)
	}
}

func TestHealthHandler_Livez(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rec := httptest.NewRecorder()

	h.Livez(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != contentTypeJSON {
		t.Errorf("esperado application/json, obteve %s", ct)
	}
}

func TestHealthHandler_Readyz(t *testing.T) {
	h := NewHealthHandler()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	h.Readyz(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperado 200, obteve %d", rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if ct != contentTypeJSON {
		t.Errorf("esperado application/json, obteve %s", ct)
	}
}
