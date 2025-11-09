package wiring

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestInit_AIDisabled(t *testing.T) {
	// Create temporary test directory with disabled AI
	tmpDir := t.TempDir()
	aiDir := filepath.Join(tmpDir, "ai")
	if err := os.MkdirAll(aiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create feature_flags.json with AI disabled
	flagsContent := `{"ai":{"enabled":false}}`
	if err := os.WriteFile(filepath.Join(aiDir, "feature_flags.json"), []byte(flagsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize AI service
	reg := prometheus.NewRegistry()
	svc, err := Init(context.Background(), Config{
		BasePathAI: aiDir,
		Registry:   reg,
	})

	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if svc == nil {
		t.Fatal("Service should not be nil")
	}

	if svc.Enabled {
		t.Error("AI should be disabled")
	}

	if svc.Router == nil {
		t.Error("Router should be initialized even when disabled")
	}
}

func TestInit_AIEnabled(t *testing.T) {
	// Create temporary test directory with enabled AI
	tmpDir := t.TempDir()
	aiDir := filepath.Join(tmpDir, "ai")
	configDir := filepath.Join(aiDir, "config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create feature_flags.json with AI enabled
	flagsContent := `{"ai":{"enabled":true,"mode":"balanced","router":"rules"}}`
	if err := os.WriteFile(filepath.Join(aiDir, "feature_flags.json"), []byte(flagsContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create ai-router.rules.json
	rulesContent := `{
		"version": "1.0",
		"default": {
			"classification": {"provider": "openai", "model": "gpt-4o-mini"},
			"generation": {"provider": "openai", "model": "gpt-4o"}
		},
		"overrides": [],
		"fallbacks": []
	}`
	if err := os.WriteFile(filepath.Join(configDir, "ai-router.rules.json"), []byte(rulesContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize AI service
	reg := prometheus.NewRegistry()
	svc, err := Init(context.Background(), Config{
		BasePathAI: aiDir,
		Registry:   reg,
	})

	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	if svc == nil {
		t.Fatal("Service should not be nil")
	}

	if !svc.Enabled {
		t.Error("AI should be enabled")
	}

	if svc.Router == nil {
		t.Fatal("Router should be initialized")
	}

	// Test router decision
	dec, err := svc.Router.Decide("generation")
	if err != nil {
		t.Fatalf("Decide failed: %v", err)
	}

	if dec.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", dec.Provider)
	}

	if dec.Model != "gpt-4o" {
		t.Errorf("Expected model 'gpt-4o', got '%s'", dec.Model)
	}

	if dec.Reason != "rule:default" {
		t.Errorf("Expected reason 'rule:default', got '%s'", dec.Reason)
	}
}

func TestInit_MissingConfig(t *testing.T) {
	// Initialize with non-existent directory
	reg := prometheus.NewRegistry()
	svc, err := Init(context.Background(), Config{
		BasePathAI: "/non/existent/path",
		Registry:   reg,
	})

	// Should not error, just return disabled service
	if err != nil {
		t.Fatalf("Init should not fail with missing config: %v", err)
	}

	if svc == nil {
		t.Fatal("Service should not be nil")
	}

	if svc.Enabled {
		t.Error("AI should be disabled when config is missing")
	}
}
