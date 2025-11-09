package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
)

// NewClient creates a new Redis client
func NewClient(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	return client
}

// Ping tests Redis connection
func Ping(client *redis.Client) error {
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("pinging Redis: %w", err)
	}
	return nil
}
