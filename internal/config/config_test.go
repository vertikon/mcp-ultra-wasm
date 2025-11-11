package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("")
	require.NoError(t, err)
	require.Equal(t, defaultTemplatesPath, cfg.TemplatesPath)
	require.Equal(t, defaultMetricsAddr, cfg.Observability.MetricsAddr)
	require.Equal(t, defaultOperationTimeout, cfg.Rendering.OperationTimeout)
}

func TestLoadFromFileAndEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`
templates_path: custom
rendering:
  max_retry_attempts: 5
`), 0o644))

	t.Setenv("LOG_LEVEL", "debug")

	cfg, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "custom", cfg.TemplatesPath)
	require.Equal(t, "debug", cfg.Logging.Level)
	require.Equal(t, 5, cfg.Rendering.MaxRetryAttempts)
	require.Equal(t, defaultOperationTimeout, cfg.Rendering.OperationTimeout)
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("non-existent.yaml")
	require.Error(t, err)
}

func TestLoadInvalidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`
templates_path: ""
observability:
  metrics_address: ""
rendering:
  operation_timeout: 0s
  max_retry_attempts: 0
`), 0o644))

	_, err := Load(path)
	require.Error(t, err)
}
