package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/vertikon/mcp-ultra-templates/internal/models"
)

func TestExecuteListCommand(t *testing.T) {
	temp := setupTemplateDir(t)

	args := []string{"list", "--config", temp.configPath}
	ctx := context.Background()

	out, restore := captureStdout(t)
	require.NoError(t, ExecuteWithArgs(ctx, args))
	restore()

	data, err := io.ReadAll(out)
	require.NoError(t, err)
	require.Contains(t, string(data), "NAME")
	_ = out.Close()
}

func TestExecuteListCommandJSON(t *testing.T) {
	temp := setupTemplateDir(t)

	args := []string{"list", "--config", temp.configPath, "--json"}
	ctx := context.Background()

	out, restore := captureStdout(t)
	require.NoError(t, ExecuteWithArgs(ctx, args))
	restore()

	data, err := io.ReadAll(out)
	require.NoError(t, err)

	var templates []models.TemplateMetadata
	require.NoError(t, json.Unmarshal(data, &templates))
	require.Len(t, templates, 1)
	require.Equal(t, "demo", templates[0].Name)

	_ = out.Close()
}

func TestExecuteRenderCommand(t *testing.T) {
	temp := setupTemplateDir(t)

	outputDir := filepath.Join(temp.root, "out")
	args := []string{
		"render",
		"--config", temp.configPath,
		"--template", "demo",
		"--output", outputDir,
		"--set", "project=demo",
	}

	ctx := context.Background()
	require.NoError(t, ExecuteWithArgs(ctx, args))

	data, err := os.ReadFile(filepath.Join(outputDir, "README.md"))
	require.NoError(t, err)
	require.Contains(t, string(data), "demo")
}

func TestExecuteRenderCommandInteractive(t *testing.T) {
	temp := setupTemplateDirNoDefaults(t)

	outputDir := filepath.Join(temp.root, "out-interactive")
	args := []string{
		"render",
		"--config", temp.configPath,
		"--template", "demo",
		"--output", outputDir,
		"--interactive",
	}

	out, restore := captureStdout(t)

	r, w, err := os.Pipe()
	require.NoError(t, err)

	origStdin := os.Stdin
	os.Stdin = r
	defer func() {
		os.Stdin = origStdin
		_ = r.Close()
	}()

	done := make(chan struct{})
	go func() {
		_, _ = fmt.Fprintln(w, "interactive-value")
		_ = w.Close()
		close(done)
	}()

	require.NoError(t, ExecuteWithArgs(context.Background(), args))

	restore()
	_ = out.Close()

	<-done

	data, err := os.ReadFile(filepath.Join(outputDir, "README.md"))
	require.NoError(t, err)
	require.Contains(t, string(data), "interactive-value")
}

func TestExecuteUsesOSArgs(t *testing.T) {
	temp := setupTemplateDir(t)

	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	os.Args = []string{"mcp-templates", "list", "--config", temp.configPath}
	require.NoError(t, Execute(context.Background()))
}

func TestMustAppPanicsWithoutContextValue(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}
	require.Panics(t, func() {
		MustApp(cmd)
	})
}

type testEnv struct {
	root       string
	configPath string
}

func setupTemplateDir(t *testing.T) testEnv {
	t.Helper()

	root := t.TempDir()
	templateDir := filepath.Join(root, "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templateDir, "demo"), 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "demo", "template.yaml"), []byte(`
name: demo
variables:
  - key: project
    required: true
defaults:
  project: sample
`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "demo", "README.md.tmpl"), []byte("{{ .project }}"), 0o644))

	cfgPath := filepath.Join(root, "config.yaml")
	configContent := fmt.Sprintf(`
templates_path: %q
observability:
  enable_metrics: false
  enable_tracing: false
`, templateDir)
	require.NoError(t, os.WriteFile(cfgPath, []byte(configContent), 0o644))

	return testEnv{
		root:       root,
		configPath: cfgPath,
	}
}

func captureStdout(t *testing.T) (*os.File, func()) {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := os.Stdout
	os.Stdout = w

	restore := func() {
		_ = w.Close()
		os.Stdout = orig
	}

	return r, restore
}

func setupTemplateDirNoDefaults(t *testing.T) testEnv {
	t.Helper()

	root := t.TempDir()
	templateDir := filepath.Join(root, "templates")
	require.NoError(t, os.MkdirAll(filepath.Join(templateDir, "demo"), 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "demo", "template.yaml"), []byte(`
name: demo
variables:
  - key: project
    required: true
`), 0o644))

	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "demo", "README.md.tmpl"), []byte("{{ .project }}"), 0o644))

	cfgPath := filepath.Join(root, "config.yaml")
	configContent := fmt.Sprintf(`
templates_path: %q
observability:
  enable_metrics: false
  enable_tracing: false
`, templateDir)
	require.NoError(t, os.WriteFile(cfgPath, []byte(configContent), 0o644))

	return testEnv{
		root:       root,
		configPath: cfgPath,
	}
}
