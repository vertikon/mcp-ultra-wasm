package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/testhelpers"
)

func TestAuthMiddleware_JWTAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AuthConfig{
		JWTSecret: testhelpers.GetTestJWTSecret(),
		JWTIssuer: "mcp-ultra-wasm-test",
		JWTExpiry: time.Hour,
		SkipPaths: []string{"/health", "/metrics"},
	}

	authMiddleware := NewAuthMiddleware(config, logger)

	t.Run("should skip authentication for configured paths", func(t *testing.T) {
		handler := authMiddleware.JWTAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return 401 for missing token", func(t *testing.T) {
		handler := authMiddleware.JWTAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should validate valid JWT token", func(t *testing.T) {
		// Generate a valid JWT token
		// TEST_USERNAME - safe test value for JWT testing
		token, err := authMiddleware.GenerateJWT("user123", "testuser", "test@example.com", []string{"user"})
		require.NoError(t, err)

		handler := authMiddleware.JWTAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user context is set
			userID := r.Context().Value("user_id")
			assert.Equal(t, "user123", userID)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject invalid JWT token", func(t *testing.T) {
		handler := authMiddleware.JWTAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthMiddleware_APIKeyAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AuthConfig{
		APIKeyHeader: "X-API-Key",
		SkipPaths:    []string{"/health"},
	}

	authMiddleware := NewAuthMiddleware(config, logger)
	publicKey, privateKey := testhelpers.GetTestAPIKeys(t)
	validAPIKeys := map[string]string{publicKey: privateKey}

	t.Run("should validate valid API key", func(t *testing.T) {
		handler := authMiddleware.APIKeyAuth(validAPIKeys)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("X-API-Key", publicKey)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject invalid API key", func(t *testing.T) {
		handler := authMiddleware.APIKeyAuth(validAPIKeys)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("X-API-Key", "invalid-key")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should reject missing API key", func(t *testing.T) {
		handler := authMiddleware.APIKeyAuth(validAPIKeys)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/api/data", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AuthConfig{
		JWTSecret: testhelpers.GetTestJWTSecret(),
		JWTIssuer: "mcp-ultra-wasm-test",
	}

	authMiddleware := NewAuthMiddleware(config, logger)

	t.Run("should allow access with correct role", func(t *testing.T) {
		claims := &AuthClaims{
			UserID: "user123",
			Roles:  []string{"user", "admin"},
		}

		handler := authMiddleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/admin", nil)
		ctx := context.WithValue(req.Context(), authClaimsKey, claims)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should deny access without correct role", func(t *testing.T) {
		claims := &AuthClaims{
			UserID: "user123",
			Roles:  []string{"user"},
		}

		handler := authMiddleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/admin", nil)
		ctx := context.WithValue(req.Context(), authClaimsKey, claims)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("should deny access without auth claims", func(t *testing.T) {
		handler := authMiddleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/admin", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAuthMiddleware_GenerateJWT(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AuthConfig{
		JWTSecret: testhelpers.GetTestJWTSecret(),
		JWTIssuer: "mcp-ultra-wasm-test",
		JWTExpiry: time.Hour,
	}

	authMiddleware := NewAuthMiddleware(config, logger)

	// TEST_USERNAME - safe test value for JWT testing
	token, err := authMiddleware.GenerateJWT("user123", "testuser", "test@example.com", []string{"user", "admin"})
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the generated token
	claims, err := authMiddleware.validateJWT(token)
	require.NoError(t, err)
	assert.Equal(t, "user123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username) // TEST_USERNAME validation
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, []string{"user", "admin"}, claims.Roles)
	assert.Equal(t, "mcp-ultra-wasm-test", claims.Issuer)
}

func TestAuthMiddleware_ExtractToken(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AuthConfig{}
	authMiddleware := NewAuthMiddleware(config, logger)

	t.Run("should extract token from Authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer test-token-123")

		token := authMiddleware.extractToken(req)
		assert.Equal(t, "test-token-123", token)
	})

	t.Run("should extract token from query parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?token=query-token-456", nil)

		token := authMiddleware.extractToken(req)
		assert.Equal(t, "query-token-456", token)
	})

	t.Run("should return empty string for missing token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		token := authMiddleware.extractToken(req)
		assert.Empty(t, token)
	})

	t.Run("should prefer Authorization header over query parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?token=query-token", nil)
		req.Header.Set("Authorization", "Bearer header-token")

		token := authMiddleware.extractToken(req)
		assert.Equal(t, "header-token", token)
	})
}

func TestAuthMiddleware_ShouldSkipAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AuthConfig{
		SkipPaths: []string{"/api/public", "/webhooks"},
	}
	authMiddleware := NewAuthMiddleware(config, logger)

	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/healthz", true},
		{"/ready", true},
		{"/metrics", true},
		{"/api/public", true},
		{"/api/public/users", true},
		{"/webhooks", true},
		{"/webhooks/github", true},
		{"/api/private", false},
		{"/admin", false},
		{"/users", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("path=%s", tt.path), func(t *testing.T) {
			result := authMiddleware.shouldSkipAuth(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthMiddleware_RateLimitByUser(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &AuthConfig{}
	authMiddleware := NewAuthMiddleware(config, logger)

	handler := authMiddleware.RateLimitByUser(2, time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Test with user context
	t.Run("should allow requests within rate limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), userIDKey, "user123")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// Second request should also be allowed
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should rate limit after exceeding limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := context.WithValue(req.Context(), userIDKey, "user456")
		req = req.WithContext(ctx)

		// Make requests up to the limit
		for i := 0; i < 2; i++ {
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Next request should be rate limited
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
	})

	t.Run("should skip rate limiting without user context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
