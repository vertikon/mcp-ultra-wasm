package sdk

import (
	"fmt"
	"sync"
	"time"

	"github.com/vertikon/mcp-ultra-wasm-wasm/sdk/sdk-ultra-wasm/pkg/types"
	"go.uber.org/zap"
)

// RegistryManager gerencia o registro de plugins e suas dependências
type RegistryManager struct {
	mu           sync.RWMutex
	plugins      map[string]*PluginEntry
	capabilities map[string][]string // capability -> [plugin names]
	logger       *zap.Logger
	config       *RegistryConfig
}

type PluginEntry struct {
	Plugin       types.Plugin
	RegisteredAt time.Time
	LastUsed     *time.Time
	UsageCount   int64
	Status       string // "active", "inactive", "error"
	ErrorMessage string
	Metadata     map[string]interface{}
}

type RegistryConfig struct {
	MaxPlugins        int           `json:"max_plugins"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	IdleTimeout       time.Duration `json:"idle_timeout"`
	EnableMetrics     bool          `json:"enable_metrics"`
	HealthCheckPeriod time.Duration `json:"health_check_period"`
}

func NewRegistryManager(config *RegistryConfig, logger *zap.Logger) *RegistryManager {
	if config == nil {
		config = &RegistryConfig{
			MaxPlugins:        100,
			CleanupInterval:   5 * time.Minute,
			IdleTimeout:       30 * time.Minute,
			EnableMetrics:     true,
			HealthCheckPeriod: 1 * time.Minute,
		}
	}

	registry := &RegistryManager{
		plugins:      make(map[string]*PluginEntry),
		capabilities: make(map[string][]string),
		logger:       logger,
		config:       config,
	}

	// Iniciar background tasks
	go registry.startBackgroundTasks()

	return registry
}

// RegisterPlugin registra um novo plugin
func (r *RegistryManager) RegisterPlugin(plugin types.Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.GetName()

	// Verificar se já existe
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin %s já registrado", name)
	}

	// Verificar limite de plugins
	if len(r.plugins) >= r.config.MaxPlugins {
		return fmt.Errorf("número máximo de plugins atingido: %d", r.config.MaxPlugins)
	}

	// Criar entrada do plugin
	entry := &PluginEntry{
		Plugin:       plugin,
		RegisteredAt: time.Now(),
		UsageCount:   0,
		Status:       "active",
		Metadata:     make(map[string]interface{}),
	}

	r.plugins[name] = entry

	// Registrar capabilities
	for _, capability := range plugin.GetCapabilities() {
		r.capabilities[capability] = append(r.capabilities[capability], name)
	}

	r.logger.Info("Plugin registrado com sucesso",
		zap.String("name", name),
		zap.String("version", plugin.GetVersion()),
		zap.Strings("capabilities", plugin.GetCapabilities()))

	// Tentar configurar o plugin
	if err := plugin.Configure(map[string]interface{}{
		"registry":      "web-wasm",
		"registered_at": entry.RegisteredAt,
	}); err != nil {
		r.logger.Warn("Erro ao configurar plugin",
			zap.String("name", name),
			zap.Error(err))
		entry.Status = "error"
		entry.ErrorMessage = err.Error()
	}

	return nil
}

// UnregisterPlugin remove um plugin do registry
func (r *RegistryManager) UnregisterPlugin(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s não encontrado", name)
	}

	// Remover capabilities
	for _, capability := range entry.Plugin.GetCapabilities() {
		if plugins, ok := r.capabilities[capability]; ok {
			// Remover plugin da lista
			for i, pluginName := range plugins {
				if pluginName == name {
					r.capabilities[capability] = append(plugins[:i], plugins[i+1:]...)
					break
				}
			}

			// Remover capability se não tiver mais plugins
			if len(r.capabilities[capability]) == 0 {
				delete(r.capabilities, capability)
			}
		}
	}

	// Fazer shutdown do plugin
	if err := entry.Plugin.Shutdown(); err != nil {
		r.logger.Error("Erro durante shutdown do plugin",
			zap.String("name", name),
			zap.Error(err))
	}

	delete(r.plugins, name)

	r.logger.Info("Plugin removido com sucesso", zap.String("name", name))
	return nil
}

// GetPlugin obtém um plugin pelo nome
func (r *RegistryManager) GetPlugin(name string) (types.Plugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if entry, exists := r.plugins[name]; exists && entry.Status == "active" {
		// Atualizar uso
		r.updateUsage(name)
		return entry.Plugin, true
	}

	return nil, false
}

// ListPlugins lista todos os plugins registrados
func (r *RegistryManager) ListPlugins() map[string]*PluginEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]*PluginEntry)
	for name, entry := range r.plugins {
		// Copiar entrada para evitar race conditions
		entryCopy := *entry
		result[name] = &entryCopy
	}

	return result
}

// GetPluginsByCapability obtém plugins que suportam uma capability específica
func (r *RegistryManager) GetPluginsByCapability(capability string) []types.Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var plugins []types.Plugin
	if pluginNames, ok := r.capabilities[capability]; ok {
		for _, name := range pluginNames {
			if entry, exists := r.plugins[name]; exists && entry.Status == "active" {
				plugins = append(plugins, entry.Plugin)
			}
		}
	}

	return plugins
}

// GetCapabilities lista todas as capabilities disponíveis
func (r *RegistryManager) GetCapabilities() map[string][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]string)
	for capability, plugins := range r.capabilities {
		result[capability] = append([]string{}, plugins...)
	}

	return result
}

// ExecutePlugin executa um método em um plugin específico
func (r *RegistryManager) ExecutePlugin(pluginName, method string, params map[string]interface{}, context map[string]string) (*types.ExecuteResponse, error) {
	r.mu.RLock()
	entry, exists := r.plugins[pluginName]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin %s não encontrado", pluginName)
	}

	if entry.Status != "active" {
		return nil, fmt.Errorf("plugin %s não está ativo (status: %s)", pluginName, entry.Status)
	}

	// Criar requisição
	req := &types.ExecuteRequest{
		Plugin:     pluginName,
		Method:     method,
		Parameters: params,
		Context:    context,
	}

	// Adicionar metadata do registry
	if req.Context == nil {
		req.Context = make(map[string]string)
	}
	req.Context["registry"] = "web-wasm"
	req.Context["request_id"] = fmt.Sprintf("req-%d", time.Now().UnixNano())

	// Executar plugin
	resp, err := entry.Plugin.Execute(req.Context, req)

	// Atualizar métricas
	r.mu.Lock()
	if err != nil {
		entry.Status = "error"
		entry.ErrorMessage = err.Error()
	} else {
		if entry.Status == "error" {
			entry.Status = "active"
			entry.ErrorMessage = ""
		}
		entry.UsageCount++
		now := time.Now()
		entry.LastUsed = &now
	}
	r.mu.Unlock()

	return resp, err
}

// HealthCheck realiza verificação de saúde de todos os plugins
func (r *RegistryManager) HealthCheck() map[string]types.HealthStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	results := make(map[string]types.HealthStatus)
	for name, entry := range r.plugins {
		if entry.Status == "active" {
			results[name] = entry.Plugin.GetHealth()
		} else {
			results[name] = types.HealthStatusUnhealthy
		}
	}

	return results
}

// GetMetrics retorna métricas do registry
func (r *RegistryManager) GetMetrics() *RegistryMetrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := &RegistryMetrics{
		TotalPlugins:      len(r.plugins),
		ActivePlugins:     0,
		InactivePlugins:   0,
		ErrorPlugins:      0,
		TotalCapabilities: len(r.capabilities),
		RegisteredAt:      time.Now(),
	}

	for _, entry := range r.plugins {
		switch entry.Status {
		case "active":
			metrics.ActivePlugins++
		case "inactive":
			metrics.InactivePlugins++
		case "error":
			metrics.ErrorPlugins++
		}
	}

	// Calcular métricas de uso
	for _, entry := range r.plugins {
		metrics.TotalUsage += entry.UsageCount
	}

	return metrics
}

// Métodos privados

func (r *RegistryManager) updateUsage(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if entry, exists := r.plugins[name]; exists {
		entry.UsageCount++
		now := time.Now()
		entry.LastUsed = &now
	}
}

func (r *RegistryManager) startBackgroundTasks() {
	// Health check periódico
	if r.config.HealthCheckPeriod > 0 {
		ticker := time.NewTicker(r.config.HealthCheckPeriod)
		go func() {
			for range ticker.C {
				r.performHealthCheck()
			}
		}()
	}

	// Cleanup periódico
	if r.config.CleanupInterval > 0 {
		ticker := time.NewTicker(r.config.CleanupInterval)
		go func() {
			for range ticker.C {
				r.performCleanup()
			}
		}()
	}
}

func (r *RegistryManager) performHealthCheck() {
	healthResults := r.HealthCheck()

	r.mu.Lock()
	defer r.mu.Unlock()

	for name, health := range healthResults {
		if entry, exists := r.plugins[name]; exists {
			if health == types.HealthStatusUnhealthy && entry.Status == "active" {
				r.logger.Warn("Plugin marcado como unhealthy", zap.String("name", name))
				entry.Status = "error"
				entry.ErrorMessage = "Health check failed"
			} else if health == types.HealthStatusHealthy && entry.Status == "error" {
				r.logger.Info("Plugin recuperado", zap.String("name", name))
				entry.Status = "active"
				entry.ErrorMessage = ""
			}
		}
	}
}

func (r *RegistryManager) performCleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	var toRemove []string

	for name, entry := range r.plugins {
		// Remover plugins inativos por muito tempo
		if entry.Status == "inactive" && entry.LastUsed != nil {
			if now.Sub(*entry.LastUsed) > r.config.IdleTimeout {
				toRemove = append(toRemove, name)
			}
		}
	}

	// Remover plugins marcados
	for _, name := range toRemove {
		r.unregisterUnsafe(name)
		r.logger.Info("Plugin removido por inatividade", zap.String("name", name))
	}
}

func (r *RegistryManager) unregisterUnsafe(name string) {
	entry := r.plugins[name]

	// Remover capabilities
	for _, capability := range entry.Plugin.GetCapabilities() {
		if plugins, ok := r.capabilities[capability]; ok {
			for i, pluginName := range plugins {
				if pluginName == name {
					r.capabilities[capability] = append(plugins[:i], plugins[i+1:]...)
					break
				}
			}

			if len(r.capabilities[capability]) == 0 {
				delete(r.capabilities, capability)
			}
		}
	}

	// Fazer shutdown
	entry.Plugin.Shutdown()
	delete(r.plugins, name)
}

// Metrics

type RegistryMetrics struct {
	TotalPlugins      int       `json:"total_plugins"`
	ActivePlugins     int       `json:"active_plugins"`
	InactivePlugins   int       `json:"inactive_plugins"`
	ErrorPlugins      int       `json:"error_plugins"`
	TotalCapabilities int       `json:"total_capabilities"`
	TotalUsage        int64     `json:"total_usage"`
	RegisteredAt      time.Time `json:"registered_at"`
}
