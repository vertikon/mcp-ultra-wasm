package security

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthManager gerencia autenticação JWT
type AuthManager struct {
	secretKey  []byte
	issuer     string
	logger     *zap.Logger
	config     *AuthConfig
	tokenStore TokenStore
}

type AuthConfig struct {
	Enabled         bool          `json:"enabled"`
	SecretKey       string        `json:"secret_key"`
	Issuer          string        `json:"issuer"`
	TokenExpiry     time.Duration `json:"token_expiry"`
	RefreshExpiry   time.Duration `json:"refresh_expiry"`
	EnableRefresh   bool          `json:"enable_refresh"`
	SkipAuthPaths   []string      `json:"skip_auth_paths"`
	RequireHTTPS    bool          `json:"require_https"`
	EnableBlacklist bool          `json:"enable_blacklist"`
	MaxTokenAge     time.Duration `json:"max_token_age"`
}

// TokenStore interface para armazenamento de tokens
type TokenStore interface {
	StoreToken(tokenID string, claims *CustomClaims) error
	RevokeToken(tokenID string) error
	IsTokenRevoked(tokenID string) (bool, error)
	CleanupExpired() error
}

// CustomClaims claims customizados para JWT
type CustomClaims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	TokenID     string   `json:"token_id"`
	jwt.RegisteredClaims
}

type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *UserInfo `json:"user"`
}

type UserInfo struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Type      string    `json:"type"`
}

func NewAuthManager(config *AuthConfig, logger *zap.Logger, tokenStore TokenStore) (*AuthManager, error) {
	if config == nil {
		config = &AuthConfig{
			Enabled:       false,
			SecretKey:     "default-secret-key-change-in-production",
			Issuer:        "wasm-server",
			TokenExpiry:   1 * time.Hour,
			RefreshExpiry: 24 * time.Hour,
			SkipAuthPaths: []string{
				"/health",
				"/metrics",
				"/ws",
			},
			RequireHTTPS:    false,
			EnableBlacklist: false,
			MaxTokenAge:     24 * time.Hour,
		}
	}

	if config.Enabled && config.SecretKey == "" {
		return nil, fmt.Errorf("secret key is required when auth is enabled")
	}

	manager := &AuthManager{
		secretKey:  []byte(config.SecretKey),
		issuer:     config.Issuer,
		logger:     logger.Named("auth"),
		config:     config,
		tokenStore: tokenStore,
	}

	// Se não fornecido token store, usar memória
	if tokenStore == nil {
		manager.tokenStore = NewMemoryTokenStore(logger)
	}

	logger.Info("Auth manager initialized",
		zap.Bool("enabled", config.Enabled),
		zap.String("issuer", config.Issuer),
		zap.Duration("token_expiry", config.TokenExpiry))

	return manager, nil
}

// GenerateToken gera um token JWT
func (am *AuthManager) GenerateToken(user *User) (*TokenResponse, error) {
	now := time.Now()
	tokenID := fmt.Sprintf("token_%d_%s", now.Unix(), user.ID)

	claims := &CustomClaims{
		UserID:      user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Roles:       user.Roles,
		Permissions: user.Permissions,
		TokenID:     tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    am.issuer,
			Subject:   user.ID,
			Audience:  []string{"wasm"},
			ExpiresAt: jwt.NewNumericDate(now.Add(am.config.TokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(am.secretKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao assinar token: %w", err)
	}

	// Armazenar token se blacklist estiver habilitado
	if am.config.EnableBlacklist {
		if err := am.tokenStore.StoreToken(tokenID, claims); err != nil {
			am.logger.Error("Erro ao armazenar token", zap.Error(err))
		}
	}

	response := &TokenResponse{
		Token:     tokenString,
		ExpiresAt: now.Add(am.config.TokenExpiry),
		Type:      "Bearer",
	}

	return response, nil
}

// GenerateRefreshToken gera um refresh token
func (am *AuthManager) GenerateRefreshToken(user *User) (*TokenResponse, error) {
	if !am.config.EnableRefresh {
		return nil, fmt.Errorf("refresh tokens not enabled")
	}

	now := time.Now()
	tokenID := fmt.Sprintf("refresh_%d_%s", now.Unix(), user.ID)

	claims := &CustomClaims{
		UserID:   user.ID,
		Username: user.Username,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    am.issuer,
			Subject:   user.ID,
			Audience:  []string{"wasm-refresh"},
			ExpiresAt: jwt.NewNumericDate(now.Add(am.config.RefreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(am.secretKey)
	if err != nil {
		return nil, fmt.Errorf("erro ao assinar refresh token: %w", err)
	}

	response := &TokenResponse{
		Token:     tokenString,
		ExpiresAt: now.Add(am.config.RefreshExpiry),
		Type:      "Refresh",
	}

	return response, nil
}

// ValidateToken valida um token JWT
func (am *AuthManager) ValidateToken(tokenString string) (*CustomClaims, error) {
	if !am.config.Enabled {
		return nil, fmt.Errorf("auth is disabled")
	}

	// Remover prefixo "Bearer "
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = tokenString[7:]
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de assinatura inválido: %v", token.Header["alg"])
		}
		return am.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("erro ao parsear token: %w", err)
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token inválido")
	}

	// Verificar se não expirou
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("token expirado")
	}

	// Verificar blacklist se habilitado
	if am.config.EnableBlacklist {
		if revoked, err := am.tokenStore.IsTokenRevoked(claims.TokenID); err != nil {
			am.logger.Error("Erro ao verificar blacklist", zap.Error(err))
		} else if revoked {
			return nil, fmt.Errorf("token revogado")
		}
	}

	// Verificar idade máxima do token
	if am.config.MaxTokenAge > 0 && claims.IssuedAt != nil {
		if time.Since(claims.IssuedAt.Time) > am.config.MaxTokenAge {
			return nil, fmt.Errorf("token muito antigo")
		}
	}

	return claims, nil
}

// RevokeToken revoga um token
func (am *AuthManager) RevokeToken(tokenID string) error {
	if !am.config.EnableBlacklist {
		return fmt.Errorf("blacklist not enabled")
	}

	return am.tokenStore.RevokeToken(tokenID)
}

// Login autentica um usuário
func (am *AuthManager) Login(req *LoginRequest) (*LoginResponse, error) {
	// TODO: Implementar autenticação real com banco de dados
	// Por enquanto, usar autenticação simulada para demonstração
	user, err := am.authenticateUser(req.Username, req.Password)
	if err != nil {
		am.logger.Warn("Falha na autenticação",
			zap.String("username", req.Username),
			zap.Error(err))
		return nil, err
	}

	// Gerar tokens
	accessToken, err := am.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar access token: %w", err)
	}

	response := &LoginResponse{
		AccessToken: accessToken.Token,
		ExpiresAt:   accessToken.ExpiresAt,
		User: &UserInfo{
			ID:          user.ID,
			Username:    user.Username,
			Email:       user.Email,
			Roles:       user.Roles,
			Permissions: user.Permissions,
		},
	}

	// Gerar refresh token se habilitado
	if am.config.EnableRefresh {
		refreshToken, err := am.GenerateRefreshToken(user)
		if err != nil {
			am.logger.Error("Erro ao gerar refresh token", zap.Error(err))
		} else {
			response.RefreshToken = refreshToken.Token
		}
	}

	am.logger.Info("Usuário autenticado com sucesso",
		zap.String("username", user.Username),
		zap.String("user_id", user.ID))

	return response, nil
}

// authenticateUser autentica usuário (implementação simulada)
func (am *AuthManager) authenticateUser(username, password string) (*User, error) {
	// TODO: Implementar autenticação real com hash de senha
	// Por enquanto, validar usuário demo
	if username == "admin" && password == "admin123" {
		return &User{
			ID:          "user_1",
			Username:    "admin",
			Email:       "admin@webwasm.local",
			Roles:       []string{"admin", "user"},
			Permissions: []string{"read", "write", "execute", "admin"},
			Active:      true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		}, nil
	}

	if username == "user" && password == "user123" {
		return &User{
			ID:          "user_2",
			Username:    "user",
			Email:       "user@webwasm.local",
			Roles:       []string{"user"},
			Permissions: []string{"read", "write"},
			Active:      true,
			CreatedAt:   time.Now().Add(-15 * 24 * time.Hour),
			UpdatedAt:   time.Now(),
		}, nil
	}

	return nil, fmt.Errorf("credenciais inválidas")
}

// CheckPermission verifica se o usuário tem permissão específica
func (am *AuthManager) CheckPermission(claims *CustomClaims, permission string) bool {
	for _, p := range claims.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// CheckRole verifica se o usuário tem role específica
func (am *AuthManager) CheckRole(claims *CustomClaims, role string) bool {
	for _, r := range claims.Roles {
		if r == role || r == "admin" {
			return true
		}
	}
	return false
}

// GetUserInfo extrai informações do usuário dos claims
func (am *AuthManager) GetUserInfo(claims *CustomClaims) *UserInfo {
	return &UserInfo{
		ID:          claims.UserID,
		Username:    claims.Username,
		Email:       claims.Email,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
	}
}

// ShouldSkipAuth verifica se a path deve pular autenticação
func (am *AuthManager) ShouldSkipAuth(path string) bool {
	for _, skipPath := range am.config.SkipAuthPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// JWTMiddleware middleware de autenticação JWT
func (am *AuthManager) JWTMiddleware() gin.HandlerFunc {
	if !am.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Verificar se deve pular autenticação
		if am.ShouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Verificar HTTPS se requerido
		if am.config.RequireHTTPS && c.Request.TLS == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "HTTPS required"})
			c.Abort()
			return
		}

		// Extrair token do header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Validar token
		claims, err := am.ValidateToken(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		// Adicionar claims ao contexto
		c.Set("user_claims", claims)
		c.Set("user_info", am.GetUserInfo(claims))

		c.Next()
	}
}

// RequirePermission middleware para verificar permissão específica
func (am *AuthManager) RequirePermission(permission string) gin.HandlerFunc {
	if !am.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*CustomClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user claims"})
			c.Abort()
			return
		}

		if !am.CheckPermission(userClaims, permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole middleware para verificar role específica
func (am *AuthManager) RequireRole(role string) gin.HandlerFunc {
	if !am.config.Enabled {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		claims, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userClaims, ok := claims.(*CustomClaims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user claims"})
			c.Abort()
			return
		}

		if !am.CheckRole(userClaims, role) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient role"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Cleanup limpa tokens expirados
func (am *AuthManager) Cleanup() error {
	if am.config.EnableBlacklist {
		return am.tokenStore.CleanupExpired()
	}
	return nil
}
