package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildValues(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	file := filepath.Join(tmp, "values.yaml")
	require.NoError(t, os.WriteFile(file, []byte("name: demo\n"), 0o644))

	values, err := buildValues(file, []string{"env=prod"})
	require.NoError(t, err)
	require.Equal(t, "demo", values["name"])
	require.Equal(t, "prod", values["env"])
}

func TestBuildValuesInvalidSet(t *testing.T) {
	t.Parallel()

	_, err := buildValues("", []string{"invalid"})
	require.Error(t, err)
}

