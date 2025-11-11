package template

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

// ErrMissingVariable indica que uma variável obrigatória não foi fornecida.
type ErrMissingVariable struct {
	Key string
}

func (e ErrMissingVariable) Error() string {
	return fmt.Sprintf("variável obrigatória ausente: %s", e.Key)
}

// RenderOptions encapsula opções de geração.
type RenderOptions struct {
	IgnoredPaths map[string]struct{}
}

// RenderDirectory processa os arquivos em src e grava em dst aplicando as variáveis.
func RenderDirectory(ctx context.Context, src, dst string, values map[string]string, opts RenderOptions) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		if _, ignore := opts.IgnoredPaths[filepath.ToSlash(rel)]; ignore {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		targetPath := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		return renderFile(path, targetPath, values)
	})
}

func renderFile(src, dst string, values map[string]string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read source file: %w", err)
	}

	isTemplate := strings.HasSuffix(dst, ".tmpl")
	if isTemplate {
		dst = strings.TrimSuffix(dst, ".tmpl")
	}

	if !isTemplate && looksBinary(data) {
		return copyBinary(dst, data, info.Mode())
	}

	tmpl, err := template.New(filepath.Base(src)).
		Funcs(funcMap()).
		Option("missingkey=error").
		Parse(string(data))
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, values); err != nil {
		return fmt.Errorf("execute template: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("ensure target dir: %w", err)
	}

	if err := os.WriteFile(dst, buf.Bytes(), info.Mode()); err != nil {
		return fmt.Errorf("write rendered file: %w", err)
	}

	return nil
}

func copyBinary(dst string, data []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("ensure target dir: %w", err)
	}

	if err := os.WriteFile(dst, data, mode); err != nil {
		return fmt.Errorf("write binary file: %w", err)
	}
	return nil
}

func looksBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	sample := data
	if len(sample) > 8000 {
		sample = sample[:8000]
	}

	for _, b := range sample {
		if b == 0 {
			return true
		}
	}
	return false
}

func funcMap() template.FuncMap {
	return template.FuncMap{
		"toUpper": strings.ToUpper,
		"toLower": strings.ToLower,
		"camel":   toCamel,
		"kebab":   toKebab,
		"snake":   toSnake,
		"title":   strings.Title, //nolint:staticcheck
	}
}

func toCamel(input string) string {
	var buf strings.Builder
	nextUpper := true
	for _, r := range input {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			nextUpper = true
			continue
		}
		if nextUpper {
			buf.WriteRune(unicode.ToUpper(r))
			nextUpper = false
		} else {
			buf.WriteRune(unicode.ToLower(r))
		}
	}
	return buf.String()
}

func toKebab(input string) string {
	parts := splitWords(input)
	return strings.ToLower(strings.Join(parts, "-"))
}

func toSnake(input string) string {
	parts := splitWords(input)
	return strings.ToLower(strings.Join(parts, "_"))
}

func splitWords(input string) []string {
	var parts []string
	var current strings.Builder

	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
			continue
		}
		if current.Len() > 0 {
			parts = append(parts, current.String())
			current.Reset()
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

// CopyFile copia arquivos sem processar template (útil para cenários específicos).
func CopyFile(dst string, r io.Reader, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("ensure target dir: %w", err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read data: %w", err)
	}

	if err := os.WriteFile(dst, data, mode); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

