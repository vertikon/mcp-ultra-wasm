package security

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	resourceTasks   = "tasks"
	resourceUnknown = "unknown"
)

// OPAConfig holds OPA configuration
type OPAConfig struct {
	URL     string        `yaml:"url"`
	Timeout time.Duration `yaml:"timeout"`
}

// OPAService handles Open Policy Agent authorization
type OPAService struct {
	config OPAConfig
	client *http.Client
	logger *zap.Logger
}

// AuthzRequest represents authorization request to OPA
type AuthzRequest struct {
	Input AuthzInput `json:"input"`
}

// AuthzInput contains the authorization input data
type AuthzInput struct {
	User     *Claims `json:"user"`
	Method   string  `json:"method"`
	Path     string  `json:"path"`
	Resource string  `json:"resource,omitempty"`
	Action   string  `json:"action,omitempty"`
}

// AuthzResponse represents OPA authorization response
type AuthzResponse struct {
	Result struct {
		Allow  bool   `json:"allow"`
		Deny   bool   `json:"deny,omitempty"`
		Reason string `json:"reason,omitempty"`
	} `json:"result"`
}

// NewOPAService creates a new OPA service
func NewOPAService(config OPAConfig, logger *zap.Logger) *OPAService {
	return &OPAService{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}
}

// IsAuthorized checks if user is authorized to perform the requested action
func (opa *OPAService) IsAuthorized(ctx context.Context, claims *Claims, method, path string) bool {
	// Build authorization input
	input := AuthzInput{
		User:   claims,
		Method: method,
		Path:   path,
	}

	// Extract resource and action from path and method
	resource, action := opa.extractResourceAction(method, path)
	input.Resource = resource
	input.Action = action

	// Create authorization request
	authzReq := AuthzRequest{Input: input}

	// Marshal request to JSON
	reqBody, err := json.Marshal(authzReq)
	if err != nil {
		opa.logger.Error("Failed to marshal authz request", zap.Error(err))
		return false
	}

	// Create HTTP request to OPA
	req, err := http.NewRequestWithContext(ctx, "POST", opa.config.URL+"/v1/data/authz", bytes.NewBuffer(reqBody))
	if err != nil {
		opa.logger.Error("Failed to create OPA request", zap.Error(err))
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := opa.client.Do(req)
	if err != nil {
		opa.logger.Error("OPA request failed", zap.Error(err))
		return false
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			opa.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		opa.logger.Warn("OPA returned non-200 status",
			zap.Int("status", resp.StatusCode),
			zap.String("user_id", claims.UserID),
			zap.String("path", path))
		return false
	}

	// Parse response
	var authzResp AuthzResponse
	if err := json.NewDecoder(resp.Body).Decode(&authzResp); err != nil {
		opa.logger.Error("Failed to decode OPA response", zap.Error(err))
		return false
	}

	// Log authorization decision
	if authzResp.Result.Allow {
		opa.logger.Debug("Authorization granted",
			zap.String("user_id", claims.UserID),
			zap.String("role", claims.Role),
			zap.String("method", method),
			zap.String("path", path))
	} else {
		opa.logger.Info("Authorization denied",
			zap.String("user_id", claims.UserID),
			zap.String("role", claims.Role),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("reason", authzResp.Result.Reason))
	}

	return authzResp.Result.Allow
}

// IsAuthorizedForResource checks authorization for specific resource action
func (opa *OPAService) IsAuthorizedForResource(ctx context.Context, claims *Claims, resource, action string) bool {
	input := AuthzInput{
		User:     claims,
		Resource: resource,
		Action:   action,
	}

	authzReq := AuthzRequest{Input: input}

	reqBody, err := json.Marshal(authzReq)
	if err != nil {
		opa.logger.Error("Failed to marshal resource authz request", zap.Error(err))
		return false
	}

	req, err := http.NewRequestWithContext(ctx, "POST", opa.config.URL+"/v1/data/authz/resource", bytes.NewBuffer(reqBody))
	if err != nil {
		opa.logger.Error("Failed to create OPA resource request", zap.Error(err))
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := opa.client.Do(req)
	if err != nil {
		opa.logger.Error("OPA resource request failed", zap.Error(err))
		return false
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			opa.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		opa.logger.Warn("OPA resource returned non-200 status",
			zap.Int("status", resp.StatusCode),
			zap.String("user_id", claims.UserID),
			zap.String("resource", resource),
			zap.String("action", action))
		return false
	}

	var authzResp AuthzResponse
	if err := json.NewDecoder(resp.Body).Decode(&authzResp); err != nil {
		opa.logger.Error("Failed to decode OPA resource response", zap.Error(err))
		return false
	}

	return authzResp.Result.Allow
}

// extractResourceAction extracts resource and action from HTTP method and path
func (opa *OPAService) extractResourceAction(method, path string) (string, string) {
	// Simple mapping for common REST patterns
	switch method {
	case "GET":
		if path == "/api/v1/tasks" {
			return resourceTasks, "list"
		}
		if len(path) > 0 && path[len(path)-1] != '/' {
			return resourceTasks, "read"
		}
		return resourceUnknown, "read"
	case "POST":
		return resourceTasks, "create"
	case "PUT", "PATCH":
		return resourceTasks, "update"
	case "DELETE":
		return resourceTasks, "delete"
	default:
		return resourceUnknown, resourceUnknown
	}
}

// HealthCheck checks if OPA is healthy
func (opa *OPAService) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", opa.config.URL+"/health", nil)
	if err != nil {
		return fmt.Errorf("creating health check request: %w", err)
	}

	resp, err := opa.client.Do(req)
	if err != nil {
		return fmt.Errorf("OPA health check failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			opa.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OPA health check returned status %d", resp.StatusCode)
	}

	return nil
}
