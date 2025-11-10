package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheRepository implements domain.CacheRepository using Redis
type CacheRepository struct {
	client *redis.Client
}

// NewCacheRepository creates a new Redis cache repository
func NewCacheRepository(client *redis.Client) *CacheRepository {
	return &CacheRepository{client: client}
}

// Set stores a value in cache with TTL
func (r *CacheRepository) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshaling value: %w", err)
	}

	expiration := time.Duration(ttl) * time.Second
	if ttl <= 0 {
		expiration = 0 // No expiration
	}

	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("setting cache value: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (r *CacheRepository) Get(ctx context.Context, key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}
	if err != nil {
		return "", fmt.Errorf("getting cache value: %w", err)
	}

	return result, nil
}

// Delete removes a key from cache
func (r *CacheRepository) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("deleting cache key: %w", err)
	}

	return nil
}

// Exists checks if a key exists in cache
func (r *CacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("checking cache key existence: %w", err)
	}

	return result > 0, nil
}

// Increment increments a counter
func (r *CacheRepository) Increment(ctx context.Context, key string) (int64, error) {
	result, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("incrementing counter: %w", err)
	}

	return result, nil
}

// SetNX sets a value only if the key doesn't exist (atomic operation)
func (r *CacheRepository) SetNX(ctx context.Context, key string, value interface{}, ttl int) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("marshaling value: %w", err)
	}

	expiration := time.Duration(ttl) * time.Second
	if ttl <= 0 {
		expiration = 0 // No expiration
	}

	result, err := r.client.SetNX(ctx, key, data, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("setting cache value with NX: %w", err)
	}

	return result, nil
}

// GetJSON retrieves and unmarshals a JSON value from cache
func (r *CacheRepository) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(data), dest)
	if err != nil {
		return fmt.Errorf("unmarshaling cached value: %w", err)
	}

	return nil
}

// SetWithExpiry sets a value with a specific expiry time
func (r *CacheRepository) SetWithExpiry(ctx context.Context, key string, value interface{}, expiry time.Time) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshaling value: %w", err)
	}

	err = r.client.SetEx(ctx, key, data, time.Until(expiry)).Err()
	if err != nil {
		return fmt.Errorf("setting cache value with expiry: %w", err)
	}

	return nil
}

// GetTTL returns the remaining time-to-live of a key
func (r *CacheRepository) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	result, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("getting TTL: %w", err)
	}

	return result, nil
}

// FlushAll removes all keys (use with caution)
func (r *CacheRepository) FlushAll(ctx context.Context) error {
	err := r.client.FlushAll(ctx).Err()
	if err != nil {
		return fmt.Errorf("flushing all cache: %w", err)
	}

	return nil
}
