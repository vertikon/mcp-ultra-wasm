package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sony/gobreaker"
	"gopkg.in/yaml.v3"

	"github.com/vertikon/mcp-ultra-templates/internal/models"
)

const metadataFile = "template.yaml"

// Repository provê acesso aos templates armazenados no filesystem.
type Repository struct {
	root    string
	breaker *gobreaker.CircuitBreaker
}

// New cria uma nova instância de Repository com circuit breaker padrão.
func New(root string) *Repository {
	settings := gobreaker.Settings{
		Name:        "fs-template-repo",
		MaxRequests: 5,
		Interval:    2 * time.Minute,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}

	return &Repository{
		root:    root,
		breaker: gobreaker.NewCircuitBreaker(settings),
	}
}

// ListTemplates lista os templates disponíveis a partir dos metadados.
func (r *Repository) ListTemplates(ctx context.Context) ([]models.TemplateMetadata, error) {
	result, err := r.breaker.Execute(func() (interface{}, error) {
		entries, err := os.ReadDir(r.root)
		if err != nil {
			return nil, fmt.Errorf("read templates dir: %w", err)
		}

		templates := make([]models.TemplateMetadata, 0, len(entries))
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			path := filepath.Join(r.root, entry.Name(), metadataFile)
			meta, err := readMetadata(path)
			if err != nil {
				return nil, err
			}
			if meta.Name == "" {
				meta.Name = entry.Name()
			}
			if meta.DisplayName == "" {
				meta.DisplayName = entry.Name()
			}
			templates = append(templates, *meta)
		}

		return templates, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]models.TemplateMetadata), nil
}

// LoadTemplate carrega os metadados e caminho do template solicitado.
func (r *Repository) LoadTemplate(ctx context.Context, name string) (*models.TemplateMetadata, string, error) {
	res, err := r.breaker.Execute(func() (interface{}, error) {
		path := filepath.Join(r.root, name)
		info, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("template not found: %s", name)
			}
			return nil, fmt.Errorf("stat template dir: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("template path is not a directory: %s", path)
		}

		metaPath := filepath.Join(path, metadataFile)
		meta, err := readMetadata(metaPath)
		if err != nil {
			return nil, err
		}
		if meta.Name == "" {
			meta.Name = name
		}
		if meta.DisplayName == "" {
			meta.DisplayName = name
		}

		return struct {
			meta *models.TemplateMetadata
			path string
		}{meta, path}, nil
	})
	if err != nil {
		return nil, "", err
	}

	result := res.(struct {
		meta *models.TemplateMetadata
		path string
	})

	return result.meta, result.path, nil
}

func readMetadata(path string) (*models.TemplateMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &models.TemplateMetadata{}, nil
		}
		return nil, fmt.Errorf("read metadata %s: %w", path, err)
	}

	var meta models.TemplateMetadata
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("unmarshal metadata %s: %w", path, err)
	}

	return &meta, nil
}

