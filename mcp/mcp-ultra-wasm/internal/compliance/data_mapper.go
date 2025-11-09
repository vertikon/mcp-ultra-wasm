package compliance

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// DataMapper provides data mapping and discovery for compliance
type DataMapper struct {
	config    Config
	logger    *zap.Logger
	dataMap   map[string]DataMapping
	inventory []DataInventoryItem
}

// DataMapping represents how data flows through the system
type DataMapping struct {
	FieldName       string               `json:"field_name"`
	DataType        DataType             `json:"data_type"`
	PIIType         PIIType              `json:"pii_type,omitempty"`
	Sensitivity     PIISensitivity       `json:"sensitivity"`
	LegalBasis      LegalBasis           `json:"legal_basis"`
	Purpose         []string             `json:"purpose"`
	Retention       RetentionRule        `json:"retention"`
	Sources         []DataSource         `json:"sources"`
	Destinations    []DataDestination    `json:"destinations"`
	Transformations []DataTransformation `json:"transformations"`
	AccessPatterns  []AccessPattern      `json:"access_patterns"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
}

// DataType represents different types of data
type DataType string

const (
	DataTypeString    DataType = "string"
	DataTypeInteger   DataType = "integer"
	DataTypeFloat     DataType = "float"
	DataTypeBoolean   DataType = "boolean"
	DataTypeDate      DataType = "date"
	DataTypeTimestamp DataType = "timestamp"
	DataTypeJSON      DataType = "json"
	DataTypeBinary    DataType = "binary"
)

// DataSource represents where data originates
type DataSource struct {
	Type        string            `json:"type"` // api, form, import, etc.
	Name        string            `json:"name"`
	Location    string            `json:"location"`
	Metadata    map[string]string `json:"metadata"`
	CollectedAt time.Time         `json:"collected_at"`
}

// DataDestination represents where data is sent or stored
type DataDestination struct {
	Type       string            `json:"type"` // database, cache, export, api, etc.
	Name       string            `json:"name"`
	Location   string            `json:"location"`
	Purpose    string            `json:"purpose"`
	Metadata   map[string]string `json:"metadata"`
	AccessedAt time.Time         `json:"accessed_at"`
}

// DataTransformation represents how data is transformed
type DataTransformation struct {
	Type      string            `json:"type"` // anonymize, encrypt, aggregate, etc.
	Method    string            `json:"method"`
	Applied   bool              `json:"applied"`
	Metadata  map[string]string `json:"metadata"`
	AppliedAt time.Time         `json:"applied_at"`
}

// AccessPattern represents how data is accessed
type AccessPattern struct {
	Actor       string    `json:"actor"`     // user, system, service
	Action      string    `json:"action"`    // read, write, delete, export
	Frequency   string    `json:"frequency"` // once, daily, weekly, etc.
	Purpose     string    `json:"purpose"`
	LastAccess  time.Time `json:"last_access"`
	AccessCount int       `json:"access_count"`
}

// RetentionRule defines how long data should be retained
type RetentionRule struct {
	Category        string        `json:"category"`
	RetentionPeriod time.Duration `json:"retention_period"`
	DeleteAfter     time.Duration `json:"delete_after"`
	ArchiveAfter    time.Duration `json:"archive_after,omitempty"`
	LegalHold       bool          `json:"legal_hold"`
	Justification   string        `json:"justification"`
}

// DataInventoryItem represents an item in the data inventory
type DataInventoryItem struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Category         string            `json:"category"`
	Owner            string            `json:"owner"`
	Steward          string            `json:"steward"`
	Location         string            `json:"location"`
	Format           string            `json:"format"`
	Volume           int64             `json:"volume"`
	PIIFields        []string          `json:"pii_fields"`
	SensitivityLevel PIISensitivity    `json:"sensitivity_level"`
	RetentionPolicy  string            `json:"retention_policy"`
	BackupLocations  []string          `json:"backup_locations"`
	AccessControls   []string          `json:"access_controls"`
	EncryptionStatus string            `json:"encryption_status"`
	LastAudit        time.Time         `json:"last_audit"`
	Metadata         map[string]string `json:"metadata"`
}

// NewDataMapper creates a new data mapper
func NewDataMapper(config Config, logger *zap.Logger) (*DataMapper, error) {
	dm := &DataMapper{
		config:    config,
		logger:    logger,
		dataMap:   make(map[string]DataMapping),
		inventory: make([]DataInventoryItem, 0),
	}

	// Initialize with default mappings
	dm.initializeDefaultMappings()

	return dm, nil
}

// MapDataField maps a data field with its compliance metadata
func (dm *DataMapper) MapDataField(_ context.Context, fieldName string, metadata DataMapping) error {
	metadata.FieldName = fieldName
	metadata.UpdatedAt = time.Now()

	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = time.Now()
	}

	dm.dataMap[fieldName] = metadata

	dm.logger.Debug("Data field mapped",
		zap.String("field", fieldName),
		zap.String("pii_type", string(metadata.PIIType)),
		zap.String("sensitivity", string(metadata.Sensitivity)))

	return nil
}

// GetDataMapping retrieves the mapping for a specific field
func (dm *DataMapper) GetDataMapping(fieldName string) (*DataMapping, bool) {
	mapping, exists := dm.dataMap[fieldName]
	return &mapping, exists
}

// GetAllMappings returns all data mappings
func (dm *DataMapper) GetAllMappings() map[string]DataMapping {
	return dm.dataMap
}

// DiscoverDataSources automatically discovers data sources and their mappings
func (dm *DataMapper) DiscoverDataSources(_ context.Context) error {
	// In a real implementation, this would scan databases, APIs, files, etc.
	// to automatically discover data sources and create mappings

	dm.logger.Info("Starting data source discovery")

	// Mock discovery process
	discoveredSources := []DataInventoryItem{
		{
			ID:               "tasks_table",
			Name:             "Tasks Table",
			Description:      "Main tasks table in PostgreSQL",
			Category:         "operational_data",
			Owner:            "data_team",
			Location:         "postgresql://tasks",
			Format:           "structured",
			Volume:           1000000,
			PIIFields:        []string{"assignee_email", "description"},
			SensitivityLevel: PIISensitivityConfidential,
			RetentionPolicy:  "2_years",
			EncryptionStatus: "encrypted_at_rest",
			LastAudit:        time.Now().AddDate(0, -1, 0),
		},
		{
			ID:               "user_cache",
			Name:             "User Cache",
			Description:      "Redis cache for user data",
			Category:         "cache_data",
			Owner:            "platform_team",
			Location:         "redis://user_cache",
			Format:           "key_value",
			Volume:           500000,
			PIIFields:        []string{"user_email", "user_name"},
			SensitivityLevel: PIISensitivityConfidential,
			RetentionPolicy:  "30_days",
			EncryptionStatus: "encrypted_in_transit",
			LastAudit:        time.Now().AddDate(0, 0, -15),
		},
	}

	dm.inventory = append(dm.inventory, discoveredSources...)

	dm.logger.Info("Data source discovery completed",
		zap.Int("sources_discovered", len(discoveredSources)))

	return nil
}

// GenerateDataMap generates a comprehensive data map
func (dm *DataMapper) GenerateDataMap(_ context.Context) (map[string]interface{}, error) {
	dataMap := map[string]interface{}{
		"generated_at":    time.Now(),
		"total_fields":    len(dm.dataMap),
		"inventory_items": len(dm.inventory),
		"mappings":        dm.dataMap,
		"inventory":       dm.inventory,
		"statistics":      dm.generateStatistics(),
	}

	return dataMap, nil
}

// TrackDataFlow tracks how data flows through the system
func (dm *DataMapper) TrackDataFlow(_ context.Context, fieldName string, source DataSource, destination DataDestination) error {
	mapping, exists := dm.dataMap[fieldName]
	if !exists {
		// Create new mapping
		mapping = DataMapping{
			FieldName:    fieldName,
			Sources:      []DataSource{source},
			Destinations: []DataDestination{destination},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
	} else {
		// Update existing mapping
		mapping.Sources = append(mapping.Sources, source)
		mapping.Destinations = append(mapping.Destinations, destination)
		mapping.UpdatedAt = time.Now()
	}

	dm.dataMap[fieldName] = mapping

	dm.logger.Debug("Data flow tracked",
		zap.String("field", fieldName),
		zap.String("source", source.Name),
		zap.String("destination", destination.Name))

	return nil
}

// RecordDataAccess records access to data for compliance tracking
func (dm *DataMapper) RecordDataAccess(_ context.Context, fieldName, actor, action, purpose string) error {
	mapping, exists := dm.dataMap[fieldName]
	if !exists {
		dm.logger.Warn("Attempted to record access for unmapped field", zap.String("field", fieldName))
		return fmt.Errorf("field not mapped: %s", fieldName)
	}

	// Find existing access pattern or create new one
	var accessPattern *AccessPattern
	for i := range mapping.AccessPatterns {
		if mapping.AccessPatterns[i].Actor == actor && mapping.AccessPatterns[i].Action == action {
			accessPattern = &mapping.AccessPatterns[i]
			break
		}
	}

	if accessPattern == nil {
		// Create new access pattern
		newPattern := AccessPattern{
			Actor:       actor,
			Action:      action,
			Purpose:     purpose,
			LastAccess:  time.Now(),
			AccessCount: 1,
		}
		mapping.AccessPatterns = append(mapping.AccessPatterns, newPattern)
	} else {
		// Update existing pattern
		accessPattern.LastAccess = time.Now()
		accessPattern.AccessCount++
	}

	mapping.UpdatedAt = time.Now()
	dm.dataMap[fieldName] = mapping

	return nil
}

// ValidateCompliance validates compliance for mapped data
func (dm *DataMapper) ValidateCompliance(_ context.Context) ([]Violation, error) {
	violations := make([]Violation, 0)

	for fieldName, mapping := range dm.dataMap {
		// Check if PII fields have proper legal basis
		if mapping.PIIType != "" && len(mapping.LegalBasis) == 0 {
			violations = append(violations, Violation{
				Type:        "missing_legal_basis",
				Field:       fieldName,
				Severity:    "high",
				Description: "PII field without legal basis",
				DetectedAt:  time.Now(),
			})
		}

		// Check retention policies
		if mapping.Retention.RetentionPeriod == 0 {
			violations = append(violations, Violation{
				Type:        "missing_retention_policy",
				Field:       fieldName,
				Severity:    "medium",
				Description: "Field without retention policy",
				DetectedAt:  time.Now(),
			})
		}

		// Check for unencrypted sensitive data
		if mapping.Sensitivity == PIISensitivityRestricted || mapping.Sensitivity == PIISensitivityConfidential {
			hasEncryption := false
			for _, transform := range mapping.Transformations {
				if transform.Type == "encrypt" {
					hasEncryption = true
					break
				}
			}
			if !hasEncryption {
				violations = append(violations, Violation{
					Type:        "unencrypted_sensitive_data",
					Field:       fieldName,
					Severity:    "critical",
					Description: "Sensitive data without encryption",
					DetectedAt:  time.Now(),
				})
			}
		}
	}

	return violations, nil
}

// Violation represents a compliance violation
type Violation struct {
	Type           string    `json:"type"`
	Field          string    `json:"field"`
	Severity       string    `json:"severity"`
	Description    string    `json:"description"`
	Recommendation string    `json:"recommendation,omitempty"`
	DetectedAt     time.Time `json:"detected_at"`
}

// Helper methods

func (dm *DataMapper) initializeDefaultMappings() {
	// Initialize common field mappings
	defaultMappings := map[string]DataMapping{
		"email": {
			FieldName:   "email",
			DataType:    DataTypeString,
			PIIType:     PIITypeEmail,
			Sensitivity: PIISensitivityConfidential,
			Purpose:     []string{"authentication", "communication"},
			Retention: RetentionRule{
				Category:        "user_data",
				RetentionPeriod: time.Hour * 24 * 365 * 2, // 2 years
				DeleteAfter:     time.Hour * 24 * 365 * 7, // 7 years
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		"phone": {
			FieldName:   "phone",
			DataType:    DataTypeString,
			PIIType:     PIITypePhone,
			Sensitivity: PIISensitivityConfidential,
			Purpose:     []string{"communication", "verification"},
			Retention: RetentionRule{
				Category:        "contact_data",
				RetentionPeriod: time.Hour * 24 * 365 * 2, // 2 years
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		"cpf": {
			FieldName:   "cpf",
			DataType:    DataTypeString,
			PIIType:     PIITypeCPF,
			Sensitivity: PIISensitivityRestricted,
			Purpose:     []string{"identification", "compliance"},
			Retention: RetentionRule{
				Category:        "identity_data",
				RetentionPeriod: time.Hour * 24 * 365 * 5, // 5 years
				LegalHold:       true,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for fieldName, mapping := range defaultMappings {
		dm.dataMap[fieldName] = mapping
	}

	dm.logger.Info("Default data mappings initialized", zap.Int("mappings", len(defaultMappings)))
}

func (dm *DataMapper) generateStatistics() map[string]interface{} {
	stats := map[string]interface{}{
		"total_mappings":    len(dm.dataMap),
		"pii_fields":        0,
		"by_sensitivity":    map[string]int{},
		"by_pii_type":       map[string]int{},
		"retention_periods": map[string]int{},
	}

	for _, mapping := range dm.dataMap {
		if mapping.PIIType != "" {
			stats["pii_fields"] = stats["pii_fields"].(int) + 1
		}

		// Count by sensitivity
		sensitivityCounts := stats["by_sensitivity"].(map[string]int)
		sensitivityCounts[string(mapping.Sensitivity)]++

		// Count by PII type
		if mapping.PIIType != "" {
			piiTypeCounts := stats["by_pii_type"].(map[string]int)
			piiTypeCounts[string(mapping.PIIType)]++
		}

		// Count retention periods
		retentionCounts := stats["retention_periods"].(map[string]int)
		if mapping.Retention.RetentionPeriod > 0 {
			period := mapping.Retention.RetentionPeriod.String()
			retentionCounts[period]++
		}
	}

	return stats
}
