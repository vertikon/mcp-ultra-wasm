package events

import (
	"context"
	"encoding/json"
	"testing"
)

// Mock publisher for testing
type mockPublisher struct {
	published []publishedEvent
}

type publishedEvent struct {
	subject string
	payload []byte
}

func (m *mockPublisher) PublishWithRetry(_ context.Context, subject string, payload []byte) error {
	m.published = append(m.published, publishedEvent{
		subject: subject,
		payload: payload,
	})
	return nil
}

func TestPublishRouterDecision(t *testing.T) {
	mock := &mockPublisher{}
	ctx := context.Background()

	event := RouterDecision{
		Base: Base{
			TenantID: "test-tenant",
			MCPID:    "test-mcp",
			SDKName:  "test-sdk",
		},
		UseCase:  "generation",
		Provider: "openai",
		Model:    "gpt-4o",
		Reason:   "rule:default",
	}

	err := PublishRouterDecision(ctx, mock, "ultra.ai.router.decision", event)
	if err != nil {
		t.Fatalf("PublishRouterDecision failed: %v", err)
	}

	if len(mock.published) != 1 {
		t.Fatalf("Expected 1 published event, got %d", len(mock.published))
	}

	pub := mock.published[0]
	if pub.subject != "ultra.ai.router.decision" {
		t.Errorf("Expected subject 'ultra.ai.router.decision', got '%s'", pub.subject)
	}

	// Unmarshal and verify
	var decoded RouterDecision
	if err := json.Unmarshal(pub.payload, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if decoded.TenantID != "test-tenant" {
		t.Errorf("Expected TenantID 'test-tenant', got '%s'", decoded.TenantID)
	}

	if decoded.UseCase != "generation" {
		t.Errorf("Expected UseCase 'generation', got '%s'", decoded.UseCase)
	}

	if decoded.Provider != "openai" {
		t.Errorf("Expected Provider 'openai', got '%s'", decoded.Provider)
	}

	if decoded.Ts == "" {
		t.Error("Timestamp should be set")
	}
}

func TestPublishPolicyBlock(t *testing.T) {
	mock := &mockPublisher{}
	ctx := context.Background()

	event := PolicyBlock{
		Base: Base{
			TenantID: "test-tenant",
			MCPID:    "test-mcp",
			SDKName:  "test-sdk",
		},
		Rule:     "pii_check",
		Severity: "high",
		Sample:   "sensitive data sample",
	}

	err := PublishPolicyBlock(ctx, mock, "ultra.ai.policy.block", event)
	if err != nil {
		t.Fatalf("PublishPolicyBlock failed: %v", err)
	}

	if len(mock.published) != 1 {
		t.Fatalf("Expected 1 published event, got %d", len(mock.published))
	}

	pub := mock.published[0]
	if pub.subject != "ultra.ai.policy.block" {
		t.Errorf("Expected subject 'ultra.ai.policy.block', got '%s'", pub.subject)
	}

	var decoded PolicyBlock
	if err := json.Unmarshal(pub.payload, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if decoded.Rule != "pii_check" {
		t.Errorf("Expected Rule 'pii_check', got '%s'", decoded.Rule)
	}

	if decoded.Severity != "high" {
		t.Errorf("Expected Severity 'high', got '%s'", decoded.Severity)
	}
}

func TestPublishInferenceError(t *testing.T) {
	mock := &mockPublisher{}
	ctx := context.Background()

	event := InferenceError{
		Base: Base{
			TenantID: "test-tenant",
			MCPID:    "test-mcp",
		},
		Provider: "openai",
		Model:    "gpt-4o",
		Code:     "RATE_LIMIT",
		Message:  "Rate limit exceeded",
	}

	err := PublishInferenceError(ctx, mock, "ultra.ai.inference.error", event)
	if err != nil {
		t.Fatalf("PublishInferenceError failed: %v", err)
	}

	if len(mock.published) != 1 {
		t.Fatalf("Expected 1 published event, got %d", len(mock.published))
	}

	var decoded InferenceError
	if err := json.Unmarshal(mock.published[0].payload, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if decoded.Code != "RATE_LIMIT" {
		t.Errorf("Expected Code 'RATE_LIMIT', got '%s'", decoded.Code)
	}

	if decoded.Message != "Rate limit exceeded" {
		t.Errorf("Expected Message 'Rate limit exceeded', got '%s'", decoded.Message)
	}
}

func TestPublishInferenceSummary(t *testing.T) {
	mock := &mockPublisher{}
	ctx := context.Background()

	event := InferenceSummary{
		Base: Base{
			TenantID: "test-tenant",
			MCPID:    "test-mcp",
			SDKName:  "test-sdk",
		},
		UseCase:   "generation",
		TokensIn:  1000,
		TokensOut: 500,
		LatencyMs: 1234,
		CostBRL:   0.25,
		Cached:    false,
	}

	err := PublishInferenceSummary(ctx, mock, "ultra.ai.inference.summary", event)
	if err != nil {
		t.Fatalf("PublishInferenceSummary failed: %v", err)
	}

	if len(mock.published) != 1 {
		t.Fatalf("Expected 1 published event, got %d", len(mock.published))
	}

	var decoded InferenceSummary
	if err := json.Unmarshal(mock.published[0].payload, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if decoded.TokensIn != 1000 {
		t.Errorf("Expected TokensIn 1000, got %d", decoded.TokensIn)
	}

	if decoded.TokensOut != 500 {
		t.Errorf("Expected TokensOut 500, got %d", decoded.TokensOut)
	}

	if decoded.CostBRL != 0.25 {
		t.Errorf("Expected CostBRL 0.25, got %f", decoded.CostBRL)
	}

	if decoded.Cached {
		t.Error("Expected Cached false, got true")
	}
}

func TestMultiplePublishes(t *testing.T) {
	mock := &mockPublisher{}
	ctx := context.Background()

	// Publish router decision
	_ = PublishRouterDecision(ctx, mock, "ultra.ai.router.decision", RouterDecision{
		Base:     Base{TenantID: "t1", MCPID: "m1"},
		UseCase:  "generation",
		Provider: "openai",
		Model:    "gpt-4o",
		Reason:   "rule",
	})

	// Publish policy block
	_ = PublishPolicyBlock(ctx, mock, "ultra.ai.policy.block", PolicyBlock{
		Base:     Base{TenantID: "t1", MCPID: "m1"},
		Rule:     "pii",
		Severity: "medium",
	})

	// Publish inference summary
	_ = PublishInferenceSummary(ctx, mock, "ultra.ai.inference.summary", InferenceSummary{
		Base:      Base{TenantID: "t1", MCPID: "m1"},
		UseCase:   "generation",
		TokensIn:  100,
		TokensOut: 50,
		LatencyMs: 500,
		CostBRL:   0.10,
	})

	if len(mock.published) != 3 {
		t.Errorf("Expected 3 published events, got %d", len(mock.published))
	}

	// Verify subjects
	expectedSubjects := []string{
		"ultra.ai.router.decision",
		"ultra.ai.policy.block",
		"ultra.ai.inference.summary",
	}

	for i, expected := range expectedSubjects {
		if mock.published[i].subject != expected {
			t.Errorf("Event %d: expected subject '%s', got '%s'", i, expected, mock.published[i].subject)
		}
	}
}
