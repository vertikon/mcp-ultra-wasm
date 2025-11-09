package domain

import (
	"context"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/types"
)

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id types.UUID) (*Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id types.UUID) error
	List(ctx context.Context, filter TaskFilter) ([]*Task, int, error)
	GetByStatus(ctx context.Context, status TaskStatus) ([]*Task, error)
	GetByAssignee(ctx context.Context, assigneeID types.UUID) ([]*Task, error)
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id types.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id types.UUID) error
	List(ctx context.Context, limit, offset int) ([]*User, int, error)
}

// EventRepository defines the interface for event data access
type EventRepository interface {
	Store(ctx context.Context, event *Event) error
	GetByAggregateID(ctx context.Context, aggregateID types.UUID) ([]*Event, error)
	GetByType(ctx context.Context, eventType string, limit, offset int) ([]*Event, error)
}

// FeatureFlagRepository defines the interface for feature flag data access
type FeatureFlagRepository interface {
	GetByKey(ctx context.Context, key string) (*FeatureFlag, error)
	List(ctx context.Context) ([]*FeatureFlag, error)
	Create(ctx context.Context, flag *FeatureFlag) error
	Update(ctx context.Context, flag *FeatureFlag) error
	Delete(ctx context.Context, key string) error
}

// CacheRepository defines the interface for cache operations
type CacheRepository interface {
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increment(ctx context.Context, key string) (int64, error)
	SetNX(ctx context.Context, key string, value interface{}, ttl int) (bool, error)
}
