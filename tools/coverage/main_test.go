package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculateCoverage(t *testing.T) {
	t.Parallel()

	path := writeCoverageFile(t, `mode: set
example.go:1.1,1.10 10 1
example.go:2.1,2.10 10 0
`)

	coverage, err := calculate(path)
	require.NoError(t, err)
	require.InDelta(t, 50.0, coverage, 0.1)
}

func TestCalculateCoverageZeroTotal(t *testing.T) {
	t.Parallel()

	path := writeCoverageFile(t, "mode: set\n")

	coverage, err := calculate(path)
	require.NoError(t, err)
	require.Equal(t, 0.0, coverage)
}

func TestCalculateCoverageMissingFile(t *testing.T) {
	t.Parallel()

	_, err := calculate("missing.out")
	require.Error(t, err)
}

func TestRunSuccessWritesOutput(t *testing.T) {
	path := writeCoverageFile(t, `mode: set
example.go:1.1,1.10 10 1
`)

	stdout, restore := captureStdout(t)

	require.NoError(t, run([]string{path}))

	restore()
	data, err := io.ReadAll(stdout)
	require.NoError(t, err)
	require.Contains(t, string(data), "coverage: 100.0%")
	require.NoError(t, stdout.Close())
}

func TestRunMissingArgs(t *testing.T) {
	err := run([]string{})
	require.Error(t, err)
}

func TestMainHandlesError(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"coverage"}

	origExit := exitFunc
	defer func() { exitFunc = origExit }()

	var exitCode int
	exitFunc = func(code int) {
		exitCode = code
	}

	stderr, restore := captureStderr(t)

	main()

	restore()
	data, err := io.ReadAll(stderr)
	require.NoError(t, err)
	require.Contains(t, string(data), "usage")
	require.Equal(t, 1, exitCode)
	require.NoError(t, stderr.Close())
}

func writeCoverageFile(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "coverage.out")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

func captureStdout(t *testing.T) (*os.File, func()) {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := os.Stdout
	os.Stdout = w

	return r, func() {
		_ = w.Close()
		os.Stdout = orig
	}
}

func captureStderr(t *testing.T) (*os.File, func()) {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := os.Stderr
	os.Stderr = w

	return r, func() {
		_ = w.Close()
		os.Stderr = orig
	}
}
