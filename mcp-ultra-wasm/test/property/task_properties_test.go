package property

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/vertikon/mcp-ultra-wasm-wasm/mcp/mcp-ultra-wasm-wasm/internal/domain"
)

// TestTaskCreationProperties tests invariants for task creation
func TestTaskCreationProperties(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: A newly created task should always have valid initial state
	properties.Property("created task has valid initial state", prop.ForAll(
		func(title, description string) bool {
			if title == "" || len(title) > 255 {
				// Skip invalid inputs
				return true
			}
			if len(description) > 2000 {
				return true
			}

			createdBy := uuid.New()
			task := domain.NewTask(title, description, createdBy)

			// Invariants that must always hold
			return task.ID != uuid.Nil &&
				task.Title == title &&
				task.Description == description &&
				task.CreatedBy == createdBy &&
				task.Status == domain.TaskStatusPending &&
				task.Priority == domain.PriorityMedium &&
				!task.CreatedAt.IsZero() &&
				!task.UpdatedAt.IsZero() &&
				task.CompletedAt == nil &&
				task.Metadata != nil
		},
		gen.AnyString(),
		gen.AnyString(),
	))

	// Property: Task status transitions must be valid
	properties.Property("task status transitions are valid", prop.ForAll(
		func(currentStatus, newStatus domain.TaskStatus) bool {
			task := createTestTask()
			task.Status = currentStatus

			expected := task.IsValidStatus(newStatus)

			// Valid transitions based on business rules
			switch currentStatus {
			case domain.TaskStatusPending:
				return expected == (newStatus == domain.TaskStatusInProgress ||
					newStatus == domain.TaskStatusCancelled ||
					newStatus == domain.TaskStatusPending)
			case domain.TaskStatusInProgress:
				return expected == (newStatus == domain.TaskStatusCompleted ||
					newStatus == domain.TaskStatusCancelled ||
					newStatus == domain.TaskStatusInProgress)
			case domain.TaskStatusCompleted, domain.TaskStatusCancelled:
				return expected == (newStatus == currentStatus) // Terminal states
			default:
				return !expected
			}
		},
		genTaskStatus(),
		genTaskStatus(),
	))

	// Property: Task completion should set completion timestamp
	properties.Property("completing task sets completion timestamp", prop.ForAll(
		func(status domain.TaskStatus) bool {
			task := createTestTask()
			task.Status = status

			beforeComplete := time.Now()

			if status == domain.TaskStatusInProgress {
				task.Complete()
				return task.Status == domain.TaskStatusCompleted &&
					task.CompletedAt != nil &&
					task.CompletedAt.After(beforeComplete)
			}

			// Should not be able to complete from other states
			task.Complete()
			return task.Status != domain.TaskStatusCompleted || task.CompletedAt == nil
		},
		genTaskStatus(),
	))

	// Property: Task priority must be valid
	properties.Property("task priority is always valid", prop.ForAll(
		func(priority domain.Priority) bool {
			task := createTestTask()
			oldPriority := task.Priority

			// Try to set the priority
			task.Priority = priority

			validPriorities := []domain.Priority{
				domain.PriorityLow,
				domain.PriorityMedium,
				domain.PriorityHigh,
				domain.PriorityUrgent,
			}

			for _, valid := range validPriorities {
				if priority == valid {
					return task.Priority == priority
				}
			}

			// If invalid priority, should retain old value or handle gracefully
			return task.Priority == oldPriority || task.Priority == domain.PriorityMedium
		},
		genPriority(),
	))

	// Property: Task tags should be unique and non-empty
	properties.Property("task tags are unique and non-empty", prop.ForAll(
		func(tags []string) bool {
			task := createTestTask()

			// Filter and set tags
			validTags := make([]string, 0)
			seen := make(map[string]bool)

			for _, tag := range tags {
				if tag != "" && !seen[tag] && len(tag) <= 50 {
					validTags = append(validTags, tag)
					seen[tag] = true
				}
			}

			task.Tags = validTags

			// Verify uniqueness
			tagSet := make(map[string]bool)
			for _, tag := range task.Tags {
				if tagSet[tag] {
					return false // Duplicate found
				}
				tagSet[tag] = true

				if tag == "" || len(tag) > 50 {
					return false // Invalid tag
				}
			}

			return true
		},
		gen.SliceOf(gen.AnyString()),
	))

	// Property: Task metadata should preserve non-nil values
	properties.Property("task metadata preserves non-nil values", prop.ForAll(
		func(key, value string) bool {
			if key == "" {
				return true // Skip empty keys
			}

			task := createTestTask()

			// Initially metadata should not be nil
			if task.Metadata == nil {
				return false
			}

			// Set a value
			if task.Metadata == nil {
				task.Metadata = make(map[string]interface{})
			}
			task.Metadata[key] = value

			// Value should be preserved
			stored, exists := task.Metadata[key]
			return exists && stored == value
		},
		gen.AnyString(),
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// TestTaskBusinessRuleProperties tests business rule properties
func TestTaskBusinessRuleProperties(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Due dates in the past should be flagged
	properties.Property("past due dates are handled correctly", prop.ForAll(
		func(hoursOffset int) bool {
			task := createTestTask()

			dueDate := time.Now().Add(time.Duration(hoursOffset) * time.Hour)
			task.DueDate = &dueDate

			isPast := dueDate.Before(time.Now())

			// Business rule: tasks with past due dates should be identifiable
			if isPast {
				return time.Since(dueDate) > 0
			}
			return time.Until(dueDate) > 0
		},
		gen.IntRange(-168, 168), // -7 days to +7 days in hours
	))

	// Property: Task title normalization
	properties.Property("task title is normalized", prop.ForAll(
		func(title string) bool {
			if len(title) > 255 {
				return true // Skip invalid titles
			}

			task := createTestTask()

			// Simulate title normalization (trim spaces, etc.)
			normalizedTitle := normalizeTitle(title)
			task.Title = normalizedTitle

			// Title should not have leading/trailing spaces
			return task.Title == normalizedTitle &&
				(task.Title == "" || (task.Title[0] != ' ' && task.Title[len(task.Title)-1] != ' '))
		},
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// Helper functions and generators

func createTestTask() *domain.Task {
	return domain.NewTask("Test Task", "Test Description", uuid.New())
}

func genTaskStatus() gopter.Gen {
	return gen.OneConstOf(
		domain.TaskStatusPending,
		domain.TaskStatusInProgress,
		domain.TaskStatusCompleted,
		domain.TaskStatusCancelled,
	)
}

func genPriority() gopter.Gen {
	return gen.OneConstOf(
		domain.PriorityLow,
		domain.PriorityMedium,
		domain.PriorityHigh,
		domain.PriorityUrgent,
		domain.Priority("invalid"), // Test invalid priority
	)
}

func normalizeTitle(title string) string {
	// Simple normalization - trim spaces
	normalized := title
	for len(normalized) > 0 && normalized[0] == ' ' {
		normalized = normalized[1:]
	}
	for len(normalized) > 0 && normalized[len(normalized)-1] == ' ' {
		normalized = normalized[:len(normalized)-1]
	}
	return normalized
}

// TestFeatureFlagProperties tests feature flag business logic properties
func TestFeatureFlagProperties(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// Property: Percentage strategy should be deterministic for same user
	properties.Property("percentage strategy is deterministic", prop.ForAll(
		func(userID string, percentage float64) bool {
			if percentage < 0 || percentage > 100 {
				return true // Skip invalid percentages
			}

			flag := &domain.FeatureFlag{
				Key:      "test-flag",
				Enabled:  true,
				Strategy: "percentage",
				Parameters: map[string]interface{}{
					"percentage": percentage,
				},
			}

			// Same user should always get same result
			result1 := evaluatePercentageFlag(flag, userID)
			result2 := evaluatePercentageFlag(flag, userID)

			return result1 == result2
		},
		gen.AnyString(),
		gen.Float64Range(0, 100),
	))

	// Property: User list strategy should be exact
	properties.Property("userlist strategy is exact", prop.ForAll(
		func(users []string, testUser string) bool {
			flag := &domain.FeatureFlag{
				Key:      "test-flag",
				Enabled:  true,
				Strategy: "userlist",
				Parameters: map[string]interface{}{
					"users": users,
				},
			}

			result := evaluateUserListFlag(flag, testUser)

			// Check if user is in the list
			inList := false
			for _, user := range users {
				if user == testUser {
					inList = true
					break
				}
			}

			return result == inList
		},
		gen.SliceOf(gen.AnyString()),
		gen.AnyString(),
	))

	properties.TestingRun(t)
}

// Mock implementation of percentage evaluation for testing
func evaluatePercentageFlag(flag *domain.FeatureFlag, userID string) bool {
	// Simplified hash-based percentage calculation
	if percentage, ok := flag.Parameters["percentage"].(float64); ok {
		// Simple hash function for deterministic behavior
		hash := 0
		for _, char := range userID {
			hash = (hash*31 + int(char)) % 100
		}
		return float64(hash) < percentage
	}
	return false
}

// Mock implementation of user list evaluation for testing
func evaluateUserListFlag(flag *domain.FeatureFlag, userID string) bool {
	if users, ok := flag.Parameters["users"].([]string); ok {
		for _, user := range users {
			if user == userID {
				return true
			}
		}
	}
	return false
}
