package compliance

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ConsentManager handles user consent for data processing
type ConsentManager struct {
	config     ConsentConfig
	logger     *zap.Logger
	repository ConsentRepository
}

// ConsentRepository interface for storing consent data
type ConsentRepository interface {
	StoreConsent(ctx context.Context, consent ConsentRecord) error
	GetConsent(ctx context.Context, subjectID, purpose string) (*ConsentRecord, error)
	GetAllConsents(ctx context.Context, subjectID string) ([]ConsentRecord, error)
	UpdateConsent(ctx context.Context, consent ConsentRecord) error
	DeleteConsent(ctx context.Context, subjectID, purpose string) error
	GetConsentHistory(ctx context.Context, subjectID, purpose string) ([]ConsentRecord, error)
}

// ConsentRecord represents a consent record in storage
type ConsentRecord struct {
	ID            string                 `json:"id" db:"id"`
	SubjectID     string                 `json:"subject_id" db:"subject_id"`
	Purpose       string                 `json:"purpose" db:"purpose"`
	Granted       bool                   `json:"granted" db:"granted"`
	LegalBasis    string                 `json:"legal_basis" db:"legal_basis"`
	ConsentSource ConsentSource          `json:"consent_source" db:"consent_source"`
	Timestamp     time.Time              `json:"timestamp" db:"timestamp"`
	ExpiresAt     *time.Time             `json:"expires_at,omitempty" db:"expires_at"`
	WithdrawnAt   *time.Time             `json:"withdrawn_at,omitempty" db:"withdrawn_at"`
	IPAddress     string                 `json:"ip_address" db:"ip_address"`
	UserAgent     string                 `json:"user_agent" db:"user_agent"`
	ConsentString string                 `json:"consent_string" db:"consent_string"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
	Version       int                    `json:"version" db:"version"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// ConsentSource represents where the consent was obtained
type ConsentSource string

const (
	ConsentSourceWeb    ConsentSource = "web"
	ConsentSourceMobile ConsentSource = "mobile"
	ConsentSourceAPI    ConsentSource = "api"
	ConsentSourcePhone  ConsentSource = "phone"
	ConsentSourceEmail  ConsentSource = "email"
	ConsentSourcePaper  ConsentSource = "paper"
	ConsentSourceImport ConsentSource = "import"
)

// ConsentRequest represents a request to grant or update consent
type ConsentRequest struct {
	SubjectID      string                 `json:"subject_id"`
	Purpose        string                 `json:"purpose"`
	Granted        bool                   `json:"granted"`
	LegalBasis     string                 `json:"legal_basis"`
	ConsentSource  ConsentSource          `json:"consent_source"`
	ExpirationDays *int                   `json:"expiration_days,omitempty"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	ConsentString  string                 `json:"consent_string"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ConsentValidationResult represents the result of consent validation
type ConsentValidationResult struct {
	Valid           bool           `json:"valid"`
	Consent         *ConsentRecord `json:"consent,omitempty"`
	Reason          string         `json:"reason,omitempty"`
	RequiredActions []string       `json:"required_actions,omitempty"`
	ExpiresIn       *time.Duration `json:"expires_in,omitempty"`
}

// LegalBasis represents the legal basis for processing personal data
type LegalBasis string

const (
	// GDPR Legal Bases (Article 6)
	LegalBasisConsent             LegalBasis = "consent"              // Article 6(1)(a)
	LegalBasisContract            LegalBasis = "contract"             // Article 6(1)(b)
	LegalBasisLegalObligation     LegalBasis = "legal_obligation"     // Article 6(1)(c)
	LegalBasisVitalInterests      LegalBasis = "vital_interests"      // Article 6(1)(d)
	LegalBasisPublicTask          LegalBasis = "public_task"          // Article 6(1)(e)
	LegalBasisLegitimateInterests LegalBasis = "legitimate_interests" // Article 6(1)(f)

	// LGPD Legal Bases (Article 7)
	LegalBasisLGPDConsent          LegalBasis = "lgpd_consent"           // Article 7(I)
	LegalBasisLGPDCompliance       LegalBasis = "lgpd_compliance"        // Article 7(II)
	LegalBasisLGPDPublicAdmin      LegalBasis = "lgpd_public_admin"      // Article 7(III)
	LegalBasisLGPDStudies          LegalBasis = "lgpd_studies"           // Article 7(IV)
	LegalBasisLGPDContractual      LegalBasis = "lgpd_contractual"       // Article 7(V)
	LegalBasisLGPDJudicial         LegalBasis = "lgpd_judicial"          // Article 7(VI)
	LegalBasisLGPDHealthLife       LegalBasis = "lgpd_health_life"       // Article 7(VII)
	LegalBasisLGPDHealthSecure     LegalBasis = "lgpd_health_secure"     // Article 7(VIII)
	LegalBasisLGPDLegitimate       LegalBasis = "lgpd_legitimate"        // Article 7(IX)
	LegalBasisLGPDCreditProtection LegalBasis = "lgpd_credit_protection" // Article 7(X)
)

// NewConsentManager creates a new consent manager
func NewConsentManager(config ConsentConfig, logger *zap.Logger) (*ConsentManager, error) {
	// In a real implementation, you would inject the repository
	// For now, we'll use a mock repository
	repository := &InMemoryConsentRepository{
		consents: make(map[string][]ConsentRecord),
	}

	return &ConsentManager{
		config:     config,
		logger:     logger,
		repository: repository,
	}, nil
}

// GrantConsent grants consent for a specific purpose
func (cm *ConsentManager) GrantConsent(ctx context.Context, request ConsentRequest) (*ConsentRecord, error) {
	if !cm.config.Enabled {
		return nil, fmt.Errorf("consent management is disabled")
	}

	// Validate the consent request
	if err := cm.validateConsentRequest(request); err != nil {
		return nil, fmt.Errorf("invalid consent request: %w", err)
	}

	// Check if consent already exists
	existing, err := cm.repository.GetConsent(ctx, request.SubjectID, request.Purpose)
	if err == nil && existing != nil {
		// Update existing consent
		return cm.updateExistingConsent(ctx, existing, request)
	}

	// Create new consent record
	consent := ConsentRecord{
		ID:            cm.generateConsentID(),
		SubjectID:     request.SubjectID,
		Purpose:       request.Purpose,
		Granted:       request.Granted,
		LegalBasis:    request.LegalBasis,
		ConsentSource: request.ConsentSource,
		Timestamp:     time.Now(),
		IPAddress:     request.IPAddress,
		UserAgent:     request.UserAgent,
		ConsentString: request.ConsentString,
		Metadata:      request.Metadata,
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Set expiration if specified
	if request.ExpirationDays != nil {
		expiresAt := time.Now().AddDate(0, 0, *request.ExpirationDays)
		consent.ExpiresAt = &expiresAt
	} else if cm.config.TTL > 0 {
		expiresAt := time.Now().Add(cm.config.TTL)
		consent.ExpiresAt = &expiresAt
	}

	// Store the consent
	if err := cm.repository.StoreConsent(ctx, consent); err != nil {
		return nil, fmt.Errorf("failed to store consent: %w", err)
	}

	cm.logger.Info("Consent granted",
		zap.String("subject_id", consent.SubjectID),
		zap.String("purpose", consent.Purpose),
		zap.Bool("granted", consent.Granted),
		zap.String("legal_basis", consent.LegalBasis),
		zap.String("source", string(consent.ConsentSource)))

	return &consent, nil
}

// RecordConsent is a convenience method to record consent with minimal parameters
func (cm *ConsentManager) RecordConsent(ctx context.Context, subjectID, purpose, source string) error {
	request := ConsentRequest{
		SubjectID:     subjectID,
		Purpose:       purpose,
		Granted:       true,
		LegalBasis:    string(LegalBasisConsent),
		ConsentSource: ConsentSource(source),
	}

	_, err := cm.GrantConsent(ctx, request)
	return err
}

// HasValidConsent checks if valid consent exists for a specific purpose
func (cm *ConsentManager) HasValidConsent(ctx context.Context, subjectID, purpose string) (bool, error) {
	if !cm.config.Enabled {
		return true, nil // Allow processing if consent management is disabled
	}

	result := cm.ValidateConsent(ctx, subjectID, purpose)
	return result.Valid, nil
}

// ValidateConsent validates consent for a specific purpose and returns detailed information
func (cm *ConsentManager) ValidateConsent(ctx context.Context, subjectID, purpose string) ConsentValidationResult {
	if !cm.config.Enabled {
		return ConsentValidationResult{
			Valid:  true,
			Reason: "consent management disabled",
		}
	}

	consent, err := cm.repository.GetConsent(ctx, subjectID, purpose)
	if err != nil {
		return ConsentValidationResult{
			Valid:           false,
			Reason:          "consent not found",
			RequiredActions: []string{"obtain_consent"},
		}
	}

	// Check if consent was granted
	if !consent.Granted {
		return ConsentValidationResult{
			Valid:           false,
			Consent:         consent,
			Reason:          "consent not granted",
			RequiredActions: []string{"request_consent"},
		}
	}

	// Check if consent was withdrawn
	if consent.WithdrawnAt != nil {
		return ConsentValidationResult{
			Valid:           false,
			Consent:         consent,
			Reason:          "consent withdrawn",
			RequiredActions: []string{"obtain_new_consent"},
		}
	}

	// Check if consent has expired
	if consent.ExpiresAt != nil && time.Now().After(*consent.ExpiresAt) {
		return ConsentValidationResult{
			Valid:           false,
			Consent:         consent,
			Reason:          "consent expired",
			RequiredActions: []string{"renew_consent"},
		}
	}

	// Calculate time until expiration
	var expiresIn *time.Duration
	if consent.ExpiresAt != nil {
		duration := time.Until(*consent.ExpiresAt)
		expiresIn = &duration
	}

	return ConsentValidationResult{
		Valid:     true,
		Consent:   consent,
		Reason:    "valid consent",
		ExpiresIn: expiresIn,
	}
}

// WithdrawConsent withdraws consent for a specific purpose
func (cm *ConsentManager) WithdrawConsent(ctx context.Context, subjectID, purpose string) error {
	if !cm.config.Enabled {
		return fmt.Errorf("consent management is disabled")
	}

	consent, err := cm.repository.GetConsent(ctx, subjectID, purpose)
	if err != nil {
		return fmt.Errorf("consent not found: %w", err)
	}

	// Mark consent as withdrawn
	now := time.Now()
	consent.WithdrawnAt = &now
	consent.UpdatedAt = now
	consent.Version++

	if err := cm.repository.UpdateConsent(ctx, *consent); err != nil {
		return fmt.Errorf("failed to withdraw consent: %w", err)
	}

	cm.logger.Info("Consent withdrawn",
		zap.String("subject_id", subjectID),
		zap.String("purpose", purpose))

	return nil
}

// GetConsentHistory returns the consent history for a subject and purpose
func (cm *ConsentManager) GetConsentHistory(ctx context.Context, subjectID, purpose string) ([]ConsentRecord, error) {
	if !cm.config.Enabled {
		return nil, fmt.Errorf("consent management is disabled")
	}

	return cm.repository.GetConsentHistory(ctx, subjectID, purpose)
}

// GetAllConsents returns all consents for a data subject
func (cm *ConsentManager) GetAllConsents(ctx context.Context, subjectID string) ([]ConsentRecord, error) {
	if !cm.config.Enabled {
		return nil, fmt.Errorf("consent management is disabled")
	}

	return cm.repository.GetAllConsents(ctx, subjectID)
}

// HealthCheck returns the health status of the consent manager
func (cm *ConsentManager) HealthCheck(_ context.Context) map[string]interface{} {
	return map[string]interface{}{
		"enabled":          cm.config.Enabled,
		"default_purposes": cm.config.DefaultPurposes,
		"ttl":              cm.config.TTL.String(),
		"granular_level":   cm.config.GranularLevel,
		"status":           "healthy",
	}
}

// Helper methods

func (cm *ConsentManager) validateConsentRequest(request ConsentRequest) error {
	if request.SubjectID == "" {
		return fmt.Errorf("subject_id is required")
	}
	if request.Purpose == "" {
		return fmt.Errorf("purpose is required")
	}
	if request.LegalBasis == "" {
		return fmt.Errorf("legal_basis is required")
	}
	return nil
}

func (cm *ConsentManager) updateExistingConsent(ctx context.Context, existing *ConsentRecord, request ConsentRequest) (*ConsentRecord, error) {
	existing.Granted = request.Granted
	existing.LegalBasis = request.LegalBasis
	existing.ConsentSource = request.ConsentSource
	existing.Timestamp = time.Now()
	existing.IPAddress = request.IPAddress
	existing.UserAgent = request.UserAgent
	existing.ConsentString = request.ConsentString
	existing.Metadata = request.Metadata
	existing.UpdatedAt = time.Now()
	existing.Version++

	// Reset withdrawn status if granting consent again
	if request.Granted {
		existing.WithdrawnAt = nil
	}

	// Update expiration
	if request.ExpirationDays != nil {
		expiresAt := time.Now().AddDate(0, 0, *request.ExpirationDays)
		existing.ExpiresAt = &expiresAt
	}

	if err := cm.repository.UpdateConsent(ctx, *existing); err != nil {
		return nil, fmt.Errorf("failed to update consent: %w", err)
	}

	return existing, nil
}

func (cm *ConsentManager) generateConsentID() string {
	return fmt.Sprintf("consent_%d", time.Now().UnixNano())
}

// InMemoryConsentRepository is a simple in-memory implementation for development/testing
type InMemoryConsentRepository struct {
	consents map[string][]ConsentRecord
}

func (r *InMemoryConsentRepository) StoreConsent(ctx context.Context, consent ConsentRecord) error {
	key := fmt.Sprintf("%s:%s", consent.SubjectID, consent.Purpose)
	r.consents[key] = append(r.consents[key], consent)
	return nil
}

func (r *InMemoryConsentRepository) GetConsent(ctx context.Context, subjectID, purpose string) (*ConsentRecord, error) {
	key := fmt.Sprintf("%s:%s", subjectID, purpose)
	consents, exists := r.consents[key]
	if !exists || len(consents) == 0 {
		return nil, fmt.Errorf("consent not found")
	}

	// Return the most recent consent
	return &consents[len(consents)-1], nil
}

func (r *InMemoryConsentRepository) GetAllConsents(ctx context.Context, subjectID string) ([]ConsentRecord, error) {
	var allConsents []ConsentRecord
	for _, consents := range r.consents {
		if len(consents) > 0 && consents[0].SubjectID == subjectID {
			// Return the most recent consent for each purpose
			allConsents = append(allConsents, consents[len(consents)-1])
		}
	}
	return allConsents, nil
}

func (r *InMemoryConsentRepository) UpdateConsent(ctx context.Context, consent ConsentRecord) error {
	key := fmt.Sprintf("%s:%s", consent.SubjectID, consent.Purpose)
	consents := r.consents[key]
	if len(consents) > 0 {
		// Replace the most recent consent
		consents[len(consents)-1] = consent
		r.consents[key] = consents
	}
	return nil
}

func (r *InMemoryConsentRepository) DeleteConsent(ctx context.Context, subjectID, purpose string) error {
	key := fmt.Sprintf("%s:%s", subjectID, purpose)
	delete(r.consents, key)
	return nil
}

func (r *InMemoryConsentRepository) GetConsentHistory(ctx context.Context, subjectID, purpose string) ([]ConsentRecord, error) {
	key := fmt.Sprintf("%s:%s", subjectID, purpose)
	consents, exists := r.consents[key]
	if !exists {
		return nil, fmt.Errorf("consent history not found")
	}
	return consents, nil
}
