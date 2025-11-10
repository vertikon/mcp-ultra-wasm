package sdk

import (
	"time"
)

// Request types para o SDK

type AnalysisRequest struct {
	TaskID      string                 `json:"task_id"`
	ProjectPath string                 `json:"project_path"`
	Type        string                 `json:"type"` // "full", "quick", "security", "performance"
	Options     map[string]interface{} `json:"options,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
}

type GenerationRequest struct {
	TaskID        string                 `json:"task_id"`
	ComponentType string                 `json:"component_type"`
	Name          string                 `json:"name"`
	Language      string                 `json:"language"`
	Specification map[string]interface{} `json:"specification"`
	Templates     []string               `json:"templates,omitempty"`
}

type ValidationRequest struct {
	TaskID     string                 `json:"task_id"`
	Type       string                 `json:"type"` // "project", "deployment", "security"
	Config     map[string]interface{} `json:"config"`
	Rules      []string               `json:"rules,omitempty"`
	StrictMode bool                   `json:"strict_mode,omitempty"`
}

// Response types do SDK

type AnalysisResult struct {
	TaskID      string                 `json:"task_id"`
	ProjectPath string                 `json:"project_path"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type GenerationResult struct {
	TaskID         string                 `json:"task_id"`
	ComponentType  string                 `json:"component_type"`
	Name           string                 `json:"name"`
	Language       string                 `json:"language"`
	Status         string                 `json:"status"`
	GeneratedCode  map[string]interface{} `json:"generated_code"`
	FilesGenerated []string               `json:"files_generated"`
	Metadata       map[string]interface{} `json:"metadata"`
}

type ValidationResult struct {
	TaskID   string                 `json:"task_id"`
	Type     string                 `json:"type"`
	Status   string                 `json:"status"`
	Valid    bool                   `json:"valid"`
	Errors   []string               `json:"errors"`
	Warnings []string               `json:"warnings"`
	Score    float64                `json:"score,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Task status para tracking

type TaskStatus struct {
	ID            string                 `json:"id"`
	CorrelationID string                 `json:"correlation_id"`
	Type          string                 `json:"type"`
	Status        string                 `json:"status"`   // "pending", "running", "completed", "failed"
	Progress      int                    `json:"progress"` // 0-100
	Message       string                 `json:"message,omitempty"`
	Result        map[string]interface{} `json:"result,omitempty"`
	Error         string                 `json:"error,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Plugin registry

type PluginInfo struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Capabilities []string               `json:"capabilities"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Enabled      bool                   `json:"enabled"`
	Health       string                 `json:"health"` // "healthy", "unhealthy", "unknown"
	LastUsed     *time.Time             `json:"last_used,omitempty"`
}

// Event types para comunicação

type TaskEvent struct {
	Type          string                 `json:"type"` // "task_created", "task_started", "task_progress", "task_completed", "task_failed"
	TaskID        string                 `json:"task_id"`
	CorrelationID string                 `json:"correlation_id"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

type SystemEvent struct {
	Type      string                 `json:"type"`   // "plugin_registered", "plugin_error", "sdk_status"
	Source    string                 `json:"source"` // "web-wasm", "sdk", "plugin"
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// Health check

type HealthStatus struct {
	Status     string                 `json:"status"` // "healthy", "degraded", "unhealthy"
	Timestamp  time.Time              `json:"timestamp"`
	Components map[string]interface{} `json:"components"`
	Version    string                 `json:"version"`
	Uptime     string                 `json:"uptime"`
}

// Metrics

type WebWasmMetrics struct {
	TasksTotal      int64     `json:"tasks_total"`
	TasksActive     int       `json:"tasks_active"`
	TasksCompleted  int64     `json:"tasks_completed"`
	TasksFailed     int64     `json:"tasks_failed"`
	AvgResponseTime float64   `json:"avg_response_time"`
	LastTaskTime    time.Time `json:"last_task_time"`
	WasmCalls       int64     `json:"wasm_calls"`
	SDKCalls        int64     `json:"sdk_calls"`
	WSErrors        int64     `json:"ws_errors"`
}

// Configuration

type WebWasmConfig struct {
	Server    ServerConfig    `json:"server"`
	WASM      WASMConfig      `json:"wasm"`
	SDK       SDKConfig       `json:"sdk"`
	NATS      NATSConfig      `json:"nats"`
	WebSocket WebSocketConfig `json:"websocket"`
	Security  SecurityConfig  `json:"security"`
}

type ServerConfig struct {
	Port         string        `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

type WASMConfig struct {
	CacheEnabled bool          `json:"cache_enabled"`
	CacheSize    int           `json:"cache_size"`
	Timeout      time.Duration `json:"timeout"`
	MaxMemory    int64         `json:"max_memory"`
}

type SDKConfig struct {
	Address         string        `json:"address"`
	Timeout         time.Duration `json:"timeout"`
	Retries         int           `json:"retries"`
	EnableCache     bool          `json:"enable_cache"`
	DefaultPlugin   string        `json:"default_plugin"`
	HealthCheckTime time.Duration `json:"health_check_time"`
}

type NATSConfig struct {
	URL           string        `json:"url"`
	MaxReconnects int           `json:"max_reconnects"`
	ReconnectWait time.Duration `json:"reconnect_wait"`
	Timeout       time.Duration `json:"timeout"`
}

type WebSocketConfig struct {
	PingInterval      time.Duration `json:"ping_interval"`
	PongWait          time.Duration `json:"pong_wait"`
	WriteWait         time.Duration `json:"write_wait"`
	MaxMessageSize    int64         `json:"max_message_size"`
	EnableCompression bool          `json:"enable_compression"`
}

type SecurityConfig struct {
	EnableAuth      bool     `json:"enable_auth"`
	EnableCORS      bool     `json:"enable_cors"`
	EnableRateLimit bool     `json:"enable_rate_limit"`
	AllowedOrigins  []string `json:"allowed_origins"`
	RateLimitRPS    int      `json:"rate_limit_rps"`
	JWTSecret       string   `json:"jwt_secret,omitempty"`
}

// Constants
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"

	TaskTypeAnalysis   = "analysis"
	TaskTypeGeneration = "generation"
	TaskTypeValidation = "validation"

	HealthStatusHealthy   = "healthy"
	HealthStatusDegraded  = "degraded"
	HealthStatusUnhealthy = "unhealthy"

	PluginHealthHealthy   = "healthy"
	PluginHealthUnhealthy = "unhealthy"
	PluginHealthUnknown   = "unknown"
)

// Helper functions

func NewTaskStatus(taskID, correlationID, taskType string) *TaskStatus {
	now := time.Now().UTC()
	return &TaskStatus{
		ID:            taskID,
		CorrelationID: correlationID,
		Type:          taskType,
		Status:        TaskStatusPending,
		Progress:      0,
		CreatedAt:     now,
		UpdatedAt:     now,
		Metadata:      make(map[string]interface{}),
	}
}

func (t *TaskStatus) UpdateProgress(progress int, message string) {
	t.Progress = progress
	t.Message = message
	t.UpdatedAt = time.Now().UTC()
}

func (t *TaskStatus) Complete(result map[string]interface{}) {
	t.Status = TaskStatusCompleted
	t.Progress = 100
	t.Result = result
	now := time.Now().UTC()
	t.CompletedAt = &now
	t.UpdatedAt = now
}

func (t *TaskStatus) Fail(err error) {
	t.Status = TaskStatusFailed
	t.Error = err.Error()
	now := time.Now().UTC()
	t.CompletedAt = &now
	t.UpdatedAt = now
}

func NewAnalysisRequest(taskID, projectPath, analysisType string) *AnalysisRequest {
	return &AnalysisRequest{
		TaskID:      taskID,
		ProjectPath: projectPath,
		Type:        analysisType,
		Options:     make(map[string]interface{}),
		Timeout:     30 * time.Second,
	}
}

func NewGenerationRequest(taskID, componentType, name, language string) *GenerationRequest {
	return &GenerationRequest{
		TaskID:        taskID,
		ComponentType: componentType,
		Name:          name,
		Language:      language,
		Specification: make(map[string]interface{}),
		Templates:     make([]string, 0),
	}
}

func NewValidationRequest(taskID, validationType string, config map[string]interface{}) *ValidationRequest {
	return &ValidationRequest{
		TaskID:     taskID,
		Type:       validationType,
		Config:     config,
		Rules:      make([]string, 0),
		StrictMode: false,
	}
}
