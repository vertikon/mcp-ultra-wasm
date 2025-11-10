package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/internal/seeds"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

// SeedSyncRequest representa uma solicitação de sincronização de seed
type SeedSyncRequest struct {
	TemplatePath string `json:"template_path,omitempty"`
}

// SeedSyncResponse representa a resposta de sincronização
type SeedSyncResponse struct {
	Status  string `json:"status"`
	Seed    string `json:"seed,omitempty"`
	Message string `json:"message,omitempty"`
}

// SeedSyncHandler sincroniza o template para a seed interna
func SeedSyncHandler(w http.ResponseWriter, r *http.Request) {
	var req SeedSyncRequest

	// Decodificar request (se houver body)
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("Failed to decode request body", "error", err)
		}
	}

	// Usar caminho padrão se não especificado
	if req.TemplatePath == "" {
		req.TemplatePath = `E:\vertikon\business\SaaS\templates\mcp-ultra-wasm`
	}

	// Executar sincronização
	err := seeds.Sync(req.TemplatePath)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		w.Header().Set("Content-Type", "application/json")
		resp := SeedSyncResponse{
			Status:  "error",
			Message: err.Error(),
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			logger.Error("Error encoding error response", "error", err)
		}
		return
	}

	// Sucesso
	w.Header().Set("Content-Type", "application/json")
	resp := SeedSyncResponse{
		Status: "ok",
		Seed:   "seeds/mcp-ultra-wasm",
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("Error encoding success response", "error", err)
	}
}

// SeedStatusHandler retorna o status da seed interna
func SeedStatusHandler(w http.ResponseWriter, _ *http.Request) {
	status := seeds.Status()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		logger.Error("Error encoding status response", "error", err)
	}
}
