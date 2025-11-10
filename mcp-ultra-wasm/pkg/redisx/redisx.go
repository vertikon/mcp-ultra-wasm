// Package redisx provides a facade for Redis operations using go-redis.
// This package encapsulates go-redis to prevent direct dependencies and
// provides a cleaner API without .Result() chains.
package redisx

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrKeyNotFound is returned when a Redis key is not found.
var ErrKeyNotFound = errors.New("redis: key not found")

// Options is an alias for redis.Options for convenience.
type Options = redis.Options

// Client wraps redis.Client and provides a cleaner API.
type Client struct {
	client *redis.Client
}

// NewClient creates a new Redis client from the given options.
func NewClient(opts *Options) *Client {
	return &Client{
		client: redis.NewClient(opts),
	}
}

// NewClientFromURL creates a new Redis client from a connection URL.
// Format: redis://user:password@host:port/db
func NewClientFromURL(url string) (*Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return NewClient(opts), nil
}

// Ping checks the connection to Redis.
func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

// Close closes the Redis connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// Get retrieves the value for the given key.
// Returns ErrKeyNotFound if the key doesn't exist.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// Set sets the value for the given key with an optional TTL in seconds.
// Use ttlSeconds=0 for no expiration.
func (c *Client) Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error {
	var ttl time.Duration
	if ttlSeconds > 0 {
		ttl = time.Duration(ttlSeconds) * time.Second
	}
	return c.client.Set(ctx, key, value, ttl).Err()
}

// SetNX sets the value for the given key only if it doesn't exist.
// Returns true if the key was set, false if it already existed.
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, ttlSeconds int) (bool, error) {
	var ttl time.Duration
	if ttlSeconds > 0 {
		ttl = time.Duration(ttlSeconds) * time.Second
	}
	return c.client.SetNX(ctx, key, value, ttl).Result()
}

// Del deletes one or more keys.
func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

// Exists checks if one or more keys exist.
// Returns the number of keys that exist.
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.client.Exists(ctx, keys...).Result()
}

// Expire sets a timeout on a key.
func (c *Client) Expire(ctx context.Context, key string, ttlSeconds int) error {
	ttl := time.Duration(ttlSeconds) * time.Second
	return c.client.Expire(ctx, key, ttl).Err()
}

// TTL returns the remaining time to live of a key.
// Returns -1 if the key exists but has no expiration.
// Returns -2 if the key does not exist.
func (c *Client) TTL(ctx context.Context, key string) (int64, error) {
	ttl, err := c.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return int64(ttl.Seconds()), nil
}

// Incr increments the integer value of a key by one.
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// IncrBy increments the integer value of a key by the given amount.
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.IncrBy(ctx, key, value).Result()
}

// Decr decrements the integer value of a key by one.
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// DecrBy decrements the integer value of a key by the given amount.
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.client.DecrBy(ctx, key, value).Result()
}

// HGet gets the value of a hash field.
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	val, err := c.client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// HSet sets the value of a hash field.
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.client.HSet(ctx, key, values...).Err()
}

// HGetAll gets all fields and values in a hash.
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HDel deletes one or more hash fields.
func (c *Client) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}

// LPush prepends one or multiple values to a list.
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, key, values...).Err()
}

// RPush appends one or multiple values to a list.
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.RPush(ctx, key, values...).Err()
}

// LPop removes and returns the first element of a list.
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	val, err := c.client.LPop(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// RPop removes and returns the last element of a list.
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	val, err := c.client.RPop(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// LLen returns the length of a list.
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// SAdd adds one or more members to a set.
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

// SMembers returns all members of a set.
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SIsMember checks if a value is a member of a set.
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// SRem removes one or more members from a set.
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, key, members...).Err()
}

// FlushDB deletes all keys in the current database.
// Use with caution!
func (c *Client) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// FlushAll deletes all keys in all databases.
// Use with extreme caution!
func (c *Client) FlushAll(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}
