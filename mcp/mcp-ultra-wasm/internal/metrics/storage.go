package metrics

import (
	"context"
	"sort"
	"sync"
	"time"
)

// MemoryMetricStorage provides in-memory metric storage
type MemoryMetricStorage struct {
	mu     sync.RWMutex
	values map[string][]MetricValue
}

// NewMemoryMetricStorage creates a new in-memory metric storage
func NewMemoryMetricStorage() *MemoryMetricStorage {
	return &MemoryMetricStorage{
		values: make(map[string][]MetricValue),
	}
}

// Store stores metric values
func (mms *MemoryMetricStorage) Store(_ context.Context, values []MetricValue) error {
	if len(values) == 0 {
		return nil
	}

	mms.mu.Lock()
	defer mms.mu.Unlock()

	for _, value := range values {
		mms.values[value.Name] = append(mms.values[value.Name], value)
	}

	return nil
}

// Query queries metric values
func (mms *MemoryMetricStorage) Query(_ context.Context, query MetricQuery) ([]MetricValue, error) {
	mms.mu.RLock()
	defer mms.mu.RUnlock()

	values, exists := mms.values[query.MetricName]
	if !exists {
		return nil, nil
	}

	filtered := make([]MetricValue, 0)
	for _, value := range values {
		if mms.matchesQuery(value, query) {
			filtered = append(filtered, value)
		}
	}

	// Sort by timestamp
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.Before(filtered[j].Timestamp)
	})

	// Apply limit
	if query.Limit > 0 && len(filtered) > query.Limit {
		filtered = filtered[len(filtered)-query.Limit:]
	}

	return filtered, nil
}

// Aggregate performs aggregations on metric values
func (mms *MemoryMetricStorage) Aggregate(ctx context.Context, query AggregationQuery) ([]AggregatedMetric, error) {
	// Get base values
	values, err := mms.Query(ctx, query.MetricQuery)
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, nil
	}

	// Group values by period and labels
	groups := mms.groupValues(values, query.GroupBy, query.Period)

	// Calculate aggregations for each group
	result := make([]AggregatedMetric, 0)

	for groupKey, groupValues := range groups {
		for _, aggType := range query.Aggregations {
			aggValue := mms.calculateAggregation(groupValues, aggType)

			// Create aggregated metric
			aggMetric := AggregatedMetric{
				MetricValue: MetricValue{
					Name:      query.MetricName,
					Value:     aggValue,
					Labels:    mms.extractLabels(groupKey, query.GroupBy),
					Timestamp: time.Now(),
					Unit:      groupValues[0].Unit,
				},
				Aggregation: aggType,
				Period:      query.Period,
				Count:       int64(len(groupValues)),
			}

			result = append(result, aggMetric)
		}
	}

	return result, nil
}

// Delete removes old metric values
func (mms *MemoryMetricStorage) Delete(_ context.Context, before time.Time) error {
	mms.mu.Lock()
	defer mms.mu.Unlock()

	for metricName, values := range mms.values {
		filtered := make([]MetricValue, 0)
		for _, value := range values {
			if value.Timestamp.After(before) {
				filtered = append(filtered, value)
			}
		}
		mms.values[metricName] = filtered
	}

	return nil
}

// Close closes the storage (no-op for memory storage)
func (mms *MemoryMetricStorage) Close() error {
	return nil
}

// Private helper methods

func (mms *MemoryMetricStorage) matchesQuery(value MetricValue, query MetricQuery) bool {
	// Check time range
	if !query.StartTime.IsZero() && value.Timestamp.Before(query.StartTime) {
		return false
	}
	if !query.EndTime.IsZero() && value.Timestamp.After(query.EndTime) {
		return false
	}

	// Check labels
	for k, v := range query.Labels {
		if value.Labels[k] != v {
			return false
		}
	}

	return true
}

func (mms *MemoryMetricStorage) groupValues(values []MetricValue, groupBy []string, period time.Duration) map[string][]MetricValue {
	groups := make(map[string][]MetricValue)

	for _, value := range values {
		groupKey := mms.buildGroupKey(value, groupBy, period)
		groups[groupKey] = append(groups[groupKey], value)
	}

	return groups
}

func (mms *MemoryMetricStorage) buildGroupKey(value MetricValue, groupBy []string, period time.Duration) string {
	key := ""

	// Add time bucket if period is specified
	if period > 0 {
		bucket := value.Timestamp.Truncate(period)
		key += bucket.Format(time.RFC3339) + "|"
	}

	// Add grouped labels
	for _, label := range groupBy {
		if labelValue, exists := value.Labels[label]; exists {
			key += label + "=" + labelValue + "|"
		}
	}

	return key
}

func (mms *MemoryMetricStorage) extractLabels(_ string, _ []string) map[string]string {
	labels := make(map[string]string)

	// This is a simplified implementation
	// In a real implementation, you'd parse the group key to extract labels

	return labels
}

func (mms *MemoryMetricStorage) calculateAggregation(values []MetricValue, aggType AggregationType) float64 {
	if len(values) == 0 {
		return 0
	}

	switch aggType {
	case AggregationSum:
		sum := 0.0
		for _, value := range values {
			sum += value.Value
		}
		return sum

	case AggregationAvg:
		sum := 0.0
		for _, value := range values {
			sum += value.Value
		}
		return sum / float64(len(values))

	case AggregationMax:
		max := values[0].Value
		for _, value := range values[1:] {
			if value.Value > max {
				max = value.Value
			}
		}
		return max

	case AggregationMin:
		min := values[0].Value
		for _, value := range values[1:] {
			if value.Value < min {
				min = value.Value
			}
		}
		return min

	case AggregationCount:
		return float64(len(values))

	case AggregationP95:
		return mms.calculatePercentile(values, 95)

	case AggregationP99:
		return mms.calculatePercentile(values, 99)

	default:
		return 0
	}
}

func (mms *MemoryMetricStorage) calculatePercentile(values []MetricValue, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Extract and sort values
	nums := make([]float64, len(values))
	for i, value := range values {
		nums[i] = value.Value
	}

	sort.Float64s(nums)

	// Calculate percentile index
	index := percentile / 100.0 * float64(len(nums)-1)

	if index == float64(int(index)) {
		return nums[int(index)]
	}

	// Interpolate between values
	lower := int(index)
	upper := lower + 1
	weight := index - float64(lower)

	if upper >= len(nums) {
		return nums[lower]
	}

	return nums[lower]*(1-weight) + nums[upper]*weight
}
