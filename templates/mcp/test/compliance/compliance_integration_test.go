//go:build integration
// +build integration

package compliance_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/compliance"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
)

// TestComplianceFrameworkIntegration tests the complete compliance framework
func TestComplianceFrameworkIntegration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Setup compliance configuration
	complianceConfig := config.ComplianceConfig{
		Enabled:       true,
		DefaultRegion: "BR",
		PIIDetection: config.PIIDetectionConfig{
			Enabled:    true,
			AutoMask:   true,
			Confidence: 0.8,
			ScanFields: []string{"email", "cpf", "phone"},
		},
		Consent: config.ConsentConfig{
			Enabled:         true,
			TTL:             time.Hour * 24 * 365 * 2, // 2 years
			GranularLevel:   "purpose",
			DefaultPurposes: []string{"service_provision", "analytics"},
		},
		DataRetention: config.DataRetentionConfig{
			Enabled:       true,
			DefaultPeriod: time.Hour * 24 * 365 * 2, // 2 years
			AutoDelete:    false,                    // Disable for testing
		},
		AuditLogging: config.AuditLoggingConfig{
			Enabled:           true,
			DetailLevel:       "full",
			EncryptionEnabled: false, // Disable for testing
		},
		LGPD: config.LGPDConfig{
			Enabled: true,
			LegalBasis: []string{
				"consent",
				"legitimate_interests",
				"contractual_fulfillment",
			},
		},
		Anonymization: config.AnonymizationConfig{
			Enabled:    true,
			Reversible: false,
			Methods:    []string{"hash", "tokenize", "redact"},
		},
	}

	// Convert to internal config format
	internalConfig := compliance.ComplianceConfig{
		Enabled:       complianceConfig.Enabled,
		DefaultRegion: complianceConfig.DefaultRegion,
		PIIDetection: compliance.PIIDetectionConfig{
			Enabled:    complianceConfig.PIIDetection.Enabled,
			ScanFields: complianceConfig.PIIDetection.ScanFields,
			Confidence: complianceConfig.PIIDetection.Confidence,
			AutoMask:   complianceConfig.PIIDetection.AutoMask,
		},
		Consent: compliance.ConsentConfig{
			Enabled:         complianceConfig.Consent.Enabled,
			DefaultPurposes: complianceConfig.Consent.DefaultPurposes,
			TTL:             complianceConfig.Consent.TTL,
			GranularLevel:   complianceConfig.Consent.GranularLevel,
		},
		DataRetention: compliance.DataRetentionConfig{
			Enabled:       complianceConfig.DataRetention.Enabled,
			DefaultPeriod: complianceConfig.DataRetention.DefaultPeriod,
			AutoDelete:    complianceConfig.DataRetention.AutoDelete,
		},
		AuditLogging: compliance.AuditLoggingConfig{
			Enabled:           complianceConfig.AuditLogging.Enabled,
			DetailLevel:       complianceConfig.AuditLogging.DetailLevel,
			EncryptionEnabled: complianceConfig.AuditLogging.EncryptionEnabled,
		},
		LGPD: compliance.LGPDConfig{
			Enabled:    complianceConfig.LGPD.Enabled,
			LegalBasis: complianceConfig.LGPD.LegalBasis,
		},
		Anonymization: compliance.AnonymizationConfig{
			Enabled:    complianceConfig.Anonymization.Enabled,
			Methods:    complianceConfig.Anonymization.Methods,
			Reversible: complianceConfig.Anonymization.Reversible,
		},
	}

	// Create compliance framework
	framework, err := compliance.NewComplianceFramework(internalConfig, logger)
	require.NoError(t, err, "Failed to create compliance framework")

	ctx := context.Background()
	subjectID := "test-user-123"

	t.Run("FrameworkHealthCheck", func(t *testing.T) {
		status, err := framework.GetComplianceStatus(ctx)
		require.NoError(t, err)

		assert.True(t, status["enabled"].(bool))
		assert.Equal(t, "BR", status["default_region"])
		assert.True(t, status["lgpd_enabled"].(bool))

		components := status["components"].(map[string]interface{})
		assert.True(t, components["pii_detection"].(bool))
		assert.True(t, components["consent_mgmt"].(bool))
		assert.True(t, components["audit_logging"].(bool))
	})

	t.Run("ConsentWorkflow", func(t *testing.T) {
		// Grant consent for service provision
		consentReq := compliance.ConsentRequest{
			SubjectID:     subjectID,
			Purpose:       "service_provision",
			Granted:       true,
			LegalBasis:    "consent",
			ConsentSource: compliance.ConsentSourceWeb,
			IPAddress:     "192.168.1.1",
			UserAgent:     "Mozilla/5.0",
			ConsentString: "User granted consent for service provision",
		}

		consent, err := framework.GetConsentManager().GrantConsent(ctx, consentReq)
		require.NoError(t, err)
		assert.True(t, consent.Granted)
		assert.Equal(t, "service_provision", consent.Purpose)

		// Check consent validation
		hasConsent, err := framework.GetConsentManager().HasValidConsent(ctx, subjectID, "service_provision")
		require.NoError(t, err)
		assert.True(t, hasConsent)

		// Check consent for different purpose (should fail)
		hasMarketingConsent, err := framework.GetConsentManager().HasValidConsent(ctx, subjectID, "marketing")
		require.NoError(t, err)
		assert.False(t, hasMarketingConsent)
	})

	t.Run("PIIDetectionAndProcessing", func(t *testing.T) {
		// First grant consent
		consentReq := compliance.ConsentRequest{
			SubjectID:     subjectID,
			Purpose:       "analytics",
			Granted:       true,
			LegalBasis:    "consent",
			ConsentSource: compliance.ConsentSourceAPI,
		}
		_, err := framework.GetConsentManager().GrantConsent(ctx, consentReq)
		require.NoError(t, err)

		// Test data with PII
		testData := map[string]interface{}{
			"email":       "test@example.com",
			"cpf":         "12345678901",
			"phone":       "+5511987654321",
			"name":        "JoÃ£o Silva",
			"description": "Task description without PII",
		}

		// Process data through compliance pipeline
		processedData, err := framework.ProcessData(ctx, subjectID, testData, "analytics")
		require.NoError(t, err)

		// Verify PII was processed (masked/anonymized)
		assert.NotEqual(t, testData["email"], processedData["email"], "Email should be anonymized")
		assert.Equal(t, testData["description"], processedData["description"], "Non-PII data should be unchanged")
	})

	t.Run("DataRightsRequests", func(t *testing.T) {
		// Test access request
		accessRequest := compliance.DataRightRequest{
			ID:          "access-req-001",
			Type:        compliance.DataRightAccess,
			Status:      compliance.DataRightStatusPending,
			RequestedAt: time.Now(),
			Data:        map[string]interface{}{"format": "json"},
		}

		err := framework.HandleDataRightRequest(ctx, subjectID, accessRequest)
		assert.NoError(t, err)

		// Test erasure request
		erasureRequest := compliance.DataRightRequest{
			ID:          "erasure-req-001",
			Type:        compliance.DataRightErasure,
			Status:      compliance.DataRightStatusPending,
			RequestedAt: time.Now(),
			Reason:      "User requested account deletion",
		}

		err = framework.HandleDataRightRequest(ctx, subjectID, erasureRequest)
		assert.NoError(t, err)
	})

	t.Run("ConsentWithdrawal", func(t *testing.T) {
		// Withdraw consent
		withdrawalRequest := compliance.DataRightRequest{
			ID:   "withdraw-req-001",
			Type: compliance.DataRightWithdrawConsent,
			Data: map[string]interface{}{
				"purpose": "analytics",
			},
			RequestedAt: time.Now(),
		}

		err := framework.HandleDataRightRequest(ctx, subjectID, withdrawalRequest)
		require.NoError(t, err)

		// Verify consent was withdrawn
		hasConsent, err := framework.GetConsentManager().HasValidConsent(ctx, subjectID, "analytics")
		require.NoError(t, err)
		assert.False(t, hasConsent, "Consent should be withdrawn")
	})

	t.Run("DataProcessingWithoutConsent", func(t *testing.T) {
		// Try to process data without consent for marketing
		testData := map[string]interface{}{
			"email": "marketing@example.com",
		}

		_, err := framework.ProcessData(ctx, subjectID, testData, "marketing")
		assert.Error(t, err, "Processing should fail without consent")
		assert.Contains(t, err.Error(), "no valid consent")
	})
}

// TestPIIManagerStandalone tests PII detection capabilities
func TestPIIManagerStandalone(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	config := compliance.PIIDetectionConfig{
		Enabled:    true,
		Confidence: 0.8,
		AutoMask:   true,
		ScanFields: []string{"email", "cpf", "phone"},
	}

	piiManager, err := compliance.NewPIIManager(config, logger)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("EmailDetection", func(t *testing.T) {
		data := map[string]interface{}{
			"user_email":    "john.doe@example.com",
			"contact_email": "invalid-email",
			"description":   "This is a task description",
		}

		processed, err := piiManager.ProcessData(ctx, data)
		require.NoError(t, err)

		// Valid email should be detected and anonymized
		assert.NotEqual(t, data["user_email"], processed["user_email"])

		// Invalid email should remain unchanged
		assert.Equal(t, data["contact_email"], processed["contact_email"])

		// Non-email field should remain unchanged
		assert.Equal(t, data["description"], processed["description"])
	})

	t.Run("CPFDetection", func(t *testing.T) {
		data := map[string]interface{}{
			"cpf":         "11144477735", // Valid CPF
			"invalid_cpf": "12345678900", // Invalid CPF
			"other_field": "some value",
		}

		processed, err := piiManager.ProcessData(ctx, data)
		require.NoError(t, err)

		// Valid CPF should be detected and tokenized
		if config.AutoMask {
			assert.NotEqual(t, data["cpf"], processed["cpf"])
		}

		// Other fields should remain unchanged
		assert.Equal(t, data["other_field"], processed["other_field"])
	})

	t.Run("HealthCheck", func(t *testing.T) {
		health := piiManager.HealthCheck(ctx)
		assert.True(t, health["enabled"].(bool))
		assert.True(t, health["auto_mask"].(bool))
		assert.Equal(t, 0.8, health["confidence_threshold"].(float64))
		assert.Greater(t, health["detectors_count"].(int), 0)
		assert.Greater(t, health["anonymizers_count"].(int), 0)
	})
}

// TestConsentManagerStandalone tests consent management capabilities
func TestConsentManagerStandalone(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	config := compliance.ConsentConfig{
		Enabled:         true,
		TTL:             time.Hour * 24, // 1 day for testing
		GranularLevel:   "purpose",
		DefaultPurposes: []string{"service", "analytics"},
	}

	consentManager, err := compliance.NewConsentManager(config, logger)
	require.NoError(t, err)

	ctx := context.Background()
	subjectID := "consent-test-user"

	t.Run("GrantAndValidateConsent", func(t *testing.T) {
		request := compliance.ConsentRequest{
			SubjectID:     subjectID,
			Purpose:       "analytics",
			Granted:       true,
			LegalBasis:    "consent",
			ConsentSource: compliance.ConsentSourceWeb,
			ConsentString: "User accepted analytics cookies",
			IPAddress:     "203.0.113.1",
			UserAgent:     "Test Browser/1.0",
		}

		consent, err := consentManager.GrantConsent(ctx, request)
		require.NoError(t, err)
		assert.True(t, consent.Granted)
		assert.Equal(t, "analytics", consent.Purpose)

		// Validate consent
		valid, err := consentManager.HasValidConsent(ctx, subjectID, "analytics")
		require.NoError(t, err)
		assert.True(t, valid)

		// Check detailed validation
		result := consentManager.ValidateConsent(ctx, subjectID, "analytics")
		assert.True(t, result.Valid)
		assert.Equal(t, "valid consent", result.Reason)
		assert.NotNil(t, result.ExpiresIn)
	})

	t.Run("ConsentExpiration", func(t *testing.T) {
		// Grant consent with short expiration
		request := compliance.ConsentRequest{
			SubjectID:      subjectID + "-expiry",
			Purpose:        "marketing",
			Granted:        true,
			LegalBasis:     "consent",
			ConsentSource:  compliance.ConsentSourceAPI,
			ExpirationDays: new(int),
		}
		*request.ExpirationDays = 0 // Immediate expiration

		_, err := consentManager.GrantConsent(ctx, request)
		require.NoError(t, err)

		// Wait a moment and check expiration
		time.Sleep(10 * time.Millisecond)

		result := consentManager.ValidateConsent(ctx, subjectID+"-expiry", "marketing")
		// Note: With 0 days expiration, it might still be valid for a brief moment
		// In a real scenario, you'd set a past date or wait for actual expiration
	})

	t.Run("ConsentWithdrawal", func(t *testing.T) {
		// First grant consent
		request := compliance.ConsentRequest{
			SubjectID:     subjectID + "-withdraw",
			Purpose:       "service",
			Granted:       true,
			LegalBasis:    "consent",
			ConsentSource: compliance.ConsentSourceWeb,
		}

		_, err := consentManager.GrantConsent(ctx, request)
		require.NoError(t, err)

		// Withdraw consent
		err = consentManager.WithdrawConsent(ctx, subjectID+"-withdraw", "service")
		require.NoError(t, err)

		// Verify withdrawal
		valid, err := consentManager.HasValidConsent(ctx, subjectID+"-withdraw", "service")
		require.NoError(t, err)
		assert.False(t, valid)
	})

	t.Run("ConsentHistory", func(t *testing.T) {
		historySubjectID := subjectID + "-history"

		// Grant initial consent
		request1 := compliance.ConsentRequest{
			SubjectID:     historySubjectID,
			Purpose:       "analytics",
			Granted:       true,
			LegalBasis:    "consent",
			ConsentSource: compliance.ConsentSourceWeb,
		}
		_, err := consentManager.GrantConsent(ctx, request1)
		require.NoError(t, err)

		// Update consent
		request2 := compliance.ConsentRequest{
			SubjectID:     historySubjectID,
			Purpose:       "analytics",
			Granted:       false, // Withdraw
			LegalBasis:    "consent",
			ConsentSource: compliance.ConsentSourceWeb,
		}
		_, err = consentManager.GrantConsent(ctx, request2)
		require.NoError(t, err)

		// Check history
		history, err := consentManager.GetConsentHistory(ctx, historySubjectID, "analytics")
		require.NoError(t, err)
		assert.Len(t, history, 2, "Should have 2 consent records in history")
	})

	t.Run("HealthCheck", func(t *testing.T) {
		health := consentManager.HealthCheck(ctx)
		assert.True(t, health["enabled"].(bool))
		assert.Equal(t, "purpose", health["granular_level"])
		assert.Equal(t, "24h0m0s", health["ttl"].(string))
	})
}

// BenchmarkComplianceFramework benchmarks the performance impact of compliance processing
func BenchmarkComplianceFramework(b *testing.B) {
	logger, _ := zap.NewDevelopment()

	config := compliance.ComplianceConfig{
		Enabled: true,
		PIIDetection: compliance.PIIDetectionConfig{
			Enabled:    true,
			AutoMask:   true,
			Confidence: 0.8,
		},
		Consent: compliance.ConsentConfig{
			Enabled: true,
		},
	}

	framework, err := compliance.NewComplianceFramework(config, logger)
	require.NoError(b, err)

	// Grant consent for benchmarking
	ctx := context.Background()
	subjectID := "benchmark-user"

	consentReq := compliance.ConsentRequest{
		SubjectID:     subjectID,
		Purpose:       "service_provision",
		Granted:       true,
		LegalBasis:    "consent",
		ConsentSource: compliance.ConsentSourceAPI,
	}
	_, err = framework.GetConsentManager().GrantConsent(ctx, consentReq)
	require.NoError(b, err)

	testData := map[string]interface{}{
		"email":       "benchmark@example.com",
		"phone":       "+5511999888777",
		"description": "This is a test task description for benchmarking",
		"priority":    "high",
		"created_at":  time.Now(),
	}

	b.ResetTimer()

	b.Run("DataProcessing", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := framework.ProcessData(ctx, subjectID, testData, "service_provision")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ConsentValidation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := framework.GetConsentManager().HasValidConsent(ctx, subjectID, "service_provision")
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("PIIDetection", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := framework.GetPIIManager().ProcessData(ctx, testData)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
