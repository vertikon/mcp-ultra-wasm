package events

import (
	"context"
	"encoding/json"
	"time"
)

// EventPublisher is an adapter for the existing NATS publisher abstraction.
// It matches a method like natsx.Publisher.PublishWithRetry(ctx, subject, payload).
type EventPublisher interface {
	PublishWithRetry(ctx context.Context, subject string, payload []byte) error
}

type Base struct {
	TenantID string `json:"tenant_id"`
	MCPID    string `json:"mcp_id"`
	SDKName  string `json:"sdk_name,omitempty"`
	Ts       string `json:"timestamp"`
}

type RouterDecision struct {
	Base
	UseCase  string `json:"use_case"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Reason   string `json:"reason"`
}

type PolicyBlock struct {
	Base
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Sample   string `json:"sample,omitempty"`
}

type InferenceError struct {
	Base
	Provider string `json:"provider,omitempty"`
	Model    string `json:"model,omitempty"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type InferenceSummary struct {
	Base
	UseCase   string  `json:"use_case"`
	TokensIn  int     `json:"tokens_in"`
	TokensOut int     `json:"tokens_out"`
	LatencyMs int     `json:"latency_ms"`
	CostBRL   float64 `json:"cost_brl"`
	Cached    bool    `json:"cached"`
}

func now() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func PublishRouterDecision(ctx context.Context, pub EventPublisher, subject string, e RouterDecision) error {
	e.Ts = now()
	b, _ := json.Marshal(e)
	return pub.PublishWithRetry(ctx, subject, b)
}

func PublishPolicyBlock(ctx context.Context, pub EventPublisher, subject string, e PolicyBlock) error {
	e.Ts = now()
	b, _ := json.Marshal(e)
	return pub.PublishWithRetry(ctx, subject, b)
}

func PublishInferenceError(ctx context.Context, pub EventPublisher, subject string, e InferenceError) error {
	e.Ts = now()
	b, _ := json.Marshal(e)
	return pub.PublishWithRetry(ctx, subject, b)
}

func PublishInferenceSummary(ctx context.Context, pub EventPublisher, subject string, e InferenceSummary) error {
	e.Ts = now()
	b, _ := json.Marshal(e)
	return pub.PublishWithRetry(ctx, subject, b)
}
