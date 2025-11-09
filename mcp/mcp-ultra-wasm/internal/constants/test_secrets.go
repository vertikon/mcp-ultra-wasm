package constants

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
)

var (
	testSecrets     map[string]string
	testSecretsOnce sync.Once
)

// GetTestSecret returns a randomly generated test secret for the given key.
// Secrets are generated once per test run and cached.
func GetTestSecret(key string) string {
	testSecretsOnce.Do(func() {
		testSecrets = map[string]string{
			"jwt":     generateRandomSecret(32),
			"api_key": generateRandomSecret(24),
			"key_id":  generateRandomSecret(16),
		}
	})

	if secret, ok := testSecrets[key]; ok {
		return secret
	}
	return generateRandomSecret(32)
}

// generateRandomSecret creates a cryptographically random string of the specified byte length
func generateRandomSecret(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fallback to a deterministic value for tests if crypto/rand fails
		return "TEST_FALLBACK_" + string(rune(length))
	}
	return base64.URLEncoding.EncodeToString(b)
}

// ResetTestSecrets clears the cached secrets (useful for test isolation)
func ResetTestSecrets() {
	testSecrets = nil
	testSecretsOnce = sync.Once{}
}
