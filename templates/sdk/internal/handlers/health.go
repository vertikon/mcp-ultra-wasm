package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

type HealthHandler struct {
	logger *slog.Logger
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}

func (h *HealthHandler) Live(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "alive"}); err != nil {
		h.logger.Error("Error encoding live response", "error", err)
	}
}

func (h *HealthHandler) Ready(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ready"}); err != nil {
		h.logger.Error("Error encoding ready response", "error", err)
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		h.logger.Error("Error encoding health response", "error", err)
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
		if _, err := w.Write([]byte("# Metrics placeholder\n")); err != nil {
			h.logger.Error("Error writing metrics response", "error", err)
		}
	})
}
