package http

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/features"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/services"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/telemetry"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/httpx"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/types"
)

// TaskService interface defines methods for task operations
type TaskService interface {
	CreateTask(ctx context.Context, req services.CreateTaskRequest) (*domain.Task, error)
	GetTask(ctx context.Context, taskID types.UUID) (*domain.Task, error)
	UpdateTask(ctx context.Context, taskID types.UUID, req services.UpdateTaskRequest) (*domain.Task, error)
	DeleteTask(ctx context.Context, taskID types.UUID) error
	ListTasks(ctx context.Context, filters domain.TaskFilter) (*domain.TaskList, error)
	CompleteTask(ctx context.Context, taskID types.UUID) (*domain.Task, error)
	GetTasksByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error)
	GetTasksByAssignee(ctx context.Context, assigneeID types.UUID) ([]*domain.Task, error)
}

// HealthServiceInterface defines the interface for health service
type HealthServiceInterface interface {
	RegisterRoutes(r httpx.Router)
}

// Router creates and configures the HTTP router
func NewRouter(
	taskService TaskService,
	flagManager *features.FlagManager,
	healthService HealthServiceInterface,
	logger *zap.Logger,
) httpx.Router {
	r := httpx.NewRouter()

	// Middleware stack
	r.Use(httpx.RequestID)
	r.Use(httpx.RealIP)
	r.Use(httpx.Logger)
	r.Use(httpx.Recoverer)
	r.Use(httpx.Timeout(60))
	r.Use(telemetry.HTTPMetrics)

	// CORS configuration
	r.Use(httpx.DefaultCORS())

	// Health endpoints using comprehensive health service
	if healthService != nil {
		healthService.RegisterRoutes(r)
	} else {
		// Fallback to basic health checks
		r.Get("/healthz", healthCheck)
		r.Get("/readyz", readinessCheck)
	}

	// API routes
	r.Route("/api/v1", func(r httpx.Router) {
		// Task routes
		r.Mount("/tasks", TaskRoutes(taskService, logger))

		// Feature flag routes
		r.Mount("/flags", FeatureFlagRoutes(flagManager, logger))
	})

	return r
}

// TaskRoutes creates task-related routes
func TaskRoutes(taskService TaskService, logger *zap.Logger) httpx.Router {
	r := httpx.NewRouter()
	handlers := NewTaskHandlers(taskService, logger)

	r.Post("/", handlers.CreateTask)
	r.Get("/", handlers.ListTasks)
	r.Get("/{id}", handlers.GetTask)
	r.Put("/{id}", handlers.UpdateTask)
	r.Delete("/{id}", handlers.DeleteTask)
	r.Post("/{id}/complete", handlers.CompleteTask)
	r.Get("/status/{status}", handlers.GetTasksByStatus)
	r.Get("/assignee/{assigneeId}", handlers.GetTasksByAssignee)

	return r
}

// FeatureFlagRoutes creates feature flag routes
func FeatureFlagRoutes(flagManager *features.FlagManager, logger *zap.Logger) httpx.Router {
	r := httpx.NewRouter()
	handlers := NewFeatureFlagHandlers(flagManager, logger)

	r.Get("/", handlers.ListFlags)
	r.Get("/{key}", handlers.GetFlag)
	r.Post("/", handlers.CreateFlag)
	r.Put("/{key}", handlers.UpdateFlag)
	r.Delete("/{key}", handlers.DeleteFlag)
	r.Post("/{key}/evaluate", handlers.EvaluateFlag)

	return r
}

// Health check endpoint
func healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "healthy", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}

// Readiness check endpoint
func readinessCheck(w http.ResponseWriter, _ *http.Request) {
	// Add checks for dependencies (database, cache, etc.)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "ready", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
}
