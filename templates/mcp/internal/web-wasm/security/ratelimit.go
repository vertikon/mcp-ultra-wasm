package security

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// RateLimiter gerencia rate limiting
type RateLimiter struct {
	limiter *rate.Limiter
	config  *RateLimitConfig
	logger  *zap.Logger
	mu      sync.RWMutex
}

type RateLimitConfig struct {
	Enabled     bool           `json:"enabled"`
	Rate        rate.Limit     `json:"rate"`         // requisições por segundo
	Burst       int            `json:"burst"`        // capacidade do bucket
	Headers     bool           `json:"headers"`      // adicionar headers de rate limit
	KeyFunc     KeyFunc        `json:"-"`            // função para extrair chave do rate limit
	SkipPaths   []string       `json:"skip_paths"`   // paths para pular rate limiting
	SkipSuccess bool           `json:"skip_success"` // pular em requisições bem-sucedidas
	Store       RateLimitStore `json:"-"`            // store para rate limit distribuído
}

// KeyFunc extrai chave para rate limiting do request
type KeyFunc func(*gin.Context) string

// RateLimitStore interface para armazenamento distribuído
type RateLimitStore interface {
	Allow(key string, limit rate.Limit, burst int) bool
	GetBucket(key string) *rate.Limiter
	SetBucket(key string, limiter *rate.Limiter)
	RemoveBucket(key string)
	CleanupExpired()
}

// DefaultRateLimitStore implementação em memória
type DefaultRateLimitStore struct {
	buckets map[string]*rate.Limiter
	mu      sync.RWMutex
	logger  *zap.Logger
}

func NewDefaultRateLimitStore(logger *zap.Logger) *DefaultRateLimitStore {
	store := &DefaultRateLimitStore{
		buckets: make(map[string]*rate.Limiter),
		logger:  logger.Named("ratelimit.store"),
	}

	// Cleanup periódico
	go store.cleanupRoutine()

	return store
}

func (s *DefaultRateLimitStore) Allow(key string, limit rate.Limit, burst int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	limiter, exists := s.buckets[key]
	if !exists {
		limiter = rate.NewLimiter(limit, burst)
		s.buckets[key] = limiter
	}

	return limiter.Allow()
}

func (s *DefaultRateLimitStore) GetBucket(key string) *rate.Limiter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.buckets[key]
}

func (s *DefaultRateLimitStore) SetBucket(key string, limiter *rate.Limiter) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buckets[key] = limiter
}

func (s *DefaultRateLimitStore) RemoveBucket(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.buckets, key)
}

func (s *DefaultRateLimitStore) CleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Para implementação em memória, não há expiração
	// TODO: Implementar lógica de expiração baseada em último acesso
}

func (s *DefaultRateLimitStore) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.CleanupExpired()
	}
}

func NewRateLimiter(config *RateLimitConfig, logger *zap.Logger) *RateLimiter {
	if config == nil {
		config = &RateLimitConfig{
			Enabled: true,
			Rate:    rate.Limit(10), // 10 requisições por segundo
			Burst:   20,             // capacidade de 20 requisições
			Headers: true,
			SkipPaths: []string{
				"/health",
				"/metrics",
			},
			SkipSuccess: false,
		}
	}

	limiter := &RateLimiter{
		config: config,
		logger: logger.Named("ratelimit"),
	}

	// Configurar função de chave padrão (IP)
	if config.KeyFunc == nil {
		config.KeyFunc = limiter.defaultKeyFunc
	}

	// Configurar store padrão
	if config.Store == nil {
		config.Store = NewDefaultRateLimitStore(logger)
	}

	logger.Info("Rate limiter initialized",
		zap.Bool("enabled", config.Enabled),
		zap.Float64("rate_per_second", float64(config.Rate)),
		zap.Int("burst", config.Burst))

	return limiter
}

// defaultKeyFunc extrai IP do cliente como chave
func (rl *RateLimiter) defaultKeyFunc(c *gin.Context) string {
	// Tentar obter IP real através de headers
	ip := c.GetHeader("X-Real-IP")
	if ip == "" {
		ip = c.GetHeader("X-Forwarded-For")
		if ip != "" {
			// Pegar primeiro IP se houver múltiplos
			if idx := len(ip); idx > 0 {
				ip = ip[:idx]
			}
		}
	}

	if ip == "" {
		ip = c.ClientIP()
	}

	return ip
}

// IPKeyFunc função de chave baseada em IP
func (rl *RateLimiter) IPKeyFunc(c *gin.Context) string {
	return rl.defaultKeyFunc(c)
}

// UserKeyFunc função de chave baseada em usuário autenticado
func (rl *RateLimiter) UserKeyFunc(c *gin.Context) string {
	// Tentar obter ID do usuário autenticado
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return fmt.Sprintf("user:%s", id)
		}
	}

	// Fallback para IP
	return fmt.Sprintf("ip:%s", rl.defaultKeyFunc(c))
}

// PathKeyFunc função de chave baseada em path + IP
func (rl *RateLimiter) PathKeyFunc(c *gin.Context) string {
	ip := rl.defaultKeyFunc(c)
	return fmt.Sprintf("%s:%s", ip, c.Request.URL.Path)
}

// Middleware retorna middleware de rate limiting
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Verificar se deve pular rate limiting
		if rl.shouldSkip(c) {
			c.Next()
			return
		}

		// Extrair chave
		key := rl.config.KeyFunc(c)

		// Verificar rate limit
		if !rl.config.Store.Allow(key, rl.config.Rate, rl.config.Burst) {
			rl.handleRateLimitExceeded(c, key)
			return
		}

		// Adicionar headers se habilitado
		if rl.config.Headers {
			rl.addRateLimitHeaders(c, key)
		}

		c.Next()
	}
}

// shouldSkip verifica se deve pular rate limiting
func (rl *RateLimiter) shouldSkip(c *gin.Context) bool {
	// Verificar paths configurados
	for _, skipPath := range rl.config.SkipPaths {
		if c.Request.URL.Path == skipPath {
			return true
		}
	}

	// Verificar se deve pular em requisições bem-sucedidas
	if rl.config.SkipSuccess {
		// Verificar status após processamento
		c.Next() // Processar requisição primeiro
		return c.Writer.Status() < 400
	}

	return false
}

// handleRateLimitExceeded trata requisições que excederam o rate limit
func (rl *RateLimiter) handleRateLimitExceeded(c *gin.Context, key string) {
	rl.logger.Warn("Rate limit exceeded",
		zap.String("key", key),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("ip", c.ClientIP()))

	// Adicionar headers de rate limit
	if rl.config.Headers {
		rl.addRateLimitHeaders(c, key)
	}

	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":       "Rate limit exceeded",
		"code":        "RATE_LIMIT_EXCEEDED",
		"retry_after": "1s",
	})
	c.Abort()
}

// addRateLimitHeaders adiciona headers informativos sobre rate limit
func (rl *RateLimiter) addRateLimitHeaders(c *gin.Context, key string) {
	limiter := rl.config.Store.GetBucket(key)
	if limiter == nil {
		return
	}

	// Tokens disponíveis no bucket
	tokens := limiter.Tokens()
	c.Header("X-RateLimit-Limit", fmt.Sprintf("%.0f", float64(rl.config.Burst)))
	c.Header("X-RateLimit-Remaining", fmt.Sprintf("%.0f", tokens))
	c.Header("X-RateLimit-Reset", fmt.Sprintf("%.0f", limiter.TokensAt(time.Now().Add(time.Second))))
}

// Allow verifica se chave específica está permitida
func (rl *RateLimiter) Allow(key string) bool {
	if !rl.config.Enabled {
		return true
	}

	return rl.config.Store.Allow(key, rl.config.Rate, rl.config.Burst)
}

// AllowWithCustomLimit verifica com limit customizado
func (rl *RateLimiter) AllowWithCustomLimit(key string, limit rate.Limit, burst int) bool {
	if !rl.config.Enabled {
		return true
	}

	return rl.config.Store.Allow(key, limit, burst)
}

// GetLimiter obtém rate limiter para chave específica
func (rl *RateLimiter) GetLimiter(key string) *rate.Limiter {
	if !rl.config.Enabled {
		return rate.NewLimiter(rate.Inf, 0)
	}

	return rl.config.Store.GetBucket(key)
}

// SetLimiter define rate limiter customizado para chave
func (rl *RateLimiter) SetLimiter(key string, limiter *rate.Limiter) {
	if rl.config.Enabled {
		rl.config.Store.SetBucket(key, limiter)
	}
}

// RemoveBucket remove bucket para chave específica
func (rl *RateLimiter) RemoveBucket(key string) {
	if rl.config.Enabled {
		rl.config.Store.RemoveBucket(key)
	}
}

// UpdateConfig atualiza configuração em runtime
func (rl *RateLimiter) UpdateConfig(config *RateLimitConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.config = config
	rl.logger.Info("Rate limit configuration updated",
		zap.Float64("rate_per_second", float64(config.Rate)),
		zap.Int("burst", config.Burst))
}

// GetConfig retorna configuração atual
func (rl *RateLimiter) GetConfig() *RateLimitConfig {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	return rl.config
}

// GetStats retorna estatísticas do rate limiter
func (rl *RateLimiter) GetStats() map[string]interface{} {
	if store, ok := rl.config.Store.(*DefaultRateLimitStore); ok {
		store.mu.RLock()
		buckets := len(store.buckets)
		store.mu.RUnlock()

		return map[string]interface{}{
			"enabled":        rl.config.Enabled,
			"rate":           float64(rl.config.Rate),
			"burst":          rl.config.Burst,
			"active_buckets": buckets,
		}
	}

	return map[string]interface{}{
		"enabled": rl.config.Enabled,
		"rate":    float64(rl.config.Rate),
		"burst":   rl.config.Burst,
	}
}

// Cleanup remove buckets expirados
func (rl *RateLimiter) Cleanup() {
	if rl.config.Enabled {
		rl.config.Store.CleanupExpired()
	}
}

// CreateCustomLimiter cria rate limiter com configuração customizada
func CreateCustomLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	config := &RateLimitConfig{
		Enabled: true,
		Rate:    rate.Limit(requestsPerSecond),
		Burst:   burst,
		Headers: true,
	}

	return NewRateLimiter(config, zap.NewNop())
}

// Per minute rate limiters comuns
func NewPerMinuteLimiter(requestsPerMinute int) *RateLimiter {
	ratePerSecond := rate.Limit(float64(requestsPerMinute) / 60.0)
	return CreateCustomLimiter(float64(ratePerSecond), requestsPerMinute)
}

func NewPerHourLimiter(requestsPerHour int) *RateLimiter {
	ratePerSecond := rate.Limit(float64(requestsPerHour) / 3600.0)
	return CreateCustomLimiter(float64(ratePerSecond), requestsPerHour)
}

func NewPerDayLimiter(requestsPerDay int) *RateLimiter {
	ratePerSecond := rate.Limit(float64(requestsPerDay) / 86400.0)
	return CreateCustomLimiter(float64(ratePerSecond), requestsPerDay)
}

// Middleware específicos para diferentes estratégias

// GlobalRateLimit rate limit global para todos os requests
func (rl *RateLimiter) GlobalRateLimit() gin.HandlerFunc {
	return rl.Middleware()
}

// PerUserRateLimit rate limit por usuário autenticado
func (rl *RateLimiter) PerUserRateLimit() gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	originalKeyFunc := rl.config.KeyFunc
	rl.config.KeyFunc = rl.UserKeyFunc

	return rl.Middleware()
}

// PerPathRateLimit rate limit por path
func (rl *RateLimiter) PerPathRateLimit() gin.HandlerFunc {
	if !rl.config.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	originalKeyFunc := rl.config.KeyFunc
	rl.config.KeyFunc = rl.PathKeyFunc

	return rl.Middleware()
}

// CustomRateLimit middleware com configuração customizada
func (rl *RateLimiter) CustomRateLimit(config *RateLimitConfig) gin.HandlerFunc {
	originalConfig := rl.config
	rl.config = config

	return func(c *gin.Context) {
		rl.Middleware()(c)
		rl.config = originalConfig
	}
}
