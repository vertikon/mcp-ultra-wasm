package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/security"
)

// Config represents the application configuration
type Config struct {
	Environment string           `yaml:"environment" envconfig:"ENVIRONMENT" default:"development"`
	Region      string           `yaml:"region" envconfig:"REGION" default:"us-east-1"`
	Datacenter  string           `yaml:"datacenter" envconfig:"DATACENTER" default:"dc1"`
	Server      ServerConfig     `yaml:"server"`
	GRPC        GRPCConfig       `yaml:"grpc"`
	Database    DatabaseConfig   `yaml:"database"`
	NATS        NATSConfig       `yaml:"nats"`
	Telemetry   TelemetryConfig  `yaml:"telemetry"`
	Features    FeaturesConfig   `yaml:"features"`
	Security    SecurityConfig   `yaml:"security"`
	Compliance  ComplianceConfig `yaml:"compliance"`
}

// ComplianceConfig holds all compliance-related configuration
type ComplianceConfig struct {
	Enabled       bool                `yaml:"enabled" envconfig:"COMPLIANCE_ENABLED" default:"true"`
	DefaultRegion string              `yaml:"default_region" envconfig:"DEFAULT_REGION" default:"BR"`
	PIIDetection  PIIDetectionConfig  `yaml:"pii_detection"`
	Consent       ConsentConfig       `yaml:"consent"`
	DataRetention DataRetentionConfig `yaml:"data_retention"`
	AuditLogging  AuditLoggingConfig  `yaml:"audit_logging"`
	LGPD          LGPDConfig          `yaml:"lgpd"`
	GDPR          GDPRConfig          `yaml:"gdpr"`
	Anonymization AnonymizationConfig `yaml:"anonymization"`
	DataRights    DataRightsConfig    `yaml:"data_rights"`
}

// PIIDetectionConfig configures PII detection and classification
type PIIDetectionConfig struct {
	Enabled           bool     `yaml:"enabled" default:"true"`
	ScanFields        []string `yaml:"scan_fields"`
	ClassificationAPI string   `yaml:"classification_api"`
	Confidence        float64  `yaml:"confidence" default:"0.8"`
	AutoMask          bool     `yaml:"auto_mask" default:"true"`
}

// ConsentConfig configures consent management
type ConsentConfig struct {
	Enabled         bool          `yaml:"enabled" default:"true"`
	DefaultPurposes []string      `yaml:"default_purposes"`
	TTL             time.Duration `yaml:"ttl" default:"2y"`
	GranularLevel   string        `yaml:"granular_level" default:"purpose"` // purpose, field, operation
}

// DataRetentionConfig configures data retention policies
type DataRetentionConfig struct {
	Enabled         bool                     `yaml:"enabled" default:"true"`
	DefaultPeriod   time.Duration            `yaml:"default_period" default:"2y"`
	CategoryPeriods map[string]time.Duration `yaml:"category_periods"`
	AutoDelete      bool                     `yaml:"auto_delete" default:"true"`
	BackupRetention time.Duration            `yaml:"backup_retention" default:"7y"`
}

// AuditLoggingConfig configures compliance audit logging
type AuditLoggingConfig struct {
	Enabled           bool          `yaml:"enabled" default:"true"`
	DetailLevel       string        `yaml:"detail_level" default:"full"` // minimal, standard, full
	RetentionPeriod   time.Duration `yaml:"retention_period" default:"7y"`
	EncryptionEnabled bool          `yaml:"encryption_enabled" default:"true"`
	ExternalLogging   bool          `yaml:"external_logging" default:"false"`
	ExternalEndpoint  string        `yaml:"external_endpoint"`
}

// LGPDConfig specific configuration for Brazilian LGPD compliance
type LGPDConfig struct {
	Enabled          bool     `yaml:"enabled" default:"true"`
	DPOContact       string   `yaml:"dpo_contact"`
	LegalBasis       []string `yaml:"legal_basis"`
	DataCategories   []string `yaml:"data_categories"`
	SharedThirdParty bool     `yaml:"shared_third_party" default:"false"`
}

// GDPRConfig specific configuration for European GDPR compliance
type GDPRConfig struct {
	Enabled             bool     `yaml:"enabled" default:"true"`
	DPOContact          string   `yaml:"dpo_contact"`
	LegalBasis          []string `yaml:"legal_basis"`
	DataCategories      []string `yaml:"data_categories"`
	CrossBorderTransfer bool     `yaml:"cross_border_transfer" default:"false"`
	AdequacyDecisions   []string `yaml:"adequacy_decisions"`
}

// AnonymizationConfig configures data anonymization
type AnonymizationConfig struct {
	Enabled    bool              `yaml:"enabled" default:"true"`
	Methods    []string          `yaml:"methods"` // hash, encrypt, tokenize, redact, generalize
	HashSalt   string            `yaml:"hash_salt"`
	Reversible bool              `yaml:"reversible" default:"false"`
	KAnonymity int               `yaml:"k_anonymity" default:"5"`
	Algorithms map[string]string `yaml:"algorithms"`
}

// DataRightsConfig configures individual data rights handling
type DataRightsConfig struct {
	Enabled              bool          `yaml:"enabled" default:"true"`
	ResponseTime         time.Duration `yaml:"response_time" default:"720h"` // 30 days
	AutoFulfillment      bool          `yaml:"auto_fulfillment" default:"false"`
	VerificationRequired bool          `yaml:"verification_required" default:"true"`
	NotificationChannels []string      `yaml:"notification_channels"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         int           `yaml:"port" envconfig:"HTTP_PORT" default:"9655"`
	ReadTimeout  time.Duration `yaml:"read_timeout" default:"30s"`
	WriteTimeout time.Duration `yaml:"write_timeout" default:"30s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" default:"120s"`
}

// GRPCConfig holds gRPC server configuration
type GRPCConfig struct {
	Port               int             `yaml:"port" envconfig:"GRPC_PORT" default:"9656"`
	MaxRecvMessageSize int             `yaml:"max_recv_message_size" default:"4194304"` // 4MB
	MaxSendMessageSize int             `yaml:"max_send_message_size" default:"4194304"` // 4MB
	ConnectionTimeout  time.Duration   `yaml:"connection_timeout" default:"30s"`
	ShutdownTimeout    time.Duration   `yaml:"shutdown_timeout" default:"30s"`
	Keepalive          KeepaliveConfig `yaml:"keepalive"`
}

// KeepaliveConfig holds gRPC keepalive configuration
type KeepaliveConfig struct {
	MaxConnectionIdle     time.Duration `yaml:"max_connection_idle" default:"15s"`
	MaxConnectionAge      time.Duration `yaml:"max_connection_age" default:"30s"`
	MaxConnectionAgeGrace time.Duration `yaml:"max_connection_age_grace" default:"5s"`
	Time                  time.Duration `yaml:"time" default:"5s"`
	Timeout               time.Duration `yaml:"timeout" default:"1s"`
	MinTime               time.Duration `yaml:"min_time" default:"10s"`
	PermitWithoutStream   bool          `yaml:"permit_without_stream" default:"false"`
}

// DatabaseConfig holds database connections configuration
type DatabaseConfig struct {
	PostgreSQL PostgreSQLConfig `yaml:"postgresql"`
	Redis      RedisConfig      `yaml:"redis"`
}

// PostgreSQLConfig holds PostgreSQL configuration
type PostgreSQLConfig struct {
	Host            string        `yaml:"host" envconfig:"POSTGRES_HOST" default:"localhost"`
	Port            int           `yaml:"port" envconfig:"POSTGRES_PORT" default:"5432"`
	Database        string        `yaml:"database" envconfig:"POSTGRES_DB" default:"mcp_ultra_wasm"`
	User            string        `yaml:"user" envconfig:"POSTGRES_USER" default:"postgres"`
	Password        string        `yaml:"password" envconfig:"POSTGRES_PASSWORD" default:"postgres"`
	SSLMode         string        `yaml:"ssl_mode" envconfig:"POSTGRES_SSLMODE" default:"disable"`
	MaxOpenConns    int           `yaml:"max_open_conns" default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" default:"5m"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string `yaml:"addr" envconfig:"REDIS_ADDR" default:"localhost:6379"`
	Password string `yaml:"password" envconfig:"REDIS_PASSWORD" default:""`
	DB       int    `yaml:"db" envconfig:"REDIS_DB" default:"0"`
	PoolSize int    `yaml:"pool_size" default:"10"`
}

// NATSConfig holds NATS configuration
type NATSConfig struct {
	URL       string `yaml:"url" envconfig:"NATS_URL" default:"nats://localhost:4222"`
	ClusterID string `yaml:"cluster_id" default:"mcp-cluster"`
	ClientID  string `yaml:"client_id" default:"mcp-ultra-wasm"`
}

// TelemetryConfig holds comprehensive telemetry configuration
type TelemetryConfig struct {
	Enabled        bool   `yaml:"enabled" envconfig:"TELEMETRY_ENABLED" default:"true"`
	ServiceName    string `yaml:"service_name" envconfig:"SERVICE_NAME" default:"mcp-ultra-wasm"`
	ServiceVersion string `yaml:"service_version" envconfig:"SERVICE_VERSION" default:"1.0.0"`
	Environment    string `yaml:"environment" envconfig:"ENVIRONMENT" default:"development"`
	Debug          bool   `yaml:"debug" envconfig:"TELEMETRY_DEBUG" default:"false"`

	// Tracing configuration
	Tracing TracingConfig `yaml:"tracing"`

	// Metrics configuration
	Metrics MetricsConfig `yaml:"metrics"`

	// Export configuration
	Exporters ExportersConfig `yaml:"exporters"`
}

// TracingConfig holds distributed tracing configuration
type TracingConfig struct {
	Enabled    bool          `yaml:"enabled" envconfig:"TRACING_ENABLED" default:"true"`
	SampleRate float64       `yaml:"sample_rate" envconfig:"TRACING_SAMPLE_RATE" default:"0.1"`
	MaxSpans   int           `yaml:"max_spans" envconfig:"TRACING_MAX_SPANS" default:"1000"`
	BatchSize  int           `yaml:"batch_size" envconfig:"TRACING_BATCH_SIZE" default:"512"`
	Timeout    time.Duration `yaml:"timeout" envconfig:"TRACING_TIMEOUT" default:"5s"`
}

// MetricsConfig holds metrics collection configuration
type MetricsConfig struct {
	Enabled          bool          `yaml:"enabled" envconfig:"METRICS_ENABLED" default:"true"`
	Port             int           `yaml:"port" envconfig:"METRICS_PORT" default:"9090"`
	Path             string        `yaml:"path" envconfig:"METRICS_PATH" default:"/metrics"`
	CollectInterval  time.Duration `yaml:"collect_interval" envconfig:"METRICS_INTERVAL" default:"15s"`
	HistogramBuckets []float64     `yaml:"histogram_buckets"`
}

// ExportersConfig holds exporter configurations
type ExportersConfig struct {
	// Jaeger exporter (deprecated but still supported)
	Jaeger JaegerConfig `yaml:"jaeger"`

	// OTLP exporter (recommended)
	OTLP OTLPConfig `yaml:"otlp"`

	// Console exporter (for debugging)
	Console ConsoleConfig `yaml:"console"`
}

// JaegerConfig holds Jaeger exporter configuration
type JaegerConfig struct {
	Enabled  bool   `yaml:"enabled" envconfig:"JAEGER_ENABLED" default:"false"`
	Endpoint string `yaml:"endpoint" envconfig:"JAEGER_ENDPOINT" default:"http://localhost:14268/api/traces"`
	User     string `yaml:"user" envconfig:"JAEGER_USER"`
	Password string `yaml:"password" envconfig:"JAEGER_PASSWORD"`
}

// OTLPConfig holds OTLP exporter configuration
type OTLPConfig struct {
	Enabled  bool              `yaml:"enabled" envconfig:"OTLP_ENABLED" default:"true"`
	Endpoint string            `yaml:"endpoint" envconfig:"OTLP_ENDPOINT" default:"http://localhost:4317"`
	Insecure bool              `yaml:"insecure" envconfig:"OTLP_INSECURE" default:"true"`
	Headers  map[string]string `yaml:"headers" envconfig:"OTLP_HEADERS"`
}

// ConsoleConfig holds console exporter configuration
type ConsoleConfig struct {
	Enabled bool `yaml:"enabled" envconfig:"CONSOLE_EXPORTER_ENABLED" default:"false"`
}

// FeaturesConfig holds feature flags configuration
type FeaturesConfig struct {
	RefreshInterval    time.Duration `yaml:"flags_refresh_interval" default:"30s"`
	ExperimentsEnabled bool          `yaml:"experiments_enabled" default:"true"`
}

// SecurityConfig holds all security-related configuration
type SecurityConfig struct {
	Auth  security.AuthConfig  `yaml:"auth"`
	OPA   security.OPAConfig   `yaml:"opa"`
	Vault security.VaultConfig `yaml:"vault"`
	TLS   security.TLSConfig   `yaml:"tls"`
}

// Load loads configuration from file and environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Load from config file if exists
	configFile := getEnv("CONFIG_FILE", "config/config.yaml")
	if _, err := os.Stat(configFile); err == nil {
		if err := loadFromFile(configFile, cfg); err != nil {
			return nil, fmt.Errorf("loading config file: %w", err)
		}
	}

	// Override with environment variables
	if err := envconfig.Process("", cfg); err != nil {
		return nil, fmt.Errorf("processing environment variables: %w", err)
	}

	return cfg, nil
}

// loadFromFile loads configuration from YAML file
func loadFromFile(filename string, cfg *Config) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log error but don't return - defer already happened
			// File was already read successfully, so this is non-critical
			log.Printf("Warning: failed to close config file %s: %v", filename, err)
		}
	}()

	decoder := yaml.NewDecoder(file)
	return decoder.Decode(cfg)
}

// getEnv returns environment variable value or default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// DSN returns PostgreSQL connection string
func (p PostgreSQLConfig) DSN() string {
	// Note: password comes from environment variable, not hardcoded
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Database, p.SSLMode)
}
