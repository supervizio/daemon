package monitoring

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

// TestHealthSummary_NewHealthSummary_internal tests internal state initialization.
func TestHealthSummary_NewHealthSummary_internal(t *testing.T) {
	tests := []struct {
		name       string
		verifyFunc func(*testing.T, HealthSummary)
	}{
		{
			name: "initializes_maps_correctly",
			verifyFunc: func(t *testing.T, h HealthSummary) {
				assert.NotNil(t, h.ByType)
				assert.NotNil(t, h.ByState)
			},
		},
		{
			name: "initializes_with_zero_total",
			verifyFunc: func(t *testing.T, h HealthSummary) {
				assert.Equal(t, 0, h.Total)
			},
		},
		{
			name: "initializes_empty_type_map",
			verifyFunc: func(t *testing.T, h HealthSummary) {
				assert.Equal(t, 0, len(h.ByType))
				assert.NotNil(t, h.ByType)
			},
		},
		{
			name: "initializes_empty_state_map",
			verifyFunc: func(t *testing.T, h HealthSummary) {
				assert.Equal(t, 0, len(h.ByState))
				assert.NotNil(t, h.ByState)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := NewHealthSummary()
			tt.verifyFunc(t, summary)
		})
	}
}

// TestHealthSummary_HealthyCount_internal tests HealthyCount method with internal state.
func TestHealthSummary_HealthyCount_internal(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() HealthSummary
		expected int
	}{
		{
			name: "returns_zero_when_state_not_present",
			setup: func() HealthSummary {
				h := NewHealthSummary()
				h.ByState[target.StateUnhealthy] = 5
				return h
			},
			expected: 0,
		},
		{
			name: "returns_correct_count_from_state_map",
			setup: func() HealthSummary {
				h := NewHealthSummary()
				h.ByState[target.StateHealthy] = 7
				return h
			},
			expected: 7,
		},
		{
			name: "ignores_other_states",
			setup: func() HealthSummary {
				h := NewHealthSummary()
				h.ByState[target.StateHealthy] = 3
				h.ByState[target.StateUnhealthy] = 2
				h.ByState[target.StateUnknown] = 1
				return h
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			result := h.HealthyCount()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHealthSummary_UnhealthyCount_internal tests UnhealthyCount method with internal state.
func TestHealthSummary_UnhealthyCount_internal(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() HealthSummary
		expected int
	}{
		{
			name: "returns_zero_when_state_not_present",
			setup: func() HealthSummary {
				h := NewHealthSummary()
				h.ByState[target.StateHealthy] = 5
				return h
			},
			expected: 0,
		},
		{
			name: "returns_correct_count_from_state_map",
			setup: func() HealthSummary {
				h := NewHealthSummary()
				h.ByState[target.StateUnhealthy] = 9
				return h
			},
			expected: 9,
		},
		{
			name: "ignores_other_states",
			setup: func() HealthSummary {
				h := NewHealthSummary()
				h.ByState[target.StateHealthy] = 3
				h.ByState[target.StateUnhealthy] = 4
				h.ByState[target.StateUnknown] = 1
				return h
			},
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			result := h.UnhealthyCount()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHealthSummary_UnknownCount_internal tests UnknownCount method with internal state.
func TestHealthSummary_UnknownCount_internal(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() HealthSummary
		expected int
	}{
		{
			name: "returns_zero_when_state_not_present",
			setup: func() HealthSummary {
				h := NewHealthSummary()
				h.ByState[target.StateHealthy] = 5
				return h
			},
			expected: 0,
		},
		{
			name: "returns_correct_count_from_state_map",
			setup: func() HealthSummary {
				h := NewHealthSummary()
				h.ByState[target.StateUnknown] = 2
				return h
			},
			expected: 2,
		},
		{
			name: "ignores_other_states",
			setup: func() HealthSummary {
				h := NewHealthSummary()
				h.ByState[target.StateHealthy] = 3
				h.ByState[target.StateUnhealthy] = 4
				h.ByState[target.StateUnknown] = 5
				return h
			},
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.setup()
			result := h.UnknownCount()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestHealthSummary_mapAccessPattern tests internal map access patterns.
func TestHealthSummary_mapAccessPattern(t *testing.T) {
	tests := []struct {
		name       string
		verifyFunc func(*testing.T)
	}{
		{
			name: "state_map_handles_missing_keys",
			verifyFunc: func(t *testing.T) {
				h := NewHealthSummary()
				// Accessing missing key should return zero value (0).
				assert.Equal(t, 0, h.ByState[target.StateHealthy])
				assert.Equal(t, 0, h.ByState[target.StateUnhealthy])
				assert.Equal(t, 0, h.ByState[target.StateUnknown])
			},
		},
		{
			name: "type_map_handles_missing_keys",
			verifyFunc: func(t *testing.T) {
				h := NewHealthSummary()
				// Accessing missing key should return zero value (0).
				assert.Equal(t, 0, h.ByType[target.TypeSystemd])
				assert.Equal(t, 0, h.ByType[target.TypeDocker])
				assert.Equal(t, 0, h.ByType[target.TypeKubernetes])
			},
		},
		{
			name: "maps_are_independent",
			verifyFunc: func(t *testing.T) {
				h := NewHealthSummary()
				h.ByState[target.StateHealthy] = 10
				h.ByType[target.TypeSystemd] = 20

				// Modifying one map should not affect the other.
				assert.Equal(t, 10, h.ByState[target.StateHealthy])
				assert.Equal(t, 20, h.ByType[target.TypeSystemd])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.verifyFunc(t)
		})
	}
}

// TestHealthSummary_multipleStates tests behavior with multiple state counts.
func TestHealthSummary_multipleStates(t *testing.T) {
	tests := []struct {
		name           string
		healthyCount   int
		unhealthyCount int
		unknownCount   int
	}{
		{
			name:           "all_states_populated",
			healthyCount:   5,
			unhealthyCount: 3,
			unknownCount:   2,
		},
		{
			name:           "only_healthy",
			healthyCount:   10,
			unhealthyCount: 0,
			unknownCount:   0,
		},
		{
			name:           "only_unhealthy",
			healthyCount:   0,
			unhealthyCount: 7,
			unknownCount:   0,
		},
		{
			name:           "mixed_counts",
			healthyCount:   1,
			unhealthyCount: 2,
			unknownCount:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHealthSummary()
			h.ByState[target.StateHealthy] = tt.healthyCount
			h.ByState[target.StateUnhealthy] = tt.unhealthyCount
			h.ByState[target.StateUnknown] = tt.unknownCount

			// Verify each count method returns correct value.
			assert.Equal(t, tt.healthyCount, h.HealthyCount())
			assert.Equal(t, tt.unhealthyCount, h.UnhealthyCount())
			assert.Equal(t, tt.unknownCount, h.UnknownCount())
		})
	}
}
