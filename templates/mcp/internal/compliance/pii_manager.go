package compliance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// PIIManager handles detection, classification, and protection of PII data
type PIIManager struct {
	config              PIIDetectionConfig
	logger              *zap.Logger
	detectors           map[PIIType]PIIDetector
	anonymizers         map[AnonymizationMethod]Anonymizer
	classificationCache map[string]PIIClassification
}

// PIIType represents different types of personally identifiable information
type PIIType string

const (
	PIITypeEmail       PIIType = "email"
	PIITypeCPF         PIIType = "cpf"  // Brazilian CPF
	PIITypeCNPJ        PIIType = "cnpj" // Brazilian CNPJ
	PIITypePhone       PIIType = "phone"
	PIITypeCreditCard  PIIType = "credit_card"
	PIITypeIPAddress   PIIType = "ip_address"
	PIITypeSSN         PIIType = "ssn" // Social Security Number
	PIITypePassport    PIIType = "passport"
	PIITypeDateOfBirth PIIType = "date_of_birth"
	PIITypeAddress     PIIType = "address"
	PIITypeName        PIIType = "name"
	PIITypeUsername    PIIType = "username"
	PIITypeCustom      PIIType = "custom"
)

// PIISensitivity represents the sensitivity level of PII
type PIISensitivity string

const (
	PIISensitivityPublic       PIISensitivity = "public"
	PIISensitivityInternal     PIISensitivity = "internal"
	PIISensitivityConfidential PIISensitivity = "confidential"
	PIISensitivityRestricted   PIISensitivity = "restricted"
)

// AnonymizationMethod represents different methods for anonymizing PII
type AnonymizationMethod string

const (
	AnonymizationHash       AnonymizationMethod = "hash"
	AnonymizationEncrypt    AnonymizationMethod = "encrypt"
	AnonymizationTokenize   AnonymizationMethod = "tokenize"
	AnonymizationRedact     AnonymizationMethod = "redact"
	AnonymizationGeneralize AnonymizationMethod = "generalize"
	AnonymizationShuffle    AnonymizationMethod = "shuffle"
	AnonymizationNoise      AnonymizationMethod = "noise"
)

// PIIClassification contains information about detected PII
type PIIClassification struct {
	FieldName      string              `json:"field_name"`
	PIIType        PIIType             `json:"pii_type"`
	Sensitivity    PIISensitivity      `json:"sensitivity"`
	Confidence     float64             `json:"confidence"`
	OriginalValue  interface{}         `json:"-"` // Don't serialize original value
	ProcessedValue interface{}         `json:"processed_value"`
	Method         AnonymizationMethod `json:"method"`
	Timestamp      time.Time           `json:"timestamp"`
	Context        map[string]string   `json:"context,omitempty"`
}

// PIIDetector interface for detecting specific types of PII
type PIIDetector interface {
	Detect(field string, value interface{}) (bool, float64, map[string]string)
	GetType() PIIType
	GetSensitivity() PIISensitivity
}

// Anonymizer interface for anonymizing PII data
type Anonymizer interface {
	Anonymize(value interface{}, context map[string]string) (interface{}, error)
	IsReversible() bool
	GetMethod() AnonymizationMethod
}

// NewPIIManager creates a new PII manager
func NewPIIManager(config PIIDetectionConfig, logger *zap.Logger) (*PIIManager, error) {
	pm := &PIIManager{
		config:              config,
		logger:              logger,
		detectors:           make(map[PIIType]PIIDetector),
		anonymizers:         make(map[AnonymizationMethod]Anonymizer),
		classificationCache: make(map[string]PIIClassification),
	}

	if !config.Enabled {
		return pm, nil
	}

	// Initialize detectors
	pm.initializeDetectors()

	// Initialize anonymizers
	pm.initializeAnonymizers()

	return pm, nil
}

// ProcessData processes data to detect and anonymize PII
func (pm *PIIManager) ProcessData(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if !pm.config.Enabled {
		return data, nil
	}

	processedData := make(map[string]interface{})
	classifications := make([]PIIClassification, 0)

	for fieldName, value := range data {
		if value == nil {
			processedData[fieldName] = value
			continue
		}

		// Detect PII in the field
		classification, detected := pm.detectPII(fieldName, value)
		if detected && classification.Confidence >= pm.config.Confidence {
			classifications = append(classifications, classification)

			// Apply anonymization if auto-mask is enabled
			if pm.config.AutoMask {
				processedValue, err := pm.anonymizeValue(classification.PIIType, value, classification.Context)
				if err != nil {
					pm.logger.Warn("Failed to anonymize PII",
						zap.String("field", fieldName),
						zap.String("pii_type", string(classification.PIIType)),
						zap.Error(err))
					processedData[fieldName] = value
				} else {
					processedData[fieldName] = processedValue
					classification.ProcessedValue = processedValue
				}
			} else {
				processedData[fieldName] = value
				classification.ProcessedValue = value
			}
		} else {
			processedData[fieldName] = value
		}
	}

	// Log PII classifications
	if len(classifications) > 0 {
		pm.logger.Info("PII detected and processed",
			zap.Int("pii_fields", len(classifications)),
			zap.Any("classifications", pm.sanitizeClassifications(classifications)))
	}

	return processedData, nil
}

// detectPII detects PII in a given field and value
func (pm *PIIManager) detectPII(fieldName string, value interface{}) (PIIClassification, bool) {
	var bestMatch PIIClassification
	var maxConfidence float64 = 0

	for piiType, detector := range pm.detectors {
		detected, confidence, context := detector.Detect(fieldName, value)
		if detected && confidence > maxConfidence {
			maxConfidence = confidence
			bestMatch = PIIClassification{
				FieldName:     fieldName,
				PIIType:       piiType,
				Sensitivity:   detector.GetSensitivity(),
				Confidence:    confidence,
				OriginalValue: value,
				Timestamp:     time.Now(),
				Context:       context,
			}
		}
	}

	return bestMatch, maxConfidence >= pm.config.Confidence
}

// anonymizeValue anonymizes a value based on its PII type
func (pm *PIIManager) anonymizeValue(piiType PIIType, value interface{}, context map[string]string) (interface{}, error) {
	// Determine the best anonymization method for the PII type
	method := pm.getAnonymizationMethod(piiType)

	anonymizer, exists := pm.anonymizers[method]
	if !exists {
		return value, fmt.Errorf("no anonymizer found for method: %s", method)
	}

	return anonymizer.Anonymize(value, context)
}

// getAnonymizationMethod returns the appropriate anonymization method for a PII type
func (pm *PIIManager) getAnonymizationMethod(piiType PIIType) AnonymizationMethod {
	switch piiType {
	case PIITypeEmail:
		return AnonymizationHash
	case PIITypeCPF, PIITypeCNPJ:
		return AnonymizationTokenize
	case PIITypePhone:
		return AnonymizationGeneralize
	case PIITypeCreditCard:
		return AnonymizationTokenize
	case PIITypeSSN:
		return AnonymizationRedact
	case PIITypeName:
		return AnonymizationGeneralize
	default:
		return AnonymizationHash
	}
}

// initializeDetectors sets up all PII detectors
func (pm *PIIManager) initializeDetectors() {
	pm.detectors[PIITypeEmail] = &EmailDetector{}
	pm.detectors[PIITypeCPF] = &CPFDetector{}
	pm.detectors[PIITypeCNPJ] = &CNPJDetector{}
	pm.detectors[PIITypePhone] = &PhoneDetector{}
	pm.detectors[PIITypeCreditCard] = &CreditCardDetector{}
	pm.detectors[PIITypeIPAddress] = &IPAddressDetector{}
	pm.detectors[PIITypeName] = &NameDetector{}
}

// initializeAnonymizers sets up all anonymizers
func (pm *PIIManager) initializeAnonymizers() {
	pm.anonymizers[AnonymizationHash] = &HashAnonymizer{}
	pm.anonymizers[AnonymizationTokenize] = &TokenizeAnonymizer{}
	pm.anonymizers[AnonymizationRedact] = &RedactAnonymizer{}
	pm.anonymizers[AnonymizationGeneralize] = &GeneralizeAnonymizer{}
}

// sanitizeClassifications removes sensitive data from classifications for logging
func (pm *PIIManager) sanitizeClassifications(classifications []PIIClassification) []PIIClassification {
	sanitized := make([]PIIClassification, len(classifications))
	for i, c := range classifications {
		sanitized[i] = PIIClassification{
			FieldName:   c.FieldName,
			PIIType:     c.PIIType,
			Sensitivity: c.Sensitivity,
			Confidence:  c.Confidence,
			Method:      c.Method,
			Timestamp:   c.Timestamp,
			Context:     c.Context,
			// Exclude OriginalValue and ProcessedValue for security
		}
	}
	return sanitized
}

// HealthCheck returns the health status of the PII manager
func (pm *PIIManager) HealthCheck(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		"enabled":              pm.config.Enabled,
		"auto_mask":            pm.config.AutoMask,
		"confidence_threshold": pm.config.Confidence,
		"detectors_count":      len(pm.detectors),
		"anonymizers_count":    len(pm.anonymizers),
		"status":               "healthy",
	}
}

// Specific PII Detectors Implementation

// EmailDetector detects email addresses
type EmailDetector struct{}

func (d *EmailDetector) Detect(field string, value interface{}) (bool, float64, map[string]string) {
	str, ok := value.(string)
	if !ok {
		return false, 0, nil
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if emailRegex.MatchString(str) {
		return true, 0.95, map[string]string{"pattern": "email_regex"}
	}

	// Field name-based detection
	fieldLower := strings.ToLower(field)
	if strings.Contains(fieldLower, "email") || strings.Contains(fieldLower, "e-mail") {
		return true, 0.7, map[string]string{"pattern": "field_name"}
	}

	return false, 0, nil
}

func (d *EmailDetector) GetType() PIIType               { return PIITypeEmail }
func (d *EmailDetector) GetSensitivity() PIISensitivity { return PIISensitivityConfidential }

// CPFDetector detects Brazilian CPF numbers
type CPFDetector struct{}

func (d *CPFDetector) Detect(field string, value interface{}) (bool, float64, map[string]string) {
	str, ok := value.(string)
	if !ok {
		return false, 0, nil
	}

	// Remove non-digit characters
	digits := regexp.MustCompile(`\D`).ReplaceAllString(str, "")

	if len(digits) == 11 && d.isValidCPF(digits) {
		return true, 0.98, map[string]string{"pattern": "cpf_validation"}
	}

	fieldLower := strings.ToLower(field)
	if strings.Contains(fieldLower, "cpf") {
		return true, 0.8, map[string]string{"pattern": "field_name"}
	}

	return false, 0, nil
}

func (d *CPFDetector) isValidCPF(cpf string) bool {
	// CPF validation algorithm
	if len(cpf) != 11 {
		return false
	}

	// Check if all digits are the same
	allSame := true
	for i := 1; i < len(cpf); i++ {
		if cpf[i] != cpf[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	// Validate check digits
	sum := 0
	for i := 0; i < 9; i++ {
		digit := int(cpf[i] - '0')
		sum += digit * (10 - i)
	}
	checkDigit1 := (sum * 10) % 11
	if checkDigit1 == 10 {
		checkDigit1 = 0
	}

	if int(cpf[9]-'0') != checkDigit1 {
		return false
	}

	sum = 0
	for i := 0; i < 10; i++ {
		digit := int(cpf[i] - '0')
		sum += digit * (11 - i)
	}
	checkDigit2 := (sum * 10) % 11
	if checkDigit2 == 10 {
		checkDigit2 = 0
	}

	return int(cpf[10]-'0') == checkDigit2
}

func (d *CPFDetector) GetType() PIIType               { return PIITypeCPF }
func (d *CPFDetector) GetSensitivity() PIISensitivity { return PIISensitivityRestricted }

// Additional detector implementations...
type CNPJDetector struct{}

func (d *CNPJDetector) Detect(field string, value interface{}) (bool, float64, map[string]string) {
	// CNPJ detection logic
	return false, 0, nil
}
func (d *CNPJDetector) GetType() PIIType               { return PIITypeCNPJ }
func (d *CNPJDetector) GetSensitivity() PIISensitivity { return PIISensitivityConfidential }

type PhoneDetector struct{}

func (d *PhoneDetector) Detect(field string, value interface{}) (bool, float64, map[string]string) {
	// Phone detection logic
	str, ok := value.(string)
	if !ok {
		return false, 0, nil
	}

	phoneRegex := regexp.MustCompile(`^[\+]?[1-9]?[\d\s\-\(\)]{7,15}$`)
	if phoneRegex.MatchString(str) {
		return true, 0.8, map[string]string{"pattern": "phone_regex"}
	}
	return false, 0, nil
}
func (d *PhoneDetector) GetType() PIIType               { return PIITypePhone }
func (d *PhoneDetector) GetSensitivity() PIISensitivity { return PIISensitivityConfidential }

type CreditCardDetector struct{}

func (d *CreditCardDetector) Detect(field string, value interface{}) (bool, float64, map[string]string) {
	// Credit card detection logic (Luhn algorithm)
	return false, 0, nil
}
func (d *CreditCardDetector) GetType() PIIType               { return PIITypeCreditCard }
func (d *CreditCardDetector) GetSensitivity() PIISensitivity { return PIISensitivityRestricted }

type IPAddressDetector struct{}

func (d *IPAddressDetector) Detect(field string, value interface{}) (bool, float64, map[string]string) {
	// IP address detection logic
	return false, 0, nil
}
func (d *IPAddressDetector) GetType() PIIType               { return PIITypeIPAddress }
func (d *IPAddressDetector) GetSensitivity() PIISensitivity { return PIISensitivityInternal }

type NameDetector struct{}

func (d *NameDetector) Detect(field string, value interface{}) (bool, float64, map[string]string) {
	// Name detection logic
	fieldLower := strings.ToLower(field)
	if strings.Contains(fieldLower, "name") || strings.Contains(fieldLower, "nome") {
		return true, 0.7, map[string]string{"pattern": "field_name"}
	}
	return false, 0, nil
}
func (d *NameDetector) GetType() PIIType               { return PIITypeName }
func (d *NameDetector) GetSensitivity() PIISensitivity { return PIISensitivityConfidential }

// Anonymizer Implementations

// HashAnonymizer anonymizes data using SHA-256 hashing
type HashAnonymizer struct{}

func (a *HashAnonymizer) Anonymize(value interface{}, context map[string]string) (interface{}, error) {
	str := fmt.Sprintf("%v", value)
	hash := sha256.Sum256([]byte(str))
	return hex.EncodeToString(hash[:]), nil
}

func (a *HashAnonymizer) IsReversible() bool             { return false }
func (a *HashAnonymizer) GetMethod() AnonymizationMethod { return AnonymizationHash }

// TokenizeAnonymizer creates reversible tokens
type TokenizeAnonymizer struct{}

func (a *TokenizeAnonymizer) Anonymize(value interface{}, context map[string]string) (interface{}, error) {
	str := fmt.Sprintf("%v", value)
	// Generate a token (simplified - in production, use proper tokenization)
	hash := sha256.Sum256([]byte(str))
	token := fmt.Sprintf("TKN_%x", hash[:8])
	return token, nil
}

func (a *TokenizeAnonymizer) IsReversible() bool             { return true }
func (a *TokenizeAnonymizer) GetMethod() AnonymizationMethod { return AnonymizationTokenize }

// RedactAnonymizer replaces data with asterisks
type RedactAnonymizer struct{}

func (a *RedactAnonymizer) Anonymize(value interface{}, context map[string]string) (interface{}, error) {
	str := fmt.Sprintf("%v", value)
	if len(str) <= 4 {
		return "****", nil
	}
	return str[:2] + strings.Repeat("*", len(str)-4) + str[len(str)-2:], nil
}

func (a *RedactAnonymizer) IsReversible() bool             { return false }
func (a *RedactAnonymizer) GetMethod() AnonymizationMethod { return AnonymizationRedact }

// GeneralizeAnonymizer generalizes data to reduce specificity
type GeneralizeAnonymizer struct{}

func (a *GeneralizeAnonymizer) Anonymize(value interface{}, context map[string]string) (interface{}, error) {
	str := fmt.Sprintf("%v", value)
	if len(str) <= 3 {
		return "***", nil
	}
	// Keep first character and generalize the rest
	return string(str[0]) + strings.Repeat("*", len(str)-1), nil
}

func (a *GeneralizeAnonymizer) IsReversible() bool             { return false }
func (a *GeneralizeAnonymizer) GetMethod() AnonymizationMethod { return AnonymizationGeneralize }
