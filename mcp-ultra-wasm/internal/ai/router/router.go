package router

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

type Flags struct {
	AI struct {
		Enabled       bool   `json:"enabled"`
		Mode          string `json:"mode"`
		CanaryPercent int    `json:"canary_percent"`
		Router        string `json:"router"`
	} `json:"ai"`
}

type Rule struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

type Rules struct {
	Version   string          `json:"version"`
	Default   map[string]Rule `json:"default"` // use_case -> rule
	Overrides []any           `json:"overrides"`
	Fallbacks []struct {
		From Rule `json:"from"`
		To   Rule `json:"to"`
	} `json:"fallbacks"`
}

type Decision struct {
	Provider string
	Model    string
	Reason   string
}

type Router struct {
	flags Flags
	rules Rules
	mu    sync.RWMutex
}

func Load(basePath string) (*Router, error) {
	r := &Router{}
	ff := filepath.Join(basePath, "feature_flags.json")
	rules := filepath.Join(basePath, "config", "ai-router.rules.json")

	if b, err := os.ReadFile(ff); err == nil {
		_ = json.Unmarshal(b, &r.flags)
	}
	if b, err := os.ReadFile(rules); err == nil {
		_ = json.Unmarshal(b, &r.rules)
	}
	return r, nil
}

func (r *Router) Enabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.flags.AI.Enabled
}

func (r *Router) Decide(useCase string) (Decision, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.flags.AI.Enabled {
		return Decision{}, errors.New("ai disabled")
	}
	if rule, ok := r.rules.Default[useCase]; ok {
		return Decision{Provider: rule.Provider, Model: rule.Model, Reason: "rule:default"}, nil
	}
	if rule, ok := r.rules.Default["generation"]; ok {
		return Decision{Provider: rule.Provider, Model: rule.Model, Reason: "fallback:generation"}, nil
	}
	return Decision{}, errors.New("no rule found")
}
