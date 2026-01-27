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
	stats := supervisor.NewServiceStats()

	// Increment some counters.
	stats.IncrementStart()
	stats.IncrementStart()
	stats.IncrementStop()
	stats.IncrementFail()
	stats.IncrementFail()
	stats.IncrementFail()
	stats.IncrementRestart()

	// Get snapshot.
	snap := stats.Snapshot()

	// Verify snapshot values.
	assert.Equal(t, 2, snap.StartCount)
	assert.Equal(t, 1, snap.StopCount)
	assert.Equal(t, 3, snap.FailCount)
	assert.Equal(t, 1, snap.RestartCount)
}

// TestServiceStats_concurrent_increments tests thread safety of atomic operations.
//
// Params:
//   - t: the testing context.
func TestServiceStats_concurrent_increments(t *testing.T) {
	stats := supervisor.NewServiceStats()
	const numGoroutines int = 100
	const incrementsPerGoroutine int = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 4) // 4 counters

	// Concurrent increments for each counter type.
	for range numGoroutines {
		go func() {
			defer wg.Done()
			for range incrementsPerGoroutine {
				stats.IncrementStart()
			}
		}()
		go func() {
			defer wg.Done()
			for range incrementsPerGoroutine {
				stats.IncrementStop()
			}
		}()
		go func() {
			defer wg.Done()
			for range incrementsPerGoroutine {
				stats.IncrementFail()
			}
		}()
		go func() {
			defer wg.Done()
			for range incrementsPerGoroutine {
				stats.IncrementRestart()
			}
		}()
	}

	wg.Wait()

	// Verify all increments were counted.
	expected := numGoroutines * incrementsPerGoroutine
	assert.Equal(t, expected, stats.StartCount())
	assert.Equal(t, expected, stats.StopCount())
	assert.Equal(t, expected, stats.FailCount())
	assert.Equal(t, expected, stats.RestartCount())
}
