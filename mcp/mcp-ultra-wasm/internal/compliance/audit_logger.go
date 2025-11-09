package compliance

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// AuditLogger handles compliance audit logging
type AuditLogger struct {
	config        AuditLoggingConfig
	logger        *zap.Logger
	auditLogger   *zap.Logger
	encryptionKey []byte
}

// AuditEvent represents an audit event
type AuditEvent struct {
	ID              string                 `json:"id"`
	Timestamp       time.Time              `json:"timestamp"`
	EventType       AuditEventType         `json:"event_type"`
	SubjectID       string                 `json:"subject_id"`
	UserID          string                 `json:"user_id,omitempty"`
	SessionID       string                 `json:"session_id,omitempty"`
	IPAddress       string                 `json:"ip_address,omitempty"`
	UserAgent       string                 `json:"user_agent,omitempty"`
	Purpose         string                 `json:"purpose,omitempty"`
	LegalBasis      string                 `json:"legal_basis,omitempty"`
	DataCategories  []string               `json:"data_categories,omitempty"`
	ProcessingType  string                 `json:"processing_type,omitempty"`
	Result          AuditResult            `json:"result"`
	Details         map[string]interface{} `json:"details,omitempty"`
	DataHash        string                 `json:"data_hash,omitempty"`
	ComplianceFlags []string               `json:"compliance_flags,omitempty"`
	Encrypted       bool                   `json:"encrypted"`
	Version         string                 `json:"version"`
	Service         string                 `json:"service"`
}

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	AuditEventDataProcessing   AuditEventType = "data_processing"
	AuditEventConsentGrant     AuditEventType = "consent_grant"
	AuditEventConsentWithdraw  AuditEventType = "consent_withdraw"
	AuditEventDataAccess       AuditEventType = "data_access"
	AuditEventDataExport       AuditEventType = "data_export"
	AuditEventDataDelete       AuditEventType = "data_delete"
	AuditEventDataRectify      AuditEventType = "data_rectify"
	AuditEventRightsRequest    AuditEventType = "rights_request"
	AuditEventPIIDetection     AuditEventType = "pii_detection"
	AuditEventAnonymization    AuditEventType = "anonymization"
	AuditEventRetentionPolicy  AuditEventType = "retention_policy"
	AuditEventSecurityIncident AuditEventType = "security_incident"
	AuditEventComplianceCheck  AuditEventType = "compliance_check"
)

// AuditResult represents the result of an audited operation
type AuditResult string

const (
	AuditResultSuccess        AuditResult = "success"
	AuditResultFailure        AuditResult = "failure"
	AuditResultPartialSuccess AuditResult = "partial_success"
	AuditResultBlocked        AuditResult = "blocked"
	AuditResultSkipped        AuditResult = "skipped"
)

// NewAuditLogger creates a new audit logger
func NewAuditLogger(config AuditLoggingConfig, logger *zap.Logger) (*AuditLogger, error) {
	if !config.Enabled {
		return &AuditLogger{
			config: config,
			logger: logger,
		}, nil
	}

	// Create dedicated audit logger with structured output
	auditConfig := zap.NewProductionConfig()
	auditConfig.OutputPaths = []string{"stdout"}

	if config.ExternalLogging && config.ExternalEndpoint != "" {
		// In a real implementation, you would configure external logging here
		// For now, we'll just log to stdout with a special format
		auditConfig.OutputPaths = append(auditConfig.OutputPaths, config.ExternalEndpoint)
	}

	// Custom encoder for audit logs
	auditConfig.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	auditLogger, err := auditConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create audit logger: %w", err)
	}

	// Generate encryption key if encryption is enabled
	var encryptionKey []byte
	if config.EncryptionEnabled {
		// Get encryption key from environment variable or secure key management system
		keyStr := os.Getenv("AUDIT_ENCRYPTION_KEY")
		if keyStr == "" {
			return nil, fmt.Errorf("AUDIT_ENCRYPTION_KEY environment variable must be set when encryption is enabled")
		}

		// Decode key from base64 or hex
		if len(keyStr) == 64 { // Hex encoded 32 bytes
			encryptionKey, err = hex.DecodeString(keyStr)
			if err != nil {
				return nil, fmt.Errorf("failed to decode hex encryption key: %w", err)
			}
		} else {
			// Try base64
			encryptionKey, err = base64.StdEncoding.DecodeString(keyStr)
			if err != nil {
				return nil, fmt.Errorf("failed to decode base64 encryption key: %w", err)
			}
		}

		// Validate key length for AES-256
		if len(encryptionKey) != 32 {
			return nil, fmt.Errorf("encryption key must be 32 bytes for AES-256, got %d bytes", len(encryptionKey))
		}
	}

	return &AuditLogger{
		config:        config,
		logger:        logger,
		auditLogger:   auditLogger.Named("compliance-audit"),
		encryptionKey: encryptionKey,
	}, nil
}

// LogDataProcessing logs data processing activities
func (al *AuditLogger) LogDataProcessing(ctx context.Context, subjectID, purpose, result string, data map[string]interface{}) error {
	if !al.config.Enabled {
		return nil
	}

	event := AuditEvent{
		ID:             al.generateEventID(),
		Timestamp:      time.Now(),
		EventType:      AuditEventDataProcessing,
		SubjectID:      subjectID,
		Purpose:        purpose,
		ProcessingType: "automated",
		Result:         AuditResult(result),
		Service:        "mcp-ultra-wasm",
		Version:        "1.0.0",
	}

	// Extract context information
	al.extractContextInfo(ctx, &event)

	// Add data hash for integrity
	if data != nil {
		event.DataHash = al.hashData(data)
		if al.config.DetailLevel == "full" {
			event.Details = al.sanitizeData(data)
		}
	}

	return al.logEvent(event)
}

// LogConsentAction logs consent-related actions
func (al *AuditLogger) LogConsentAction(ctx context.Context, subjectID string, consent ConsentRecord, action string) error {
	if !al.config.Enabled {
		return nil
	}

	eventType := AuditEventConsentGrant
	if action == "withdraw" {
		eventType = AuditEventConsentWithdraw
	}

	event := AuditEvent{
		ID:         al.generateEventID(),
		Timestamp:  time.Now(),
		EventType:  eventType,
		SubjectID:  subjectID,
		Purpose:    consent.Purpose,
		LegalBasis: consent.LegalBasis,
		Result:     AuditResultSuccess,
		Details: map[string]interface{}{
			"consent_id":     consent.ID,
			"consent_source": consent.ConsentSource,
			"granted":        consent.Granted,
			"version":        consent.Version,
		},
		Service: "mcp-ultra-wasm",
		Version: "1.0.0",
	}

	al.extractContextInfo(ctx, &event)
	return al.logEvent(event)
}

// LogConsent is a convenience method to log consent actions with minimal parameters
func (al *AuditLogger) LogConsent(ctx context.Context, subjectID string, purposes []string, source, action string) error {
	if !al.config.Enabled {
		return nil
	}

	for _, purpose := range purposes {
		consent := ConsentRecord{
			SubjectID:     subjectID,
			Purpose:       purpose,
			ConsentSource: ConsentSource(source),
			Granted:       action == "granted",
		}
		if err := al.LogConsentAction(ctx, subjectID, consent, action); err != nil {
			return err
		}
	}

	return nil
}

// LogDataRightsRequest logs data subject rights requests
func (al *AuditLogger) LogDataRightsRequest(ctx context.Context, subjectID string, request DataRightRequest) error {
	if !al.config.Enabled {
		return nil
	}

	event := AuditEvent{
		ID:        al.generateEventID(),
		Timestamp: time.Now(),
		EventType: AuditEventRightsRequest,
		SubjectID: subjectID,
		Result:    AuditResult(request.Status),
		Details: map[string]interface{}{
			"request_id":   request.ID,
			"request_type": request.Type,
			"status":       request.Status,
			"requested_at": request.RequestedAt,
		},
		Service: "mcp-ultra-wasm",
		Version: "1.0.0",
	}

	if request.CompletedAt != nil {
		event.Details["completed_at"] = *request.CompletedAt
	}

	al.extractContextInfo(ctx, &event)
	return al.logEvent(event)
}

// LogPIIDetection logs PII detection events
func (al *AuditLogger) LogPIIDetection(ctx context.Context, subjectID string, classifications []PIIClassification) error {
	if !al.config.Enabled || len(classifications) == 0 {
		return nil
	}

	event := AuditEvent{
		ID:        al.generateEventID(),
		Timestamp: time.Now(),
		EventType: AuditEventPIIDetection,
		SubjectID: subjectID,
		Result:    AuditResultSuccess,
		Details: map[string]interface{}{
			"pii_count":       len(classifications),
			"classifications": al.sanitizeClassifications(classifications),
		},
		ComplianceFlags: []string{"pii_detected"},
		Service:         "mcp-ultra-wasm",
		Version:         "1.0.0",
	}

	al.extractContextInfo(ctx, &event)
	return al.logEvent(event)
}

// LogSecurityIncident logs security incidents related to compliance
func (al *AuditLogger) LogSecurityIncident(ctx context.Context, incidentType, description string, severity string) error {
	if !al.config.Enabled {
		return nil
	}

	event := AuditEvent{
		ID:        al.generateEventID(),
		Timestamp: time.Now(),
		EventType: AuditEventSecurityIncident,
		Result:    AuditResultFailure,
		Details: map[string]interface{}{
			"incident_type": incidentType,
			"description":   description,
			"severity":      severity,
		},
		ComplianceFlags: []string{"security_incident"},
		Service:         "mcp-ultra-wasm",
		Version:         "1.0.0",
	}

	al.extractContextInfo(ctx, &event)
	return al.logEvent(event)
}

// LogComplianceCheck logs compliance validation results
func (al *AuditLogger) LogComplianceCheck(ctx context.Context, checkType string, result bool, details map[string]interface{}) error {
	if !al.config.Enabled {
		return nil
	}

	auditResult := AuditResultSuccess
	if !result {
		auditResult = AuditResultFailure
	}

	event := AuditEvent{
		ID:        al.generateEventID(),
		Timestamp: time.Now(),
		EventType: AuditEventComplianceCheck,
		Result:    auditResult,
		Details: map[string]interface{}{
			"check_type": checkType,
			"passed":     result,
		},
		Service: "mcp-ultra-wasm",
		Version: "1.0.0",
	}

	// Merge additional details
	for k, v := range details {
		event.Details[k] = v
	}

	al.extractContextInfo(ctx, &event)
	return al.logEvent(event)
}

// QueryAuditLogs queries audit logs (simplified implementation)
func (al *AuditLogger) QueryAuditLogs(_ context.Context, _ map[string]interface{}, _ int) ([]AuditEvent, error) {
	if !al.config.Enabled {
		return nil, fmt.Errorf("audit logging is disabled")
	}

	// In a real implementation, this would query from a persistent audit store
	// For now, return empty results
	return []AuditEvent{}, nil
}

// logEvent writes an audit event to the audit log
func (al *AuditLogger) logEvent(event AuditEvent) error {
	// Encrypt sensitive details if encryption is enabled
	if al.config.EncryptionEnabled && al.encryptionKey != nil {
		if err := al.encryptSensitiveData(&event); err != nil {
			al.logger.Warn("Failed to encrypt audit event", zap.Error(err))
		} else {
			event.Encrypted = true
		}
	}

	// Convert event to JSON for structured logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Log the event
	switch al.config.DetailLevel {
	case "minimal":
		al.auditLogger.Info("compliance_audit",
			zap.String("event_id", event.ID),
			zap.String("event_type", string(event.EventType)),
			zap.String("subject_id", event.SubjectID),
			zap.String("result", string(event.Result)))
	case "standard":
		al.auditLogger.Info("compliance_audit",
			zap.String("event_id", event.ID),
			zap.Time("timestamp", event.Timestamp),
			zap.String("event_type", string(event.EventType)),
			zap.String("subject_id", event.SubjectID),
			zap.String("purpose", event.Purpose),
			zap.String("result", string(event.Result)))
	case "full":
		al.auditLogger.Info("compliance_audit",
			zap.String("audit_event", string(eventJSON)))
	}

	return nil
}

// Helper methods

func (al *AuditLogger) generateEventID() string {
	return fmt.Sprintf("audit_%d_%x", time.Now().UnixNano(), time.Now().UnixNano()%1000)
}

func (al *AuditLogger) extractContextInfo(ctx context.Context, event *AuditEvent) {
	// Extract information from context
	if userID := ctx.Value("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			event.UserID = uid
		}
	}

	if sessionID := ctx.Value("session_id"); sessionID != nil {
		if sid, ok := sessionID.(string); ok {
			event.SessionID = sid
		}
	}

	if ipAddress := ctx.Value("ip_address"); ipAddress != nil {
		if ip, ok := ipAddress.(string); ok {
			event.IPAddress = ip
		}
	}

	if userAgent := ctx.Value("user_agent"); userAgent != nil {
		if ua, ok := userAgent.(string); ok {
			event.UserAgent = ua
		}
	}
}

func (al *AuditLogger) hashData(data map[string]interface{}) string {
	// Create a deterministic hash of the data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

func (al *AuditLogger) sanitizeData(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range data {
		// Remove or mask sensitive fields
		keyLower := strings.ToLower(key)
		if strings.Contains(keyLower, "password") ||
			strings.Contains(keyLower, "secret") ||
			strings.Contains(keyLower, "token") ||
			strings.Contains(keyLower, "key") {
			sanitized[key] = "[REDACTED]"
		} else {
			sanitized[key] = value
		}
	}

	return sanitized
}

func (al *AuditLogger) sanitizeClassifications(classifications []PIIClassification) []map[string]interface{} {
	sanitized := make([]map[string]interface{}, len(classifications))

	for i, c := range classifications {
		sanitized[i] = map[string]interface{}{
			"field_name":  c.FieldName,
			"pii_type":    c.PIIType,
			"sensitivity": c.Sensitivity,
			"confidence":  c.Confidence,
			"method":      c.Method,
			"timestamp":   c.Timestamp,
		}
	}

	return sanitized
}

func (al *AuditLogger) encryptSensitiveData(event *AuditEvent) error {
	// In a real implementation, you would use proper encryption like AES-GCM
	// This is a simplified example
	if event.Details != nil {
		// Convert details to JSON and encrypt
		detailsJSON, err := json.Marshal(event.Details)
		if err != nil {
			return err
		}

		// Simple XOR encryption for demonstration (use proper encryption in production)
		encrypted := make([]byte, len(detailsJSON))
		for i, b := range detailsJSON {
			encrypted[i] = b ^ al.encryptionKey[i%len(al.encryptionKey)]
		}

		// Store as hex string
		event.Details = map[string]interface{}{
			"encrypted_data": hex.EncodeToString(encrypted),
		}
	}

	return nil
}
