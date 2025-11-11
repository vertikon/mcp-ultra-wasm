package compliance

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// RetentionManager handles data retention policies and lifecycle management
type RetentionManager struct {
	config     DataRetentionConfig
	logger     *zap.Logger
	policies   map[string]RetentionPolicy
	scheduler  *RetentionScheduler
	repository RetentionRepository
}

// RetentionRepository interface for persistence
type RetentionRepository interface {
	StoreRetentionRecord(ctx context.Context, record RetentionRecord) error
	GetRetentionRecords(ctx context.Context, subjectID string) ([]RetentionRecord, error)
	GetExpiredRecords(ctx context.Context, beforeTime time.Time) ([]RetentionRecord, error)
	UpdateRetentionRecord(ctx context.Context, record RetentionRecord) error
	DeleteRetentionRecord(ctx context.Context, recordID string) error
}

// RetentionPolicy defines a data retention policy
type RetentionPolicy struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Category        string                 `json:"category"`
	RetentionPeriod time.Duration          `json:"retention_period"`
	GracePeriod     time.Duration          `json:"grace_period"`
	Action          RetentionAction        `json:"action"`
	Priority        int                    `json:"priority"`
	Conditions      []RetentionCondition   `json:"conditions"`
	Exceptions      []RetentionException   `json:"exceptions"`
	LegalBasis      []LegalBasis           `json:"legal_basis"`
	Jurisdictions   []string               `json:"jurisdictions"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	IsActive        bool                   `json:"is_active"`
}

// RetentionAction defines what happens when data retention period expires
type RetentionAction string

const (
	RetentionActionDelete    RetentionAction = "delete"
	RetentionActionArchive   RetentionAction = "archive"
	RetentionActionAnonymize RetentionAction = "anonymize"
	RetentionActionNotify    RetentionAction = "notify"
	RetentionActionReview    RetentionAction = "review"
	RetentionActionPurge     RetentionAction = "purge"
)

// RetentionCondition defines conditions for applying retention policies
type RetentionCondition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
	LogicType string      `json:"logic_type"` // AND, OR
}

// RetentionException defines exceptions to retention policies
type RetentionException struct {
	Reason     string                 `json:"reason"`
	ExtendBy   time.Duration          `json:"extend_by"`
	Conditions []RetentionCondition   `json:"conditions"`
	ApprovedBy string                 `json:"approved_by"`
	ApprovedAt time.Time              `json:"approved_at"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// RetentionRecord tracks retention information for specific data
type RetentionRecord struct {
	ID              string                 `json:"id"`
	SubjectID       string                 `json:"subject_id"`
	DataType        string                 `json:"data_type"`
	PolicyID        string                 `json:"policy_id"`
	CreatedAt       time.Time              `json:"created_at"`
	RetentionStart  time.Time              `json:"retention_start"`
	RetentionEnd    time.Time              `json:"retention_end"`
	GraceEnd        *time.Time             `json:"grace_end,omitempty"`
	Status          RetentionStatus        `json:"status"`
	Action          RetentionAction        `json:"action"`
	ActionTaken     bool                   `json:"action_taken"`
	ActionTakenAt   *time.Time             `json:"action_taken_at,omitempty"`
	LegalHold       bool                   `json:"legal_hold"`
	LegalHoldReason string                 `json:"legal_hold_reason,omitempty"`
	Extensions      []RetentionExtension   `json:"extensions"`
	Metadata        map[string]interface{} `json:"metadata"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// RetentionStatus represents the current status of a retention record
type RetentionStatus string

const (
	RetentionStatusActive     RetentionStatus = "active"
	RetentionStatusExpired    RetentionStatus = "expired"
	RetentionStatusProcessing RetentionStatus = "processing"
	RetentionStatusCompleted  RetentionStatus = "completed"
	RetentionStatusOnHold     RetentionStatus = "on_hold"
	RetentionStatusExtended   RetentionStatus = "extended"
)

// RetentionExtension represents an extension to a retention period
type RetentionExtension struct {
	Reason     string        `json:"reason"`
	ExtendBy   time.Duration `json:"extend_by"`
	ExtendedBy string        `json:"extended_by"`
	ExtendedAt time.Time     `json:"extended_at"`
	ExpiresAt  time.Time     `json:"expires_at"`
	Approved   bool          `json:"approved"`
}

// RetentionScheduler handles scheduling of retention actions
type RetentionScheduler struct {
	manager *RetentionManager
	ticker  *time.Ticker
	stop    chan bool
}

// NewRetentionManager creates a new retention manager
func NewRetentionManager(config DataRetentionConfig, logger *zap.Logger) (*RetentionManager, error) {
	rm := &RetentionManager{
		config:   config,
		logger:   logger,
		policies: make(map[string]RetentionPolicy),
		repository: &InMemoryRetentionRepository{
			records: make(map[string][]RetentionRecord),
		},
	}

	if !config.Enabled {
		return rm, nil
	}

	// Initialize default retention policies
	rm.initializeDefaultPolicies()

	// Create and start scheduler if auto-delete is enabled
	if config.AutoDelete {
		rm.scheduler = &RetentionScheduler{
			manager: rm,
			ticker:  time.NewTicker(24 * time.Hour), // Run daily
			stop:    make(chan bool),
		}
		go rm.scheduler.run()
	}

	return rm, nil
}

// ApplyRetentionPolicy applies retention policy to specific data
func (rm *RetentionManager) ApplyRetentionPolicy(ctx context.Context, subjectID string, data map[string]interface{}) error {
	if !rm.config.Enabled {
		return nil
	}

	// Determine applicable policies based on data
	policies := rm.getApplicablePolicies(data)

	for _, policy := range policies {
		// Create retention record
		record := RetentionRecord{
			ID:             rm.generateRecordID(),
			SubjectID:      subjectID,
			DataType:       rm.inferDataType(data),
			PolicyID:       policy.ID,
			CreatedAt:      time.Now(),
			RetentionStart: time.Now(),
			RetentionEnd:   time.Now().Add(policy.RetentionPeriod),
			Status:         RetentionStatusActive,
			Action:         policy.Action,
			UpdatedAt:      time.Now(),
		}

		// Add grace period if configured
		if policy.GracePeriod > 0 {
			graceEnd := record.RetentionEnd.Add(policy.GracePeriod)
			record.GraceEnd = &graceEnd
		}

		// Store retention record
		if err := rm.repository.StoreRetentionRecord(ctx, record); err != nil {
			rm.logger.Error("Failed to store retention record",
				zap.String("subject_id", subjectID),
				zap.String("policy_id", policy.ID),
				zap.Error(err))
			continue
		}

		rm.logger.Debug("Retention policy applied",
			zap.String("subject_id", subjectID),
			zap.String("policy_id", policy.ID),
			zap.String("action", string(policy.Action)),
			zap.Time("retention_end", record.RetentionEnd))
	}

	return nil
}

// ProcessExpiredRetentions processes expired retention records
func (rm *RetentionManager) ProcessExpiredRetentions(ctx context.Context) error {
	if !rm.config.Enabled {
		return nil
	}

	// Get expired records
	expiredRecords, err := rm.repository.GetExpiredRecords(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to get expired records: %w", err)
	}

	rm.logger.Info("Processing expired retentions", zap.Int("count", len(expiredRecords)))

	for _, record := range expiredRecords {
		if record.LegalHold {
			rm.logger.Info("Skipping record on legal hold",
				zap.String("record_id", record.ID),
				zap.String("subject_id", record.SubjectID))
			continue
		}

		// Check grace period
		if record.GraceEnd != nil && time.Now().Before(*record.GraceEnd) {
			continue
		}

		// Execute retention action
		if err := rm.executeRetentionAction(ctx, record); err != nil {
			rm.logger.Error("Failed to execute retention action",
				zap.String("record_id", record.ID),
				zap.String("action", string(record.Action)),
				zap.Error(err))
			continue
		}

		// Update record status
		record.Status = RetentionStatusCompleted
		record.ActionTaken = true
		now := time.Now()
		record.ActionTakenAt = &now
		record.UpdatedAt = now

		if err := rm.repository.UpdateRetentionRecord(ctx, record); err != nil {
			rm.logger.Error("Failed to update retention record",
				zap.String("record_id", record.ID),
				zap.Error(err))
		}
	}

	return nil
}

// ExtendRetention extends the retention period for specific data
func (rm *RetentionManager) ExtendRetention(ctx context.Context, subjectID, reason string, extendBy time.Duration, approvedBy string) error {
	if !rm.config.Enabled {
		return fmt.Errorf("retention management is disabled")
	}

	records, err := rm.repository.GetRetentionRecords(ctx, subjectID)
	if err != nil {
		return fmt.Errorf("failed to get retention records: %w", err)
	}

	for _, record := range records {
		if record.Status == RetentionStatusActive || record.Status == RetentionStatusExtended {
			extension := RetentionExtension{
				Reason:     reason,
				ExtendBy:   extendBy,
				ExtendedBy: approvedBy,
				ExtendedAt: time.Now(),
				ExpiresAt:  record.RetentionEnd.Add(extendBy),
				Approved:   true,
			}

			record.Extensions = append(record.Extensions, extension)
			record.RetentionEnd = extension.ExpiresAt
			record.Status = RetentionStatusExtended
			record.UpdatedAt = time.Now()

			if err := rm.repository.UpdateRetentionRecord(ctx, record); err != nil {
				rm.logger.Error("Failed to extend retention",
					zap.String("record_id", record.ID),
					zap.Error(err))
				continue
			}
		}
	}

	rm.logger.Info("Retention extended",
		zap.String("subject_id", subjectID),
		zap.String("reason", reason),
		zap.Duration("extend_by", extendBy))

	return nil
}

// PlaceLegalHold places a legal hold on data to prevent deletion
func (rm *RetentionManager) PlaceLegalHold(ctx context.Context, subjectID, reason string) error {
	if !rm.config.Enabled {
		return fmt.Errorf("retention management is disabled")
	}

	records, err := rm.repository.GetRetentionRecords(ctx, subjectID)
	if err != nil {
		return fmt.Errorf("failed to get retention records: %w", err)
	}

	for _, record := range records {
		record.LegalHold = true
		record.LegalHoldReason = reason
		record.Status = RetentionStatusOnHold
		record.UpdatedAt = time.Now()

		if err := rm.repository.UpdateRetentionRecord(ctx, record); err != nil {
			rm.logger.Error("Failed to place legal hold",
				zap.String("record_id", record.ID),
				zap.Error(err))
		}
	}

	rm.logger.Info("Legal hold placed",
		zap.String("subject_id", subjectID),
		zap.String("reason", reason))

	return nil
}

// RemoveLegalHold removes a legal hold from data
func (rm *RetentionManager) RemoveLegalHold(ctx context.Context, subjectID string) error {
	if !rm.config.Enabled {
		return fmt.Errorf("retention management is disabled")
	}

	records, err := rm.repository.GetRetentionRecords(ctx, subjectID)
	if err != nil {
		return fmt.Errorf("failed to get retention records: %w", err)
	}

	for _, record := range records {
		if record.LegalHold {
			record.LegalHold = false
			record.LegalHoldReason = ""
			record.Status = RetentionStatusActive
			record.UpdatedAt = time.Now()

			if err := rm.repository.UpdateRetentionRecord(ctx, record); err != nil {
				rm.logger.Error("Failed to remove legal hold",
					zap.String("record_id", record.ID),
					zap.Error(err))
			}
		}
	}

	rm.logger.Info("Legal hold removed", zap.String("subject_id", subjectID))
	return nil
}

// GetRetentionStatus returns the retention status for a data subject
func (rm *RetentionManager) GetRetentionStatus(ctx context.Context, subjectID string) ([]RetentionRecord, error) {
	if !rm.config.Enabled {
		return nil, fmt.Errorf("retention management is disabled")
	}

	return rm.repository.GetRetentionRecords(ctx, subjectID)
}

// GetPolicies returns all retention policies
func (rm *RetentionManager) GetPolicies() map[string]RetentionPolicy {
	return rm.policies
}

// RecordDataCreation records data creation for retention tracking
func (rm *RetentionManager) RecordDataCreation(ctx context.Context, subjectID, dataCategory string, data map[string]interface{}) error {
	if !rm.config.Enabled {
		return nil
	}

	// Add category metadata to data
	dataWithCategory := make(map[string]interface{})
	for k, v := range data {
		dataWithCategory[k] = v
	}
	dataWithCategory["_category"] = dataCategory

	// Apply retention policy
	return rm.ApplyRetentionPolicy(ctx, subjectID, dataWithCategory)
}

// ShouldDeleteData checks if data should be deleted based on retention policy
func (rm *RetentionManager) ShouldDeleteData(ctx context.Context, subjectID, dataCategory string) (bool, error) {
	if !rm.config.Enabled {
		return false, nil
	}

	records, err := rm.repository.GetRetentionRecords(ctx, subjectID)
	if err != nil {
		return false, fmt.Errorf("failed to get retention records: %w", err)
	}

	now := time.Now()
	for _, record := range records {
		// Skip records on legal hold
		if record.LegalHold {
			continue
		}

		// Check if record matches the data category and is expired
		if record.DataType == dataCategory || dataCategory == "" {
			// Check if retention period has expired
			if now.After(record.RetentionEnd) {
				// Check if grace period has also expired (if applicable)
				if record.GraceEnd == nil || now.After(*record.GraceEnd) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// Helper methods

func (rm *RetentionManager) executeRetentionAction(ctx context.Context, record RetentionRecord) error {
	switch record.Action {
	case RetentionActionDelete:
		return rm.deleteData(ctx, record)
	case RetentionActionArchive:
		return rm.archiveData(ctx, record)
	case RetentionActionAnonymize:
		return rm.anonymizeData(ctx, record)
	case RetentionActionPurge:
		return rm.purgeData(ctx, record)
	case RetentionActionNotify:
		return rm.notifyRetention(ctx, record)
	case RetentionActionReview:
		return rm.scheduleReview(ctx, record)
	default:
		return fmt.Errorf("unknown retention action: %s", record.Action)
	}
}

func (rm *RetentionManager) deleteData(ctx context.Context, record RetentionRecord) error {
	rm.logger.Info("Deleting data for retention policy",
		zap.String("subject_id", record.SubjectID),
		zap.String("data_type", record.DataType))
	// Implementation would delete actual data
	return nil
}

func (rm *RetentionManager) archiveData(ctx context.Context, record RetentionRecord) error {
	rm.logger.Info("Archiving data for retention policy",
		zap.String("subject_id", record.SubjectID),
		zap.String("data_type", record.DataType))
	// Implementation would archive data
	return nil
}

func (rm *RetentionManager) anonymizeData(ctx context.Context, record RetentionRecord) error {
	rm.logger.Info("Anonymizing data for retention policy",
		zap.String("subject_id", record.SubjectID),
		zap.String("data_type", record.DataType))
	// Implementation would anonymize data
	return nil
}

func (rm *RetentionManager) purgeData(ctx context.Context, record RetentionRecord) error {
	rm.logger.Info("Purging data for retention policy",
		zap.String("subject_id", record.SubjectID),
		zap.String("data_type", record.DataType))
	// Implementation would permanently delete all traces
	return nil
}

func (rm *RetentionManager) notifyRetention(ctx context.Context, record RetentionRecord) error {
	rm.logger.Info("Sending retention notification",
		zap.String("subject_id", record.SubjectID),
		zap.String("data_type", record.DataType))
	// Implementation would send notifications
	return nil
}

func (rm *RetentionManager) scheduleReview(ctx context.Context, record RetentionRecord) error {
	rm.logger.Info("Scheduling retention review",
		zap.String("subject_id", record.SubjectID),
		zap.String("data_type", record.DataType))
	// Implementation would schedule human review
	return nil
}

func (rm *RetentionManager) getApplicablePolicies(data map[string]interface{}) []RetentionPolicy {
	var applicable []RetentionPolicy

	for _, policy := range rm.policies {
		if rm.policyApplies(policy, data) {
			applicable = append(applicable, policy)
		}
	}

	return applicable
}

func (rm *RetentionManager) policyApplies(policy RetentionPolicy, data map[string]interface{}) bool {
	// Simplified policy matching - in production, this would be more sophisticated
	return policy.IsActive
}

func (rm *RetentionManager) inferDataType(data map[string]interface{}) string {
	// Infer data type based on fields present
	if _, exists := data["email"]; exists {
		return "user_data"
	}
	if _, exists := data["task_id"]; exists {
		return "task_data"
	}
	return "general_data"
}

func (rm *RetentionManager) generateRecordID() string {
	return fmt.Sprintf("retention_%d", time.Now().UnixNano())
}

func (rm *RetentionManager) initializeDefaultPolicies() {
	defaultPolicies := []RetentionPolicy{
		{
			ID:              "user_data_policy",
			Name:            "User Data Retention",
			Description:     "Standard retention policy for user data",
			Category:        "user_data",
			RetentionPeriod: rm.config.DefaultPeriod,
			GracePeriod:     30 * 24 * time.Hour, // 30 days
			Action:          RetentionActionDelete,
			Priority:        1,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "task_data_policy",
			Name:            "Task Data Retention",
			Description:     "Retention policy for task-related data",
			Category:        "operational_data",
			RetentionPeriod: time.Hour * 24 * 365, // 1 year
			Action:          RetentionActionArchive,
			Priority:        2,
			IsActive:        true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	for _, policy := range defaultPolicies {
		rm.policies[policy.ID] = policy
	}

	rm.logger.Info("Default retention policies initialized", zap.Int("policies", len(defaultPolicies)))
}

// RetentionScheduler methods

func (rs *RetentionScheduler) run() {
	for {
		select {
		case <-rs.ticker.C:
			ctx := context.Background()
			if err := rs.manager.ProcessExpiredRetentions(ctx); err != nil {
				rs.manager.logger.Error("Error processing expired retentions", zap.Error(err))
			}
		case <-rs.stop:
			rs.ticker.Stop()
			return
		}
	}
}

func (rs *RetentionScheduler) Stop() {
	rs.stop <- true
}

// InMemoryRetentionRepository implementation

type InMemoryRetentionRepository struct {
	records map[string][]RetentionRecord
}

func (r *InMemoryRetentionRepository) StoreRetentionRecord(ctx context.Context, record RetentionRecord) error {
	r.records[record.SubjectID] = append(r.records[record.SubjectID], record)
	return nil
}

func (r *InMemoryRetentionRepository) GetRetentionRecords(ctx context.Context, subjectID string) ([]RetentionRecord, error) {
	records, exists := r.records[subjectID]
	if !exists {
		return []RetentionRecord{}, nil
	}
	return records, nil
}

func (r *InMemoryRetentionRepository) GetExpiredRecords(ctx context.Context, beforeTime time.Time) ([]RetentionRecord, error) {
	var expired []RetentionRecord
	for _, recordList := range r.records {
		for _, record := range recordList {
			if record.Status == RetentionStatusActive && record.RetentionEnd.Before(beforeTime) {
				expired = append(expired, record)
			}
		}
	}
	return expired, nil
}

func (r *InMemoryRetentionRepository) UpdateRetentionRecord(ctx context.Context, record RetentionRecord) error {
	records := r.records[record.SubjectID]
	for i, existing := range records {
		if existing.ID == record.ID {
			records[i] = record
			break
		}
	}
	return nil
}

func (r *InMemoryRetentionRepository) DeleteRetentionRecord(ctx context.Context, recordID string) error {
	for subjectID, records := range r.records {
		for i, record := range records {
			if record.ID == recordID {
				r.records[subjectID] = append(records[:i], records[i+1:]...)
				return nil
			}
		}
	}
	return fmt.Errorf("record not found: %s", recordID)
}
