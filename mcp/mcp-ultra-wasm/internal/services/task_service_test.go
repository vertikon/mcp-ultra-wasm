package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/types"
)

// Mock repositories
type mockTaskRepository struct {
	mock.Mock
}

func (m *mockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockTaskRepository) GetByID(ctx context.Context, id types.UUID) (*domain.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *mockTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *mockTaskRepository) Delete(ctx context.Context, id types.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTaskRepository) List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.Task), args.Int(1), args.Error(2)
}

func (m *mockTaskRepository) GetByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *mockTaskRepository) GetByAssignee(ctx context.Context, assigneeID types.UUID) ([]*domain.Task, error) {
	args := m.Called(ctx, assigneeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id types.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) Delete(ctx context.Context, id types.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, int, error) {
	args := m.Called(ctx, offset, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.User), args.Int(1), args.Error(2)
}

type mockEventRepository struct {
	mock.Mock
}

func (m *mockEventRepository) Save(ctx context.Context, event *domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *mockEventRepository) GetByAggregateID(ctx context.Context, aggregateID types.UUID) ([]*domain.Event, error) {
	args := m.Called(ctx, aggregateID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Event), args.Error(1)
}

func (m *mockEventRepository) GetByType(ctx context.Context, eventType string, limit int, offset int) ([]*domain.Event, error) {
	args := m.Called(ctx, eventType, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Event), args.Error(1)
}

func (m *mockEventRepository) Store(ctx context.Context, event *domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type mockCacheRepository struct {
	mock.Mock
}

func (m *mockCacheRepository) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *mockCacheRepository) SetNX(ctx context.Context, key string, value interface{}, ttl int) (bool, error) {
	args := m.Called(ctx, key, value, ttl)
	return args.Bool(0), args.Error(1)
}

func (m *mockCacheRepository) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *mockCacheRepository) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *mockCacheRepository) Clear(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *mockCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	args := m.Called(ctx, key)
	return args.Bool(0), args.Error(1)
}

func (m *mockCacheRepository) Increment(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

type mockEventBus struct {
	mock.Mock
}

func (m *mockEventBus) Publish(ctx context.Context, event *domain.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// Test helper functions
func createTestTaskService() (*TaskService, *mockTaskRepository, *mockUserRepository, *mockEventRepository, *mockCacheRepository, *mockEventBus) {
	taskRepo := &mockTaskRepository{}
	userRepo := &mockUserRepository{}
	eventRepo := &mockEventRepository{}
	cacheRepo := &mockCacheRepository{}
	eventBus := &mockEventBus{}
	logger := zap.NewNop()

	service := NewTaskService(taskRepo, userRepo, eventRepo, cacheRepo, logger, eventBus)
	return service, taskRepo, userRepo, eventRepo, cacheRepo, eventBus
}

func createTestUser() *domain.User {
	return &domain.User{
		ID:    types.New(),
		Email: "test@example.com",
		Name:  "Test User",
	}
}

func createTestTask() *domain.Task {
	userID := types.New()
	return &domain.Task{
		ID:          types.New(),
		Title:       "Test Task",
		Description: "Test Description",
		Status:      domain.TaskStatusPending,
		Priority:    domain.PriorityMedium,
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Test cases
func TestTaskService_CreateTask_Success(t *testing.T) {
	service, taskRepo, userRepo, eventRepo, cacheRepo, eventBus := createTestTaskService()

	creator := createTestUser()
	assignee := createTestUser()

	req := CreateTaskRequest{
		Title:       "New Task",
		Description: "Task Description",
		Priority:    domain.PriorityHigh,
		CreatedBy:   creator.ID,
		AssigneeID:  &assignee.ID,
		Tags:        []string{"urgent", "project-a"},
	}

	ctx := context.Background()

	// Set up expectations
	userRepo.On("GetByID", ctx, creator.ID).Return(creator, nil)
	userRepo.On("GetByID", ctx, assignee.ID).Return(assignee, nil)
	taskRepo.On("Create", ctx, mock.AnythingOfType("*domain.Task")).Return(nil)
	eventRepo.On("Store", ctx, mock.AnythingOfType("*domain.Event")).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("*domain.Event")).Return(nil)

	// Execute
	result, err := service.CreateTask(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "New Task", result.Title)
	assert.Equal(t, "Task Description", result.Description)
	assert.Equal(t, domain.PriorityHigh, result.Priority)
	assert.Equal(t, creator.ID, result.CreatedBy)
	assert.Equal(t, &assignee.ID, result.AssigneeID)
	assert.Equal(t, []string{"urgent", "project-a"}, result.Tags)
	assert.Equal(t, domain.TaskStatusPending, result.Status)

	// Verify all expectations were met
	taskRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
	cacheRepo.AssertExpectations(t)
}

func TestTaskService_CreateTask_ValidationError(t *testing.T) {
	service, _, _, _, _, _ := createTestTaskService()

	req := CreateTaskRequest{
		Title:     "", // Empty title should cause validation error
		CreatedBy: types.New(),
	}

	ctx := context.Background()

	// Execute
	result, err := service.CreateTask(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "title is required")
}

func TestTaskService_CreateTask_CreatorNotFound(t *testing.T) {
	service, _, userRepo, _, _, _ := createTestTaskService()

	creatorID := types.New()
	req := CreateTaskRequest{
		Title:     "Test Task",
		CreatedBy: creatorID,
	}

	ctx := context.Background()

	// Set up expectations
	userRepo.On("GetByID", ctx, creatorID).Return(nil, errors.New("user not found"))

	// Execute
	result, err := service.CreateTask(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "creator not found")

	userRepo.AssertExpectations(t)
}

func TestTaskService_CreateTask_AssigneeNotFound(t *testing.T) {
	service, _, userRepo, _, _, _ := createTestTaskService()

	creator := createTestUser()
	assigneeID := types.New()

	req := CreateTaskRequest{
		Title:      "Test Task",
		CreatedBy:  creator.ID,
		AssigneeID: &assigneeID,
	}

	ctx := context.Background()

	// Set up expectations
	userRepo.On("GetByID", ctx, creator.ID).Return(creator, nil)
	userRepo.On("GetByID", ctx, assigneeID).Return(nil, errors.New("user not found"))

	// Execute
	result, err := service.CreateTask(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "assignee not found")

	userRepo.AssertExpectations(t)
}

func TestTaskService_UpdateTask_Success(t *testing.T) {
	service, taskRepo, userRepo, eventRepo, _, eventBus := createTestTaskService()

	existingTask := createTestTask()
	assignee := createTestUser()

	newTitle := "Updated Task"
	newPriority := domain.PriorityLow

	req := UpdateTaskRequest{
		Title:      &newTitle,
		Priority:   &newPriority,
		AssigneeID: &assignee.ID,
	}

	ctx := context.Background()

	// Set up expectations
	taskRepo.On("GetByID", ctx, existingTask.ID).Return(existingTask, nil)
	userRepo.On("GetByID", ctx, assignee.ID).Return(assignee, nil)
	taskRepo.On("Update", ctx, mock.AnythingOfType("*domain.Task")).Return(nil)
	eventRepo.On("Store", ctx, mock.AnythingOfType("*domain.Event")).Return(nil)
	eventBus.On("Publish", ctx, mock.AnythingOfType("*domain.Event")).Return(nil)

	// Execute
	result, err := service.UpdateTask(ctx, existingTask.ID, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Updated Task", result.Title)
	assert.Equal(t, domain.PriorityLow, result.Priority)
	assert.Equal(t, &assignee.ID, result.AssigneeID)

	// Verify all expectations were met
	taskRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	eventBus.AssertExpectations(t)
}

func TestTaskService_UpdateTask_TaskNotFound(t *testing.T) {
	service, taskRepo, _, _, _, _ := createTestTaskService()

	taskID := types.New()
	req := UpdateTaskRequest{}

	ctx := context.Background()

	// Set up expectations
	taskRepo.On("GetByID", ctx, taskID).Return(nil, errors.New("task not found"))

	// Execute
	result, err := service.UpdateTask(ctx, taskID, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "task not found")

	taskRepo.AssertExpectations(t)
}

func TestCreateTaskRequest_Validate_Success(t *testing.T) {
	req := CreateTaskRequest{
		Title:     "Valid Task",
		CreatedBy: types.New(),
	}

	err := req.Validate()
	assert.NoError(t, err)
}

func TestCreateTaskRequest_Validate_EmptyTitle(t *testing.T) {
	req := CreateTaskRequest{
		Title:     "",
		CreatedBy: types.New(),
	}

	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "title is required")
}

func TestCreateTaskRequest_Validate_EmptyCreatedBy(t *testing.T) {
	req := CreateTaskRequest{
		Title:     "Valid Task",
		CreatedBy: types.Nil,
	}

	err := req.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "created_by is required")
}
