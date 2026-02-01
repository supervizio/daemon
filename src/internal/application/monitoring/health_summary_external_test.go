package monitoring_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/application/monitoring"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

func TestHealthSummary(t *testing.T) {
	// testCase defines a test case for HealthSummary operations.
	type testCase struct {
		name      string
		setupFunc func() monitoring.HealthSummary
		testFunc  func(*testing.T, monitoring.HealthSummary)
	}

	// tests defines all test cases for HealthSummary.
	tests := []testCase{
		{
			name: "NewHealthSummary initializes correctly",
			setupFunc: func() monitoring.HealthSummary {
				// create new health summary
				return monitoring.NewHealthSummary()
			},
			testFunc: func(t *testing.T, summary monitoring.HealthSummary) {
				// verify initial state
				assert.Equal(t, 0, summary.Total)
				assert.NotNil(t, summary.ByType)
				assert.NotNil(t, summary.ByState)
			},
		},
		{
			name: "HealthyCount returns correct count",
			setupFunc: func() monitoring.HealthSummary {
				// create summary with healthy targets
				summary := monitoring.NewHealthSummary()
				summary.ByState[target.StateHealthy] = 5
				return summary
			},
			testFunc: func(t *testing.T, summary monitoring.HealthSummary) {
				// verify healthy count
				assert.Equal(t, 5, summary.HealthyCount())
			},
		},
		{
			name: "UnhealthyCount returns correct count",
			setupFunc: func() monitoring.HealthSummary {
				// create summary with unhealthy targets
				summary := monitoring.NewHealthSummary()
				summary.ByState[target.StateUnhealthy] = 3
				return summary
			},
			testFunc: func(t *testing.T, summary monitoring.HealthSummary) {
				// verify unhealthy count
				assert.Equal(t, 3, summary.UnhealthyCount())
			},
		},
		{
			name: "UnknownCount returns correct count",
			setupFunc: func() monitoring.HealthSummary {
				// create summary with unknown targets
				summary := monitoring.NewHealthSummary()
				summary.ByState[target.StateUnknown] = 2
				return summary
			},
			testFunc: func(t *testing.T, summary monitoring.HealthSummary) {
				// verify unknown count
				assert.Equal(t, 2, summary.UnknownCount())
			},
		},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary := tc.setupFunc()
			tc.testFunc(t, summary)
		})
	}
}

// TestNewHealthSummary tests the NewHealthSummary constructor.
func TestNewHealthSummary(t *testing.T) {
	// testCase defines a test case for NewHealthSummary.
	type testCase struct {
		name   string
		verify func(*testing.T, monitoring.HealthSummary)
	}

	tests := []testCase{
		{
			name: "creates summary with initialized maps",
			verify: func(t *testing.T, summary monitoring.HealthSummary) {
				assert.Equal(t, 0, summary.Total)
				assert.NotNil(t, summary.ByType)
				assert.NotNil(t, summary.ByState)
			},
		},
		{
			name: "initializes with zero counts",
			verify: func(t *testing.T, summary monitoring.HealthSummary) {
				assert.Equal(t, 0, summary.HealthyCount())
				assert.Equal(t, 0, summary.UnhealthyCount())
				assert.Equal(t, 0, summary.UnknownCount())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary := monitoring.NewHealthSummary()
			tc.verify(t, summary)
		})
	}
}

// TestHealthyCount tests the HealthyCount method.
func TestHealthyCount(t *testing.T) {
	// testCase defines a test case for HealthyCount.
	type testCase struct {
		name     string
		setup    func() monitoring.HealthSummary
		expected int
	}

	tests := []testCase{
		{
			name: "returns zero for empty summary",
			setup: func() monitoring.HealthSummary {
				return monitoring.NewHealthSummary()
			},
			expected: 0,
		},
		{
			name: "returns correct count when healthy targets exist",
			setup: func() monitoring.HealthSummary {
				summary := monitoring.NewHealthSummary()
				summary.ByState[target.StateHealthy] = 5
				return summary
			},
			expected: 5,
		},
		{
			name: "ignores other states",
			setup: func() monitoring.HealthSummary {
				summary := monitoring.NewHealthSummary()
				summary.ByState[target.StateHealthy] = 3
				summary.ByState[target.StateUnhealthy] = 7
				summary.ByState[target.StateUnknown] = 2
				return summary
			},
			expected: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary := tc.setup()
			assert.Equal(t, tc.expected, summary.HealthyCount())
		})
	}
}

// TestUnhealthyCount tests the UnhealthyCount method.
func TestUnhealthyCount(t *testing.T) {
	// testCase defines a test case for UnhealthyCount.
	type testCase struct {
		name     string
		setup    func() monitoring.HealthSummary
		expected int
	}

	tests := []testCase{
		{
			name: "returns zero for empty summary",
			setup: func() monitoring.HealthSummary {
				return monitoring.NewHealthSummary()
			},
			expected: 0,
		},
		{
			name: "returns correct count when unhealthy targets exist",
			setup: func() monitoring.HealthSummary {
				summary := monitoring.NewHealthSummary()
				summary.ByState[target.StateUnhealthy] = 4
				return summary
			},
			expected: 4,
		},
		{
			name: "ignores other states",
			setup: func() monitoring.HealthSummary {
				summary := monitoring.NewHealthSummary()
				summary.ByState[target.StateHealthy] = 5
				summary.ByState[target.StateUnhealthy] = 8
				summary.ByState[target.StateUnknown] = 1
				return summary
			},
			expected: 8,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary := tc.setup()
			assert.Equal(t, tc.expected, summary.UnhealthyCount())
		})
	}
}

// TestUnknownCount tests the UnknownCount method.
func TestUnknownCount(t *testing.T) {
	// testCase defines a test case for UnknownCount.
	type testCase struct {
		name     string
		setup    func() monitoring.HealthSummary
		expected int
	}

	tests := []testCase{
		{
			name: "returns zero for empty summary",
			setup: func() monitoring.HealthSummary {
				return monitoring.NewHealthSummary()
			},
			expected: 0,
		},
		{
			name: "returns correct count when unknown targets exist",
			setup: func() monitoring.HealthSummary {
				summary := monitoring.NewHealthSummary()
				summary.ByState[target.StateUnknown] = 6
				return summary
			},
			expected: 6,
		},
		{
			name: "ignores other states",
			setup: func() monitoring.HealthSummary {
				summary := monitoring.NewHealthSummary()
				summary.ByState[target.StateHealthy] = 2
				summary.ByState[target.StateUnhealthy] = 3
				summary.ByState[target.StateUnknown] = 9
				return summary
			},
			expected: 9,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary := tc.setup()
			assert.Equal(t, tc.expected, summary.UnknownCount())
		})
	}
}
