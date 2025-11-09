//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/config"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/constants"
	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
	postgresRepo "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/repository/postgres"
	redisRepo "github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/repository/redis"
)

// DatabaseIntegrationTestSuite tests database operations with real databases
type DatabaseIntegrationTestSuite struct {
	suite.Suite

	// Test containers
	postgresContainer *postgres.PostgresContainer
	redisContainer    *redis.RedisContainer

	// Repositories
	taskRepo  *postgresRepo.TaskRepository
	cacheRepo *redisRepo.CacheRepository

	// Test context
	ctx context.Context
}

func (suite *DatabaseIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Get test database credentials from environment or use defaults
	// NOTE: These are test-only values used with containerized databases
	testDBUser := os.Getenv("TEST_DB_USER")
	if testDBUser == "" {
		testDBUser = constants.TestDBUser // Safe test value for containerized testing
	}

	testDBPassword := os.Getenv("TEST_DB_PASSWORD")
	if testDBPassword == "" {
		testDBPassword = constants.TestDBPassword // Safe test value for containerized testing
	}

	testDBName := os.Getenv("TEST_DB_NAME")
	if testDBName == "" {
		testDBName = "test_mcp_ultra_wasm" // TEST_DB_NAME - safe test database name
	}

	// Start PostgreSQL container
	postgresContainer, err := postgres.RunContainer(suite.ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPassword),
		testcontainers.WithWaitStrategy(
			testcontainers.NewLogWaitStrategy("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	require.NoError(suite.T(), err)
	suite.postgresContainer = postgresContainer

	// Start Redis container
	redisContainer, err := redis.RunContainer(suite.ctx,
		testcontainers.WithImage("redis:7-alpine"),
	)
	require.NoError(suite.T(), err)
	suite.redisContainer = redisContainer

	// Setup database connection
	host, err := postgresContainer.Host(suite.ctx)
	require.NoError(suite.T(), err)

	port, err := postgresContainer.MappedPort(suite.ctx, "5432")
	require.NoError(suite.T(), err)

	pgConfig := config.PostgreSQLConfig{
		Host:     host,
		Port:     port.Int(),
		Database: testDBName,
		User:     testDBUser,
		Password: testDBPassword,
		SSLMode:  "disable",
	}

	db, err := postgresRepo.Connect(pgConfig)
	require.NoError(suite.T(), err)

	suite.taskRepo = postgresRepo.NewTaskRepository(db)

	// Setup Redis connection
	redisHost, err := redisContainer.Host(suite.ctx)
	require.NoError(suite.T(), err)

	redisPort, err := redisContainer.MappedPort(suite.ctx, "6379")
	require.NoError(suite.T(), err)

	redisConfig := config.RedisConfig{
		Addr: redisHost + ":" + redisPort.Port(),
		DB:   0,
	}

	redisClient := redisRepo.NewClient(redisConfig)
	suite.cacheRepo = redisRepo.NewCacheRepository(redisClient)

	// Run migrations
	err = postgresRepo.RunMigrations(db)
	require.NoError(suite.T(), err)
}

func (suite *DatabaseIntegrationTestSuite) TearDownSuite() {
	if suite.postgresContainer != nil {
		suite.postgresContainer.Terminate(suite.ctx)
	}
	if suite.redisContainer != nil {
		suite.redisContainer.Terminate(suite.ctx)
	}
}

func (suite *DatabaseIntegrationTestSuite) SetupTest() {
	// Clean database state before each test
	suite.cleanDatabase()
}

func (suite *DatabaseIntegrationTestSuite) cleanDatabase() {
	// Clean PostgreSQL
	db := suite.taskRepo.DB()
	_, err := db.Exec("TRUNCATE TABLE tasks CASCADE")
	require.NoError(suite.T(), err)

	// Clean Redis
	client := suite.cacheRepo.Client()
	err = client.FlushDB(suite.ctx).Err()
	require.NoError(suite.T(), err)
}

// Test complete task lifecycle with database persistence
func (suite *DatabaseIntegrationTestSuite) TestTaskLifecycleWithPersistence() {
	ctx := suite.ctx
	userID := uuid.New()

	// 1. Create task
	task := &domain.Task{
		ID:          uuid.New(),
		Title:       "Integration Test Task",
		Description: "Testing complete task lifecycle",
		Status:      domain.TaskStatusPending,
		Priority:    domain.PriorityHigh,
		Tags:        []string{"integration", "test"},
		CreatedBy:   userID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    map[string]interface{}{"test": true},
	}

	createdTask, err := suite.taskRepo.Create(ctx, task)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), task.Title, createdTask.Title)
	assert.Equal(suite.T(), task.Description, createdTask.Description)
	assert.NotEqual(suite.T(), uuid.Nil, createdTask.ID)

	// 2. Retrieve task
	retrievedTask, err := suite.taskRepo.GetByID(ctx, createdTask.ID)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdTask.Title, retrievedTask.Title)
	assert.Equal(suite.T(), createdTask.Description, retrievedTask.Description)
	assert.Equal(suite.T(), createdTask.Status, retrievedTask.Status)
	assert.Equal(suite.T(), createdTask.Priority, retrievedTask.Priority)

	// 3. Update task status
	retrievedTask.Status = domain.TaskStatusInProgress
	retrievedTask.UpdatedAt = time.Now()

	updatedTask, err := suite.taskRepo.Update(ctx, retrievedTask)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), domain.TaskStatusInProgress, updatedTask.Status)

	// 4. Complete task
	updatedTask.Complete()

	completedTask, err := suite.taskRepo.Update(ctx, updatedTask)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), domain.TaskStatusCompleted, completedTask.Status)
	assert.NotNil(suite.T(), completedTask.CompletedAt)

	// 5. List tasks with filters
	filter := &domain.TaskFilter{
		UserID: userID,
		Status: domain.TaskStatusCompleted,
		Limit:  10,
		Offset: 0,
	}

	tasks, total, err := suite.taskRepo.List(ctx, filter)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), tasks, 1)
	assert.Equal(suite.T(), int64(1), total)
	assert.Equal(suite.T(), completedTask.ID, tasks[0].ID)

	// 6. Delete task
	err = suite.taskRepo.Delete(ctx, completedTask.ID)
	require.NoError(suite.T(), err)

	// 7. Verify deletion
	_, err = suite.taskRepo.GetByID(ctx, completedTask.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), domain.ErrNotFound, err)
}

// Test cache-database synchronization
func (suite *DatabaseIntegrationTestSuite) TestCacheDatabaseSync() {
	ctx := suite.ctx
	userID := uuid.New()

	// Create task in database
	task := &domain.Task{
		ID:        uuid.New(),
		Title:     "Cache Test Task",
		CreatedBy: userID,
		Status:    domain.TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdTask, err := suite.taskRepo.Create(ctx, task)
	require.NoError(suite.T(), err)

	cacheKey := "task:" + createdTask.ID.String()

	// Initially not in cache
	_, err = suite.cacheRepo.Get(ctx, cacheKey)
	assert.Error(suite.T(), err)

	// Cache the task
	err = suite.cacheRepo.Set(ctx, cacheKey, createdTask, 300)
	require.NoError(suite.T(), err)

	// Retrieve from cache
	var cachedTask domain.Task
	err = suite.cacheRepo.Get(ctx, cacheKey)
	require.NoError(suite.T(), err)

	// Update database directly (simulating concurrent update)
	createdTask.Title = "Updated Title"
	createdTask.UpdatedAt = time.Now()
	updatedTask, err := suite.taskRepo.Update(ctx, createdTask)
	require.NoError(suite.T(), err)

	// Cache should be invalidated
	err = suite.cacheRepo.Delete(ctx, cacheKey)
	require.NoError(suite.T(), err)

	// Verify cache is empty
	_, err = suite.cacheRepo.Get(ctx, cacheKey)
	assert.Error(suite.T(), err)

	// Re-cache updated task
	err = suite.cacheRepo.Set(ctx, cacheKey, updatedTask, 300)
	require.NoError(suite.T(), err)

	// Verify updated data in cache
	err = suite.cacheRepo.Get(ctx, cacheKey)
	require.NoError(suite.T(), err)
}

// Test transaction rollback scenarios
func (suite *DatabaseIntegrationTestSuite) TestTransactionRollback() {
	ctx := suite.ctx
	userID := uuid.New()

	// Start transaction
	tx, err := suite.taskRepo.DB().BeginTx(ctx, nil)
	require.NoError(suite.T(), err)

	// Create task in transaction
	task := &domain.Task{
		ID:        uuid.New(),
		Title:     "Transaction Test Task",
		CreatedBy: userID,
		Status:    domain.TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// This would typically be done through a transactional repository method
	_, err = tx.ExecContext(ctx, `
		INSERT INTO tasks (id, title, description, status, priority, created_by, created_at, updated_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		task.ID, task.Title, task.Description, task.Status, task.Priority,
		task.CreatedBy, task.CreatedAt, task.UpdatedAt, task.Metadata)
	require.NoError(suite.T(), err)

	// Rollback transaction
	err = tx.Rollback()
	require.NoError(suite.T(), err)

	// Verify task was not persisted
	_, err = suite.taskRepo.GetByID(ctx, task.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), domain.ErrNotFound, err)
}

// Test concurrent database operations
func (suite *DatabaseIntegrationTestSuite) TestConcurrentOperations() {
	ctx := suite.ctx
	userID := uuid.New()
	numGoroutines := 10

	// Create initial task
	baseTask := &domain.Task{
		ID:        uuid.New(),
		Title:     "Concurrent Test Task",
		CreatedBy: userID,
		Status:    domain.TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdTask, err := suite.taskRepo.Create(ctx, baseTask)
	require.NoError(suite.T(), err)

	// Channel to collect results
	results := make(chan error, numGoroutines)

	// Launch concurrent updates
	for i := 0; i < numGoroutines; i++ {
		go func(iteration int) {
			// Get task
			task, err := suite.taskRepo.GetByID(ctx, createdTask.ID)
			if err != nil {
				results <- err
				return
			}

			// Modify task
			task.Title = fmt.Sprintf("Updated by goroutine %d", iteration)
			task.UpdatedAt = time.Now()

			// Update task
			_, err = suite.taskRepo.Update(ctx, task)
			results <- err
		}(i)
	}

	// Collect results
	errorCount := 0
	successCount := 0

	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err != nil {
			errorCount++
		} else {
			successCount++
		}
	}

	// At least one should succeed (last writer wins)
	assert.True(suite.T(), successCount >= 1, "At least one concurrent update should succeed")

	// Verify final state
	finalTask, err := suite.taskRepo.GetByID(ctx, createdTask.ID)
	require.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), baseTask.Title, finalTask.Title)
}

// Test large dataset operations
func (suite *DatabaseIntegrationTestSuite) TestLargeDatasetOperations() {
	ctx := suite.ctx
	userID := uuid.New()
	taskCount := 1000

	// Create many tasks
	tasks := make([]*domain.Task, taskCount)
	for i := 0; i < taskCount; i++ {
		task := &domain.Task{
			ID:          uuid.New(),
			Title:       fmt.Sprintf("Bulk Task %d", i),
			Description: fmt.Sprintf("Description for task %d", i),
			Status:      domain.TaskStatusPending,
			Priority:    domain.PriorityMedium,
			CreatedBy:   userID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Add some tags to every 10th task
		if i%10 == 0 {
			task.Tags = []string{fmt.Sprintf("batch-%d", i/10)}
		}

		tasks[i] = task
	}

	// Bulk insert (this would require implementing batch operations)
	start := time.Now()
	for _, task := range tasks {
		_, err := suite.taskRepo.Create(ctx, task)
		require.NoError(suite.T(), err)
	}
	insertDuration := time.Since(start)

	suite.T().Logf("Inserted %d tasks in %v (%.2f tasks/sec)",
		taskCount, insertDuration, float64(taskCount)/insertDuration.Seconds())

	// Test pagination with large dataset
	pageSize := 50
	totalRetrieved := 0

	for page := 0; page*pageSize < taskCount; page++ {
		filter := &domain.TaskFilter{
			UserID: userID,
			Limit:  pageSize,
			Offset: page * pageSize,
		}

		pageTasks, total, err := suite.taskRepo.List(ctx, filter)
		require.NoError(suite.T(), err)

		totalRetrieved += len(pageTasks)
		assert.Equal(suite.T(), int64(taskCount), total)

		if page == 0 {
			// First page should be full (unless taskCount < pageSize)
			expectedPageSize := min(pageSize, taskCount)
			assert.Len(suite.T(), pageTasks, expectedPageSize)
		}
	}

	assert.Equal(suite.T(), taskCount, totalRetrieved)

	// Test filtering on large dataset
	filter := &domain.TaskFilter{
		UserID: userID,
		Tags:   []string{"batch-10"}, // Should match one task
		Limit:  100,
	}

	filteredTasks, total, err := suite.taskRepo.List(ctx, filter)
	require.NoError(suite.T(), err)
	assert.Len(suite.T(), filteredTasks, 1)
	assert.Equal(suite.T(), int64(1), total)
	assert.Contains(suite.T(), filteredTasks[0].Tags, "batch-10")
}

// Test database constraints and validation
func (suite *DatabaseIntegrationTestSuite) TestDatabaseConstraints() {
	ctx := suite.ctx

	// Test duplicate ID constraint
	taskID := uuid.New()
	userID := uuid.New()

	task1 := &domain.Task{
		ID:        taskID,
		Title:     "First Task",
		CreatedBy: userID,
		Status:    domain.TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	task2 := &domain.Task{
		ID:        taskID, // Same ID
		Title:     "Second Task",
		CreatedBy: userID,
		Status:    domain.TaskStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// First insert should succeed
	_, err := suite.taskRepo.Create(ctx, task1)
	require.NoError(suite.T(), err)

	// Second insert with same ID should fail
	_, err = suite.taskRepo.Create(ctx, task2)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "duplicate")
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Run the test suite
func TestDatabaseIntegrationSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	suite.Run(t, new(DatabaseIntegrationTestSuite))
}
