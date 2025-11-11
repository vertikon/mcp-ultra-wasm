// internal/http/router.go
package httpserver

import (
	"encoding/json"
	"net/http"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/features"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/hello", hello)
	mux.HandleFunc("/api/v1/flags/evaluate", evaluateFlag)
}

func hello(w http.ResponseWriter, _ *http.Request) {
	resp := map[string]any{"message": "hello from mcp-model-ultra"}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Handle encoding error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

var fm = features.NewInMemoryManager()

type evalRequest struct {
	Flag   string         `json:"flag"`
	UserID string         `json:"user_id"`
	Attrs  map[string]any `json:"attrs"`
}

func evaluateFlag(w http.ResponseWriter, r *http.Request) {
	var req evalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	val := fm.Evaluate(req.Flag, features.EvalContext{UserID: req.UserID, Attributes: req.Attrs})
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"flag": req.Flag, "value": val}); err != nil {
		// Handle encoding error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
