package security

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Mock OPA Service
type MockOPAService struct {
	mock.Mock
}

func (m *MockOPAService) IsAuthorized(ctx context.Context, claims *Claims, method, path string) bool {
	args := m.Called(ctx, claims, method, path)
	return args.Bool(0)
}

func (m *MockOPAService) IsAuthorizedForResource(ctx context.Context, claims *Claims, resource, action string) bool {
	args := m.Called(ctx, claims, resource, action)
	return args.Bool(0)
}

func (m *MockOPAService) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func TestNewAuthService(t *testing.T) {
	config := AuthConfig{
		Mode:     "jwt",
		JWKSUrl:  "",
		Issuer:   "test-issuer",   // TEST_ISSUER - safe test value
		Audience: "test-audience", // TEST_AUDIENCE - safe test value
	}
	logger := zap.NewNop()
	opa := &MockOPAService{}

	authService := NewAuthService(config, logger, opa)

	assert.NotNil(t, authService)
	assert.Equal(t, config, authService.config)
	assert.Equal(t, logger, authService.logger)
	assert.Equal(t, opa, authService.opa)
	assert.NotNil(t, authService.publicKeys)
}

func TestJWTMiddleware_Success(t *testing.T) {
	// Setup
	config := AuthConfig{
		Issuer:   "test-issuer",   // TEST_ISSUER - safe test value
		Audience: "test-audience", // TEST_AUDIENCE - safe test value
	}
	logger := zap.NewNop()
	opa := &MockOPAService{}

	authService := NewAuthService(config, logger, opa)

	// Create a test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Add the public key to the auth service
	authService.publicKeys["test-kid"] = &privateKey.PublicKey // TEST_KEY_ID - safe test value

	// Create test claims
	claims := &Claims{
		UserID:   "user123",
		Email:    "test@example.com",
		Role:     "user",
		TenantID: "tenant123",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",                     // TEST_ISSUER - safe test value
			Audience:  jwt.ClaimStrings{"test-audience"}, // TEST_AUDIENCE - safe test value
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-kid" // TEST_KEY_ID - safe test value
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// Mock OPA authorization
	opa.On("IsAuthorized", mock.Anything, mock.AnythingOfType("*security.Claims"), "GET", "/api/v1/tasks").Return(true, nil)

	// Create test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify user context is set
		user, err := GetUserFromContext(r.Context())
		assert.NoError(t, err)
		assert.Equal(t, "user123", user.UserID)

		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write([]byte("success")); writeErr != nil {
			t.Logf("Warning: failed to write response: %v", writeErr)
		}
	})

	// Create request with JWT token
	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create recorder
	rr := httptest.NewRecorder()

	// Execute middleware
	middleware := authService.JWTMiddleware(testHandler)
	middleware.ServeHTTP(rr, req)

	// Assert response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "success", rr.Body.String())

	// Verify headers are set
	assert.Equal(t, "user123", rr.Header().Get("X-User-ID"))
	assert.Equal(t, "tenant123", rr.Header().Get("X-Tenant-ID"))

	opa.AssertExpectations(t)
}

func TestJWTMiddleware_NoAuthHeader(t *testing.T) {
	config := AuthConfig{
		Issuer:   "test-issuer",   // TEST_ISSUER - safe test value
		Audience: "test-audience", // TEST_AUDIENCE - safe test value
	}
	logger := zap.NewNop()
	opa := &MockOPAService{}

	authService := NewAuthService(config, logger, opa)

	testHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("Handler should not be called")
	})

	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	rr := httptest.NewRecorder()

	middleware := authService.JWTMiddleware(testHandler)
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	config := AuthConfig{
		Issuer:   "test-issuer",   // TEST_ISSUER - safe test value
		Audience: "test-audience", // TEST_AUDIENCE - safe test value
	}
	logger := zap.NewNop()
	opa := &MockOPAService{}

	authService := NewAuthService(config, logger, opa)

	testHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("Handler should not be called")
	})

	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	middleware := authService.JWTMiddleware(testHandler)
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestJWTMiddleware_AuthorizationDenied(t *testing.T) {
	config := AuthConfig{
		Issuer:   "test-issuer",   // TEST_ISSUER - safe test value
		Audience: "test-audience", // TEST_AUDIENCE - safe test value
	}
	logger := zap.NewNop()
	opa := &MockOPAService{}

	authService := NewAuthService(config, logger, opa)

	// Create a test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	authService.publicKeys["test-kid"] = &privateKey.PublicKey // TEST_KEY_ID - safe test value

	// Create test claims
	claims := &Claims{
		UserID:   "user123",
		Email:    "test@example.com",
		Role:     "user",
		TenantID: "tenant123",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "test-issuer",                     // TEST_ISSUER - safe test value
			Audience:  jwt.ClaimStrings{"test-audience"}, // TEST_AUDIENCE - safe test value
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create and sign token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-kid" // TEST_KEY_ID - safe test value
	tokenString, err := token.SignedString(privateKey)
	require.NoError(t, err)

	// Mock OPA authorization - return false
	opa.On("IsAuthorized", mock.Anything, mock.AnythingOfType("*security.Claims"), "GET", "/api/v1/tasks").Return(false)

	testHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("Handler should not be called")
	})

	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	middleware := authService.JWTMiddleware(testHandler)
	middleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	opa.AssertExpectations(t)
}

func TestJWTMiddleware_HealthEndpoints(t *testing.T) {
	config := AuthConfig{
		Issuer:   "test-issuer",   // TEST_ISSUER - safe test value
		Audience: "test-audience", // TEST_AUDIENCE - safe test value
	}
	logger := zap.NewNop()
	opa := &MockOPAService{}

	authService := NewAuthService(config, logger, opa)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write([]byte("healthy")); writeErr != nil {
			t.Logf("Warning: failed to write response: %v", writeErr)
		}
	})

	healthEndpoints := []string{"/healthz", "/readyz", "/metrics"}

	for _, endpoint := range healthEndpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			rr := httptest.NewRecorder()

			middleware := authService.JWTMiddleware(testHandler)
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "healthy", rr.Body.String())
		})
	}
}

func TestJWKToRSA(t *testing.T) {
	config := AuthConfig{}
	logger := zap.NewNop()
	opa := &MockOPAService{}

	authService := NewAuthService(config, logger, opa)

	// TEST ONLY: JWK parameters for unit testing RSA validation
	// These are public test values from RFC 7517 examples - NOT secret keys
	n := "0vx7agoebGcQSuuPiLJXZptN9nndrQmbXEps2aiAFbWhM78LhWx4cbbfAAtVT86zwu1RK7aPFFxuhDR1L6tSoc_BJECPebWKRXjBZCiFV4n3oknjhMstn64tZ_2W-5JsGY4Hc5n9yBXArwl93lqt7_RN5w6Cf0h4QyQ5v-65YGjQR0_FDW2QvzqY368QQMicAtaSqzs8KJZgnYb9c7d0zgdAZHzu6qMQvRL5hajrn1n91CbOpbISD08qNLyrdkt-bFTWhAI4vMQFh6WeZu0fM4lFd2NcRwr3XPksINHaQ-G_xBniIqbw0Ls1jF44-csFCur-kEgU8awapJzKnqDKgw" // TEST_RSA_MODULUS - RFC 7517 example
	e := "AQAB"                                                                                                                                                                                                                                                                                                                                                   // TEST_RSA_EXPONENT - Standard exponent

	publicKey, err := authService.jwkToRSA(n, e)

	assert.NoError(t, err)
	assert.NotNil(t, publicKey)
	assert.IsType(t, &rsa.PublicKey{}, publicKey)
}

func TestRequireScope(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write([]byte("authorized")); writeErr != nil {
			t.Logf("Warning: failed to write response: %v", writeErr)
		}
	})

	// Test with user having required scope
	t.Run("HasRequiredScope", func(t *testing.T) {
		claims := &Claims{
			UserID: "user123",
			Scopes: []string{"tasks:read", "tasks:write"},
		}

		req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
		ctx := context.WithValue(req.Context(), userKey, claims)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		middleware := RequireScope("tasks:read")
		handler := middleware(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "authorized", rr.Body.String())
	})

	// Test with user missing required scope
	t.Run("MissingRequiredScope", func(t *testing.T) {
		claims := &Claims{
			UserID: "user123",
			Scopes: []string{"tasks:read"},
		}

		req := httptest.NewRequest("DELETE", "/api/v1/tasks/123", nil)
		ctx := context.WithValue(req.Context(), userKey, claims)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		middleware := RequireScope("tasks:delete")
		handler := middleware(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	// Test with no user context
	t.Run("NoUserContext", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
		rr := httptest.NewRecorder()

		middleware := RequireScope("tasks:read")
		handler := middleware(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})
}

func TestRequireRole(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write([]byte("authorized")); writeErr != nil {
			t.Logf("Warning: failed to write response: %v", writeErr)
		}
	})

	// Test with user having required role
	t.Run("HasRequiredRole", func(t *testing.T) {
		claims := &Claims{
			UserID: "user123",
			Role:   "admin",
		}

		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		ctx := context.WithValue(req.Context(), userKey, claims)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		middleware := RequireRole("admin")
		handler := middleware(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "authorized", rr.Body.String())
	})

	// Test with user having insufficient role
	t.Run("InsufficientRole", func(t *testing.T) {
		claims := &Claims{
			UserID: "user123",
			Role:   "user",
		}

		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		ctx := context.WithValue(req.Context(), userKey, claims)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		middleware := RequireRole("admin")
		handler := middleware(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	// Test admin override
	t.Run("AdminOverride", func(t *testing.T) {
		claims := &Claims{
			UserID: "test_admin_user",
			Role:   "admin",
		}

		req := httptest.NewRequest("GET", "/api/v1/manager", nil)
		ctx := context.WithValue(req.Context(), userKey, claims)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()

		middleware := RequireRole("manager")
		handler := middleware(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "authorized", rr.Body.String())
	})
}

func TestGetUserFromContext(t *testing.T) {
	// Test with valid user context
	t.Run("ValidUserContext", func(t *testing.T) {
		claims := &Claims{
			UserID: "user123",
			Email:  "test@example.com",
			Role:   "user",
		}

		ctx := context.WithValue(context.Background(), userKey, claims)

		user, err := GetUserFromContext(ctx)
		assert.NoError(t, err)
		assert.Equal(t, claims, user)
	})

	// Test with no user context
	t.Run("NoUserContext", func(t *testing.T) {
		ctx := context.Background()

		user, err := GetUserFromContext(ctx)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found in context")
	})

	// Test with invalid user context type
	t.Run("InvalidUserContext", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), userKey, "invalid")

		user, err := GetUserFromContext(ctx)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user not found in context")
	})
}
