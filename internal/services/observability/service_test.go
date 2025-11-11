package observability

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/vertikon/mcp-ultra-templates/internal/config"
)

func TestServiceStartAndShutdownMetrics(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(io.Discard)

	cfg := config.ObservabilityConfig{
		EnableMetrics: true,
		MetricsNS:     "test",
		EnableTracing: false,
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	cfg.MetricsAddr = ln.Addr().String()
	require.NoError(t, ln.Close())

	svc := New(cfg, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, svc.Start(ctx))

	client := &http.Client{Timeout: time.Second}
	require.Eventually(t, func() bool {
		resp, err := client.Get(fmt.Sprintf("http://%s/metrics", cfg.MetricsAddr))
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, time.Second, 50*time.Millisecond)

	require.NotNil(t, svc.Registry())

	svc.Shutdown(context.Background())
}

func TestServiceStartWithTracing(t *testing.T) {
	t.Parallel()

	logger := zerolog.New(io.Discard)
	cfg := config.ObservabilityConfig{
		EnableMetrics: false,
		EnableTracing: true,
		OTLPEndpoint:  "127.0.0.1:4317",
		ServiceName:   "test",
		MetricsNS:     "test",
		MetricsAddr:   "127.0.0.1:0",
	}

	svc := New(cfg, logger)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	require.NoError(t, svc.Start(ctx))
	svc.Shutdown(context.Background())
}
