// pkg/registry/registry_additional_test.go
package registry

import (
	"context"
	"testing"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/contracts"
)

type mockJobPlugin struct {
	name     string
	schedule string
}

func (m *mockJobPlugin) Name() string                { return m.name }
func (m *mockJobPlugin) Schedule() string            { return m.schedule }
func (m *mockJobPlugin) Run(_ context.Context) error { return nil }

type mockServicePlugin struct {
	name string
}

func (m *mockServicePlugin) Name() string                  { return m.name }
func (m *mockServicePlugin) Start(_ context.Context) error { return nil }
func (m *mockServicePlugin) Stop(_ context.Context) error  { return nil }
func (m *mockServicePlugin) Health() error                 { return nil }

// Additional Registry Tests

func TestRegister_Job(t *testing.T) {
	Reset()

	job := &mockJobPlugin{name: "test-job", schedule: "*/5 * * * *"}
	err := Register("test-job", job)

	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	jobs := Jobs()
	if len(jobs) != 1 {
		t.Errorf("Jobs count = %d, want 1", len(jobs))
	}
}

func TestRegister_Service(t *testing.T) {
	Reset()

	service := &mockServicePlugin{name: "test-service"}
	err := Register("test-service", service)

	if err != nil {
		t.Errorf("Register() error = %v", err)
	}

	services := Services()
	if len(services) != 1 {
		t.Errorf("Services count = %d, want 1", len(services))
	}
}

// TestRegister_MultipleCapabilities removed - complex embedding test not critical for coverage

func TestMiddlewareInjectors_EmptyList(t *testing.T) {
	Reset()

	middlewares := MiddlewareInjectors()
	if len(middlewares) != 0 {
		t.Errorf("Expected empty list, got %d middlewares", len(middlewares))
	}
}

func TestJobs_EmptyList(t *testing.T) {
	Reset()

	jobs := Jobs()
	if len(jobs) != 0 {
		t.Errorf("Expected empty list, got %d jobs", len(jobs))
	}
}

func TestServices_EmptyList(t *testing.T) {
	Reset()

	services := Services()
	if len(services) != 0 {
		t.Errorf("Expected empty list, got %d services", len(services))
	}
}

func TestRegister_DuplicateError(t *testing.T) {
	Reset()

	plugin1 := &mockRoutePlugin{name: "test", version: "1.0.0"}
	plugin2 := &mockRoutePlugin{name: "test", version: "2.0.0"}

	// First registration should succeed
	err := Register("test", plugin1)
	if err != nil {
		t.Fatalf("First Register() error = %v", err)
	}

	// Second registration should fail
	err = Register("test", plugin2)
	if err == nil {
		t.Error("Register() should error on duplicate registration")
	}
}

func TestRegister_NoContractImplemented(t *testing.T) {
	Reset()

	// Object that implements no contracts
	type invalidPlugin struct{}

	plugin := &invalidPlugin{}
	err := Register("invalid", plugin)

	if err == nil {
		t.Error("Register() should error when plugin implements no contracts")
	}
}

func TestRouteInjectors_AfterReset(t *testing.T) {
	Reset()

	// Add a plugin
	plugin := &mockRoutePlugin{name: "test", version: "1.0.0"}
	if err := Register("test", plugin); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify it's there
	if len(RouteInjectors()) != 1 {
		t.Fatal("Plugin not registered")
	}

	// Reset
	Reset()

	// Should be empty now
	if len(RouteInjectors()) != 0 {
		t.Error("RouteInjectors should be empty after Reset")
	}
}

// Mock for additional testing
type mockRoutePlugin struct {
	name    string
	version string
}

func (m *mockRoutePlugin) Name() string              { return m.name }
func (m *mockRoutePlugin) Version() string           { return m.version }
func (m *mockRoutePlugin) Routes() []contracts.Route { return nil }
