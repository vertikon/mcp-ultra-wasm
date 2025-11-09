package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Context keys for auth data
type contextKey string

const (
	userIDKey     contextKey = "user_id"
	usernameKey   contextKey = "username"
	userRolesKey  contextKey = "user_roles"
	sessionIDKey  contextKey = "session_id"
	authClaimsKey contextKey = "auth_claims"
	clientNameKey contextKey = "client_name"
	apiKeyKey     contextKey = "api_key"
)

type AuthConfig struct {
	JWTSecret     string        `yaml:"jwt_secret" envconfig:"JWT_SECRET" required:"true"`
	JWTExpiry     time.Duration `yaml:"jwt_expiry" envconfig:"JWT_TOKEN_EXPIRY" default:"24h"`
	JWTIssuer     string        `yaml:"jwt_issuer" envconfig:"JWT_ISSUER" default:"mcp-ultra-wasm"`
	APIKeyHeader  string        `yaml:"api_key_header" envconfig:"API_KEY_HEADER" default:"X-API-Key"`
	RequiredRoles []string      `yaml:"required_roles" envconfig:"REQUIRED_ROLES"`
	SkipPaths     []string      `yaml:"skip_paths"`
}

type AuthClaims struct {
	UserID    string   `json:"user_id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	Scope     []string `json:"scope"`
	SessionID string   `json:"session_id"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	config *AuthConfig
	logger *zap.Logger
	tracer trace.Tracer
}

func NewAuthMiddleware(config *AuthConfig, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		config: config,
		logger: logger,
		tracer: otel.Tracer("mcp-ultra-wasm/auth"),
	}
}

// JWTAuth middleware for JWT token authentication
func (a *AuthMiddleware) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := a.tracer.Start(r.Context(), "auth.jwt_validation")
		defer span.End()

		// Skip authentication for certain paths
		if a.shouldSkipAuth(r.URL.Path) {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Extract token from Authorization header
		token := a.extractToken(r)
		if token == "" {
			span.SetStatus(codes.Error, "missing authorization token")
			a.logger.Warn("Missing authorization token",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("remote_addr", r.RemoteAddr))
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}

		// Validate JWT token
		claims, err := a.validateJWT(token)
		if err != nil {
			span.SetStatus(codes.Error, fmt.Sprintf("invalid token: %v", err))
			a.logger.Warn("Invalid JWT token",
				zap.String("token_prefix", token[:min(len(token), 10)]),
				zap.String("path", r.URL.Path),
				zap.Error(err))
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to span attributes
		span.SetAttributes(
			attribute.String("user.id", claims.UserID),
			attribute.String("user.username", claims.Username),
			attribute.String("session.id", claims.SessionID),
			attribute.StringSlice("user.roles", claims.Roles),
		)

		// Add user context
		ctx = context.WithValue(ctx, userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, usernameKey, claims.Username)
		ctx = context.WithValue(ctx, userRolesKey, claims.Roles)
		ctx = context.WithValue(ctx, sessionIDKey, claims.SessionID)
		ctx = context.WithValue(ctx, authClaimsKey, claims)

		a.logger.Debug("JWT authentication successful",
			zap.String("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.Strings("roles", claims.Roles),
			zap.String("path", r.URL.Path))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// APIKeyAuth middleware for API key authentication
func (a *AuthMiddleware) APIKeyAuth(validAPIKeys map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, span := a.tracer.Start(r.Context(), "auth.api_key_validation")
			defer span.End()

			// Skip authentication for certain paths
			if a.shouldSkipAuth(r.URL.Path) {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			apiKey := r.Header.Get(a.config.APIKeyHeader)
			if apiKey == "" {
				span.SetStatus(codes.Error, "missing api key")
				http.Error(w, "Unauthorized: missing API key", http.StatusUnauthorized)
				return
			}

			clientName, exists := validAPIKeys[apiKey]
			if !exists {
				span.SetStatus(codes.Error, "invalid api key")
				a.logger.Warn("Invalid API key",
					zap.String("key_prefix", apiKey[:min(len(apiKey), 8)]),
					zap.String("path", r.URL.Path))
				http.Error(w, "Unauthorized: invalid API key", http.StatusUnauthorized)
				return
			}

			// Add client context
			ctx = context.WithValue(ctx, clientNameKey, clientName)
			ctx = context.WithValue(ctx, apiKeyKey, apiKey[:8]+"...")

			span.SetAttributes(
				attribute.String("client.name", clientName),
				attribute.String("auth.method", "api_key"),
			)

			a.logger.Debug("API key authentication successful",
				zap.String("client", clientName),
				zap.String("path", r.URL.Path))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware to check if user has required role
func (a *AuthMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, span := a.tracer.Start(r.Context(), "auth.role_check")
			defer span.End()

			claims, ok := r.Context().Value(authClaimsKey).(*AuthClaims)
			if !ok {
				span.SetStatus(codes.Error, "missing auth claims")
				http.Error(w, "Forbidden: authentication required", http.StatusForbidden)
				return
			}

			hasRole := false
			for _, role := range claims.Roles {
				if role == requiredRole || role == "admin" {
					hasRole = true
					break
				}
			}

			if !hasRole {
				span.SetStatus(codes.Error, "insufficient permissions")
				a.logger.Warn("Access denied - insufficient role",
					zap.String("user_id", claims.UserID),
					zap.String("required_role", requiredRole),
					zap.Strings("user_roles", claims.Roles))
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			span.SetAttributes(
				attribute.String("required.role", requiredRole),
				attribute.Bool("access.granted", true),
			)

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByUser applies rate limiting per user
func (a *AuthMiddleware) RateLimitByUser(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter (for production, use Redis)
	userLimits := make(map[string]*rateLimitInfo)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(userIDKey)
			if userID == nil {
				// If no user ID, skip rate limiting
				next.ServeHTTP(w, r)
				return
			}

			userKey := userID.(string)
			now := time.Now()

			info, exists := userLimits[userKey]
			if !exists || now.Sub(info.windowStart) > window {
				userLimits[userKey] = &rateLimitInfo{
					requests:    1,
					windowStart: now,
				}
				next.ServeHTTP(w, r)
				return
			}

			if info.requests >= maxRequests {
				a.logger.Warn("Rate limit exceeded",
					zap.String("user_id", userKey),
					zap.Int("requests", info.requests),
					zap.Int("max_requests", maxRequests))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			info.requests++
			next.ServeHTTP(w, r)
		})
	}
}

type rateLimitInfo struct {
	requests    int
	windowStart time.Time
}

// Helper functions
func (a *AuthMiddleware) extractToken(r *http.Request) string {
	// Try Authorization header first
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Try query parameter as fallback
	return r.URL.Query().Get("token")
}

func (a *AuthMiddleware) validateJWT(tokenString string) (*AuthClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*AuthClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	if claims.Issuer != a.config.JWTIssuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", a.config.JWTIssuer, claims.Issuer)
	}

	// Validate expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

func (a *AuthMiddleware) shouldSkipAuth(path string) bool {
	defaultSkipPaths := []string{
		"/health", "/healthz", "/ready", "/readyz", "/live", "/livez",
		"/metrics", "/ping", "/status", "/version",
		"/swagger", "/docs", "/api-docs",
	}

	allSkipPaths := append(defaultSkipPaths, a.config.SkipPaths...)

	for _, skipPath := range allSkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// LoggingMiddleware logs HTTP requests with auth context
func (a *AuthMiddleware) LoggingMiddleware(next http.Handler) http.Handler {
	return middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: &zapLoggerWrapper{logger: a.logger},
	})(next)
}

type zapLoggerWrapper struct {
	logger *zap.Logger
}

func (z *zapLoggerWrapper) Print(v ...interface{}) {
	z.logger.Info(fmt.Sprint(v...))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateJWT creates a new JWT token for a user
func (a *AuthMiddleware) GenerateJWT(userID, username, email string, roles []string) (string, error) {
	now := time.Now()

	claims := &AuthClaims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		Roles:     roles,
		SessionID: fmt.Sprintf("sess_%d_%s", now.Unix(), userID),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.config.JWTIssuer,
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(now.Add(a.config.JWTExpiry)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        fmt.Sprintf("jwt_%d", now.UnixNano()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.JWTSecret))
}
