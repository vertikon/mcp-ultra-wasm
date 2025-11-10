package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm/internal/wasm/nats"
)

type APIHandler struct {
	publisher *nats.Publisher
	logger    *zap.Logger
}

type TaskRequest struct {
	Type      string                 `json:"type" binding:"required"`
	Data      map[string]interface{} `json:"data"`
	WASMFunc  string                 `json:"wasm_function,omitempty"`
	Priority  string                 `json:"priority,omitempty"`
	Metadata  map[string]string      `json:"metadata,omitempty"`
}

type TaskResponse struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       string                 `json:"error,omitempty"`
	CorrelationID string               `json:"correlation_id"`
}

func NewAPIHandler(publisher *nats.Publisher, logger *zap.Logger) *APIHandler {
	return &APIHandler{
		publisher: publisher,
		logger:    logger,
	}
}

// CreateTask cria uma nova task e publica no NATS
func (h *APIHandler) CreateTask(c *gin.Context) {
	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Erro ao validar request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Gerar IDs
	taskID := uuid.New().String()
	correlationID := uuid.New().String()

	// Criar task
	task := map[string]interface{}{
		"id":             taskID,
		"correlation_id": correlationID,
		"type":           req.Type,
		"data":           req.Data,
		"wasm_function":  req.WASMFunc,
		"priority":       req.Priority,
		"metadata":       req.Metadata,
		"created_at":     time.Now().UTC(),
		"status":         "pending",
	}

	// Publicar no NATS
	subject := "web.wasm.task.requested"
	if err := h.publisher.Publish(subject, task); err != nil {
		h.logger.Error("Erro ao publicar task no NATS", 
			zap.String("subject", subject),
			zap.String("task_id", taskID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao criar task"})
		return
	}

	h.logger.Info("Task criada com sucesso", 
		zap.String("task_id", taskID),
		zap.String("correlation_id", correlationID),
		zap.String("type", req.Type))

	// Retornar response imediata
	c.JSON(http.StatusAccepted, TaskResponse{
		ID:            taskID,
		Status:        "pending",
		CreatedAt:     time.Now().UTC(),
		CorrelationID: correlationID,
	})
}

// GetTask retorna informações de uma task específica
func (h *APIHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da task é obrigatório"})
		return
	}

	// TODO: Implementar busca de task no storage
	// Por enquanto, retorna resposta simulada
	c.JSON(http.StatusOK, TaskResponse{
		ID:        taskID,
		Status:    "completed",
		CreatedAt: time.Now().UTC().Add(-5 * time.Minute),
		Data: map[string]interface{}{
			"result": "Task executada com sucesso",
		},
	})
}

// ListTasks lista todas as tasks
func (h *APIHandler) ListTasks(c *gin.Context) {
	// Parâmetros de paginação
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "20")
	status := c.Query("status")

	h.logger.Info("Listando tasks", 
		zap.String("page", page),
		zap.String("limit", limit),
		zap.String("status", status))

	// TODO: Implementar listagem real no storage
	// Por enquanto, retorna lista simulada
	tasks := []TaskResponse{
		{
			ID:            uuid.New().String(),
			Status:        "completed",
			CreatedAt:     time.Now().UTC().Add(-10 * time.Minute),
			CorrelationID: uuid.New().String(),
			Data: map[string]interface{}{
				"result": "Análise concluída",
			},
		},
		{
			ID:            uuid.New().String(),
			Status:        "running",
			CreatedAt:     time.Now().UTC().Add(-2 * time.Minute),
			CorrelationID: uuid.New().String(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"total": len(tasks),
		"page":  page,
		"limit": limit,
	})
}

// CancelTask cancela uma task
func (h *APIHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da task é obrigatório"})
		return
	}

	 Criar evento de cancelamento
	cancelEvent := map[string]interface{}{
		"id":         taskID,
		"action":     "cancel",
		"cancelled_at": time.Now().UTC(),
		"reason":     "user_request",
	}

	// Publicar no NATS
	subject := "web.wasm.task.cancelled"
	if err := h.publisher.Publish(subject, cancelEvent); err != nil {
		h.logger.Error("Erro ao publicar cancelamento no NATS", 
			zap.String("subject", subject),
			zap.String("task_id", taskID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao cancelar task"})
		return
	}

	h.logger.Info("Task cancelada", zap.String("task_id", taskID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Task cancelada com sucesso",
		"task_id": taskID,
	})
}

// GetHealth retorna status da API
func (h *APIHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "wasm-api",
		"version":   "1.0.0",
	})
}