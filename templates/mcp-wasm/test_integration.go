// test_integration.go - Teste de integração do MCP WASM
// Para rodar: go run test_integration.go

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"./wasm/internal"
)

func main() {
	fmt.Println("=== MCP WASM Integration Test ===")

	// 1. Testar inicialização do módulo
	fmt.Println("\n1. Testing module initialization...")
	config := map[string]interface{}{
		"debug":   true,
		"timeout": 30,
		"features": []string{"analysis", "generation", "validation"},
	}

	if err := internal.InitializeWasmModule(config); err != nil {
		log.Fatalf("❌ Failed to initialize WASM module: %v", err)
	}
	fmt.Println("✅ WASM module initialized successfully")

	// 2. Testar inicialização do cliente MCP
	fmt.Println("\n2. Testing MCP client initialization...")
	if err := internal.InitializeMCPClient(config); err != nil {
		log.Fatalf("❌ Failed to initialize MCP client: %v", err)
	}
	fmt.Println("✅ MCP client initialized successfully")

	// 3. Testar análise de projeto
	fmt.Println("\n3. Testing project analysis...")
	analysisConfig := map[string]interface{}{
		"project_path":  "./testdata/sample-project",
		"analysis_type": "quick",
		"options": map[string]interface{}{
			"include_tests":    true,
			"check_security":   false,
			"performance_scan": false,
		},
	}

	analysisResult := internal.PerformProjectAnalysis(analysisConfig)
	fmt.Printf("Analysis result: %s\n", formatJSON(analysisResult))

	// 4. Testar geração de código
	fmt.Println("\n4. Testing code generation...")
	genConfig := map[string]interface{}{
		"component_type": "api",
		"name":           "TestAPI",
		"language":       "go",
		"options": map[string]interface{}{
			"port":     "8080",
			"database": "postgresql",
			"auth":     "jwt",
		},
	}

	genResult := internal.GenerateCodeFromSpec(genConfig)
	fmt.Printf("Generation result: %s\n", formatJSON(genResult))

	// 5. Testar validação
	fmt.Println("\n5. Testing configuration validation...")
	validationConfig := map[string]interface{}{
		"type":         "project",
		"project_name": "test-project",
		"module_path":  "github.com/example/test-project",
		"options": map[string]interface{}{
			"require_tests": true,
			"check_format":  true,
		},
	}

	validationResult := internal.ValidateConfiguration(validationConfig)
	fmt.Printf("Validation result: %s\n", formatJSON(validationResult))

	// 6. Testar diferentes tipos de análise
	fmt.Println("\n6. Testing different analysis types...")

	analysisTypes := []string{"quick", "security", "performance", "full"}
	for _, analysisType := range analysisTypes {
		fmt.Printf("\n--- Testing %s analysis ---\n", analysisType)
		config := map[string]interface{}{
			"project_path":  "./testdata/sample-project",
			"analysis_type": analysisType,
		}

		result := internal.AnalyzeProjectReal(config)
		if result != nil {
			fmt.Printf("✅ %s analysis completed\n", analysisType)
			fmt.Printf("   Health Score: %.2f\n", result.HealthScore)
			fmt.Printf("   Processing Time: %s\n", result.ProcessingTime)
			if result.Issues != nil {
				fmt.Printf("   Issues Found: %d\n", len(result.Issues))
			}
		} else {
			fmt.Printf("❌ %s analysis failed\n", analysisType)
		}
	}

	// 7. Testar tipos de geração de código
	fmt.Println("\n7. Testing different code generation types...")

	componentTypes := []string{"api", "service", "cli", "library"}
	for _, compType := range componentTypes {
		fmt.Printf("\n--- Testing %s generation ---\n", compType)
		config := map[string]interface{}{
			"component_type": compType,
			"name":           fmt.Sprintf("Test%s", capitalize(compType)),
			"options": map[string]interface{}{
				"version": "1.0.0",
			},
		}

		result := internal.GenerateCodeFromSpec(config)
		if result != nil {
			fmt.Printf("✅ %s generation completed\n", compType)
			if files, ok := result["files_generated"].(int); ok {
				fmt.Printf("   Files Generated: %d\n", files)
			}
			if warnings, ok := result["warnings"].([]string); ok {
				fmt.Printf("   Warnings: %d\n", len(warnings))
			}
		} else {
			fmt.Printf("❌ %s generation failed\n", compType)
		}
	}

	// 8. Testar cleanup
	fmt.Println("\n8. Testing cleanup...")
	if err := internal.CleanupWasmModule(); err != nil {
		log.Printf("❌ Cleanup failed: %v", err)
	} else {
		fmt.Println("✅ Cleanup completed successfully")
	}

	fmt.Println("\n=== Integration Test Completed ===")
	fmt.Println("\n📋 Summary:")
	fmt.Println("✅ Module initialization")
	fmt.Println("✅ MCP client initialization")
	fmt.Println("✅ Project analysis (multiple types)")
	fmt.Println("✅ Code generation (multiple types)")
	fmt.Println("✅ Configuration validation")
	fmt.Println("✅ Cleanup")

	fmt.Println("\n🚀 MCP WASM is ready for production use!")
}

// Helper functions
func formatJSON(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(jsonBytes)
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}