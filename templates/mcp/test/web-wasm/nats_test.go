package nats

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm/internal/wasm/nats"
)

// MockConnection é um mock para a conexão NATS
type MockConnection struct {
	mock.Mock
}

func (m *MockConnection) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	args := m.Called(subject, handler)
	return args.Get(0).(*nats.Subscription), args.Error(1)
}

func (m *MockConnection) QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	args := m.Called(subject, queue, handler)
	return args.Get(0).(*nats.Subscription), args.Error(1)
}

func (m *MockConnection) Publish(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func (m *MockConnection) PublishAsync(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func (m *MockConnection) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	args := m.Called(subject, data, timeout)
	return args.Get(0).(*nats.Msg), args.Error(1)
}

func (m *MockConnection) Flush() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConnection) IsClosed() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockConnection) Stats() nats.Stats {
	args := m.Called()
	return args.Get(0).(nats.Stats)
}

func (m *MockConnection) ConnectedServers() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockConnection) ConnectedUrl() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockConnection) Status() nats.ConnStatus {
	args := m.Called()
	return args.Get(0).(nats.ConnStatus)
}

// MockJetStreamContext é um mock para o JetStream
type MockJetStreamContext struct {
	mock.Mock
}

func (m *MockJetStreamContext) AddStream(cfg *nats.StreamConfig) (*nats.StreamInfo, error) {
	args := m.Called(cfg)
	return args.Get(0).(*nats.StreamInfo), args.Error(1)
}

func (m *MockJetStreamContext) UpdateStream(cfg *nats.StreamConfig) (*nats.StreamInfo, error) {
	args := m.Called(cfg)
	return args.Get(0).(*nats.StreamInfo), args.Error(1)
}

func (m *MockJetStreamContext) DeleteStream(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockJetStreamContext) StreamInfo(name string) (*nats.StreamInfo, error) {
	args := m.Called(name)
	return args.Get(0).(*nats.StreamInfo), args.Error(1)
}

func (m *MockJetStreamContext) StreamsInfo() (*nats.StreamInfoList, error) {
	args := m.Called()
	return args.Get(0).(*nats.StreamInfoList), args.Error(1)
}

func (m *MockJetStreamContext) AddConsumer(stream string, cfg *nats.ConsumerConfig) (*nats.ConsumerInfo, error) {
	args := m.Called(stream, cfg)
	return args.Get(0).(*nats.ConsumerInfo), args.Error(1)
}

func (m *MockJetStreamContext) UpdateConsumer(stream string, cfg *nats.ConsumerConfig) (*nats.ConsumerInfo, error) {
	args := m.Called(stream, cfg)
	return args.Get(0).(*nats.ConsumerInfo), args.Error(1)
}

func (m *MockJetStreamContext) DeleteConsumer(stream, name string) error {
	args := m.Called(stream, name)
	return args.Error(0)
}

func (m *MockJetStreamContext) ConsumerInfo(stream string, name string) (*nats.ConsumerInfo, error) {
	args := m.Called(stream, name)
	return args.Get(0).(*nats.ConsumerInfo), args.Error(1)
}

func (m *MockJetStreamContext) ConsumersInfo(stream string) (*nats.ConsumerInfoList, error) {
	args := m.Called(stream)
	return args.Get(0).(*nats.ConsumerInfoList), args.Error(1)
}

func TestNewClient_Success(t *testing.T) {
	// Este teste requer uma conexão NATS real
	// Para ambientes de CI/CD, pode ser pulado ou mockado
	t.Skip("Requer conexão NATS real - implementar mock se necessário")

	logger := zaptest.NewLogger(t)
	client, err := NewClient("nats://localhost:4222", logger)

	if err != nil {
		t.Skipf("NATS não disponível: %v", err)
		return
	}

	require.NotNil(t, client)
	assert.NotNil(t, client.GetConnection())
	assert.NotNil(t, client.GetJetStream())

	// Cleanup
	client.Close()
}

func TestNewClient_InvalidURL(t *testing.T) {
	logger := zaptest.NewLogger(t)
	client, err := NewClient("invalid-url", logger)

	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestClient_GetConnection(t *testing.T) {
	// Setup mock connection
	mockConn := &MockConnection{}
	logger := zaptest.NewLogger(t)

	client := &Client{
		conn:   mockConn,
		logger: logger,
	}

	// Test
	conn := client.GetConnection()
	assert.Equal(t, mockConn, conn)
}

func TestClient_GetJetStream(t *testing.T) {
	// Setup mock jetstream
	mockJS := &MockJetStreamContext{}
	logger := zaptest.NewLogger(t)

	client := &Client{
		js:     mockJS,
		logger: logger,
	}

	// Test
	js := client.GetJetStream()
	assert.Equal(t, mockJS, js)
}

func TestClient_CreateStream(t *testing.T) {
	// Setup mock jetstream
	mockJS := &MockJetStreamContext{}
	mockStreamInfo := &nats.StreamInfo{Config: nats.StreamConfig{Name: "test-stream"}}
	mockJS.On("StreamInfo", "test-stream").Return(mockStreamInfo, nil)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(mockStreamInfo, nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		js:     mockJS,
		logger: logger,
	}

	// Test
	config := nats.StreamConfig{
		Name:     "test-stream",
		Subjects: []string{"test.>"},
		Storage:  nats.FileStorage,
	}

	err := client.CreateStream(config)

	// Assertions
	assert.NoError(t, err)
	mockJS.AssertExpectations(t)
}

func TestClient_CreateStream_AlreadyExists(t *testing.T) {
	// Setup mock jetstream
	mockJS := &MockJetStreamContext{}
	mockStreamInfo := &nats.StreamInfo{Config: nats.StreamConfig{Name: "test-stream"}}
	mockJS.On("StreamInfo", "test-stream").Return(mockStreamInfo, nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		js:     mockJS,
		logger: logger,
	}

	// Test
	config := nats.StreamConfig{
		Name:     "test-stream",
		Subjects: []string{"test.>"},
		Storage:  nats.FileStorage,
	}

	err := client.CreateStream(config)

	// Assertions
	assert.NoError(t, err) // Should not error if stream already exists
	mockJS.AssertExpectations(t)
}

func TestClient_CreateStream_Error(t *testing.T) {
	// Setup mock jetstream
	mockJS := &MockJetStreamContext{}
	mockJS.On("StreamInfo", "test-stream").Return(nil, nats.ErrStreamNotFound)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(nil, assert.AnError)

	logger := zaptest.NewLogger(t)
	client := &Client{
		js:     mockJS,
		logger: logger,
	}

	// Test
	config := nats.StreamConfig{
		Name:     "test-stream",
		Subjects: []string{"test.>"},
		Storage:  nats.FileStorage,
	}

	err := client.CreateStream(config)

	// Assertions
	assert.Error(t, err)
	mockJS.AssertExpectations(t)
}

func TestClient_Publish(t *testing.T) {
	// Setup mock connection
	mockConn := &MockConnection{}
	mockConn.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		logger: logger,
	}

	// Test
	subject := "test.subject"
	data := map[string]interface{}{"key": "value"}

	err := client.Publish(subject, data)

	// Assertions
	assert.NoError(t, err)
	mockConn.AssertCalled(t, "Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"))
}

func TestClient_Publish_Error(t *testing.T) {
	// Setup mock connection
	mockConn := &MockConnection{}
	mockConn.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(assert.AnError)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		logger: logger,
	}

	// Test
	subject := "test.subject"
	data := map[string]interface{}{"key": "value"}

	err := client.Publish(subject, data)

	// Assertions
	assert.Error(t, err)
	mockConn.AssertCalled(t, "Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"))
}

func TestClient_PublishAsync(t *testing.T) {
	// Setup mock connection
	mockConn := &MockConnection{}
	mockConn.On("PublishAsync", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		logger: logger,
	}

	// Test
	subject := "test.subject"
	data := map[string]interface{}{"key": "value"}

	err := client.PublishAsync(subject, data)

	// Assertions
	assert.NoError(t, err)
	mockConn.AssertCalled(t, "PublishAsync", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"))
}

func TestClient_Request(t *testing.T) {
	// Setup mock connection
	mockConn := &MockConnection{}
	mockMsg := &nats.Msg{Data: []byte("response")}
	mockConn.On("Request", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"), mock.AnythingOfType("time.Duration")).Return(mockMsg, nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		logger: logger,
	}

	// Test
	subject := "test.subject"
	data := map[string]interface{}{"key": "value"}
	timeout := 5 * time.Second

	msg, err := client.Request(subject, data, timeout)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, mockMsg, msg)
	assert.Equal(t, []byte("response"), msg.Data)
	mockConn.AssertCalled(t, "Request", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8"), mock.AnythingOfType("time.Duration"))
}

func TestClient_Health(t *testing.T) {
	// Setup mock connection
	mockConn := &MockConnection{}
	mockConn.On("IsClosed").Return(false)
	mockConn.On("Status").Return(nats.ConnStatus(nats.CONNECTED))

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		logger: logger,
	}

	// Test
	err := client.Health()

	// Assertions
	assert.NoError(t, err)
	mockConn.AssertCalled(t, "IsClosed")
	mockConn.AssertCalled(t, "Status")
}

func TestClient_Health_Closed(t *testing.T) {
	// Setup mock connection
	mockConn := &MockConnection{}
	mockConn.On("IsClosed").Return(true)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		logger: logger,
	}

	// Test
	err := client.Health()

	// Assertions
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "conexão NATS está fechada")
	mockConn.AssertCalled(t, "IsClosed")
}

func TestClient_GetStats(t *testing.T) {
	// Setup mock connection
	mockConn := &MockConnection{}
	expectedStats := nats.Stats{
		Connects:   1,
		Reconnects: 0,
		Errors:     0,
		InMsgs:     10,
		OutMsgs:    5,
		InBytes:    1024,
		OutBytes:   512,
	}
	mockConn.On("Stats").Return(expectedStats)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		logger: logger,
	}

	// Test
	stats := client.GetStats()

	// Assertions
	assert.NotNil(t, stats)
	assert.Equal(t, expectedStats.Connects, stats["connects"])
	assert.Equal(t, expectedStats.Reconnects, stats["reconnects"])
	assert.Equal(t, expectedStats.Errors, stats["errors"])
	assert.Equal(t, expectedStats.InMsgs, stats["in_msgs"])
	assert.Equal(t, expectedStats.OutMsgs, stats["out_msgs"])
	assert.Equal(t, expectedStats.InBytes, stats["in_bytes"])
	assert.Equal(t, expectedStats.OutBytes, stats["out_bytes"])
	mockConn.AssertCalled(t, "Stats")
}

func TestPublisher_PublishTask(t *testing.T) {
	// Setup mock client
	mockConn := &MockConnection{}
	mockConn.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	mockJS := &MockJetStreamContext{}
	mockStreamInfo := &nats.StreamInfo{Config: nats.StreamConfig{Name: "WEB_WASM_TASKS"}}
	mockJS.On("StreamInfo", "WEB_WASM_TASKS").Return(mockStreamInfo, nil)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(mockStreamInfo, nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		js:     mockJS,
		logger: logger,
	}
	publisher := NewPublisher(client, logger)

	// Initialize streams
	err := publisher.InitializeStreams()
	require.NoError(t, err)

	// Test
	task := map[string]interface{}{
		"id":   "test-task-id",
		"type": "analyze",
		"data": map[string]interface{}{"project": "test"},
	}

	err = publisher.PublishTask(task)

	// Assertions
	assert.NoError(t, err)
	mockConn.AssertCalled(t, "Publish")
}

func TestPublisher_PublishTaskResult(t *testing.T) {
	// Setup mock client
	mockConn := &MockConnection{}
	mockConn.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	mockJS := &MockJetStreamContext{}
	mockStreamInfo := &nats.StreamInfo{Config: nats.StreamConfig{Name: "WEB_WASM_TASKS"}}
	mockJS.On("StreamInfo", "WEB_WASM_TASKS").Return(mockStreamInfo, nil)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(mockStreamInfo, nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		js:     mockJS,
		logger: logger,
	}
	publisher := NewPublisher(client, logger)

	// Initialize streams
	err := publisher.InitializeStreams()
	require.NoError(t, err)

	// Test
	taskID := "test-task-id"
	correlationID := "test-correlation-id"
	result := map[string]interface{}{"output": "success"}

	err = publisher.PublishTaskResult(taskID, correlationID, result, nil)

	// Assertions
	assert.NoError(t, err)
	mockConn.AssertCalled(t, "Publish")
}

func TestPublisher_PublishTaskError(t *testing.T) {
	// Setup mock client
	mockConn := &MockConnection{}
	mockConn.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	mockJS := &MockJetStreamContext{}
	mockStreamInfo := &nats.StreamInfo{Config: nats.StreamConfig{Name: "WEB_WASM_TASKS"}}
	mockJS.On("StreamInfo", "WEB_WASM_TASKS").Return(mockStreamInfo, nil)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(mockStreamInfo, nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		js:     mockJS,
		logger: logger,
	}
	publisher := NewPublisher(client, logger)

	// Initialize streams
	err := publisher.InitializeStreams()
	require.NoError(t, err)

	// Test
	taskID := "test-task-id"
	correlationID := "test-correlation-id"
	taskErr := assert.AnError

	err = publisher.PublishTaskResult(taskID, correlationID, nil, taskErr)

	// Assertions
	assert.NoError(t, err)
	mockConn.AssertCalled(t, "Publish")
}

func TestPublisher_PublishTaskProgress(t *testing.T) {
	// Setup mock client
	mockConn := &MockConnection{}
	mockConn.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	mockJS := &MockJetStreamContext{}
	mockStreamInfo := &nats.StreamInfo{Config: nats.StreamConfig{Name: "WEB_WASM_TASKS"}}
	mockJS.On("StreamInfo", "WEB_WASM_TASKS").Return(mockStreamInfo, nil)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(mockStreamInfo, nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		js:     mockJS,
		logger: logger,
	}
	publisher := NewPublisher(client, logger)

	// Initialize streams
	err := publisher.InitializeStreams()
	require.NoError(t, err)

	// Test
	taskID := "test-task-id"
	correlationID := "test-correlation-id"
	progress := 50
	message := "Processing step 1/2"

	err = publisher.PublishTaskProgress(taskID, correlationID, progress, message)

	// Assertions
	assert.NoError(t, err)
	mockConn.AssertCalled(t, "Publish")
}

func TestPublisher_PublishWASMEvent(t *testing.T) {
	// Setup mock client
	mockConn := &MockConnection{}
	mockConn.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	mockJS := &MockJetStreamContext{}
	mockStreamInfo := &nats.StreamInfo{Config: nats.StreamConfig{Name: "WEB_WASM_EVENTS"}}
	mockJS.On("StreamInfo", "WEB_WASM_EVENTS").Return(mockStreamInfo, nil)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(mockStreamInfo, nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		js:     mockJS,
		logger: logger,
	}
	publisher := NewPublisher(client, logger)

	// Initialize streams
	err := publisher.InitializeStreams()
	require.NoError(t, err)

	// Test
	eventType := "module_loaded"
	data := map[string]interface{}{
		"module": "test-module",
		"size":   1024,
	}

	err = publisher.PublishWASMEvent(eventType, data)

	// Assertions
	assert.NoError(t, err)
	mockConn.AssertCalled(t, "Publish")
}

func TestPublisher_InitializeStreams(t *testing.T) {
	// Setup mock client
	mockConn := &MockConnection{}

	mockJS := &MockJetStreamContext{}
	mockStreamInfo := &nats.StreamInfo{Config: nats.StreamConfig{Name: "WEB_WASM_TASKS"}}
	mockJS.On("StreamInfo", "WEB_WASM_TASKS").Return(nil, nats.ErrStreamNotFound)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(mockStreamInfo, nil)

	mockStreamInfo2 := &nats.StreamInfo{Config: nats.StreamConfig{Name: "WEB_WASM_EVENTS"}}
	mockJS.On("StreamInfo", "WEB_WASM_EVENTS").Return(nil, nats.ErrStreamNotFound)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(mockStreamInfo2, nil)

	mockStreamInfo3 := &nats.StreamInfo{Config: nats.StreamConfig{Name: "WEB_WASM_SYSTEM"}}
	mockJS.On("StreamInfo", "WEB_WASM_SYSTEM").Return(nil, nats.ErrStreamNotFound)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(mockStreamInfo3, nil)

	logger := zaptest.NewLogger(t)
	client := &Client{
		conn:   mockConn,
		js:     mockJS,
		logger: logger,
	}
	publisher := NewPublisher(client, logger)

	// Test
	err := publisher.InitializeStreams()

	// Assertions
	assert.NoError(t, err)
	mockJS.AssertCalled(t, "StreamInfo", "WEB_WASM_TASKS")
	mockJS.AssertCalled(t, "AddStream", mock.AnythingOfType("*nats.StreamConfig"))
	mockJS.AssertCalled(t, "StreamInfo", "WEB_WASM_EVENTS")
	mockJS.AssertCalled(t, "AddStream", mock.AnythingOfType("*nats.StreamConfig"))
	mockJS.AssertCalled(t, "StreamInfo", "WEB_WASM_SYSTEM")
	mockJS.AssertCalled(t, "AddStream", mock.AnythingOfType("*nats.StreamConfig"))
}

// Benchmark tests
func BenchmarkClient_Publish(b *testing.B) {
	// Setup mock connection
	mockConn := &MockConnection{}
	mockConn.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	logger := zaptest.NewLogger(b)
	client := &Client{
		conn:   mockConn,
		logger: logger,
	}

	// Prepare data
	subject := "test.subject"
	data := map[string]interface{}{"key": "value"}

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Publish(subject, data)
	}
}

func BenchmarkPublisher_PublishTask(b *testing.B) {
	// Setup mock client
	mockConn := &MockConnection{}
	mockConn.On("Publish", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)

	mockJS := &MockJetStreamContext{}
	mockStreamInfo := &nats.StreamInfo{Config: nats.StreamConfig{Name: "WEB_WASM_TASKS"}}
	mockJS.On("StreamInfo", "WEB_WASM_TASKS").Return(mockStreamInfo, nil)
	mockJS.On("AddStream", mock.AnythingOfType("*nats.StreamConfig")).Return(mockStreamInfo, nil)

	logger := zaptest.NewLogger(b)
	client := &Client{
		conn:   mockConn,
		js:     mockJS,
		logger: logger,
	}
	publisher := NewPublisher(client, logger)

	// Initialize streams
	publisher.InitializeStreams()

	// Prepare task
	task := map[string]interface{}{
		"id":   "test-task-id",
		"type": "analyze",
		"data": map[string]interface{}{"project": "test"},
	}

	// Benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		publisher.PublishTask(task)
	}
}
