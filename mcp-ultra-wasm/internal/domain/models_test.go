package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/types"
)

func TestNewTask(t *testing.T) {
	title := "Test Task"
	description := "Test Description"
	createdBy := types.New()

	task := NewTask(title, description, createdBy)

	assert.Equal(t, title, task.Title)
	assert.Equal(t, description, task.Description)
	assert.Equal(t, createdBy, task.CreatedBy)
	assert.Equal(t, TaskStatusPending, task.Status)
	assert.Equal(t, PriorityMedium, task.Priority)
	assert.NotEqual(t, types.Nil, task.ID)
	assert.False(t, task.CreatedAt.IsZero())
	assert.False(t, task.UpdatedAt.IsZero())
	assert.Empty(t, task.Tags)
	assert.NotNil(t, task.Metadata)
}

func TestTaskComplete(t *testing.T) {
	task := NewTask("Test", "Description", types.New())
	task.Status = TaskStatusInProgress

	beforeComplete := time.Now()
	task.Complete()
	afterComplete := time.Now()

	assert.Equal(t, TaskStatusCompleted, task.Status)
	assert.NotNil(t, task.CompletedAt)
	assert.True(t, !task.CompletedAt.Before(beforeComplete), "CompletedAt should be after or equal to beforeComplete")
	assert.True(t, !task.CompletedAt.After(afterComplete), "CompletedAt should be before or equal to afterComplete")
	assert.True(t, !task.UpdatedAt.Before(beforeComplete), "UpdatedAt should be after or equal to beforeComplete")
}

func TestTaskCancel(t *testing.T) {
	task := NewTask("Test", "Description", types.New())

	beforeCancel := time.Now()
	task.Cancel()
	afterCancel := time.Now()

	assert.Equal(t, TaskStatusCancelled, task.Status)
	assert.True(t, !task.UpdatedAt.Before(beforeCancel), "UpdatedAt should be after or equal to beforeCancel")
	assert.True(t, !task.UpdatedAt.After(afterCancel), "UpdatedAt should be before or equal to afterCancel")
}

func TestTaskUpdateStatus(t *testing.T) {
	task := NewTask("Test", "Description", types.New())

	beforeUpdate := time.Now()
	task.UpdateStatus(TaskStatusInProgress)
	afterUpdate := time.Now()

	assert.Equal(t, TaskStatusInProgress, task.Status)
	assert.True(t, !task.UpdatedAt.Before(beforeUpdate), "UpdatedAt should be after or equal to beforeUpdate")
	assert.True(t, !task.UpdatedAt.After(afterUpdate), "UpdatedAt should be before or equal to afterUpdate")
}

func TestTaskIsValidStatus(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus TaskStatus
		newStatus     TaskStatus
		expected      bool
	}{
		{
			name:          "Pending to InProgress",
			currentStatus: TaskStatusPending,
			newStatus:     TaskStatusInProgress,
			expected:      true,
		},
		{
			name:          "Pending to Cancelled",
			currentStatus: TaskStatusPending,
			newStatus:     TaskStatusCancelled,
			expected:      true,
		},
		{
			name:          "Pending to Completed",
			currentStatus: TaskStatusPending,
			newStatus:     TaskStatusCompleted,
			expected:      false,
		},
		{
			name:          "InProgress to Completed",
			currentStatus: TaskStatusInProgress,
			newStatus:     TaskStatusCompleted,
			expected:      true,
		},
		{
			name:          "InProgress to Cancelled",
			currentStatus: TaskStatusInProgress,
			newStatus:     TaskStatusCancelled,
			expected:      true,
		},
		{
			name:          "Completed to InProgress",
			currentStatus: TaskStatusCompleted,
			newStatus:     TaskStatusInProgress,
			expected:      false,
		},
		{
			name:          "Cancelled to InProgress",
			currentStatus: TaskStatusCancelled,
			newStatus:     TaskStatusInProgress,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("Test", "Description", types.New())
			task.Status = tt.currentStatus

			result := task.IsValidStatus(tt.newStatus)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTaskStatuses(t *testing.T) {
	assert.Equal(t, TaskStatus("pending"), TaskStatusPending)
	assert.Equal(t, TaskStatus("in_progress"), TaskStatusInProgress)
	assert.Equal(t, TaskStatus("completed"), TaskStatusCompleted)
	assert.Equal(t, TaskStatus("cancelled"), TaskStatusCancelled)
}

func TestPriorities(t *testing.T) {
	assert.Equal(t, Priority("low"), PriorityLow)
	assert.Equal(t, Priority("medium"), PriorityMedium)
	assert.Equal(t, Priority("high"), PriorityHigh)
	assert.Equal(t, Priority("urgent"), PriorityUrgent)
}

func TestUserRoles(t *testing.T) {
	assert.Equal(t, Role("admin"), RoleAdmin)
	assert.Equal(t, Role("user"), RoleUser)
}
