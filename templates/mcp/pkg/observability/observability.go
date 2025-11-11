// Package observability provides a facade for OpenTelemetry tracing and metrics.
// This package encapsulates otel to prevent direct dependencies throughout
// the codebase.
package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// Tracer is an interface for creating spans.
type Tracer interface {
	// Start creates a span and a context containing the span.
	Start(ctx context.Context, spanName string, opts ...SpanStartOption) (context.Context, Span)
}

// Span is an interface for span operations.
type Span interface {
	// End completes the span.
	End(opts ...SpanEndOption)

	// AddEvent adds an event to the span.
	AddEvent(name string, opts ...EventOption)

	// SetAttributes sets attributes on the span.
	SetAttributes(attrs ...Attribute)

	// SetStatus sets the status of the span.
	SetStatus(code StatusCode, description string)

	// SetName sets the name of the span.
	SetName(name string)

	// RecordError records an error as a span event.
	RecordError(err error, opts ...EventOption)

	// SpanContext returns the SpanContext of the span.
	SpanContext() SpanContext

	// IsRecording returns true if the span is recording.
	IsRecording() bool
}

// SpanContext holds the identifying trace information.
type SpanContext struct {
	traceID    string
	spanID     string
	traceFlags string
}

// TraceID returns the trace ID as a string.
func (sc SpanContext) TraceID() string {
	return sc.traceID
}

// SpanID returns the span ID as a string.
func (sc SpanContext) SpanID() string {
	return sc.spanID
}

// StatusCode represents the status of a span.
type StatusCode int

const (
	// StatusUnset is the default status.
	StatusUnset StatusCode = iota
	// StatusError indicates the operation contains an error.
	StatusError
	// StatusOK indicates the operation completed successfully.
	StatusOK
)

// Attribute is a key-value pair for span attributes.
type Attribute = attribute.KeyValue

// SpanStartOption configures a Span start.
type SpanStartOption = trace.SpanStartOption

// SpanEndOption configures a Span end.
type SpanEndOption = trace.SpanEndOption

// EventOption configures an Event.
type EventOption = trace.EventOption

// tracer wraps an otel tracer.
type tracer struct {
	trace.Tracer
}

// span wraps an otel span.
type span struct {
	trace.Span
}

// Start creates a new span.
func (t *tracer) Start(ctx context.Context, spanName string, opts ...SpanStartOption) (context.Context, Span) {
	ctx, s := t.Tracer.Start(ctx, spanName, opts...)
	return ctx, &span{s}
}

// End completes the span.
func (s *span) End(opts ...SpanEndOption) {
	s.Span.End(opts...)
}

// AddEvent adds an event to the span.
func (s *span) AddEvent(name string, opts ...EventOption) {
	s.Span.AddEvent(name, opts...)
}

// SetAttributes sets attributes on the span.
func (s *span) SetAttributes(attrs ...Attribute) {
	s.Span.SetAttributes(attrs...)
}

// SetStatus sets the status of the span.
func (s *span) SetStatus(code StatusCode, description string) {
	var otelCode codes.Code
	switch code {
	case StatusError:
		otelCode = codes.Error
	case StatusOK:
		otelCode = codes.Ok
	default:
		otelCode = codes.Unset
	}
	s.Span.SetStatus(otelCode, description)
}

// SetName sets the name of the span.
func (s *span) SetName(name string) {
	s.Span.SetName(name)
}

// RecordError records an error as a span event.
func (s *span) RecordError(err error, opts ...EventOption) {
	s.Span.RecordError(err, opts...)
}

// SpanContext returns the SpanContext of the span.
func (s *span) SpanContext() SpanContext {
	sc := s.Span.SpanContext()
	return SpanContext{
		traceID:    sc.TraceID().String(),
		spanID:     sc.SpanID().String(),
		traceFlags: sc.TraceFlags().String(),
	}
}

// IsRecording returns true if the span is recording.
func (s *span) IsRecording() bool {
	return s.Span.IsRecording()
}

// GetTracer returns a named tracer.
func GetTracer(name string, opts ...trace.TracerOption) Tracer {
	return &tracer{otel.Tracer(name, opts...)}
}

// NewNoopTracer returns a tracer that does nothing.
func NewNoopTracer() Tracer {
	return &tracer{noop.NewTracerProvider().Tracer("")}
}

// Attribute constructors

// String creates a string attribute.
func String(key, value string) Attribute {
	return attribute.String(key, value)
}

// Int creates an int attribute.
func Int(key string, value int) Attribute {
	return attribute.Int(key, value)
}

// Int64 creates an int64 attribute.
func Int64(key string, value int64) Attribute {
	return attribute.Int64(key, value)
}

// Float64 creates a float64 attribute.
func Float64(key string, value float64) Attribute {
	return attribute.Float64(key, value)
}

// Bool creates a bool attribute.
func Bool(key string, value bool) Attribute {
	return attribute.Bool(key, value)
}

// StringSlice creates a string slice attribute.
func StringSlice(key string, value []string) Attribute {
	return attribute.StringSlice(key, value)
}

// IntSlice creates an int slice attribute.
func IntSlice(key string, value []int) Attribute {
	return attribute.IntSlice(key, value)
}

// Int64Slice creates an int64 slice attribute.
func Int64Slice(key string, value []int64) Attribute {
	return attribute.Int64Slice(key, value)
}

// Float64Slice creates a float64 slice attribute.
func Float64Slice(key string, value []float64) Attribute {
	return attribute.Float64Slice(key, value)
}

// BoolSlice creates a bool slice attribute.
func BoolSlice(key string, value []bool) Attribute {
	return attribute.BoolSlice(key, value)
}

// SpanKind options

// WithSpanKind configures the span kind.
func WithSpanKind(kind trace.SpanKind) SpanStartOption {
	return trace.WithSpanKind(kind)
}

// SpanKind constants
const (
	SpanKindInternal = trace.SpanKindInternal
	SpanKindServer   = trace.SpanKindServer
	SpanKindClient   = trace.SpanKindClient
	SpanKindProducer = trace.SpanKindProducer
	SpanKindConsumer = trace.SpanKindConsumer
)
