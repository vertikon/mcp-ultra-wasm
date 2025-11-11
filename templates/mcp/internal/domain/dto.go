// Package domain contém DTOs mínimos exigidos por handlers e testes.
package domain

// CreateTaskRequest é o DTO para criação de tasks
type CreateTaskRequest struct {
	Title       string
	Description string
}

// UpdateTaskRequest é o DTO para atualização de tasks
type UpdateTaskRequest struct {
	Title       *string
	Description *string
}

// TaskFilters representa filtros para listagem de tasks
type TaskFilters struct {
	TenantKey string
	Limit     int
	Offset    int
}

// TaskList representa uma lista paginada de tasks
// Usa o tipo Task já definido em models.go
type TaskList struct {
	Items []*Task
	Total int
}
