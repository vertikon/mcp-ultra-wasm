package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/services"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/httpx"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/pkg/types"
)

// TaskHandlers handles HTTP requests for tasks
type TaskHandlers struct {
	taskService TaskService
	logger      *zap.Logger
}

// NewTaskHandlers creates new task handlers
func NewTaskHandlers(taskService TaskService, logger *zap.Logger) *TaskHandlers {
	return &TaskHandlers{
		taskService: taskService,
		logger:      logger,
	}
}

// CreateTask handles task creation
func (h *TaskHandlers) CreateTask(w http.ResponseWriter, r *http.Request) {
	var req services.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	task, err := h.taskService.CreateTask(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create task", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create task", err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, task)
}

// GetTask handles task retrieval
func (h *TaskHandlers) GetTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := httpx.URLParam(r, "id")
	taskID, err := types.Parse(taskIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	task, err := h.taskService.GetTask(r.Context(), taskID)
	if err != nil {
		h.logger.Error("Failed to get task", zap.Error(err))
		h.writeErrorResponse(w, http.StatusNotFound, "Task not found", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, task)
}

// UpdateTask handles task updates
func (h *TaskHandlers) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := httpx.URLParam(r, "id")
	taskID, err := types.Parse(taskIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	var req services.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	task, err := h.taskService.UpdateTask(r.Context(), taskID, req)
	if err != nil {
		h.logger.Error("Failed to update task", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to update task", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, task)
}

// CompleteTask handles task completion
func (h *TaskHandlers) CompleteTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := httpx.URLParam(r, "id")
	taskID, err := types.Parse(taskIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	task, err := h.taskService.CompleteTask(r.Context(), taskID)
	if err != nil {
		h.logger.Error("Failed to complete task", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to complete task", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, task)
}

// DeleteTask handles task deletion
func (h *TaskHandlers) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskIDStr := httpx.URLParam(r, "id")
	taskID, err := types.Parse(taskIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid task ID", err)
		return
	}

	if err := h.taskService.DeleteTask(r.Context(), taskID); err != nil {
		h.logger.Error("Failed to delete task", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete task", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListTasks handles task listing with filters
func (h *TaskHandlers) ListTasks(w http.ResponseWriter, r *http.Request) {
	filter := h.parseTaskFilter(r)

	taskList, err := h.taskService.ListTasks(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list tasks", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list tasks", err)
		return
	}

	response := TaskListResponse{
		Tasks: taskList.Items,
		Total: taskList.Total,
		Page:  filter.Offset/filter.Limit + 1,
		Limit: filter.Limit,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetTasksByStatus handles retrieving tasks by status
func (h *TaskHandlers) GetTasksByStatus(w http.ResponseWriter, r *http.Request) {
	statusStr := httpx.URLParam(r, "status")
	status := domain.TaskStatus(statusStr)

	tasks, err := h.taskService.GetTasksByStatus(r.Context(), status)
	if err != nil {
		h.logger.Error("Failed to get tasks by status", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get tasks", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, tasks)
}

// GetTasksByAssignee handles retrieving tasks by assignee
func (h *TaskHandlers) GetTasksByAssignee(w http.ResponseWriter, r *http.Request) {
	assigneeIDStr := httpx.URLParam(r, "assigneeId")
	assigneeID, err := types.Parse(assigneeIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid assignee ID", err)
		return
	}

	tasks, err := h.taskService.GetTasksByAssignee(r.Context(), assigneeID)
	if err != nil {
		h.logger.Error("Failed to get tasks by assignee", zap.Error(err))
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get tasks", err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, tasks)
}

// parseTaskFilter parses query parameters into TaskFilter
func (h *TaskHandlers) parseTaskFilter(r *http.Request) domain.TaskFilter {
	filter := domain.TaskFilter{}

	// Status filter
	if statusParams := r.URL.Query()["status"]; len(statusParams) > 0 {
		for _, status := range statusParams {
			filter.Status = append(filter.Status, domain.TaskStatus(status))
		}
	}

	// Priority filter
	if priorityParams := r.URL.Query()["priority"]; len(priorityParams) > 0 {
		for _, priority := range priorityParams {
			filter.Priority = append(filter.Priority, domain.Priority(priority))
		}
	}

	// Assignee filter
	if assigneeID := r.URL.Query().Get("assignee_id"); assigneeID != "" {
		if id, err := types.Parse(assigneeID); err == nil {
			filter.AssigneeID = &id
		}
	}

	// Creator filter
	if createdBy := r.URL.Query().Get("created_by"); createdBy != "" {
		if id, err := types.Parse(createdBy); err == nil {
			filter.CreatedBy = &id
		}
	}

	// Date filters
	if fromDate := r.URL.Query().Get("from_date"); fromDate != "" {
		if t, err := time.Parse(time.RFC3339, fromDate); err == nil {
			filter.FromDate = &t
		}
	}

	if toDate := r.URL.Query().Get("to_date"); toDate != "" {
		if t, err := time.Parse(time.RFC3339, toDate); err == nil {
			filter.ToDate = &t
		}
	}

	// Pagination
	filter.Limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}

	filter.Offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Tags filter
	filter.Tags = r.URL.Query()["tags"]

	return filter
}

// writeJSONResponse writes a JSON response
func (h *TaskHandlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

// writeErrorResponse writes an error response
func (h *TaskHandlers) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := ErrorResponse{
		Error:   message,
		Details: err.Error(),
		Code:    statusCode,
	}

	if encodeErr := json.NewEncoder(w).Encode(errorResponse); encodeErr != nil {
		h.logger.Error("Failed to encode error response", zap.Error(encodeErr))
	}
}

// Response types
type TaskListResponse struct {
	Tasks []*domain.Task `json:"tasks"`
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
	Code    int    `json:"code"`
}
