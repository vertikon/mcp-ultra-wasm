// Package metrics provides a facade for Prometheus metrics.
// This package encapsulates prometheus client to prevent direct dependencies
// and provides a cleaner API.
package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Counter is an interface for counter metrics.
type Counter interface {
	// Inc increments the counter by 1.
	Inc()
	// Add adds the given value to the counter.
	Add(float64)
}

// Gauge is an interface for gauge metrics.
type Gauge interface {
	// Set sets the gauge to the given value.
	Set(float64)
	// Inc increments the gauge by 1.
	Inc()
	// Dec decrements the gauge by 1.
	Dec()
	// Add adds the given value to the gauge.
	Add(float64)
	// Sub subtracts the given value from the gauge.
	Sub(float64)
}

// Histogram is an interface for histogram metrics.
type Histogram interface {
	// Observe adds a single observation to the histogram.
	Observe(float64)
}

// CounterVec is an interface for counter vector metrics.
type CounterVec interface {
	// WithLabelValues returns a Counter for the given label values.
	WithLabelValues(lvs ...string) Counter
}

// GaugeVec is an interface for gauge vector metrics.
type GaugeVec interface {
	// WithLabelValues returns a Gauge for the given label values.
	WithLabelValues(lvs ...string) Gauge
}

// HistogramVec is an interface for histogram vector metrics.
type HistogramVec interface {
	// WithLabelValues returns a Histogram for the given label values.
	WithLabelValues(lvs ...string) Histogram
}

// Internal implementations wrapping prometheus types

type counter struct{ prometheus.Counter }

func (c *counter) Inc()          { c.Counter.Inc() }
func (c *counter) Add(v float64) { c.Counter.Add(v) }

type gauge struct{ prometheus.Gauge }

func (g *gauge) Set(v float64) { g.Gauge.Set(v) }
func (g *gauge) Inc()          { g.Gauge.Inc() }
func (g *gauge) Dec()          { g.Gauge.Dec() }
func (g *gauge) Add(v float64) { g.Gauge.Add(v) }
func (g *gauge) Sub(v float64) { g.Gauge.Sub(v) }

type histogram struct{ prometheus.Histogram }

func (h *histogram) Observe(v float64) { h.Histogram.Observe(v) }

type counterVec struct{ *prometheus.CounterVec }

func (cv *counterVec) WithLabelValues(lvs ...string) Counter {
	return &counter{cv.CounterVec.WithLabelValues(lvs...)}
}

type gaugeVec struct{ *prometheus.GaugeVec }

func (gv *gaugeVec) WithLabelValues(lvs ...string) Gauge {
	return &gauge{gv.GaugeVec.WithLabelValues(lvs...)}
}

type histogramVec struct{ *prometheus.HistogramVec }

func (hv *histogramVec) WithLabelValues(lvs ...string) Histogram {
	obs := hv.HistogramVec.WithLabelValues(lvs...)
	// Type assertion to get the underlying prometheus.Histogram
	if h, ok := obs.(prometheus.Histogram); ok {
		return &histogram{h}
	}
	// Fallback - this shouldn't happen but prevents panic
	return &histogram{prometheus.NewHistogram(prometheus.HistogramOpts{})}
}

// NewCounter creates a new Counter metric.
func NewCounter(name, help string) Counter {
	return &counter{
		promauto.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		}),
	}
}

// NewCounterVec creates a new CounterVec metric with the given label names.
func NewCounterVec(name, help string, labelNames []string) CounterVec {
	return &counterVec{
		promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: name,
				Help: help,
			},
			labelNames,
		),
	}
}

// NewGauge creates a new Gauge metric.
func NewGauge(name, help string) Gauge {
	return &gauge{
		promauto.NewGauge(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		}),
	}
}

// NewGaugeVec creates a new GaugeVec metric with the given label names.
func NewGaugeVec(name, help string, labelNames []string) GaugeVec {
	return &gaugeVec{
		promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: name,
				Help: help,
			},
			labelNames,
		),
	}
}

// NewHistogram creates a new Histogram metric with default buckets.
func NewHistogram(name, help string) Histogram {
	return &histogram{
		promauto.NewHistogram(prometheus.HistogramOpts{
			Name: name,
			Help: help,
		}),
	}
}

// NewHistogramWithBuckets creates a new Histogram metric with custom buckets.
func NewHistogramWithBuckets(name, help string, buckets []float64) Histogram {
	return &histogram{
		promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		}),
	}
}

// NewHistogramVec creates a new HistogramVec metric with the given label names.
func NewHistogramVec(name, help string, labelNames []string) HistogramVec {
	return &histogramVec{
		promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: name,
				Help: help,
			},
			labelNames,
		),
	}
}

// NewHistogramVecWithBuckets creates a new HistogramVec with custom buckets.
func NewHistogramVecWithBuckets(name, help string, labelNames []string, buckets []float64) HistogramVec {
	return &histogramVec{
		promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    name,
				Help:    help,
				Buckets: buckets,
			},
			labelNames,
		),
	}
}

// Handler returns an HTTP handler for the Prometheus metrics endpoint.
func Handler() http.Handler {
	return promhttp.Handler()
}

// DefaultBuckets returns the default histogram buckets.
func DefaultBuckets() []float64 {
	return prometheus.DefBuckets
}

// LinearBuckets creates 'count' buckets, each 'width' wide, starting at 'start'.
func LinearBuckets(start, width float64, count int) []float64 {
	return prometheus.LinearBuckets(start, width, count)
}

// ExponentialBuckets creates 'count' buckets, where the lowest bucket has an
// upper bound of 'start' and each following bucket's upper bound is 'factor'
// times the previous bucket's upper bound.
func ExponentialBuckets(start, factor float64, count int) []float64 {
	return prometheus.ExponentialBuckets(start, factor, count)
}
