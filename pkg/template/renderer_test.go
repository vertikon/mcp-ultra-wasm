package template

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRenderDirectory(t *testing.T) {
	t.Parallel()

	tmpSrc := t.TempDir()
	tmpDst := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(tmpSrc, "nested"), 0o755))

	textTemplatePath := filepath.Join(tmpSrc, "nested", "config.yaml.tmpl")
	err := os.WriteFile(textTemplatePath, []byte("service: {{ toUpper .service_name }}"), 0o644)
	require.NoError(t, err)

	binaryPath := filepath.Join(tmpSrc, "module.wasm")
	require.NoError(t, os.WriteFile(binaryPath, []byte{0x00, 0x61, 0x62, 0x63}, 0o644))

	err = RenderDirectory(context.Background(), tmpSrc, tmpDst, map[string]string{
		"service_name": "demo",
	}, RenderOptions{
		IgnoredPaths: map[string]struct{}{},
	})
	require.NoError(t, err)

	renderedText := filepath.Join(tmpDst, "nested", "config.yaml")
	data, err := os.ReadFile(renderedText)
	require.NoError(t, err)
	require.Equal(t, "service: DEMO", string(data))

	renderedBinary := filepath.Join(tmpDst, "module.wasm")
	binData, err := os.ReadFile(renderedBinary)
	require.NoError(t, err)
	require.Equal(t, []byte{0x00, 0x61, 0x62, 0x63}, binData)
}

func TestRenderDirectoryMissingVariable(t *testing.T) {
	t.Parallel()

	tmpSrc := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpSrc, "config.yaml.tmpl"), []byte("{{ .missing }}"), 0o644))

	err := RenderDirectory(context.Background(), tmpSrc, t.TempDir(), map[string]string{}, RenderOptions{
		IgnoredPaths: map[string]struct{}{},
	})
	require.Error(t, err)
}

func TestCopyFile(t *testing.T) {
	t.Parallel()

	r := bytes.NewReader([]byte("sample"))
	target := filepath.Join(t.TempDir(), "file.txt")
	require.NoError(t, CopyFile(target, r, 0o600))

	data, err := os.ReadFile(target)
	require.NoError(t, err)
	require.Equal(t, "sample", string(data))
}

func TestFuncMapTransforms(t *testing.T) {
	t.Parallel()

	fns := funcMap()
	require.Equal(t, "HELLO", fns["toUpper"].(func(string) string)("hello"))
	require.Equal(t, "hello-world", fns["kebab"].(func(string) string)("Hello World"))
	require.Equal(t, "HelloWorld", fns["camel"].(func(string) string)("hello world"))
	require.Equal(t, "hello_world", fns["snake"].(func(string) string)("Hello World"))
}

func TestErrMissingVariable(t *testing.T) {
	t.Parallel()

	err := ErrMissingVariable{Key: "name"}
	require.EqualError(t, err, "variável obrigatória ausente: name")
}

func TestCopyFileError(t *testing.T) {
	t.Parallel()

	temp := t.TempDir()
	file := filepath.Join(temp, "existing.txt")
	require.NoError(t, os.WriteFile(file, []byte("data"), 0o644))

	dst := filepath.Join(file, "nested.txt")
	err := CopyFile(dst, bytes.NewReader([]byte("sample")), 0o644)
	require.Error(t, err)
}

func TestRenderDirectoryIgnorePaths(t *testing.T) {
	t.Parallel()

	tmpSrc := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(tmpSrc, "skip"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpSrc, "keep"), 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(tmpSrc, "skip", "ignored.txt"), []byte("ignore"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpSrc, "keep", "value.txt.tmpl"), []byte("value: {{ .val }}"), 0o644))

	tmpDst := t.TempDir()
	err := RenderDirectory(context.Background(), tmpSrc, tmpDst, map[string]string{"val": "ok"}, RenderOptions{
		IgnoredPaths: map[string]struct{}{
			"skip": {},
		},
	})
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(tmpDst, "skip"))
	require.True(t, os.IsNotExist(err))

	data, err := os.ReadFile(filepath.Join(tmpDst, "keep", "value.txt"))
	require.NoError(t, err)
	require.Equal(t, "value: ok", string(data))
}

func TestRenderDirectoryContextCanceled(t *testing.T) {
	t.Parallel()

	tmpSrc := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpSrc, "file.txt.tmpl"), []byte("content"), 0o644))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := RenderDirectory(ctx, tmpSrc, t.TempDir(), map[string]string{}, RenderOptions{
		IgnoredPaths: map[string]struct{}{},
	})
	require.ErrorIs(t, err, context.Canceled)
}
