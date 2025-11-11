// Package logger provides a compatibility wrapper around zap.Logger
// to maintain backward compatibility with the old logger API.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger and provides a backward-compatible API
type Logger struct {
	*zap.Logger
}

// NewLogger creates a new production logger
func NewLogger() (*Logger, error) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return &Logger{Logger: zapLogger}, nil
}

// NewDevelopment creates a new development logger
func NewDevelopment() (*Logger, error) {
	zapLogger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return &Logger{Logger: zapLogger}, nil
}

// FromZap wraps an existing zap.Logger
func FromZap(zapLogger *zap.Logger) *Logger {
	return &Logger{Logger: zapLogger}
}

// Info logs with key-value pairs (backward compatible API)
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Info(msg, convertToZapFields(keysAndValues)...)
}

// Error logs with key-value pairs (backward compatible API)
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.Logger.Error(msg, convertToZapFields(keysAndValues)...)
}

// Debug logs with key-value pairs (backward compatible API)
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.Logger.Debug(msg, convertToZapFields(keysAndValues)...)
}

// Warn logs with key-value pairs (backward compatible API)
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.Logger.Warn(msg, convertToZapFields(keysAndValues)...)
}

// Fatal logs with key-value pairs and exits (backward compatible API)
func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.Logger.Fatal(msg, convertToZapFields(keysAndValues)...)
}

// convertToZapFields converts alternating key-value pairs to zap.Field slice
func convertToZapFields(keysAndValues []interface{}) []zap.Field {
	if len(keysAndValues) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 >= len(keysAndValues) {
			// Odd number of arguments, skip the last one
			break
		}

		key, ok := keysAndValues[i].(string)
		if !ok {
			// Key is not a string, skip this pair
			continue
		}

		value := keysAndValues[i+1]
		fields = append(fields, anyToField(key, value))
	}

	return fields
}

// anyToField converts any value to a zap.Field
func anyToField(key string, value interface{}) zap.Field {
	switch v := value.(type) {
	case string:
		return zap.String(key, v)
	case int:
		return zap.Int(key, v)
	case int64:
		return zap.Int64(key, v)
	case uint:
		return zap.Uint(key, v)
	case uint64:
		return zap.Uint64(key, v)
	case float64:
		return zap.Float64(key, v)
	case bool:
		return zap.Bool(key, v)
	case error:
		return zap.Error(v)
	case zapcore.Field:
		return v
	default:
		return zap.Any(key, v)
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}
