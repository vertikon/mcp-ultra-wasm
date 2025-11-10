package security

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const authMethodToken = "token"

// VaultConfig holds Vault configuration
type VaultConfig struct {
	Address   string        `yaml:"address"`
	Token     string        `yaml:"token"`
	Namespace string        `yaml:"namespace,omitempty"`
	Timeout   time.Duration `yaml:"timeout"`
	// Auth method configuration
	AuthMethod string `yaml:"auth_method"` // token, k8s, aws, etc.
	Role       string `yaml:"role,omitempty"`
}

// VaultService provides secure secret management using HashiCorp Vault
type VaultService struct {
	config   VaultConfig
	client   *http.Client
	logger   *zap.Logger
	token    string
	tokenMux sync.RWMutex
}

// SecretData represents secret data from Vault
type SecretData struct {
	Data     map[string]interface{} `json:"data"`
	Metadata SecretMetadata         `json:"metadata"`
}

// SecretMetadata contains secret metadata
type SecretMetadata struct {
	Version     int       `json:"version"`
	CreatedTime time.Time `json:"created_time"`
	UpdatedTime time.Time `json:"updated_time"`
	Destroyed   bool      `json:"destroyed"`
}

// VaultResponse represents a generic Vault API response
type VaultResponse struct {
	Data     json.RawMessage `json:"data"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// VaultAuth represents Vault authentication response
type VaultAuth struct {
	ClientToken   string   `json:"client_token"`
	Accessor      string   `json:"accessor"`
	LeaseDuration int      `json:"lease_duration"`
	Renewable     bool     `json:"renewable"`
	Policies      []string `json:"policies"`
}

// VaultAuthResponse represents authentication response
type VaultAuthResponse struct {
	Auth VaultAuth `json:"auth"`
}

// NewVaultService creates a new Vault service
func NewVaultService(config VaultConfig, logger *zap.Logger) *VaultService {
	vs := &VaultService{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
		token:  config.Token,
	}

	// Start token renewal goroutine if using token auth
	if config.AuthMethod == authMethodToken && config.Token != "" {
		go vs.renewToken(context.Background())
	}

	return vs
}

// GetSecret retrieves a secret from Vault
func (vs *VaultService) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
	vs.tokenMux.RLock()
	token := vs.token
	vs.tokenMux.RUnlock()

	if token == "" {
		return nil, fmt.Errorf("no valid Vault token available")
	}

	// Construct URL
	url := fmt.Sprintf("%s/v1/%s", vs.config.Address, path)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Set("X-Vault-Token", token)
	if vs.config.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", vs.config.Namespace)
	}

	// Execute request
	resp, err := vs.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vs.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	// Handle response
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("secret not found: %s", path)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Vault returned status %d for path %s", resp.StatusCode, path)
	}

	var vaultResp VaultResponse
	if err := json.NewDecoder(resp.Body).Decode(&vaultResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// For KV v2, the actual data is nested under "data"
	if strings.HasPrefix(path, "secret/data/") {
		var secretData SecretData
		if err := json.Unmarshal(vaultResp.Data, &secretData); err != nil {
			return nil, fmt.Errorf("unmarshaling secret data: %w", err)
		}
		return secretData.Data, nil
	}

	// For other engines, return the data directly
	var data map[string]interface{}
	if err := json.Unmarshal(vaultResp.Data, &data); err != nil {
		return nil, fmt.Errorf("unmarshaling data: %w", err)
	}

	return data, nil
}

// PutSecret stores a secret in Vault
func (vs *VaultService) PutSecret(ctx context.Context, path string, data map[string]interface{}) error {
	vs.tokenMux.RLock()
	token := vs.token
	vs.tokenMux.RUnlock()

	if token == "" {
		return fmt.Errorf("no valid Vault token available")
	}

	// For KV v2, wrap the data
	var payload interface{}
	if strings.HasPrefix(path, "secret/data/") {
		payload = map[string]interface{}{
			"data": data,
		}
	} else {
		payload = data
	}

	// Marshal payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	// Construct URL
	url := fmt.Sprintf("%s/v1/%s", vs.config.Address, path)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Vault-Token", token)
	if vs.config.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", vs.config.Namespace)
	}

	// Execute request
	resp, err := vs.client.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vs.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Vault returned status %d for path %s", resp.StatusCode, path)
	}

	vs.logger.Info("Secret stored in Vault", zap.String("path", path))
	return nil
}

// GetDatabaseCredentials retrieves database credentials from Vault
func (vs *VaultService) GetDatabaseCredentials(ctx context.Context, role string) (string, string, error) {
	path := fmt.Sprintf("database/creds/%s", role)

	data, err := vs.GetSecret(ctx, path)
	if err != nil {
		return "", "", fmt.Errorf("getting database credentials: %w", err)
	}

	username, ok := data["username"].(string)
	if !ok {
		return "", "", fmt.Errorf("username not found in database credentials")
	}

	// Extract password value from Vault response (not a hardcoded secret)
	password, ok := data["password"].(string)
	if !ok {
		return "", "", fmt.Errorf("password field not found in database credentials")
	}

	return username, password, nil
}

// renewToken renews the Vault token periodically
func (vs *VaultService) renewToken(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute) // Renew every 30 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := vs.renewCurrentToken(ctx); err != nil {
				vs.logger.Error("Failed to renew Vault token", zap.Error(err))
			}
		}
	}
}

// renewCurrentToken renews the current token
func (vs *VaultService) renewCurrentToken(ctx context.Context) error {
	vs.tokenMux.RLock()
	token := vs.token
	vs.tokenMux.RUnlock()

	if token == "" {
		return fmt.Errorf("no token to renew")
	}

	url := fmt.Sprintf("%s/v1/auth/token/renew-self", vs.config.Address)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("creating renew request: %w", err)
	}

	req.Header.Set("X-Vault-Token", token)
	if vs.config.Namespace != "" {
		req.Header.Set("X-Vault-Namespace", vs.config.Namespace)
	}

	resp, err := vs.client.Do(req)
	if err != nil {
		return fmt.Errorf("executing renew request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vs.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token renewal failed with status %d", resp.StatusCode)
	}

	vs.logger.Debug("Vault token renewed successfully")
	return nil
}

// HealthCheck checks if Vault is healthy and accessible
func (vs *VaultService) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/v1/sys/health", vs.config.Address)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating health check request: %w", err)
	}

	resp, err := vs.client.Do(req)
	if err != nil {
		return fmt.Errorf("Vault health check failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vs.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	// Vault health endpoint returns 200 when initialized and unsealed
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Vault health check returned status %d", resp.StatusCode)
	}

	return nil
}

// GetJWTSigningKey retrieves JWT signing key from Vault
func (vs *VaultService) GetJWTSigningKey(ctx context.Context) (string, error) {
	data, err := vs.GetSecret(ctx, "secret/data/jwt")
	if err != nil {
		return "", fmt.Errorf("getting JWT signing key: %w", err)
	}

	signingKey, ok := data["signing_key"].(string)
	if !ok {
		return "", fmt.Errorf("signing key not found in secret")
	}

	return signingKey, nil
}

// Close closes the Vault service
func (vs *VaultService) Close() error {
	// Revoke token if needed
	if vs.config.AuthMethod != "token" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = vs.revokeToken(ctx)
	}
	return nil
}

// revokeToken revokes the current token
func (vs *VaultService) revokeToken(ctx context.Context) error {
	vs.tokenMux.RLock()
	token := vs.token
	vs.tokenMux.RUnlock()

	if token == "" {
		return nil
	}

	url := fmt.Sprintf("%s/v1/auth/token/revoke-self", vs.config.Address)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("creating revoke request: %w", err)
	}

	req.Header.Set("X-Vault-Token", token)

	resp, err := vs.client.Do(req)
	if err != nil {
		return fmt.Errorf("executing revoke request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vs.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	return nil
}
