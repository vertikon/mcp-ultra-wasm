package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/logger"
)

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()
	zapLog := zaptest.NewLogger(t)
	return logger.FromZap(zapLog)
}

func createTestDistributedCache(t *testing.T) (*DistributedCache, *miniredis.Miniredis) {
	s, err := miniredis.Run()
	require.NoError(t, err)

	config := Config{
		Addrs:                 []string{s.Addr()},
		Password:              "",
		DB:                    0,
		PoolSize:              10,
		MinIdleConns:          5,
		MaxConnAge:            30 * time.Minute,
		PoolTimeout:           5 * time.Second,
		IdleTimeout:           10 * time.Minute,
		IdleCheckFrequency:    time.Minute,
		DefaultTTL:            5 * time.Minute,
		MaxMemory:             1024 * 1024 * 1024,
		Strategy:              StrategyWriteThrough,
		EvictionPolicy:        EvictionLRU,
		ReadPreference:        "primary",
		WriteConsistency:      "strong",
		ReplicationFactor:     3,
		CompressionEnabled:    false,
		CompressionLevel:      6,
		SerializationMode:     "json",
		EnableMetrics:         true,
		EnableTracing:         false,
		SlowQueryThreshold:    100 * time.Millisecond,
		EnableSharding:        false,
		ShardingStrategy:      "hash",
		VirtualNodes:          150,
		CircuitBreakerEnabled: false,
		FailureThreshold:      5,
		RecoveryTimeout:       30 * time.Second,
		HalfOpenMaxRequests:   3,
	}

	testLog := newTestLogger(t)
	cache, err := NewDistributedCache(config, testLog, nil)
	require.NoError(t, err)

	return cache, s
}

func TestDistributedCache_SetAndGet(t *testing.T) {
	cache, miniredis := createTestDistributedCache(t)
	defer miniredis.Close()

	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	// Test Set
	err := cache.Set(ctx, key, value, time.Minute)
	assert.NoError(t, err)

	// Test Get
	resultVal, found, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, resultVal)
}

func TestDistributedCache_SetWithTTL(t *testing.T) {
	cache, miniredis := createTestDistributedCache(t)
	defer miniredis.Close()

	ctx := context.Background()
	key := "ttl_key"
	value := "ttl_value"
	ttl := 100 * time.Millisecond

	// Set value with short TTL
	err := cache.Set(ctx, key, value, ttl)
	assert.NoError(t, err)

	// Verify value exists initially
	resultVal, found, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, resultVal)

	// Wait for TTL to expire
	time.Sleep(150 * time.Millisecond)

	// Verify value no longer exists
	_, found, err = cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestDistributedCache_Delete(t *testing.T) {
	cache, miniredis := createTestDistributedCache(t)
	defer miniredis.Close()

	ctx := context.Background()
	key := "delete_key"
	value := "delete_value"

	// Set value
	err := cache.Set(ctx, key, value, time.Minute)
	assert.NoError(t, err)

	// Verify value exists
	resultVal, found, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, resultVal)

	// Delete value
	err = cache.Delete(ctx, key)
	assert.NoError(t, err)

	// Verify value no longer exists
	_, found, err = cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestDistributedCache_Clear(t *testing.T) {
	cache, miniredis := createTestDistributedCache(t)
	defer miniredis.Close()

	ctx := context.Background()

	// Set multiple values
	keys := []string{"clear_key1", "clear_key2", "clear_key3"}
	for _, key := range keys {
		err := cache.Set(ctx, key, "value", time.Minute)
		assert.NoError(t, err)
	}

	// Clear all keys matching pattern
	err := cache.Clear(ctx, "clear_*")
	assert.NoError(t, err)

	// Verify all keys are deleted
	for _, key := range keys {
		_, found, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.False(t, found)
	}
}

func TestDistributedCache_GetNonExistentKey(t *testing.T) {
	cache, miniredis := createTestDistributedCache(t)
	defer miniredis.Close()

	ctx := context.Background()
	key := "non_existent_key"

	_, found, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestDistributedCache_SetComplexObject(t *testing.T) {
	cache, miniredis := createTestDistributedCache(t)
	defer miniredis.Close()

	ctx := context.Background()
	key := "complex_object"

	type ComplexObject struct {
		ID     int      `json:"id"`
		Name   string   `json:"name"`
		Tags   []string `json:"tags"`
		Active bool     `json:"active"`
	}

	originalObject := ComplexObject{
		ID:     123,
		Name:   "Test Object",
		Tags:   []string{"tag1", "tag2", "tag3"},
		Active: true,
	}

	// Set complex object
	err := cache.Set(ctx, key, originalObject, time.Minute)
	assert.NoError(t, err)

	// Get complex object
	resultVal, found, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.True(t, found)

	// Convert result to ComplexObject
	resultMap, ok := resultVal.(map[string]interface{})
	assert.True(t, ok)

	retrievedObject := ComplexObject{
		ID:     int(resultMap["id"].(float64)),
		Name:   resultMap["name"].(string),
		Active: resultMap["active"].(bool),
	}

	// Convert tags
	tagsInterface := resultMap["tags"].([]interface{})
	tags := make([]string, len(tagsInterface))
	for i, tag := range tagsInterface {
		tags[i] = tag.(string)
	}
	retrievedObject.Tags = tags

	assert.Equal(t, originalObject, retrievedObject)
}

func TestDistributedCache_ConcurrentOperations(t *testing.T) {
	cache, miniredis := createTestDistributedCache(t)
	defer miniredis.Close()

	ctx := context.Background()
	numOperations := 100

	// Run concurrent set operations
	done := make(chan bool, numOperations)
	for i := 0; i < numOperations; i++ {
		go func(i int) {
			key := fmt.Sprintf("concurrent_key_%d", i)
			value := fmt.Sprintf("concurrent_value_%d", i)
			err := cache.Set(ctx, key, value, time.Minute)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numOperations; i++ {
		<-done
	}

	// Verify all values were set correctly
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("concurrent_key_%d", i)
		expectedValue := fmt.Sprintf("concurrent_value_%d", i)

		actualValue, found, err := cache.Get(ctx, key)
		assert.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, expectedValue, actualValue)
	}
}

func TestDistributedCache_Namespace(t *testing.T) {
	cache, miniredis := createTestDistributedCache(t)
	defer miniredis.Close()

	ctx := context.Background()
	key := "namespace_key"
	value := "namespace_value"

	// Set value (should be prefixed with namespace)
	err := cache.Set(ctx, key, value, time.Minute)
	assert.NoError(t, err)

	// Verify the key exists with the namespace prefix in Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: miniredis.Addr(),
	})
	defer func() { _ = redisClient.Close() }()

	namespacedKey := "test:" + key
	exists := redisClient.Exists(ctx, namespacedKey)
	assert.Equal(t, int64(1), exists.Val())

	// Get value through cache (should handle namespace automatically)
	resultVal, found, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, resultVal)
}

func TestCacheStrategy_WriteThrough(t *testing.T) {
	cache, miniredis := createTestDistributedCache(t)
	defer miniredis.Close()

	// WriteThrough strategy should write to both cache and backing store
	// For this test, we'll just verify the cache behavior
	ctx := context.Background()
	key := "write_through_key"
	value := "write_through_value"

	err := cache.Set(ctx, key, value, time.Minute)
	assert.NoError(t, err)

	resultVal, found, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, resultVal)
}

func TestDistributedCache_InvalidKey(t *testing.T) {
	cache, miniredis := createTestDistributedCache(t)
	defer miniredis.Close()

	ctx := context.Background()

	// Test with empty key
	err := cache.Set(ctx, "", "value", time.Minute)
	assert.Error(t, err)

	// Test with very long key (exceeding MaxKeySize)
	longKey := string(make([]byte, 2000))
	err = cache.Set(ctx, longKey, "value", time.Minute)
	assert.Error(t, err)
}
