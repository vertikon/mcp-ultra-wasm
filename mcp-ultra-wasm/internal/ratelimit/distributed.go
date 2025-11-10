package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/logger"
)

// Algorithm represents different rate limiting algorithms
type Algorithm string

const (
	AlgorithmTokenBucket   Algorithm = "token_bucket"
	AlgorithmLeakyBucket   Algorithm = "leaky_bucket"
	AlgorithmFixedWindow   Algorithm = "fixed_window"
	AlgorithmSlidingWindow Algorithm = "sliding_window"
	AlgorithmConcurrency   Algorithm = "concurrency"
	AlgorithmAdaptive      Algorithm = "adaptive"
)

// DistributedRateLimiter provides distributed rate limiting capabilities
type DistributedRateLimiter struct {
	client    redis.Cmdable
	config    Config
	logger    *logger.Logger
	telemetry *observability.TelemetryService

	// State
	limiters map[string]Limiter
	scripts  *LuaScripts

	// Background tasks
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Config configures the distributed rate limiter
type Config struct {
	// Redis configuration
	RedisKeyPrefix string        `yaml:"redis_key_prefix"`
	RedisKeyTTL    time.Duration `yaml:"redis_key_ttl"`

	// Default limits
	DefaultAlgorithm Algorithm     `yaml:"default_algorithm"`
	DefaultLimit     int64         `yaml:"default_limit"`
	DefaultWindow    time.Duration `yaml:"default_window"`

	// Behavior
	AllowBursts          bool `yaml:"allow_bursts"`
	SkipFailedLimits     bool `yaml:"skip_failed_limits"`
	SkipSuccessfulLimits bool `yaml:"skip_successful_limits"`

	// Performance
	MaxConcurrency    int           `yaml:"max_concurrency"`
	LocalCacheEnabled bool          `yaml:"local_cache_enabled"`
	LocalCacheTTL     time.Duration `yaml:"local_cache_ttl"`

	// Monitoring
	EnableMetrics bool `yaml:"enable_metrics"`
	EnableTracing bool `yaml:"enable_tracing"`

	// Adaptive behavior
	AdaptiveEnabled   bool          `yaml:"adaptive_enabled"`
	AdaptiveWindow    time.Duration `yaml:"adaptive_window"`
	AdaptiveThreshold float64       `yaml:"adaptive_threshold"`
}

// Rule defines a rate limiting rule
type Rule struct {
	ID          string        `json:"id" yaml:"id"`
	Name        string        `json:"name" yaml:"name"`
	Description string        `json:"description" yaml:"description"`
	Algorithm   Algorithm     `json:"algorithm" yaml:"algorithm"`
	Limit       int64         `json:"limit" yaml:"limit"`
	Window      time.Duration `json:"window" yaml:"window"`

	// Key generation
	KeyTemplate string   `json:"key_template" yaml:"key_template"`
	KeyFields   []string `json:"key_fields" yaml:"key_fields"`

	// Conditions
	Conditions []Condition `json:"conditions" yaml:"conditions"`

	// Behavior
	Priority int  `json:"priority" yaml:"priority"`
	Enabled  bool `json:"enabled" yaml:"enabled"`
	FailOpen bool `json:"fail_open" yaml:"fail_open"`

	// Adaptive settings
	Adaptive bool  `json:"adaptive" yaml:"adaptive"`
	MinLimit int64 `json:"min_limit" yaml:"min_limit"`
	MaxLimit int64 `json:"max_limit" yaml:"max_limit"`

	// Metadata
	Tags      []string  `json:"tags" yaml:"tags"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
}

// Condition represents a condition for rule application
type Condition struct {
	Field    string      `json:"field" yaml:"field"`
	Operator string      `json:"operator" yaml:"operator"`
	Value    interface{} `json:"value" yaml:"value"`
	Type     string      `json:"type" yaml:"type"`
}

// Request represents a rate limiting request
type Request struct {
	Key        string                 `json:"key"`
	UserID     string                 `json:"user_id,omitempty"`
	IP         string                 `json:"ip,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Response represents a rate limiting response
type Response struct {
	Allowed    bool          `json:"allowed"`
	Limit      int64         `json:"limit"`
	Remaining  int64         `json:"remaining"`
	ResetTime  time.Time     `json:"reset_time"`
	RetryAfter time.Duration `json:"retry_after,omitempty"`

	// Additional info
	Algorithm Algorithm     `json:"algorithm"`
	RuleID    string        `json:"rule_id,omitempty"`
	RuleName  string        `json:"rule_name,omitempty"`
	Window    time.Duration `json:"window"`

	// Metadata
	RequestID      string        `json:"request_id,omitempty"`
	ProcessingTime time.Duration `json:"processing_time"`
	FromCache      bool          `json:"from_cache"`
}

// Limiter interface for different rate limiting algorithms
type Limiter interface {
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (*Response, error)
	Reset(ctx context.Context, key string) error
	GetUsage(ctx context.Context, key string) (int64, error)
}

// TokenBucketLimiter implements token bucket algorithm
type TokenBucketLimiter struct {
	client redis.Cmdable
	script string
}

// SlidingWindowLimiter implements sliding window algorithm
type SlidingWindowLimiter struct {
	client redis.Cmdable
	script string
}

// AdaptiveLimiter implements adaptive rate limiting
type AdaptiveLimiter struct {
	client redis.Cmdable
	config Config
	logger *logger.Logger

	mu            sync.RWMutex
	adaptiveState map[string]*AdaptiveState
}

// AdaptiveState tracks adaptive rate limiting state
type AdaptiveState struct {
	CurrentLimit   int64     `json:"current_limit"`
	BaseLimit      int64     `json:"base_limit"`
	MinLimit       int64     `json:"min_limit"`
	MaxLimit       int64     `json:"max_limit"`
	SuccessCount   int64     `json:"success_count"`
	ErrorCount     int64     `json:"error_count"`
	LastAdjustment time.Time `json:"last_adjustment"`
	AdjustmentRate float64   `json:"adjustment_rate"`
}

// LuaScripts contains Lua scripts for atomic operations
type LuaScripts struct {
	tokenBucket   *redis.Script
	slidingWindow *redis.Script
	fixedWindow   *redis.Script
	leakyBucket   *redis.Script
	concurrency   *redis.Script
}

// DefaultConfig returns default rate limiter configuration
func DefaultConfig() Config {
	return Config{
		RedisKeyPrefix:       "ratelimit:",
		RedisKeyTTL:          time.Hour,
		DefaultAlgorithm:     AlgorithmSlidingWindow,
		DefaultLimit:         1000,
		DefaultWindow:        time.Minute,
		AllowBursts:          true,
		SkipFailedLimits:     false,
		SkipSuccessfulLimits: false,
		MaxConcurrency:       100,
		LocalCacheEnabled:    true,
		LocalCacheTTL:        time.Second,
		EnableMetrics:        true,
		EnableTracing:        true,
		AdaptiveEnabled:      false,
		AdaptiveWindow:       5 * time.Minute,
		AdaptiveThreshold:    0.8,
	}
}

// NewDistributedRateLimiter creates a new distributed rate limiter
func NewDistributedRateLimiter(client redis.Cmdable, config Config, logger *logger.Logger, telemetry *observability.TelemetryService) (*DistributedRateLimiter, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scripts := &LuaScripts{
		tokenBucket:   redis.NewScript(tokenBucketScript),
		slidingWindow: redis.NewScript(slidingWindowScript),
		fixedWindow:   redis.NewScript(fixedWindowScript),
		leakyBucket:   redis.NewScript(leakyBucketScript),
		concurrency:   redis.NewScript(concurrencyScript),
	}

	limiter := &DistributedRateLimiter{
		client:    client,
		config:    config,
		logger:    logger,
		telemetry: telemetry,
		limiters:  make(map[string]Limiter),
		scripts:   scripts,
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initialize algorithm-specific limiters
	limiter.limiters[string(AlgorithmTokenBucket)] = &TokenBucketLimiter{
		client: client,
		script: tokenBucketScript,
	}

	limiter.limiters[string(AlgorithmSlidingWindow)] = &SlidingWindowLimiter{
		client: client,
		script: slidingWindowScript,
	}

	limiter.limiters[string(AlgorithmAdaptive)] = &AdaptiveLimiter{
		client:        client,
		config:        config,
		logger:        logger,
		adaptiveState: make(map[string]*AdaptiveState),
	}

	// Start background tasks
	limiter.startBackgroundTasks()

	logger.Info("Distributed rate limiter initialized",
		"default_algorithm", config.DefaultAlgorithm,
		"default_limit", config.DefaultLimit,
		"default_window", config.DefaultWindow,
		"adaptive_enabled", config.AdaptiveEnabled,
	)

	return limiter, nil
}

// Allow checks if a request should be allowed
func (drl *DistributedRateLimiter) Allow(ctx context.Context, request Request) (*Response, error) {
	start := time.Now()

	// Use default values if not specified
	key := request.Key
	if key == "" {
		key = drl.generateKey(request)
	}

	// Get appropriate limiter
	algorithm := drl.config.DefaultAlgorithm
	limiter, exists := drl.limiters[string(algorithm)]
	if !exists {
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	// Apply rate limiting
	response, err := limiter.Allow(ctx, key, drl.config.DefaultLimit, drl.config.DefaultWindow)
	if err != nil {
		drl.recordMetrics("error", algorithm, key, 0)
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	response.Algorithm = algorithm
	response.ProcessingTime = time.Since(start)

	// Record metrics
	status := "allowed"
	if !response.Allowed {
		status = "denied"
	}
	drl.recordMetrics(status, algorithm, key, response.Remaining)

	return response, nil
}

// AllowWithRule checks if a request should be allowed using a specific rule
func (drl *DistributedRateLimiter) AllowWithRule(ctx context.Context, request Request, rule Rule) (*Response, error) {
	start := time.Now()

	// Check if rule conditions match
	if !drl.evaluateConditions(rule.Conditions, request) {
		return &Response{
			Allowed:        true,
			Limit:          rule.Limit,
			Remaining:      rule.Limit,
			ResetTime:      time.Now().Add(rule.Window),
			Algorithm:      rule.Algorithm,
			RuleID:         rule.ID,
			RuleName:       rule.Name,
			Window:         rule.Window,
			ProcessingTime: time.Since(start),
		}, nil
	}

	// Generate key based on rule template
	key := drl.generateRuleKey(rule, request)

	// Get appropriate limiter
	limiter, exists := drl.limiters[string(rule.Algorithm)]
	if !exists {
		if rule.FailOpen {
			return &Response{Allowed: true}, nil
		}
		return nil, fmt.Errorf("unsupported algorithm: %s", rule.Algorithm)
	}

	// Apply adaptive limits if enabled
	limit := rule.Limit
	if rule.Adaptive && drl.config.AdaptiveEnabled {
		limit = drl.getAdaptiveLimit(key, rule)
	}

	// Apply rate limiting
	response, err := limiter.Allow(ctx, key, limit, rule.Window)
	if err != nil {
		if rule.FailOpen {
			return &Response{Allowed: true}, nil
		}
		return nil, fmt.Errorf("rate limit check failed: %w", err)
	}

	response.Algorithm = rule.Algorithm
	response.RuleID = rule.ID
	response.RuleName = rule.Name
	response.Window = rule.Window
	response.ProcessingTime = time.Since(start)

	// Update adaptive state
	if rule.Adaptive && drl.config.AdaptiveEnabled {
		drl.updateAdaptiveState(key, rule, response.Allowed)
	}

	// Record metrics
	status := "allowed"
	if !response.Allowed {
		status = "denied"
	}
	drl.recordMetrics(status, rule.Algorithm, key, response.Remaining)

	return response, nil
}

// Reset resets the rate limit for a key
func (drl *DistributedRateLimiter) Reset(ctx context.Context, key string) error {
	for _, limiter := range drl.limiters {
		if err := limiter.Reset(ctx, key); err != nil {
			drl.logger.Error("Failed to reset rate limit", "key", key, "error", err)
			return err
		}
	}
	return nil
}

// GetUsage returns current usage for a key
func (drl *DistributedRateLimiter) GetUsage(ctx context.Context, key string, algorithm Algorithm) (int64, error) {
	limiter, exists := drl.limiters[string(algorithm)]
	if !exists {
		return 0, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	return limiter.GetUsage(ctx, key)
}

// GetStats returns rate limiting statistics
func (drl *DistributedRateLimiter) GetStats() Stats {
	// Implementation would collect stats from Redis and internal state
	return Stats{
		TotalRequests:   0,
		AllowedRequests: 0,
		DeniedRequests:  0,
		ErrorRate:       0,
		AvgLatency:      0,
		LastReset:       time.Now(),
	}
}

// Close gracefully shuts down the rate limiter
func (drl *DistributedRateLimiter) Close() error {
	drl.logger.Info("Shutting down distributed rate limiter")

	drl.cancel()
	drl.wg.Wait()

	return nil
}

// Stats contains rate limiting statistics
type Stats struct {
	TotalRequests   int64         `json:"total_requests"`
	AllowedRequests int64         `json:"allowed_requests"`
	DeniedRequests  int64         `json:"denied_requests"`
	ErrorRate       float64       `json:"error_rate"`
	AvgLatency      time.Duration `json:"avg_latency"`
	LastReset       time.Time     `json:"last_reset"`
}

// Private methods

func (drl *DistributedRateLimiter) generateKey(request Request) string {
	// Simple key generation based on available fields
	if request.UserID != "" {
		return fmt.Sprintf("%suser:%s", drl.config.RedisKeyPrefix, request.UserID)
	}
	if request.IP != "" {
		return fmt.Sprintf("%sip:%s", drl.config.RedisKeyPrefix, request.IP)
	}
	return fmt.Sprintf("%sdefault", drl.config.RedisKeyPrefix)
}

func (drl *DistributedRateLimiter) generateRuleKey(rule Rule, request Request) string {
	key := rule.KeyTemplate

	// Replace template variables
	for _, field := range rule.KeyFields {
		value := drl.getRequestField(request, field)
		key = fmt.Sprintf("%s:%s", key, value)
	}

	return fmt.Sprintf("%s%s", drl.config.RedisKeyPrefix, key)
}

func (drl *DistributedRateLimiter) getRequestField(request Request, field string) string {
	switch field {
	case "user_id":
		return request.UserID
	case "ip":
		return request.IP
	case "path":
		return request.Path
	case "method":
		return request.Method
	default:
		if value, exists := request.Attributes[field]; exists {
			return fmt.Sprintf("%v", value)
		}
		return ""
	}
}

func (drl *DistributedRateLimiter) evaluateConditions(conditions []Condition, request Request) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, condition := range conditions {
		if !drl.evaluateCondition(condition, request) {
			return false
		}
	}

	return true
}

func (drl *DistributedRateLimiter) evaluateCondition(condition Condition, request Request) bool {
	requestValue := drl.getRequestField(request, condition.Field)

	switch condition.Operator {
	case "equals":
		return requestValue == fmt.Sprintf("%v", condition.Value)
	case "not_equals":
		return requestValue != fmt.Sprintf("%v", condition.Value)
	case "contains":
		return len(requestValue) > 0 && len(fmt.Sprintf("%v", condition.Value)) > 0
	case "starts_with":
		return len(requestValue) > 0 && fmt.Sprintf("%v", condition.Value) != ""
	case "ends_with":
		return len(requestValue) > 0 && fmt.Sprintf("%v", condition.Value) != ""
	default:
		return false
	}
}

func (drl *DistributedRateLimiter) getAdaptiveLimit(key string, rule Rule) int64 {
	if adaptive, exists := drl.limiters[string(AlgorithmAdaptive)]; exists {
		if adaptiveLimiter, ok := adaptive.(*AdaptiveLimiter); ok {
			return adaptiveLimiter.getAdaptiveLimit(key, rule)
		}
	}
	return rule.Limit
}

func (drl *DistributedRateLimiter) updateAdaptiveState(key string, rule Rule, allowed bool) {
	if adaptive, exists := drl.limiters[string(AlgorithmAdaptive)]; exists {
		if adaptiveLimiter, ok := adaptive.(*AdaptiveLimiter); ok {
			adaptiveLimiter.updateState(key, rule, allowed)
		}
	}
}

func (drl *DistributedRateLimiter) recordMetrics(status string, algorithm Algorithm, _ string, remaining int64) {
	if drl.telemetry != nil && drl.config.EnableMetrics {
		drl.telemetry.RecordCounter("rate_limit_requests_total", 1, map[string]string{
			"status":    status,
			"algorithm": string(algorithm),
		})

		drl.telemetry.RecordGauge("rate_limit_remaining", float64(remaining), map[string]string{
			"algorithm": string(algorithm),
		})
	}
}

func (drl *DistributedRateLimiter) startBackgroundTasks() {
	// Adaptive adjustment task
	if drl.config.AdaptiveEnabled {
		drl.wg.Add(1)
		go drl.adaptiveAdjustmentTask()
	}

	// Cleanup task
	drl.wg.Add(1)
	go drl.cleanupTask()
}

func (drl *DistributedRateLimiter) adaptiveAdjustmentTask() {
	defer drl.wg.Done()

	ticker := time.NewTicker(drl.config.AdaptiveWindow / 4)
	defer ticker.Stop()

	for {
		select {
		case <-drl.ctx.Done():
			return
		case <-ticker.C:
			drl.performAdaptiveAdjustments()
		}
	}
}

func (drl *DistributedRateLimiter) performAdaptiveAdjustments() {
	if adaptive, exists := drl.limiters[string(AlgorithmAdaptive)]; exists {
		if adaptiveLimiter, ok := adaptive.(*AdaptiveLimiter); ok {
			adaptiveLimiter.performAdjustments()
		}
	}
}

func (drl *DistributedRateLimiter) cleanupTask() {
	defer drl.wg.Done()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-drl.ctx.Done():
			return
		case <-ticker.C:
			drl.performCleanup()
		}
	}
}

func (drl *DistributedRateLimiter) performCleanup() {
	// Clean up expired keys and adaptive state
	drl.logger.Debug("Performing rate limiter cleanup")
}

// TokenBucketLimiter implementation

func (tbl *TokenBucketLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (*Response, error) {
	now := time.Now()
	result, err := tbl.client.Eval(ctx, tbl.script, []string{key}, limit, window.Seconds(), now.Unix()).Result()
	if err != nil {
		return nil, err
	}

	values := result.([]interface{})
	allowed := values[0].(int64) == 1
	remaining := values[1].(int64)
	resetTime := time.Unix(values[2].(int64), 0)

	response := &Response{
		Allowed:   allowed,
		Limit:     limit,
		Remaining: remaining,
		ResetTime: resetTime,
		Window:    window,
	}

	if !allowed {
		response.RetryAfter = resetTime.Sub(now)
	}

	return response, nil
}

func (tbl *TokenBucketLimiter) Reset(ctx context.Context, key string) error {
	return tbl.client.Del(ctx, key).Err()
}

func (tbl *TokenBucketLimiter) GetUsage(ctx context.Context, key string) (int64, error) {
	result, err := tbl.client.HGet(ctx, key, "tokens").Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	tokens, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return 0, err
	}

	return tokens, nil
}

// SlidingWindowLimiter implementation

func (swl *SlidingWindowLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (*Response, error) {
	now := time.Now()
	result, err := swl.client.Eval(ctx, swl.script, []string{key}, limit, window.Milliseconds(), now.UnixNano()/1000000).Result()
	if err != nil {
		return nil, err
	}

	values := result.([]interface{})
	allowed := values[0].(int64) == 1
	count := values[1].(int64)
	remaining := limit - count
	resetTime := now.Add(window)

	response := &Response{
		Allowed:   allowed,
		Limit:     limit,
		Remaining: remaining,
		ResetTime: resetTime,
		Window:    window,
	}

	if !allowed {
		response.RetryAfter = window
	}

	return response, nil
}

func (swl *SlidingWindowLimiter) Reset(ctx context.Context, key string) error {
	return swl.client.Del(ctx, key).Err()
}

func (swl *SlidingWindowLimiter) GetUsage(ctx context.Context, key string) (int64, error) {
	now := time.Now().UnixNano() / 1000000
	count, err := swl.client.ZCount(ctx, key, fmt.Sprintf("%d", now-60000), "+inf").Result()
	return count, err
}

// AdaptiveLimiter implementation

func (al *AdaptiveLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (*Response, error) {
	// Use sliding window as base algorithm
	swl := &SlidingWindowLimiter{
		client: al.client,
		script: slidingWindowScript,
	}

	return swl.Allow(ctx, key, limit, window)
}

func (al *AdaptiveLimiter) Reset(ctx context.Context, key string) error {
	al.mu.Lock()
	delete(al.adaptiveState, key)
	al.mu.Unlock()

	return al.client.Del(ctx, key).Err()
}

func (al *AdaptiveLimiter) GetUsage(ctx context.Context, key string) (int64, error) {
	swl := &SlidingWindowLimiter{client: al.client}
	return swl.GetUsage(ctx, key)
}

func (al *AdaptiveLimiter) getAdaptiveLimit(key string, rule Rule) int64 {
	al.mu.RLock()
	state, exists := al.adaptiveState[key]
	al.mu.RUnlock()

	if !exists {
		state = &AdaptiveState{
			CurrentLimit:   rule.Limit,
			BaseLimit:      rule.Limit,
			MinLimit:       rule.MinLimit,
			MaxLimit:       rule.MaxLimit,
			AdjustmentRate: 0.1, // 10% adjustments
		}

		al.mu.Lock()
		al.adaptiveState[key] = state
		al.mu.Unlock()
	}

	return state.CurrentLimit
}

func (al *AdaptiveLimiter) updateState(key string, _ Rule, allowed bool) {
	al.mu.Lock()
	defer al.mu.Unlock()

	state, exists := al.adaptiveState[key]
	if !exists {
		return
	}

	if allowed {
		state.SuccessCount++
	} else {
		state.ErrorCount++
	}
}

func (al *AdaptiveLimiter) performAdjustments() {
	al.mu.Lock()
	defer al.mu.Unlock()

	now := time.Now()

	for key, state := range al.adaptiveState {
		if now.Sub(state.LastAdjustment) < al.config.AdaptiveWindow {
			continue
		}

		total := state.SuccessCount + state.ErrorCount
		if total == 0 {
			continue
		}

		errorRate := float64(state.ErrorCount) / float64(total)

		// Adjust limits based on error rate
		if errorRate > al.config.AdaptiveThreshold {
			// High error rate - decrease limit
			newLimit := int64(float64(state.CurrentLimit) * (1 - state.AdjustmentRate))
			if newLimit >= state.MinLimit {
				state.CurrentLimit = newLimit
			}
		} else if errorRate < al.config.AdaptiveThreshold/2 {
			// Low error rate - increase limit
			newLimit := int64(float64(state.CurrentLimit) * (1 + state.AdjustmentRate))
			if newLimit <= state.MaxLimit {
				state.CurrentLimit = newLimit
			}
		}

		// Reset counters
		state.SuccessCount = 0
		state.ErrorCount = 0
		state.LastAdjustment = now

		al.logger.Debug("Adaptive limit adjusted",
			"key", key,
			"new_limit", state.CurrentLimit,
			"error_rate", errorRate,
		)
	}
}

// Lua Scripts for atomic operations

const tokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
local tokens = tonumber(bucket[1]) or capacity
local last_refill = tonumber(bucket[2]) or now

-- Calculate tokens to add based on time elapsed
local elapsed = math.max(0, now - last_refill)
local tokens_to_add = math.floor(elapsed * capacity / window)
tokens = math.min(capacity, tokens + tokens_to_add)

local allowed = 0
local reset_time = now + window

if tokens > 0 then
    allowed = 1
    tokens = tokens - 1
end

-- Update bucket state
redis.call('HMSET', key, 'tokens', tokens, 'last_refill', now)
redis.call('EXPIRE', key, window + 1)

return {allowed, tokens, reset_time}
`

const slidingWindowScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

-- Remove expired entries
local expired_before = now - window
redis.call('ZREMRANGEBYSCORE', key, 0, expired_before)

-- Count current entries
local current = redis.call('ZCARD', key)

local allowed = 0
if current < limit then
    allowed = 1
    -- Add current request
    redis.call('ZADD', key, now, now .. math.random())
    current = current + 1
end

-- Set expiration
redis.call('EXPIRE', key, math.ceil(window / 1000) + 1)

return {allowed, current}
`

const fixedWindowScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

-- Create window-specific key
local window_start = math.floor(now / window) * window
local window_key = key .. ':' .. window_start

local current = redis.call('GET', window_key) or 0
current = tonumber(current)

local allowed = 0
if current < limit then
    allowed = 1
    current = redis.call('INCR', window_key)
    redis.call('EXPIRE', window_key, window + 1)
end

local reset_time = window_start + window

return {allowed, current, reset_time}
`

const leakyBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local leak_rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

local bucket = redis.call('HMGET', key, 'volume', 'last_leak')
local volume = tonumber(bucket[1]) or 0
local last_leak = tonumber(bucket[2]) or now

-- Calculate leaked volume
local elapsed = math.max(0, now - last_leak)
local leaked = elapsed * leak_rate
volume = math.max(0, volume - leaked)

local allowed = 0
if volume < capacity then
    allowed = 1
    volume = volume + 1
end

-- Update bucket state
redis.call('HMSET', key, 'volume', volume, 'last_leak', now)
redis.call('EXPIRE', key, capacity / leak_rate + 1)

local retry_after = 0
if allowed == 0 then
    retry_after = (volume - capacity + 1) / leak_rate
end

return {allowed, capacity - volume, retry_after}
`

const concurrencyScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local ttl = tonumber(ARGV[2])

local current = redis.call('GET', key) or 0
current = tonumber(current)

local allowed = 0
if current < limit then
    allowed = 1
    current = redis.call('INCR', key)
    redis.call('EXPIRE', key, ttl)
end

return {allowed, current, limit - current}
`
