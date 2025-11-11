package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vertikon/mcp-ultra-templates/internal/config"
)

func TestAppContextCancel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		TemplatesPath: t.TempDir(),
		Observability: config.ObservabilityConfig{
			EnableMetrics: false,
			EnableTracing: false,
			MetricsNS:     "test",
		},
	}
	app := NewApp(cfg)

	ctx, cancel := app.Context()
	cancel()

	select {
	case <-ctx.Done():
	case <-time.After(500 * time.Millisecond):
		t.Fatal("context was not canceled")
	}
}

func TestAppLogger(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		TemplatesPath: t.TempDir(),
		Observability: config.ObservabilityConfig{
			EnableMetrics: false,
			EnableTracing: false,
			MetricsNS:     "test",
		},
	}
	app := NewApp(cfg)

	logger := app.Logger()
	require.NotNil(t, &logger)
	app.TemplateService()
}
