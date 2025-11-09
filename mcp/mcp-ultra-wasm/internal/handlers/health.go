package handlers

import (
	"encoding/json"
	"net/http"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "alive"}); err != nil {
		// Handle encoding error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *HealthHandler) Ready(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ready"}); err != nil {
		// Handle encoding error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		// Handle encoding error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *HealthHandler) Livez(w http.ResponseWriter, r *http.Request) {
	h.Live(w, r)
}

func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	h.Ready(w, r)
}

func (h *HealthHandler) Metrics() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("# Metrics placeholder\n"))
	})
}
