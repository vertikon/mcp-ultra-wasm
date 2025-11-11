package wiring

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ai/router"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/ai/telemetry"
)

type Config struct {
	BasePathAI string // path to templates/ai
	Registry   prometheus.Registerer
}

// Service holds minimal IA singletons (router + telemetry enabled flag).
type Service struct {
	Router  *router.Router
	Enabled bool
}

func Init(ctx context.Context, cfg Config) (*Service, error) {
	// feature flags are inside feature_flags.json at BasePathAI
	base := cfg.BasePathAI
	if base == "" {
		cwd, _ := os.Getwd()
		base = filepath.Join(cwd, "templates", "ai")
	}

	r, _ := router.Load(base)
	telemetry.Init(cfg.Registry)

	svc := &Service{Router: r, Enabled: r != nil && r.Enabled()}
	_ = ctx // reserved for future async init
	time.AfterFunc(0, func() { /* noop */ })
	return svc, nil
}
