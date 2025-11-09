package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
)

const (
	tlsVersion12 = "1.2"
	tlsVersion13 = "1.3"
)

type TLSConfig struct {
	Enabled            bool     `yaml:"enabled" envconfig:"TLS_ENABLED" default:"false"`
	CertFile           string   `yaml:"cert_file" envconfig:"TLS_CERT_FILE" default:"./certs/server.crt"`
	KeyFile            string   `yaml:"key_file" envconfig:"TLS_KEY_FILE" default:"./certs/server.key"`
	CAFile             string   `yaml:"ca_file" envconfig:"TLS_CA_FILE"`
	ClientAuth         string   `yaml:"client_auth" envconfig:"TLS_CLIENT_AUTH" default:"none"`
	MinVersion         string   `yaml:"min_version" envconfig:"TLS_MIN_VERSION" default:"1.2"`
	MaxVersion         string   `yaml:"max_version" envconfig:"TLS_MAX_VERSION" default:"1.3"`
	CipherSuites       []string `yaml:"cipher_suites" envconfig:"TLS_CIPHER_SUITES"`
	InsecureSkipVerify bool     `yaml:"insecure_skip_verify" envconfig:"TLS_INSECURE_SKIP_VERIFY" default:"false"`
	ServerName         string   `yaml:"server_name" envconfig:"TLS_SERVER_NAME"`

	// mTLS Configuration
	ClientCerts                []string `yaml:"client_certs" envconfig:"TLS_CLIENT_CERTS"`
	RequireAndVerifyClientCert bool     `yaml:"require_client_cert" envconfig:"TLS_REQUIRE_CLIENT_CERT" default:"false"`

	// Certificate rotation
	AutoReload     bool          `yaml:"auto_reload" envconfig:"TLS_AUTO_RELOAD" default:"true"`
	ReloadInterval time.Duration `yaml:"reload_interval" envconfig:"TLS_RELOAD_INTERVAL" default:"1h"`
}

type TLSManager struct {
	config      *TLSConfig
	logger      *zap.Logger
	tlsConfig   *tls.Config
	certWatcher *certWatcher
}

type certWatcher struct {
	certFile     string
	keyFile      string
	lastModified time.Time
	ticker       *time.Ticker
	stopCh       chan struct{}
	reloadFn     func()
}

func NewTLSManager(config *TLSConfig, logger *zap.Logger) (*TLSManager, error) {
	if !config.Enabled {
		logger.Info("TLS is disabled")
		return &TLSManager{
			config: config,
			logger: logger,
		}, nil
	}

	manager := &TLSManager{
		config: config,
		logger: logger,
	}

	// Load initial TLS configuration
	if err := manager.loadTLSConfig(); err != nil {
		return nil, fmt.Errorf("failed to load TLS configuration: %w", err)
	}

	// Start certificate watcher if auto-reload is enabled
	if config.AutoReload {
		manager.startCertWatcher()
	}

	logger.Info("TLS manager initialized",
		zap.String("cert_file", config.CertFile),
		zap.String("key_file", config.KeyFile),
		zap.String("client_auth", config.ClientAuth),
		zap.String("min_version", config.MinVersion),
		zap.Bool("auto_reload", config.AutoReload))

	return manager, nil
}

func (tm *TLSManager) loadTLSConfig() error {
	if !tm.config.Enabled {
		return nil
	}

	// Load certificate and key
	cert, err := tls.LoadX509KeyPair(tm.config.CertFile, tm.config.KeyFile)
	if err != nil {
		return fmt.Errorf("failed to load certificate pair: %w", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates:             []tls.Certificate{cert},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.X25519,
			tls.CurveP256,
			tls.CurveP384,
			tls.CurveP521,
		},
	}

	// Set TLS version constraints
	if err := tm.setTLSVersions(tlsConfig); err != nil {
		return fmt.Errorf("failed to set TLS versions: %w", err)
	}

	// Set cipher suites
	if err := tm.setCipherSuites(tlsConfig); err != nil {
		return fmt.Errorf("failed to set cipher suites: %w", err)
	}

	// Configure client authentication
	if err := tm.configureClientAuth(tlsConfig); err != nil {
		return fmt.Errorf("failed to configure client auth: %w", err)
	}

	// Set server name
	if tm.config.ServerName != "" {
		tlsConfig.ServerName = tm.config.ServerName
	}

	// Set insecure skip verify (for development only)
	tlsConfig.InsecureSkipVerify = tm.config.InsecureSkipVerify
	if tm.config.InsecureSkipVerify {
		tm.logger.Warn("TLS certificate verification is disabled - not recommended for production")
	}

	tm.tlsConfig = tlsConfig
	tm.logger.Info("TLS configuration loaded successfully")
	return nil
}

func (tm *TLSManager) setTLSVersions(tlsConfig *tls.Config) error {
	// Set minimum TLS version
	switch tm.config.MinVersion {
	case "1.0":
		tlsConfig.MinVersion = tls.VersionTLS10
	case "1.1":
		tlsConfig.MinVersion = tls.VersionTLS11
	case tlsVersion12:
		tlsConfig.MinVersion = tls.VersionTLS12
	case tlsVersion13:
		tlsConfig.MinVersion = tls.VersionTLS13
	default:
		return fmt.Errorf("unsupported minimum TLS version: %s", tm.config.MinVersion)
	}

	// Set maximum TLS version
	switch tm.config.MaxVersion {
	case "1.0":
		tlsConfig.MaxVersion = tls.VersionTLS10
	case "1.1":
		tlsConfig.MaxVersion = tls.VersionTLS11
	case tlsVersion12:
		tlsConfig.MaxVersion = tls.VersionTLS12
	case tlsVersion13:
		tlsConfig.MaxVersion = tls.VersionTLS13
	default:
		return fmt.Errorf("unsupported maximum TLS version: %s", tm.config.MaxVersion)
	}

	if tlsConfig.MinVersion > tlsConfig.MaxVersion {
		return fmt.Errorf("minimum TLS version (%s) is higher than maximum (%s)",
			tm.config.MinVersion, tm.config.MaxVersion)
	}

	return nil
}

func (tm *TLSManager) setCipherSuites(tlsConfig *tls.Config) error {
	if len(tm.config.CipherSuites) == 0 {
		// Use secure default cipher suites for TLS 1.2
		tlsConfig.CipherSuites = []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		}
		return nil
	}

	// Parse configured cipher suites
	cipherSuiteMap := map[string]uint16{
		"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":         tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256":   tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":         tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		"TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":       tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		"TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256": tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
		"TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":       tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	}

	var cipherSuites []uint16
	for _, suite := range tm.config.CipherSuites {
		if cipherSuite, exists := cipherSuiteMap[suite]; exists {
			cipherSuites = append(cipherSuites, cipherSuite)
		} else {
			tm.logger.Warn("Unknown cipher suite ignored", zap.String("cipher_suite", suite))
		}
	}

	if len(cipherSuites) == 0 {
		return fmt.Errorf("no valid cipher suites configured")
	}

	tlsConfig.CipherSuites = cipherSuites
	return nil
}

func (tm *TLSManager) configureClientAuth(tlsConfig *tls.Config) error {
	switch tm.config.ClientAuth {
	case "none", "":
		tlsConfig.ClientAuth = tls.NoClientCert
	case "request":
		tlsConfig.ClientAuth = tls.RequestClientCert
	case "require":
		tlsConfig.ClientAuth = tls.RequireAnyClientCert
	case "verify":
		tlsConfig.ClientAuth = tls.VerifyClientCertIfGiven
	case "require-and-verify":
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	default:
		return fmt.Errorf("unsupported client auth mode: %s", tm.config.ClientAuth)
	}

	// Load CA certificate for client verification
	if tm.config.CAFile != "" && tlsConfig.ClientAuth != tls.NoClientCert {
		caCert, err := os.ReadFile(tm.config.CAFile)
		if err != nil {
			return fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return fmt.Errorf("failed to parse CA certificate")
		}

		tlsConfig.ClientCAs = caCertPool
		tm.logger.Info("CA certificate loaded for client verification",
			zap.String("ca_file", tm.config.CAFile))
	}

	// Load additional client certificates
	for _, certFile := range tm.config.ClientCerts {
		clientCert, err := os.ReadFile(certFile)
		if err != nil {
			tm.logger.Warn("Failed to read client certificate",
				zap.String("cert_file", certFile),
				zap.Error(err))
			continue
		}

		if tlsConfig.ClientCAs == nil {
			tlsConfig.ClientCAs = x509.NewCertPool()
		}

		if !tlsConfig.ClientCAs.AppendCertsFromPEM(clientCert) {
			tm.logger.Warn("Failed to parse client certificate",
				zap.String("cert_file", certFile))
		} else {
			tm.logger.Info("Client certificate loaded",
				zap.String("cert_file", certFile))
		}
	}

	return nil
}

func (tm *TLSManager) startCertWatcher() {
	if tm.certWatcher != nil {
		return // Already watching
	}

	// Get initial file modification times
	certStat, err := os.Stat(tm.config.CertFile)
	if err != nil {
		tm.logger.Error("Failed to stat certificate file", zap.Error(err))
		return
	}

	keyStat, err := os.Stat(tm.config.KeyFile)
	if err != nil {
		tm.logger.Error("Failed to stat key file", zap.Error(err))
		return
	}

	lastModified := certStat.ModTime()
	if keyStat.ModTime().After(lastModified) {
		lastModified = keyStat.ModTime()
	}

	tm.certWatcher = &certWatcher{
		certFile:     tm.config.CertFile,
		keyFile:      tm.config.KeyFile,
		lastModified: lastModified,
		ticker:       time.NewTicker(tm.config.ReloadInterval),
		stopCh:       make(chan struct{}),
		reloadFn: func() {
			if err := tm.loadTLSConfig(); err != nil {
				tm.logger.Error("Failed to reload TLS configuration", zap.Error(err))
			} else {
				tm.logger.Info("TLS configuration reloaded successfully")
			}
		},
	}

	go tm.runCertWatcher()
	tm.logger.Info("Certificate watcher started",
		zap.Duration("interval", tm.config.ReloadInterval))
}

func (tm *TLSManager) runCertWatcher() {
	for {
		select {
		case <-tm.certWatcher.ticker.C:
			tm.checkAndReloadCerts()
		case <-tm.certWatcher.stopCh:
			tm.certWatcher.ticker.Stop()
			return
		}
	}
}

func (tm *TLSManager) checkAndReloadCerts() {
	certStat, err := os.Stat(tm.certWatcher.certFile)
	if err != nil {
		tm.logger.Error("Failed to stat certificate file", zap.Error(err))
		return
	}

	keyStat, err := os.Stat(tm.certWatcher.keyFile)
	if err != nil {
		tm.logger.Error("Failed to stat key file", zap.Error(err))
		return
	}

	lastModified := certStat.ModTime()
	if keyStat.ModTime().After(lastModified) {
		lastModified = keyStat.ModTime()
	}

	if lastModified.After(tm.certWatcher.lastModified) {
		tm.logger.Info("Certificate files changed, reloading TLS configuration")
		tm.certWatcher.lastModified = lastModified
		tm.certWatcher.reloadFn()
	}
}

// GetTLSConfig returns the current TLS configuration
func (tm *TLSManager) GetTLSConfig() *tls.Config {
	if tm.tlsConfig == nil {
		return nil
	}

	// Return a copy to prevent external modifications
	config := tm.tlsConfig.Clone()
	return config
}

// IsEnabled returns whether TLS is enabled
func (tm *TLSManager) IsEnabled() bool {
	return tm.config.Enabled
}

// Stop stops the certificate watcher
func (tm *TLSManager) Stop() {
	if tm.certWatcher != nil {
		close(tm.certWatcher.stopCh)
		tm.certWatcher = nil
		tm.logger.Info("Certificate watcher stopped")
	}
}

// ValidateConfig validates the TLS configuration
func (config *TLSConfig) ValidateConfig() error {
	if !config.Enabled {
		return nil
	}

	// Check if certificate and key files exist
	if config.CertFile == "" {
		return fmt.Errorf("certificate file is required when TLS is enabled")
	}

	if config.KeyFile == "" {
		return fmt.Errorf("key file is required when TLS is enabled")
	}

	if _, err := os.Stat(config.CertFile); os.IsNotExist(err) {
		return fmt.Errorf("certificate file does not exist: %s", config.CertFile)
	}

	if _, err := os.Stat(config.KeyFile); os.IsNotExist(err) {
		return fmt.Errorf("key file does not exist: %s", config.KeyFile)
	}

	// Validate TLS versions
	validVersions := []string{"1.0", "1.1", tlsVersion12, tlsVersion13}
	if !contains(validVersions, config.MinVersion) {
		return fmt.Errorf("invalid minimum TLS version: %s", config.MinVersion)
	}

	if !contains(validVersions, config.MaxVersion) {
		return fmt.Errorf("invalid maximum TLS version: %s", config.MaxVersion)
	}

	// Validate client auth mode
	validAuthModes := []string{"none", "request", "require", "verify", "require-and-verify"}
	if !contains(validAuthModes, config.ClientAuth) {
		return fmt.Errorf("invalid client auth mode: %s", config.ClientAuth)
	}

	// Check CA file if client auth is enabled
	if config.ClientAuth != "none" && config.ClientAuth != "" && config.CAFile != "" {
		if _, err := os.Stat(config.CAFile); os.IsNotExist(err) {
			return fmt.Errorf("CA file does not exist: %s", config.CAFile)
		}
	}

	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
