package security

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"
)

// TLSConfig holds TLS configuration
type TLSConfig struct {
	CertFile   string `yaml:"cert_file"`
	KeyFile    string `yaml:"key_file"`
	CAFile     string `yaml:"ca_file,omitempty"`
	ClientAuth bool   `yaml:"client_auth"`
	MinVersion string `yaml:"min_version"`
}

// TLSManager manages TLS configuration for secure communications
type TLSManager struct {
	config TLSConfig
	logger *zap.Logger
}

// NewTLSManager creates a new TLS manager
func NewTLSManager(config TLSConfig, logger *zap.Logger) *TLSManager {
	return &TLSManager{
		config: config,
		logger: logger,
	}
}

// GetServerTLSConfig returns TLS configuration for server
func (tm *TLSManager) GetServerTLSConfig() (*tls.Config, error) {
	// Load server certificate and key
	cert, err := tls.LoadX509KeyPair(tm.config.CertFile, tm.config.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("loading server certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tm.getTLSVersion(tm.config.MinVersion),
		CipherSuites: tm.getSecureCipherSuites(),
	}

	// Configure client authentication if enabled
	if tm.config.ClientAuth && tm.config.CAFile != "" {
		caCert, err := os.ReadFile(tm.config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("reading CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = caCertPool

		tm.logger.Info("mTLS enabled - client certificate authentication required")
	}

	tm.logger.Info("Server TLS configuration loaded",
		zap.String("min_version", tm.config.MinVersion),
		zap.Bool("client_auth", tm.config.ClientAuth))

	return tlsConfig, nil
}

// GetClientTLSConfig returns TLS configuration for client connections
func (tm *TLSManager) GetClientTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion:   tm.getTLSVersion(tm.config.MinVersion),
		CipherSuites: tm.getSecureCipherSuites(),
	}

	// Load client certificate if configured for mTLS
	if tm.config.CertFile != "" && tm.config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(tm.config.CertFile, tm.config.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Configure CA for server verification
	if tm.config.CAFile != "" {
		caCert, err := os.ReadFile(tm.config.CAFile)
		if err != nil {
			return nil, fmt.Errorf("reading CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}

// MTLSMiddleware validates client certificates
func (tm *TLSManager) MTLSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client certificate is present
		if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
			tm.logger.Warn("mTLS: No client certificate provided",
				zap.String("remote_addr", r.RemoteAddr))
			http.Error(w, "Client certificate required", http.StatusUnauthorized)
			return
		}

		// Get client certificate
		clientCert := r.TLS.PeerCertificates[0]

		// Extract client information
		clientID := tm.extractClientID(clientCert)
		if clientID == "" {
			tm.logger.Warn("mTLS: Unable to extract client ID from certificate",
				zap.String("subject", clientCert.Subject.String()))
			http.Error(w, "Invalid client certificate", http.StatusUnauthorized)
			return
		}

		// Add client information to request headers
		w.Header().Set("X-Client-ID", clientID)
		w.Header().Set("X-Client-Subject", clientCert.Subject.String())

		tm.logger.Debug("mTLS: Client authenticated",
			zap.String("client_id", clientID),
			zap.String("subject", clientCert.Subject.String()))

		next.ServeHTTP(w, r)
	})
}

// extractClientID extracts client identifier from certificate
func (tm *TLSManager) extractClientID(cert *x509.Certificate) string {
	// Try Common Name first
	if cert.Subject.CommonName != "" {
		return cert.Subject.CommonName
	}

	// Try Organization
	if len(cert.Subject.Organization) > 0 {
		return cert.Subject.Organization[0]
	}

	// Try Organizational Unit
	if len(cert.Subject.OrganizationalUnit) > 0 {
		return cert.Subject.OrganizationalUnit[0]
	}

	// Try Subject Alternative Names
	for _, dnsName := range cert.DNSNames {
		if dnsName != "" {
			return dnsName
		}
	}

	return ""
}

// getTLSVersion converts version string to tls constant
func (tm *TLSManager) getTLSVersion(version string) uint16 {
	switch version {
	case "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		tm.logger.Warn("Unknown TLS version, defaulting to 1.2", zap.String("version", version))
		return tls.VersionTLS12
	}
}

// getSecureCipherSuites returns secure cipher suites
func (tm *TLSManager) getSecureCipherSuites() []uint16 {
	return []uint16{
		// TLS 1.3 cipher suites (preferred)
		tls.TLS_AES_256_GCM_SHA384,
		tls.TLS_AES_128_GCM_SHA256,
		tls.TLS_CHACHA20_POLY1305_SHA256,

		// TLS 1.2 cipher suites
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
	}
}

// ValidateCertificate validates a certificate against custom rules
func (tm *TLSManager) ValidateCertificate(cert *x509.Certificate) error {
	// Check certificate validity period
	if cert.NotBefore.After(cert.NotAfter) {
		return fmt.Errorf("certificate has invalid validity period")
	}

	// Check key usage
	if cert.KeyUsage&x509.KeyUsageDigitalSignature == 0 {
		return fmt.Errorf("certificate missing required key usage: digital signature")
	}

	// Check extended key usage for client certificates
	hasClientAuth := false
	for _, usage := range cert.ExtKeyUsage {
		if usage == x509.ExtKeyUsageClientAuth {
			hasClientAuth = true
			break
		}
	}

	if !hasClientAuth {
		return fmt.Errorf("certificate missing required extended key usage: client authentication")
	}

	tm.logger.Debug("Certificate validation passed",
		zap.String("subject", cert.Subject.String()),
		zap.Time("not_before", cert.NotBefore),
		zap.Time("not_after", cert.NotAfter))

	return nil
}

// GetHTTPClient returns an HTTP client configured with TLS
func (tm *TLSManager) GetHTTPClient() (*http.Client, error) {
	tlsConfig, err := tm.GetClientTLSConfig()
	if err != nil {
		return nil, fmt.Errorf("creating client TLS config: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return client, nil
}
