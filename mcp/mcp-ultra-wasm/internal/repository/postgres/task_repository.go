package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/types"
)

// TaskRepository implements domain.TaskRepository using PostgreSQL
type TaskRepository struct {
	db *sql.DB
}

// NewTaskRepository creates a new PostgreSQL task repository
func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create inserts a new task
func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, title, description, status, priority, assignee_id, created_by, created_at, updated_at, due_date, tags, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	tagsJSON, _ := json.Marshal(task.Tags)
	metadataJSON, _ := json.Marshal(task.Metadata)

	_, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.Priority,
		task.AssigneeID, task.CreatedBy, task.CreatedAt, task.UpdatedAt,
		task.DueDate, tagsJSON, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("creating task: %w", err)
	}

	return nil
}

// GetByID retrieves a task by ID
func (r *TaskRepository) GetByID(ctx context.Context, id types.UUID) (*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by,
		       created_at, updated_at, completed_at, due_date, tags, metadata
		FROM tasks WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	return r.scanTask(row)
}

// Update updates an existing task
func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks SET
			title = $2, description = $3, status = $4, priority = $5,
			assignee_id = $6, updated_at = $7, completed_at = $8, due_date = $9,
			tags = $10, metadata = $11
		WHERE id = $1
	`

	tagsJSON, _ := json.Marshal(task.Tags)
	metadataJSON, _ := json.Marshal(task.Metadata)

	result, err := r.db.ExecContext(ctx, query,
		task.ID, task.Title, task.Description, task.Status, task.Priority,
		task.AssigneeID, task.UpdatedAt, task.CompletedAt, task.DueDate,
		tagsJSON, metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("updating task: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// Delete removes a task
func (r *TaskRepository) Delete(ctx context.Context, id types.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

// List retrieves tasks with filtering and pagination
func (r *TaskRepository) List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error) {
	// Build WHERE clause
	conditions := make([]string, 0)
	args := make([]interface{}, 0)
	argIndex := 1

	if len(filter.Status) > 0 {
		placeholders := make([]string, len(filter.Status))
		for i, status := range filter.Status {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, status)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("status IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.Priority) > 0 {
		placeholders := make([]string, len(filter.Priority))
		for i, priority := range filter.Priority {
			placeholders[i] = fmt.Sprintf("$%d", argIndex)
			args = append(args, priority)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("priority IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.AssigneeID != nil {
		conditions = append(conditions, fmt.Sprintf("assignee_id = $%d", argIndex))
		args = append(args, *filter.AssigneeID)
		argIndex++
	}

	if filter.CreatedBy != nil {
		conditions = append(conditions, fmt.Sprintf("created_by = $%d", argIndex))
		args = append(args, *filter.CreatedBy)
		argIndex++
	}

	if filter.FromDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, *filter.FromDate)
		argIndex++
	}

	if filter.ToDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, *filter.ToDate)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(*) FROM tasks " + whereClause
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting tasks: %w", err)
	}

	// Data query
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by,
		       created_at, updated_at, completed_at, due_date, tags, metadata
		FROM tasks ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(argIndex) + ` OFFSET $` + strconv.Itoa(argIndex+1)

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying tasks: %w", err)
	}
	defer func() {
		_ = rows.Close() // Explicitly ignore error in defer
	}()

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

// GetByStatus retrieves tasks by status
func (r *TaskRepository) GetByStatus(ctx context.Context, status domain.TaskStatus) ([]*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by,
		       created_at, updated_at, completed_at, due_date, tags, metadata
		FROM tasks WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, fmt.Errorf("querying tasks by status: %w", err)
	}
	defer func() {
		_ = rows.Close() // Explicitly ignore error in defer
	}()

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetByAssignee retrieves tasks assigned to a specific user
func (r *TaskRepository) GetByAssignee(ctx context.Context, assigneeID types.UUID) ([]*domain.Task, error) {
	query := `
		SELECT id, title, description, status, priority, assignee_id, created_by,
		       created_at, updated_at, completed_at, due_date, tags, metadata
		FROM tasks WHERE assignee_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, assigneeID)
	if err != nil {
		return nil, fmt.Errorf("querying tasks by assignee: %w", err)
	}
	defer func() {
		_ = rows.Close() // Explicitly ignore error in defer
	}()

	tasks := make([]*domain.Task, 0)
	for rows.Next() {
		task, err := r.scanTask(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// scanTask scans a database row into a Task struct
func (r *TaskRepository) scanTask(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.Task, error) {
	var task domain.Task
	var tagsJSON, metadataJSON []byte

	err := scanner.Scan(
		&task.ID, &task.Title, &task.Description, &task.Status, &task.Priority,
		&task.AssigneeID, &task.CreatedBy, &task.CreatedAt, &task.UpdatedAt,
		&task.CompletedAt, &task.DueDate, &tagsJSON, &metadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}

	// Unmarshal JSON fields
	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &task.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	} else {
		task.Tags = make([]string, 0)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &task.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	} else {
		task.Metadata = make(map[string]interface{})
	}

	return &task, nil
}
