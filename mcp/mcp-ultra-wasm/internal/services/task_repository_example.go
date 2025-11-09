//go:build example || testpatch

// Package services contains service-layer contracts and examples.
// This file is excluded from production builds by the build tag above.
package services

import "context"

// TaskFilters represents filter criteria for listing tasks.
type TaskFilters struct {
	TenantKey string
	Limit     int
	Offset    int
}

// Task is a minimal example entity used by the TaskRepository interface.
type Task struct {
	ID    string
	Title string
}

// TaskRepository documents the final v11.1 signature used across tests/mocks.
// NOTE: This is an example contract; real implementations live elsewhere.
type TaskRepository interface {
	// List returns the tasks for a given filter, along with the total count.
	List(ctx context.Context, filter TaskFilters) ([]*Task, int, error)
	// Create persists a new task and returns its ID.
	Create(ctx context.Context, t *Task) (string, error)
	// Exists checks if a task ID exists; used by some tests/mocks.
	Exists(ctx context.Context, id string) (bool, error)
}
