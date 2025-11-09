package observability

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// MetricsCollector coleta e expõe métricas
type MetricsCollector struct {
	registry *prometheus.Registry
	logger   *zap.Logger
	config   *MetricsConfig

	// Métricas HTTP
	httpRequestsTotal   *prometheus.CounterVec
	httpRequestDuration *prometheus.HistogramVec
	httpResponseSize    *prometheus.HistogramVec

	// Métricas WASM
	wasmExecutionsTotal   *prometheus.CounterVec
	wasmExecutionDuration *prometheus.HistogramVec
	wasmMemoryUsage       *prometheus.GaugeVec
	wasmLoadedModules     *prometheus.GaugeVec

	// Métricas SDK
	sdkRequestsTotal     *prometheus.CounterVec
	sdkRequestDuration   *prometheus.HistogramVec
	sdkPluginHealth      *prometheus.GaugeVec
	sdkActiveConnections *prometheus.GaugeVec

	// Métricas WebSocket
	wsConnections        *prometheus.GaugeVec
	wsMessagesTotal      *prometheus.CounterVec
	wsConnectionDuration *prometheus.HistogramVec

	// Métricas NATS
	natsMessagesTotal    *prometheus.CounterVec
	natsMessageSize      *prometheus.HistogramVec
	natsConnectionErrors *prometheus.CounterVec

	// Métricas do sistema
	systemUptime      prometheus.Gauge
	systemMemoryUsage prometheus.Gauge
	systemCPUUsage    prometheus.Gauge

	startTime time.Time
	mu        sync.RWMutex
}

type MetricsConfig struct {
	Enabled           bool                `json:"enabled"`
	Port              string              `json:"port"`
	Path              string              `json:"path"`
	Namespace         string              `json:"namespace"`
	Subsystem         string              `json:"subsystem"`
	Buckets           []float64           `json:"buckets"`
	Objectives        map[float64]float64 `json:"objectives"`
	PrometheusAddress string              `json:"prometheus_address"`
	ScrapeInterval    time.Duration       `json:"scrape_interval"`
}

func NewMetricsCollector(config *MetricsConfig, logger *zap.Logger) *MetricsCollector {
	if config == nil {
		config = &MetricsConfig{
			Enabled:    true,
			Port:       "9090",
			Path:       "/metrics",
			Namespace:  "web_wasm",
			Subsystem:  "mcp",
			Buckets:    []float64{0.1, 0.5, 1, 2.5, 5, 10, 30, 60, 120, 300},
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		}
	}

	collector := &MetricsCollector{
		registry:  prometheus.NewRegistry(),
		logger:    logger.Named("metrics"),
		config:    config,
		startTime: time.Now(),
	}

	if config.Enabled {
		collector.initializeMetrics()
		collector.startMetricsServer()
	}

	return collector
}

func (mc *MetricsCollector) initializeMetrics() {
	// Inicializar métricas HTTP
	mc.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code", "handler"},
	)

	mc.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   mc.config.Buckets,
		},
		[]string{"method", "path", "status_code", "handler"},
	)

	mc.httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   []float64{100, 1000, 10000, 100000, 1000000, 10000000},
		},
		[]string{"method", "path", "status_code"},
	)

	// Inicializar métricas WASM
	mc.wasmExecutionsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "wasm_executions_total",
			Help:      "Total number of WASM executions",
		},
		[]string{"function", "status"},
	)

	mc.wasmExecutionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "wasm_execution_duration_seconds",
			Help:      "WASM execution duration in seconds",
			Buckets:   mc.config.Buckets,
		},
		[]string{"function"},
	)

	mc.wasmMemoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "wasm_memory_usage_bytes",
			Help:      "WASM memory usage in bytes",
		},
		[]string{"module"},
	)

	mc.wasmLoadedModules = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "wasm_loaded_modules",
			Help:      "Number of loaded WASM modules",
		},
		[]string{"status"},
	)

	// Inicializar métricas SDK
	mc.sdkRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "sdk_requests_total",
			Help:      "Total number of SDK requests",
		},
		[]string{"plugin", "method", "status"},
	)

	mc.sdkRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "sdk_request_duration_seconds",
			Help:      "SDK request duration in seconds",
			Buckets:   mc.config.Buckets,
		},
		[]string{"plugin", "method"},
	)

	mc.sdkPluginHealth = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "sdk_plugin_health",
			Help:      "Health status of SDK plugins (1=healthy, 0=unhealthy)",
		},
		[]string{"plugin", "version"},
	)

	mc.sdkActiveConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "sdk_active_connections",
			Help:      "Number of active SDK connections",
		},
		[]string{"type"},
	)

	// Inicializar métricas WebSocket
	mc.wsConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "websocket_connections",
			Help:      "Number of active WebSocket connections",
		},
		[]string{"status"},
	)

	mc.wsMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "websocket_messages_total",
			Help:      "Total number of WebSocket messages",
		},
		[]string{"direction", "type"},
	)

	mc.wsConnectionDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "websocket_connection_duration_seconds",
			Help:      "WebSocket connection duration in seconds",
			Buckets:   []float64{10, 60, 300, 900, 3600, 7200},
		},
		[]string{"status"},
	)

	// Inicializar métricas NATS
	mc.natsMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "nats_messages_total",
			Help:      "Total number of NATS messages",
		},
		[]string{"subject", "direction", "status"},
	)

	mc.natsMessageSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "nats_message_size_bytes",
			Help:      "NATS message size in bytes",
			Buckets:   []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"subject"},
	)

	mc.natsConnectionErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "nats_connection_errors_total",
			Help:      "Total number of NATS connection errors",
		},
		[]string{"error_type"},
	)

	// Inicializar métricas do sistema
	mc.systemUptime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "uptime_seconds",
			Help:      "System uptime in seconds",
		},
	)

	mc.systemMemoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "memory_usage_bytes",
			Help:      "System memory usage in bytes",
		},
	)

	mc.systemCPUUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: mc.config.Namespace,
			Subsystem: mc.config.Subsystem,
			Name:      "cpu_usage_percent",
			Help:      "System CPU usage percentage",
		},
	)

	// Registrar métricas
	mc.registerMetrics()
}

func (mc *MetricsCollector) registerMetrics() {
	mc.registry.MustRegister(mc.httpRequestsTotal)
	mc.registry.MustRegister(mc.httpRequestDuration)
	mc.registry.MustRegister(mc.httpResponseSize)
	mc.registry.MustRegister(mc.wasmExecutionsTotal)
	mc.registry.MustRegister(mc.wasmExecutionDuration)
	mc.registry.MustRegister(mc.wasmMemoryUsage)
	mc.registry.MustRegister(mc.wasmLoadedModules)
	mc.registry.MustRegister(mc.sdkRequestsTotal)
	mc.registry.MustRegister(mc.sdkRequestDuration)
	mc.registry.MustRegister(mc.sdkPluginHealth)
	mc.registry.MustRegister(mc.sdkActiveConnections)
	mc.registry.MustRegister(mc.wsConnections)
	mc.registry.MustRegister(mc.wsMessagesTotal)
	mc.registry.MustRegister(mc.wsConnectionDuration)
	mc.registry.MustRegister(mc.natsMessagesTotal)
	mc.registry.MustRegister(mc.natsMessageSize)
	mc.registry.MustRegister(mc.natsConnectionErrors)
	mc.registry.MustRegister(mc.systemUptime)
	mc.registry.MustRegister(mc.systemMemoryUsage)
	mc.registry.MustRegister(mc.systemCPUUsage)
}

func (mc *MetricsCollector) startMetricsServer() {
	if !mc.config.Enabled {
		return
	}

	mux := http.NewServeMux()
	mux.Handle(mc.config.Path, promhttp.HandlerFor(mc.registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:    ":" + mc.config.Port,
		Handler: mux,
	}

	go func() {
		mc.logger.Info("Metrics server started",
			zap.String("port", mc.config.Port),
			zap.String("path", mc.config.Path))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			mc.logger.Error("Metrics server error", zap.Error(err))
		}
	}()

	// Atualizar métricas do sistema periodicamente
	go mc.updateSystemMetrics()
}

// Métodos para registrar métricas HTTP
func (mc *MetricsCollector) RecordHTTPRequest(method, path, statusCode, handler string, duration float64, size int64) {
	mc.httpRequestsTotal.WithLabelValues(method, path, statusCode, handler).Inc()
	mc.httpRequestDuration.WithLabelValues(method, path, statusCode, handler).Observe(duration)
	mc.httpResponseSize.WithLabelValues(method, path, statusCode).Observe(float64(size))
}

// Métodos para registrar métricas WASM
func (mc *MetricsCollector) RecordWASMExecution(function, status string, duration float64) {
	mc.wasmExecutionsTotal.WithLabelValues(function, status).Inc()
	mc.wasmExecutionDuration.WithLabelValues(function).Observe(duration)
}

func (mc *MetricsCollector) UpdateWASMMemoryUsage(module string, bytes int64) {
	mc.wasmMemoryUsage.WithLabelValues(module).Set(float64(bytes))
}

func (mc *MetricsCollector) UpdateWASMLoadedModules(status string, count int) {
	mc.wasmLoadedModules.WithLabelValues(status).Set(float64(count))
}

// Métodos para registrar métricas SDK
func (mc *MetricsCollector) RecordSDKRequest(plugin, method, status string, duration float64) {
	mc.sdkRequestsTotal.WithLabelValues(plugin, method, status).Inc()
	mc.sdkRequestDuration.WithLabelValues(plugin, method).Observe(duration)
}

func (mc *MetricsCollector) UpdateSDKPluginHealth(plugin, version string, healthy bool) {
	value := 0.0
	if healthy {
		value = 1.0
	}
	mc.sdkPluginHealth.WithLabelValues(plugin, version).Set(value)
}

func (mc *MetricsCollector) UpdateSDKActiveConnections(connectionType string, count int) {
	mc.sdkActiveConnections.WithLabelValues(connectionType).Set(float64(count))
}

// Métodos para registrar métricas WebSocket
func (mc *MetricsCollector) UpdateWebSocketConnections(status string, count int) {
	mc.wsConnections.WithLabelValues(status).Set(float64(count))
}

func (mc *MetricsCollector) RecordWebSocketMessage(direction, messageType string) {
	mc.wsMessagesTotal.WithLabelValues(direction, messageType).Inc()
}

func (mc *MetricsCollector) RecordWebSocketConnection(status string, duration float64) {
	mc.wsConnectionDuration.WithLabelValues(status).Observe(duration)
}

// Métodos para registrar métricas NATS
func (mc *MetricsCollector) RecordNATSMessage(subject, direction, status string, size int64) {
	mc.natsMessagesTotal.WithLabelValues(subject, direction, status).Inc()
	mc.natsMessageSize.WithLabelValues(subject).Observe(float64(size))
}

func (mc *MetricsCollector) RecordNATSConnectionError(errorType string) {
	mc.natsConnectionErrors.WithLabelValues(errorType).Inc()
}

// Método para obter métricas do Prometheus
func (mc *MetricsCollector) GetPrometheusMetrics(ctx context.Context, query string) (interface{}, error) {
	if mc.config.PrometheusAddress == "" {
		return nil, fmt.Errorf("Prometheus address not configured")
	}

	client, err := api.NewClient(api.Config{
		Address: mc.config.PrometheusAddress,
	})
	if err != nil {
		return nil, err
	}

	v1api := v1.NewAPI(client)
	result, warnings, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"result":   result,
		"warnings": warnings,
	}, nil
}

// updateSystemMetrics atualiza métricas do sistema
func (mc *MetricsCollector) updateSystemMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		mc.mu.Lock()

		// Atualizar uptime
		mc.systemUptime.Set(time.Since(mc.startTime).Seconds())

		// TODO: Implementar coleta de métricas reais do sistema
		// - Memória usada pelo processo
		// - CPU usage
		// - Número de goroutines
		// - Garbage collector stats

		mc.mu.Unlock()
	}
}

// GetRegistry retorna o registry do Prometheus
func (mc *MetricsCollector) GetRegistry() *prometheus.Registry {
	return mc.registry
}

// GetConfig retorna a configuração das métricas
func (mc *MetricsCollector) GetConfig() *MetricsConfig {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return mc.config
}
