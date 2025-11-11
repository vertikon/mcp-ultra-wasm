package template

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/vertikon/mcp-ultra-templates/internal/config"
	"github.com/vertikon/mcp-ultra-templates/internal/models"
	"github.com/vertikon/mcp-ultra-templates/internal/services/template/mocks"
)

func TestServiceRenderSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	meta := &models.TemplateMetadata{
		Name:      "demo",
		Variables: []models.TemplateVariable{{Key: "service_name", Required: true}},
		Defaults:  map[string]string{"service_name": "fallback"},
	}

	templateDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "main.go.tmpl"), []byte(`package main
const Service = "{{ toUpper .service_name }}"
`), 0o644))

	outputDir := t.TempDir()

	mockRepo.EXPECT().
		LoadTemplate(gomock.Any(), "demo").
		Return(meta, templateDir, nil).
		Times(1)

	cfg := config.RenderingConfig{
		OperationTimeout: 5 * time.Second,
		MaxRetryAttempts: 3,
	}

	logger := testLogger()
	reg := prometheus.NewRegistry()

	service := New(cfg, logger, reg, mockRepo)

	resp, err := service.Render(context.Background(), RenderRequest{
		TemplateName: "demo",
		OutputDir:    outputDir,
		Values: map[string]string{
			"service_name": "ultra",
		},
		Overwrite: true,
	})

	require.NoError(t, err)
	require.Equal(t, "demo", resp.Template.Name)

	renderedData, err := os.ReadFile(filepath.Join(outputDir, "main.go"))
	require.NoError(t, err)
	require.Contains(t, string(renderedData), `const Service = "ULTRA"`)
}

func TestServiceRenderMissingVariable(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	meta := &models.TemplateMetadata{
		Name:      "demo",
		Variables: []models.TemplateVariable{{Key: "env", Required: true}},
	}

	templateDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "config.yaml.tmpl"), []byte(`env: {{ .env }}`), 0o644))

	mockRepo.EXPECT().
		LoadTemplate(gomock.Any(), "demo").
		Return(meta, templateDir, nil).
		Times(1)

	cfg := config.RenderingConfig{
		OperationTimeout: 5 * time.Second,
		MaxRetryAttempts: 1,
	}

	logger := testLogger()
	reg := prometheus.NewRegistry()
	service := New(cfg, logger, reg, mockRepo)

	_, err := service.Render(context.Background(), RenderRequest{
		TemplateName: "demo",
		OutputDir:    t.TempDir(),
		Values:       map[string]string{},
		Overwrite:    true,
	})

	require.Error(t, err)
}

func TestServiceList(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	meta := models.TemplateMetadata{Name: "demo"}
	mockRepo.EXPECT().
		ListTemplates(gomock.Any()).
		Return([]models.TemplateMetadata{meta}, nil).
		Times(1)

	cfg := config.RenderingConfig{
		OperationTimeout: 5 * time.Second,
		MaxRetryAttempts: 1,
	}

	logger := testLogger()
	reg := prometheus.NewRegistry()
	service := New(cfg, logger, reg, mockRepo)

	templates, err := service.List(context.Background())
	require.NoError(t, err)
	require.Len(t, templates, 1)
	require.Equal(t, "demo", templates[0].Name)
}

func TestServiceRenderOutputNotEmpty(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	meta := &models.TemplateMetadata{Name: "demo"}
	templateDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "README.md"), []byte("plain"), 0o644))

	mockRepo.EXPECT().
		LoadTemplate(gomock.Any(), "demo").
		Return(meta, templateDir, nil).
		Times(1)

	outputDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "existing.txt"), []byte("content"), 0o644))

	cfg := config.RenderingConfig{
		OperationTimeout: 5 * time.Second,
		MaxRetryAttempts: 1,
	}

	logger := testLogger()
	reg := prometheus.NewRegistry()
	service := New(cfg, logger, reg, mockRepo)

	_, err := service.Render(context.Background(), RenderRequest{
		TemplateName: "demo",
		OutputDir:    outputDir,
		Values:       map[string]string{},
		Overwrite:    false,
	})

	require.Error(t, err)
}

func TestServiceRenderMissingTemplateName(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := config.RenderingConfig{
		OperationTimeout: 5 * time.Second,
		MaxRetryAttempts: 1,
	}

	logger := testLogger()
	reg := prometheus.NewRegistry()
	service := New(cfg, logger, reg, mocks.NewMockRepository(ctrl))

	_, err := service.Render(context.Background(), RenderRequest{
		TemplateName: "",
		OutputDir:    t.TempDir(),
		Overwrite:    true,
	})

	require.Error(t, err)
}

func TestServiceRenderRepositoryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockRepo.EXPECT().
		LoadTemplate(gomock.Any(), "demo").
		Return(nil, "", assert.AnError).
		Times(1)

	cfg := config.RenderingConfig{
		OperationTimeout: 5 * time.Second,
		MaxRetryAttempts: 1,
	}

	logger := testLogger()
	reg := prometheus.NewRegistry()
	service := New(cfg, logger, reg, mockRepo)

	_, err := service.Render(context.Background(), RenderRequest{
		TemplateName: "demo",
		OutputDir:    t.TempDir(),
		Overwrite:    true,
	})

	require.Error(t, err)
}

func TestNewDefaultRepository(t *testing.T) {
	t.Parallel()

	repo := NewDefaultRepository(t.TempDir())
	require.NotNil(t, repo)
}

func TestServiceRenderCreatesOutputDir(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	meta := &models.TemplateMetadata{Name: "demo"}
	templateDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "file.txt"), []byte("data"), 0o644))

	mockRepo.EXPECT().
		LoadTemplate(gomock.Any(), "demo").
		Return(meta, templateDir, nil).
		Times(1)

	cfg := config.RenderingConfig{
		OperationTimeout: 5 * time.Second,
		MaxRetryAttempts: 1,
	}

	logger := testLogger()
	reg := prometheus.NewRegistry()
	service := New(cfg, logger, reg, mockRepo)

	outputDir := filepath.Join(t.TempDir(), "new-project")

	resp, err := service.Render(context.Background(), RenderRequest{
		TemplateName: "demo",
		OutputDir:    outputDir,
		Overwrite:    true,
	})

	require.NoError(t, err)
	require.Equal(t, outputDir, resp.Output)

	info, err := os.Stat(outputDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestServiceRenderOutputIsFile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	meta := &models.TemplateMetadata{Name: "demo"}
	templateDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "file.txt"), []byte("data"), 0o644))

	mockRepo.EXPECT().
		LoadTemplate(gomock.Any(), "demo").
		Return(meta, templateDir, nil).
		Times(1)

	cfg := config.RenderingConfig{
		OperationTimeout: 5 * time.Second,
		MaxRetryAttempts: 1,
	}

	logger := testLogger()
	reg := prometheus.NewRegistry()
	service := New(cfg, logger, reg, mockRepo)

	outputFile := filepath.Join(t.TempDir(), "output.txt")
	require.NoError(t, os.WriteFile(outputFile, []byte("content"), 0o644))

	_, err := service.Render(context.Background(), RenderRequest{
		TemplateName: "demo",
		OutputDir:    outputFile,
		Overwrite:    true,
	})

	require.Error(t, err)
}

func testLogger() zerolog.Logger {
	var buf bytes.Buffer
	return zerolog.New(&buf).With().Timestamp().Logger()
}
