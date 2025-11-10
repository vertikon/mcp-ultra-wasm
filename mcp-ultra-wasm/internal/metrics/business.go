package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/observability"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/logger"
)

// MetricType represents different types of business metrics
type MetricType string

const (
	MetricCounter   MetricType = "counter"
	MetricGauge     MetricType = "gauge"
	MetricHistogram MetricType = "histogram"
	MetricSummary   MetricType = "summary"
)

// AggregationType represents how metrics should be aggregated
type AggregationType string

const (
	AggregationSum AggregationType = "sum"
	AggregationAvg AggregationType = "avg"
	AggregationMax AggregationType = "max"
	AggregationMin AggregationType = "min"
)

// AlertStateType represents the state of an alert
type AlertStateType string

const (
	AlertStateFiring   AlertStateType  = "firing"
	AlertStatePending  AlertStateType  = "pending"
	AlertStateResolved AlertStateType  = "resolved"
	AggregationCount   AggregationType = "count"
	AggregationP95     AggregationType = "p95"
	AggregationP99     AggregationType = "p99"
)

// BusinessMetric defines a business metric configuration
type BusinessMetric struct {
	Name         string              `yaml:"name" json:"name"`
	Type         MetricType          `yaml:"type" json:"type"`
	Description  string              `yaml:"description" json:"description"`
	Unit         string              `yaml:"unit" json:"unit"`
	Labels       []string            `yaml:"labels" json:"labels"`
	Aggregations []AggregationType   `yaml:"aggregations" json:"aggregations"`
	Buckets      []float64           `yaml:"buckets,omitempty" json:"buckets,omitempty"`       // For histograms
	Objectives   map[float64]float64 `yaml:"objectives,omitempty" json:"objectives,omitempty"` // For summaries
	TTL          time.Duration       `yaml:"ttl" json:"ttl"`
	Category     string              `yaml:"category" json:"category"`
	Priority     int                 `yaml:"priority" json:"priority"`
	Enabled      bool                `yaml:"enabled" json:"enabled"`
}

// BusinessMetricsConfig configures business metrics collection
type BusinessMetricsConfig struct {
	Enabled            bool              `yaml:"enabled"`
	CollectionInterval time.Duration     `yaml:"collection_interval"`
	RetentionPeriod    time.Duration     `yaml:"retention_period"`
	CustomMetrics      []BusinessMetric  `yaml:"custom_metrics"`
	DefaultLabels      map[string]string `yaml:"default_labels"`

	// Storage configuration
	StorageBackend string                 `yaml:"storage_backend"` // "memory", "redis", "postgres"
	StorageConfig  map[string]interface{} `yaml:"storage_config"`

	// Alerting configuration
	AlertingEnabled bool              `yaml:"alerting_enabled"`
	AlertRules      []MetricAlertRule `yaml:"alert_rules"`

	// Export configuration
	ExportEnabled  bool          `yaml:"export_enabled"`
	ExportInterval time.Duration `yaml:"export_interval"`
	ExportFormat   string        `yaml:"export_format"` // "prometheus", "json", "csv"
	ExportEndpoint string        `yaml:"export_endpoint"`
}

// MetricAlertRule defines alerting rules for business metrics
type MetricAlertRule struct {
	MetricName  string            `yaml:"metric_name"`
	Condition   string            `yaml:"condition"` // ">", "<", ">=", "<=", "==", "!="
	Threshold   float64           `yaml:"threshold"`
	Duration    time.Duration     `yaml:"duration"` // How long condition must be true
	Severity    string            `yaml:"severity"` // "critical", "warning", "info"
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
	Enabled     bool              `yaml:"enabled"`
}

// MetricValue represents a metric measurement
type MetricValue struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
	Unit      string            `json:"unit"`
}

// AggregatedMetric represents an aggregated metric value
type AggregatedMetric struct {
	MetricValue
	Aggregation AggregationType `json:"aggregation"`
	Period      time.Duration   `json:"period"`
	Count       int64           `json:"count"`
}

// BusinessMetricsCollector collects and manages business metrics
type BusinessMetricsCollector struct {
	config    BusinessMetricsConfig
	logger    *logger.Logger
	telemetry *observability.TelemetryService

	// State
	mu           sync.RWMutex
	metrics      map[string]*BusinessMetric
	values       map[string][]MetricValue
	aggregations map[string]map[AggregationType]AggregatedMetric
	alertStates  map[string]AlertState

	// Background processing
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Storage
	storage MetricStorage
}

// AlertState tracks the state of metric alerts
type AlertState struct {
	MetricName  string            `json:"metric_name"`
	RuleName    string            `json:"rule_name"`
	State       string            `json:"state"` // "firing", "pending", "resolved"
	Value       float64           `json:"value"`
	Threshold   float64           `json:"threshold"`
	Since       time.Time         `json:"since"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

// MetricStorage interface for metric storage backends
type MetricStorage interface {
	Store(ctx context.Context, values []MetricValue) error
	Query(ctx context.Context, query MetricQuery) ([]MetricValue, error)
	Aggregate(ctx context.Context, query AggregationQuery) ([]AggregatedMetric, error)
	Delete(ctx context.Context, before time.Time) error
	Close() error
}

// MetricQuery defines a metric query
type MetricQuery struct {
	MetricName string            `json:"metric_name"`
	Labels     map[string]string `json:"labels"`
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Limit      int               `json:"limit"`
}

// AggregationQuery defines an aggregation query
type AggregationQuery struct {
	MetricQuery
	Aggregations []AggregationType `json:"aggregations"`
	GroupBy      []string          `json:"group_by"`
	Period       time.Duration     `json:"period"`
}

// DefaultBusinessMetricsConfig returns default configuration
func DefaultBusinessMetricsConfig() BusinessMetricsConfig {
	return BusinessMetricsConfig{
		Enabled:            true,
		CollectionInterval: 30 * time.Second,
		RetentionPeriod:    7 * 24 * time.Hour, // 7 days
		DefaultLabels: map[string]string{
			"service":     "mcp-ultra-wasm",
			"environment": "production",
		},
		StorageBackend:  "memory",
		StorageConfig:   make(map[string]interface{}),
		AlertingEnabled: true,
		ExportEnabled:   true,
		ExportInterval:  5 * time.Minute,
		ExportFormat:    "prometheus",
		CustomMetrics:   DefaultBusinessMetrics(),
		AlertRules:      DefaultAlertRules(),
	}
}

// DefaultBusinessMetrics returns default business metrics
func DefaultBusinessMetrics() []BusinessMetric {
	return []BusinessMetric{
		// API Performance Metrics
		{
			Name:        "api_request_duration",
			Type:        MetricHistogram,
			Description: "Duration of API requests",
			Unit:        "seconds",
			Labels:      []string{"method", "endpoint", "status_code"},
			Buckets:     []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			Category:    "performance",
			Priority:    1,
			Enabled:     true,
		},
		{
			Name:        "api_requests_total",
			Type:        MetricCounter,
			Description: "Total number of API requests",
			Unit:        "count",
			Labels:      []string{"method", "endpoint", "status_code"},
			Category:    "performance",
			Priority:    1,
			Enabled:     true,
		},

		// Business KPIs
		{
			Name:        "active_users",
			Type:        MetricGauge,
			Description: "Number of currently active users",
			Unit:        "count",
			Labels:      []string{"user_type", "subscription_tier"},
			Category:    "business",
			Priority:    1,
			Enabled:     true,
		},
		{
			Name:        "tasks_completed",
			Type:        MetricCounter,
			Description: "Number of completed tasks",
			Unit:        "count",
			Labels:      []string{"task_type", "priority", "user_id"},
			Category:    "business",
			Priority:    2,
			Enabled:     true,
		},
		{
			Name:        "revenue_generated",
			Type:        MetricCounter,
			Description: "Revenue generated from completed transactions",
			Unit:        "currency",
			Labels:      []string{"subscription_tier", "payment_method"},
			Category:    "business",
			Priority:    1,
			Enabled:     true,
		},

		// System Health Metrics
		{
			Name:        "database_query_duration",
			Type:        MetricHistogram,
			Description: "Duration of database queries",
			Unit:        "seconds",
			Labels:      []string{"query_type", "table", "operation"},
			Buckets:     []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5},
			Category:    "system",
			Priority:    2,
			Enabled:     true,
		},
		{
			Name:        "cache_operations",
			Type:        MetricCounter,
			Description: "Number of cache operations",
			Unit:        "count",
			Labels:      []string{"operation", "result", "cache_type"},
			Category:    "system",
			Priority:    2,
			Enabled:     true,
		},

		// Compliance Metrics
		{
			Name:        "compliance_events",
			Type:        MetricCounter,
			Description: "Number of compliance-related events",
			Unit:        "count",
			Labels:      []string{"event_type", "regulation", "status"},
			Category:    "compliance",
			Priority:    1,
			Enabled:     true,
		},
		{
			Name:        "data_processing_requests",
			Type:        MetricCounter,
			Description: "Number of data processing requests (GDPR/LGPD)",
			Unit:        "count",
			Labels:      []string{"request_type", "regulation", "status"},
			Category:    "compliance",
			Priority:    1,
			Enabled:     true,
		},

		// Feature Usage Metrics
		{
			Name:        "feature_usage",
			Type:        MetricCounter,
			Description: "Usage count for various features",
			Unit:        "count",
			Labels:      []string{"feature_name", "user_type", "subscription_tier"},
			Category:    "feature",
			Priority:    3,
			Enabled:     true,
		},
		{
			Name:        "experiment_exposures",
			Type:        MetricCounter,
			Description: "Number of users exposed to experiments",
			Unit:        "count",
			Labels:      []string{"experiment_name", "variant", "user_segment"},
			Category:    "feature",
			Priority:    3,
			Enabled:     true,
		},
	}
}

// DefaultAlertRules returns default alert rules
func DefaultAlertRules() []MetricAlertRule {
	return []MetricAlertRule{
		{
			MetricName: "api_request_duration",
			Condition:  ">",
			Threshold:  2.0, // 2 seconds
			Duration:   5 * time.Minute,
			Severity:   "warning",
			Labels: map[string]string{
				"team": "platform",
			},
			Annotations: map[string]string{
				"summary":     "High API response time",
				"description": "API response time is above 2 seconds",
			},
			Enabled: true,
		},
		{
			MetricName: "api_requests_total",
			Condition:  "<",
			Threshold:  10, // Less than 10 requests per minute
			Duration:   10 * time.Minute,
			Severity:   "critical",
			Labels: map[string]string{
				"team": "platform",
			},
			Annotations: map[string]string{
				"summary":     "Low API traffic",
				"description": "API traffic is unusually low",
			},
			Enabled: true,
		},
		{
			MetricName: "compliance_events",
			Condition:  ">",
			Threshold:  100, // More than 100 compliance events per hour
			Duration:   time.Hour,
			Severity:   "warning",
			Labels: map[string]string{
				"team": "compliance",
			},
			Annotations: map[string]string{
				"summary":     "High compliance activity",
				"description": "Unusual increase in compliance events",
			},
			Enabled: true,
		},
	}
}

// NewBusinessMetricsCollector creates a new business metrics collector
func NewBusinessMetricsCollector(
	config BusinessMetricsConfig,
	logger *logger.Logger,
	telemetry *observability.TelemetryService,
) (*BusinessMetricsCollector, error) {

	// Initialize storage backend
	storage, err := NewMetricStorage(config.StorageBackend, config.StorageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize metric storage: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	collector := &BusinessMetricsCollector{
		config:       config,
		logger:       logger,
		telemetry:    telemetry,
		metrics:      make(map[string]*BusinessMetric),
		values:       make(map[string][]MetricValue),
		aggregations: make(map[string]map[AggregationType]AggregatedMetric),
		alertStates:  make(map[string]AlertState),
		ctx:          ctx,
		cancel:       cancel,
		storage:      storage,
	}

	// Initialize metrics
	for i := range config.CustomMetrics {
		metric := &config.CustomMetrics[i]
		collector.metrics[metric.Name] = metric
		collector.values[metric.Name] = make([]MetricValue, 0)
		collector.aggregations[metric.Name] = make(map[AggregationType]AggregatedMetric)
	}

	// Start background tasks
	collector.startBackgroundTasks()

	logger.Info("Business metrics collector initialized",
		"metrics_count", len(config.CustomMetrics),
		"storage_backend", config.StorageBackend,
		"alerting_enabled", config.AlertingEnabled,
	)

	return collector, nil
}

// RecordCounter records a counter metric
func (bmc *BusinessMetricsCollector) RecordCounter(name string, value float64, labels map[string]string) {
	bmc.recordMetric(name, value, labels)
}

// RecordGauge records a gauge metric
func (bmc *BusinessMetricsCollector) RecordGauge(name string, value float64, labels map[string]string) {
	bmc.recordMetric(name, value, labels)
}

// RecordHistogram records a histogram metric
func (bmc *BusinessMetricsCollector) RecordHistogram(name string, value float64, labels map[string]string) {
	bmc.recordMetric(name, value, labels)
}

// RecordSummary records a summary metric
func (bmc *BusinessMetricsCollector) RecordSummary(name string, value float64, labels map[string]string) {
	bmc.recordMetric(name, value, labels)
}

// recordMetric is the internal method to record any metric
func (bmc *BusinessMetricsCollector) recordMetric(name string, value float64, labels map[string]string) {
	bmc.mu.Lock()
	defer bmc.mu.Unlock()

	metric, exists := bmc.metrics[name]
	if !exists || !metric.Enabled {
		return
	}

	// Merge with default labels
	allLabels := make(map[string]string)
	for k, v := range bmc.config.DefaultLabels {
		allLabels[k] = v
	}
	for k, v := range labels {
		allLabels[k] = v
	}

	metricValue := MetricValue{
		Name:      name,
		Value:     value,
		Labels:    allLabels,
		Timestamp: time.Now(),
		Unit:      metric.Unit,
	}

	// Store in memory
	bmc.values[name] = append(bmc.values[name], metricValue)

	// Also record in telemetry service
	if bmc.telemetry != nil {
		switch metric.Type {
		case MetricCounter:
			bmc.telemetry.RecordCounter(name, value, allLabels)
		case MetricGauge:
			bmc.telemetry.RecordGauge(name, value, allLabels)
		case MetricHistogram, MetricSummary:
			bmc.telemetry.RecordHistogram(name, value, allLabels)
		}
	}

	// Trigger immediate aggregation for high-priority metrics
	if metric.Priority == 1 {
		bmc.aggregateMetric(name, metricValue)
	}
}

// GetMetricValues returns raw metric values
func (bmc *BusinessMetricsCollector) GetMetricValues(query MetricQuery) ([]MetricValue, error) {
	if bmc.storage != nil {
		return bmc.storage.Query(bmc.ctx, query)
	}

	// Fallback to in-memory values
	bmc.mu.RLock()
	defer bmc.mu.RUnlock()

	values, exists := bmc.values[query.MetricName]
	if !exists {
		return nil, nil
	}

	filtered := make([]MetricValue, 0)
	for _, value := range values {
		if bmc.matchesQuery(value, query) {
			filtered = append(filtered, value)
		}
	}

	return filtered, nil
}

// GetAggregatedMetrics returns aggregated metrics
func (bmc *BusinessMetricsCollector) GetAggregatedMetrics(query AggregationQuery) ([]AggregatedMetric, error) {
	if bmc.storage != nil {
		return bmc.storage.Aggregate(bmc.ctx, query)
	}

	// Fallback to in-memory aggregations
	bmc.mu.RLock()
	defer bmc.mu.RUnlock()

	aggregations, exists := bmc.aggregations[query.MetricName]
	if !exists {
		return nil, nil
	}

	result := make([]AggregatedMetric, 0)
	for _, aggType := range query.Aggregations {
		if aggMetric, exists := aggregations[aggType]; exists {
			result = append(result, aggMetric)
		}
	}

	return result, nil
}

// GetAlertStates returns current alert states
func (bmc *BusinessMetricsCollector) GetAlertStates() []AlertState {
	bmc.mu.RLock()
	defer bmc.mu.RUnlock()

	states := make([]AlertState, 0, len(bmc.alertStates))
	for _, state := range bmc.alertStates {
		states = append(states, state)
	}

	return states
}

// GetMetrics returns all configured metrics
func (bmc *BusinessMetricsCollector) GetMetrics() []BusinessMetric {
	bmc.mu.RLock()
	defer bmc.mu.RUnlock()

	metrics := make([]BusinessMetric, 0, len(bmc.metrics))
	for _, metric := range bmc.metrics {
		metrics = append(metrics, *metric)
	}

	return metrics
}

// Close gracefully shuts down the collector
func (bmc *BusinessMetricsCollector) Close() error {
	bmc.logger.Info("Shutting down business metrics collector")

	// Cancel context and wait for background tasks
	bmc.cancel()
	bmc.wg.Wait()

	// Close storage
	if bmc.storage != nil {
		return bmc.storage.Close()
	}

	return nil
}

// Private methods

func (bmc *BusinessMetricsCollector) startBackgroundTasks() {
	// Collection and aggregation task
	bmc.wg.Add(1)
	go bmc.collectionTask()

	// Alert processing task
	if bmc.config.AlertingEnabled {
		bmc.wg.Add(1)
		go bmc.alertTask()
	}

	// Export task
	if bmc.config.ExportEnabled {
		bmc.wg.Add(1)
		go bmc.exportTask()
	}

	// Cleanup task
	bmc.wg.Add(1)
	go bmc.cleanupTask()
}

func (bmc *BusinessMetricsCollector) collectionTask() {
	defer bmc.wg.Done()

	ticker := time.NewTicker(bmc.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bmc.ctx.Done():
			return
		case <-ticker.C:
			bmc.performCollection()
		}
	}
}

func (bmc *BusinessMetricsCollector) performCollection() {
	bmc.mu.Lock()
	defer bmc.mu.Unlock()

	// Process aggregations for all metrics
	for metricName := range bmc.metrics {
		values := bmc.values[metricName]
		for _, value := range values {
			bmc.aggregateMetric(metricName, value)
		}
	}

	// Store values in persistent storage
	if bmc.storage != nil {
		for metricName, values := range bmc.values {
			if len(values) > 0 {
				if err := bmc.storage.Store(bmc.ctx, values); err != nil {
					bmc.logger.Error("Failed to store metric values",
						"metric", metricName,
						"count", len(values),
						"error", err,
					)
				}

				// Clear in-memory values after successful storage
				bmc.values[metricName] = make([]MetricValue, 0)
			}
		}
	}
}

func (bmc *BusinessMetricsCollector) aggregateMetric(metricName string, value MetricValue) {
	metric := bmc.metrics[metricName]
	if metric == nil {
		return
	}

	aggregations := bmc.aggregations[metricName]

	for _, aggType := range metric.Aggregations {
		existing, exists := aggregations[aggType]
		if !exists {
			existing = AggregatedMetric{
				MetricValue: MetricValue{
					Name:      metricName,
					Labels:    value.Labels,
					Timestamp: value.Timestamp,
					Unit:      value.Unit,
				},
				Aggregation: aggType,
				Period:      bmc.config.CollectionInterval,
				Count:       0,
			}
		}

		// Apply aggregation logic
		switch aggType {
		case AggregationSum:
			existing.Value += value.Value
		case AggregationAvg:
			existing.Value = (existing.Value*float64(existing.Count) + value.Value) / float64(existing.Count+1)
		case AggregationMax:
			if value.Value > existing.Value {
				existing.Value = value.Value
			}
		case AggregationMin:
			if existing.Count == 0 || value.Value < existing.Value {
				existing.Value = value.Value
			}
		case AggregationCount:
			existing.Value = float64(existing.Count + 1)
		}

		existing.Count++
		existing.Timestamp = time.Now()

		aggregations[aggType] = existing
	}
}

func (bmc *BusinessMetricsCollector) alertTask() {
	defer bmc.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-bmc.ctx.Done():
			return
		case <-ticker.C:
			bmc.processAlerts()
		}
	}
}

func (bmc *BusinessMetricsCollector) processAlerts() {
	for _, rule := range bmc.config.AlertRules {
		if !rule.Enabled {
			continue
		}

		bmc.evaluateAlertRule(rule)
	}
}

func (bmc *BusinessMetricsCollector) evaluateAlertRule(rule MetricAlertRule) {
	// Get current metric value
	query := MetricQuery{
		MetricName: rule.MetricName,
		StartTime:  time.Now().Add(-time.Minute),
		EndTime:    time.Now(),
		Limit:      1,
	}

	values, err := bmc.GetMetricValues(query)
	if err != nil || len(values) == 0 {
		return
	}

	currentValue := values[len(values)-1].Value

	// Evaluate condition
	conditionMet := false
	switch rule.Condition {
	case ">":
		conditionMet = currentValue > rule.Threshold
	case "<":
		conditionMet = currentValue < rule.Threshold
	case ">=":
		conditionMet = currentValue >= rule.Threshold
	case "<=":
		conditionMet = currentValue <= rule.Threshold
	case "==":
		conditionMet = currentValue == rule.Threshold
	case "!=":
		conditionMet = currentValue != rule.Threshold
	}

	// Update alert state
	alertKey := fmt.Sprintf("%s-%s", rule.MetricName, rule.Severity)

	bmc.mu.Lock()
	existingState, exists := bmc.alertStates[alertKey]
	bmc.mu.Unlock()

	if conditionMet {
		if !exists || existingState.State == string(AlertStateResolved) {
			// New alert or previously resolved
			newState := AlertState{
				MetricName:  rule.MetricName,
				RuleName:    alertKey,
				State:       "pending",
				Value:       currentValue,
				Threshold:   rule.Threshold,
				Since:       time.Now(),
				Labels:      rule.Labels,
				Annotations: rule.Annotations,
			}

			bmc.mu.Lock()
			bmc.alertStates[alertKey] = newState
			bmc.mu.Unlock()

		} else if existingState.State == "pending" && time.Since(existingState.Since) >= rule.Duration {
			// Promote to firing
			existingState.State = "firing"
			existingState.Value = currentValue

			bmc.mu.Lock()
			bmc.alertStates[alertKey] = existingState
			bmc.mu.Unlock()

			bmc.logger.Warn("Alert firing",
				"metric", rule.MetricName,
				"condition", fmt.Sprintf("%s %f", rule.Condition, rule.Threshold),
				"current_value", currentValue,
				"severity", rule.Severity,
			)
		}
	} else if exists && existingState.State != string(AlertStateResolved) {
		// Resolve alert
		existingState.State = string(AlertStateResolved)
		existingState.Value = currentValue

		bmc.mu.Lock()
		bmc.alertStates[alertKey] = existingState
		bmc.mu.Unlock()

		bmc.logger.Info("Alert resolved",
			"metric", rule.MetricName,
			"current_value", currentValue,
		)
	}
}

func (bmc *BusinessMetricsCollector) exportTask() {
	defer bmc.wg.Done()

	ticker := time.NewTicker(bmc.config.ExportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bmc.ctx.Done():
			return
		case <-ticker.C:
			bmc.performExport()
		}
	}
}

func (bmc *BusinessMetricsCollector) performExport() {
	// Implementation would depend on export format and endpoint
	bmc.logger.Debug("Performing metrics export",
		"format", bmc.config.ExportFormat,
		"endpoint", bmc.config.ExportEndpoint,
	)
}

func (bmc *BusinessMetricsCollector) cleanupTask() {
	defer bmc.wg.Done()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-bmc.ctx.Done():
			return
		case <-ticker.C:
			bmc.performCleanup()
		}
	}
}

func (bmc *BusinessMetricsCollector) performCleanup() {
	cutoff := time.Now().Add(-bmc.config.RetentionPeriod)

	// Clean up in-memory values
	bmc.mu.Lock()
	for metricName, values := range bmc.values {
		filtered := make([]MetricValue, 0)
		for _, value := range values {
			if value.Timestamp.After(cutoff) {
				filtered = append(filtered, value)
			}
		}
		bmc.values[metricName] = filtered
	}
	bmc.mu.Unlock()

	// Clean up persistent storage
	if bmc.storage != nil {
		if err := bmc.storage.Delete(bmc.ctx, cutoff); err != nil {
			bmc.logger.Error("Failed to clean up old metrics", "error", err)
		}
	}
}

func (bmc *BusinessMetricsCollector) matchesQuery(value MetricValue, query MetricQuery) bool {
	// Check time range
	if !query.StartTime.IsZero() && value.Timestamp.Before(query.StartTime) {
		return false
	}
	if !query.EndTime.IsZero() && value.Timestamp.After(query.EndTime) {
		return false
	}

	// Check labels
	for k, v := range query.Labels {
		if value.Labels[k] != v {
			return false
		}
	}

	return true
}

// NewMetricStorage creates a new metric storage backend
func NewMetricStorage(backend string, _ map[string]interface{}) (MetricStorage, error) {
	switch backend {
	case "memory":
		return NewMemoryMetricStorage(), nil
	case "redis":
		// TODO: Implement Redis storage
		return NewMemoryMetricStorage(), nil
	case "postgres":
		// TODO: Implement PostgreSQL storage
		return NewMemoryMetricStorage(), nil
	default:
		return NewMemoryMetricStorage(), nil
	}
}
