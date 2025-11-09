package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/types"
)

// TaskService handles task business logic
type TaskService struct {
	taskRepo  domain.TaskRepository
	userRepo  domain.UserRepository
	eventRepo domain.EventRepository
	cacheRepo domain.CacheRepository
	logger    *zap.Logger
	eventBus  EventBus
}

// EventBus defines interface for publishing events
type EventBus interface {
	Publish(ctx context.Context, event *domain.Event) error
}

// NewTaskService creates a new task service
func NewTaskService(
	taskRepo domain.TaskRepository,
	userRepo domain.UserRepository,
	eventRepo domain.EventRepository,
	cacheRepo domain.CacheRepository,
	logger *zap.Logger,
	eventBus EventBus,
) *TaskService {
	return &TaskService{
		taskRepo:  taskRepo,
		userRepo:  userRepo,
		eventRepo: eventRepo,
		cacheRepo: cacheRepo,
		logger:    logger,
		eventBus:  eventBus,
	}
}

// CreateTask creates a new task
func (s *TaskService) CreateTask(ctx context.Context, req CreateTaskRequest) (*domain.Task, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Verify creator exists
	creator, err := s.userRepo.GetByID(ctx, req.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("creator not found: %w", err)
	}

	// Verify assignee exists if provided
	if req.AssigneeID != nil {
		if _, err := s.userRepo.GetByID(ctx, *req.AssigneeID); err != nil {
			return nil, fmt.Errorf("assignee not found: %w", err)
		}
	}

	// Create task
	task := domain.NewTask(req.Title, req.Description, creator.ID)
	task.Priority = req.Priority
	task.AssigneeID = req.AssigneeID
	task.DueDate = req.DueDate
	task.Tags = req.Tags

	// Save to repository
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("creating task: %w", err)
	}

	// Publish event
	event := &domain.Event{
		ID:          types.New(),
		Type:        "task.created",
		AggregateID: task.ID,
		Data: map[string]interface{}{
			"task_id":     task.ID,
			"title":       task.Title,
			"created_by":  task.CreatedBy,
			"assignee_id": task.AssigneeID,
			"priority":    task.Priority,
		},
		OccurredAt: time.Now(),
		Version:    1,
	}

	if err := s.publishEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish task created event", zap.Error(err))
	}

	// Clear cache
	s.invalidateTaskCache(ctx)

	s.logger.Info("Task created",
		zap.String("task_id", task.ID.String()),
		zap.String("title", task.Title),
		zap.String("created_by", creator.Email))

	return task, nil
}

// UpdateTask updates an existing task
func (s *TaskService) UpdateTask(ctx context.Context, id types.UUID, req UpdateTaskRequest) (*domain.Task, error) {
	// Get existing task
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	// Update fields if provided
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.AssigneeID != nil {
		// Verify assignee exists
		if _, err := s.userRepo.GetByID(ctx, *req.AssigneeID); err != nil {
			return nil, fmt.Errorf("assignee not found: %w", err)
		}
		task.AssigneeID = req.AssigneeID
	}
	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}
	if req.Tags != nil {
		task.Tags = req.Tags
	}

	task.UpdatedAt = time.Now()

	// Save changes
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("updating task: %w", err)
	}

	// Publish event
	event := &domain.Event{
		ID:          types.New(),
		Type:        "task.updated",
		AggregateID: task.ID,
		Data: map[string]interface{}{
			"task_id": task.ID,
			"changes": req,
		},
		OccurredAt: time.Now(),
		Version:    1,
	}

	if err := s.publishEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish task updated event", zap.Error(err))
	}

	// Clear cache
	s.invalidateTaskCache(ctx)

	s.logger.Info("Task updated", zap.String("task_id", task.ID.String()))

	return task, nil
}

// CompleteTask marks a task as completed
func (s *TaskService) CompleteTask(ctx context.Context, id types.UUID) (*domain.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	if !task.IsValidStatus(domain.TaskStatusCompleted) {
		return nil, fmt.Errorf("cannot complete task in status: %s", task.Status)
	}

	task.Complete()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("completing task: %w", err)
	}

	// Publish event
	event := &domain.Event{
		ID:          types.New(),
		Type:        "task.completed",
		AggregateID: task.ID,
		Data: map[string]interface{}{
			"task_id":      task.ID,
			"completed_at": task.CompletedAt,
		},
		OccurredAt: time.Now(),
		Version:    1,
	}

	if err := s.publishEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish task completed event", zap.Error(err))
	}

	// Clear cache
	s.invalidateTaskCache(ctx)

	s.logger.Info("Task completed", zap.String("task_id", task.ID.String()))

	return task, nil
}

// GetTask retrieves a task by ID with caching
func (s *TaskService) GetTask(ctx context.Context, id types.UUID) (*domain.Task, error) {
	// Try cache first
	cacheKey := fmt.Sprintf("task:%s", id.String())
	cachedData, err := s.cacheRepo.Get(ctx, cacheKey)
	if err == nil {
		var cachedTask domain.Task
		if json.Unmarshal([]byte(cachedData), &cachedTask) == nil {
			return &cachedTask, nil
		}
	}

	// Get from repository
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	if err := s.cacheRepo.Set(ctx, cacheKey, task, 300); err != nil {
		s.logger.Error("Failed to cache task", zap.Error(err))
	}

	return task, nil
}

// ListTasks lists tasks with filtering
func (s *TaskService) ListTasks(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error) {
	return s.taskRepo.List(ctx, filter)
}

// DeleteTask deletes a task
func (s *TaskService) DeleteTask(ctx context.Context, id types.UUID) error {
	// Verify task exists
	if _, err := s.taskRepo.GetByID(ctx, id); err != nil {
		return fmt.Errorf("task not found: %w", err)
	}

	if err := s.taskRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}

	// Publish event
	event := &domain.Event{
		ID:          types.New(),
		Type:        "task.deleted",
		AggregateID: id,
		Data: map[string]interface{}{
			"task_id": id,
		},
		OccurredAt: time.Now(),
		Version:    1,
	}

	if err := s.publishEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish task deleted event", zap.Error(err))
	}

	// Clear cache
	s.invalidateTaskCache(ctx)

	s.logger.Info("Task deleted", zap.String("task_id", id.String()))

	return nil
}

// GetTasksByStatus retrieves tasks by status
func (s *TaskService) GetTasksByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	return s.taskRepo.GetByStatus(ctx, status)
}

// GetTasksByAssignee retrieves tasks assigned to a user
func (s *TaskService) GetTasksByAssignee(ctx context.Context, assigneeID types.UUID) ([]*domain.Task, error) {
	return s.taskRepo.GetByAssignee(ctx, assigneeID)
}

// publishEvent publishes an event to the event store and event bus
func (s *TaskService) publishEvent(ctx context.Context, event *domain.Event) error {
	// Store in event store
	if err := s.eventRepo.Store(ctx, event); err != nil {
		return fmt.Errorf("storing event: %w", err)
	}

	// Publish to event bus
	if err := s.eventBus.Publish(ctx, event); err != nil {
		return fmt.Errorf("publishing event: %w", err)
	}

	return nil
}

// invalidateTaskCache clears task-related cache entries
func (s *TaskService) invalidateTaskCache(_ context.Context) {
	// Implementation would depend on cache invalidation strategy
	// For now, we'll just log it
	s.logger.Debug("Task cache invalidated")
}

// Request and Response types
type CreateTaskRequest struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Priority    domain.Priority `json:"priority"`
	AssigneeID  *types.UUID     `json:"assignee_id"`
	CreatedBy   types.UUID      `json:"created_by"`
	DueDate     *time.Time      `json:"due_date"`
	Tags        []string        `json:"tags"`
}

func (r CreateTaskRequest) Validate() error {
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}
	if r.CreatedBy == types.Nil {
		return fmt.Errorf("created_by is required")
	}
	return nil
}

type UpdateTaskRequest struct {
	Title       *string          `json:"title"`
	Description *string          `json:"description"`
	Priority    *domain.Priority `json:"priority"`
	AssigneeID  *types.UUID      `json:"assignee_id"`
	DueDate     *time.Time       `json:"due_date"`
	Tags        []string         `json:"tags"`
}
