package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepositoryListAndLoad(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	templateDir := filepath.Join(root, "demo")
	require.NoError(t, os.MkdirAll(templateDir, 0o755))

	err := os.WriteFile(filepath.Join(templateDir, "template.yaml"), []byte(`
name: demo
description: demo template
`), 0o644)
	require.NoError(t, err)

	repo := New(root)
	results, err := repo.ListTemplates(context.Background())
	require.NoError(t, err)
	require.Len(t, results, 1)
	require.Equal(t, "demo", results[0].Name)

	meta, path, err := repo.LoadTemplate(context.Background(), "demo")
	require.NoError(t, err)
	require.Equal(t, "demo", meta.Name)
	require.Equal(t, templateDir, path)
}

func TestRepositoryLoadMissing(t *testing.T) {
	t.Parallel()

	repo := New(t.TempDir())
	_, _, err := repo.LoadTemplate(context.Background(), "missing")
	require.Error(t, err)
}

func TestRepositoryListWithoutMetadata(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "empty"), 0o755))

	repo := New(root)
	templates, err := repo.ListTemplates(context.Background())
	require.NoError(t, err)
	require.Len(t, templates, 1)
	require.Equal(t, "empty", templates[0].Name)
	require.Empty(t, templates[0].Description)
}
