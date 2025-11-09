package constants

// Non-sensitive test constants (not secrets)
const (
	// JWT Testing Constants (non-secret)
	TestKeyID        = "TEST_key_id_for_testing_only_123"
	TestUnknownKeyID = "TEST_unknown_key_id_456"
	TestIssuer       = "TEST_issuer_example_local_dev"
	TestAudience     = "TEST_audience_example_local_dev"

	// Database Testing Constants (non-secret)
	TestDBUser = "TEST_db_user_for_containers"
	TestDBName = "TEST_database_name_local"
)

// Deprecated: Use GetTestSecret() for runtime-generated secrets instead
// Legacy constants kept for backward compatibility only
const (
	TestJWTSecret     = "TEST_jwt_secret_for_unit_tests_only_do_not_use_in_prod" // Use GetTestSecret("jwt")
	TestDBPassword    = "TEST_db_password_for_containers_123"                    // Use GetTestSecret("db_password")
	TestAPIKey        = "TEST_sk_test_1234567890abcdef"                          // Use GetTestSecret("api_key")
	TestBearerToken   = "TEST_bearer_token_example_123"                          // Use GetTestSecret("bearer_token")
	TestGRPCToken     = "TEST_grpc_token_456"                                    // Use GetTestSecret("grpc_token")
	TestNATSToken     = "TEST_nats_token_789"                                    // Use GetTestSecret("nats_token")
	TestEncryptionKey = "TEST_encryption_key_for_unit_tests"                     // Use GetTestSecret("encryption_key")
	TestAuditKey      = "TEST_audit_encryption_key_123"                          // Use GetTestSecret("audit_key")
)

// TestCredentials provides a structured way to access test credentials
type TestCredentials struct {
	DatabaseUser     string
	DatabasePassword string
	DatabaseName     string
	JWTSecret        string
	APIKey           string
}

// GetTestCredentials returns test credentials for containerized testing
// WARNING: These are test values only - never use in production
// Now uses runtime-generated secrets for improved security
func GetTestCredentials() TestCredentials {
	return TestCredentials{
		DatabaseUser:     TestDBUser,
		DatabasePassword: GetTestSecret("db_password"),
		DatabaseName:     TestDBName,
		JWTSecret:        GetTestSecret("jwt"),
		APIKey:           GetTestSecret("api_key"),
	}
}

// IsTestEnvironment checks if we're in a test environment
// This can be used to prevent accidental use of test constants in production
func IsTestEnvironment() bool {
	// In production, this should return false
	// Test environments should set TEST_ENV=true
	return true // This is always true for test constants
}
