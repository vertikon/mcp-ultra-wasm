package dashboard

import (
	"time"
)

// SystemOverview represents the overall system status
type SystemOverview struct {
	Timestamp    time.Time         `json:"timestamp"`
	SystemHealth SystemHealth      `json:"system_health"`
	Components   []ComponentStatus `json:"components"`
	Metrics      OverviewMetrics   `json:"metrics"`
	Alerts       []Alert           `json:"alerts"`
}

// SystemHealth represents overall system health status
type SystemHealth struct {
	Status        string        `json:"status"`        // healthy, degraded, unhealthy
	OverallScore  float64       `json:"overall_score"` // 0-100
	Uptime        time.Duration `json:"uptime"`
	LastIncident  *time.Time    `json:"last_incident,omitempty"`
	SLOCompliance float64       `json:"slo_compliance"` // Percentage
}

// ComponentStatus represents individual component status
type ComponentStatus struct {
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	Status       string             `json:"status"`
	Health       float64            `json:"health"`
	LastCheck    time.Time          `json:"last_check"`
	Message      string             `json:"message,omitempty"`
	Dependencies []string           `json:"dependencies,omitempty"`
	Metrics      map[string]float64 `json:"metrics,omitempty"`
}

// OverviewMetrics represents key system metrics
type OverviewMetrics struct {
	RequestRate    float64 `json:"request_rate"` // requests per second
	ErrorRate      float64 `json:"error_rate"`   // percentage
	AvgLatency     float64 `json:"avg_latency"`  // milliseconds
	P99Latency     float64 `json:"p99_latency"`  // milliseconds
	Throughput     float64 `json:"throughput"`   // operations per second
	ActiveSessions int64   `json:"active_sessions"`
	CacheHitRate   float64 `json:"cache_hit_rate"` // percentage
	CPUUsage       float64 `json:"cpu_usage"`      // percentage
	MemoryUsage    float64 `json:"memory_usage"`   // percentage
}

// Alert represents system alerts
type Alert struct {
	ID          string            `json:"id"`
	Type        AlertType         `json:"type"`
	Severity    AlertSeverity     `json:"severity"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Component   string            `json:"component,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Status      AlertStatus       `json:"status"`
	Labels      map[string]string `json:"labels,omitempty"`
	Actions     []AlertAction     `json:"actions,omitempty"`
}

// AlertType represents different types of alerts
type AlertType string

const (
	AlertTypeSystem      AlertType = "system"
	AlertTypePerformance AlertType = "performance"
	AlertTypeBusiness    AlertType = "business"
	AlertTypeSecurity    AlertType = "security"
	AlertTypeCompliance  AlertType = "compliance"
)

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	AlertSeverityInfo      AlertSeverity = "info"
	AlertSeverityWarning   AlertSeverity = "warning"
	AlertSeverityCritical  AlertSeverity = "critical"
	AlertSeverityEmergency AlertSeverity = "emergency"
)

// AlertStatus represents alert status
type AlertStatus string

const (
	AlertStatusActive       AlertStatus = "active"
	AlertStatusAcknowledged AlertStatus = "acknowledged"
	AlertStatusResolved     AlertStatus = "resolved"
	AlertStatusSuppressed   AlertStatus = "suppressed"
)

// AlertAction represents available actions for alerts
type AlertAction struct {
	Type    string `json:"type"`
	Label   string `json:"label"`
	URL     string `json:"url,omitempty"`
	Confirm bool   `json:"confirm,omitempty"`
}

// RealtimeMetrics represents real-time system metrics
type RealtimeMetrics struct {
	Timestamp     time.Time          `json:"timestamp"`
	SystemMetrics SystemMetrics      `json:"system_metrics"`
	Performance   PerformanceMetrics `json:"performance"`
	Errors        ErrorMetrics       `json:"errors"`
	Traffic       TrafficMetrics     `json:"traffic"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	CPU       CPUMetrics     `json:"cpu"`
	Memory    MemoryMetrics  `json:"memory"`
	Disk      DiskMetrics    `json:"disk"`
	Network   NetworkMetrics `json:"network"`
	Processes ProcessMetrics `json:"processes"`
}

// CPUMetrics represents CPU usage metrics
type CPUMetrics struct {
	Usage       float64 `json:"usage"` // percentage
	LoadAvg1m   float64 `json:"load_avg_1m"`
	LoadAvg5m   float64 `json:"load_avg_5m"`
	LoadAvg15m  float64 `json:"load_avg_15m"`
	Cores       int     `json:"cores"`
	Temperature float64 `json:"temperature,omitempty"`
}

// MemoryMetrics represents memory usage metrics
type MemoryMetrics struct {
	Total     int64   `json:"total"`      // bytes
	Used      int64   `json:"used"`       // bytes
	Available int64   `json:"available"`  // bytes
	Usage     float64 `json:"usage"`      // percentage
	SwapTotal int64   `json:"swap_total"` // bytes
	SwapUsed  int64   `json:"swap_used"`  // bytes
	SwapUsage float64 `json:"swap_usage"` // percentage
}

// DiskMetrics represents disk usage metrics
type DiskMetrics struct {
	Total     int64   `json:"total"`      // bytes
	Used      int64   `json:"used"`       // bytes
	Available int64   `json:"available"`  // bytes
	Usage     float64 `json:"usage"`      // percentage
	IOPS      float64 `json:"iops"`       // operations per second
	ReadRate  float64 `json:"read_rate"`  // bytes per second
	WriteRate float64 `json:"write_rate"` // bytes per second
}

// NetworkMetrics represents network usage metrics
type NetworkMetrics struct {
	BytesReceived   int64   `json:"bytes_received"`
	BytesSent       int64   `json:"bytes_sent"`
	PacketsReceived int64   `json:"packets_received"`
	PacketsSent     int64   `json:"packets_sent"`
	ReceiveRate     float64 `json:"receive_rate"`  // bytes per second
	TransmitRate    float64 `json:"transmit_rate"` // bytes per second
	Connections     int     `json:"connections"`
	DroppedPackets  int64   `json:"dropped_packets"`
	ErrorPackets    int64   `json:"error_packets"`
}

// ProcessMetrics represents process-level metrics
type ProcessMetrics struct {
	Total    int `json:"total"`
	Running  int `json:"running"`
	Sleeping int `json:"sleeping"`
	Zombie   int `json:"zombie"`
	Stopped  int `json:"stopped"`
}

// PerformanceMetrics represents application performance metrics
type PerformanceMetrics struct {
	RequestRate     float64             `json:"request_rate"` // requests per second
	ResponseTime    ResponseTimeMetrics `json:"response_time"`
	Throughput      float64             `json:"throughput"`  // operations per second
	Concurrency     int                 `json:"concurrency"` // active concurrent operations
	QueueDepth      int                 `json:"queue_depth"` // pending operations
	DatabaseMetrics DatabaseMetrics     `json:"database"`
	CacheMetrics    CacheMetricsData    `json:"cache"`
}

// ResponseTimeMetrics represents response time statistics
type ResponseTimeMetrics struct {
	Mean   float64 `json:"mean"`   // milliseconds
	Median float64 `json:"median"` // milliseconds
	P95    float64 `json:"p95"`    // milliseconds
	P99    float64 `json:"p99"`    // milliseconds
	P999   float64 `json:"p999"`   // milliseconds
	Min    float64 `json:"min"`    // milliseconds
	Max    float64 `json:"max"`    // milliseconds
}

// DatabaseMetrics represents database performance metrics
type DatabaseMetrics struct {
	Connections       int     `json:"connections"`
	ActiveConnections int     `json:"active_connections"`
	QueryRate         float64 `json:"query_rate"`                // queries per second
	SlowQueries       int64   `json:"slow_queries"`              // count
	Deadlocks         int64   `json:"deadlocks"`                 // count
	LockWaitTime      float64 `json:"lock_wait_time"`            // milliseconds
	ReplicationLag    float64 `json:"replication_lag,omitempty"` // milliseconds
}

// CacheMetricsData represents cache performance metrics
type CacheMetricsData struct {
	HitRate       float64 `json:"hit_rate"`  // percentage
	MissRate      float64 `json:"miss_rate"` // percentage
	Evictions     int64   `json:"evictions"` // count
	KeyCount      int64   `json:"key_count"`
	MemoryUsage   int64   `json:"memory_usage"`   // bytes
	OperationRate float64 `json:"operation_rate"` // operations per second
}

// ErrorMetrics represents error tracking metrics
type ErrorMetrics struct {
	TotalErrors    int64            `json:"total_errors"`
	ErrorRate      float64          `json:"error_rate"` // percentage
	ErrorsByType   map[string]int64 `json:"errors_by_type"`
	ErrorsByCode   map[string]int64 `json:"errors_by_code"`
	CriticalErrors int64            `json:"critical_errors"`
	RecentErrors   []RecentError    `json:"recent_errors"`
}

// RecentError represents recent error information
type RecentError struct {
	Timestamp time.Time         `json:"timestamp"`
	Type      string            `json:"type"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Component string            `json:"component"`
	TraceID   string            `json:"trace_id,omitempty"`
	Count     int               `json:"count"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// TrafficMetrics represents traffic and usage metrics
type TrafficMetrics struct {
	TotalRequests    int64            `json:"total_requests"`
	ActiveSessions   int64            `json:"active_sessions"`
	UniqueUsers      int64            `json:"unique_users"`
	TrafficBySource  map[string]int64 `json:"traffic_by_source"`
	TrafficByRegion  map[string]int64 `json:"traffic_by_region"`
	TrafficByChannel map[string]int64 `json:"traffic_by_channel"`
	PeakTraffic      TrafficPeak      `json:"peak_traffic"`
	Bandwidth        BandwidthMetrics `json:"bandwidth"`
}

// TrafficPeak represents peak traffic information
type TrafficPeak struct {
	Timestamp   time.Time `json:"timestamp"`
	RequestRate float64   `json:"request_rate"` // requests per second
	Users       int64     `json:"users"`
	Sessions    int64     `json:"sessions"`
}

// BandwidthMetrics represents bandwidth usage
type BandwidthMetrics struct {
	Incoming float64 `json:"incoming"` // bytes per second
	Outgoing float64 `json:"outgoing"` // bytes per second
	Total    float64 `json:"total"`    // bytes per second
	Peak     float64 `json:"peak"`     // bytes per second
	Usage    float64 `json:"usage"`    // percentage of available
}

// ChartData represents time-series data for charts
type ChartData struct {
	Labels   []string  `json:"labels"`
	Datasets []Dataset `json:"datasets"`
}

// Dataset represents a data series in a chart
type Dataset struct {
	Label           string    `json:"label"`
	Data            []float64 `json:"data"`
	BackgroundColor string    `json:"backgroundColor,omitempty"`
	BorderColor     string    `json:"borderColor,omitempty"`
	Fill            bool      `json:"fill,omitempty"`
}

// Widget represents a dashboard widget configuration
type Widget struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Size     WidgetSize             `json:"size"`
	Position WidgetPosition         `json:"position"`
	Config   map[string]interface{} `json:"config"`
	Data     interface{}            `json:"data,omitempty"`
}

// WidgetSize represents widget dimensions
type WidgetSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// WidgetPosition represents widget position
type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WebSocketMessage represents messages sent via WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
	ClientID  string      `json:"client_id,omitempty"`
}

// SubscriptionRequest represents WebSocket subscription requests
type SubscriptionRequest struct {
	Type     string                 `json:"type"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
	Interval time.Duration          `json:"interval,omitempty"`
}
