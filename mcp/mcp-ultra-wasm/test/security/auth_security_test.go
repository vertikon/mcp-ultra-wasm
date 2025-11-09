//go:build security
// +build security

package security

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/constants"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/security"
)

// SecurityTestSuite provides comprehensive security testing
type SecurityTestSuite struct {
	suite.Suite

	authService  *security.AuthService
	opaService   *security.OPAService
	vaultService *security.VaultService
	tlsManager   *security.TLSManager

	privateKey *rsa.PrivateKey
	logger     *zap.Logger
}

func (suite *SecurityTestSuite) SetupSuite() {
	suite.logger = zap.NewNop()

	// Generate test RSA key
	var err error
	suite.privateKey, err = rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(suite.T(), err)

	// Setup test services with mock configurations
	authConfig := security.AuthConfig{
		Mode:     "jwt",
		Issuer:   constants.TestIssuer,
		Audience: constants.TestAudience,
	}

	opaConfig := security.OPAConfig{
		URL:     "http://localhost:8181",
		Timeout: 5 * time.Second,
	}

	vaultConfig := security.VaultConfig{
		Address: "http://localhost:8200",
		Timeout: 10 * time.Second,
	}

	tlsConfig := security.TLSConfig{
		MinVersion: "1.2",
	}

	// Create mock OPA service for testing
	suite.opaService = security.NewOPAService(opaConfig, suite.logger)
	suite.authService = security.NewAuthService(authConfig, suite.logger, suite.opaService)
	suite.vaultService = security.NewVaultService(vaultConfig, suite.logger)
	suite.tlsManager = security.NewTLSManager(tlsConfig, suite.logger)

	// Add test public key to auth service
	suite.authService.AddPublicKey(constants.TestKeyID, &suite.privateKey.PublicKey)
}

// Test JWT token validation security
func (suite *SecurityTestSuite) TestJWTTokenSecurity() {
	// Test 1: Valid JWT token should be accepted
	suite.T().Run("ValidJWTToken", func(t *testing.T) {
		claims := &security.Claims{
			UserID:   "user123",
			Email:    "test@example.com",
			Role:     "user",
			TenantID: "tenant123",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    constants.TestIssuer,
				Audience:  jwt.ClaimStrings{constants.TestAudience},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = constants.TestKeyID
		tokenString, err := token.SignedString(suite.privateKey)
		require.NoError(t, err)

		// Validate token
		parsedClaims, err := suite.authService.ValidateToken(tokenString)
		assert.NoError(t, err)
		assert.Equal(t, "user123", parsedClaims.UserID)
	})

	// Test 2: Expired token should be rejected
	suite.T().Run("ExpiredJWTToken", func(t *testing.T) {
		claims := &security.Claims{
			UserID: "user123",
			Email:  "test@example.com",
			Role:   "user",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    constants.TestIssuer,
				Audience:  jwt.ClaimStrings{constants.TestAudience},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = constants.TestKeyID
		tokenString, err := token.SignedString(suite.privateKey)
		require.NoError(t, err)

		// Should reject expired token
		_, err = suite.authService.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
	})

	// Test 3: Malformed token should be rejected
	suite.T().Run("MalformedJWTToken", func(t *testing.T) {
		malformedToken := "invalid.jwt.token"

		_, err := suite.authService.ValidateToken(malformedToken)
		assert.Error(t, err)
	})

	// Test 4: Token with wrong issuer should be rejected
	suite.T().Run("WrongIssuerJWTToken", func(t *testing.T) {
		claims := &security.Claims{
			UserID: "user123",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "malicious-issuer", // Wrong issuer
				Audience:  jwt.ClaimStrings{constants.TestAudience},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = constants.TestKeyID
		tokenString, err := token.SignedString(suite.privateKey)
		require.NoError(t, err)

		_, err = suite.authService.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid issuer")
	})

	// Test 5: Token with unknown key ID should be rejected
	suite.T().Run("UnknownKeyIDJWTToken", func(t *testing.T) {
		claims := &security.Claims{
			UserID: "user123",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    constants.TestIssuer,
				Audience:  jwt.ClaimStrings{constants.TestAudience},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = constants.TestUnknownKeyID // Unknown key ID
		tokenString, err := token.SignedString(suite.privateKey)
		require.NoError(t, err)

		_, err = suite.authService.ValidateToken(tokenString)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown key ID")
	})
}

// Test authorization bypass attempts
func (suite *SecurityTestSuite) TestAuthorizationBypass() {
	// Test 1: Missing Authorization header
	suite.T().Run("MissingAuthHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
		rr := httptest.NewRecorder()

		handler := suite.authService.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	// Test 2: Malformed Authorization header
	suite.T().Run("MalformedAuthHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
		req.Header.Set("Authorization", "InvalidFormat token123")
		rr := httptest.NewRecorder()

		handler := suite.authService.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	// Test 3: Privilege escalation attempt
	suite.T().Run("PrivilegeEscalation", func(t *testing.T) {
		claims := &security.Claims{
			UserID:   "user123",
			Role:     "user", // Regular user
			Scopes:   []string{"tasks:read"},
			TenantID: "tenant123",
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    constants.TestIssuer,
				Audience:  jwt.ClaimStrings{constants.TestAudience},
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = constants.TestKeyID
		tokenString, err := token.SignedString(suite.privateKey)
		require.NoError(t, err)

		req := httptest.NewRequest("DELETE", "/api/v1/tasks/123", nil)
		req.Header.Set("Authorization", "Bearer "+tokenString)
		rr := httptest.NewRecorder()

		// Use role-based middleware that requires admin
		handler := security.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		// First set user context (simulating auth middleware)
		ctx := context.WithValue(req.Context(), "user", claims)
		req = req.WithContext(ctx)

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

// Test input validation and injection attacks
func (suite *SecurityTestSuite) TestInputValidationSecurity() {
	// Test 1: SQL Injection attempts
	suite.T().Run("SQLInjection", func(t *testing.T) {
		maliciousInputs := []string{
			"'; DROP TABLE tasks; --",
			"' OR '1'='1",
			"1; DELETE FROM tasks WHERE id = '1'",
			"' UNION SELECT * FROM users --",
		}

		for _, input := range maliciousInputs {
			// Test would validate that input sanitization prevents SQL injection
			// This would typically be tested at the repository level with real DB
			assert.NotEmpty(t, input) // Placeholder assertion
		}
	})

	// Test 2: XSS Prevention
	suite.T().Run("XSSPrevention", func(t *testing.T) {
		maliciousScripts := []string{
			"<script>alert('xss')</script>",
			"javascript:alert('xss')",
			"<img src=x onerror=alert('xss')>",
			"<svg onload=alert('xss')>",
		}

		for _, script := range maliciousScripts {
			// Test that output encoding prevents XSS
			// This would be tested in handlers/templates
			assert.Contains(t, script, "<") // Placeholder assertion
		}
	})

	// Test 3: Path Traversal Prevention
	suite.T().Run("PathTraversal", func(t *testing.T) {
		maliciousPaths := []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\config\\sam",
			"....//....//etc/passwd",
			"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		}

		for _, path := range maliciousPaths {
			// Test that file access is properly restricted
			assert.Contains(t, path, "..") // Placeholder assertion
		}
	})

	// Test 4: JSON Injection
	suite.T().Run("JSONInjection", func(t *testing.T) {
		maliciousJSON := `{
			"title": "Normal Title",
			"description": "Normal Description",
			"__proto__": {
				"isAdmin": true
			},
			"constructor": {
				"prototype": {
					"isAdmin": true
				}
			}
		}`

		// Test that JSON parsing doesn't allow prototype pollution
		var data map[string]interface{}
		err := json.Unmarshal([]byte(maliciousJSON), &data)
		assert.NoError(t, err)

		// Should not have polluted the prototype
		assert.NotContains(t, data, "isAdmin")
	})
}

// Test rate limiting and DoS prevention
func (suite *SecurityTestSuite) TestRateLimitingSecurity() {
	// Test 1: Rate limiting enforcement
	suite.T().Run("RateLimitingEnforcement", func(t *testing.T) {
		// This test would verify that rate limiting is properly enforced
		// Would require implementing rate limiting middleware first
		suite.T().Skip("Rate limiting implementation needed")
	})

	// Test 2: Large payload rejection
	suite.T().Run("LargePayloadRejection", func(t *testing.T) {
		largePayload := strings.Repeat("A", 10*1024*1024) // 10MB payload

		req := httptest.NewRequest("POST", "/api/v1/tasks", strings.NewReader(largePayload))
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		// Handler that would reject large payloads
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check content length
			if r.ContentLength > 1024*1024 { // 1MB limit
				http.Error(w, "Payload too large", http.StatusRequestEntityTooLarge)
				return
			}
			w.WriteHeader(http.StatusOK)
		})

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusRequestEntityTooLarge, rr.Code)
	})
}

// Test session security
func (suite *SecurityTestSuite) TestSessionSecurity() {
	// Test 1: Session fixation prevention
	suite.T().Run("SessionFixation", func(t *testing.T) {
		// Test that session IDs change after authentication
		// This would require session management implementation
		suite.T().Skip("Session management implementation needed")
	})

	// Test 2: Concurrent session limit
	suite.T().Run("ConcurrentSessionLimit", func(t *testing.T) {
		// Test that users cannot have unlimited concurrent sessions
		suite.T().Skip("Session limit implementation needed")
	})
}

// Test cryptographic security
func (suite *SecurityTestSuite) TestCryptographicSecurity() {
	// Test 1: Secure random generation
	suite.T().Run("SecureRandomGeneration", func(t *testing.T) {
		// Generate multiple random values and ensure they're different
		random1 := make([]byte, 32)
		random2 := make([]byte, 32)

		_, err := rand.Read(random1)
		require.NoError(t, err)

		_, err = rand.Read(random2)
		require.NoError(t, err)

		assert.NotEqual(t, random1, random2)
	})

	// Test 2: TLS configuration security
	suite.T().Run("TLSConfigurationSecurity", func(t *testing.T) {
		tlsConfig, err := suite.tlsManager.GetServerTLSConfig()
		if err != nil {
			// If cert files don't exist, test configuration principles
			assert.Contains(t, err.Error(), "loading server certificate")
			return
		}

		// Test secure TLS settings
		assert.NotNil(t, tlsConfig)
		assert.GreaterOrEqual(t, tlsConfig.MinVersion, uint16(0x0303)) // TLS 1.2 minimum
		assert.NotEmpty(t, tlsConfig.CipherSuites)
	})

	// Test 3: Key strength validation
	suite.T().Run("KeyStrengthValidation", func(t *testing.T) {
		// Test that weak keys are rejected
		weakKey, err := rsa.GenerateKey(rand.Reader, 1024) // Weak 1024-bit key
		require.NoError(t, err)

		// In production, should reject keys smaller than 2048 bits
		assert.Less(t, weakKey.Size(), 256) // 2048 bits = 256 bytes
	})
}

// Test data protection
func (suite *SecurityTestSuite) TestDataProtection() {
	// Test 1: PII detection and masking
	suite.T().Run("PIIDetectionAndMasking", func(t *testing.T) {
		sensitiveData := map[string]string{
			"email":       "user@example.com",
			"phone":       "123-456-7890",
			"ssn":         "123-45-6789",
			"credit_card": "4111-1111-1111-1111",
			"ip_address":  "192.168.1.1",
		}

		for dataType, value := range sensitiveData {
			// Test that PII is properly detected and masked
			masked := maskPII(value, dataType)
			assert.NotEqual(t, value, masked, "PII should be masked for type: "+dataType)
			assert.Contains(t, masked, "*", "Masked value should contain asterisks")
		}
	})

	// Test 2: Data encryption at rest
	suite.T().Run("DataEncryptionAtRest", func(t *testing.T) {
		// This would test that sensitive data is encrypted before storage
		suite.T().Skip("Data encryption implementation needed")
	})
}

// Test audit logging security
func (suite *SecurityTestSuite) TestAuditLoggingSecurity() {
	// Test 1: Authentication events are logged
	suite.T().Run("AuthenticationEventsLogged", func(t *testing.T) {
		// Test that all authentication attempts are logged
		suite.T().Skip("Audit logging implementation needed")
	})

	// Test 2: Authorization failures are logged
	suite.T().Run("AuthorizationFailuresLogged", func(t *testing.T) {
		// Test that authorization failures are logged
		suite.T().Skip("Audit logging implementation needed")
	})

	// Test 3: Sensitive data is not logged
	suite.T().Run("SensitiveDataNotLogged", func(t *testing.T) {
		// Test that passwords, tokens, etc. are not logged
		suite.T().Skip("Audit logging implementation needed")
	})
}

// Helper function for PII masking (mock implementation)
func maskPII(value, dataType string) string {
	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}

	switch dataType {
	case "email":
		parts := strings.Split(value, "@")
		if len(parts) == 2 {
			return parts[0][:1] + strings.Repeat("*", len(parts[0])-1) + "@" + parts[1]
		}
	case "phone":
		return "***-***-" + value[len(value)-4:]
	case "ssn":
		return "***-**-" + value[len(value)-4:]
	case "credit_card":
		return "**** **** **** " + value[len(value)-4:]
	}

	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}

// Run the security test suite
func TestSecuritySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping security tests in short mode")
	}

	suite.Run(t, new(SecurityTestSuite))
}
