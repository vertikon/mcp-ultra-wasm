package security

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Context keys for auth data
type contextKey string

const (
	userKey     contextKey = "user"
	userIDKey   contextKey = "user_id"
	tenantIDKey contextKey = "tenant_id"
)

// OPAAuthorizer is the interface for OPA authorization
type OPAAuthorizer interface {
	IsAuthorized(ctx context.Context, claims *Claims, method, path string) bool
}

// Claims represents JWT claims
type Claims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Role     string   `json:"role"`
	Scopes   []string `json:"scopes"`
	TenantID string   `json:"tenant_id"`
	jwt.RegisteredClaims
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Mode          string        `yaml:"mode"` // jwt, oauth2, api-key
	JWKSUrl       string        `yaml:"jwks_url"`
	Issuer        string        `yaml:"issuer"`
	Audience      string        `yaml:"audience"`
	TokenExpiry   time.Duration `yaml:"token_expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
}

// AuthService handles JWT authentication and authorization
type AuthService struct {
	config     AuthConfig
	publicKeys map[string]*rsa.PublicKey
	logger     *zap.Logger
	opa        OPAAuthorizer
}

// NewAuthService creates a new authentication service
func NewAuthService(config AuthConfig, logger *zap.Logger, opa OPAAuthorizer) *AuthService {
	as := &AuthService{
		config:     config,
		publicKeys: make(map[string]*rsa.PublicKey),
		logger:     logger,
		opa:        opa,
	}

	// Load JWKS on startup
	if err := as.loadJWKS(); err != nil {
		logger.Error("Failed to load JWKS", zap.Error(err))
	}

	// Refresh JWKS every hour
	go as.refreshJWKS(context.Background())

	return as
}

// JWTMiddleware validates JWT tokens and sets user context
func (as *AuthService) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoints
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" || r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			as.writeUnauthorized(w, "missing authorization header")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			as.writeUnauthorized(w, "invalid authorization header format")
			return
		}

		// Parse and validate token
		claims, err := as.validateToken(tokenString)
		if err != nil {
			as.logger.Warn("Token validation failed", zap.Error(err))
			as.writeUnauthorized(w, "invalid token")
			return
		}

		// Check OPA authorization
		if !as.opa.IsAuthorized(r.Context(), claims, r.Method, r.URL.Path) {
			as.writeForbidden(w, "insufficient permissions")
			return
		}

		// Add user context
		ctx := context.WithValue(r.Context(), userKey, claims)
		ctx = context.WithValue(ctx, userIDKey, claims.UserID)
		ctx = context.WithValue(ctx, tenantIDKey, claims.TenantID)

		// Set security headers
		w.Header().Set("X-User-ID", claims.UserID)
		w.Header().Set("X-Tenant-ID", claims.TenantID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken parses and validates JWT token
func (as *AuthService) validateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure token uses RSA
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing key ID in token header")
		}

		// Get public key for this key ID
		publicKey, ok := as.publicKeys[kid]
		if !ok {
			return nil, fmt.Errorf("unknown key ID: %s", kid)
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate standard claims
	if claims.Issuer != as.config.Issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != as.config.Audience {
		return nil, fmt.Errorf("invalid audience")
	}

	return claims, nil
}

// loadJWKS loads JSON Web Key Set from configured URL
func (as *AuthService) loadJWKS() error {
	if as.config.JWKSUrl == "" {
		return fmt.Errorf("JWKS URL not configured")
	}

	resp, err := http.Get(as.config.JWKSUrl)
	if err != nil {
		return fmt.Errorf("fetching JWKS: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			as.logger.Warn("Failed to close response body", zap.Error(closeErr))
		}
	}()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decoding JWKS: %w", err)
	}

	// Convert JWKS to RSA public keys
	for _, key := range jwks.Keys {
		publicKey, err := as.jwkToRSA(key.N, key.E)
		if err != nil {
			as.logger.Warn("Failed to convert JWK to RSA", zap.String("kid", key.Kid), zap.Error(err))
			continue
		}
		as.publicKeys[key.Kid] = publicKey
	}

	as.logger.Info("Loaded JWKS", zap.Int("keys_count", len(as.publicKeys)))
	return nil
}

// refreshJWKS periodically refreshes JWKS
func (as *AuthService) refreshJWKS(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := as.loadJWKS(); err != nil {
				as.logger.Error("Failed to refresh JWKS", zap.Error(err))
			}
		}
	}
}

// jwkToRSA converts JWK parameters to RSA public key
func (as *AuthService) jwkToRSA(n, e string) (*rsa.PublicKey, error) {
	// Decode base64url encoded modulus (n)
	nBytes, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, fmt.Errorf("decoding modulus: %w", err)
	}

	// Decode base64url encoded exponent (e)
	eBytes, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, fmt.Errorf("decoding exponent: %w", err)
	}

	// Convert bytes to big integers
	nBig := new(big.Int).SetBytes(nBytes)
	eBig := new(big.Int).SetBytes(eBytes)

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: nBig,
		E: int(eBig.Int64()),
	}

	return publicKey, nil
}

// GetUserFromContext extracts user claims from request context
func GetUserFromContext(ctx context.Context) (*Claims, error) {
	user, ok := ctx.Value("user").(*Claims)
	if !ok {
		return nil, fmt.Errorf("user not found in context")
	}
	return user, nil
}

// writeUnauthorized writes 401 Unauthorized response
func (as *AuthService) writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error":   "unauthorized",
		"message": message,
	}); err != nil {
		as.logger.Warn("Failed to encode unauthorized response", zap.Error(err))
	}
}

// writeForbidden writes 403 Forbidden response
func (as *AuthService) writeForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"error":   "forbidden",
		"message": message,
	}); err != nil {
		as.logger.Warn("Failed to encode forbidden response", zap.Error(err))
	}
}

// RequireScope middleware ensures user has required scope
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromContext(r.Context())
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user has required scope
			for _, userScope := range user.Scopes {
				if userScope == scope {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "insufficient scope", http.StatusForbidden)
		})
	}
}

// RequireRole middleware ensures user has required role
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserFromContext(r.Context())
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if user.Role != role && user.Role != "admin" {
				http.Error(w, "insufficient role", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
