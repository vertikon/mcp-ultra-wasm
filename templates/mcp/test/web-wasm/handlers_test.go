package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm/internal/wasm/nats"
)

// MockPublisher é um mock para o NATS Publisher
type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(subject string, data interface{}) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func (m *MockPublisher) PublishTask(task map[string]interface{}) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockPublisher) PublishTaskResult(taskID, correlationID string, result map[string]interface{}, err error) error {
	args := m.Called(taskID, correlationID, result, err)
	return args.Error(0)
}

func (m *MockPublisher) PublishTaskProgress(taskID, correlationID string, progress int, message string) error {
	args := m.Called(taskID, correlationID, progress, message)
	return args.Error(0)
}

func (m *MockPublisher) PublishTaskCancel(taskID string, reason string) error {
	args := m.Called(taskID, reason)
	return args.Error(0)
}

func (m *MockPublisher) PublishWASMEvent(eventType string, data map[string]interface{}) error {
	args := m.Called(eventType, data)
	return args.Error(0)
}

func (m *MockPublisher) PublishSystemEvent(eventType string, data map[string]interface{}) error {
	args := m.Called(eventType, data)
	return args.Error(0)
}

func (m *MockPublisher) Publish(subject string, data interface{}) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func (m *MockPublisher) PublishAsync(subject string, data interface{}) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func (m *MockPublisher) InitializeStreams() error {
	args := m.Called()
	return args.Error(0)
}

func TestUIHandler_ServeIndex(t *testing.T) {
	// Setup
	logger := zaptest.NewLogger(t)
	handler := NewUIHandler("./static", "./wasm", logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", handler.ServeIndex)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Perform request
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "<!DOCTYPE html>")
	assert.Contains(t, w.Body.String(), "MCP Ultra WASM")
	assert.Contains(t, w.Body.String(), "wasm-server")
}

func TestUIHandler_ServeWASM(t *testing.T) {
	// Setup
	logger := zaptest.NewLogger(t)
	handler := NewUIHandler("./static", "./wasm", logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/wasm/*filepath", handler.ServeWASM)

	// Test case: valid WASM file
	req := httptest.NewRequest(http.MethodGet, "/wasm/main.wasm", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should return 404 for now since the file doesn't exist
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test case: invalid path
	req = httptest.NewRequest(http.MethodGet, "/wasm/../../../etc/passwd", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPIHandler_CreateTask(t *testing.T) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	mockPublisher.On("PublishTask", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	logger := zaptest.NewLogger(t)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/tasks", handler.CreateTask)

	// Test case: valid request
	taskReq := TaskRequest{
		Type:     "analyze",
		Data:     map[string]interface{}{"project_path": "/test"},
		WASMFunc: "analyzeProject",
	}

	jsonBody, _ := json.Marshal(taskReq)
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusAccepted, w.Code)

	var response TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.CorrelationID)
	assert.Equal(t, "pending", response.Status)
	assert.Equal(t, "analyze", response.Type)

	// Verify mock was called correctly
	mockPublisher.AssertExpectations(t)
}

func TestAPIHandler_CreateTask_InvalidRequest(t *testing.T) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	logger := zaptest.NewLogger(t)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/tasks", handler.CreateTask)

	// Test case: invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

func TestAPIHandler_CreateTask_MissingFields(t *testing.T) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	logger := zaptest.NewLogger(t)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/tasks", handler.CreateTask)

	// Test case: missing type field
	taskReq := map[string]interface{}{
		"data": map[string]interface{}{"project_path": "/test"},
	}

	jsonBody, _ := json.Marshal(taskReq)
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "required")
}

func TestAPIHandler_GetTask(t *testing.T) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	logger := zaptest.NewLogger(t)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tasks/:id", handler.GetTask)

	// Test case: valid task ID
	taskID := uuid.New().String()
	req := httptest.NewRequest(http.MethodGet, "/tasks/"+taskID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, taskID, response.ID)
	assert.Equal(t, "completed", response.Status)
	assert.NotNil(t, response.Data)
}

func TestAPIHandler_GetTask_InvalidID(t *testing.T) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	logger := zaptest.NewLogger(t)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tasks/:id", handler.GetTask)

	// Test case: empty task ID
	req := httptest.NewRequest(http.MethodGet, "/tasks/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should get 404 for empty ID
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPIHandler_ListTasks(t *testing.T) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	logger := zaptest.NewLogger(t)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tasks", handler.ListTasks)

	// Test case: default parameters
	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "tasks")
	assert.Contains(t, response, "total")
	assert.Equal(t, "1", response["page"])
	assert.Equal(t, "20", response["limit"])
}

func TestAPIHandler_ListTasks_WithParameters(t *testing.T) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	logger := zaptest.NewLogger(t)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/tasks", handler.ListTasks)

	// Test case: custom parameters
	req := httptest.NewRequest(http.MethodGet, "/tasks?page=2&limit=10&status=running", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "2", response["page"])
	assert.Equal(t, "10", response["limit"])
}

func TestAPIHandler_CancelTask(t *testing.T) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	mockPublisher.On("PublishTaskCancel", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

	logger := zaptest.NewLogger(t)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/tasks/:id", handler.CancelTask)

	// Test case: valid task ID
	taskID := uuid.New().String()
	req := httptest.NewRequest(http.MethodDelete, "/tasks/"+taskID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Task cancelada com sucesso", response["message"])
	assert.Equal(t, taskID, response["task_id"])

	// Verify mock was called correctly
	mockPublisher.AssertExpectations(t)
}

func TestAPIHandler_CancelTask_InvalidID(t *testing.T) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	logger := zaptest.NewLogger(t)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/tasks/:id", handler.CancelTask)

	// Test case: empty task ID
	req := httptest.NewRequest(http.MethodDelete, "/tasks/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should get 404 for empty ID
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPIHandler_GetHealth(t *testing.T) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	logger := zaptest.NewLogger(t)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/health", handler.GetHealth)

	// Test case: health check
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "wasm-api", response["service"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.NotEmpty(t, response["timestamp"])
}

// Benchmark tests
func BenchmarkAPIHandler_CreateTask(b *testing.B) {
	// Setup mock
	mockPublisher := &MockPublisher{}
	mockPublisher.On("PublishTask", mock.AnythingOfType("map[string]interface {}")).Return(nil)

	logger := zaptest.NewLogger(b)
	handler := NewAPIHandler(mockPublisher, logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/tasks", handler.CreateTask)

	// Prepare request
	taskReq := TaskRequest{
		Type:     "analyze",
		Data:     map[string]interface{}{"project_path": "/test"},
		WASMFunc: "analyzeProject",
	}
	jsonBody, _ := json.Marshal(taskReq)

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkUIHandler_ServeIndex(b *testing.B) {
	// Setup
	logger := zaptest.NewLogger(b)
	handler := NewUIHandler("./static", "./wasm", logger)

	// Create gin router with the handler
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", handler.ServeIndex)

	// Prepare request
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
