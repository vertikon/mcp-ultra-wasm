package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/features"
)

// FeatureFlagHandlers handles HTTP requests for feature flags
type FeatureFlagHandlers struct {
	flagManager *features.FlagManager
	logger      *zap.Logger
}

// NewFeatureFlagHandlers creates new feature flag handlers
func NewFeatureFlagHandlers(flagManager *features.FlagManager, logger *zap.Logger) *FeatureFlagHandlers {
	return &FeatureFlagHandlers{
		flagManager: flagManager,
		logger:      logger,
	}
}

// ListFlags handles listing all feature flags
func (h *FeatureFlagHandlers) ListFlags(w http.ResponseWriter, r *http.Request) {
	flags, err := h.flagManager.ListFlags(r.Context())
	if err != nil {
		h.logger.Error("Failed to list feature flags", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list flags", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, flags)
}

// GetFlag handles retrieving a specific feature flag
func (h *FeatureFlagHandlers) GetFlag(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Flag key is required", nil)
		return
	}

	flag, err := h.flagManager.GetFlag(r.Context(), key)
	if err != nil {
		h.logger.Error("Failed to get feature flag", zap.String("key", key), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Flag not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, flag)
}

// CreateFlag handles creating a new feature flag
func (h *FeatureFlagHandlers) CreateFlag(w http.ResponseWriter, r *http.Request) {
	var req CreateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if err := req.Validate(); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request", err)
		return
	}

	flag := &domain.FeatureFlag{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		Strategy:    req.Strategy,
		Parameters:  req.Parameters,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.flagManager.SetFlag(r.Context(), flag); err != nil {
		h.logger.Error("Failed to create feature flag", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create flag", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, flag)
}

// UpdateFlag handles updating an existing feature flag
func (h *FeatureFlagHandlers) UpdateFlag(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Flag key is required", nil)
		return
	}

	var req UpdateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Get existing flag
	flag, err := h.flagManager.GetFlag(r.Context(), key)
	if err != nil {
		h.logger.Error("Failed to get feature flag for update", zap.String("key", key), zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Flag not found", err)
		return
	}

	// Update fields if provided
	if req.Name != nil {
		flag.Name = *req.Name
	}
	if req.Description != nil {
		flag.Description = *req.Description
	}
	if req.Enabled != nil {
		flag.Enabled = *req.Enabled
	}
	if req.Strategy != nil {
		flag.Strategy = *req.Strategy
	}
	if req.Parameters != nil {
		flag.Parameters = req.Parameters
	}

	flag.UpdatedAt = time.Now()

	if err := h.flagManager.SetFlag(r.Context(), flag); err != nil {
		h.logger.Error("Failed to update feature flag", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update flag", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, flag)
}

// DeleteFlag handles deleting a feature flag
func (h *FeatureFlagHandlers) DeleteFlag(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Flag key is required", nil)
		return
	}

	if err := h.flagManager.DeleteFlag(r.Context(), key); err != nil {
		h.logger.Error("Failed to delete feature flag", zap.String("key", key), zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete flag", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// EvaluateFlag handles feature flag evaluation
func (h *FeatureFlagHandlers) EvaluateFlag(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Flag key is required", nil)
		return
	}

	var req EvaluateFlagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	if req.UserID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	result := h.flagManager.EvaluateFlag(r.Context(), key, req.UserID, req.Attributes)

	response := EvaluateFlagResponse{
		Key:        key,
		UserID:     req.UserID,
		Enabled:    result,
		Attributes: req.Attributes,
		Timestamp:  time.Now(),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// writeJSONResponse writes a JSON response
func (h *FeatureFlagHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// writeErrorResponse writes an error response
func (h *FeatureFlagHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Error: message,
		Code:  statusCode,
	}

	if err != nil {
		errorResponse.Details = err.Error()
	}

	if encodeErr := json.NewEncoder(w).Encode(errorResponse); encodeErr != nil {
		h.logger.Error("Failed to encode error response", zap.Error(encodeErr))
	}
}

// Request and Response types
type CreateFlagRequest struct {
	Key         string                 `json:"key"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Strategy    string                 `json:"strategy"`
	Parameters  map[string]interface{} `json:"parameters"`
}

func (r CreateFlagRequest) Validate() error {
	if r.Key == "" {
		return fmt.Errorf("key is required")
	}
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

type UpdateFlagRequest struct {
	Name        *string                `json:"name"`
	Description *string                `json:"description"`
	Enabled     *bool                  `json:"enabled"`
	Strategy    *string                `json:"strategy"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type EvaluateFlagRequest struct {
	UserID     string                 `json:"user_id"`
	Attributes map[string]interface{} `json:"attributes"`
}

type EvaluateFlagResponse struct {
	Key        string                 `json:"key"`
	UserID     string                 `json:"user_id"`
	Enabled    bool                   `json:"enabled"`
	Attributes map[string]interface{} `json:"attributes"`
	Timestamp  time.Time              `json:"timestamp"`
}
