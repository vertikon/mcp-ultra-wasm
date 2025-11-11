package metrics

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestRegistryServeAndShutdown(t *testing.T) {
	t.Parallel()

	reg := NewRegistry("test")
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_total",
		Help: "counter",
	})
	reg.MustRegister(counter)

	server := &Registry{}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	server.Serve(addr, reg)
	counter.Inc()

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/metrics", addr))
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(body), "test_total")

	require.NoError(t, server.Shutdown())
}

