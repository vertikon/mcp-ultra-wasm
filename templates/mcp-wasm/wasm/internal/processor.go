package internal

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ProcessorState mantém o estado do processador
type ProcessorState struct {
	Initialized bool
	Config      map[string]interface{}
	StartTime   time.Time
}

var state ProcessorState

func init() {
	rand.Seed(time.Now().UnixNano())
}

// InitializeWasmModule inicializa o módulo WASM
func InitializeWasmModule(config map[string]interface{}) error {
	state = ProcessorState{
		Initialized: true,
		Config:      config,
		StartTime:   time.Now(),
	}
	return nil
}

// CleanupWasmModule limpa recursos do módulo
func CleanupWasmModule() error {
	state = ProcessorState{}
	return nil
}

// PerformProjectAnalysis realiza análise de projeto real usando MCP
func PerformProjectAnalysis(config map[string]interface{}) map[string]interface{} {
	if !state.Initialized {
		return map[string]interface{}{
			"error": "Módulo não inicializado",
		}
	}

	// Tentar usar MCP primeiro
	if mcpClient.Initialized {
		response := AnalyzeWithMCP(config)
		if response.Success {
			// Converter resposta para o formato esperado
			result := map[string]interface{}{
				"status":        "completed",
				"project_path":  getString(config, "project_path"),
				"analysis_type": getString(config, "analysis_type"),
				"mcp_analysis":  response,
				"processing_time": response.ProcessingTime.String(),
				"timestamp":     response.Timestamp,
			}

			// Adicionar dados do MCP se disponíveis
			if response.Data != nil {
				result["data"] = response.Data
			}
			if response.Score != nil {
				result["score"] = response.Score
			}
			if response.Metrics != nil {
				result["metrics"] = response.Metrics
			}

			return result
		}

		// Se MCP falhar, continuar com análise local
		fmt.Printf("MCP analysis failed: %s, falling back to local analysis\n", response.Error)
	}

	// Análise local usando as novas funções reais
	analysis := AnalyzeProjectReal(config)

	// Converter para formato de mapa
	result, err := analysis.ToJSON()
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Erro ao converter análise para JSON: %v", err),
		}
	}

	// Parse do JSON para mapa (para manter compatibilidade)
	var resultMap map[string]interface{}
	if err := json.Unmarshal([]byte(result), &resultMap); err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Erro ao processar resultado: %v", err),
		}
	}

	// Adicionar campos de compatibilidade
	resultMap["status"] = "completed"
	resultMap["analysis"] = resultMap

	return resultMap
}

// GenerateCodeFromSpec gera código baseado em especificação usando MCP
func GenerateCodeFromSpec(spec map[string]interface{}) map[string]interface{} {
	if !state.Initialized {
		return map[string]interface{}{
			"error": "Módulo não inicializado",
		}
	}

	// Tentar usar MCP primeiro
	if mcpClient.Initialized {
		response := GenerateCodeWithMCP(spec)
		if response.Success {
			// Converter resposta para o formato esperado
			result := map[string]interface{}{
				"status":           "completed",
				"component_type":   getString(spec, "component_type"),
				"name":             getString(spec, "name"),
				"language":         getString(spec, "language"),
				"mcp_generation":   response,
				"files_generated":  len(response.Files),
				"warnings":         response.Warnings,
				"instructions":     response.Instructions,
				"timestamp":        time.Now().UTC(),
			}

			// Se houver arquivos gerados
			if len(response.Files) > 0 {
				// Gerar código combinado dos arquivos
				var combinedCode strings.Builder
				for _, file := range response.Files {
					combinedCode.WriteString(fmt.Sprintf("// File: %s\n", file.Path))
					combinedCode.WriteString(file.Content)
					combinedCode.WriteString("\n\n")
				}
				result["generated_code"] = combinedCode.String()
			} else if response.GeneratedCode != "" {
				result["generated_code"] = response.GeneratedCode
			}

			return result
		}

		// Se MCP falhar, continuar com geração local
		fmt.Printf("MCP code generation failed: %s, falling back to local generation\n", response.Error)
	}

	// Geração local (mantendo compatibilidade)
	componentType := getString(spec, "component_type")
	name := getString(spec, "name")
	language := getString(spec, "language")
	if language == "" {
		language = "go"
	}

	// Gerar código simulado melhorado
	generatedCode := generateLocalCode(componentType, name, spec)

	return map[string]interface{}{
		"status":           "completed",
		"component_type":   componentType,
		"name":             name,
		"language":         language,
		"generated_code":   generatedCode,
		"files_generated":  1,
		"fallback":         "local_generation",
		"timestamp":        time.Now().UTC(),
	}
}

// ValidateConfiguration valida configurações usando MCP
func ValidateConfiguration(config map[string]interface{}) map[string]interface{} {
	if !state.Initialized {
		return map[string]interface{}{
			"error": "Módulo não inicializado",
		}
	}

	// Tentar usar MCP primeiro
	if mcpClient.Initialized {
		response := ValidateWithMCP(config)
		if response.Success {
			// Converter resposta para o formato esperado
			result := map[string]interface{}{
				"status":       "completed",
				"mcp_validation": response,
				"timestamp":    time.Now().UTC(),
			}

			// Extrair dados de validação
			if response.Data != nil {
				if valid, ok := response.Data["valid"].(bool); ok {
					result["valid"] = valid
				}
				if issuesCount, ok := response.Data["issues_count"].(int); ok {
					result["errors_count"] = issuesCount
				}
				result["validation_data"] = response.Data
			}

			// Adicionar score se disponível
			if response.Score != nil {
				result["validation_score"] = response.Score
			}

			// Adicionar warnings
			if len(response.Warnings) > 0 {
				result["warnings"] = response.Warnings
			}

			return result
		}

		// Se MCP falhar, continuar com validação local
		fmt.Printf("MCP validation failed: %s, falling back to local validation\n", response.Error)
	}

	// Validação local (mantendo compatibilidade)
	configType := getString(config, "type")
	validationErrors := []string{}
	warnings := []string{}

	// Simular validação
	switch configType {
	case "project":
		if getString(config, "project_name") == "" {
			validationErrors = append(validationErrors, "project_name is required")
		}
		if getString(config, "module_path") == "" {
			validationErrors = append(validationErrors, "module_path is required")
		}
	case "deployment":
		if getString(config, "target") == "" {
			validationErrors = append(validationErrors, "deployment target is required")
		}
		warnings = append(warnings, "Consider using HTTPS in production")
	}

	isValid := len(validationErrors) == 0

	return map[string]interface{}{
		"status":           "completed",
		"valid":            isValid,
		"errors":           validationErrors,
		"warnings":         warnings,
		"validated_fields": len(config),
		"fallback":         "local_validation",
		"timestamp":        time.Now().UTC(),
	}
}

// ProcessGenericTask processa tasks genéricas
func ProcessGenericTask(task map[string]interface{}) map[string]interface{} {
	if !state.Initialized {
		return map[string]interface{}{
			"error": "Módulo não inicializado",
		}
	}

	// Simular tempo de processamento baseado no tipo de task
	taskType := getString(task, "type")
	var processingTime time.Duration
	
	switch taskType {
	case "heavy":
		processingTime = time.Duration(1000+rand.Intn(2000)) * time.Millisecond
	case "medium":
		processingTime = time.Duration(500+rand.Intn(1000)) * time.Millisecond
	default:
		processingTime = time.Duration(200+rand.Intn(500)) * time.Millisecond
	}

	time.Sleep(processingTime)

	// Simular resultados diferentes baseados no tipo
	var result map[string]interface{}
	switch taskType {
	case "analysis":
		result = map[string]interface{}{
			"issues_found":    rand.Intn(10),
			"suggestions":     rand.Intn(5) + 1,
			"score":          float64(70+rand.Intn(30)),
		}
	case "generation":
		result = map[string]interface{}{
			"files_created":   rand.Intn(10) + 1,
			"lines_written":   rand.Intn(500) + 100,
			"complexity":     []string{"low", "medium", "high"}[rand.Intn(3)],
		}
	default:
		result = map[string]interface{}{
			"processed_items": rand.Intn(20) + 1,
			"success_rate":    float64(80+rand.Intn(20)),
		}
	}

	return map[string]interface{}{
		"status":          "completed",
		"task_type":       taskType,
		"result":          result,
		"processing_time": processingTime.String(),
		"timestamp":       time.Now().UTC(),
	}
}

// Funções helper para gerar análises simuladas
func generateFullAnalysis(projectPath string) map[string]interface{} {
	return map[string]interface{}{
		"metrics": map[string]interface{}{
			"lines_of_code":    rand.Intn(10000) + 1000,
			"files_count":      rand.Intn(100) + 10,
			"packages_count":   rand.Intn(20) + 1,
			"test_coverage":    float64(60+rand.Intn(40)),
		},
		"quality": map[string]interface{}{
			"code_smells":      rand.Intn(20),
			"duplicated_code":  rand.Intn(5),
			"maintainability": "A",
			"reliability":      "B",
		},
		"security": map[string]interface{}{
			"vulnerabilities": rand.Intn(5),
			"hotspots":       rand.Intn(3),
			"security_rating": "Good",
		},
	}
}

func generateQuickAnalysis(projectPath string) map[string]interface{} {
	return map[string]interface{}{
		"basic_metrics": map[string]interface{}{
			"files_count":    rand.Intn(50) + 5,
			"lines_of_code":  rand.Intn(5000) + 500,
			"go_modules":     rand.Intn(10) + 1,
		},
		"health_score": float64(70+rand.Intn(30)),
	}
}

func generateSecurityAnalysis(projectPath string) map[string]interface{} {
	return map[string]interface{}{
		"vulnerabilities": map[string]interface{}{
			"critical": rand.Intn(2),
			"high":     rand.Intn(3),
			"medium":   rand.Intn(8),
			"low":      rand.Intn(15),
		},
		"security_hotspots": rand.Intn(10) + 1,
		"compliance_status": []string{"compliant", "non_compliant"}[rand.Intn(2)],
	}
}

func generatePerformanceAnalysis(projectPath string) map[string]interface{} {
	return map[string]interface{}{
		"performance_metrics": map[string]interface{}{
			"cpu_usage":      float64(20+rand.Intn(60)),
			"memory_usage":   float64(30+rand.Intn(50)),
			"response_time":  float64(10+rand.Intn(100)),
		},
		"bottlenecks": rand.Intn(5),
		"optimization_suggestions": rand.Intn(10) + 1,
	}
}

// generateLocalCode gera código local quando MCP não está disponível
func generateLocalCode(componentType, name string, spec map[string]interface{}) string {
	// Gerar código baseado no tipo de componente
	switch componentType {
	case "api":
		return generateAPICode(name, spec)
	case "service":
		return generateServiceCode(name, spec)
	case "cli":
		return generateCLICode(name, spec)
	default:
		return generateBasicLocalCode(name, spec)
	}
}

// generateAPICode gera código para componente API
func generateAPICode(name string, spec map[string]interface{}) string {
	port := "8080"
	if p, ok := spec["port"].(string); ok && p != "" {
		port = p
	}

	return fmt.Sprintf(`// Generated API code for %s
package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type %sServer struct {
	router *gin.Engine
}

func New%sServer() *%sServer {
	r := gin.Default()

	server := &%sServer{
		router: r,
	}

	server.setupRoutes()
	return server
}

func (s *%sServer) setupRoutes() {
	api := s.router.Group("/api/v1")
	{
		api.GET("/health", s.healthCheck)
		api.GET("/%s", s.getAll)
		api.POST("/%s", s.create)
		api.GET("/%s/:id", s.getByID)
		api.PUT("/%s/:id", s.update)
		api.DELETE("/%s/:id", s.delete)
	}
}

func (s *%sServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "%s",
		"port": "%s",
	})
}

// TODO: Implement CRUD operations
func (s *%sServer) getAll(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": []interface{}{}})
}

func (s *%sServer) create(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func (s *%sServer) getByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"id": c.Param("id")})
}

func (s *%sServer) update(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (s *%sServer) delete(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (s *%sServer) Run() error {
	return s.router.Run(":" + "%s")
}

func main() {
	server := New%sServer()
	if err := server.Run(); err != nil {
		panic(err)
	}
}`, name, name, name, name, name, name,
		strings.ToLower(name), strings.ToLower(name), strings.ToLower(name),
		strings.ToLower(name), strings.ToLower(name), name, name, port,
		name, name, name, name, name, port, name)
}

// generateServiceCode gera código para componente Service
func generateServiceCode(name string, spec map[string]interface{}) string {
	return fmt.Sprintf(`// Generated Service code for %s
package main

import (
	"context"
	"log"
	"time"
)

type %sService struct {
	name    string
	started time.Time
}

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	DoWork(ctx context.Context, input interface{}) (interface{}, error)
}

func New%sService() Service {
	return &%sService{
		name:    "%s",
		started: time.Now(),
	}
}

func (s *%sService) Start(ctx context.Context) error {
	log.Printf("Starting %s service...")

	// TODO: Add service initialization logic

	log.Printf("%s service started successfully")
	return nil
}

func (s *%sService) Stop(ctx context.Context) error {
	log.Printf("Stopping %s service...")

	// TODO: Add graceful shutdown logic

	log.Printf("%s service stopped")
	return nil
}

func (s *%sService) DoWork(ctx context.Context, input interface{}) (interface{}, error) {
	log.Printf("%s processing work...")

	// TODO: Add business logic

	return map[string]interface{}{
		"service": s.name,
		"result":  "processed",
		"uptime":  time.Since(s.started).String(),
	}, nil
}

func main() {
	ctx := context.Background()

	service := New%sService()

	if err := service.Start(ctx); err != nil {
		log.Fatalf("Failed to start service: %%v", err)
	}

	// TODO: Add service loop or signal handling

	// Simulate some work
	result, err := service.DoWork(ctx, map[string]interface{}{"test": "data"})
	if err != nil {
		log.Printf("Work failed: %%v", err)
	} else {
		log.Printf("Work result: %%+v", result)
	}

	if err := service.Stop(ctx); err != nil {
		log.Printf("Error stopping service: %%v", err)
	}
}`, name, name, name, name, name, name, name, name, name, name, name, name)
}

// generateCLICode gera código para componente CLI
func generateCLICode(name string, spec map[string]interface{}) string {
	return fmt.Sprintf(`// Generated CLI code for %s
package main

import (
	"flag"
	"fmt"
	"os"
)

type %sCLI struct {
	name    string
	version string
}

func New%sCLI() *%sCLI {
	return &%sCLI{
		name:    "%s",
		version: "1.0.0",
	}
}

func (cli *%sCLI) Run(args []string) error {
	var version bool
	var help bool
	var input string

	flagSet := flag.NewFlagSet(cli.name, flag.ContinueOnError)
	flagSet.BoolVar(&version, "version", false, "Show version information")
	flagSet.BoolVar(&help, "help", false, "Show help information")
	flagSet.StringVar(&input, "input", "", "Input file or data")

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if version {
		fmt.Printf("%%s v%%s\n", cli.name, cli.version)
		return nil
	}

	if help {
		cli.showHelp()
		return nil
	}

	if input == "" {
		return fmt.Errorf("input is required")
	}

	return cli.processInput(input)
}

func (cli *%sCLI) processInput(input string) error {
	fmt.Printf("Processing input: %%s\n", input)

	// TODO: Add main CLI logic

	result := map[string]interface{}{
		"processed": true,
		"input":     input,
		"output":    "processed_" + input,
	}

	fmt.Printf("Result: %%+v\n", result)
	return nil
}

func (cli *%sCLI) showHelp() {
	fmt.Printf(` + "`" + `Usage: %s [options]

Options:
  -version    Show version information
  -help       Show this help message
  -input      Input file or data (required)

Examples:
  %s -input data.json
  %s -input "raw data"
` + "`" + `, cli.name, cli.name, cli.name)
}

func main() {
	cli := New%sCLI()

	if err := cli.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %%v\n", err)
		os.Exit(1)
	}
}`, name, name, name, name, name, name, name, name, name, name)
}

// generateBasicLocalCode gera código básico
func generateBasicLocalCode(name string, spec map[string]interface{}) string {
	return fmt.Sprintf(`// Generated code for %s
package main

import (
	"fmt"
	"log"
)

type %s struct {
	Name    string
	Version string
}

func New%s(name string) *%s {
	return &%s{
		Name:    name,
		Version: "1.0.0",
	}
}

func (x *%s) Process(input interface{}) interface{} {
	fmt.Printf("%%s processing: %%v\n", x.Name, input)

	// TODO: Add processing logic

	result := map[string]interface{}{
		"processor": x.Name,
		"version":   x.Version,
		"input":     input,
		"output":    fmt.Sprintf("processed_by_%%s", x.Name),
	}

	return result
}

func (x *%s) Run() error {
	log.Printf("Running %s v%s", x.Name, x.Version)

	// TODO: Add main logic

	data := map[string]interface{}{
		"test":  true,
		"value": 42,
	}

	result := x.Process(data)
	fmt.Printf("Result: %%+v\n", result)

	return nil
}

func main() {
	x := New%s("%s")

	if err := x.Run(); err != nil {
		log.Fatalf("Error: %%v", err)
	}
}`, name, name, name, name, name, name, name, name, name, name)
}

// Helper function para extrair string de map com segurança
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}