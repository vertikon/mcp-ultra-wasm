package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/logger"
)

// OperationType represents different types of operations
type OperationType string

const (
	OperationMaintenance   OperationType = "maintenance"
	OperationUpgrade       OperationType = "upgrade"
	OperationScaling       OperationType = "scaling"
	OperationBackup        OperationType = "backup"
	OperationRestore       OperationType = "restore"
	OperationDiagnostics   OperationType = "diagnostics"
	OperationCleanup       OperationType = "cleanup"
	OperationConfiguration OperationType = "configuration"
	OperationSecurityPatch OperationType = "security_patch"
)

// OperationStatus represents the status of an operation
type OperationStatus string

const (
	StatusPending   OperationStatus = "pending"
	StatusRunning   OperationStatus = "running"
	StatusCompleted OperationStatus = "completed"
	StatusFailed    OperationStatus = "failed"
	StatusCanceled  OperationStatus = "canceled"
)

// Operation represents a system operation
type Operation struct {
	ID          string          `json:"id"`
	Type        OperationType   `json:"type"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      OperationStatus `json:"status"`

	// Timing
	CreatedAt   time.Time     `json:"created_at"`
	StartedAt   *time.Time    `json:"started_at,omitempty"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	Duration    time.Duration `json:"duration"`
	Timeout     time.Duration `json:"timeout"`

	// Execution details
	Steps       []OperationStep `json:"steps"`
	CurrentStep int             `json:"current_step"`
	Progress    float64         `json:"progress"` // 0-100

	// Metadata
	Parameters map[string]interface{} `json:"parameters"`
	Context    map[string]string      `json:"context"`
	Tags       []string               `json:"tags"`

	// Results
	Result map[string]interface{} `json:"result,omitempty"`
	Logs   []string               `json:"logs"`
	Errors []string               `json:"errors"`

	// Control
	Cancelable   bool `json:"cancelable"`
	Rollbackable bool `json:"rollbackable"`

	// Execution control
	ctx    context.Context
	cancel context.CancelFunc
}

// OperationStep represents a step within an operation
type OperationStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      OperationStatus        `json:"status"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Retryable   bool                   `json:"retryable"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
}

// OperationExecutor defines the interface for operation execution
type OperationExecutor interface {
	Execute(ctx context.Context, operation *Operation) error
	Rollback(ctx context.Context, operation *Operation) error
	Validate(operation *Operation) error
}

// OperationsManager manages system operations and procedures
type OperationsManager struct {
	mu sync.RWMutex

	// Operations tracking
	operations map[string]*Operation
	history    []Operation
	maxHistory int

	// Executors
	executors map[OperationType]OperationExecutor

	// Configuration
	config OperationsConfig
	logger *logger.Logger

	// Background processing
	workerPool chan *Operation
	workers    int
	stopCh     chan struct{}
	running    bool
}

// OperationsConfig configures operations management
type OperationsConfig struct {
	MaxConcurrentOps int           `json:"max_concurrent_ops"`
	DefaultTimeout   time.Duration `json:"default_timeout"`
	MaxHistorySize   int           `json:"max_history_size"`
	EnableMetrics    bool          `json:"enable_metrics"`
	EnableAuditLog   bool          `json:"enable_audit_log"`
	WorkerPoolSize   int           `json:"worker_pool_size"`
	OperationRetries int           `json:"operation_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
}

// DefaultOperationsConfig returns default operations configuration
func DefaultOperationsConfig() OperationsConfig {
	return OperationsConfig{
		MaxConcurrentOps: 5,
		DefaultTimeout:   30 * time.Minute,
		MaxHistorySize:   100,
		EnableMetrics:    true,
		EnableAuditLog:   true,
		WorkerPoolSize:   3,
		OperationRetries: 3,
		RetryDelay:       5 * time.Second,
	}
}

// NewOperationsManager creates a new operations manager
func NewOperationsManager(config OperationsConfig, logger *logger.Logger) *OperationsManager {
	return &OperationsManager{
		operations: make(map[string]*Operation),
		history:    make([]Operation, 0, config.MaxHistorySize),
		maxHistory: config.MaxHistorySize,
		executors:  make(map[OperationType]OperationExecutor),
		config:     config,
		logger:     logger,
		workerPool: make(chan *Operation, config.MaxConcurrentOps),
		workers:    config.WorkerPoolSize,
		stopCh:     make(chan struct{}),
	}
}

// RegisterExecutor registers an operation executor
func (om *OperationsManager) RegisterExecutor(opType OperationType, executor OperationExecutor) {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.executors[opType] = executor
	om.logger.Info("Operation executor registered", "type", opType)
}

// Start starts the operations manager
func (om *OperationsManager) Start() error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if om.running {
		return fmt.Errorf("operations manager already running")
	}

	om.running = true

	// Start worker goroutines
	for i := 0; i < om.workers; i++ {
		go om.worker()
	}

	om.logger.Info("Operations manager started",
		"workers", om.workers,
		"max_concurrent", om.config.MaxConcurrentOps,
	)

	return nil
}

// Stop stops the operations manager
func (om *OperationsManager) Stop() error {
	om.mu.Lock()
	defer om.mu.Unlock()

	if !om.running {
		return nil
	}

	om.running = false
	close(om.stopCh)

	// Cancel all running operations
	for _, op := range om.operations {
		if op.Status == StatusRunning && op.cancel != nil {
			op.cancel()
		}
	}

	om.logger.Info("Operations manager stopped")
	return nil
}

// CreateOperation creates a new operation
func (om *OperationsManager) CreateOperation(
	opType OperationType,
	name, description string,
	parameters map[string]interface{},
	steps []OperationStep,
) (*Operation, error) {

	om.mu.Lock()
	defer om.mu.Unlock()

	// Check if executor exists
	if _, exists := om.executors[opType]; !exists {
		return nil, fmt.Errorf("no executor registered for operation type: %s", opType)
	}

	// Generate unique ID
	id := fmt.Sprintf("%s-%d", opType, time.Now().Unix())

	ctx, cancel := context.WithTimeout(context.Background(), om.config.DefaultTimeout)

	operation := &Operation{
		ID:          id,
		Type:        opType,
		Name:        name,
		Description: description,
		Status:      StatusPending,
		CreatedAt:   time.Now(),
		Steps:       steps,
		CurrentStep: -1,
		Progress:    0,
		Parameters:  parameters,
		Context:     make(map[string]string),
		Tags:        make([]string, 0),
		Result:      make(map[string]interface{}),
		Logs:        make([]string, 0),
		Errors:      make([]string, 0),
		Cancelable:  true,
		Timeout:     om.config.DefaultTimeout,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Validate operation
	if executor, exists := om.executors[opType]; exists {
		if err := executor.Validate(operation); err != nil {
			cancel()
			return nil, fmt.Errorf("operation validation failed: %w", err)
		}
	}

	om.operations[id] = operation

	om.logger.Info("Operation created",
		"id", id,
		"type", opType,
		"name", name,
	)

	return operation, nil
}

// ExecuteOperation executes an operation asynchronously
func (om *OperationsManager) ExecuteOperation(id string) error {
	om.mu.RLock()
	operation, exists := om.operations[id]
	om.mu.RUnlock()

	if !exists {
		return fmt.Errorf("operation not found: %s", id)
	}

	if operation.Status != StatusPending {
		return fmt.Errorf("operation %s is not in pending state: %s", id, operation.Status)
	}

	// Check if we can accept more operations
	select {
	case om.workerPool <- operation:
		om.logger.Info("Operation queued for execution", "id", id)
		return nil
	default:
		return fmt.Errorf("maximum concurrent operations reached")
	}
}

// CancelOperation cancels a running operation
func (om *OperationsManager) CancelOperation(id string) error {
	om.mu.RLock()
	operation, exists := om.operations[id]
	om.mu.RUnlock()

	if !exists {
		return fmt.Errorf("operation not found: %s", id)
	}

	if !operation.Cancelable {
		return fmt.Errorf("operation %s is not cancelable", id)
	}

	if operation.Status != StatusRunning {
		return fmt.Errorf("operation %s is not running: %s", id, operation.Status)
	}

	if operation.cancel != nil {
		operation.cancel()
		operation.Status = StatusCanceled
		now := time.Now()
		operation.CompletedAt = &now
		if operation.StartedAt != nil {
			operation.Duration = now.Sub(*operation.StartedAt)
		}

		om.addLog(operation, "Operation canceled by user")
		om.logger.Info("Operation canceled", "id", id)
	}

	return nil
}

// GetOperation returns an operation by ID
func (om *OperationsManager) GetOperation(id string) (*Operation, error) {
	om.mu.RLock()
	defer om.mu.RUnlock()

	operation, exists := om.operations[id]
	if !exists {
		return nil, fmt.Errorf("operation not found: %s", id)
	}

	// Return a copy to prevent external modifications
	opCopy := *operation
	return &opCopy, nil
}

// ListOperations returns all operations with optional filtering
func (om *OperationsManager) ListOperations(filter OperationFilter) []Operation {
	om.mu.RLock()
	defer om.mu.RUnlock()

	operations := make([]Operation, 0)

	for _, op := range om.operations {
		if filter.Matches(op) {
			opCopy := *op
			operations = append(operations, opCopy)
		}
	}

	return operations
}

// GetOperationHistory returns operation history
func (om *OperationsManager) GetOperationHistory(limit int) []Operation {
	om.mu.RLock()
	defer om.mu.RUnlock()

	if limit <= 0 || limit > len(om.history) {
		limit = len(om.history)
	}

	// Return most recent operations
	start := len(om.history) - limit
	history := make([]Operation, limit)
	copy(history, om.history[start:])

	return history
}

// OperationFilter for filtering operations
type OperationFilter struct {
	Type     *OperationType
	Status   *OperationStatus
	FromDate *time.Time
	ToDate   *time.Time
	Tags     []string
}

// Matches checks if an operation matches the filter
func (of *OperationFilter) Matches(op *Operation) bool {
	if of.Type != nil && op.Type != *of.Type {
		return false
	}

	if of.Status != nil && op.Status != *of.Status {
		return false
	}

	if of.FromDate != nil && op.CreatedAt.Before(*of.FromDate) {
		return false
	}

	if of.ToDate != nil && op.CreatedAt.After(*of.ToDate) {
		return false
	}

	if len(of.Tags) > 0 {
		tagMap := make(map[string]bool)
		for _, tag := range op.Tags {
			tagMap[tag] = true
		}

		for _, requiredTag := range of.Tags {
			if !tagMap[requiredTag] {
				return false
			}
		}
	}

	return true
}

// Private methods

func (om *OperationsManager) worker() {
	for {
		select {
		case <-om.stopCh:
			return
		case operation := <-om.workerPool:
			om.executeOperationWithRetry(operation)
		}
	}
}

func (om *OperationsManager) executeOperationWithRetry(operation *Operation) {
	executor, exists := om.executors[operation.Type]
	if !exists {
		om.failOperation(operation, fmt.Errorf("no executor found for operation type: %s", operation.Type))
		return
	}

	operation.Status = StatusRunning
	now := time.Now()
	operation.StartedAt = &now

	om.addLog(operation, fmt.Sprintf("Starting operation execution with executor: %T", executor))

	var lastErr error
	for attempt := 1; attempt <= om.config.OperationRetries; attempt++ {
		om.addLog(operation, fmt.Sprintf("Execution attempt %d/%d", attempt, om.config.OperationRetries))

		err := executor.Execute(operation.ctx, operation)
		if err == nil {
			// Success
			operation.Status = StatusCompleted
			now := time.Now()
			operation.CompletedAt = &now
			operation.Duration = now.Sub(*operation.StartedAt)
			operation.Progress = 100

			om.addLog(operation, fmt.Sprintf("Operation completed successfully in %v", operation.Duration))
			om.moveToHistory(operation)
			return
		}

		lastErr = err
		om.addError(operation, fmt.Sprintf("Attempt %d failed: %v", attempt, err))

		// Check if context was canceled
		select {
		case <-operation.ctx.Done():
			if operation.ctx.Err() == context.Canceled {
				operation.Status = StatusCanceled
			} else {
				operation.Status = StatusFailed
			}
			om.finalizeOperation(operation, operation.ctx.Err())
			return
		default:
		}

		// Wait before retry (except on last attempt)
		if attempt < om.config.OperationRetries {
			om.addLog(operation, fmt.Sprintf("Retrying in %v", om.config.RetryDelay))
			select {
			case <-operation.ctx.Done():
				om.finalizeOperation(operation, operation.ctx.Err())
				return
			case <-time.After(om.config.RetryDelay):
			}
		}
	}

	// All attempts failed
	om.failOperation(operation, fmt.Errorf("operation failed after %d attempts: %w", om.config.OperationRetries, lastErr))
}

func (om *OperationsManager) failOperation(operation *Operation, err error) {
	operation.Status = StatusFailed
	now := time.Now()
	operation.CompletedAt = &now
	if operation.StartedAt != nil {
		operation.Duration = now.Sub(*operation.StartedAt)
	}

	om.addError(operation, fmt.Sprintf("Operation failed: %v", err))
	om.logger.Error("Operation failed",
		"id", operation.ID,
		"type", operation.Type,
		"error", err,
	)

	om.moveToHistory(operation)
}

func (om *OperationsManager) finalizeOperation(operation *Operation, err error) {
	now := time.Now()
	operation.CompletedAt = &now
	if operation.StartedAt != nil {
		operation.Duration = now.Sub(*operation.StartedAt)
	}

	if err != nil {
		om.addError(operation, fmt.Sprintf("Operation finalized with error: %v", err))
	}

	om.moveToHistory(operation)
}

func (om *OperationsManager) moveToHistory(operation *Operation) {
	om.mu.Lock()
	defer om.mu.Unlock()

	// Add to history
	opCopy := *operation
	om.history = append(om.history, opCopy)

	// Maintain history size limit
	if len(om.history) > om.maxHistory {
		om.history = om.history[len(om.history)-om.maxHistory:]
	}

	// Remove from active operations
	delete(om.operations, operation.ID)

	// Cancel context to free resources
	if operation.cancel != nil {
		operation.cancel()
	}
}

func (om *OperationsManager) addLog(operation *Operation, message string) {
	logEntry := fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), message)
	operation.Logs = append(operation.Logs, logEntry)

	om.logger.Info(message,
		"operation_id", operation.ID,
		"operation_type", operation.Type,
	)
}

func (om *OperationsManager) addError(operation *Operation, message string) {
	errorEntry := fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), message)
	operation.Errors = append(operation.Errors, errorEntry)

	om.logger.Error(message,
		"operation_id", operation.ID,
		"operation_type", operation.Type,
	)
}

// Built-in operation executors

// MaintenanceExecutor handles maintenance operations
type MaintenanceExecutor struct {
	logger *logger.Logger
}

func NewMaintenanceExecutor(logger *logger.Logger) *MaintenanceExecutor {
	return &MaintenanceExecutor{logger: logger}
}

func (me *MaintenanceExecutor) Execute(ctx context.Context, operation *Operation) error {
	me.logger.Info("Executing maintenance operation", "id", operation.ID)

	// Simulate maintenance tasks
	for i, step := range operation.Steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		operation.CurrentStep = i
		operation.Progress = float64(i+1) / float64(len(operation.Steps)) * 100

		// Simulate step execution
		time.Sleep(time.Second)

		step.Status = StatusCompleted
		now := time.Now()
		step.CompletedAt = &now
		operation.Steps[i] = step
	}

	return nil
}

func (me *MaintenanceExecutor) Rollback(_ context.Context, operation *Operation) error {
	me.logger.Info("Rolling back maintenance operation", "id", operation.ID)
	return nil
}

func (me *MaintenanceExecutor) Validate(operation *Operation) error {
	if len(operation.Steps) == 0 {
		return fmt.Errorf("maintenance operation must have at least one step")
	}
	return nil
}
