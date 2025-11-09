package observability

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// -------------------- COMPAT LAYER (sem ctx) --------------------
// Mantém compatibilidade com código legado que não passa context

func (ts *TelemetryService) RecordCounter(name string, value float64, labels map[string]string) {
	ts.RecordCounterWithContext(context.Background(), name, value, labels)
}

func (ts *TelemetryService) RecordGauge(name string, value float64, labels map[string]string) {
	ts.RecordGaugeWithContext(context.Background(), name, value, labels)
}

func (ts *TelemetryService) RecordHistogram(name string, value float64, labels map[string]string) {
	ts.RecordHistogramWithContext(context.Background(), name, value, labels)
}

// --------------- MODERNA (com ctx) — use quando puder ---------------

// RecordCounterWithContext increments a counter metric with context propagation
func (ts *TelemetryService) RecordCounterWithContext(ctx context.Context, name string, value float64, labels map[string]string) {
	if ts == nil || ts.meter == nil {
		return
	}

	counter, err := ts.meter.Int64Counter(name)
	if err != nil {
		if ts.logger != nil {
			ts.logger.Warn("failed to create counter",
				zap.String("metric", name),
				zap.Error(err))
		}
		return
	}

	attrs := labelsToAttributes(labels)
	counter.Add(ctx, int64(value), metric.WithAttributes(attrs...))
}

// RecordGaugeWithContext sets a gauge metric with context propagation
func (ts *TelemetryService) RecordGaugeWithContext(ctx context.Context, name string, value float64, labels map[string]string) {
	if ts == nil || ts.meter == nil {
		return
	}

	// Note: OpenTelemetry doesn't have simple Gauge, we use UpDownCounter as approximation
	gauge, err := ts.meter.Int64UpDownCounter(name)
	if err != nil {
		if ts.logger != nil {
			ts.logger.Warn("failed to create gauge",
				zap.String("metric", name),
				zap.Error(err))
		}
		return
	}

	attrs := labelsToAttributes(labels)
	gauge.Add(ctx, int64(value), metric.WithAttributes(attrs...))
}

// RecordHistogramWithContext records a histogram observation with context propagation
func (ts *TelemetryService) RecordHistogramWithContext(ctx context.Context, name string, value float64, labels map[string]string) {
	if ts == nil || ts.meter == nil {
		return
	}

	histogram, err := ts.meter.Float64Histogram(name)
	if err != nil {
		if ts.logger != nil {
			ts.logger.Warn("failed to create histogram",
				zap.String("metric", name),
				zap.Error(err))
		}
		return
	}

	attrs := labelsToAttributes(labels)
	histogram.Record(ctx, value, metric.WithAttributes(attrs...))
}

// labelsToAttributes converts map[string]string to []attribute.KeyValue
func labelsToAttributes(labels map[string]string) []attribute.KeyValue {
	if len(labels) == 0 {
		return nil
	}

	attrs := make([]attribute.KeyValue, 0, len(labels))
	for k, v := range labels {
		attrs = append(attrs, attribute.String(k, v))
	}
	return attrs
}
