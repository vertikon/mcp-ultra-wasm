package telemetry

import (
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	// Use default registry since sync.Once prevents re-registration
	Init(nil)

	// Verify metrics are accessible (not nil)
	if infRequests == nil {
		t.Error("infRequests should be initialized")
	}
	if infLatency == nil {
		t.Error("infLatency should be initialized")
	}
	if tokensIn == nil {
		t.Error("tokensIn should be initialized")
	}
	if tokensOut == nil {
		t.Error("tokensOut should be initialized")
	}
	if costBRL == nil {
		t.Error("costBRL should be initialized")
	}
	if policyBlocks == nil {
		t.Error("policyBlocks should be initialized")
	}
	if routerDecisions == nil {
		t.Error("routerDecisions should be initialized")
	}
	if budgetBreaches == nil {
		t.Error("budgetBreaches should be initialized")
	}
}

func TestObserveInference(t *testing.T) {
	Init(nil)

	start := time.Now()
	time.Sleep(10 * time.Millisecond)

	meta := InferenceMeta{
		Labels: Labels{
			TenantID: "test-tenant",
			MCPID:    "test-mcp",
			SDKName:  "test-sdk",
			Provider: "openai",
			Model:    "gpt-4o",
			UseCase:  "generation",
		},
		TokensIn:  100,
		TokensOut: 200,
		CostBRL:   0.50,
		Start:     start,
		End:       time.Now(),
	}

	// Should not panic
	ObserveInference(meta)

	// Verify metrics are not nil (init was successful)
	if infRequests == nil {
		t.Fatal("Metrics should be initialized")
	}
}

func TestIncPolicyBlock(t *testing.T) {
	Init(nil)

	labels := Labels{
		TenantID: "test-tenant",
		MCPID:    "test-mcp",
		SDKName:  "test-sdk",
		Rule:     "pii_check",
		Severity: "high",
	}

	// Should not panic
	IncPolicyBlock(labels)
	IncPolicyBlock(labels)

	// Verify metric is initialized
	if policyBlocks == nil {
		t.Fatal("Policy blocks metric should be initialized")
	}
}

func TestIncRouterDecision(t *testing.T) {
	Init(nil)

	labels := Labels{
		TenantID: "test-tenant",
		MCPID:    "test-mcp",
		SDKName:  "test-sdk",
		Provider: "openai",
		Model:    "gpt-4o",
		Reason:   "rule:default",
	}

	// Should not panic
	IncRouterDecision(labels)

	// Verify metric is initialized
	if routerDecisions == nil {
		t.Fatal("Router decisions metric should be initialized")
	}
}

func TestIncBudgetBreach(t *testing.T) {
	Init(nil)

	// Should not panic
	IncBudgetBreach("global")
	IncBudgetBreach("tenant")
	IncBudgetBreach("global")

	// Verify metric is initialized
	if budgetBreaches == nil {
		t.Fatal("Budget breaches metric should be initialized")
	}
}

func TestNoOpWhenNotInitialized(_ *testing.T) {
	// Create a new registry to isolate this test
	// Don't reset the global once - it would break other tests

	// These should not panic when metrics are not initialized
	// (they check for nil before using)
	meta := InferenceMeta{
		Labels: Labels{
			TenantID: "test",
			MCPID:    "test",
			SDKName:  "test",
			Provider: "test",
			Model:    "test",
			UseCase:  "test",
		},
		Start: time.Now(),
		End:   time.Now(),
	}

	// Save current state
	oldRequests := infRequests
	infRequests = nil

	// These should not panic
	ObserveInference(meta)
	IncPolicyBlock(Labels{})
	IncRouterDecision(Labels{})
	IncBudgetBreach("global")

	// Restore state
	infRequests = oldRequests
}
