// Package supervisor_test provides external tests for service_stats.go.
// It tests the public API of the ServiceStats type using black-box testing.
package supervisor_test

import (
	"sync"
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

			// Verify all fields are initialized to zero via getter methods.
			assert.NotNil(t, stats)
			assert.Equal(t, 0, stats.StartCount())
			assert.Equal(t, 0, stats.StopCount())
			assert.Equal(t, 0, stats.FailCount())
			assert.Equal(t, 0, stats.RestartCount())
		})
	}
}

// TestServiceStats_atomic_increments tests the atomic increment methods.
//
// Params:
//   - t: the testing context.
func TestServiceStats_atomic_increments(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startCount is the number of times to call IncrementStart.
		startCount int
		// stopCount is the number of times to call IncrementStop.
		stopCount int
		// failCount is the number of times to call IncrementFail.
		failCount int
		// restartCount is the number of times to call IncrementRestart.
		restartCount int
	}{
		{
			name:         "single_increments",
			startCount:   1,
			stopCount:    1,
			failCount:    1,
			restartCount: 1,
		},
		{
			name:         "multiple_increments",
			startCount:   5,
			stopCount:    3,
			failCount:    2,
			restartCount: 4,
		},
		{
			name:         "zero_increments",
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

			// Call increment methods the specified number of times.
			for range tt.startCount {
				stats.IncrementStart()
			}
			for range tt.stopCount {
				stats.IncrementStop()
			}
			for range tt.failCount {
				stats.IncrementFail()
			}
			for range tt.restartCount {
				stats.IncrementRestart()
			}

			// Verify the counts via getter methods.
			assert.Equal(t, tt.startCount, stats.StartCount())
			assert.Equal(t, tt.stopCount, stats.StopCount())
			assert.Equal(t, tt.failCount, stats.FailCount())
			assert.Equal(t, tt.restartCount, stats.RestartCount())
		})
	}
}

// TestServiceStats_Snapshot tests the Snapshot method returns consistent values.
//
// Params:
//   - t: the testing context.
func TestServiceStats_Snapshot(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startCount is the number of starts.
		startCount int
		// stopCount is the number of stops.
		stopCount int
		// failCount is the number of fails.
		failCount int
		// restartCount is the number of restarts.
		restartCount int
	}{
		{
			name:         "snapshot_with_mixed_counts",
			startCount:   2,
			stopCount:    1,
			failCount:    3,
			restartCount: 1,
		},
		{
			name:         "snapshot_with_zero_counts",
			startCount:   0,
			stopCount:    0,
			failCount:    0,
			restartCount: 0,
		},
		{
			name:         "snapshot_with_high_counts",
			startCount:   10,
			stopCount:    5,
			failCount:    7,
			restartCount: 3,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			// Increment counters.
			for range tt.startCount {
				stats.IncrementStart()
			}
			for range tt.stopCount {
				stats.IncrementStop()
			}
			for range tt.failCount {
				stats.IncrementFail()
			}
			for range tt.restartCount {
				stats.IncrementRestart()
			}

			// Get snapshot.
			snap := stats.Snapshot()

			// Verify snapshot values.
			assert.Equal(t, tt.startCount, snap.StartCount)
			assert.Equal(t, tt.stopCount, snap.StopCount)
			assert.Equal(t, tt.failCount, snap.FailCount)
			assert.Equal(t, tt.restartCount, snap.RestartCount)
		})
	}
}

// TestServiceStats_concurrent_increments tests thread safety of atomic operations.
//
// Params:
//   - t: the testing context.
func TestServiceStats_concurrent_increments(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// numGoroutines is the number of concurrent goroutines.
		numGoroutines int
		// incrementsPerGoroutine is the increments per goroutine.
		incrementsPerGoroutine int
	}{
		{
			name:                   "concurrent_with_100_goroutines",
			numGoroutines:          100,
			incrementsPerGoroutine: 100,
		},
		{
			name:                   "concurrent_with_10_goroutines",
			numGoroutines:          10,
			incrementsPerGoroutine: 50,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			var wg sync.WaitGroup
			wg.Add(tt.numGoroutines * 4) // 4 counters

			// Concurrent increments for each counter type.
			for range tt.numGoroutines {
				go func() {
					defer wg.Done()
					for range tt.incrementsPerGoroutine {
						stats.IncrementStart()
					}
				}()
				go func() {
					defer wg.Done()
					for range tt.incrementsPerGoroutine {
						stats.IncrementStop()
					}
				}()
				go func() {
					defer wg.Done()
					for range tt.incrementsPerGoroutine {
						stats.IncrementFail()
					}
				}()
				go func() {
					defer wg.Done()
					for range tt.incrementsPerGoroutine {
						stats.IncrementRestart()
					}
				}()
			}

			wg.Wait()

			// Verify all increments were counted.
			expected := tt.numGoroutines * tt.incrementsPerGoroutine
			assert.Equal(t, expected, stats.StartCount())
			assert.Equal(t, expected, stats.StopCount())
			assert.Equal(t, expected, stats.FailCount())
			assert.Equal(t, expected, stats.RestartCount())
		})
	}
}

// TestServiceStats_IncrementStart tests the IncrementStart method.
//
// Params:
//   - t: the testing context.
func TestServiceStats_IncrementStart(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// count is the number of increments.
		count int
		// expected is the expected start count.
		expected int
	}{
		{
			name:     "increment_once",
			count:    1,
			expected: 1,
		},
		{
			name:     "increment_multiple_times",
			count:    5,
			expected: 5,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			for range tt.count {
				stats.IncrementStart()
			}

			assert.Equal(t, tt.expected, stats.StartCount())
		})
	}
}

// TestServiceStats_IncrementStop tests the IncrementStop method.
//
// Params:
//   - t: the testing context.
func TestServiceStats_IncrementStop(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// count is the number of increments.
		count int
		// expected is the expected stop count.
		expected int
	}{
		{
			name:     "increment_once",
			count:    1,
			expected: 1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			for range tt.count {
				stats.IncrementStop()
			}

			assert.Equal(t, tt.expected, stats.StopCount())
		})
	}
}

// TestServiceStats_IncrementFail tests the IncrementFail method.
//
// Params:
//   - t: the testing context.
func TestServiceStats_IncrementFail(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// count is the number of increments.
		count int
		// expected is the expected fail count.
		expected int
	}{
		{
			name:     "increment_once",
			count:    1,
			expected: 1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			for range tt.count {
				stats.IncrementFail()
			}

			assert.Equal(t, tt.expected, stats.FailCount())
		})
	}
}

// TestServiceStats_IncrementRestart tests the IncrementRestart method.
//
// Params:
//   - t: the testing context.
func TestServiceStats_IncrementRestart(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// count is the number of increments.
		count int
		// expected is the expected restart count.
		expected int
	}{
		{
			name:     "increment_once",
			count:    1,
			expected: 1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			for range tt.count {
				stats.IncrementRestart()
			}

			assert.Equal(t, tt.expected, stats.RestartCount())
		})
	}
}

// TestServiceStats_StartCount tests the StartCount method.
//
// Params:
//   - t: the testing context.
func TestServiceStats_StartCount(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// increments is the number of increments.
		increments int
		// expected is the expected count.
		expected int
	}{
		{
			name:       "zero_count",
			increments: 0,
			expected:   0,
		},
		{
			name:       "positive_count",
			increments: 3,
			expected:   3,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			for range tt.increments {
				stats.IncrementStart()
			}

			assert.Equal(t, tt.expected, stats.StartCount())
		})
	}
}

// TestServiceStats_StopCount tests the StopCount method.
//
// Params:
//   - t: the testing context.
func TestServiceStats_StopCount(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// increments is the number of increments.
		increments int
		// expected is the expected count.
		expected int
	}{
		{
			name:       "zero_count",
			increments: 0,
			expected:   0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			for range tt.increments {
				stats.IncrementStop()
			}

			assert.Equal(t, tt.expected, stats.StopCount())
		})
	}
}

// TestServiceStats_FailCount tests the FailCount method.
//
// Params:
//   - t: the testing context.
func TestServiceStats_FailCount(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// increments is the number of increments.
		increments int
		// expected is the expected count.
		expected int
	}{
		{
			name:       "zero_count",
			increments: 0,
			expected:   0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			for range tt.increments {
				stats.IncrementFail()
			}

			assert.Equal(t, tt.expected, stats.FailCount())
		})
	}
}

// TestServiceStats_RestartCount tests the RestartCount method.
//
// Params:
//   - t: the testing context.
func TestServiceStats_RestartCount(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// increments is the number of increments.
		increments int
		// expected is the expected count.
		expected int
	}{
		{
			name:       "zero_count",
			increments: 0,
			expected:   0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			for range tt.increments {
				stats.IncrementRestart()
			}

			assert.Equal(t, tt.expected, stats.RestartCount())
		})
	}
}

// TestServiceStats_SnapshotPtr tests the SnapshotPtr method.
//
// Params:
//   - t: the testing context.
func TestServiceStats_SnapshotPtr(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startCount is the number of starts.
		startCount int
		// stopCount is the number of stops.
		stopCount int
		// failCount is the number of fails.
		failCount int
		// restartCount is the number of restarts.
		restartCount int
	}{
		{
			name:         "snapshot_with_zero_counts",
			startCount:   0,
			stopCount:    0,
			failCount:    0,
			restartCount: 0,
		},
		{
			name:         "snapshot_with_mixed_counts",
			startCount:   2,
			stopCount:    1,
			failCount:    1,
			restartCount: 1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			stats := supervisor.NewServiceStats()

			for range tt.startCount {
				stats.IncrementStart()
			}
			for range tt.stopCount {
				stats.IncrementStop()
			}
			for range tt.failCount {
				stats.IncrementFail()
			}
			for range tt.restartCount {
				stats.IncrementRestart()
			}

			snap := stats.SnapshotPtr()

			assert.NotNil(t, snap)
			assert.Equal(t, tt.startCount, snap.StartCount)
			assert.Equal(t, tt.stopCount, snap.StopCount)
			assert.Equal(t, tt.failCount, snap.FailCount)
			assert.Equal(t, tt.restartCount, snap.RestartCount)
		})
	}
}
