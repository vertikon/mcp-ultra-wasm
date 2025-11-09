package domain

import (
	"time"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/types"
)

// Task represents a task in the system
type Task struct {
	ID          types.UUID             `json:"id" db:"id"`
	Title       string                 `json:"title" db:"title"`
	Description string                 `json:"description" db:"description"`
	Status      TaskStatus             `json:"status" db:"status"`
	Priority    Priority               `json:"priority" db:"priority"`
	AssigneeID  *types.UUID            `json:"assignee_id" db:"assignee_id"`
	CreatedBy   types.UUID             `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at" db:"completed_at"`
	DueDate     *time.Time             `json:"due_date" db:"due_date"`
	Tags        []string               `json:"tags" db:"tags"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Priority represents task priority
type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// User represents a user in the system
type User struct {
	ID        types.UUID `json:"id" db:"id"`
	Email     string     `json:"email" db:"email"`
	Name      string     `json:"name" db:"name"`
	Role      Role       `json:"role" db:"role"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	Active    bool       `json:"active" db:"active"`
}

// Role represents user role
type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

// Event represents a domain event
type Event struct {
	ID          types.UUID             `json:"id"`
	Type        string                 `json:"type"`
	AggregateID types.UUID             `json:"aggregate_id"`
	Data        map[string]interface{} `json:"data"`
	OccurredAt  time.Time              `json:"occurred_at"`
	Version     int                    `json:"version"`
}

// FeatureFlag represents a feature flag
type FeatureFlag struct {
	Key         string                 `json:"key" db:"key"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	Strategy    string                 `json:"strategy" db:"strategy"`
	Parameters  map[string]interface{} `json:"parameters" db:"parameters"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// TaskFilter represents filters for task queries
type TaskFilter struct {
	Status     []TaskStatus
	Priority   []Priority
	AssigneeID *types.UUID
	CreatedBy  *types.UUID
	Tags       []string
	FromDate   *time.Time
	ToDate     *time.Time
	Limit      int
	Offset     int
}

// NewTask creates a new task with default values
func NewTask(title, description string, createdBy types.UUID) *Task {
	return &Task{
		ID:          types.New(),
		Title:       title,
		Description: description,
		Status:      TaskStatusPending,
		Priority:    PriorityMedium,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        make([]string, 0),
		Metadata:    make(map[string]interface{}),
	}
}

// Complete marks a task as completed
func (t *Task) Complete() {
	now := time.Now()
	t.Status = TaskStatusCompleted
	t.UpdatedAt = now
	t.CompletedAt = &now
}

// Cancel marks a task as cancelled
func (t *Task) Cancel() {
	t.Status = TaskStatusCancelled
	t.UpdatedAt = time.Now()
}

// UpdateStatus updates task status
func (t *Task) UpdateStatus(status TaskStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
}

// IsValidStatus checks if status transition is valid
func (t *Task) IsValidStatus(newStatus TaskStatus) bool {
	switch t.Status {
	case TaskStatusPending:
		return newStatus == TaskStatusInProgress || newStatus == TaskStatusCancelled
	case TaskStatusInProgress:
		return newStatus == TaskStatusCompleted || newStatus == TaskStatusCancelled
	case TaskStatusCompleted, TaskStatusCancelled:
		return false
	default:
		return false
	}
}
