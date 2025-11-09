package observability

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger cria um novo logger estruturado
func NewLogger(level string) (*zap.Logger, error) {
	// Configurar nível de log
	logLevel := zapcore.InfoLevel
	switch level {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	}

	// Configurar encoder
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Verificar se estamos em ambiente de desenvolvimento
	if os.Getenv("ENV") == "development" || level == "debug" {
		// Encoder mais legível para desenvolvimento
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		// Encoder JSON para produção
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Configurar core
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		logLevel,
	)

	// Adicionar caller e stack trace para debug e error
	if level == "debug" || level == "error" {
		core = zapcore.NewTee(
			core,
			zapcore.NewCore(
				zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
				zapcore.AddSync(os.Stderr),
				zapcore.ErrorLevel,
			),
		)
	}

	// Criar logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

// NewLoggerWithConfig cria logger com configuração customizada
func NewLoggerWithConfig(config *LoggerConfig) (*zap.Logger, error) {
	if config == nil {
		return NewLogger("info")
	}

	// Configurar nível de log
	logLevel := zapcore.InfoLevel
	switch config.Level {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	}

	// Configurar encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if config.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var encoder zapcore.Encoder
	if config.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Configurar output
	var writeSyncer zapcore.WriteSyncer
	if config.OutputPath != "" {
		file, err := os.OpenFile(config.OutputPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		writeSyncer = zapcore.AddSync(file)
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// Configurar core
	core := zapcore.NewCore(
		encoder,
		writeSyncer,
		logLevel,
	)

	// Configurar opções
	var options []zap.Option
	if config.EnableCaller {
		options = append(options, zap.AddCaller())
	}
	if config.EnableStackTrace {
		options = append(options, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Adicionar fields globais
	if len(config.Fields) > 0 {
		options = append(options, zap.Fields(config.Fields))
	}

	return zap.New(core, options...), nil
}

// LoggerConfig configuração do logger
type LoggerConfig struct {
	Level            string      `json:"level"`
	Format           string      `json:"format"` // "json" ou "console"
	OutputPath       string      `json:"output_path"`
	EnableCaller     bool        `json:"enable_caller"`
	EnableStackTrace bool        `json:"enable_stack_trace"`
	Fields           []zap.Field `json:"fields"`
}

// ContextLogger wrapper para adicionar contexto automaticamente
type ContextLogger struct {
	logger *zap.Logger
	fields map[string]interface{}
}

func NewContextLogger(logger *zap.Logger, fields map[string]interface{}) *ContextLogger {
	return &ContextLogger{
		logger: logger,
		fields: fields,
	}
}

func (cl *ContextLogger) Debug(message string, fields ...zap.Field) {
	cl.logger.Debug(message, cl.appendFields(fields)...)
}

func (cl *ContextLogger) Info(message string, fields ...zap.Field) {
	cl.logger.Info(message, cl.appendFields(fields)...)
}

func (cl *ContextLogger) Warn(message string, fields ...zap.Field) {
	cl.logger.Warn(message, cl.appendFields(fields)...)
}

func (cl *ContextLogger) Error(message string, fields ...zap.Field) {
	cl.logger.Error(message, cl.appendFields(fields)...)
}

func (cl *ContextLogger) Fatal(message string, fields ...zap.Field) {
	cl.logger.Fatal(message, cl.appendFields(fields)...)
}

func (cl *ContextLogger) Sync() error {
	return cl.logger.Sync()
}

func (cl *ContextLogger) With(fields ...zap.Field) *ContextLogger {
	newFields := make(map[string]interface{})
	for k, v := range cl.fields {
		newFields[k] = v
	}

	// Adicionar novos fields (simplificado - em produção usaria reflection)
	for _, field := range fields {
		// TODO: Extrair chave e valor do zap.Field
	}

	return NewContextLogger(cl.logger.With(fields...), newFields)
}

func (cl *ContextLogger) appendFields(fields []zap.Field) []zap.Field {
	// Converter fields do mapa para zap.Fields
	zapFields := make([]zap.Field, 0, len(cl.fields)+len(fields))

	for key, value := range cl.fields {
		switch v := value.(type) {
		case string:
			zapFields = append(zapFields, zap.String(key, v))
		case int:
			zapFields = append(zapFields, zap.Int(key, v))
		case int64:
			zapFields = append(zapFields, zap.Int64(key, v))
		case float64:
			zapFields = append(zapFields, zap.Float64(key, v))
		case bool:
			zapFields = append(zapFields, zap.Bool(key, v))
		default:
			zapFields = append(zapFields, zap.Any(key, v))
		}
	}

	zapFields = append(zapFields, fields...)
	return zapFields
}

// LoggingMiddleware middleware para logging de requisições HTTP
func LoggingMiddleware(logger *zap.Logger) func(interface{}) interface{} {
	// Esta função será implementada no handlers
	return func(next interface{}) interface{} {
		return next
	}
}

// RequestLogger logger específico para requisições
type RequestLogger struct {
	logger *zap.Logger
}

func NewRequestLogger(logger *zap.Logger) *RequestLogger {
	return &RequestLogger{
		logger: logger.Named("request"),
	}
}

func (rl *RequestLogger) LogRequest(method, path string, statusCode int, duration int64, requestID string) {
	rl.logger.Info("HTTP Request",
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Int64("duration_ms", duration),
		zap.String("request_id", requestID),
	)
}

func (rl *RequestLogger) LogError(method, path string, err error, requestID string) {
	rl.logger.Error("HTTP Request Error",
		zap.String("method", method),
		zap.String("path", path),
		zap.Error(err),
		zap.String("request_id", requestID),
	)
}

// WASMLogger logger específico para operações WASM
type WASMLogger struct {
	logger *zap.Logger
}

func NewWASMLogger(logger *zap.Logger) *WASMLogger {
	return &WASMLogger{
		logger: logger.Named("wasm"),
	}
}

func (wl *WASMLogger) LogExecution(function string, duration int64, success bool, requestID string) {
	fields := []zap.Field{
		zap.String("function", function),
		zap.Int64("duration_ms", duration),
		zap.Bool("success", success),
		zap.String("request_id", requestID),
	}

	if success {
		wl.logger.Info("WASM Execution Success", fields...)
	} else {
		wl.logger.Error("WASM Execution Failed", fields...)
	}
}

func (wl *WASMLogger) LogLoad(moduleName string, size int64, loadTime int64) {
	wl.logger.Info("WASM Module Loaded",
		zap.String("module", moduleName),
		zap.Int64("size_bytes", size),
		zap.Int64("load_time_ms", loadTime),
	)
}

func (wl *WASMLogger) LogError(function string, err error, requestID string) {
	wl.logger.Error("WASM Execution Error",
		zap.String("function", function),
		zap.Error(err),
		zap.String("request_id", requestID),
	)
}

// SDKLogger logger específico para operações SDK
type SDKLogger struct {
	logger *zap.Logger
}

func NewSDKLogger(logger *zap.Logger) *SDKLogger {
	return &SDKLogger{
		logger: logger.Named("sdk"),
	}
}

func (sl *SDKLogger) LogPluginExecution(plugin, method string, duration int64, success bool, requestID string) {
	fields := []zap.Field{
		zap.String("plugin", plugin),
		zap.String("method", method),
		zap.Int64("duration_ms", duration),
		zap.Bool("success", success),
		zap.String("request_id", requestID),
	}

	if success {
		sl.logger.Info("SDK Plugin Execution Success", fields...)
	} else {
		sl.logger.Error("SDK Plugin Execution Failed", fields...)
	}
}

func (sl *SDKLogger) LogPluginRegistration(plugin, version string, success bool) {
	fields := []zap.Field{
		zap.String("plugin", plugin),
		zap.String("version", version),
		zap.Bool("success", success),
	}

	if success {
		sl.logger.Info("SDK Plugin Registration Success", fields...)
	} else {
		sl.logger.Error("SDK Plugin Registration Failed", fields...)
	}
}

func (sl *SDKLogger) LogConnection(address string, success bool, err error) {
	fields := []zap.Field{
		zap.String("address", address),
		zap.Bool("success", success),
	}

	if success {
		sl.logger.Info("SDK Connection Success", fields...)
	} else {
		fields = append(fields, zap.Error(err))
		sl.logger.Error("SDK Connection Failed", fields...)
	}
}
