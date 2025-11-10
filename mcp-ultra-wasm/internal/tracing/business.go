package tracing

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/logger"
)

// BusinessTransactionTracer provides advanced tracing for critical business transactions
type BusinessTransactionTracer struct {
	tracer    trace.Tracer
	config    Config
	logger    *logger.Logger
	telemetry *observability.TelemetryService

	// State
	mu           sync.RWMutex
	transactions map[string]*BusinessTransaction
	templates    map[string]*TransactionTemplate
	correlations map[string][]string

	// Background processing
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// Config configures business transaction tracing
type Config struct {
	// General settings
	Enabled        bool   `yaml:"enabled"`
	ServiceName    string `yaml:"service_name"`
	ServiceVersion string `yaml:"service_version"`
	Environment    string `yaml:"environment"`

	// Sampling
	SamplingRate         float64 `yaml:"sampling_rate"`
	CriticalSamplingRate float64 `yaml:"critical_sampling_rate"`
	ErrorSamplingRate    float64 `yaml:"error_sampling_rate"`

	// Business transaction settings
	AutoInstrumentation  bool          `yaml:"auto_instrumentation"`
	TransactionThreshold time.Duration `yaml:"transaction_threshold"`
	MaxTransactionAge    time.Duration `yaml:"max_transaction_age"`

	// Correlation
	CorrelationEnabled bool     `yaml:"correlation_enabled"`
	CorrelationFields  []string `yaml:"correlation_fields"`
	MaxCorrelations    int      `yaml:"max_correlations"`

	// Storage
	RetentionPeriod time.Duration `yaml:"retention_period"`
	MaxTransactions int           `yaml:"max_transactions"`

	// Performance
	AsyncProcessing bool          `yaml:"async_processing"`
	BatchSize       int           `yaml:"batch_size"`
	FlushInterval   time.Duration `yaml:"flush_interval"`

	// Alerting
	AlertingEnabled bool            `yaml:"alerting_enabled"`
	AlertThresholds AlertThresholds `yaml:"alert_thresholds"`
}

// AlertThresholds defines alerting thresholds
type AlertThresholds struct {
	HighLatency       time.Duration `yaml:"high_latency"`
	VeryHighLatency   time.Duration `yaml:"very_high_latency"`
	ErrorRate         float64       `yaml:"error_rate"`
	FailureRate       float64       `yaml:"failure_rate"`
	TransactionVolume int64         `yaml:"transaction_volume"`
}

// BusinessTransaction represents a high-level business transaction
type BusinessTransaction struct {
	ID          string            `json:"id"`
	Type        TransactionType   `json:"type"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      TransactionStatus `json:"status"`

	// Timing
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`

	// Tracing
	TraceID      string `json:"trace_id"`
	SpanID       string `json:"span_id"`
	ParentSpanID string `json:"parent_span_id,omitempty"`

	// Business context
	UserID        string `json:"user_id,omitempty"`
	SessionID     string `json:"session_id,omitempty"`
	RequestID     string `json:"request_id,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`

	// Metadata
	Attributes map[string]interface{} `json:"attributes"`
	Tags       map[string]string      `json:"tags"`
	Metrics    TransactionMetrics     `json:"metrics"`

	// Steps and events
	Steps  []TransactionStep  `json:"steps"`
	Events []TransactionEvent `json:"events"`
	Errors []TransactionError `json:"errors"`

	// Classification
	Critical bool   `json:"critical"`
	Priority int    `json:"priority"`
	Category string `json:"category"`

	// Context
	Context context.Context `json:"-"`
	Span    trace.Span      `json:"-"`
}

// TransactionType represents different types of business transactions
type TransactionType string

const (
	TransactionTypeAPI        TransactionType = "api"
	TransactionTypeDatabase   TransactionType = "database"
	TransactionTypeMessage    TransactionType = "message"
	TransactionTypeFile       TransactionType = "file"
	TransactionTypeExternal   TransactionType = "external"
	TransactionTypeAuth       TransactionType = "auth"
	TransactionTypePayment    TransactionType = "payment"
	TransactionTypeCompliance TransactionType = "compliance"
	TransactionTypeWorkflow   TransactionType = "workflow"
	TransactionTypeBatch      TransactionType = "batch"
)

// TransactionStatus represents transaction status
type TransactionStatus string

const (
	TransactionStatusStarted    TransactionStatus = "started"
	TransactionStatusInProgress TransactionStatus = "in_progress"
	TransactionStatusCompleted  TransactionStatus = "completed"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusTimeout    TransactionStatus = "timeout"
	TransactionStatusCanceled   TransactionStatus = "canceled"
)

// TransactionStep represents a step within a business transaction
type TransactionStep struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Status     TransactionStatus      `json:"status"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	Duration   time.Duration          `json:"duration"`
	Attributes map[string]interface{} `json:"attributes"`
	Error      *TransactionError      `json:"error,omitempty"`

	// Tracing
	SpanID string `json:"span_id"`
}

// TransactionEvent represents an event within a transaction
type TransactionEvent struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Timestamp  time.Time              `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes"`
	Level      EventLevel             `json:"level"`
}

// TransactionError represents an error within a transaction
type TransactionError struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Message     string                 `json:"message"`
	Code        string                 `json:"code,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Attributes  map[string]interface{} `json:"attributes"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	Recoverable bool                   `json:"recoverable"`
}

// TransactionMetrics contains transaction performance metrics
type TransactionMetrics struct {
	DatabaseQueries int           `json:"database_queries"`
	DatabaseTime    time.Duration `json:"database_time"`
	ExternalCalls   int           `json:"external_calls"`
	ExternalTime    time.Duration `json:"external_time"`
	MemoryUsage     int64         `json:"memory_usage"`
	CPUTime         time.Duration `json:"cpu_time"`
	NetworkBytes    int64         `json:"network_bytes"`
	CacheHits       int           `json:"cache_hits"`
	CacheMisses     int           `json:"cache_misses"`
}

// TransactionTemplate defines a template for transaction creation
type TransactionTemplate struct {
	Type         TransactionType        `json:"type"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Critical     bool                   `json:"critical"`
	SamplingRate float64                `json:"sampling_rate"`
	Attributes   map[string]interface{} `json:"attributes"`
	Tags         map[string]string      `json:"tags"`

	// SLA settings
	SLAThreshold   time.Duration `json:"sla_threshold"`
	AlertThreshold time.Duration `json:"alert_threshold"`

	// Steps configuration
	ExpectedSteps []string `json:"expected_steps"`
	OptionalSteps []string `json:"optional_steps"`
}

// EventLevel represents the severity level of an event
type EventLevel string

const (
	EventLevelDebug    EventLevel = "debug"
	EventLevelInfo     EventLevel = "info"
	EventLevelWarn     EventLevel = "warn"
	EventLevelError    EventLevel = "error"
	EventLevelCritical EventLevel = "critical"
)

// DefaultConfig returns default tracing configuration
func DefaultConfig() Config {
	return Config{
		Enabled:              true,
		ServiceName:          "mcp-ultra-wasm",
		ServiceVersion:       "1.0.0",
		Environment:          "production",
		SamplingRate:         0.1, // 10% sampling
		CriticalSamplingRate: 1.0, // 100% for critical transactions
		ErrorSamplingRate:    1.0, // 100% for transactions with errors
		AutoInstrumentation:  true,
		TransactionThreshold: 100 * time.Millisecond,
		MaxTransactionAge:    1 * time.Hour,
		CorrelationEnabled:   true,
		CorrelationFields:    []string{"user_id", "session_id", "request_id"},
		MaxCorrelations:      1000,
		RetentionPeriod:      24 * time.Hour,
		MaxTransactions:      10000,
		AsyncProcessing:      true,
		BatchSize:            100,
		FlushInterval:        30 * time.Second,
		AlertingEnabled:      true,
		AlertThresholds: AlertThresholds{
			HighLatency:       1 * time.Second,
			VeryHighLatency:   5 * time.Second,
			ErrorRate:         0.05, // 5%
			FailureRate:       0.01, // 1%
			TransactionVolume: 1000,
		},
	}
}

// NewBusinessTransactionTracer creates a new business transaction tracer
func NewBusinessTransactionTracer(config Config, logger *logger.Logger, telemetry *observability.TelemetryService) (*BusinessTransactionTracer, error) {
	tracer := otel.Tracer(config.ServiceName)

	ctx, cancel := context.WithCancel(context.Background())

	btt := &BusinessTransactionTracer{
		tracer:       tracer,
		config:       config,
		logger:       logger,
		telemetry:    telemetry,
		transactions: make(map[string]*BusinessTransaction),
		templates:    make(map[string]*TransactionTemplate),
		correlations: make(map[string][]string),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Initialize default templates
	btt.initializeDefaultTemplates()

	// Start background processing
	if config.AsyncProcessing {
		btt.startBackgroundProcessing()
	}

	logger.Info("Business transaction tracer initialized",
		"service_name", config.ServiceName,
		"sampling_rate", config.SamplingRate,
		"auto_instrumentation", config.AutoInstrumentation,
		"correlation_enabled", config.CorrelationEnabled,
	)

	return btt, nil
}

// StartTransaction starts a new business transaction
func (btt *BusinessTransactionTracer) StartTransaction(ctx context.Context, transactionType TransactionType, name string, attributes map[string]interface{}) (*BusinessTransaction, context.Context) {
	template := btt.getTemplate(string(transactionType))

	// Determine sampling decision
	shouldSample := btt.shouldSample(template, attributes)
	if !shouldSample {
		// Return a lightweight transaction for non-sampled requests
		return btt.createLightweightTransaction(ctx, transactionType, name, attributes)
	}

	// Create span
	spanName := fmt.Sprintf("%s.%s", transactionType, name)
	spanCtx, span := btt.tracer.Start(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(btt.convertAttributes(attributes)...),
	)

	// Create transaction
	transaction := &BusinessTransaction{
		ID:         btt.generateTransactionID(),
		Type:       transactionType,
		Name:       name,
		Status:     TransactionStatusStarted,
		StartTime:  time.Now(),
		TraceID:    span.SpanContext().TraceID().String(),
		SpanID:     span.SpanContext().SpanID().String(),
		Attributes: attributes,
		Tags:       make(map[string]string),
		Steps:      make([]TransactionStep, 0),
		Events:     make([]TransactionEvent, 0),
		Errors:     make([]TransactionError, 0),
		Context:    spanCtx,
		Span:       span,
		Critical:   template != nil && template.Critical,
	}

	// Apply template settings
	if template != nil {
		transaction.Description = template.Description
		transaction.Priority = 1
		if template.Critical {
			transaction.Priority = 0 // Higher priority
		}

		// Merge template attributes
		for k, v := range template.Attributes {
			if _, exists := transaction.Attributes[k]; !exists {
				transaction.Attributes[k] = v
			}
		}

		// Merge template tags
		for k, v := range template.Tags {
			transaction.Tags[k] = v
		}
	}

	// Extract correlation fields
	btt.extractCorrelationFields(transaction, attributes)

	// Store transaction
	btt.mu.Lock()
	btt.transactions[transaction.ID] = transaction
	btt.mu.Unlock()

	// Add to baggage
	btt.addToBaggage(spanCtx, transaction)

	// Record metrics
	btt.recordTransactionStart(transaction)

	btt.logger.Debug("Transaction started",
		"transaction_id", transaction.ID,
		"type", transactionType,
		"name", name,
		"trace_id", transaction.TraceID,
		"critical", transaction.Critical,
	)

	return transaction, spanCtx
}

// EndTransaction ends a business transaction
func (btt *BusinessTransactionTracer) EndTransaction(transaction *BusinessTransaction, err error) {
	if transaction == nil {
		return
	}

	// Set end time and status
	transaction.EndTime = time.Now()
	transaction.Duration = transaction.EndTime.Sub(transaction.StartTime)

	if err != nil {
		transaction.Status = TransactionStatusFailed
		btt.addError(transaction, "transaction_error", err.Error(), err, true)

		// Set span error status
		if transaction.Span != nil {
			transaction.Span.SetStatus(codes.Error, err.Error())
		}
	} else {
		transaction.Status = TransactionStatusCompleted

		// Set span OK status
		if transaction.Span != nil {
			transaction.Span.SetStatus(codes.Ok, "Transaction completed successfully")
		}
	}

	// Add final attributes
	if transaction.Span != nil {
		transaction.Span.SetAttributes(
			attribute.Int64("transaction.duration_ms", transaction.Duration.Milliseconds()),
			attribute.String("transaction.status", string(transaction.Status)),
			attribute.Int("transaction.steps_count", len(transaction.Steps)),
			attribute.Int("transaction.events_count", len(transaction.Events)),
			attribute.Int("transaction.errors_count", len(transaction.Errors)),
		)

		// End span
		transaction.Span.End()
	}

	// Record metrics
	btt.recordTransactionEnd(transaction)

	// Check for alerts
	if btt.config.AlertingEnabled {
		btt.checkAlerts(transaction)
	}

	// Update correlations
	btt.updateCorrelations(transaction)

	btt.logger.Debug("Transaction ended",
		"transaction_id", transaction.ID,
		"status", transaction.Status,
		"duration", transaction.Duration,
		"steps", len(transaction.Steps),
		"errors", len(transaction.Errors),
	)

	// Schedule for cleanup
	if btt.config.AsyncProcessing {
		btt.scheduleCleanup(transaction.ID)
	}
}

// StartStep starts a new step within a transaction
func (btt *BusinessTransactionTracer) StartStep(transaction *BusinessTransaction, stepName, stepType string, attributes map[string]interface{}) *TransactionStep {
	if transaction == nil {
		return nil
	}

	// Create child span
	spanName := fmt.Sprintf("%s.step.%s", transaction.Name, stepName)
	stepCtx, stepSpan := btt.tracer.Start(transaction.Context, spanName,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(btt.convertAttributes(attributes)...),
	)

	step := &TransactionStep{
		ID:         btt.generateStepID(transaction.ID),
		Name:       stepName,
		Type:       stepType,
		Status:     TransactionStatusStarted,
		StartTime:  time.Now(),
		Attributes: attributes,
		SpanID:     stepSpan.SpanContext().SpanID().String(),
	}

	// Store step context for later use
	transaction.Context = stepCtx

	// Add to transaction
	btt.mu.Lock()
	transaction.Steps = append(transaction.Steps, *step)
	btt.mu.Unlock()

	return step
}

// EndStep ends a transaction step
func (btt *BusinessTransactionTracer) EndStep(transaction *BusinessTransaction, step *TransactionStep, err error) {
	if transaction == nil || step == nil {
		return
	}

	step.EndTime = time.Now()
	step.Duration = step.EndTime.Sub(step.StartTime)

	if err != nil {
		step.Status = TransactionStatusFailed
		step.Error = &TransactionError{
			ID:        btt.generateErrorID(),
			Type:      "step_error",
			Message:   err.Error(),
			Timestamp: time.Now(),
		}
	} else {
		step.Status = TransactionStatusCompleted
	}

	// Find and update the step in transaction
	btt.mu.Lock()
	for i, s := range transaction.Steps {
		if s.ID == step.ID {
			transaction.Steps[i] = *step
			break
		}
	}
	btt.mu.Unlock()

	// Find and end the corresponding span
	// This is simplified - in reality you'd maintain span references
	span := trace.SpanFromContext(transaction.Context)
	if span != nil {
		if err != nil {
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "Step completed successfully")
		}
		span.End()
	}
}

// AddEvent adds an event to a transaction
func (btt *BusinessTransactionTracer) AddEvent(transaction *BusinessTransaction, eventType, eventName string, level EventLevel, attributes map[string]interface{}) {
	if transaction == nil {
		return
	}

	event := TransactionEvent{
		ID:         btt.generateEventID(),
		Type:       eventType,
		Name:       eventName,
		Timestamp:  time.Now(),
		Attributes: attributes,
		Level:      level,
	}

	btt.mu.Lock()
	transaction.Events = append(transaction.Events, event)
	btt.mu.Unlock()

	// Add to span
	if transaction.Span != nil {
		transaction.Span.AddEvent(eventName, trace.WithAttributes(btt.convertAttributes(attributes)...))
	}

	// Log high-level events
	if level == EventLevelError || level == EventLevelCritical {
		btt.logger.Error("Transaction event",
			"transaction_id", transaction.ID,
			"event_type", eventType,
			"event_name", eventName,
			"level", level,
		)
	}
}

// AddError adds an error to a transaction
func (btt *BusinessTransactionTracer) AddError(transaction *BusinessTransaction, errorType, message string, err error) {
	btt.addError(transaction, errorType, message, err, true)
}

// GetTransaction retrieves a transaction by ID
func (btt *BusinessTransactionTracer) GetTransaction(transactionID string) *BusinessTransaction {
	btt.mu.RLock()
	defer btt.mu.RUnlock()

	transaction, exists := btt.transactions[transactionID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modifications
	transactionCopy := *transaction
	return &transactionCopy
}

// ListActiveTransactions returns all currently active transactions
func (btt *BusinessTransactionTracer) ListActiveTransactions() []*BusinessTransaction {
	btt.mu.RLock()
	defer btt.mu.RUnlock()

	transactions := make([]*BusinessTransaction, 0, len(btt.transactions))
	for _, transaction := range btt.transactions {
		if transaction.Status == TransactionStatusStarted || transaction.Status == TransactionStatusInProgress {
			transactionCopy := *transaction
			transactions = append(transactions, &transactionCopy)
		}
	}

	return transactions
}

// GetTransactionMetrics returns aggregated metrics for transactions
func (btt *BusinessTransactionTracer) GetTransactionMetrics() TransactionAnalytics {
	btt.mu.RLock()
	defer btt.mu.RUnlock()

	analytics := TransactionAnalytics{
		TotalTransactions: int64(len(btt.transactions)),
		ByType:            make(map[string]int64),
		ByStatus:          make(map[string]int64),
		AvgDuration:       0,
		MaxDuration:       0,
		MinDuration:       0,
	}

	var totalDuration time.Duration
	first := true

	for _, transaction := range btt.transactions {
		analytics.ByType[string(transaction.Type)]++
		analytics.ByStatus[string(transaction.Status)]++

		if transaction.Duration > 0 {
			totalDuration += transaction.Duration

			if first {
				analytics.MinDuration = transaction.Duration
				analytics.MaxDuration = transaction.Duration
				first = false
			} else {
				if transaction.Duration < analytics.MinDuration {
					analytics.MinDuration = transaction.Duration
				}
				if transaction.Duration > analytics.MaxDuration {
					analytics.MaxDuration = transaction.Duration
				}
			}
		}
	}

	if analytics.TotalTransactions > 0 {
		analytics.AvgDuration = totalDuration / time.Duration(analytics.TotalTransactions)
	}

	return analytics
}

// RegisterTemplate registers a transaction template
func (btt *BusinessTransactionTracer) RegisterTemplate(template *TransactionTemplate) {
	btt.mu.Lock()
	defer btt.mu.Unlock()

	btt.templates[template.Name] = template

	btt.logger.Info("Transaction template registered",
		"template_name", template.Name,
		"type", template.Type,
		"critical", template.Critical,
	)
}

// Close gracefully shuts down the tracer
func (btt *BusinessTransactionTracer) Close() error {
	btt.logger.Info("Shutting down business transaction tracer")

	btt.cancel()
	btt.wg.Wait()

	return nil
}

// TransactionAnalytics contains transaction analytics
type TransactionAnalytics struct {
	TotalTransactions int64            `json:"total_transactions"`
	ByType            map[string]int64 `json:"by_type"`
	ByStatus          map[string]int64 `json:"by_status"`
	AvgDuration       time.Duration    `json:"avg_duration"`
	MaxDuration       time.Duration    `json:"max_duration"`
	MinDuration       time.Duration    `json:"min_duration"`
}

// Private methods

func (btt *BusinessTransactionTracer) initializeDefaultTemplates() {
	templates := []*TransactionTemplate{
		{
			Type:           TransactionTypeAPI,
			Name:           "api_request",
			Description:    "HTTP API request processing",
			Critical:       false,
			SamplingRate:   btt.config.SamplingRate,
			SLAThreshold:   500 * time.Millisecond,
			AlertThreshold: 2 * time.Second,
			ExpectedSteps:  []string{"validation", "processing", "response"},
		},
		{
			Type:           TransactionTypeAuth,
			Name:           "authentication",
			Description:    "User authentication flow",
			Critical:       true,
			SamplingRate:   btt.config.CriticalSamplingRate,
			SLAThreshold:   200 * time.Millisecond,
			AlertThreshold: 1 * time.Second,
			ExpectedSteps:  []string{"credential_validation", "token_generation"},
		},
		{
			Type:           TransactionTypePayment,
			Name:           "payment_processing",
			Description:    "Payment processing transaction",
			Critical:       true,
			SamplingRate:   1.0, // Always sample payments
			SLAThreshold:   2 * time.Second,
			AlertThreshold: 5 * time.Second,
			ExpectedSteps:  []string{"validation", "authorization", "capture", "confirmation"},
		},
		{
			Type:           TransactionTypeCompliance,
			Name:           "compliance_check",
			Description:    "LGPD/GDPR compliance processing",
			Critical:       true,
			SamplingRate:   1.0, // Always sample compliance
			SLAThreshold:   1 * time.Second,
			AlertThreshold: 3 * time.Second,
			ExpectedSteps:  []string{"data_classification", "policy_evaluation", "audit_logging"},
		},
	}

	for _, template := range templates {
		btt.templates[template.Name] = template
	}
}

func (btt *BusinessTransactionTracer) shouldSample(template *TransactionTemplate, _ map[string]interface{}) bool {
	if template != nil && template.Critical {
		return true // Always sample critical transactions
	}

	samplingRate := btt.config.SamplingRate
	if template != nil && template.SamplingRate > 0 {
		samplingRate = template.SamplingRate
	}

	// Simple random sampling
	return btt.generateRandomFloat() < samplingRate
}

func (btt *BusinessTransactionTracer) createLightweightTransaction(ctx context.Context, transactionType TransactionType, name string, attributes map[string]interface{}) (*BusinessTransaction, context.Context) {
	// Create minimal transaction for non-sampled requests
	transaction := &BusinessTransaction{
		ID:         btt.generateTransactionID(),
		Type:       transactionType,
		Name:       name,
		Status:     TransactionStatusStarted,
		StartTime:  time.Now(),
		Attributes: attributes,
		Tags:       make(map[string]string),
		Context:    ctx,
	}

	return transaction, ctx
}

func (btt *BusinessTransactionTracer) getTemplate(templateName string) *TransactionTemplate {
	btt.mu.RLock()
	defer btt.mu.RUnlock()

	return btt.templates[templateName]
}

func (btt *BusinessTransactionTracer) extractCorrelationFields(transaction *BusinessTransaction, attributes map[string]interface{}) {
	if !btt.config.CorrelationEnabled {
		return
	}

	for _, field := range btt.config.CorrelationFields {
		if value, exists := attributes[field]; exists {
			switch field {
			case "user_id":
				if userID, ok := value.(string); ok {
					transaction.UserID = userID
				}
			case "session_id":
				if sessionID, ok := value.(string); ok {
					transaction.SessionID = sessionID
				}
			case "request_id":
				if requestID, ok := value.(string); ok {
					transaction.RequestID = requestID
				}
			case "correlation_id":
				if correlationID, ok := value.(string); ok {
					transaction.CorrelationID = correlationID
				}
			}
		}
	}
}

func (btt *BusinessTransactionTracer) addToBaggage(ctx context.Context, transaction *BusinessTransaction) context.Context {
	bag, _ := baggage.Parse(fmt.Sprintf("transaction.id=%s,transaction.type=%s,transaction.name=%s",
		transaction.ID, string(transaction.Type), transaction.Name))

	if transaction.UserID != "" {
		member, _ := baggage.NewMember("user.id", transaction.UserID)
		bag, _ = bag.SetMember(member)
	}

	return baggage.ContextWithBaggage(ctx, bag)
}

func (btt *BusinessTransactionTracer) addError(transaction *BusinessTransaction, errorType, message string, err error, recoverable bool) {
	if transaction == nil {
		return
	}

	transactionError := TransactionError{
		ID:          btt.generateErrorID(),
		Type:        errorType,
		Message:     message,
		Timestamp:   time.Now(),
		Attributes:  make(map[string]interface{}),
		Recoverable: recoverable,
	}

	if err != nil {
		transactionError.Code = fmt.Sprintf("%T", err)
		// In a real implementation, you'd extract stack trace here
	}

	btt.mu.Lock()
	transaction.Errors = append(transaction.Errors, transactionError)
	btt.mu.Unlock()

	// Record in span
	if transaction.Span != nil {
		transaction.Span.AddEvent("error",
			trace.WithAttributes(
				attribute.String("error.type", errorType),
				attribute.String("error.message", message),
				attribute.Bool("error.recoverable", recoverable),
			),
		)
	}
}

func (btt *BusinessTransactionTracer) recordTransactionStart(transaction *BusinessTransaction) {
	if btt.telemetry != nil {
		btt.telemetry.RecordCounter("business_transactions_started_total", 1, map[string]string{
			"type":     string(transaction.Type),
			"name":     transaction.Name,
			"critical": fmt.Sprintf("%t", transaction.Critical),
		})
	}
}

func (btt *BusinessTransactionTracer) recordTransactionEnd(transaction *BusinessTransaction) {
	if btt.telemetry != nil {
		btt.telemetry.RecordCounter("business_transactions_completed_total", 1, map[string]string{
			"type":   string(transaction.Type),
			"name":   transaction.Name,
			"status": string(transaction.Status),
		})

		btt.telemetry.RecordHistogram("business_transaction_duration", float64(transaction.Duration.Milliseconds()), map[string]string{
			"type": string(transaction.Type),
			"name": transaction.Name,
		})

		if len(transaction.Errors) > 0 {
			btt.telemetry.RecordCounter("business_transaction_errors_total", float64(len(transaction.Errors)), map[string]string{
				"type": string(transaction.Type),
				"name": transaction.Name,
			})
		}
	}
}

func (btt *BusinessTransactionTracer) checkAlerts(transaction *BusinessTransaction) {
	// Check duration alerts
	if transaction.Duration > btt.config.AlertThresholds.VeryHighLatency {
		btt.logger.Error("Very high transaction latency detected",
			"transaction_id", transaction.ID,
			"type", transaction.Type,
			"duration", transaction.Duration,
			"threshold", btt.config.AlertThresholds.VeryHighLatency,
		)
	} else if transaction.Duration > btt.config.AlertThresholds.HighLatency {
		btt.logger.Warn("High transaction latency detected",
			"transaction_id", transaction.ID,
			"type", transaction.Type,
			"duration", transaction.Duration,
			"threshold", btt.config.AlertThresholds.HighLatency,
		)
	}

	// Check error alerts
	if len(transaction.Errors) > 0 {
		btt.logger.Error("Transaction completed with errors",
			"transaction_id", transaction.ID,
			"type", transaction.Type,
			"error_count", len(transaction.Errors),
		)
	}
}

func (btt *BusinessTransactionTracer) updateCorrelations(transaction *BusinessTransaction) {
	if !btt.config.CorrelationEnabled || transaction.CorrelationID == "" {
		return
	}

	btt.mu.Lock()
	defer btt.mu.Unlock()

	correlations := btt.correlations[transaction.CorrelationID]
	correlations = append(correlations, transaction.ID)

	// Limit correlation size
	if len(correlations) > btt.config.MaxCorrelations {
		correlations = correlations[len(correlations)-btt.config.MaxCorrelations:]
	}

	btt.correlations[transaction.CorrelationID] = correlations
}

func (btt *BusinessTransactionTracer) convertAttributes(attributes map[string]interface{}) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, len(attributes))

	for k, v := range attributes {
		switch val := v.(type) {
		case string:
			attrs = append(attrs, attribute.String(k, val))
		case int:
			attrs = append(attrs, attribute.Int(k, val))
		case int64:
			attrs = append(attrs, attribute.Int64(k, val))
		case float64:
			attrs = append(attrs, attribute.Float64(k, val))
		case bool:
			attrs = append(attrs, attribute.Bool(k, val))
		default:
			// Convert to string for complex types
			if jsonBytes, err := json.Marshal(v); err == nil {
				attrs = append(attrs, attribute.String(k, string(jsonBytes)))
			}
		}
	}

	return attrs
}

func (btt *BusinessTransactionTracer) generateTransactionID() string {
	return fmt.Sprintf("txn_%d", time.Now().UnixNano())
}

func (btt *BusinessTransactionTracer) generateStepID(transactionID string) string {
	return fmt.Sprintf("%s_step_%d", transactionID, time.Now().UnixNano())
}

func (btt *BusinessTransactionTracer) generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}

func (btt *BusinessTransactionTracer) generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}

func (btt *BusinessTransactionTracer) generateRandomFloat() float64 {
	// Simple random number generation - in production use crypto/rand
	return float64(time.Now().UnixNano()%1000) / 1000.0
}

func (btt *BusinessTransactionTracer) startBackgroundProcessing() {
	// Cleanup task
	btt.wg.Add(1)
	go btt.cleanupTask()

	// Analytics task
	btt.wg.Add(1)
	go btt.analyticsTask()
}

func (btt *BusinessTransactionTracer) cleanupTask() {
	defer btt.wg.Done()

	ticker := time.NewTicker(btt.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-btt.ctx.Done():
			return
		case <-ticker.C:
			btt.performCleanup()
		}
	}
}

func (btt *BusinessTransactionTracer) analyticsTask() {
	defer btt.wg.Done()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-btt.ctx.Done():
			return
		case <-ticker.C:
			btt.computeAnalytics()
		}
	}
}

func (btt *BusinessTransactionTracer) performCleanup() {
	btt.mu.Lock()
	defer btt.mu.Unlock()

	cutoff := time.Now().Add(-btt.config.MaxTransactionAge)

	// Remove old transactions
	for id, transaction := range btt.transactions {
		if transaction.EndTime.Before(cutoff) {
			delete(btt.transactions, id)
		}
	}

	// Limit transaction count
	if len(btt.transactions) > btt.config.MaxTransactions {
		// Remove oldest transactions (simplified)
		count := len(btt.transactions) - btt.config.MaxTransactions
		for id := range btt.transactions {
			if count <= 0 {
				break
			}
			delete(btt.transactions, id)
			count--
		}
	}
}

func (btt *BusinessTransactionTracer) scheduleCleanup(transactionID string) {
	// Schedule individual transaction cleanup (simplified)
	go func() {
		time.Sleep(btt.config.MaxTransactionAge)
		btt.mu.Lock()
		delete(btt.transactions, transactionID)
		btt.mu.Unlock()
	}()
}

func (btt *BusinessTransactionTracer) computeAnalytics() {
	analytics := btt.GetTransactionMetrics()

	// Record analytics metrics
	if btt.telemetry != nil {
		btt.telemetry.RecordGauge("business_transactions_total", float64(analytics.TotalTransactions), nil)
		btt.telemetry.RecordGauge("business_transactions_avg_duration_ms", float64(analytics.AvgDuration.Milliseconds()), nil)
		btt.telemetry.RecordGauge("business_transactions_max_duration_ms", float64(analytics.MaxDuration.Milliseconds()), nil)

		for transactionType, count := range analytics.ByType {
			btt.telemetry.RecordGauge("business_transactions_by_type", float64(count), map[string]string{
				"type": transactionType,
			})
		}

		for status, count := range analytics.ByStatus {
			btt.telemetry.RecordGauge("business_transactions_by_status", float64(count), map[string]string{
				"status": status,
			})
		}
	}
}
