package security

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SecurityMiddleware gerencia todos os middlewares de segurança
type SecurityMiddleware struct {
	auth      *AuthManager
	cors      *CORSManager
	rateLimit *RateLimiter
	logger    *zap.Logger
	config    *SecurityConfig
}

type SecurityConfig struct {
	Auth      *AuthConfig      `json:"auth"`
	CORS      *CORSConfig      `json:"cors"`
	RateLimit *RateLimitConfig `json:"rate_limit"`
	Security  *GeneralConfig   `json:"security"`
}

type GeneralConfig struct {
	Enabled               bool              `json:"enabled"`
	TrustedProxies        []string          `json:"trusted_proxies"`
	RequestTimeout        time.Duration     `json:"request_timeout"`
	MaxHeaderBytes        int               `json:"max_header_bytes"`
	MaxRequestBodyBytes   int64             `json:"max_request_body_bytes"`
	EnableHSTS            bool              `json:"enable_hsts"`
	HSTSMaxAge            time.Duration     `json:"sts_max_age"`
	EnableContentSniffing bool              `json:"enable_content_sniffing"`
	EnableXSSProtection   bool              `json:"enable_xss_protection"`
	EnableFrameOptions    bool              `json:"enable_frame_options"`
	ContentTypeNosniff    bool              `json:"content_type_nosniff"`
	FrameOptions          string            `json:"frame_options"`
	XSSProtection         string            `json:"xss_protection"`
	CustomHeaders         map[string]string `json:"custom_headers"`
}

func NewSecurityMiddleware(config *SecurityConfig, logger *zap.Logger) (*SecurityMiddleware, error) {
	if config == nil {
		config = &SecurityConfig{
			Auth:      &AuthConfig{Enabled: false},
			CORS:      &CORSConfig{Enabled: false},
			RateLimit: &RateLimitConfig{Enabled: false},
			Security: &GeneralConfig{
				Enabled:               true,
				RequestTimeout:        30 * time.Second,
				MaxHeaderBytes:        1 << 20,  // 1MB
				MaxRequestBodyBytes:   10 << 20, // 10MB
				EnableHSTS:            true,
				HSTSMaxAge:            365 * 24 * time.Hour,
				EnableContentSniffing: true,
				EnableXSSProtection:   true,
				EnableFrameOptions:    true,
				ContentTypeNosniff:    true,
				FrameOptions:          "DENY",
				XSSProtection:         "1; mode=block",
				CustomHeaders:         make(map[string]string),
			},
		}
	}

	// Criar componentes individuais
	authManager, err := NewAuthManager(config.Auth, logger, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar auth manager: %w", err)
	}

	corsManager := NewCORSManager(config.CORS, logger)
	rateLimiter := NewRateLimiter(config.RateLimit, logger)

	middleware := &SecurityMiddleware{
		auth:      authManager,
		cors:      corsManager,
		rateLimit: rateLimiter,
		logger:    logger.Named("security"),
		config:    config,
	}

	logger.Info("Security middleware initialized",
		zap.Bool("auth_enabled", config.Auth.Enabled),
		zap.Bool("cors_enabled", config.CORS.Enabled),
		zap.Bool("rate_limit_enabled", config.RateLimit.Enabled),
		zap.Bool("security_enabled", config.Security.Enabled))

	return middleware, nil
}

// Middleware retorna middleware combinado de segurança
func (sm *SecurityMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Aplicar headers de segurança
		if sm.config.Security.Enabled {
			sm.applySecurityHeaders(c)
		}

		// Aplicar CORS
		if sm.config.CORS.Enabled {
			sm.cors.CORSMiddleware()(c)
			if c.IsAborted() {
				return
			}
		}

		// Aplicar Rate Limiting
		if sm.config.RateLimit.Enabled {
			sm.rateLimit.Middleware()(c)
			if c.IsAborted() {
				return
			}
		}

		// Aplicar Autenticação JWT
		if sm.config.Auth.Enabled {
			sm.auth.JWTMiddleware()(c)
			if c.IsAborted() {
				return
			}
		}

		c.Next()
	}
}

// applySecurityHeaders aplica headers de segurança HTTP
func (sm *SecurityMiddleware) applySecurityHeaders(c *gin.Context) {
	security := sm.config.Security

	// X-Content-Type-Options
	if security.ContentTypeNosniff {
		c.Header("X-Content-Type-Options", "nosniff")
	}

	// X-Frame-Options
	if security.EnableFrameOptions && security.FrameOptions != "" {
		c.Header("X-Frame-Options", security.FrameOptions)
	}

	// X-XSS-Protection
	if security.EnableXSSProtection && security.XSSProtection != "" {
		c.Header("X-XSS-Protection", security.XSSProtection)
	}

	// Strict-Transport-Security (HSTS)
	if security.EnableHSTS && c.Request.TLS != nil {
		maxAge := int(security.HSTSMaxAge.Seconds())
		hstsValue := fmt.Sprintf("max-age=%d; includeSubDomains", maxAge)
		c.Header("Strict-Transport-Security", hstsValue)
	}

	// Content-Security-Policy (básico)
	if security.EnableContentSniffing {
		csp := "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self' ws: wss:;"
		c.Header("Content-Security-Policy", csp)
	}

	// Referrer-Policy
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

	// Permissions-Policy (antes Feature-Policy)
	permissionsPolicy := "geolocation=(), microphone=(), camera=(), payment=(), usb=()"
	c.Header("Permissions-Policy", permissionsPolicy)

	// Custom headers
	for key, value := range security.CustomHeaders {
		c.Header(key, value)
	}
}

// AuthMiddleware retorna apenas middleware de autenticação
func (sm *SecurityMiddleware) AuthMiddleware() gin.HandlerFunc {
	if sm.config.Auth.Enabled {
		return sm.auth.JWTMiddleware()
	}
	return func(c *gin.Context) { c.Next() }
}

// CORSMiddleware retorna apenas middleware CORS
func (sm *SecurityMiddleware) CORSMiddleware() gin.HandlerFunc {
	if sm.config.CORS.Enabled {
		return sm.cors.CORSMiddleware()
	}
	return func(c *gin.Context) { c.Next() }
}

// RateLimitMiddleware retorna apenas middleware de rate limiting
func (sm *SecurityMiddleware) RateLimitMiddleware() gin.HandlerFunc {
	if sm.config.RateLimit.Enabled {
		return sm.rateLimit.Middleware()
	}
	return func(c *gin.Context) { c.Next() }
}

// RequirePermission middleware para verificar permissão específica
func (sm *SecurityMiddleware) RequirePermission(permission string) gin.HandlerFunc {
	if sm.config.Auth.Enabled {
		return sm.auth.RequirePermission(permission)
	}
	return func(c *gin.Context) { c.Next() }
}

// RequireRole middleware para verificar role específica
func (sm *SecurityMiddleware) RequireRole(role string) gin.HandlerFunc {
	if sm.config.Auth.Enabled {
		return sm.auth.RequireRole(role)
	}
	return func(c *gin.Context) { c.Next() }
}

// RequireAdmin middleware para verificar se é admin
func (sm *SecurityMiddleware) RequireAdmin() gin.HandlerFunc {
	return sm.RequireRole("admin")
}

// LoginHandler handler para login
func (sm *SecurityMiddleware) LoginHandler(c *gin.Context) {
	if !sm.config.Auth.Enabled {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Auth not enabled"})
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := sm.auth.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// RefreshTokenHandler handler para refresh token
func (sm *SecurityMiddleware) RefreshTokenHandler(c *gin.Context) {
	if !sm.config.Auth.Enabled {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Auth not enabled"})
		return
	}

	if !sm.config.Auth.EnableRefresh {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Refresh tokens not enabled"})
		return
	}

	// Obter refresh token do request
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validar refresh token
	claims, err := sm.auth.ValidateToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Verificar se é refresh token
	if claims.Audience != nil && len(claims.Audience) > 0 && claims.Audience[0] != "web-wasm-refresh" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not a refresh token"})
		return
	}

	// Gerar novos tokens
	user := &User{
		ID:          claims.UserID,
		Username:    claims.Username,
		Email:       claims.Email,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}

	newToken, err := sm.auth.GenerateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := gin.H{
		"access_token": newToken.Token,
		"expires_at":   newToken.ExpiresAt,
	}

	c.JSON(http.StatusOK, response)
}

// LogoutHandler handler para logout
func (sm *SecurityMiddleware) LogoutHandler(c *gin.Context) {
	if !sm.config.Auth.Enabled {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Auth not enabled"})
		return
	}

	// Obter claims do contexto
	claimsInterface, exists := c.Get("user_claims")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	claims, ok := claimsInterface.(*CustomClaims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user claims"})
		return
	}

	// Revogar token se blacklist estiver habilitado
	if sm.config.Auth.EnableBlacklist {
		if err := sm.auth.RevokeToken(claims.TokenID); err != nil {
			sm.logger.Error("Erro ao revogar token", zap.Error(err))
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// UserInfoHandler handler para obter informações do usuário
func (sm *SecurityMiddleware) UserInfoHandler(c *gin.Context) {
	if !sm.config.Auth.Enabled {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "Auth not enabled"})
		return
	}

	userInfoInterface, exists := c.Get("user_info")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	userInfo, ok := userInfoInterface.(*UserInfo)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user info"})
		return
	}

	c.JSON(http.StatusOK, userInfo)
}

// SecurityInfoHandler handler para obter informações de segurança
func (sm *SecurityMiddleware) SecurityInfoHandler(c *gin.Context) {
	info := gin.H{
		"security": gin.H{
			"enabled": sm.config.Security.Enabled,
		},
		"auth": gin.H{
			"enabled":      sm.config.Auth.Enabled,
			"issuer":       sm.config.Auth.Issuer,
			"token_expiry": sm.config.Auth.TokenExpiry.String(),
		},
		"cors":       sm.cors.CORSInfo(),
		"rate_limit": sm.rateLimit.GetStats(),
	}

	c.JSON(http.StatusOK, info)
}

// GetAuthManager retorna o auth manager
func (sm *SecurityMiddleware) GetAuthManager() *AuthManager {
	return sm.auth
}

// GetCORSManager retorna o CORS manager
func (sm *SecurityMiddleware) GetCORSManager() *CORSManager {
	return sm.cors
}

// GetRateLimiter retorna o rate limiter
func (sm *SecurityMiddleware) GetRateLimiter() *RateLimiter {
	return sm.rateLimit
}

// GetConfig retorna configuração de segurança
func (sm *SecurityMiddleware) GetConfig() *SecurityConfig {
	return sm.config
}

// UpdateConfig atualiza configuração de segurança
func (sm *SecurityMiddleware) UpdateConfig(config *SecurityConfig) {
	sm.config = config
	sm.logger.Info("Security configuration updated")
}

// Cleanup realiza cleanup dos componentes de segurança
func (sm *SecurityMiddleware) Cleanup() error {
	var errors []error

	// Cleanup auth
	if err := sm.auth.Cleanup(); err != nil {
		errors = append(errors, fmt.Errorf("auth cleanup error: %w", err))
	}

	// Cleanup rate limiter
	sm.rateLimit.Cleanup()

	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}

	return nil
}

// HealthCheck verifica saúde dos componentes de segurança
func (sm *SecurityMiddleware) HealthCheck() gin.H {
	status := gin.H{
		"status": "healthy",
		"components": gin.H{
			"auth":       "disabled",
			"cors":       "disabled",
			"rate_limit": "disabled",
			"security":   "disabled",
		},
	}

	if sm.config.Auth.Enabled {
		status["components"].(gin.H)["auth"] = "enabled"
	}

	if sm.config.CORS.Enabled {
		status["components"].(gin.H)["cors"] = "enabled"
	}

	if sm.config.RateLimit.Enabled {
		status["components"].(gin.H)["rate_limit"] = "enabled"
	}

	if sm.config.Security.Enabled {
		status["components"].(gin.H)["security"] = "enabled"
	}

	return status
}

// RegisterRoutes registra rotas de segurança no router
func (sm *SecurityMiddleware) RegisterRoutes(router *gin.RouterGroup) {
	if !sm.config.Auth.Enabled {
		return
	}

	auth := router.Group("/auth")
	{
		auth.POST("/login", sm.LoginHandler)
		auth.POST("/refresh", sm.RefreshTokenHandler)
		auth.POST("/logout", sm.RequireAuth().Middleware(), sm.LogoutHandler)
		auth.GET("/user", sm.RequireAuth().Middleware(), sm.UserInfoHandler)
	}

	router.GET("/security/info", sm.SecurityInfoHandler)
}

// RequireAuth retorna wrapper para autenticação
func (sm *SecurityMiddleware) RequireAuth() *SecurityMiddleware {
	return &SecurityMiddleware{
		auth:      sm.auth,
		cors:      sm.cors,
		rateLimit: sm.rateLimit,
		logger:    sm.logger,
		config:    sm.config,
	}
}

// Middleware helper para criar chains de middlewares
func (sm *SecurityMiddleware) Chain(middlewares ...gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, middleware := range middlewares {
			middleware(c)
			if c.IsAborted() {
				return
			}
		}
		c.Next()
	}
}
