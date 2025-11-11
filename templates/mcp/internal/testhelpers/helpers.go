package testhelpers

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

// GetTestJWTSecret returns a safe test JWT secret
func GetTestJWTSecret() string {
	return "TEST_jwt_secret_for_testing_only_not_secure_32bytes"
}

// GenerateTestSecret generates a random test secret
func GenerateTestSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "TEST_fallback_secret_" + hex.EncodeToString([]byte("test"))
	}
	return "TEST_" + hex.EncodeToString(bytes)
}

// GetTestDatabaseURL returns a test database URL
func GetTestDatabaseURL() string {
	return "postgres://TEST_user:TEST_pass@localhost:5432/TEST_db?sslmode=disable"
}

// GetTestRedisURL returns a test Redis URL
func GetTestRedisURL() string {
	return "redis://localhost:6379/0"
}

// GetTestNATSURL returns a test NATS URL
func GetTestNATSURL() string {
	return "nats://localhost:4222"
}

// GetTestAPIKeys returns test API keys for authentication testing
func GetTestAPIKeys(t *testing.T) (publicKey, privateKey string) {
	t.Helper()
	return "test-public-key-" + hex.EncodeToString([]byte("public")),
		"test-private-key-" + hex.EncodeToString([]byte("private"))
}
