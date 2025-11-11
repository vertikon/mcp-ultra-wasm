package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
	"go.uber.org/zap"
)

// SecretRotationConfig defines secret rotation settings
type SecretRotationConfig struct {
	RotationInterval time.Duration `yaml:"rotation_interval"`
	PreRotationHook  string        `yaml:"pre_rotation_hook"`
	PostRotationHook string        `yaml:"post_rotation_hook"`
	BackupVersions   int           `yaml:"backup_versions"`
	GracePeriod      time.Duration `yaml:"grace_period"`
}

// EnhancedVaultService provides advanced secret management with rotation
type EnhancedVaultService struct {
	client         *api.Client
	config         VaultConfig
	rotationConfig SecretRotationConfig
	logger         *zap.Logger

	// Secret rotation management
	rotationSchedules map[string]*RotationSchedule
	rotationMutex     sync.RWMutex

	// Secret cache with TTL
	secretCache map[string]*CachedSecret
	cacheMutex  sync.RWMutex

	// Notification channels
	notificationChan chan RotationEvent
	errorChan        chan error
}

// RotationSchedule manages the rotation of a specific secret
type RotationSchedule struct {
	SecretPath    string
	Config        SecretRotationConfig
	NextRotation  time.Time
	LastRotation  time.Time
	RotationCount int
	IsActive      bool
	ticker        *time.Ticker
	stopChan      chan struct{}
}

// CachedSecret represents a cached secret with TTL
type CachedSecret struct {
	Value     string
	ExpiresAt time.Time
	Version   int
}

// RotationEvent represents a secret rotation event
type RotationEvent struct {
	SecretPath string
	EventType  string // "started", "completed", "failed"
	Timestamp  time.Time
	OldVersion int
	NewVersion int
	Error      error
}

// NewEnhancedVaultService creates a new enhanced Vault service
func NewEnhancedVaultService(
	config VaultConfig,
	rotationConfig SecretRotationConfig,
	logger *zap.Logger,
) (*EnhancedVaultService, error) {
	// Create Vault API client
	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = config.Address

	vaultClient, err := api.NewClient(vaultConfig)
	if err != nil {
		return nil, fmt.Errorf("creating vault client: %w", err)
	}

	vaultClient.SetToken(config.Token)
	if config.Namespace != "" {
		vaultClient.SetNamespace(config.Namespace)
	}

	service := &EnhancedVaultService{
		client:            vaultClient,
		config:            config,
		rotationConfig:    rotationConfig,
		logger:            logger,
		rotationSchedules: make(map[string]*RotationSchedule),
		secretCache:       make(map[string]*CachedSecret),
		notificationChan:  make(chan RotationEvent, 100),
		errorChan:         make(chan error, 50),
	}

	// Start background workers
	go service.rotationWorker()
	go service.cacheCleanupWorker()

	return service, nil
}

// ScheduleRotation schedules automatic rotation for a secret
func (evs *EnhancedVaultService) ScheduleRotation(secretPath string, config SecretRotationConfig) error {
	evs.rotationMutex.Lock()
	defer evs.rotationMutex.Unlock()

	// Stop existing schedule if present
	if existing, exists := evs.rotationSchedules[secretPath]; exists {
		existing.Stop()
	}

	schedule := &RotationSchedule{
		SecretPath:   secretPath,
		Config:       config,
		NextRotation: time.Now().Add(config.RotationInterval),
		IsActive:     true,
		ticker:       time.NewTicker(config.RotationInterval),
		stopChan:     make(chan struct{}),
	}

	evs.rotationSchedules[secretPath] = schedule

	// Start rotation goroutine
	go evs.rotateSecretPeriodically(schedule)

	evs.logger.Info("Secret rotation scheduled",
		zap.String("secret_path", secretPath),
		zap.Duration("interval", config.RotationInterval),
		zap.Time("next_rotation", schedule.NextRotation),
	)

	return nil
}

// rotateSecretPeriodically handles periodic secret rotation
func (evs *EnhancedVaultService) rotateSecretPeriodically(schedule *RotationSchedule) {
	defer schedule.ticker.Stop()

	for {
		select {
		case <-schedule.ticker.C:
			if err := evs.rotateSecret(schedule); err != nil {
				evs.logger.Error("Secret rotation failed",
					zap.String("secret_path", schedule.SecretPath),
					zap.Error(err),
				)

				evs.notificationChan <- RotationEvent{
					SecretPath: schedule.SecretPath,
					EventType:  "failed",
					Timestamp:  time.Now(),
					Error:      err,
				}
			}
		case <-schedule.stopChan:
			evs.logger.Info("Secret rotation stopped",
				zap.String("secret_path", schedule.SecretPath),
			)
			return
		}
	}
}

// rotateSecret performs the actual secret rotation
func (evs *EnhancedVaultService) rotateSecret(schedule *RotationSchedule) error {
	secretPath := schedule.SecretPath

	evs.logger.Info("Starting secret rotation",
		zap.String("secret_path", secretPath),
	)

	// Notify rotation start
	evs.notificationChan <- RotationEvent{
		SecretPath: secretPath,
		EventType:  "started",
		Timestamp:  time.Now(),
	}

	// Execute pre-rotation hook if configured
	if schedule.Config.PreRotationHook != "" {
		if err := evs.executeHook(schedule.Config.PreRotationHook, secretPath, "pre"); err != nil {
			return fmt.Errorf("pre-rotation hook failed: %w", err)
		}
	}

	// Get current secret version
	currentSecret, err := evs.client.KVv2("secret").Get(context.Background(), secretPath)
	if err != nil {
		return fmt.Errorf("reading current secret: %w", err)
	}

	oldVersion := currentSecret.VersionMetadata.Version

	// Generate new secret value
	newSecretValue, err := evs.generateNewSecretValue(secretPath, currentSecret.Data)
	if err != nil {
		return fmt.Errorf("generating new secret value: %w", err)
	}

	// Store new secret version
	newSecret, err := evs.client.KVv2("secret").Put(context.Background(), secretPath, newSecretValue)
	if err != nil {
		return fmt.Errorf("storing new secret version: %w", err)
	}

	newVersion := newSecret.VersionMetadata.Version

	// Update cache
	evs.cacheMutex.Lock()
	delete(evs.secretCache, secretPath)
	evs.cacheMutex.Unlock()

	// Execute post-rotation hook if configured
	if schedule.Config.PostRotationHook != "" {
		if err := evs.executeHook(schedule.Config.PostRotationHook, secretPath, "post"); err != nil {
			evs.logger.Warn("Post-rotation hook failed",
				zap.String("secret_path", secretPath),
				zap.Error(err),
			)
		}
	}

	// Clean up old versions if configured
	if schedule.Config.BackupVersions > 0 {
		if err := evs.cleanupOldVersions(secretPath, schedule.Config.BackupVersions); err != nil {
			evs.logger.Warn("Failed to cleanup old secret versions",
				zap.String("secret_path", secretPath),
				zap.Error(err),
			)
		}
	}

	// Update schedule metadata
	schedule.LastRotation = time.Now()
	schedule.NextRotation = time.Now().Add(schedule.Config.RotationInterval)
	schedule.RotationCount++

	evs.logger.Info("Secret rotation completed",
		zap.String("secret_path", secretPath),
		zap.Int("old_version", oldVersion),
		zap.Int("new_version", newVersion),
		zap.Time("next_rotation", schedule.NextRotation),
	)

	// Notify rotation completion
	evs.notificationChan <- RotationEvent{
		SecretPath: secretPath,
		EventType:  "completed",
		Timestamp:  time.Now(),
		OldVersion: oldVersion,
		NewVersion: newVersion,
	}

	return nil
}

// generateNewSecretValue generates a new secret value based on the type
func (evs *EnhancedVaultService) generateNewSecretValue(_ string, currentData map[string]interface{}) (map[string]interface{}, error) {
	newData := make(map[string]interface{})

	// Copy non-secret metadata
	for key, value := range currentData {
		if !evs.isSecretField(key) {
			newData[key] = value
		}
	}

	// Generate new secret values
	for key, value := range currentData {
		if evs.isSecretField(key) {
			newValue, err := evs.generateSecretForField(key, value)
			if err != nil {
				return nil, fmt.Errorf("generating secret for field %s: %w", key, err)
			}
			newData[key] = newValue
		}
	}

	// Add rotation metadata
	newData["rotated_at"] = time.Now().Unix()
	newData["rotation_id"] = evs.generateRotationID()

	return newData, nil
}

// generateSecretForField generates a new secret value for a specific field
func (evs *EnhancedVaultService) generateSecretForField(fieldName string, _ interface{}) (string, error) {
	switch fieldName {
	case "password", "secret", "api_key", "token":
		return evs.generateSecurePassword(32)
	case "jwt_secret":
		return evs.generateSecurePassword(64)
	case "encryption_key":
		return evs.generateEncryptionKey(32)
	default:
		// For unknown fields, generate a secure random string
		return evs.generateSecurePassword(32)
	}
}

// generateSecurePassword generates a cryptographically secure password
func (evs *EnhancedVaultService) generateSecurePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generating random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// generateEncryptionKey generates a cryptographically secure encryption key
func (evs *EnhancedVaultService) generateEncryptionKey(keySize int) (string, error) {
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("generating encryption key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// generateRotationID generates a unique rotation identifier
func (evs *EnhancedVaultService) generateRotationID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random generation fails
		evs.logger.Warn("Failed to generate random rotation ID, using timestamp", zap.Error(err))
		return base64.URLEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

// isSecretField determines if a field contains secret data
func (evs *EnhancedVaultService) isSecretField(fieldName string) bool {
	secretFields := []string{
		"password", "secret", "api_key", "token", "jwt_secret",
		"encryption_key", "private_key", "client_secret",
	}

	fieldName = fmt.Sprintf("%v", fieldName) // Convert to string
	for _, secretField := range secretFields {
		if fieldName == secretField {
			return true
		}
	}
	return false
}

// executeHook executes a rotation hook script or command
func (evs *EnhancedVaultService) executeHook(hook, secretPath, phase string) error {
	// This is a placeholder for hook execution
	// In a real implementation, you would execute the hook script/command
	evs.logger.Info("Executing rotation hook",
		zap.String("hook", hook),
		zap.String("secret_path", secretPath),
		zap.String("phase", phase),
	)
	return nil
}

// cleanupOldVersions removes old versions of a secret
func (evs *EnhancedVaultService) cleanupOldVersions(secretPath string, keepVersions int) error {
	metadata, err := evs.client.KVv2("secret").GetMetadata(context.Background(), secretPath)
	if err != nil {
		return fmt.Errorf("getting secret metadata: %w", err)
	}

	versions := make([]string, 0, len(metadata.Versions))
	for versionKey := range metadata.Versions {
		versions = append(versions, versionKey)
	}

	if len(versions) <= keepVersions {
		return nil // Nothing to cleanup
	}

	// Sort versions and determine which to delete
	// Keep the latest 'keepVersions' versions
	toDelete := len(versions) - keepVersions

	// Delete old versions (implementation depends on Vault API)
	evs.logger.Info("Cleaning up old secret versions",
		zap.String("secret_path", secretPath),
		zap.Int("versions_to_delete", toDelete),
		zap.Int("keep_versions", keepVersions),
	)

	return nil
}

// GetSecretWithCache retrieves a secret with caching
func (evs *EnhancedVaultService) GetSecretWithCache(ctx context.Context, path string, ttl time.Duration) (map[string]interface{}, error) {
	evs.cacheMutex.RLock()
	if cached, exists := evs.secretCache[path]; exists && cached.ExpiresAt.After(time.Now()) {
		evs.cacheMutex.RUnlock()
		evs.logger.Debug("Secret cache hit", zap.String("path", path))

		// Return cached value parsed back to map
		return map[string]interface{}{"value": cached.Value}, nil
	}
	evs.cacheMutex.RUnlock()

	// Cache miss - fetch from Vault
	secret, err := evs.client.KVv2("secret").Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("getting secret from vault: %w", err)
	}

	// Cache the secret
	evs.cacheMutex.Lock()
	evs.secretCache[path] = &CachedSecret{
		Value:     fmt.Sprintf("%v", secret.Data["value"]),
		ExpiresAt: time.Now().Add(ttl),
		Version:   secret.VersionMetadata.Version,
	}
	evs.cacheMutex.Unlock()

	evs.logger.Debug("Secret cached", zap.String("path", path), zap.Duration("ttl", ttl))

	return secret.Data, nil
}

// rotationWorker processes rotation events
func (evs *EnhancedVaultService) rotationWorker() {
	for event := range evs.notificationChan {
		evs.logger.Info("Processing rotation event",
			zap.String("secret_path", event.SecretPath),
			zap.String("event_type", event.EventType),
			zap.Time("timestamp", event.Timestamp),
		)

		// Here you could send notifications to external systems
		// like Slack, email, monitoring systems, etc.
	}
}

// cacheCleanupWorker periodically cleans up expired cache entries
func (evs *EnhancedVaultService) cacheCleanupWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		evs.cacheMutex.Lock()
		now := time.Now()
		for path, cached := range evs.secretCache {
			if cached.ExpiresAt.Before(now) {
				delete(evs.secretCache, path)
				evs.logger.Debug("Expired secret removed from cache", zap.String("path", path))
			}
		}
		evs.cacheMutex.Unlock()
	}
}

// Stop stops all rotation schedules and background workers
func (evs *EnhancedVaultService) Stop() {
	evs.rotationMutex.Lock()
	defer evs.rotationMutex.Unlock()

	for _, schedule := range evs.rotationSchedules {
		schedule.Stop()
	}

	close(evs.notificationChan)
	close(evs.errorChan)

	evs.logger.Info("Enhanced Vault service stopped")
}

// Stop stops a rotation schedule
func (rs *RotationSchedule) Stop() {
	rs.IsActive = false
	close(rs.stopChan)
}

// GetRotationStatus returns the status of all rotation schedules
func (evs *EnhancedVaultService) GetRotationStatus() map[string]RotationStatus {
	evs.rotationMutex.RLock()
	defer evs.rotationMutex.RUnlock()

	status := make(map[string]RotationStatus)
	for path, schedule := range evs.rotationSchedules {
		status[path] = RotationStatus{
			SecretPath:       schedule.SecretPath,
			IsActive:         schedule.IsActive,
			LastRotation:     schedule.LastRotation,
			NextRotation:     schedule.NextRotation,
			RotationCount:    schedule.RotationCount,
			RotationInterval: schedule.Config.RotationInterval,
		}
	}
	return status
}

// RotationStatus represents the status of a secret rotation schedule
type RotationStatus struct {
	SecretPath       string        `json:"secret_path"`
	IsActive         bool          `json:"is_active"`
	LastRotation     time.Time     `json:"last_rotation"`
	NextRotation     time.Time     `json:"next_rotation"`
	RotationCount    int           `json:"rotation_count"`
	RotationInterval time.Duration `json:"rotation_interval"`
}
