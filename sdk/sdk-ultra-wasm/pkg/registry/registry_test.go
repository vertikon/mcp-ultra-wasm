// pkg/registry/registry_test.go
package registry_test

import (
	"net/http"
	"testing"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/contracts"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm-wasm/sdk/sdk-ultra-wasm/pkg/registry"
)

type testRoutePlugin struct{}

func (r testRoutePlugin) Name() string    { return "test-route" }
func (r testRoutePlugin) Version() string { return "1.0.0" }
func (r testRoutePlugin) Routes() []contracts.Route {
	return []contracts.Route{}
}

type testMiddlewarePlugin struct{ priority int }

func (m testMiddlewarePlugin) Name() string  { return "test-middleware" }
func (m testMiddlewarePlugin) Priority() int { return m.priority }
func (m testMiddlewarePlugin) Middleware() func(http.Handler) http.Handler {
	return nil
}

func TestRegisterAndList(t *testing.T) {
	registry.Reset()

	if err := registry.Register("test", testRoutePlugin{}); err != nil {
		t.Fatalf("Erro ao registrar plugin: %v", err)
	}

	got := registry.RouteInjectors()
	if len(got) != 1 {
		t.Fatalf("Esperava 1 route injector, obteve %d", len(got))
	}

	if got[0].Name() != "test-route" {
		t.Errorf("Esperava nome 'test-route', obteve '%s'", got[0].Name())
	}
}

func TestRegisterDuplicate(t *testing.T) {
	registry.Reset()

	if err := registry.Register("test", testRoutePlugin{}); err != nil {
		t.Fatalf("Erro ao registrar plugin: %v", err)
	}

	err := registry.Register("test", testRoutePlugin{})
	if err == nil {
		t.Error("Esperava erro ao registrar plugin duplicado")
	}
}

func TestMiddlewarePriority(t *testing.T) {
	registry.Reset()

	if err := registry.Register("high", testMiddlewarePlugin{priority: 10}); err != nil {
		t.Fatalf("Failed to register high priority middleware: %v", err)
	}
	if err := registry.Register("low", testMiddlewarePlugin{priority: 1}); err != nil {
		t.Fatalf("Failed to register low priority middleware: %v", err)
	}
	if err := registry.Register("medium", testMiddlewarePlugin{priority: 5}); err != nil {
		t.Fatalf("Failed to register medium priority middleware: %v", err)
	}

	middlewares := registry.MiddlewareInjectors()

	if len(middlewares) != 3 {
		t.Fatalf("Esperava 3 middlewares, obteve %d", len(middlewares))
	}

	// Verifica ordenação por prioridade
	if middlewares[0].Priority() != 1 {
		t.Errorf("Primeiro middleware deveria ter prioridade 1, obteve %d", middlewares[0].Priority())
	}
	if middlewares[1].Priority() != 5 {
		t.Errorf("Segundo middleware deveria ter prioridade 5, obteve %d", middlewares[1].Priority())
	}
	if middlewares[2].Priority() != 10 {
		t.Errorf("Terceiro middleware deveria ter prioridade 10, obteve %d", middlewares[2].Priority())
	}
}

func TestReset(t *testing.T) {
	registry.Reset()

	if err := registry.Register("test", testRoutePlugin{}); err != nil {
		t.Fatalf("Failed to register test plugin: %v", err)
	}

	if len(registry.RouteInjectors()) != 1 {
		t.Fatal("Plugin deveria estar registrado")
	}

	registry.Reset()

	if len(registry.RouteInjectors()) != 0 {
		t.Error("Registry deveria estar vazio após Reset()")
	}
}
