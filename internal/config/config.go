package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v10"
	"gopkg.in/yaml.v3"
)

const (
	defaultTemplatesPath    = "templates"
	defaultMetricsAddr      = ":2112"
	defaultMetricsNamespace = "mcp_ultra_templates"
	defaultOTLPEndpoint     = "localhost:4317"
	defaultLogLevel         = "info"
	defaultOperationTimeout = 30 * time.Second
	defaultRetryAttempts    = 3
)

// Config define os parâmetros de configuração globais carregados via arquivo YAML e variáveis de ambiente.
type Config struct {
	TemplatesPath string              `yaml:"templates_path" env:"TEMPLATES_PATH"`
	Logging       LoggingConfig       `yaml:"logging"`
	Observability ObservabilityConfig `yaml:"observability"`
	Rendering     RenderingConfig     `yaml:"rendering"`
}

// LoggingConfig encapsula definições de logging estruturado.
type LoggingConfig struct {
	Level  string `yaml:"level" env:"LOG_LEVEL"`
	Pretty bool   `yaml:"pretty" env:"LOG_PRETTY"`
}

// ObservabilityConfig abrange métricas Prometheus e traces OpenTelemetry.
type ObservabilityConfig struct {
	EnableMetrics bool   `yaml:"enable_metrics" env:"OBS_ENABLE_METRICS"`
	MetricsAddr   string `yaml:"metrics_address" env:"OBS_METRICS_ADDRESS"`
	MetricsNS     string `yaml:"metrics_namespace" env:"OBS_METRICS_NAMESPACE"`

	EnableTracing bool   `yaml:"enable_tracing" env:"OBS_ENABLE_TRACING"`
	OTLPEndpoint  string `yaml:"otlp_endpoint" env:"OBS_OTLP_ENDPOINT"`
	ServiceName   string `yaml:"service_name" env:"OBS_SERVICE_NAME"`
}

// RenderingConfig controla comportamento de geração dos templates.
type RenderingConfig struct {
	OperationTimeout time.Duration `yaml:"operation_timeout" env:"RENDER_TIMEOUT"`
	MaxRetryAttempts int           `yaml:"max_retry_attempts" env:"RENDER_MAX_RETRY_ATTEMPTS"`
}

// Load carrega a configuração padrão, opcionalmente mesclando com um arquivo YAML e variáveis de ambiente.
func Load(path string) (*Config, error) {
	cfg := &Config{
		TemplatesPath: defaultTemplatesPath,
		Logging: LoggingConfig{
			Level: defaultLogLevel,
		},
		Observability: ObservabilityConfig{
			EnableMetrics: true,
			MetricsAddr:   defaultMetricsAddr,
			MetricsNS:     defaultMetricsNamespace,
			EnableTracing: false,
			OTLPEndpoint:  defaultOTLPEndpoint,
			ServiceName:   "mcp-ultra-template-cli",
		},
		Rendering: RenderingConfig{
			OperationTimeout: defaultOperationTimeout,
			MaxRetryAttempts: defaultRetryAttempts,
		},
	}

	if path != "" {
		if err := mergeFile(cfg, path); err != nil {
			return nil, err
		}
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse env config: %w", err)
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func mergeFile(cfg *Config, path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("config file not found: %s", path)
		}
		return fmt.Errorf("read config file: %w", err)
	}

	if err := yaml.Unmarshal(content, cfg); err != nil {
		return fmt.Errorf("unmarshal config file: %w", err)
	}

	return nil
}

func validate(cfg *Config) error {
	if cfg.TemplatesPath == "" {
		return errors.New("templates_path must not be empty")
	}
	if cfg.Observability.MetricsAddr == "" {
		return errors.New("observability.metrics_address must not be empty")
	}
	if cfg.Observability.MetricsNS == "" {
		return errors.New("observability.metrics_namespace must not be empty")
	}
	if cfg.Rendering.OperationTimeout <= 0 {
		return errors.New("rendering.operation_timeout must be positive")
	}
	if cfg.Rendering.MaxRetryAttempts <= 0 {
		return errors.New("rendering.max_retry_attempts must be positive")
	}
	return nil
}
