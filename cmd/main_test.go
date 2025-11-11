package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunExecutesListCommand(t *testing.T) {
	t.Parallel()

	cfgPath := prepareTemplateConfig(t)
	args := []string{"list", "--config", cfgPath}
	require.NoError(t, run(context.Background(), args))
}

func TestRunReturnsErrorForUnknownCommand(t *testing.T) {
	t.Parallel()

	err := run(context.Background(), []string{"unknown"})
	require.Error(t, err)
}

func TestMainSuccessDoesNotExit(t *testing.T) {
	cfgPath := prepareTemplateConfig(t)

	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"mcp-templates", "list", "--config", cfgPath}

	origExit := exitFunc
	defer func() { exitFunc = origExit }()

	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}

	main()
	require.False(t, exitCalled)
}

func TestMainErrorExits(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"mcp-templates", "unknown"}

	origExit := exitFunc
	defer func() { exitFunc = origExit }()

	var exitCode int
	exitFunc = func(code int) {
		exitCode = code
	}

	main()
	require.Equal(t, 1, exitCode)
}

func prepareTemplateConfig(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	templateDir := filepath.Join(root, "templates", "demo")
	require.NoError(t, os.MkdirAll(templateDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "template.yaml"), []byte("name: demo\n"), 0o644))

	cfgPath := filepath.Join(root, "config.yaml")
	configContent := fmt.Sprintf("templates_path: %q\nobservability:\n  enable_metrics: false\n  enable_tracing: false\n", filepath.Dir(templateDir))
	require.NoError(t, os.WriteFile(cfgPath, []byte(configContent), 0o644))

	return cfgPath
}
