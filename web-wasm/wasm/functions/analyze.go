package functions

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm/web-wasm/wasm/internal"
)

// AnalyzeProject analisa um projeto Go e retorna métricas e insights
func AnalyzeProject(this js.Value, args []js.Value) interface{} {
	// Callback function para retornar resultado async
	callback := args[len(args)-1]
	if callback.Type() != js.TypeFunction {
		return map[string]interface{}{
			"error": "Último argumento deve ser uma função de callback",
		}
	}

	// Extrair configuração do primeiro argumento
	var config map[string]interface{}
	if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
		configStr := args[0].String()
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			callback.Invoke(map[string]interface{}{
				"error": fmt.Sprintf("Erro ao parsear configuração: %v", err),
			})
			return nil
		}
	}

	// Simular processamento assíncrono
	go func() {
		result := internal.PerformProjectAnalysis(config)
		callback.Invoke(result)
	}()

	return nil
}

// GenerateCode gera código baseado em especificações
func GenerateCode(this js.Value, args []js.Value) interface{} {
	callback := args[len(args)-1]
	if callback.Type() != js.TypeFunction {
		return map[string]interface{}{
			"error": "Último argumento deve ser uma função de callback",
		}
	}

	var spec map[string]interface{}
	if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
		specStr := args[0].String()
		if err := json.Unmarshal([]byte(specStr), &spec); err != nil {
			callback.Invoke(map[string]interface{}{
				"error": fmt.Sprintf("Erro ao parsear especificação: %v", err),
			})
			return nil
		}
	}

	go func() {
		result := internal.GenerateCodeFromSpec(spec)
		callback.Invoke(result)
	}()

	return nil
}

// ValidateConfig valida configurações de projeto
func ValidateConfig(this js.Value, args []js.Value) interface{} {
	callback := args[len(args)-1]
	if callback.Type() != js.TypeFunction {
		return map[string]interface{}{
			"error": "Último argumento deve ser uma função de callback",
		}
	}

	var config map[string]interface{}
	if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
		configStr := args[0].String()
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			callback.Invoke(map[string]interface{}{
				"error": fmt.Sprintf("Erro ao parsear configuração: %v", err),
			})
			return nil
		}
	}

	go func() {
		result := internal.ValidateConfiguration(config)
		callback.Invoke(result)
	}()

	return nil
}

// ProcessTask processa tasks genéricas
func ProcessTask(this js.Value, args []js.Value) interface{} {
	callback := args[len(args)-1]
	if callback.Type() != js.TypeFunction {
		return map[string]interface{}{
			"error": "Último argumento deve ser uma função de callback",
		}
	}

	var task map[string]interface{}
	if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
		taskStr := args[0].String()
		if err := json.Unmarshal([]byte(taskStr), &task); err != nil {
			callback.Invoke(map[string]interface{}{
				"error": fmt.Sprintf("Erro ao parsear task: %v", err),
			})
			return nil
		}
	}

	go func() {
		result := internal.ProcessGenericTask(task)
		callback.Invoke(result)
	}()

	return nil
}

// GetVersion retorna a versão do módulo WASM
func GetVersion(this js.Value, args []js.Value) interface{} {
	return map[string]interface{}{
		"version":     "1.0.0",
		"build_time":  "2025-11-09T15:00:00Z",
		"go_version":  "go1.24.0",
		"module_name": "mcp-ultra-wasm-web",
		"features": []string{
			"project_analysis",
			"code_generation",
			"config_validation",
			"task_processing",
		},
	}
}

// Initialize inicializa o módulo WASM com configurações
func Initialize(this js.Value, args []js.Value) interface{} {
	config := make(map[string]interface{})
	
	if len(args) > 0 && !args[0].IsUndefined() && !args[0].IsNull() {
		configStr := args[0].String()
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			return map[string]interface{}{
				"error": fmt.Sprintf("Erro ao parsear configuração: %v", err),
			}
		}
	}

	// Inicializar módulo interno
	if err := internal.InitializeWasmModule(config); err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Erro ao inicializar módulo: %v", err),
		}
	}

	return map[string]interface{}{
		"status":  "initialized",
		"message": "Módulo WASM inicializado com sucesso",
	}
}

// Cleanup limpa recursos do módulo WASM
func Cleanup(this js.Value, args []js.Value) interface{} {
	if err := internal.CleanupWasmModule(); err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Erro durante cleanup: %v", err),
		}
	}

	return map[string]interface{}{
		"status":  "cleaned_up",
		"message": "Recursos do módulo WASM liberados com sucesso",
	}
}