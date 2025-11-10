package sdk

import (
	"context"
	"fmt"
	"time"

	"github.com/vertikon/mcp-ultra-wasm-wasm/sdk/sdk-ultra-wasm/pkg/client"
	"github.com/vertikon/mcp-ultra-wasm-wasm/sdk/sdk-ultra-wasm/pkg/types"
	"go.uber.org/zap"
)

type SDKClient struct {
	client  *client.Client
	logger  *zap.Logger
	plugins map[string]types.Plugin
	config  *Config
}

type Config struct {
	SDKAddress       string        `json:"sdk_address"`
	Timeout          time.Duration `json:"timeout"`
	Retries          int           `json:"retries"`
	EnableCache      bool          `json:"enable_cache"`
	CacheSize        int           `json:"cache_size"`
	DefaultNamespace string        `json:"default_namespace"`
}

func NewSDKClient(config *Config, logger *zap.Logger) (*SDKClient, error) {
	if config == nil {
		config = &Config{
			SDKAddress:       "localhost:9090",
			Timeout:          30 * time.Second,
			Retries:          3,
			EnableCache:      true,
			CacheSize:        1000,
			DefaultNamespace: "wasm",
		}
	}

	// Criar cliente SDK
	clientConfig := &client.Config{
		Address:          config.SDKAddress,
		Timeout:          config.Timeout,
		Retries:          config.Retries,
		EnableCache:      config.EnableCache,
		CacheSize:        config.CacheSize,
		DefaultNamespace: config.DefaultNamespace,
	}

	sdkClient, err := client.New(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar cliente SDK: %w", err)
	}

	sdkWrapper := &SDKClient{
		client:  sdkClient,
		logger:  logger,
		plugins: make(map[string]types.Plugin),
		config:  config,
	}

	// Inicializar plugins padrão
	if err := sdkWrapper.initializePlugins(); err != nil {
		logger.Warn("Erro ao inicializar plugins", zap.Error(err))
	}

	logger.Info("SDK client inicializado com sucesso",
		zap.String("address", config.SDKAddress),
		zap.String("namespace", config.DefaultNamespace))

	return sdkWrapper, nil
}

func (s *SDKClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *SDKClient) initializePlugins() error {
	// Registrar o wasm como plugin do SDK
	webWasmPlugin := &WebWasmPlugin{
		name:    "wasm",
		version: "1.0.0",
		client:  s,
	}

	if err := s.client.RegisterPlugin(webWasmPlugin); err != nil {
		return fmt.Errorf("erro ao registrar plugin wasm: %w", err)
	}

	s.plugins["wasm"] = webWasmPlugin

	// Carregar plugins disponíveis
	availablePlugins, err := s.client.ListPlugins()
	if err != nil {
		s.logger.Error("Erro ao listar plugins", zap.Error(err))
		return err
	}

	for _, plugin := range availablePlugins {
		s.plugins[plugin.GetName()] = plugin
		s.logger.Info("Plugin carregado",
			zap.String("name", plugin.GetName()),
			zap.String("version", plugin.GetVersion()))
	}

	return nil
}

// ExecuteProjectAnalysis executa análise de projeto usando o SDK
func (s *SDKClient) ExecuteProjectAnalysis(ctx context.Context, req *AnalysisRequest) (*AnalysisResult, error) {
	s.logger.Info("Executando análise de projeto via SDK",
		zap.String("project_path", req.ProjectPath),
		zap.String("analysis_type", req.Type))

	// Criar requisição para o SDK
	sdkReq := &types.ExecuteRequest{
		Plugin: "analyzer",
		Method: "analyze",
		Parameters: map[string]interface{}{
			"project_path":  req.ProjectPath,
			"analysis_type": req.Type,
			"options":       req.Options,
			"timeout":       req.Timeout.Seconds(),
		},
		Context: map[string]string{
			"source":    "wasm",
			"task_id":   req.TaskID,
			"namespace": s.config.DefaultNamespace,
		},
	}

	// Executar via SDK
	resp, err := s.client.Execute(ctx, sdkReq)
	if err != nil {
		s.logger.Error("Erro na análise via SDK", zap.Error(err))
		return nil, fmt.Errorf("erro na análise: %w", err)
	}

	// Converter resultado
	result := &AnalysisResult{
		TaskID:      req.TaskID,
		ProjectPath: req.ProjectPath,
		Type:        req.Type,
		Status:      "completed",
		Data:        resp.Result,
		Metadata: map[string]interface{}{
			"execution_time": resp.ExecutionTime,
			"plugin_used":    resp.Plugin,
			"method_used":    resp.Method,
			"timestamp":      time.Now().UTC(),
		},
	}

	s.logger.Info("Análise concluída via SDK",
		zap.String("task_id", req.TaskID),
		zap.Duration("execution_time", resp.ExecutionTime))

	return result, nil
}

// GenerateCode gera código usando o SDK
func (s *SDKClient) GenerateCode(ctx context.Context, req *GenerationRequest) (*GenerationResult, error) {
	s.logger.Info("Gerando código via SDK",
		zap.String("component_type", req.ComponentType),
		zap.String("language", req.Language))

	// Criar requisição para o SDK
	sdkReq := &types.ExecuteRequest{
		Plugin: "generator",
		Method: "generate",
		Parameters: map[string]interface{}{
			"component_type": req.ComponentType,
			"name":           req.Name,
			"language":       req.Language,
			"specification":  req.Specification,
			"templates":      req.Templates,
		},
		Context: map[string]string{
			"source":    "wasm",
			"task_id":   req.TaskID,
			"namespace": s.config.DefaultNamespace,
		},
	}

	// Executar via SDK
	resp, err := s.client.Execute(ctx, sdkReq)
	if err != nil {
		s.logger.Error("Erro na geração via SDK", zap.Error(err))
		return nil, fmt.Errorf("erro na geração: %w", err)
	}

	// Converter resultado
	result := &GenerationResult{
		TaskID:         req.TaskID,
		ComponentType:  req.ComponentType,
		Name:           req.Name,
		Language:       req.Language,
		Status:         "completed",
		GeneratedCode:  resp.Result,
		FilesGenerated: s.extractFilesGenerated(resp.Result),
		Metadata: map[string]interface{}{
			"execution_time": resp.ExecutionTime,
			"plugin_used":    resp.Plugin,
			"method_used":    resp.Method,
			"timestamp":      time.Now().UTC(),
		},
	}

	s.logger.Info("Geração concluída via SDK",
		zap.String("task_id", req.TaskID),
		zap.Int("files_generated", len(result.FilesGenerated)))

	return result, nil
}

// ValidateConfiguration valida configuração usando o SDK
func (s *SDKClient) ValidateConfiguration(ctx context.Context, req *ValidationRequest) (*ValidationResult, error) {
	s.logger.Info("Validando configuração via SDK",
		zap.String("config_type", req.Type))

	// Criar requisição para o SDK
	sdkReq := &types.ExecuteRequest{
		Plugin: "validator",
		Method: "validate",
		Parameters: map[string]interface{}{
			"type":        req.Type,
			"config":      req.Config,
			"rules":       req.Rules,
			"strict_mode": req.StrictMode,
		},
		Context: map[string]string{
			"source":    "wasm",
			"task_id":   req.TaskID,
			"namespace": s.config.DefaultNamespace,
		},
	}

	// Executar via SDK
	resp, err := s.client.Execute(ctx, sdkReq)
	if err != nil {
		s.logger.Error("Erro na validação via SDK", zap.Error(err))
		return nil, fmt.Errorf("erro na validação: %w", err)
	}

	// Converter resultado
	result := s.convertValidationResult(req.TaskID, req.Type, resp.Result)

	s.logger.Info("Validação concluída via SDK",
		zap.String("task_id", req.TaskID),
		zap.Bool("valid", result.Valid))

	return result, nil
}

// ListPlugins lista todos os plugins disponíveis
func (s *SDKClient) ListPlugins() map[string]types.Plugin {
	return s.plugins
}

// GetPlugin obtém um plugin específico
func (s *SDKClient) GetPlugin(name string) (types.Plugin, bool) {
	plugin, exists := s.plugins[name]
	return plugin, exists
}

// GetMetrics obtém métricas do SDK
func (s *SDKClient) GetMetrics(ctx context.Context) (*types.Metrics, error) {
	return s.client.GetMetrics(ctx)
}

// Helper functions
func (s *SDKClient) extractFilesGenerated(result map[string]interface{}) []string {
	var files []string

	if filesData, ok := result["files"].([]interface{}); ok {
		for _, file := range filesData {
			if fileName, ok := file.(string); ok {
				files = append(files, fileName)
			}
		}
	}

	return files
}

func (s *SDKClient) convertValidationResult(taskID, configType string, result map[string]interface{}) *ValidationResult {
	validationResult := &ValidationResult{
		TaskID: taskID,
		Type:   configType,
		Status: "completed",
		Metadata: map[string]interface{}{
			"timestamp": time.Now().UTC(),
		},
	}

	// Extrair campos do resultado
	if valid, ok := result["valid"].(bool); ok {
		validationResult.Valid = valid
	}

	if errors, ok := result["errors"].([]interface{}); ok {
		for _, err := range errors {
			if errMsg, ok := err.(string); ok {
				validationResult.Errors = append(validationResult.Errors, errMsg)
			}
		}
	}

	if warnings, ok := result["warnings"].([]interface{}); ok {
		for _, warn := range warnings {
			if warnMsg, ok := warn.(string); ok {
				validationResult.Warnings = append(validationResult.Warnings, warnMsg)
			}
		}
	}

	if score, ok := result["score"].(float64); ok {
		validationResult.Score = score
	}

	return validationResult
}
