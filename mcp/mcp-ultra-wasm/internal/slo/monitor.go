package slo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"go.uber.org/zap"
)

// Type represents the type of SLO being monitored
type Type string

const (
	TypeAvailability Type = "availability"
	TypeLatency      Type = "latency"
	TypeThroughput   Type = "throughput"
	TypeErrorRate    Type = "error_rate"
	TypeAccuracy     Type = "accuracy"
)

// Status represents the current status of an SLO
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusCritical  Status = "critical"
	StatusViolation Status = "violation"
)

// SLO represents a Service Level Objective
type SLO struct {
	// Basic identification
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        Type   `json:"type"`
	Service     string `json:"service"`
	Component   string `json:"component"`

	// SLO targets and thresholds
	Target            float64 `json:"target"`             // Primary SLO target (e.g., 99.9%)
	WarningThreshold  float64 `json:"warning_threshold"`  // Warning threshold (e.g., 99.5%)
	CriticalThreshold float64 `json:"critical_threshold"` // Critical threshold (e.g., 99.0%)

	// Queries and measurement
	Query            string `json:"query"`              // Prometheus query for measurement
	ErrorBudgetQuery string `json:"error_budget_query"` // Query for error budget calculation
	BurnRateQuery    string `json:"burn_rate_query"`    // Query for burn rate calculation

	// Time windows for evaluation
	EvaluationWindow time.Duration `json:"evaluation_window"` // Window for SLO evaluation (e.g., 5m)
	ComplianceWindow time.Duration `json:"compliance_window"` // Window for compliance measurement (e.g., 30d)

	// Alerting configuration
	AlertingRules []AlertRule `json:"alerting_rules"` // Associated alerting rules

	// Metadata
	Tags      map[string]string `json:"tags"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
	Enabled   bool              `json:"enabled"`
}

// Result represents the result of an SLO evaluation
type Result struct {
	SLO               *SLO              `json:"slo"`
	Status            Status            `json:"status"`
	CurrentValue      float64           `json:"current_value"`
	Target            float64           `json:"target"`
	ErrorBudget       ErrorBudget       `json:"error_budget"`
	BurnRate          BurnRate          `json:"burn_rate"`
	Timestamp         time.Time         `json:"timestamp"`
	ComplianceHistory []CompliancePoint `json:"compliance_history"`
}

// ErrorBudget represents the error budget information
type ErrorBudget struct {
	Total      float64   `json:"total"`      // Total error budget for the period
	Remaining  float64   `json:"remaining"`  // Remaining error budget
	Consumed   float64   `json:"consumed"`   // Consumed error budget
	Percentage float64   `json:"percentage"` // Percentage remaining
	ExhaustAt  time.Time `json:"exhaust_at"` // Estimated exhaustion time
}

// BurnRate represents burn rate information
type BurnRate struct {
	Current      float64 `json:"current"`       // Current burn rate
	Fast         float64 `json:"fast"`          // Fast burn rate (1h window)
	Slow         float64 `json:"slow"`          // Slow burn rate (6h window)
	Alerting     bool    `json:"alerting"`      // Whether burn rate is alerting
	FastAlerting bool    `json:"fast_alerting"` // Fast burn rate alerting
	SlowAlerting bool    `json:"slow_alerting"` // Slow burn rate alerting
}

// CompliancePoint represents a point in time compliance measurement
type CompliancePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Status    Status    `json:"status"`
}

// AlertRule represents an alerting rule for an SLO
type AlertRule struct {
	Name        string            `json:"name"`
	Expression  string            `json:"expression"`
	For         time.Duration     `json:"for"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Enabled     bool              `json:"enabled"`
}

// Monitor manages SLO monitoring and evaluation
type Monitor struct {
	logger     *zap.Logger
	promClient v1.API
	slos       map[string]*SLO
	results    map[string]*Result
	mu         sync.RWMutex

	// Configuration
	evaluationInterval time.Duration
	retentionPeriod    time.Duration

	// Channels for communication
	alertChan  chan AlertEvent
	statusChan chan StatusEvent
	stopChan   chan struct{}
}

// AlertEvent represents an SLO alert event
type AlertEvent struct {
	SLOName     string            `json:"slo_name"`
	Type        string            `json:"type"`
	Severity    string            `json:"severity"`
	Message     string            `json:"message"`
	Timestamp   time.Time         `json:"timestamp"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// StatusEvent represents an SLO status change event
type StatusEvent struct {
	SLOName        string    `json:"slo_name"`
	PreviousStatus Status    `json:"previous_status"`
	CurrentStatus  Status    `json:"current_status"`
	Timestamp      time.Time `json:"timestamp"`
	Reason         string    `json:"reason"`
}

// NewMonitor creates a new SLO monitor
func NewMonitor(promClient api.Client, logger *zap.Logger) (*Monitor, error) {
	v1api := v1.NewAPI(promClient)

	return &Monitor{
		logger:             logger,
		promClient:         v1api,
		slos:               make(map[string]*SLO),
		results:            make(map[string]*Result),
		evaluationInterval: 1 * time.Minute,
		retentionPeriod:    30 * 24 * time.Hour, // 30 days
		alertChan:          make(chan AlertEvent, 100),
		statusChan:         make(chan StatusEvent, 100),
		stopChan:           make(chan struct{}),
	}, nil
}

// AddSLO adds an SLO to the monitor
func (m *Monitor) AddSLO(slo *SLO) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if slo.Name == "" {
		return fmt.Errorf("SLO name cannot be empty")
	}

	if slo.Target <= 0 || slo.Target > 100 {
		return fmt.Errorf("SLO target must be between 0 and 100")
	}

	if slo.Query == "" {
		return fmt.Errorf("SLO query cannot be empty")
	}

	// Set default values
	if slo.EvaluationWindow == 0 {
		slo.EvaluationWindow = 5 * time.Minute
	}

	if slo.ComplianceWindow == 0 {
		slo.ComplianceWindow = 30 * 24 * time.Hour
	}

	if slo.CreatedAt.IsZero() {
		slo.CreatedAt = time.Now()
	}
	slo.UpdatedAt = time.Now()

	m.slos[slo.Name] = slo
	m.logger.Info("Added SLO", zap.String("name", slo.Name), zap.String("type", string(slo.Type)))

	return nil
}

// RemoveSLO removes an SLO from monitoring
func (m *Monitor) RemoveSLO(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.slos, name)
	delete(m.results, name)
	m.logger.Info("Removed SLO", zap.String("name", name))
}

// GetSLO retrieves an SLO by name
func (m *Monitor) GetSLO(name string) (*SLO, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	slo, exists := m.slos[name]
	return slo, exists
}

// GetAllSLOs returns all configured SLOs
func (m *Monitor) GetAllSLOs() map[string]*SLO {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*SLO, len(m.slos))
	for name, slo := range m.slos {
		result[name] = slo
	}
	return result
}

// GetResult retrieves the latest SLO evaluation result
func (m *Monitor) GetResult(name string) (*Result, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result, exists := m.results[name]
	return result, exists
}

// GetAllResults returns all SLO evaluation results
func (m *Monitor) GetAllResults() map[string]*Result {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*Result, len(m.results))
	for name, sloResult := range m.results {
		result[name] = sloResult
	}
	return result
}

// Start begins SLO monitoring
func (m *Monitor) Start(ctx context.Context) error {
	m.logger.Info("Starting SLO monitor")

	ticker := time.NewTicker(m.evaluationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("SLO monitor stopped by context")
			return ctx.Err()
		case <-m.stopChan:
			m.logger.Info("SLO monitor stopped")
			return nil
		case <-ticker.C:
			m.evaluateAllSLOs(ctx)
		}
	}
}

// Stop stops SLO monitoring
func (m *Monitor) Stop() {
	close(m.stopChan)
}

// AlertChannel returns the alert event channel
func (m *Monitor) AlertChannel() <-chan AlertEvent {
	return m.alertChan
}

// StatusChannel returns the status change event channel
func (m *Monitor) StatusChannel() <-chan StatusEvent {
	return m.statusChan
}

// evaluateAllSLOs evaluates all configured SLOs
func (m *Monitor) evaluateAllSLOs(ctx context.Context) {
	m.mu.RLock()
	slos := make([]*SLO, 0, len(m.slos))
	for _, slo := range m.slos {
		if slo.Enabled {
			slos = append(slos, slo)
		}
	}
	m.mu.RUnlock()

	for _, slo := range slos {
		if err := m.evaluateSLO(ctx, slo); err != nil {
			m.logger.Error("Failed to evaluate SLO",
				zap.String("slo", slo.Name),
				zap.Error(err))
		}
	}
}

// evaluateSLO evaluates a single SLO
func (m *Monitor) evaluateSLO(ctx context.Context, slo *SLO) error {
	// Query current value
	currentValue, err := m.queryPrometheus(ctx, slo.Query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to query current value: %w", err)
	}

	// Calculate error budget
	errorBudget, err := m.calculateErrorBudget(ctx, slo)
	if err != nil {
		return fmt.Errorf("failed to calculate error budget: %w", err)
	}

	// Calculate burn rate
	burnRate, err := m.calculateBurnRate(ctx, slo)
	if err != nil {
		return fmt.Errorf("failed to calculate burn rate: %w", err)
	}

	// Determine status
	status := m.determineStatus(slo, currentValue, errorBudget, burnRate)

	// Create result
	result := &Result{
		SLO:          slo,
		Status:       status,
		CurrentValue: currentValue,
		Target:       slo.Target,
		ErrorBudget:  errorBudget,
		BurnRate:     burnRate,
		Timestamp:    time.Now(),
	}

	// Get compliance history
	history, err := m.getComplianceHistory(ctx, slo)
	if err != nil {
		m.logger.Warn("Failed to get compliance history",
			zap.String("slo", slo.Name),
			zap.Error(err))
	} else {
		result.ComplianceHistory = history
	}

	// Store result and check for status changes
	m.storeResult(slo.Name, result)

	// Generate alerts if necessary
	m.checkAndGenerateAlerts(result)

	return nil
}

// queryPrometheus executes a Prometheus query
func (m *Monitor) queryPrometheus(ctx context.Context, query string, timestamp time.Time) (float64, error) {
	result, warnings, err := m.promClient.Query(ctx, query, timestamp)
	if err != nil {
		return 0, err
	}

	if len(warnings) > 0 {
		m.logger.Warn("Prometheus query warnings", zap.Strings("warnings", warnings))
	}

	switch result.Type() {
	case model.ValVector:
		vector, ok := result.(model.Vector)
		if !ok {
			return 0, fmt.Errorf("failed to cast result to Vector")
		}
		if len(vector) == 0 {
			return 0, fmt.Errorf("no data points returned")
		}
		return float64(vector[0].Value), nil
	case model.ValScalar:
		scalar, ok := result.(*model.Scalar)
		if !ok {
			return 0, fmt.Errorf("failed to cast result to Scalar")
		}
		return float64(scalar.Value), nil
	default:
		return 0, fmt.Errorf("unexpected result type: %s", result.Type())
	}
}

// calculateErrorBudget calculates the error budget for an SLO
func (m *Monitor) calculateErrorBudget(ctx context.Context, slo *SLO) (ErrorBudget, error) {
	var errorBudget ErrorBudget

	if slo.ErrorBudgetQuery == "" {
		// Calculate based on target and compliance window
		totalBudget := (100 - slo.Target) * float64(slo.ComplianceWindow.Minutes()) / 100
		errorBudget.Total = totalBudget

		// Query current error rate to calculate consumed budget
		errorQuery := fmt.Sprintf("(1 - (%s)) * 100", slo.Query)
		currentError, err := m.queryPrometheus(ctx, errorQuery, time.Now())
		if err != nil {
			return errorBudget, err
		}

		consumed := currentError * float64(slo.ComplianceWindow.Minutes()) / 100
		errorBudget.Consumed = consumed
		errorBudget.Remaining = totalBudget - consumed

		if totalBudget > 0 {
			errorBudget.Percentage = (errorBudget.Remaining / totalBudget) * 100
		}

		// Estimate exhaustion time
		if currentError > 0 {
			remainingMinutes := errorBudget.Remaining / (currentError / 100)
			errorBudget.ExhaustAt = time.Now().Add(time.Duration(remainingMinutes) * time.Minute)
		}
	} else {
		// Use custom error budget query
		value, err := m.queryPrometheus(ctx, slo.ErrorBudgetQuery, time.Now())
		if err != nil {
			return errorBudget, err
		}
		errorBudget.Remaining = value
		errorBudget.Percentage = value // Assuming query returns percentage
	}

	return errorBudget, nil
}

// calculateBurnRate calculates the burn rate for an SLO
func (m *Monitor) calculateBurnRate(ctx context.Context, slo *SLO) (BurnRate, error) {
	var burnRate BurnRate

	if slo.BurnRateQuery != "" {
		// Use custom burn rate query
		value, err := m.queryPrometheus(ctx, slo.BurnRateQuery, time.Now())
		if err != nil {
			return burnRate, err
		}
		burnRate.Current = value
	} else {
		// Calculate burn rate based on error rate
		errorQuery := fmt.Sprintf("rate((%s)[1h:])", slo.Query)
		errorRate, err := m.queryPrometheus(ctx, errorQuery, time.Now())
		if err != nil {
			return burnRate, err
		}

		allowedErrorRate := (100 - slo.Target) / 100
		if allowedErrorRate > 0 {
			burnRate.Current = errorRate / allowedErrorRate
		}
	}

	// Calculate fast burn rate (1h window)
	fastQuery := fmt.Sprintf("rate((%s)[1h:])", slo.Query)
	if fast, err := m.queryPrometheus(ctx, fastQuery, time.Now()); err == nil {
		burnRate.Fast = fast
		burnRate.FastAlerting = fast > 14.4 // 2% budget burn in 1 hour
	}

	// Calculate slow burn rate (6h window)
	slowQuery := fmt.Sprintf("rate((%s)[6h:])", slo.Query)
	if slow, err := m.queryPrometheus(ctx, slowQuery, time.Now()); err == nil {
		burnRate.Slow = slow
		burnRate.SlowAlerting = slow > 6 // 5% budget burn in 6 hours
	}

	burnRate.Alerting = burnRate.FastAlerting || burnRate.SlowAlerting

	return burnRate, nil
}

// determineStatus determines the SLO status based on current metrics
func (m *Monitor) determineStatus(slo *SLO, currentValue float64, errorBudget ErrorBudget, burnRate BurnRate) Status {
	// Check for violation (below critical threshold)
	if currentValue < slo.CriticalThreshold {
		return StatusViolation
	}

	// Check for critical status (burn rate alerting or low error budget)
	if burnRate.Alerting || errorBudget.Percentage < 10 {
		return StatusCritical
	}

	// Check for degraded status (below warning threshold)
	if currentValue < slo.WarningThreshold {
		return StatusDegraded
	}

	return StatusHealthy
}

// getComplianceHistory retrieves historical compliance data
func (m *Monitor) getComplianceHistory(ctx context.Context, slo *SLO) ([]CompliancePoint, error) {
	// Query historical data for the compliance window
	end := time.Now()
	start := end.Add(-slo.ComplianceWindow)
	step := slo.ComplianceWindow / 100 // 100 data points

	queryRange := v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	result, warnings, err := m.promClient.QueryRange(ctx, slo.Query, queryRange)
	if err != nil {
		return nil, err
	}

	if len(warnings) > 0 {
		m.logger.Warn("Compliance history query warnings", zap.Strings("warnings", warnings))
	}

	var history []CompliancePoint

	if matrix, ok := result.(model.Matrix); ok {
		for _, series := range matrix {
			for _, sample := range series.Values {
				value := float64(sample.Value)
				timestamp := sample.Timestamp.Time()

				status := StatusHealthy
				if value < slo.CriticalThreshold {
					status = StatusViolation
				} else if value < slo.WarningThreshold {
					status = StatusDegraded
				}

				history = append(history, CompliancePoint{
					Timestamp: timestamp,
					Value:     value,
					Status:    status,
				})
			}
		}
	}

	return history, nil
}

// storeResult stores an SLO evaluation result and checks for status changes
func (m *Monitor) storeResult(name string, result *Result) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for status change
	if previous, exists := m.results[name]; exists {
		if previous.Status != result.Status {
			// Status changed, emit event
			statusEvent := StatusEvent{
				SLOName:        name,
				PreviousStatus: previous.Status,
				CurrentStatus:  result.Status,
				Timestamp:      result.Timestamp,
				Reason:         fmt.Sprintf("Value changed from %.2f to %.2f", previous.CurrentValue, result.CurrentValue),
			}

			select {
			case m.statusChan <- statusEvent:
			default:
				m.logger.Warn("Status channel full, dropping event", zap.String("slo", name))
			}
		}
	}

	m.results[name] = result
}

// checkAndGenerateAlerts checks if alerts should be generated for an SLO result
func (m *Monitor) checkAndGenerateAlerts(result *Result) {
	for _, rule := range result.SLO.AlertingRules {
		if !rule.Enabled {
			continue
		}

		shouldAlert := false
		var alertMessage string

		switch rule.Name {
		case "SLOErrorBudgetLow":
			if result.ErrorBudget.Percentage < 10 {
				shouldAlert = true
				alertMessage = fmt.Sprintf("Error budget low: %.2f%% remaining", result.ErrorBudget.Percentage)
			}
		case "SLOBurnRateHigh":
			if result.BurnRate.Alerting {
				shouldAlert = true
				alertMessage = fmt.Sprintf("Burn rate high: %.2f", result.BurnRate.Current)
			}
		case "SLOViolation":
			if result.Status == StatusViolation {
				shouldAlert = true
				alertMessage = fmt.Sprintf("SLO violation: %.2f%% < %.2f%%", result.CurrentValue, result.SLO.CriticalThreshold)
			}
		}

		if shouldAlert {
			alertEvent := AlertEvent{
				SLOName:     result.SLO.Name,
				Type:        rule.Name,
				Severity:    rule.Severity,
				Message:     alertMessage,
				Timestamp:   result.Timestamp,
				Labels:      rule.Labels,
				Annotations: rule.Annotations,
			}

			select {
			case m.alertChan <- alertEvent:
			default:
				m.logger.Warn("Alert channel full, dropping alert",
					zap.String("slo", result.SLO.Name),
					zap.String("alert", rule.Name))
			}
		}
	}
}
