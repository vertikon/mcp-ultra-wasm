package config

import (
	"crypto/tls"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const invalidValue = "invalid"

func TestTLSConfig_ValidateConfig(t *testing.T) {
	t.Run("should validate disabled TLS", func(t *testing.T) {
		config := &TLSConfig{
			Enabled: false,
		}

		err := config.ValidateConfig()
		assert.NoError(t, err)
	})

	t.Run("should require certificate file when enabled", func(t *testing.T) {
		config := &TLSConfig{
			Enabled:  true,
			CertFile: "",
		}

		err := config.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "certificate file is required")
	})

	t.Run("should require key file when enabled", func(t *testing.T) {
		config := &TLSConfig{
			Enabled:  true,
			CertFile: "cert.pem",
			KeyFile:  "",
		}

		err := config.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key file is required")
	})

	t.Run("should validate TLS versions", func(t *testing.T) {
		config := &TLSConfig{
			Enabled:    true,
			CertFile:   createTempFile(t, "cert", testCert),
			KeyFile:    createTempFile(t, "key", testKey),
			MinVersion: invalidValue,
		}

		err := config.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid minimum TLS version")
	})

	t.Run("should validate client auth mode", func(t *testing.T) {
		config := &TLSConfig{
			Enabled:    true,
			CertFile:   createTempFile(t, "cert", testCert),
			KeyFile:    createTempFile(t, "key", testKey),
			MinVersion: "1.2",
			MaxVersion: tlsVersion13,
			ClientAuth: invalidValue,
		}

		err := config.ValidateConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid client auth mode")
	})

	t.Run("should validate existing files", func(t *testing.T) {
		config := &TLSConfig{
			Enabled:    true,
			CertFile:   createTempFile(t, "cert", testCert),
			KeyFile:    createTempFile(t, "key", testKey),
			MinVersion: "1.2",
			MaxVersion: tlsVersion13,
			ClientAuth: "none",
		}

		err := config.ValidateConfig()
		assert.NoError(t, err)
	})
}

func TestNewTLSManager(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("should create manager with disabled TLS", func(t *testing.T) {
		config := &TLSConfig{
			Enabled: false,
		}

		manager, err := NewTLSManager(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, manager)
		assert.False(t, manager.IsEnabled())
		assert.Nil(t, manager.GetTLSConfig())
	})

	t.Run("should create manager with valid TLS config", func(t *testing.T) {
		certFile := createTempFile(t, "cert", testCert)
		keyFile := createTempFile(t, "key", testKey)

		config := &TLSConfig{
			Enabled:    true,
			CertFile:   certFile,
			KeyFile:    keyFile,
			MinVersion: "1.2",
			MaxVersion: tlsVersion13,
			ClientAuth: "none",
			AutoReload: false, // Disable for testing
		}

		manager, err := NewTLSManager(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, manager)
		assert.True(t, manager.IsEnabled())
		assert.NotNil(t, manager.GetTLSConfig())
	})

	t.Run("should fail with invalid certificate", func(t *testing.T) {
		certFile := createTempFile(t, "cert", "invalid-cert")
		keyFile := createTempFile(t, "key", testKey)

		config := &TLSConfig{
			Enabled:  true,
			CertFile: certFile,
			KeyFile:  keyFile,
		}

		manager, err := NewTLSManager(config, logger)
		assert.Error(t, err)
		assert.Nil(t, manager)
	})
}

func TestTLSManager_SetTLSVersions(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TLSConfig{Enabled: false}
	manager := &TLSManager{config: config, logger: logger}

	t.Run("should set valid TLS versions", func(t *testing.T) {
		tlsConfig := &tls.Config{}
		manager.config.MinVersion = tlsVersion12
		manager.config.MaxVersion = tlsVersion13

		err := manager.setTLSVersions(tlsConfig)
		assert.NoError(t, err)
		assert.Equal(t, uint16(tls.VersionTLS12), tlsConfig.MinVersion)
		assert.Equal(t, uint16(tls.VersionTLS13), tlsConfig.MaxVersion)
	})

	t.Run("should reject invalid minimum version", func(t *testing.T) {
		tlsConfig := &tls.Config{}
		manager.config.MinVersion = invalidValue
		manager.config.MaxVersion = tlsVersion13

		err := manager.setTLSVersions(tlsConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported minimum TLS version")
	})

	t.Run("should reject invalid maximum version", func(t *testing.T) {
		tlsConfig := &tls.Config{}
		manager.config.MinVersion = "1.2"
		manager.config.MaxVersion = invalidValue

		err := manager.setTLSVersions(tlsConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported maximum TLS version")
	})

	t.Run("should reject min version higher than max", func(t *testing.T) {
		tlsConfig := &tls.Config{}
		manager.config.MinVersion = tlsVersion13
		manager.config.MaxVersion = "1.2"

		err := manager.setTLSVersions(tlsConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "minimum TLS version")
		assert.Contains(t, err.Error(), "is higher than maximum")
	})
}

func TestTLSManager_SetCipherSuites(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TLSConfig{Enabled: false}
	manager := &TLSManager{config: config, logger: logger}

	t.Run("should use default cipher suites when none configured", func(t *testing.T) {
		tlsConfig := &tls.Config{}
		manager.config.CipherSuites = nil

		err := manager.setCipherSuites(tlsConfig)
		assert.NoError(t, err)
		assert.NotEmpty(t, tlsConfig.CipherSuites)
		// Should include secure defaults
		assert.Contains(t, tlsConfig.CipherSuites, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
	})

	t.Run("should set valid cipher suites", func(t *testing.T) {
		tlsConfig := &tls.Config{}
		manager.config.CipherSuites = []string{
			"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
		}

		err := manager.setCipherSuites(tlsConfig)
		assert.NoError(t, err)
		assert.Len(t, tlsConfig.CipherSuites, 2)
		assert.Contains(t, tlsConfig.CipherSuites, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
		assert.Contains(t, tlsConfig.CipherSuites, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
	})

	t.Run("should ignore unknown cipher suites", func(t *testing.T) {
		tlsConfig := &tls.Config{}
		manager.config.CipherSuites = []string{
			"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
			"UNKNOWN_CIPHER_SUITE",
		}

		err := manager.setCipherSuites(tlsConfig)
		assert.NoError(t, err)
		assert.Len(t, tlsConfig.CipherSuites, 1) // Only the valid one
	})

	t.Run("should fail with no valid cipher suites", func(t *testing.T) {
		tlsConfig := &tls.Config{}
		manager.config.CipherSuites = []string{
			"UNKNOWN_CIPHER_SUITE_1",
			"UNKNOWN_CIPHER_SUITE_2",
		}

		err := manager.setCipherSuites(tlsConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid cipher suites")
	})
}

func TestTLSManager_ConfigureClientAuth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &TLSConfig{Enabled: false}
	manager := &TLSManager{config: config, logger: logger}

	tests := []struct {
		clientAuth string
		expected   tls.ClientAuthType
	}{
		{"none", tls.NoClientCert},
		{"", tls.NoClientCert},
		{"request", tls.RequestClientCert},
		{"require", tls.RequireAnyClientCert},
		{"verify", tls.VerifyClientCertIfGiven},
		{"require-and-verify", tls.RequireAndVerifyClientCert},
	}

	for _, tt := range tests {
		t.Run(tt.clientAuth, func(t *testing.T) {
			tlsConfig := &tls.Config{}
			manager.config.ClientAuth = tt.clientAuth

			err := manager.configureClientAuth(tlsConfig)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, tlsConfig.ClientAuth)
		})
	}

	t.Run("should reject invalid client auth mode", func(t *testing.T) {
		tlsConfig := &tls.Config{}
		manager.config.ClientAuth = invalidValue

		err := manager.configureClientAuth(tlsConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported client auth mode")
	})
}

func TestTLSManager_GetTLSConfig(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("should return nil for disabled TLS", func(t *testing.T) {
		config := &TLSConfig{Enabled: false}
		manager := &TLSManager{config: config, logger: logger}

		tlsConfig := manager.GetTLSConfig()
		assert.Nil(t, tlsConfig)
	})

	t.Run("should return copy of TLS config", func(t *testing.T) {
		certFile := createTempFile(t, "cert", testCert)
		keyFile := createTempFile(t, "key", testKey)

		config := &TLSConfig{
			Enabled:    true,
			CertFile:   certFile,
			KeyFile:    keyFile,
			AutoReload: false,
		}

		manager, err := NewTLSManager(config, logger)
		require.NoError(t, err)

		tlsConfig1 := manager.GetTLSConfig()
		tlsConfig2 := manager.GetTLSConfig()

		assert.NotNil(t, tlsConfig1)
		assert.NotNil(t, tlsConfig2)
		// Should be different instances (copies)
		assert.NotSame(t, tlsConfig1, tlsConfig2)
	})
}

func TestTLSManager_Stop(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("should stop certificate watcher", func(t *testing.T) {
		certFile := createTempFile(t, "cert", testCert)
		keyFile := createTempFile(t, "key", testKey)

		config := &TLSConfig{
			Enabled:        true,
			CertFile:       certFile,
			KeyFile:        keyFile,
			AutoReload:     true,
			ReloadInterval: time.Millisecond, // Very short for testing
		}

		manager, err := NewTLSManager(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, manager.certWatcher)

		manager.Stop()
		assert.Nil(t, manager.certWatcher)
	})

	t.Run("should handle multiple stops", func(t *testing.T) {
		config := &TLSConfig{Enabled: false}
		manager := &TLSManager{config: config, logger: logger}

		// Should not panic
		manager.Stop()
		manager.Stop()
	})
}

// Helper function to create temporary files for testing
func createTempFile(t *testing.T, prefix, content string) string {
	file, err := os.CreateTemp("", prefix+"*.pem")
	require.NoError(t, err)

	_, err = file.WriteString(content)
	require.NoError(t, err)

	err = file.Close()
	require.NoError(t, err)

	// Clean up after test
	t.Cleanup(func() {
		_ = os.Remove(file.Name())
	})

	return file.Name()
}

// Test certificate and key (for testing purposes only)
const testCert = `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAOAOOOOOOOOOMA0GCSqGSIb3DQEBCwUAMEUxCzAJBgNV
BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
aWRnaXRzIFB0eSBMdGQwHhcNMjAwMTAxMDAwMDAwWhcNMzAwMTAxMDAwMDAwWjBF
MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
CgKCAQEAujQx/9SFHWZx3+fjVvg6HNjRoNZQQnA1Z7wl1OJ+3+HJjSFAcNHhAtUO
xSg5wKgYvYEUvfO2N9PgN1YW7TcdE5J5J0A8kP7w9U0Q1Jz7g2v2pV7j0Y2N8Fk9
ckYz2Kv2Y5QjWgIrKN9z7zK1FXtNQ5kKQJO5I/tF5lQJiN5YTI7QJ8y6J8VfKs8q
Lb9cP5FyQI8KJm9dYgQW3UvY8pQWuKF7LmL5X8sQJL7PwG2k5yHqW2O4K9vQ7w5o
qJ4F5C2d3v7E8h5wGgxU2qE8LgV9Q1v1W2K9L9xzBQK2o9K6QYqQ1W0oQ9V0B8Yc
V5yQo8XG7w0oYFP5YW6jY3oP3A1JgwIDAQABo1AwTjAdBgNVHQ4EFgQUA5Z5Q9L0
8YYlFzQ4vQ4k1Y6N9nQwHwYDVR0jBBgwFoAUA5Z5Q9L08YYlFzQ4vQ4k1Y6N9nQw
DAYDVR0TBAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAW6E0H/7SSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SS==
-----END CERTIFICATE-----`

const testKey = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC6NDH/1IUdZnHf
5+NW+Doc2NGg1lBCcDVnvCXU4n7f4cmNIUBw0eEC1Q7FKDnAqBi9gRS987Y30+A3
VhbtNx0TknknQDyQ/vD1TRDUnPuDa/alXuPRjY3wWT1yRjPYq/ZjlCNaAiso33Pv
MrUVe01DmQpAk7kj+0XmVAmI3lhMjtAnzLonxV8qzypbcJFWWWWWWWWWWWWWWWWW
WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW
WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW
WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW
WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW
WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWW
WWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWWg==
-----END PRIVATE KEY-----`
