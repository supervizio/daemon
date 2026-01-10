// Package supervisor_test provides external tests for service_stats.go.
// It tests the public API of the ServiceStats type using black-box testing.
package supervisor_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/application/supervisor"
)

// TestNewServiceStats tests the NewServiceStats constructor function.
//
// Params:
//   - t: the testing context.
func TestNewServiceStats(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "creates_stats_with_zero_values",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			// Verify all fields are initialized to zero.
			assert.NotNil(t, stats)
			assert.Equal(t, 0, stats.StartCount)
			assert.Equal(t, 0, stats.StopCount)
			assert.Equal(t, 0, stats.FailCount)
			assert.Equal(t, 0, stats.RestartCount)
		})
	}
}

// TestServiceStats_fields tests that ServiceStats fields can be modified.
//
// Params:
//   - t: the testing context.
func TestServiceStats_fields(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startCount is the value to set for StartCount.
		startCount int
		// stopCount is the value to set for StopCount.
		stopCount int
		// failCount is the value to set for FailCount.
		failCount int
		// restartCount is the value to set for RestartCount.
		restartCount int
	}{
		{
			name:         "fields_can_be_set_and_read",
			startCount:   5,
			stopCount:    3,
			failCount:    2,
			restartCount: 4,
		},
		{
			name:         "fields_can_be_zero",
			startCount:   0,
			stopCount:    0,
			failCount:    0,
			restartCount: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			// Set the fields.
			stats.StartCount = tt.startCount
			stats.StopCount = tt.stopCount
			stats.FailCount = tt.failCount
			stats.RestartCount = tt.restartCount

			// Verify the fields were set correctly.
			assert.Equal(t, tt.startCount, stats.StartCount)
			assert.Equal(t, tt.stopCount, stats.StopCount)
			assert.Equal(t, tt.failCount, stats.FailCount)
			assert.Equal(t, tt.restartCount, stats.RestartCount)
		})
	}
}
