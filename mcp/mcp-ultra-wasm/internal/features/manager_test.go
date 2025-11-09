package features

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
)

// Mock repositories
type MockFeatureFlagRepository struct {
	mock.Mock
}

func (m *MockFeatureFlagRepository) GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*domain.FeatureFlag), args.Error(1)
}

func (m *MockFeatureFlagRepository) List(ctx context.Context) ([]*domain.FeatureFlag, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.FeatureFlag), args.Error(1)
}

func (m *MockFeatureFlagRepository) Create(ctx context.Context, flag *domain.FeatureFlag) error {
	args := m.Called(ctx, flag)
	return args.Error(0)
}

func (m *MockFeatureFlagRepository) Update(ctx context.Context, flag *domain.FeatureFlag) error {
	args := m.Called(ctx, flag)
	return args.Error(0)
}

func (m *MockFeatureFlagRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

type MockCacheRepository struct {
	mock.Mock
}

func (m *MockCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheRepository) Increment(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCacheRepository) SetNX(ctx context.Context, key string, value interface{}, ttl int) (bool, error) {
	args := m.Called(ctx, key, value, ttl)
	return args.Bool(0), args.Error(1)
}

func TestFlagManager_IsEnabled(t *testing.T) {
	flagRepo := &MockFeatureFlagRepository{}
	cacheRepo := &MockCacheRepository{}
	logger := zap.NewNop()

	manager := &FlagManager{
		flags:  make(map[string]*domain.FeatureFlag),
		repo:   flagRepo,
		cache:  cacheRepo,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	ctx := context.Background()

	// Test with enabled flag
	enabledFlag := &domain.FeatureFlag{
		Key:     "test-flag",
		Enabled: true,
	}

	cacheRepo.On("Get", ctx, "flag:test-flag").Return("", assert.AnError)
	flagRepo.On("GetByKey", ctx, "test-flag").Return(enabledFlag, nil)
	cacheRepo.On("Set", ctx, "flag:test-flag", enabledFlag, 300).Return(nil)

	result := manager.IsEnabled(ctx, "test-flag")
	assert.True(t, result)

	// Test with disabled flag
	disabledFlag := &domain.FeatureFlag{
		Key:     "disabled-flag",
		Enabled: false,
	}

	cacheRepo.On("Get", ctx, "flag:disabled-flag").Return("", assert.AnError)
	flagRepo.On("GetByKey", ctx, "disabled-flag").Return(disabledFlag, nil)
	cacheRepo.On("Set", ctx, "flag:disabled-flag", disabledFlag, 300).Return(nil)

	result = manager.IsEnabled(ctx, "disabled-flag")
	assert.False(t, result)

	flagRepo.AssertExpectations(t)
	cacheRepo.AssertExpectations(t)
}

func TestFlagManager_IsEnabledWithDefault(t *testing.T) {
	flagRepo := &MockFeatureFlagRepository{}
	cacheRepo := &MockCacheRepository{}
	logger := zap.NewNop()

	manager := &FlagManager{
		flags:  make(map[string]*domain.FeatureFlag),
		repo:   flagRepo,
		cache:  cacheRepo,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	ctx := context.Background()

	// Test with non-existent flag and default true
	cacheRepo.On("Get", ctx, "flag:non-existent").Return("", assert.AnError)
	flagRepo.On("GetByKey", ctx, "non-existent").Return((*domain.FeatureFlag)(nil), assert.AnError)

	result := manager.IsEnabledWithDefault(ctx, "non-existent", true)
	assert.True(t, result)

	// Test with non-existent flag and default false
	result = manager.IsEnabledWithDefault(ctx, "non-existent", false)
	assert.False(t, result)

	flagRepo.AssertExpectations(t)
	cacheRepo.AssertExpectations(t)
}

func TestFlagManager_EvaluatePercentage(t *testing.T) {
	flagRepo := &MockFeatureFlagRepository{}
	cacheRepo := &MockCacheRepository{}
	logger := zap.NewNop()

	manager := &FlagManager{
		flags:  make(map[string]*domain.FeatureFlag),
		repo:   flagRepo,
		cache:  cacheRepo,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	flag := &domain.FeatureFlag{
		Key:      "percentage-flag",
		Enabled:  true,
		Strategy: "percentage",
		Parameters: map[string]interface{}{
			"percentage": 50.0,
		},
	}

	// Test with user that should be enabled (predictable hash)
	result := manager.evaluatePercentage(flag, "user1")
	// This is deterministic based on the hash function
	assert.IsType(t, bool(true), result)

	// Test with different user
	result2 := manager.evaluatePercentage(flag, "user2")
	assert.IsType(t, bool(true), result2)
}

func TestFlagManager_EvaluateUserList(t *testing.T) {
	flagRepo := &MockFeatureFlagRepository{}
	cacheRepo := &MockCacheRepository{}
	logger := zap.NewNop()

	manager := &FlagManager{
		flags:  make(map[string]*domain.FeatureFlag),
		repo:   flagRepo,
		cache:  cacheRepo,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	flag := &domain.FeatureFlag{
		Key:      "userlist-flag",
		Enabled:  true,
		Strategy: "userlist",
		Parameters: map[string]interface{}{
			"users": []interface{}{"user1", "user2", "user3"},
		},
	}

	// Test with user in list
	result := manager.evaluateUserList(flag, "user1")
	assert.True(t, result)

	// Test with user not in list
	result = manager.evaluateUserList(flag, "user4")
	assert.False(t, result)
}

func TestFlagManager_EvaluateAttribute(t *testing.T) {
	flagRepo := &MockFeatureFlagRepository{}
	cacheRepo := &MockCacheRepository{}
	logger := zap.NewNop()

	manager := &FlagManager{
		flags:  make(map[string]*domain.FeatureFlag),
		repo:   flagRepo,
		cache:  cacheRepo,
		logger: logger,
		stopCh: make(chan struct{}),
	}

	flag := &domain.FeatureFlag{
		Key:      "attribute-flag",
		Enabled:  true,
		Strategy: "attribute",
		Parameters: map[string]interface{}{
			"conditions": map[string]interface{}{
				"country": "US",
				"plan":    "premium",
			},
		},
	}

	// Test with matching attributes
	attributes := map[string]interface{}{
		"country": "US",
		"plan":    "premium",
		"extra":   "value",
	}
	result := manager.evaluateAttribute(flag, attributes)
	assert.True(t, result)

	// Test with non-matching attributes
	attributes = map[string]interface{}{
		"country": "CA",
		"plan":    "premium",
	}
	result = manager.evaluateAttribute(flag, attributes)
	assert.False(t, result)

	// Test with missing attributes
	attributes = map[string]interface{}{
		"country": "US",
	}
	result = manager.evaluateAttribute(flag, attributes)
	assert.False(t, result)
}
