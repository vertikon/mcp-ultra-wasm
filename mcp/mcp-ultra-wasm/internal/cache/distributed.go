package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/logger"
)

// Strategy represents different caching strategies
type Strategy string

const (
	StrategyWriteThrough Strategy = "write_through"
	StrategyWriteBehind  Strategy = "write_behind"
	StrategyWriteAround  Strategy = "write_around"
	StrategyReadThrough  Strategy = "read_through"
)

// EvictionPolicy represents cache eviction policies
type EvictionPolicy string

const (
	EvictionLRU    EvictionPolicy = "lru"
	EvictionLFU    EvictionPolicy = "lfu"
	EvictionTTL    EvictionPolicy = "ttl"
	EvictionRandom EvictionPolicy = "random"
)

// Config configures the distributed cache system
type Config struct {
	// Redis Cluster Configuration
	Addrs              []string      `yaml:"addrs"`
	Password           string        `yaml:"password"`
	DB                 int           `yaml:"db"`
	PoolSize           int           `yaml:"pool_size"`
	MinIdleConns       int           `yaml:"min_idle_conns"`
	MaxConnAge         time.Duration `yaml:"max_conn_age"`
	PoolTimeout        time.Duration `yaml:"pool_timeout"`
	IdleTimeout        time.Duration `yaml:"idle_timeout"`
	IdleCheckFrequency time.Duration `yaml:"idle_check_frequency"`

	// Cache Settings
	DefaultTTL     time.Duration  `yaml:"default_ttl"`
	MaxMemory      int64          `yaml:"max_memory"`
	Strategy       Strategy       `yaml:"strategy"`
	EvictionPolicy EvictionPolicy `yaml:"eviction_policy"`

	// Consistency Settings
	ReadPreference    string `yaml:"read_preference"`   // "primary", "secondary", "nearest"
	WriteConsistency  string `yaml:"write_consistency"` // "strong", "eventual"
	ReplicationFactor int    `yaml:"replication_factor"`

	// Performance Settings
	CompressionEnabled bool   `yaml:"compression_enabled"`
	CompressionLevel   int    `yaml:"compression_level"`
	SerializationMode  string `yaml:"serialization_mode"` // "json", "msgpack", "protobuf"

	// Monitoring
	EnableMetrics      bool          `yaml:"enable_metrics"`
	EnableTracing      bool          `yaml:"enable_tracing"`
	SlowQueryThreshold time.Duration `yaml:"slow_query_threshold"`

	// Partitioning
	EnableSharding   bool   `yaml:"enable_sharding"`
	ShardingStrategy string `yaml:"sharding_strategy"` // "hash", "range", "directory"
	VirtualNodes     int    `yaml:"virtual_nodes"`

	// Circuit Breaker
	CircuitBreakerEnabled bool          `yaml:"circuit_breaker_enabled"`
	FailureThreshold      int           `yaml:"failure_threshold"`
	RecoveryTimeout       time.Duration `yaml:"recovery_timeout"`
	HalfOpenMaxRequests   int           `yaml:"half_open_max_requests"`
}

// DefaultConfig returns default cache configuration
func DefaultConfig() Config {
	return Config{
		Addrs:                 []string{"localhost:6379"},
		PoolSize:              10,
		MinIdleConns:          5,
		MaxConnAge:            time.Hour,
		PoolTimeout:           30 * time.Second,
		IdleTimeout:           5 * time.Minute,
		IdleCheckFrequency:    time.Minute,
		DefaultTTL:            time.Hour,
		MaxMemory:             1024 * 1024 * 1024, // 1GB
		Strategy:              StrategyWriteThrough,
		EvictionPolicy:        EvictionLRU,
		ReadPreference:        "primary",
		WriteConsistency:      "strong",
		ReplicationFactor:     3,
		CompressionEnabled:    true,
		CompressionLevel:      6,
		SerializationMode:     "json",
		EnableMetrics:         true,
		EnableTracing:         true,
		SlowQueryThreshold:    100 * time.Millisecond,
		EnableSharding:        true,
		ShardingStrategy:      "hash",
		VirtualNodes:          150,
		CircuitBreakerEnabled: true,
		FailureThreshold:      5,
		RecoveryTimeout:       30 * time.Second,
		HalfOpenMaxRequests:   3,
	}
}

// DistributedCache provides distributed caching capabilities
type DistributedCache struct {
	client    *redis.ClusterClient
	config    Config
	logger    *logger.Logger
	telemetry *observability.TelemetryService

	// State tracking
	mu         sync.RWMutex
	shards     []Shard
	consistent *ConsistentHash
	breaker    *CircuitBreaker
	stats      Stats

	// Background tasks
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Write-behind buffer
	writeBuffer chan WriteOperation
}

// Shard represents a cache shard
type Shard struct {
	ID       string
	Node     string
	Weight   int
	Healthy  bool
	LastSeen time.Time
}

// WriteOperation represents a write operation in write-behind mode
type WriteOperation struct {
	Key       string
	Value     interface{}
	TTL       time.Duration
	Operation string // "set", "del", "expire"
	Timestamp time.Time
}

// Stats tracks cache performance metrics
type Stats struct {
	Hits            int64         `json:"hits"`
	Misses          int64         `json:"misses"`
	Sets            int64         `json:"sets"`
	Deletes         int64         `json:"deletes"`
	Evictions       int64         `json:"evictions"`
	Errors          int64         `json:"errors"`
	TotalOperations int64         `json:"total_operations"`
	AvgLatency      time.Duration `json:"avg_latency"`
	P95Latency      time.Duration `json:"p95_latency"`
	P99Latency      time.Duration `json:"p99_latency"`
	LastReset       time.Time     `json:"last_reset"`
	MemoryUsage     int64         `json:"memory_usage"`
	ConnectionCount int           `json:"connection_count"`
}

// Entry represents a cached item with metadata
type Entry struct {
	Key         string        `json:"key"`
	Value       interface{}   `json:"value"`
	TTL         time.Duration `json:"ttl"`
	CreatedAt   time.Time     `json:"created_at"`
	ExpiresAt   time.Time     `json:"expires_at"`
	AccessCount int           `json:"access_count"`
	LastAccess  time.Time     `json:"last_access"`
	Size        int64         `json:"size"`
	Compressed  bool          `json:"compressed"`
}

// NewDistributedCache creates a new distributed cache instance
func NewDistributedCache(config Config, log *logger.Logger, telemetry *observability.TelemetryService) (*DistributedCache, error) {
	// Validate configuration
	if len(config.Addrs) == 0 {
		return nil, fmt.Errorf("at least one Redis address is required")
	}

	// Create Redis cluster client
	rdb := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        config.Addrs,
		Password:     config.Password,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		// MaxConnAge removed in v9 (managed automatically)
		PoolTimeout: config.PoolTimeout,
		// IdleTimeout removed in v9 (managed automatically)
		// IdleCheckFrequency removed in v9 (managed automatically)
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		RouteByLatency: true,
		RouteRandomly:  true,
	})

	// Test connection
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis cluster: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	cache := &DistributedCache{
		client:      rdb,
		config:      config,
		logger:      log,
		telemetry:   telemetry,
		shards:      make([]Shard, 0),
		consistent:  NewConsistentHash(config.VirtualNodes),
		breaker:     NewCircuitBreaker(config.FailureThreshold, config.RecoveryTimeout, config.HalfOpenMaxRequests),
		stats:       Stats{LastReset: time.Now()},
		ctx:         ctx,
		cancel:      cancel,
		writeBuffer: make(chan WriteOperation, 1000),
	}

	// Initialize sharding if enabled
	if config.EnableSharding {
		if err := cache.initializeSharding(ctx); err != nil {
			return nil, fmt.Errorf("failed to initialize sharding: %w", err)
		}
	}

	// Start background tasks
	cache.startBackgroundTasks()

	log.Info("Distributed cache initialized",
		"strategy", config.Strategy,
		"eviction_policy", config.EvictionPolicy,
		"sharding_enabled", config.EnableSharding,
		"compression_enabled", config.CompressionEnabled,
	)

	return cache, nil
}

// Set stores a value in the cache with the specified TTL
func (dc *DistributedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		dc.recordLatency("set", time.Since(start))
		dc.incrementCounter("sets")
	}()

	// Check circuit breaker
	if !dc.breaker.Allow() {
		dc.incrementCounter("errors")
		return fmt.Errorf("cache circuit breaker is open")
	}

	// Serialize value
	data, err := dc.serialize(value)
	if err != nil {
		dc.incrementCounter("errors")
		dc.breaker.RecordFailure()
		return fmt.Errorf("serialization failed: %w", err)
	}

	// Compress if enabled
	if dc.config.CompressionEnabled {
		data, err = dc.compress(data)
		if err != nil {
			dc.incrementCounter("errors")
			dc.breaker.RecordFailure()
			return fmt.Errorf("compression failed: %w", err)
		}
	}

	// Apply caching strategy
	switch dc.config.Strategy {
	case StrategyWriteThrough:
		err = dc.setWriteThrough(ctx, key, data, ttl)
	case StrategyWriteBehind:
		err = dc.setWriteBehind(ctx, key, value, ttl)
	case StrategyWriteAround:
		err = dc.setWriteAround(ctx, key, data, ttl)
	default:
		err = dc.setDirect(ctx, key, data, ttl)
	}

	if err != nil {
		dc.incrementCounter("errors")
		dc.breaker.RecordFailure()
		return err
	}

	dc.breaker.RecordSuccess()

	// Record metrics
	if dc.telemetry != nil && dc.config.EnableMetrics {
		dc.telemetry.RecordCounter("cache_operations_total", 1, map[string]string{
			"operation": "set",
			"strategy":  string(dc.config.Strategy),
		})
	}

	return nil
}

// Get retrieves a value from the cache
func (dc *DistributedCache) Get(ctx context.Context, key string) (interface{}, bool, error) {
	start := time.Now()
	defer func() {
		dc.recordLatency("get", time.Since(start))
	}()

	// Check circuit breaker
	if !dc.breaker.Allow() {
		dc.incrementCounter("errors")
		return nil, false, fmt.Errorf("cache circuit breaker is open")
	}

	// Apply read strategy
	data, found, err := dc.getDirect(ctx, key)
	if err != nil {
		dc.incrementCounter("errors")
		dc.incrementCounter("misses")
		dc.breaker.RecordFailure()
		return nil, false, err
	}

	if !found {
		dc.incrementCounter("misses")

		// Try read-through if configured
		if dc.config.Strategy == StrategyReadThrough {
			return dc.getReadThrough(ctx, key)
		}

		return nil, false, nil
	}

	dc.incrementCounter("hits")
	dc.breaker.RecordSuccess()

	// Decompress if needed
	if dc.config.CompressionEnabled {
		data, err = dc.decompress(data)
		if err != nil {
			dc.incrementCounter("errors")
			return nil, false, fmt.Errorf("decompression failed: %w", err)
		}
	}

	// Deserialize
	value, err := dc.deserialize(data)
	if err != nil {
		dc.incrementCounter("errors")
		return nil, false, fmt.Errorf("deserialization failed: %w", err)
	}

	// Record metrics
	if dc.telemetry != nil && dc.config.EnableMetrics {
		dc.telemetry.RecordCounter("cache_operations_total", 1, map[string]string{
			"operation": "get",
			"result":    "hit",
		})
	}

	return value, true, nil
}

// Delete removes a key from the cache
func (dc *DistributedCache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		dc.recordLatency("delete", time.Since(start))
		dc.incrementCounter("deletes")
	}()

	// Check circuit breaker
	if !dc.breaker.Allow() {
		dc.incrementCounter("errors")
		return fmt.Errorf("cache circuit breaker is open")
	}

	err := dc.client.Del(ctx, key).Err()
	if err != nil {
		dc.incrementCounter("errors")
		dc.breaker.RecordFailure()
		return fmt.Errorf("delete failed: %w", err)
	}

	dc.breaker.RecordSuccess()

	// Record metrics
	if dc.telemetry != nil && dc.config.EnableMetrics {
		dc.telemetry.RecordCounter("cache_operations_total", 1, map[string]string{
			"operation": "delete",
		})
	}

	return nil
}

// Exists checks if a key exists in the cache
func (dc *DistributedCache) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()
	defer func() {
		dc.recordLatency("exists", time.Since(start))
	}()

	count, err := dc.client.Exists(ctx, key).Result()
	if err != nil {
		dc.incrementCounter("errors")
		return false, fmt.Errorf("exists check failed: %w", err)
	}

	return count > 0, nil
}

// Expire sets the TTL for a key
func (dc *DistributedCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		dc.recordLatency("expire", time.Since(start))
	}()

	err := dc.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		dc.incrementCounter("errors")
		return fmt.Errorf("expire failed: %w", err)
	}

	return nil
}

// Clear removes all keys matching the pattern
func (dc *DistributedCache) Clear(ctx context.Context, pattern string) error {
	start := time.Now()
	defer func() {
		dc.recordLatency("clear", time.Since(start))
	}()

	// Check circuit breaker
	if !dc.breaker.Allow() {
		dc.incrementCounter("errors")
		return fmt.Errorf("cache circuit breaker is open")
	}

	// Use SCAN to find keys matching the pattern
	var cursor uint64
	var keys []string

	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = dc.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			dc.incrementCounter("errors")
			dc.breaker.RecordFailure()
			return fmt.Errorf("scan failed: %w", err)
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	// Delete all matched keys
	if len(keys) > 0 {
		err := dc.client.Del(ctx, keys...).Err()
		if err != nil {
			dc.incrementCounter("errors")
			dc.breaker.RecordFailure()
			return fmt.Errorf("delete failed: %w", err)
		}
	}

	dc.breaker.RecordSuccess()

	// Record metrics
	if dc.telemetry != nil && dc.config.EnableMetrics {
		dc.telemetry.RecordCounter("cache_operations_total", float64(len(keys)), map[string]string{
			"operation": "clear",
		})
	}

	return nil
}

// GetStats returns cache performance statistics
func (dc *DistributedCache) GetStats() Stats {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	stats := dc.stats

	// Add real-time memory usage
	if info, err := dc.client.Info(context.Background(), "memory").Result(); err == nil {
		// Parse memory usage from Redis INFO command
		for _, line := range strings.Split(info, "\r\n") {
			if strings.HasPrefix(line, "used_memory:") {
				// Extract memory usage
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					// Parse memory usage (simplified)
					stats.MemoryUsage = int64(len(parts[1])) // Placeholder
				}
			}
		}
	}

	// Add connection count
	if poolStats := dc.client.PoolStats(); poolStats != nil {
		stats.ConnectionCount = int(poolStats.TotalConns)
	}

	return stats
}

// ResetStats resets cache statistics
func (dc *DistributedCache) ResetStats() {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	dc.stats = Stats{LastReset: time.Now()}
}

// Close gracefully shuts down the cache
func (dc *DistributedCache) Close() error {
	dc.logger.Info("Shutting down distributed cache")

	// Cancel context and wait for background tasks
	dc.cancel()
	dc.wg.Wait()

	// Close write buffer
	close(dc.writeBuffer)

	// Close Redis client
	return dc.client.Close()
}

// Health check for the cache
func (dc *DistributedCache) HealthCheck(ctx context.Context) error {
	// Test basic connectivity
	if err := dc.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Check cluster health
	if err := dc.checkClusterHealth(ctx); err != nil {
		return fmt.Errorf("cluster health check failed: %w", err)
	}

	// Check circuit breaker state
	if dc.breaker.State() == CircuitBreakerOpen {
		return fmt.Errorf("circuit breaker is open")
	}

	return nil
}

// Private methods

func (dc *DistributedCache) setDirect(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = dc.config.DefaultTTL
	}
	return dc.client.Set(ctx, key, data, ttl).Err()
}

func (dc *DistributedCache) setWriteThrough(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	// In write-through, we write to cache and backing store simultaneously
	// For this example, we'll just write to cache
	return dc.setDirect(ctx, key, data, ttl)
}

func (dc *DistributedCache) setWriteBehind(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Queue write operation for background processing
	select {
	case dc.writeBuffer <- WriteOperation{
		Key:       key,
		Value:     value,
		TTL:       ttl,
		Operation: "set",
		Timestamp: time.Now(),
	}:
		return nil
	default:
		// Buffer full, fall back to direct write
		data, err := dc.serialize(value)
		if err != nil {
			return err
		}
		if dc.config.CompressionEnabled {
			data, err = dc.compress(data)
			if err != nil {
				return err
			}
		}
		return dc.setDirect(ctx, key, data, ttl)
	}
}

func (dc *DistributedCache) setWriteAround(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	// In write-around, we skip the cache and write directly to backing store
	// For this example, we'll still write to cache but with shorter TTL
	shortTTL := ttl / 4
	if shortTTL < time.Minute {
		shortTTL = time.Minute
	}
	return dc.setDirect(ctx, key, data, shortTTL)
}

func (dc *DistributedCache) getDirect(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := dc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return []byte(val), true, nil
}

func (dc *DistributedCache) getReadThrough(_ context.Context, _ string) (interface{}, bool, error) {
	// In read-through, if cache miss, we load from backing store
	// For this example, we'll return cache miss
	return nil, false, nil
}

func (dc *DistributedCache) serialize(value interface{}) ([]byte, error) {
	switch dc.config.SerializationMode {
	case "json":
		return json.Marshal(value)
	case "msgpack":
		// TODO: Implement MessagePack serialization
		return json.Marshal(value)
	case "protobuf":
		// TODO: Implement Protocol Buffers serialization
		return json.Marshal(value)
	default:
		return json.Marshal(value)
	}
}

func (dc *DistributedCache) deserialize(data []byte) (interface{}, error) {
	var value interface{}
	err := json.Unmarshal(data, &value)
	return value, err
}

func (dc *DistributedCache) compress(data []byte) ([]byte, error) {
	// TODO: Implement compression (gzip, lz4, etc.)
	return data, nil
}

func (dc *DistributedCache) decompress(data []byte) ([]byte, error) {
	// TODO: Implement decompression
	return data, nil
}

func (dc *DistributedCache) initializeSharding(ctx context.Context) error {
	// Get cluster nodes
	nodes, err := dc.client.ClusterNodes(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to get cluster nodes: %w", err)
	}

	// Parse nodes and initialize shards
	for _, line := range strings.Split(nodes, "\n") {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 8 {
			continue
		}

		nodeID := parts[0]
		nodeAddr := parts[1]

		shard := Shard{
			ID:       nodeID,
			Node:     nodeAddr,
			Weight:   1,
			Healthy:  true,
			LastSeen: time.Now(),
		}

		dc.shards = append(dc.shards, shard)
		dc.consistent.Add(nodeID, 1)
	}

	dc.logger.Info("Sharding initialized", "shards_count", len(dc.shards))
	return nil
}

func (dc *DistributedCache) checkClusterHealth(ctx context.Context) error {
	nodes, err := dc.client.ClusterNodes(ctx).Result()
	if err != nil {
		return err
	}

	healthyNodes := 0
	totalNodes := 0

	for _, line := range strings.Split(nodes, "\n") {
		if line == "" {
			continue
		}
		totalNodes++

		if strings.Contains(line, "connected") {
			healthyNodes++
		}
	}

	if healthyNodes == 0 {
		return fmt.Errorf("no healthy nodes found")
	}

	healthRatio := float64(healthyNodes) / float64(totalNodes)
	if healthRatio < 0.5 {
		return fmt.Errorf("cluster health below threshold: %.2f", healthRatio)
	}

	return nil
}

func (dc *DistributedCache) recordLatency(operation string, latency time.Duration) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Simple moving average for demonstration
	dc.stats.AvgLatency = (dc.stats.AvgLatency + latency) / 2

	// Update P95/P99 (simplified)
	if latency > dc.stats.P95Latency {
		dc.stats.P95Latency = latency
	}
	if latency > dc.stats.P99Latency {
		dc.stats.P99Latency = latency
	}

	// Record slow queries
	if latency > dc.config.SlowQueryThreshold {
		dc.logger.Warn("Slow cache operation detected",
			"operation", operation,
			"latency", latency,
			"threshold", dc.config.SlowQueryThreshold,
		)
	}
}

func (dc *DistributedCache) incrementCounter(counter string) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	switch counter {
	case "hits":
		dc.stats.Hits++
	case "misses":
		dc.stats.Misses++
	case "sets":
		dc.stats.Sets++
	case "deletes":
		dc.stats.Deletes++
	case "errors":
		dc.stats.Errors++
	}
	dc.stats.TotalOperations++
}

func (dc *DistributedCache) startBackgroundTasks() {
	// Write-behind processor
	if dc.config.Strategy == StrategyWriteBehind {
		dc.wg.Add(1)
		go dc.writeBehindProcessor()
	}

	// Stats collector
	if dc.config.EnableMetrics {
		dc.wg.Add(1)
		go dc.statsCollector()
	}

	// Health monitor
	dc.wg.Add(1)
	go dc.healthMonitor()
}

func (dc *DistributedCache) writeBehindProcessor() {
	defer dc.wg.Done()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	batch := make([]WriteOperation, 0, 100)

	for {
		select {
		case <-dc.ctx.Done():
			// Process remaining operations
			dc.processBatch(batch)
			return
		case op := <-dc.writeBuffer:
			batch = append(batch, op)
			if len(batch) >= 100 {
				dc.processBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				dc.processBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

func (dc *DistributedCache) processBatch(batch []WriteOperation) {
	if len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pipe := dc.client.Pipeline()

	for _, op := range batch {
		switch op.Operation {
		case "set":
			data, err := dc.serialize(op.Value)
			if err != nil {
				dc.logger.Error("Serialization failed in batch", "key", op.Key, "error", err)
				continue
			}

			if dc.config.CompressionEnabled {
				data, err = dc.compress(data)
				if err != nil {
					dc.logger.Error("Compression failed in batch", "key", op.Key, "error", err)
					continue
				}
			}

			pipe.Set(ctx, op.Key, data, op.TTL)
		case "del":
			pipe.Del(ctx, op.Key)
		case "expire":
			pipe.Expire(ctx, op.Key, op.TTL)
		}
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		dc.logger.Error("Batch write failed", "batch_size", len(batch), "error", err)
	} else {
		dc.logger.Debug("Batch write completed", "batch_size", len(batch))
	}
}

func (dc *DistributedCache) statsCollector() {
	defer dc.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dc.ctx.Done():
			return
		case <-ticker.C:
			dc.collectAndReportMetrics()
		}
	}
}

func (dc *DistributedCache) collectAndReportMetrics() {
	stats := dc.GetStats()

	if dc.telemetry != nil {
		dc.telemetry.RecordGauge("cache_hits_total", float64(stats.Hits), nil)
		dc.telemetry.RecordGauge("cache_misses_total", float64(stats.Misses), nil)
		dc.telemetry.RecordGauge("cache_memory_usage_bytes", float64(stats.MemoryUsage), nil)
		dc.telemetry.RecordGauge("cache_connections", float64(stats.ConnectionCount), nil)

		// Hit rate calculation
		total := stats.Hits + stats.Misses
		if total > 0 {
			hitRate := float64(stats.Hits) / float64(total) * 100
			dc.telemetry.RecordGauge("cache_hit_rate_percent", hitRate, nil)
		}
	}
}

func (dc *DistributedCache) healthMonitor() {
	defer dc.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dc.ctx.Done():
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := dc.HealthCheck(ctx); err != nil {
				dc.logger.Error("Cache health check failed", "error", err)
			}
			cancel()
		}
	}
}
