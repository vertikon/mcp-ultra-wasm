package features

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
)

// FlagManager manages feature flags with persistence
type FlagManager struct {
	flags     map[string]*domain.FeatureFlag
	mu        sync.RWMutex
	repo      domain.FeatureFlagRepository
	cache     domain.CacheRepository
	logger    *zap.Logger
	refresher *time.Ticker
	stopCh    chan struct{}
}

// NewFlagManager creates a new feature flag manager
func NewFlagManager(repo domain.FeatureFlagRepository, cache domain.CacheRepository, logger *zap.Logger) *FlagManager {
	manager := &FlagManager{
		flags:  make(map[string]*domain.FeatureFlag),
		repo:   repo,
		cache:  cache,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	// Start background refresh
	go manager.startRefresh(30 * time.Second)

	return manager
}

// IsEnabled checks if a feature flag is enabled
func (m *FlagManager) IsEnabled(ctx context.Context, key string) bool {
	flag, err := m.GetFlag(ctx, key)
	if err != nil {
		m.logger.Debug("Feature flag not found", zap.String("key", key))
		return false
	}

	return flag.Enabled
}

// IsEnabledWithDefault checks if a feature flag is enabled with a default value
func (m *FlagManager) IsEnabledWithDefault(ctx context.Context, key string, defaultValue bool) bool {
	flag, err := m.GetFlag(ctx, key)
	if err != nil {
		return defaultValue
	}

	return flag.Enabled
}

// GetFlag retrieves a feature flag
func (m *FlagManager) GetFlag(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	// Try memory cache first
	m.mu.RLock()
	if flag, exists := m.flags[key]; exists {
		m.mu.RUnlock()
		return flag, nil
	}
	m.mu.RUnlock()

	// Try Redis cache
	cacheKey := fmt.Sprintf("flag:%s", key)
	if cachedData, err := m.cache.Get(ctx, cacheKey); err == nil {
		var flag domain.FeatureFlag
		if json.Unmarshal([]byte(cachedData), &flag) == nil {
			// Update memory cache
			m.mu.Lock()
			m.flags[key] = &flag
			m.mu.Unlock()
			return &flag, nil
		}
	}

	// Get from repository
	flag, err := m.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("getting feature flag: %w", err)
	}

	// Cache in Redis for 5 minutes
	if err := m.cache.Set(ctx, cacheKey, flag, 300); err != nil {
		m.logger.Error("Failed to cache feature flag", zap.Error(err))
	}

	// Update memory cache
	m.mu.Lock()
	m.flags[key] = flag
	m.mu.Unlock()

	return flag, nil
}

// SetFlag creates or updates a feature flag
func (m *FlagManager) SetFlag(ctx context.Context, flag *domain.FeatureFlag) error {
	// Save to repository
	existingFlag, err := m.repo.GetByKey(ctx, flag.Key)
	if err != nil {
		// Create new flag
		if err := m.repo.Create(ctx, flag); err != nil {
			return fmt.Errorf("creating feature flag: %w", err)
		}
	} else {
		// Update existing flag
		existingFlag.Name = flag.Name
		existingFlag.Description = flag.Description
		existingFlag.Enabled = flag.Enabled
		existingFlag.Strategy = flag.Strategy
		existingFlag.Parameters = flag.Parameters
		existingFlag.UpdatedAt = time.Now()

		if err := m.repo.Update(ctx, existingFlag); err != nil {
			return fmt.Errorf("updating feature flag: %w", err)
		}
		flag = existingFlag
	}

	// Update caches
	m.mu.Lock()
	m.flags[flag.Key] = flag
	m.mu.Unlock()

	cacheKey := fmt.Sprintf("flag:%s", flag.Key)
	if err := m.cache.Set(ctx, cacheKey, flag, 300); err != nil {
		m.logger.Error("Failed to cache feature flag", zap.Error(err))
	}

	m.logger.Info("Feature flag updated",
		zap.String("key", flag.Key),
		zap.Bool("enabled", flag.Enabled))

	return nil
}

// ListFlags returns all feature flags
func (m *FlagManager) ListFlags(ctx context.Context) ([]*domain.FeatureFlag, error) {
	return m.repo.List(ctx)
}

// DeleteFlag deletes a feature flag
func (m *FlagManager) DeleteFlag(ctx context.Context, key string) error {
	if err := m.repo.Delete(ctx, key); err != nil {
		return fmt.Errorf("deleting feature flag: %w", err)
	}

	// Remove from caches
	m.mu.Lock()
	delete(m.flags, key)
	m.mu.Unlock()

	cacheKey := fmt.Sprintf("flag:%s", key)
	if err := m.cache.Delete(ctx, cacheKey); err != nil {
		m.logger.Error("Failed to delete cached feature flag", zap.Error(err))
	}

	m.logger.Info("Feature flag deleted", zap.String("key", key))

	return nil
}

// RefreshFlags reloads all flags from the repository
func (m *FlagManager) RefreshFlags(ctx context.Context) error {
	flags, err := m.repo.List(ctx)
	if err != nil {
		return fmt.Errorf("refreshing feature flags: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Clear existing flags
	m.flags = make(map[string]*domain.FeatureFlag)

	// Load new flags
	for _, flag := range flags {
		m.flags[flag.Key] = flag

		// Update Redis cache
		cacheKey := fmt.Sprintf("flag:%s", flag.Key)
		if err := m.cache.Set(ctx, cacheKey, flag, 300); err != nil {
			m.logger.Error("Failed to cache feature flag during refresh", zap.Error(err))
		}
	}

	m.logger.Info("Feature flags refreshed", zap.Int("count", len(flags)))

	return nil
}

// startRefresh starts background refresh of feature flags
func (m *FlagManager) startRefresh(interval time.Duration) {
	m.refresher = time.NewTicker(interval)
	defer m.refresher.Stop()

	for {
		select {
		case <-m.refresher.C:
			ctx := context.Background()
			if err := m.RefreshFlags(ctx); err != nil {
				m.logger.Error("Failed to refresh feature flags", zap.Error(err))
			}
		case <-m.stopCh:
			return
		}
	}
}

// Stop stops the background refresh
func (m *FlagManager) Stop() {
	if m.refresher != nil {
		m.refresher.Stop()
	}
	close(m.stopCh)
}

// EvaluateFlag evaluates a feature flag with strategy
func (m *FlagManager) EvaluateFlag(ctx context.Context, key string, userID string, attributes map[string]interface{}) bool {
	flag, err := m.GetFlag(ctx, key)
	if err != nil {
		return false
	}

	if !flag.Enabled {
		return false
	}

	switch flag.Strategy {
	case "simple":
		return true
	case "percentage":
		return m.evaluatePercentage(flag, userID)
	case "userlist":
		return m.evaluateUserList(flag, userID)
	case "attribute":
		return m.evaluateAttribute(flag, attributes)
	default:
		m.logger.Warn("Unknown flag strategy",
			zap.String("key", key),
			zap.String("strategy", flag.Strategy))
		return false
	}
}

// evaluatePercentage evaluates percentage-based rollout
func (m *FlagManager) evaluatePercentage(flag *domain.FeatureFlag, userID string) bool {
	if percentage, ok := flag.Parameters["percentage"].(float64); ok {
		// Simple hash-based percentage calculation
		hash := 0
		for _, c := range userID {
			hash = hash*31 + int(c)
		}
		return (hash % 100) < int(percentage)
	}
	return false
}

// evaluateUserList evaluates user list strategy
func (m *FlagManager) evaluateUserList(flag *domain.FeatureFlag, userID string) bool {
	if userListInterface, ok := flag.Parameters["users"]; ok {
		if userList, ok := userListInterface.([]interface{}); ok {
			for _, user := range userList {
				if userStr, ok := user.(string); ok && userStr == userID {
					return true
				}
			}
		}
	}
	return false
}

// evaluateAttribute evaluates attribute-based strategy
func (m *FlagManager) evaluateAttribute(flag *domain.FeatureFlag, attributes map[string]interface{}) bool {
	if conditions, ok := flag.Parameters["conditions"].(map[string]interface{}); ok {
		for key, expectedValue := range conditions {
			if actualValue, exists := attributes[key]; !exists || actualValue != expectedValue {
				return false
			}
		}
		return true
	}
	return false
}
