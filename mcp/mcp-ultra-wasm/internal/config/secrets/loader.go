package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"gopkg.in/yaml.v3"
)

// SecretsBackendType define o tipo de backend de secrets
type SecretsBackendType string

const (
	SecretsBackendEnv   SecretsBackendType = "env"
	SecretsBackendVault SecretsBackendType = "vault"
	SecretsBackendK8s   SecretsBackendType = "k8s"
)

// SecretsConfig representa a configuração de secrets
type SecretsConfig struct {
	Version        string               `yaml:"version"`
	SecretsBackend SecretsBackendConfig `yaml:"secrets_backend"`
	Database       DatabaseSecrets      `yaml:"database"`
	NATS           NATSSecrets          `yaml:"nats"`
	Telemetry      TelemetrySecrets     `yaml:"telemetry"`
	Auth           AuthSecrets          `yaml:"auth"`
	Encryption     EncryptionSecrets    `yaml:"encryption"`
	Required       []string             `yaml:"required_secrets"`
	Optional       map[string]string    `yaml:"optional_secrets"`
}

// SecretsBackendConfig configura o backend de secrets
type SecretsBackendConfig struct {
	Type  string       `yaml:"type"`
	Vault *VaultConfig `yaml:"vault,omitempty"`
}

// VaultConfig configuração do Vault
type VaultConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
	Path    string `yaml:"path"`
}

// DatabaseSecrets secrets do banco de dados
type DatabaseSecrets struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"` // NUNCA logar este campo
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
}

// NATSSecrets secrets do NATS
type NATSSecrets struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"` // NUNCA logar
	Token    string `yaml:"token"`    // NUNCA logar
}

// TelemetrySecrets secrets de telemetria
type TelemetrySecrets struct {
	OTLP       OTLPSecrets       `yaml:"otlp"`
	Prometheus PrometheusSecrets `yaml:"prometheus"`
}

// OTLPSecrets configuração OTLP
type OTLPSecrets struct {
	Endpoint string            `yaml:"endpoint"`
	Headers  map[string]string `yaml:"headers"`
}

// PrometheusSecrets configuração Prometheus
type PrometheusSecrets struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"` // NUNCA logar
}

// AuthSecrets secrets de autenticação
type AuthSecrets struct {
	JWTSecret string `yaml:"jwt_secret"` // NUNCA logar
	APIKeys   string `yaml:"api_keys"`   // NUNCA logar
}

// EncryptionSecrets secrets de criptografia
type EncryptionSecrets struct {
	MasterKey       string `yaml:"master_key"` // NUNCA logar
	KeyRotationDays int    `yaml:"key_rotation_days"`
}

// SecretsLoader carrega secrets de diferentes fontes
type SecretsLoader struct {
	config        *SecretsConfig
	backendType   SecretsBackendType
	vaultClient   *api.Client
	cacheEnabled  bool
	cacheDuration time.Duration
}

// NewSecretsLoader cria um novo loader de secrets
func NewSecretsLoader(configPath string) (*SecretsLoader, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo de configuração: %w", err)
	}

	// Expandir variáveis de ambiente no YAML
	expanded := os.ExpandEnv(string(data))

	var config SecretsConfig
	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return nil, fmt.Errorf("erro ao parsear configuração: %w", err)
	}

	loader := &SecretsLoader{
		config:        &config,
		backendType:   SecretsBackendType(config.SecretsBackend.Type),
		cacheEnabled:  true,
		cacheDuration: 5 * time.Minute,
	}

	// Inicializar backend específico
	switch loader.backendType {
	case SecretsBackendVault:
		if err := loader.initVaultClient(); err != nil {
			return nil, fmt.Errorf("erro ao inicializar Vault: %w", err)
		}
	case SecretsBackendK8s:
		// TODO: Implementar K8s secrets
	case SecretsBackendEnv:
		// Nada a inicializar
	default:
		return nil, fmt.Errorf("backend de secrets não suportado: %s", loader.backendType)
	}

	return loader, nil
}

// Load carrega todos os secrets
func (sl *SecretsLoader) Load(ctx context.Context) (*SecretsConfig, error) {
	// Validar secrets obrigatórios
	if err := sl.validateRequiredSecrets(); err != nil {
		return nil, fmt.Errorf("validação de secrets falhou: %w", err)
	}

	// Carregar do backend específico
	switch sl.backendType {
	case SecretsBackendVault:
		return sl.loadFromVault(ctx)
	case SecretsBackendK8s:
		return sl.loadFromK8s(ctx)
	default:
		// Env já foi expandido no NewSecretsLoader
		return sl.config, nil
	}
}

// initVaultClient inicializa o cliente Vault
func (sl *SecretsLoader) initVaultClient() error {
	if sl.config.SecretsBackend.Vault == nil {
		return fmt.Errorf("configuração do Vault não encontrada")
	}

	vaultConfig := api.DefaultConfig()
	vaultConfig.Address = sl.config.SecretsBackend.Vault.Address

	client, err := api.NewClient(vaultConfig)
	if err != nil {
		return fmt.Errorf("erro ao criar cliente Vault: %w", err)
	}

	client.SetToken(sl.config.SecretsBackend.Vault.Token)
	sl.vaultClient = client

	return nil
}

// loadFromVault carrega secrets do Vault
func (sl *SecretsLoader) loadFromVault(ctx context.Context) (*SecretsConfig, error) {
	if sl.vaultClient == nil {
		return nil, fmt.Errorf("cliente Vault não inicializado")
	}

	path := sl.config.SecretsBackend.Vault.Path
	secret, err := sl.vaultClient.Logical().ReadWithContext(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler secrets do Vault: %w", err)
	}

	if secret == nil {
		return nil, fmt.Errorf("secret não encontrado no path: %s", path)
	}

	// Mesclar secrets do Vault com configuração
	// TODO: Implementar merge de secrets

	return sl.config, nil
}

// loadFromK8s carrega secrets do Kubernetes
func (sl *SecretsLoader) loadFromK8s(_ context.Context) (*SecretsConfig, error) {
	// TODO: Implementar carregamento do K8s secrets
	return sl.config, fmt.Errorf("K8s secrets não implementado ainda")
}

// validateRequiredSecrets valida se todos os secrets obrigatórios estão presentes
func (sl *SecretsLoader) validateRequiredSecrets() error {
	var missing []string

	for _, secretName := range sl.config.Required {
		value := os.Getenv(secretName)
		if value == "" {
			missing = append(missing, secretName)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("secrets obrigatórios não definidos: %s", strings.Join(missing, ", "))
	}

	return nil
}

// GetDatabaseDSN retorna a DSN do banco de dados de forma segura
func (sl *SecretsLoader) GetDatabaseDSN() string {
	db := sl.config.Database
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.Username, db.Password, db.Database, db.SSLMode,
	)
}

// GetNATSConnection retorna a string de conexão NATS
func (sl *SecretsLoader) GetNATSConnection() string {
	nats := sl.config.NATS
	if nats.Token != "" {
		return fmt.Sprintf("%s?auth_token=%s", nats.URL, nats.Token)
	}
	if nats.Username != "" && nats.Password != "" {
		return fmt.Sprintf("nats://%s:%s@%s", nats.Username, nats.Password,
			strings.TrimPrefix(nats.URL, "nats://"))
	}
	return nats.URL
}

// Redact remove informações sensíveis para logs
func (sl *SecretsLoader) Redact(value string) string {
	if len(value) == 0 {
		return "<empty>"
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "..." + value[len(value)-4:]
}

// SecureString representa uma string segura que não aparece em logs
type SecureString struct {
	value string
}

// NewSecureString cria uma nova string segura
func NewSecureString(value string) *SecureString {
	return &SecureString{value: value}
}

// String implementa Stringer e redact o valor
func (s *SecureString) String() string {
	return "***REDACTED***"
}

// Value retorna o valor real (use com cuidado!)
func (s *SecureString) Value() string {
	return s.value
}

// MarshalJSON implementa json.Marshaler
func (s *SecureString) MarshalJSON() ([]byte, error) {
	return []byte(`"***REDACTED***"`), nil
}
