package observability

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// HTTPTelemetryMiddleware provides HTTP request instrumentation
func (ts *TelemetryService) HTTPTelemetryMiddleware(next http.Handler) http.Handler {
	if !ts.config.Enabled {
		return next
	}

	// Use OpenTelemetry HTTP instrumentation
	return otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Increment active connections
			ts.IncrementActiveConnections()
			defer ts.DecrementActiveConnections()

			// Create custom response writer to capture status code
			rw := &middlewareResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			// Get trace span from context
			span := trace.SpanFromContext(r.Context())

			// Add request attributes to span
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.scheme", r.URL.Scheme),
				attribute.String("http.host", r.Host),
				attribute.String("http.user_agent", r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
				attribute.Int64("http.request_content_length", r.ContentLength),
			)

			// Add custom business attributes
			if userID := r.Header.Get("X-User-ID"); userID != "" {
				span.SetAttributes(attribute.String("user.id", userID))
			}
			if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
				span.SetAttributes(attribute.String("tenant.id", tenantID))
			}
			if traceID := r.Header.Get("X-Trace-ID"); traceID != "" {
				span.SetAttributes(attribute.String("trace.id", traceID))
			}

			// Call next handler
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			statusCode := rw.statusCode
			statusStr := strconv.Itoa(statusCode)

			// Add response attributes to span
			span.SetAttributes(
				attribute.Int("http.status_code", statusCode),
				attribute.Int64("http.response_content_length", rw.bytesWritten),
				attribute.Float64("http.duration_ms", float64(duration.Nanoseconds())/1000000),
			)

			// Set span status based on HTTP status code
			if statusCode >= 400 {
				span.SetStatus(codes.Error, http.StatusText(statusCode))
			} else {
				span.SetStatus(codes.Ok, "")
			}

			// Record metrics
			ts.RecordHTTPRequest(r.Method, r.URL.Path, statusStr, duration)

			// Log request (structured logging)
			fields := []zap.Field{
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", statusCode),
				zap.Duration("duration", duration),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			}

			if userID := r.Header.Get("X-User-ID"); userID != "" {
				fields = append(fields, zap.String("user_id", userID))
			}
			if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
				fields = append(fields, zap.String("tenant_id", tenantID))
			}

			if statusCode >= 400 {
				ts.logger.Warn("HTTP request completed with error", fields...)
			} else if ts.config.Debug {
				ts.logger.Debug("HTTP request completed", fields...)
			}
		}),
		ts.config.ServiceName,
		otelhttp.WithTracerProvider(ts.tracerProvider),
		otelhttp.WithMeterProvider(ts.meterProvider),
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)
}

// middlewareResponseWriter wraps http.ResponseWriter to capture response data
type middlewareResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (rw *middlewareResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *middlewareResponseWriter) Write(data []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(data)
	rw.bytesWritten += int64(n)
	return n, err
}

// DatabaseTelemetryWrapper provides database operation instrumentation
type DatabaseTelemetryWrapper struct {
	telemetry *TelemetryService
}

// NewDatabaseTelemetryWrapper creates a new database telemetry wrapper
func NewDatabaseTelemetryWrapper(telemetry *TelemetryService) *DatabaseTelemetryWrapper {
	return &DatabaseTelemetryWrapper{
		telemetry: telemetry,
	}
}

// WrapDatabaseOperation wraps a database operation with telemetry
func (dtw *DatabaseTelemetryWrapper) WrapDatabaseOperation(
	ctx context.Context,
	operation string,
	table string,
	query string,
	fn func(context.Context) error,
) error {
	if !dtw.telemetry.config.Enabled {
		return fn(ctx)
	}

	spanName := fmt.Sprintf("db.%s.%s", operation, table)
	ctx, span := dtw.telemetry.StartSpan(ctx, spanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", operation),
			attribute.String("db.sql.table", table),
			attribute.String("db.statement", query),
		),
	)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	// Add timing attribute
	span.SetAttributes(attribute.Float64("db.duration_ms", float64(duration.Nanoseconds())/1000000))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		dtw.telemetry.RecordError("database_error", "database")
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}

// instrumentedOperation is a generic helper for wrapping operations with telemetry
func (ts *TelemetryService) instrumentedOperation(
	ctx context.Context,
	spanName string,
	spanKind trace.SpanKind,
	attrs []attribute.KeyValue,
	errorCategory string,
	fn func(context.Context) error,
) error {
	if !ts.config.Enabled {
		return fn(ctx)
	}

	ctx, span := ts.StartSpan(ctx, spanName,
		trace.WithSpanKind(spanKind),
		trace.WithAttributes(attrs...),
	)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	span.SetAttributes(attribute.Float64(fmt.Sprintf("%s.duration_ms", errorCategory), float64(duration.Nanoseconds())/1000000))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ts.RecordError(fmt.Sprintf("%s_error", errorCategory), errorCategory)
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}

// CacheOperation wrapper for cache operations
func (ts *TelemetryService) CacheOperation(
	ctx context.Context,
	operation string,
	key string,
	fn func(context.Context) error,
) error {
	return ts.instrumentedOperation(
		ctx,
		fmt.Sprintf("cache.%s", operation),
		trace.SpanKindClient,
		[]attribute.KeyValue{
			attribute.String("cache.system", "redis"),
			attribute.String("cache.operation", operation),
			attribute.String("cache.key", key),
		},
		"cache",
		fn,
	)
}

// MessageQueueOperation wrapper for message queue operations
func (ts *TelemetryService) MessageQueueOperation(
	ctx context.Context,
	operation string,
	subject string,
	fn func(context.Context) error,
) error {
	return ts.instrumentedOperation(
		ctx,
		fmt.Sprintf("messaging.%s", operation),
		trace.SpanKindProducer,
		[]attribute.KeyValue{
			attribute.String("messaging.system", "nats"),
			attribute.String("messaging.operation", operation),
			attribute.String("messaging.destination", subject),
		},
		"messaging",
		fn,
	)
}

// BusinessOperation wrapper for general business operations
func (ts *TelemetryService) BusinessOperation(
	ctx context.Context,
	operationName string,
	attributes []attribute.KeyValue,
	fn func(context.Context) error,
) error {
	if !ts.config.Enabled {
		return fn(ctx)
	}

	ctx, span := ts.StartSpan(ctx, operationName,
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithAttributes(attributes...),
	)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	span.SetAttributes(attribute.Float64("operation.duration_ms", float64(duration.Nanoseconds())/1000000))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		ts.RecordError("business_logic_error", "business")
	} else {
		span.SetStatus(codes.Ok, "")
	}

	return err
}

// AddSpanEvent adds an event to the current span
func (ts *TelemetryService) AddSpanEvent(ctx context.Context, name string, attributes ...attribute.KeyValue) {
	if !ts.config.Enabled {
		return
	}

	span := trace.SpanFromContext(ctx)
	span.AddEvent(name, trace.WithAttributes(attributes...))
}

// LogEvent logs a structured event with tracing context
func (ts *TelemetryService) LogEvent(ctx context.Context, level string, message string, fields ...zap.Field) {
	// Add trace context to log fields
	span := trace.SpanFromContext(ctx)
	spanContext := span.SpanContext()

	if spanContext.IsValid() {
		fields = append(fields,
			zap.String("trace_id", spanContext.TraceID().String()),
			zap.String("span_id", spanContext.SpanID().String()),
		)
	}

	switch level {
	case "debug":
		ts.logger.Debug(message, fields...)
	case "info":
		ts.logger.Info(message, fields...)
	case "warn":
		ts.logger.Warn(message, fields...)
	case "error":
		ts.logger.Error(message, fields...)
		// Also add span event for errors
		if ts.config.Enabled {
			span.AddEvent("error", trace.WithAttributes(
				attribute.String("error.message", message),
			))
		}
	}
}
