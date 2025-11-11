package security

// This file was removed to eliminate duplication with auth_test.go
// All test functions (TestNewAuthService, TestGetUserFromContext, TestRequireScope, TestRequireRole)
// and the MockOPAService mock were already defined in auth_test.go
//
// Keeping this file prevented compilation due to:
// - MockOPAService redeclared
// - TestNewAuthService redeclared
// - TestGetUserFromContext redeclared
// - TestRequireScope redeclared
// - TestRequireRole redeclared
//
// All tests have been consolidated in auth_test.go
