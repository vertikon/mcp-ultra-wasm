package slo

import (
	"time"
)

// DefaultSLOs returns the default SLO configuration for MCP Ultra
func DefaultSLOs() []*SLO {
	return []*SLO{
		// API Availability SLO - 99.9% uptime
		{
			Name:        "api_availability",
			Description: "API availability - percentage of successful requests",
			Type:        TypeAvailability,
			Service:     "mcp-ultra-wasm",
			Component:   "api",

			Target:            99.9,
			WarningThreshold:  99.5,
			CriticalThreshold: 99.0,

			Query: `(
				sum(rate(http_requests_total{job="mcp-ultra-wasm", code!~"5.."}[5m])) /
				sum(rate(http_requests_total{job="mcp-ultra-wasm"}[5m]))
			) * 100`,

			ErrorBudgetQuery: `(
				(1 - (
					sum(rate(http_requests_total{job="mcp-ultra-wasm", code!~"5.."}[30d])) /
					sum(rate(http_requests_total{job="mcp-ultra-wasm"}[30d]))
				)) * 100
			) / 0.1 * 100`, // 0.1% error budget for 99.9% target

			BurnRateQuery: `(
				1 - (
					sum(rate(http_requests_total{job="mcp-ultra-wasm", code!~"5.."}[1h])) /
					sum(rate(http_requests_total{job="mcp-ultra-wasm"}[1h]))
				)
			) / 0.001`, // Burn rate relative to 99.9% target

			EvaluationWindow: 5 * time.Minute,
			ComplianceWindow: 30 * 24 * time.Hour,

			AlertingRules: []AlertRule{
				{
					Name:       "SLOErrorBudgetLow",
					Expression: "mcp_ultra_wasm_slo_error_budget_remaining_percent{slo=\"api_availability\"} < 10",
					For:        5 * time.Minute,
					Severity:   "warning",
					Labels: map[string]string{
						"service":   "mcp-ultra-wasm",
						"component": "api",
						"slo_type":  "availability",
					},
					Annotations: map[string]string{
						"summary":     "API availability error budget is low",
						"description": "The error budget for API availability is below 10%",
						"runbook":     "https://runbooks.mcp-ultra-wasm.com/slo/api-availability",
					},
					Enabled: true,
				},
				{
					Name:       "SLOBurnRateHigh",
					Expression: "mcp_ultra_wasm_slo_burn_rate{slo=\"api_availability\"} > 14.4",
					For:        2 * time.Minute,
					Severity:   "critical",
					Labels: map[string]string{
						"service":   "mcp-ultra-wasm",
						"component": "api",
						"slo_type":  "availability",
					},
					Annotations: map[string]string{
						"summary":     "API availability burn rate is high",
						"description": "The burn rate for API availability indicates rapid error budget consumption",
						"runbook":     "https://runbooks.mcp-ultra-wasm.com/slo/burn-rate",
					},
					Enabled: true,
				},
			},

			Tags: map[string]string{
				"team":        "platform",
				"tier":        "1",
				"criticality": "high",
			},

			Enabled: true,
		},

		// API Latency SLO - 95% of requests under 500ms
		{
			Name:        "api_latency_p95",
			Description: "API latency - 95th percentile under 500ms",
			Type:        TypeLatency,
			Service:     "mcp-ultra-wasm",
			Component:   "api",

			Target:            95.0,
			WarningThreshold:  90.0,
			CriticalThreshold: 85.0,

			Query: `(
				histogram_quantile(0.95, 
					sum(rate(http_request_duration_seconds_bucket{job="mcp-ultra-wasm"}[5m])) by (le)
				) < 0.5
			) * 100`,

			EvaluationWindow: 5 * time.Minute,
			ComplianceWindow: 7 * 24 * time.Hour, // Weekly window for latency

			AlertingRules: []AlertRule{
				{
					Name:       "SLOLatencyDegraded",
					Expression: "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job=\"mcp-ultra-wasm\"}[5m])) by (le)) > 0.5",
					For:        10 * time.Minute,
					Severity:   "warning",
					Labels: map[string]string{
						"service":   "mcp-ultra-wasm",
						"component": "api",
						"slo_type":  "latency",
					},
					Annotations: map[string]string{
						"summary":     "API latency SLO is degraded",
						"description": "95th percentile latency is above 500ms",
						"runbook":     "https://runbooks.mcp-ultra-wasm.com/slo/latency",
					},
					Enabled: true,
				},
			},

			Tags: map[string]string{
				"team":        "platform",
				"tier":        "2",
				"criticality": "medium",
			},

			Enabled: true,
		},

		// gRPC Availability SLO - 99.9% success rate
		{
			Name:        "grpc_availability",
			Description: "gRPC service availability - percentage of successful requests",
			Type:        TypeAvailability,
			Service:     "mcp-ultra-wasm",
			Component:   "grpc",

			Target:            99.9,
			WarningThreshold:  99.5,
			CriticalThreshold: 99.0,

			Query: `(
				sum(rate(grpc_server_handled_total{job="mcp-ultra-wasm", grpc_code="OK"}[5m])) /
				sum(rate(grpc_server_handled_total{job="mcp-ultra-wasm"}[5m]))
			) * 100`,

			EvaluationWindow: 5 * time.Minute,
			ComplianceWindow: 30 * 24 * time.Hour,

			AlertingRules: []AlertRule{
				{
					Name:       "SLOgRPCErrorBudgetLow",
					Expression: "mcp_ultra_wasm_slo_error_budget_remaining_percent{slo=\"grpc_availability\"} < 10",
					For:        5 * time.Minute,
					Severity:   "warning",
					Labels: map[string]string{
						"service":   "mcp-ultra-wasm",
						"component": "grpc",
						"slo_type":  "availability",
					},
					Annotations: map[string]string{
						"summary":     "gRPC availability error budget is low",
						"description": "The error budget for gRPC availability is below 10%",
						"runbook":     "https://runbooks.mcp-ultra-wasm.com/slo/grpc-availability",
					},
					Enabled: true,
				},
			},

			Tags: map[string]string{
				"team":        "platform",
				"tier":        "1",
				"criticality": "high",
			},

			Enabled: true,
		},

		// Database Connection Health SLO - 99.5% availability
		{
			Name:        "database_availability",
			Description: "Database connection availability - percentage of successful connections",
			Type:        TypeAvailability,
			Service:     "mcp-ultra-wasm",
			Component:   "database",

			Target:            99.5,
			WarningThreshold:  99.0,
			CriticalThreshold: 98.0,

			Query: `(
				up{job="postgres"} and 
				on (instance) pg_up{job="postgres"}
			) * 100`,

			EvaluationWindow: 1 * time.Minute,
			ComplianceWindow: 30 * 24 * time.Hour,

			AlertingRules: []AlertRule{
				{
					Name:       "SLODatabaseDown",
					Expression: "up{job=\"postgres\"} == 0 or pg_up{job=\"postgres\"} == 0",
					For:        1 * time.Minute,
					Severity:   "critical",
					Labels: map[string]string{
						"service":   "mcp-ultra-wasm",
						"component": "database",
						"slo_type":  "availability",
					},
					Annotations: map[string]string{
						"summary":     "Database availability SLO violation",
						"description": "Database is not responding",
						"runbook":     "https://runbooks.mcp-ultra-wasm.com/slo/database",
					},
					Enabled: true,
				},
			},

			Tags: map[string]string{
				"team":        "platform",
				"tier":        "1",
				"criticality": "critical",
			},

			Enabled: true,
		},

		// Cache Hit Rate SLO - 90% cache hit rate
		{
			Name:        "cache_hit_rate",
			Description: "Redis cache hit rate - percentage of cache hits",
			Type:        TypeAccuracy,
			Service:     "mcp-ultra-wasm",
			Component:   "cache",

			Target:            90.0,
			WarningThreshold:  85.0,
			CriticalThreshold: 80.0,

			Query: `(
				sum(rate(redis_keyspace_hits_total{job="redis"}[5m])) /
				(sum(rate(redis_keyspace_hits_total{job="redis"}[5m])) + 
				 sum(rate(redis_keyspace_misses_total{job="redis"}[5m])))
			) * 100`,

			EvaluationWindow: 5 * time.Minute,
			ComplianceWindow: 7 * 24 * time.Hour,

			AlertingRules: []AlertRule{
				{
					Name:       "SLOCacheHitRateLow",
					Expression: "mcp_ultra_wasm_slo_current_value{slo=\"cache_hit_rate\"} < 85",
					For:        10 * time.Minute,
					Severity:   "warning",
					Labels: map[string]string{
						"service":   "mcp-ultra-wasm",
						"component": "cache",
						"slo_type":  "accuracy",
					},
					Annotations: map[string]string{
						"summary":     "Cache hit rate is low",
						"description": "Redis cache hit rate is below expected threshold",
						"runbook":     "https://runbooks.mcp-ultra-wasm.com/slo/cache",
					},
					Enabled: true,
				},
			},

			Tags: map[string]string{
				"team":        "platform",
				"tier":        "3",
				"criticality": "low",
			},

			Enabled: true,
		},

		// Task Processing Throughput SLO - 1000 tasks/minute
		{
			Name:        "task_processing_throughput",
			Description: "Task processing throughput - tasks processed per minute",
			Type:        TypeThroughput,
			Service:     "mcp-ultra-wasm",
			Component:   "task_processor",

			Target:            95.0, // 95% of target throughput
			WarningThreshold:  90.0,
			CriticalThreshold: 80.0,

			Query: `(
				sum(rate(task_processed_total{job="mcp-ultra-wasm"}[5m])) * 60 >= 950
			) * 100`, // At least 950 tasks/minute (95% of 1000)

			EvaluationWindow: 5 * time.Minute,
			ComplianceWindow: 24 * time.Hour, // Daily window for throughput

			AlertingRules: []AlertRule{
				{
					Name:       "SLOThroughputLow",
					Expression: "sum(rate(task_processed_total{job=\"mcp-ultra-wasm\"}[5m])) * 60 < 800",
					For:        5 * time.Minute,
					Severity:   "warning",
					Labels: map[string]string{
						"service":   "mcp-ultra-wasm",
						"component": "task_processor",
						"slo_type":  "throughput",
					},
					Annotations: map[string]string{
						"summary":     "Task processing throughput is low",
						"description": "Task processing rate is below 800 tasks/minute",
						"runbook":     "https://runbooks.mcp-ultra-wasm.com/slo/throughput",
					},
					Enabled: true,
				},
			},

			Tags: map[string]string{
				"team":        "platform",
				"tier":        "2",
				"criticality": "medium",
			},

			Enabled: true,
		},

		// Compliance Data Processing SLO - 99.9% accuracy
		{
			Name:        "compliance_accuracy",
			Description: "Compliance data processing accuracy - percentage of correctly processed compliance events",
			Type:        TypeAccuracy,
			Service:     "mcp-ultra-wasm",
			Component:   "compliance",

			Target:            99.9,
			WarningThreshold:  99.5,
			CriticalThreshold: 99.0,

			Query: `(
				sum(rate(compliance_events_processed_total{job="mcp-ultra-wasm", status="success"}[5m])) /
				sum(rate(compliance_events_processed_total{job="mcp-ultra-wasm"}[5m]))
			) * 100`,

			EvaluationWindow: 5 * time.Minute,
			ComplianceWindow: 30 * 24 * time.Hour,

			AlertingRules: []AlertRule{
				{
					Name:       "SLOComplianceAccuracyDegraded",
					Expression: "mcp_ultra_wasm_slo_current_value{slo=\"compliance_accuracy\"} < 99.5",
					For:        5 * time.Minute,
					Severity:   "critical",
					Labels: map[string]string{
						"service":   "mcp-ultra-wasm",
						"component": "compliance",
						"slo_type":  "accuracy",
					},
					Annotations: map[string]string{
						"summary":     "Compliance processing accuracy is degraded",
						"description": "Compliance data processing accuracy is below threshold",
						"runbook":     "https://runbooks.mcp-ultra-wasm.com/slo/compliance",
					},
					Enabled: true,
				},
			},

			Tags: map[string]string{
				"team":        "compliance",
				"tier":        "1",
				"criticality": "critical",
			},

			Enabled: true,
		},
	}
}

// GetSLOsByService returns SLOs filtered by service name
func GetSLOsByService(slos []*SLO, service string) []*SLO {
	var filtered []*SLO
	for _, slo := range slos {
		if slo.Service == service {
			filtered = append(filtered, slo)
		}
	}
	return filtered
}

// GetSLOsByComponent returns SLOs filtered by component name
func GetSLOsByComponent(slos []*SLO, component string) []*SLO {
	var filtered []*SLO
	for _, slo := range slos {
		if slo.Component == component {
			filtered = append(filtered, slo)
		}
	}
	return filtered
}

// GetSLOsByType returns SLOs filtered by type
func GetSLOsByType(slos []*SLO, sloType Type) []*SLO {
	var filtered []*SLO
	for _, slo := range slos {
		if slo.Type == sloType {
			filtered = append(filtered, slo)
		}
	}
	return filtered
}

// GetCriticalSLOs returns SLOs marked as critical
func GetCriticalSLOs(slos []*SLO) []*SLO {
	var critical []*SLO
	for _, slo := range slos {
		if criticality, exists := slo.Tags["criticality"]; exists && criticality == "critical" {
			critical = append(critical, slo)
		}
	}
	return critical
}
