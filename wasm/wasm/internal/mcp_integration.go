package internal

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"
	"time"
)

// MCPClient representa um cliente para interagir com o MCP Go-Architect
type MCPClient struct {
	Initialized bool
	Config      map[string]interface{}
	LastUsed    time.Time
}

// AnalysisRequest representa uma requisição de análise para o MCP
type AnalysisRequest struct {
	ProjectPath    string      `json:"project_path"`
	AnalysisType   string      `json:"analysis_type"`
	Options        map[string]interface{} `json:"options,omitempty"`
	Format         string      `json:"format"`
	IncludeDetails bool        `json:"include_details"`
}

// AnalysisResponse representa a resposta do MCP
type AnalysisResponse struct {
	Success        bool                   `json:"success"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Error          string                 `json:"error,omitempty"`
	Warnings       []string               `json:"warnings,omitempty"`
	Metrics        map[string]interface{} `json:"metrics,omitempty"`
	Score          *ScoreDetails          `json:"score,omitempty"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
}

// ScoreDetails representa detalhes do Score Vertikon
type ScoreDetails struct {
	Score      float64                `json:"score"`
	Band       string                 `json:"band"`
	Categories map[string]float64     `json:"categories"`
	Issues     []ScoreIssue           `json:"issues"`
	Features   []string               `json:"features"`
	Rules      map[string]RuleResult  `json:"rules"`
}

// ScoreIssue representa uma issue do Score Vertikon
type ScoreIssue struct {
	Rule     string `json:"rule"`
	Category string `json:"category"`
	Level    string `json:"level"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
}

// RuleResult representa o resultado de uma regra
type RuleResult struct {
	Status     string        `json:"status"`
	Message    string        `json:"message"`
	Details    interface{}   `json:"details,omitempty"`
	Suggestions []string     `json:"suggestions,omitempty"`
}

// CodeGenerationRequest representa uma requisição de geração de código
type CodeGenerationRequest struct {
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Options      map[string]interface{} `json:"options,omitempty"`
	Template     string                 `json:"template,omitempty"`
	Features     []string               `json:"features,omitempty"`
}

// CodeGenerationResponse representa a resposta de geração de código
type CodeGenerationResponse struct {
	Success       bool                   `json:"success"`
	GeneratedCode string                 `json:"generated_code,omitempty"`
	Files         []GeneratedFile        `json:"files,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Warnings      []string               `json:"warnings,omitempty"`
	Instructions  []string               `json:"instructions,omitempty"`
}

// GeneratedFile representa um arquivo gerado
type GeneratedFile struct {
	Path     string `json:"path"`
	Content  string `json:"content"`
	Type     string `json:"type"`
	Size     int    `json:"size"`
	ReadOnly bool   `json:"read_only"`
}

var mcpClient MCPClient

// InitializeMCPClient inicializa o cliente MCP
func InitializeMCPClient(config map[string]interface{}) error {
	mcpClient = MCPClient{
		Initialized: true,
		Config:      config,
		LastUsed:    time.Now(),
	}
	return nil
}

// AnalyzeWithMCP realiza análise usando o MCP Go-Architect
func AnalyzeWithMCP(config map[string]interface{}) *AnalysisResponse {
	if !mcpClient.Initialized {
		return &AnalysisResponse{
			Success: false,
			Error:   "MCP client not initialized",
		}
	}

	startTime := time.Now()

	projectPath := getString(config, "project_path")
	analysisType := getString(config, "analysis_type")
	if analysisType == "" {
		analysisType = "full"
	}

	request := AnalysisRequest{
		ProjectPath:    projectPath,
		AnalysisType:   analysisType,
		Options:        getOptions(config),
		Format:         "json",
		IncludeDetails: true,
	}

	// Simular chamada ao MCP Go-Architect
	response := simulateMCPAnalysis(request)

	response.ProcessingTime = time.Since(startTime)
	response.Timestamp = startTime

	mcpClient.LastUsed = time.Now()
	return response
}

// GenerateCodeWithMCP gera código usando o MCP Go-Architect
func GenerateCodeWithMCP(spec map[string]interface{}) *CodeGenerationResponse {
	if !mcpClient.Initialized {
		return &CodeGenerationResponse{
			Success: false,
			Error:   "MCP client not initialized",
		}
	}

	componentType := getString(spec, "component_type")
	name := getString(spec, "name")
	if componentType == "" {
		componentType = "service"
	}
	if name == "" {
		name = "GeneratedService"
	}

	request := CodeGenerationRequest{
		Type:     componentType,
		Name:     name,
		Options:  spec,
		Template: "default",
		Features: extractFeatures(spec),
	}

	// Simular geração de código
	response := simulateCodeGeneration(request)

	return response
}

// ValidateWithMCP valida configurações usando o MCP
func ValidateWithMCP(config map[string]interface{}) *AnalysisResponse {
	if !mcpClient.Initialized {
		return &AnalysisResponse{
			Success: false,
			Error:   "MCP client not initialized",
		}
	}

	startTime := time.Now()

	validationType := getString(config, "type")

	// Simular validação
	response := simulateMCPValidation(config, validationType)

	response.ProcessingTime = time.Since(startTime)
	response.Timestamp = startTime

	mcpClient.LastUsed = time.Now()
	return response
}

// simulateMCPAnalysis simula uma chamada ao MCP Go-Architect
func simulateMCPAnalysis(request AnalysisRequest) *AnalysisResponse {
	// Simular delay de processamento
	time.Sleep(time.Duration(100+rand.Intn(500)) * time.Millisecond)

	// Gerar score baseado no tipo de análise
	var score *ScoreDetails
	var data map[string]interface{}

	switch request.AnalysisType {
	case "quick":
		score = generateQuickScore(request.ProjectPath)
		data = map[string]interface{}{
			"analysis_type": "quick",
			"summary":       "Quick project analysis completed",
			"recommendations": generateQuickRecommendationsMCP(request.ProjectPath),
		}
	case "security":
		score = generateSecurityScore(request.ProjectPath)
		data = map[string]interface{}{
			"analysis_type": "security",
			"summary":       "Security analysis completed",
			"vulnerabilities": generateSecurityFindings(request.ProjectPath),
		}
	case "performance":
		score = generatePerformanceScore(request.ProjectPath)
		data = map[string]interface{}{
			"analysis_type": "performance",
			"summary":       "Performance analysis completed",
			"bottlenecks": generateBottlenecks(request.ProjectPath),
		}
	default:
		score = generateFullScore(request.ProjectPath)
		data = map[string]interface{}{
			"analysis_type": "full",
			"summary":       "Comprehensive project analysis completed",
			"features":      detectFeatures(request.ProjectPath),
			"architecture":  analyzeArchitecture(request.ProjectPath),
		}
	}

	return &AnalysisResponse{
		Success: true,
		Data:    data,
		Score:   score,
		Metrics: map[string]interface{}{
			"files_analyzed":    rand.Intn(50) + 10,
			"lines_processed":   rand.Intn(10000) + 1000,
			"rules_executed":    rand.Intn(100) + 50,
			"issues_found":      len(score.Issues),
			"warnings_count":    len(score.Issues) / 3,
		},
	}
}

// simulateCodeGeneration simula geração de código
func simulateCodeGeneration(request CodeGenerationRequest) *CodeGenerationResponse {
	// Simular delay
	time.Sleep(time.Duration(200+rand.Intn(800)) * time.Millisecond)

	files := []GeneratedFile{}

	// Gerar arquivos baseados no tipo
	switch request.Type {
	case "api":
		files = generateAPIFiles(request.Name, request.Options)
	case "service":
		files = generateServiceFiles(request.Name, request.Options)
	case "cli":
		files = generateCLIFiles(request.Name, request.Options)
	default:
		files = generateBasicFiles(request.Name, request.Options)
	}

	return &CodeGenerationResponse{
		Success: true,
		Files:   files,
		Warnings: []string{
			"Review generated code before using in production",
			"Add appropriate error handling",
			"Consider adding unit tests",
		},
		Instructions: []string{
			"1. Review the generated files",
			"2. Add business logic to service layer",
			"3. Configure database connections",
			"4. Add proper error handling",
			"5. Write unit tests",
		},
	}
}

// simulateMCPValidation simula validação com MCP
func simulateMCPValidation(config map[string]interface{}, validationType string) *AnalysisResponse {
	// Simular delay
	time.Sleep(time.Duration(50+rand.Intn(200)) * time.Millisecond)

	issues := []ScoreIssue{}

	// Validar basedo no tipo
	switch validationType {
	case "project":
		issues = validateProjectConfig(config)
	case "deployment":
		issues = validateDeploymentConfig(config)
	case "module":
		issues = validateModuleConfig(config)
	default:
		issues = validateGenericConfig(config)
	}

	score := calculateValidationScore(issues)

	return &AnalysisResponse{
		Success: true,
		Data: map[string]interface{}{
			"validation_type": validationType,
			"valid":           len(issues) == 0,
			"issues_count":    len(issues),
		},
		Score: score,
		Warnings: extractWarnings(issues),
	}
}

// Funções auxiliares para geração de scores

func generateQuickScore(projectPath string) *ScoreDetails {
	return &ScoreDetails{
		Score:  70 + rand.Float64()*20,
		Band:   []string{"A", "B", "C"}[rand.Intn(3)],
		Categories: map[string]float64{
			"structure":   70 + rand.Float64()*30,
			"naming":      80 + rand.Float64()*20,
			"formatting":  90 + rand.Float64()*10,
		},
		Issues:   generateQuickIssuesMCP(projectPath),
		Features: []string{"quick", "structure", "basic"},
		Rules:    generateRuleResults(5),
	}
}

func generateSecurityScore(projectPath string) *ScoreDetails {
	return &ScoreDetails{
		Score:  60 + rand.Float64()*30,
		Band:   []string{"B", "C", "D"}[rand.Intn(3)],
		Categories: map[string]float64{
			"security":    50 + rand.Float64()*40,
			"dependencies": 60 + rand.Float64()*30,
			"data_flow":   70 + rand.Float64()*20,
		},
		Issues:   generateSecurityIssuesMCP(projectPath),
		Features: []string{"security", "vulnerabilities", "dependencies"},
		Rules:    generateRuleResults(15),
	}
}

func generatePerformanceScore(projectPath string) *ScoreDetails {
	return &ScoreDetails{
		Score:  65 + rand.Float64()*25,
		Band:   []string{"B", "C"}[rand.Intn(2)],
		Categories: map[string]float64{
			"performance": 60 + rand.Float64()*30,
			"memory":      70 + rand.Float64()*20,
			"concurrency": 50 + rand.Float64()*40,
		},
		Issues:   generatePerformanceIssuesMCP(projectPath),
		Features: []string{"performance", "memory", "concurrency"},
		Rules:    generateRuleResults(12),
	}
}

func generateFullScore(projectPath string) *ScoreDetails {
	return &ScoreDetails{
		Score:  50 + rand.Float64()*40,
		Band:   []string{"C", "B", "A"}[rand.Intn(3)],
		Categories: map[string]float64{
			"architecture":   60 + rand.Float64()*30,
			"design":        70 + rand.Float64()*20,
			"error_handling": 50 + rand.Float40()*40,
			"testing":        40 + rand.Float40()*50,
			"documentation":  30 + rand.Float40()*60,
		},
		Issues:   generateFullIssuesMCP(projectPath),
		Features: []string{"full", "architecture", "design", "testing", "documentation"},
		Rules:    generateRuleResults(50),
	}
}

// Funções auxiliares para geração de arquivos

func generateAPIFiles(name string, options map[string]interface{}) []GeneratedFile {
	return []GeneratedFile{
		{
			Path:    fmt.Sprintf("cmd/%s/main.go", strings.ToLower(name)),
			Content: generateMainTemplate(name, "api"),
			Type:    "main",
			Size:    len(generateMainTemplate(name, "api")),
		},
		{
			Path:    fmt.Sprintf("internal/%s/handler.go", strings.ToLower(name)),
			Content: generateHandlerTemplate(name),
			Type:    "handler",
			Size:    len(generateHandlerTemplate(name)),
		},
		{
			Path:    fmt.Sprintf("internal/%s/service.go", strings.ToLower(name)),
			Content: generateServiceTemplate(name),
			Type:    "service",
			Size:    len(generateServiceTemplate(name)),
		},
		{
			Path:    fmt.Sprintf("internal/%s/repository.go", strings.ToLower(name)),
			Content: generateRepositoryTemplate(name),
			Type:    "repository",
			Size:    len(generateRepositoryTemplate(name)),
		},
	}
}

func generateServiceFiles(name string, options map[string]interface{}) []GeneratedFile {
	return []GeneratedFile{
		{
			Path:    fmt.Sprintf("cmd/%s/main.go", strings.ToLower(name)),
			Content: generateMainTemplate(name, "service"),
			Type:    "main",
			Size:    len(generateMainTemplate(name, "service")),
		},
		{
			Path:    fmt.Sprintf("internal/%s/service.go", strings.ToLower(name)),
			Content: generateServiceTemplate(name),
			Type:    "service",
			Size:    len(generateServiceTemplate(name)),
		},
		{
			Path:    fmt.Sprintf("internal/%s/worker.go", strings.ToLower(name)),
			Content: generateWorkerTemplate(name),
			Type:    "worker",
			Size:    len(generateWorkerTemplate(name)),
		},
	}
}

func generateCLIFiles(name string, options map[string]interface{}) []GeneratedFile {
	return []GeneratedFile{
		{
			Path:    fmt.Sprintf("cmd/%s/main.go", strings.ToLower(name)),
			Content: generateMainTemplate(name, "cli"),
			Type:    "main",
			Size:    len(generateMainTemplate(name, "cli")),
		},
		{
			Path:    fmt.Sprintf("cmd/%s/root.go", strings.ToLower(name)),
			Content: generateRootCommandTemplate(name),
			Type:    "command",
			Size:    len(generateRootCommandTemplate(name)),
		},
		{
			Path:    fmt.Sprintf("internal/%s/processor.go", strings.ToLower(name)),
			Content: generateProcessorTemplate(name),
			Type:    "processor",
			Size:    len(generateProcessorTemplate(name)),
		},
	}
}

func generateBasicFiles(name string, options map[string]interface{}) []GeneratedFile {
	return []GeneratedFile{
		{
			Path:    fmt.Sprintf("%s.go", strings.ToLower(name)),
			Content: generateBasicTemplate(name),
			Type:    "implementation",
			Size:    len(generateBasicTemplate(name)),
		},
		{
			Path:    fmt.Sprintf("%s_test.go", strings.ToLower(name)),
			Content: generateTestTemplate(name),
			Type:    "test",
			Size:    len(generateTestTemplate(name)),
		},
	}
}

// Templates de código (simulados)

func generateMainTemplate(name, appType string) string {
	return fmt.Sprintf(`// Generated %s main for %s
package main

import (
    "fmt"
    "log"
    "os"
)

func main() {
    fmt.Println("%s %s starting...")

    // TODO: Add initialization logic

    if err := run(); err != nil {
        log.Fatalf("Application error: %%v", err)
        os.Exit(1)
    }
}

func run() error {
    // TODO: Add main application logic
    return nil
}`, appType, name, name)
}

func generateHandlerTemplate(name string) string {
	return fmt.Sprintf(`package %s

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type Handler struct {
    service Service
}

func NewHandler(service Service) *Handler {
    return &Handler{
        service: service,
    }
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
    api := r.Group("/api/v1")
    {
        api.GET("/%s", h.GetAll)
        api.GET("/%s/:id", h.GetByID)
        api.POST("/%s", h.Create)
        api.PUT("/%s/:id", h.Update)
        api.DELETE("/%s/:id", h.Delete)
    }
}

// TODO: Implement handler methods
`, strings.ToLower(name), strings.ToLower(name), strings.ToLower(name), strings.ToLower(name), strings.ToLower(name), strings.ToLower(name))
}

func generateServiceTemplate(name string) string {
	return fmt.Sprintf(`package %s

import (
    "context"
    "fmt"
)

type Service interface {
    // TODO: Define service interface
}

type service struct {
    // TODO: Add dependencies
}

func NewService() Service {
    return &service{
        // TODO: Initialize dependencies
    }
}

// TODO: Implement service methods
func (s *service) DoSomething(ctx context.Context) error {
    return fmt.Errorf("not implemented")
}`, strings.ToLower(name))
}

func generateRepositoryTemplate(name string) string {
	return fmt.Sprintf(`package %s

import (
    "context"
)

type Repository interface {
    // TODO: Define repository interface
}

type repository struct {
    // TODO: Add database connection
}

func NewRepository() Repository {
    return &repository{
        // TODO: Initialize database connection
    }
}

// TODO: Implement repository methods
func (r *repository) FindByID(ctx context.Context, id string) (interface{}, error) {
    return nil, fmt.Errorf("not implemented")
}`, strings.ToLower(name))
}

func generateWorkerTemplate(name string) string {
	return fmt.Sprintf(`package %s

import (
    "context"
    "log"
)

type Worker struct {
    // TODO: Add worker dependencies
}

func NewWorker() *Worker {
    return &Worker{
        // TODO: Initialize worker
    }
}

func (w *Worker) Start(ctx context.Context) error {
    log.Println("Worker starting...")

    // TODO: Add worker logic

    <-ctx.Done()
    log.Println("Worker stopped")
    return nil
}`, strings.ToLower(name))
}

func generateRootCommandTemplate(name string) string {
	return fmt.Sprintf(`package main

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "%s",
    Short: "%s CLI tool",
    Long:  `%s is a command-line interface tool`,
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("%s called with args:", args)
    },
}

func Execute() error {
    return rootCmd.Execute()
}

func init() {
    // TODO: Add command flags
}

func main() {
    if err := Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %%v\\n", err)
        os.Exit(1)
    }
}`, strings.ToLower(name), name, name, name)
}

func generateProcessorTemplate(name string) string {
	return fmt.Sprintf(`package %s

import (
    "context"
    "log"
)

type Processor struct {
    // TODO: Add processor dependencies
}

func NewProcessor() *Processor {
    return &Processor{
        // TODO: Initialize processor
    }
}

func (p *Processor) Process(ctx context.Context, input interface{}) (interface{}, error) {
    log.Printf("Processing input: %%v", input)

    // TODO: Add processing logic

    return nil, fmt.Errorf("not implemented")
}`, strings.ToLower(name))
}

func generateBasicTemplate(name string) string {
	return fmt.Sprintf(`// Generated %s implementation
package main

import "fmt"

type %s struct {
    // TODO: Add fields
}

func New%s() *%s {
    return &%s{
        // TODO: Initialize fields
    }
}

func (x *%s) Do() error {
    fmt.Println("%s doing work...")

    // TODO: Add implementation

    return nil
}`, name, name, name, name, name, name, name)
}

func generateTestTemplate(name string) string {
	return fmt.Sprintf(`package main

import (
    "testing"
)

func Test%s(t *testing.T) {
    // TODO: Add unit tests

    x := New%s()

    if err := x.Do(); err != nil {
        t.Errorf("Expected no error, got: %%v", err)
    }
}`, name, name)
}

// Funções auxiliares

func getOptions(config map[string]interface{}) map[string]interface{} {
	if opts, ok := config["options"].(map[string]interface{}); ok {
		return opts
	}
	return make(map[string]interface{})
}

func extractFeatures(spec map[string]interface{}) []string {
	if features, ok := spec["features"].([]string); ok {
		return features
	}
	return []string{"basic"}
}

func generateQuickRecommendationsMCP(projectPath string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"type":        "structure",
			"priority":    "medium",
			"title":       "Consider organizing packages",
			"description": "Project structure could benefit from better package organization",
		},
		{
			"type":        "testing",
			"priority":    "high",
			"title":       "Add unit tests",
			"description": "Consider adding comprehensive unit tests",
		},
	}
}

func generateSecurityFindings(projectPath string) map[string]interface{} {
	return map[string]interface{}{
		"vulnerabilities": map[string]int{
			"critical": rand.Intn(2),
			"high":     rand.Intn(3),
			"medium":   rand.Intn(8),
			"low":      rand.Intn(15),
		},
		"hotspots": rand.Intn(10) + 1,
	}
}

func generateBottlenecks(projectPath string) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"area":        "database",
			"impact":      "high",
			"description": "Potential database query optimization needed",
		},
		{
			"area":        "memory",
			"impact":      "medium",
			"description": "Consider implementing memory pooling",
		},
	}
}

func detectFeatures(projectPath string) []string {
	features := []string{"go-mod"}

	if rand.Intn(2) == 0 {
		features = append(features, "gin")
	}
	if rand.Intn(2) == 0 {
		features = append(features, "nats")
	}
	if rand.Intn(2) == 0 {
		features = append(features, "postgresql")
	}
	if rand.Intn(2) == 0 {
		features = append(features, "docker")
	}

	return features
}

func analyzeArchitecture(projectPath string) map[string]interface{} {
	return map[string]interface{}{
		"style":     []string{"layered", "hexagonal", "clean"}[rand.Intn(3)],
		"layers":    rand.Intn(4) + 3,
		"packages":  rand.Intn(15) + 5,
		"coupling":  []string{"low", "medium"}[rand.Intn(2)],
		"cohesion":  []string{"high", "medium"}[rand.Intn(2)],
	}
}

// Outras funções auxiliares...

func generateQuickIssuesMCP(projectPath string) []ScoreIssue {
	return []ScoreIssue{
		{
			Rule:     "go-missing-tests",
			Category: "testing",
			Level:    "warning",
			File:     "main.go",
			Line:     1,
			Message:  "File lacks unit tests",
		},
	}
}

func generateSecurityIssuesMCP(projectPath string) []ScoreIssue {
	return []ScoreIssue{
		{
			Rule:     "go-hardcoded-credentials",
			Category: "security",
			Level:    "error",
			File:     "config.go",
			Line:     15,
			Message:  "Potential hardcoded credentials detected",
		},
	}
}

func generatePerformanceIssuesMCP(projectPath string) []ScoreIssue {
	return []ScoreIssue{
		{
			Rule:     "go-inefficient-loop",
			Category: "performance",
			Level:    "warning",
			File:     "processor.go",
			Line:     42,
			Message:  "Inefficient loop detected",
		},
	}
}

func generateFullIssuesMCP(projectPath string) []ScoreIssue {
	issues := generateQuickIssuesMCP(projectPath)
	issues = append(issues, generateSecurityIssuesMCP(projectPath)...)
	issues = append(issues, generatePerformanceIssuesMCP(projectPath)...)
	return issues
}

func generateRuleResults(count int) map[string]RuleResult {
	rules := make(map[string]RuleResult)
	ruleNames := []string{
		"go-fmt", "go-vet", "go-lint", "go-errcheck",
		"go-staticcheck", "go-gosec", "go-misspell",
		"go-imports", "go-exports", "go-dupl",
	}

	for i := 0; i < count && i < len(ruleNames); i++ {
		rules[ruleNames[i]] = RuleResult{
			Status:  []string{"pass", "fail", "warn"}[rand.Intn(3)],
			Message: fmt.Sprintf("Rule %s executed", ruleNames[i]),
		}
	}

	return rules
}

func validateProjectConfig(config map[string]interface{}) []ScoreIssue {
	issues := []ScoreIssue{}

	if getString(config, "project_name") == "" {
		issues = append(issues, ScoreIssue{
			Rule:     "project-name-required",
			Category: "validation",
			Level:    "error",
			Message:  "Project name is required",
		})
	}

	if getString(config, "module_path") == "" {
		issues = append(issues, ScoreIssue{
			Rule:     "module-path-required",
			Category: "validation",
			Level:    "error",
			Message:  "Module path is required",
		})
	}

	return issues
}

func validateDeploymentConfig(config map[string]interface{}) []ScoreIssue {
	issues := []ScoreIssue{}

	if getString(config, "target") == "" {
		issues = append(issues, ScoreIssue{
			Rule:     "deployment-target-required",
			Category: "validation",
			Level:    "error",
			Message:  "Deployment target is required",
		})
	}

	return issues
}

func validateModuleConfig(config map[string]interface{}) []ScoreIssue {
	issues := []ScoreIssue{}

	if getString(config, "name") == "" {
		issues = append(issues, ScoreIssue{
			Rule:     "module-name-required",
			Category: "validation",
			Level:    "error",
			Message:  "Module name is required",
		})
	}

	return issues
}

func validateGenericConfig(config map[string]interface{}) []ScoreIssue {
	return []ScoreIssue{}
}

func calculateValidationScore(issues []ScoreIssue) *ScoreDetails {
	score := 100.0
	for _, issue := range issues {
		switch issue.Level {
		case "error":
			score -= 20
		case "warning":
			score -= 10
		case "info":
			score -= 5
		}
	}

	if score < 0 {
		score = 0
	}

	band := "F"
	if score >= 90 {
		band = "A"
	} else if score >= 80 {
		band = "B"
	} else if score >= 70 {
		band = "C"
	} else if score >= 60 {
		band = "D"
	}

	return &ScoreDetails{
		Score:    score,
		Band:     band,
		Issues:   issues,
		Features: []string{"validation"},
	}
}

func extractWarnings(issues []ScoreIssue) []string {
	warnings := []string{}
	for _, issue := range issues {
		if issue.Level == "warning" {
			warnings = append(warnings, issue.Message)
		}
	}
	return warnings
}