// Package metrics provides internal tests for tracker.go.
// It tests internal implementation details using white-box testing.
package metrics

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
	domain "github.com/kodflow/daemon/internal/domain/process"
)

// mockCollectorInternal implements MetricsCollector for internal testing.
type mockCollectorInternal struct {
	mu       sync.Mutex
	cpuCalls int
	memCalls int
	cpuErr   error
	memErr   error
	cpu      domainmetrics.ProcessCPU
	mem      domainmetrics.ProcessMemory
}

// CollectCPU collects CPU metrics for a process.
//
// Params:
//   - ctx: the context for the collection
//   - pid: the process ID
//
// Returns:
//   - ProcessCPU: the CPU metrics
//   - error: any error that occurred
func (m *mockCollectorInternal) CollectCPU(_ context.Context, pid int) (domainmetrics.ProcessCPU, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cpuCalls++
	if m.cpuErr != nil {
		return domainmetrics.ProcessCPU{}, m.cpuErr
	}
	cpu := m.cpu
	cpu.PID = pid
	return cpu, nil
}

// CollectMemory collects memory metrics for a process.
//
// Params:
//   - ctx: the context for the collection
//   - pid: the process ID
//
// Returns:
//   - ProcessMemory: the memory metrics
//   - error: any error that occurred
func (m *mockCollectorInternal) CollectMemory(_ context.Context, pid int) (domainmetrics.ProcessMemory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.memCalls++
	if m.memErr != nil {
		return domainmetrics.ProcessMemory{}, m.memErr
	}
	mem := m.mem
	mem.PID = pid
	return mem, nil
}

// Test_Tracker_calculateCPUPercent tests the calculateCPUPercent method.
//
// Params:
//   - t: the testing context.
func Test_Tracker_calculateCPUPercent(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prevCPU is the previous CPU snapshot.
		prevCPU domainmetrics.ProcessCPU
		// currCPU is the current CPU snapshot.
		currCPU domainmetrics.ProcessCPU
		// prevTime is the time of previous snapshot.
		prevTime time.Time
		// currTime is the time of current snapshot.
		currTime time.Time
		// expected is the expected CPU percentage.
		expected float64
	}{
		{
			name: "calculates_cpu_percent_for_1_second_interval",
			prevCPU: domainmetrics.ProcessCPU{
				User:   100, // 1 second of user time (100 jiffies at 100 Hz)
				System: 100, // 1 second of system time
			},
			currCPU: domainmetrics.ProcessCPU{
				User:   200, // 2 seconds total
				System: 200, // 2 seconds total
			},
			prevTime: time.Now(),
			currTime: time.Now().Add(1 * time.Second),
			expected: 200.0, // (200 jiffies / 100 Hz / 1 second) * 100 = 200%
		},
		{
			name: "returns_zero_for_zero_elapsed_time",
			prevCPU: domainmetrics.ProcessCPU{
				User:   100,
				System: 100,
			},
			currCPU: domainmetrics.ProcessCPU{
				User:   200,
				System: 200,
			},
			prevTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			currTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), // Same time
			expected: 0.0,
		},
		{
			name: "returns_zero_for_negative_elapsed_time",
			prevCPU: domainmetrics.ProcessCPU{
				User:   100,
				System: 100,
			},
			currCPU: domainmetrics.ProcessCPU{
				User:   200,
				System: 200,
			},
			prevTime: time.Now(),
			currTime: time.Now().Add(-1 * time.Second), // Earlier time
			expected: 0.0,
		},
		{
			name: "returns_zero_for_counter_wrap",
			prevCPU: domainmetrics.ProcessCPU{
				User:   200,
				System: 200,
			},
			currCPU: domainmetrics.ProcessCPU{
				User:   100, // Wrapped/reset
				System: 100,
			},
			prevTime: time.Now(),
			currTime: time.Now().Add(1 * time.Second),
			expected: 0.0, // Underflow detected
		},
		{
			name: "calculates_low_cpu_usage",
			prevCPU: domainmetrics.ProcessCPU{
				User:   10,
				System: 10,
			},
			currCPU: domainmetrics.ProcessCPU{
				User:   15,
				System: 15,
			},
			prevTime: time.Now(),
			currTime: time.Now().Add(1 * time.Second),
			expected: 10.0, // (10 jiffies / 100 Hz / 1 second) * 100 = 10%
		},
		{
			name: "zero_cpu_usage",
			prevCPU: domainmetrics.ProcessCPU{
				User:   100,
				System: 100,
			},
			currCPU: domainmetrics.ProcessCPU{
				User:   100, // No change
				System: 100,
			},
			prevTime: time.Now(),
			currTime: time.Now().Add(1 * time.Second),
			expected: 0.0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker(nil)

			percent := tracker.calculateCPUPercent(tt.prevCPU, tt.currCPU, tt.prevTime, tt.currTime)

			// Use delta comparison for floating point.
			assert.InDelta(t, tt.expected, percent, 0.1)
		})
	}
}

// Test_Tracker_buildMetrics tests the buildMetrics method.
//
// Params:
//   - t: the testing context.
func Test_Tracker_buildMetrics(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// proc is the tracked process.
		proc *trackedProcess
		// expectedServiceName is the expected service name.
		expectedServiceName string
		// expectedPID is the expected PID.
		expectedPID int
	}{
		{
			name: "builds_metrics_with_zero_pid",
			proc: &trackedProcess{
				serviceName:  "test-service",
				pid:          0,
				state:        domain.StateRunning,
				healthy:      true,
				startTime:    time.Now(),
				restartCount: 0,
				lastMetrics:  domainmetrics.ProcessMetrics{},
			},
			expectedServiceName: "test-service",
			expectedPID:         0,
		},
		{
			name: "builds_metrics_with_positive_pid",
			proc: &trackedProcess{
				serviceName:  "web-service",
				pid:          1234,
				state:        domain.StateRunning,
				healthy:      true,
				startTime:    time.Now().Add(-1 * time.Minute),
				restartCount: 2,
				lastMetrics: domainmetrics.ProcessMetrics{
					CPU: domainmetrics.ProcessCPU{
						UsagePercent: 50.0,
					},
					Memory: domainmetrics.ProcessMemory{
						RSS: 1024 * 1024,
					},
				},
			},
			expectedServiceName: "web-service",
			expectedPID:         1234,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker(nil)
			now := time.Now()

			metrics := tracker.buildMetrics(tt.proc, now)

			assert.Equal(t, tt.expectedServiceName, metrics.ServiceName)
			assert.Equal(t, tt.expectedPID, metrics.PID)
			assert.Equal(t, tt.proc.state, metrics.State)
			assert.Equal(t, tt.proc.healthy, metrics.Healthy)
			assert.Equal(t, now, metrics.Timestamp)
		})
	}
}

// Test_Tracker_collectLoop tests the collectLoop method.
//
// Params:

// Goroutine lifecycle:
//   - Spawns one goroutine for testing collectLoop.
//   - Goroutine exits when context is cancelled.
//   - Test blocks until goroutine exits or timeout.
//   - t: the testing context.
func Test_Tracker_collectLoop(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// interval is the collection interval.
		interval time.Duration
	}{
		{
			name:     "exits_when_context_cancelled",
			interval: 100 * time.Millisecond,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker(nil, WithCollectionInterval(tt.interval))
			ctx, cancel := context.WithCancel(context.Background())
			tracker.ctx = ctx
			tracker.cancel = cancel

			// Start collectLoop in background.
			done := make(chan struct{})
			go func() {
				tracker.collectLoop()
				close(done)
			}()

			// Cancel context to stop loop.
			cancel()

			// Wait for goroutine to exit.
			select {
			case <-done:
				// Success - goroutine exited.
			case <-time.After(1 * time.Second):
				t.Fatal("collectLoop did not exit after context cancellation")
			}
		})
	}
}

// Test_Tracker_collectAll tests the collectAll method.
//
// Params:
//   - t: the testing context.
func Test_Tracker_collectAll(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// numProcesses is the number of processes to track.
		numProcesses int
	}{
		{
			name:         "collects_zero_processes",
			numProcesses: 0,
		},
		{
			name:         "collects_single_process",
			numProcesses: 1,
		},
		{
			name:         "collects_multiple_processes",
			numProcesses: 3,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			collector := &mockCollectorInternal{}
			tracker := NewTracker(collector)
			tracker.ctx = context.Background()
			tracker.interval = 1 * time.Second

			// Add processes to tracker.
			for i := range tt.numProcesses {
				proc := &trackedProcess{
					serviceName:  fmt.Sprintf("service-%d", i),
					pid:          i + 1,
					state:        domain.StateRunning,
					healthy:      true,
					startTime:    time.Now(),
					restartCount: 0,
					lastMetrics:  domainmetrics.ProcessMetrics{},
				}
				tracker.processes[proc.serviceName] = proc
			}

			// Call collectAll - should not panic.
			tracker.collectAll()

			// Verify collector was called for each process with valid PID.
			assert.Equal(t, tt.numProcesses, collector.cpuCalls)
			assert.Equal(t, tt.numProcesses, collector.memCalls)
		})
	}
}

// Test_Tracker_collectProcess tests the collectProcess method.
//
// Params:
//   - t: the testing context.
func Test_Tracker_collectProcess(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// pid is the process PID.
		pid int
	}{
		{
			name: "handles_zero_pid",
			pid:  0,
		},
		{
			name: "handles_negative_pid",
			pid:  -1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker(nil)
			tracker.ctx = context.Background()
			tracker.interval = 1 * time.Second

			proc := &trackedProcess{
				serviceName:  "test-service",
				pid:          tt.pid,
				state:        domain.StateRunning,
				healthy:      true,
				startTime:    time.Now(),
				restartCount: 0,
				lastMetrics:  domainmetrics.ProcessMetrics{},
			}

			// Call collectProcess - should not panic.
			tracker.collectProcess(proc)
		})
	}
}

// Test_Tracker_updateProcessMetrics tests the updateProcessMetrics method.
//
// Params:
//   - t: the testing context.
func Test_Tracker_updateProcessMetrics(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// cpu is the CPU metrics.
		cpu domainmetrics.ProcessCPU
		// mem is the memory metrics.
		mem domainmetrics.ProcessMemory
	}{
		{
			name: "updates_metrics_with_zero_values",
			cpu:  domainmetrics.ProcessCPU{},
			mem:  domainmetrics.ProcessMemory{},
		},
		{
			name: "updates_metrics_with_positive_values",
			cpu: domainmetrics.ProcessCPU{
				UsagePercent: 75.5,
				User:         100,
				System:       50,
			},
			mem: domainmetrics.ProcessMemory{
				RSS:  2 * 1024 * 1024,
				VMS:  4 * 1024 * 1024,
				Swap: 1024 * 1024,
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTracker(nil)

			proc := &trackedProcess{
				serviceName:  "test-service",
				pid:          1234,
				state:        domain.StateRunning,
				healthy:      true,
				startTime:    time.Now().Add(-1 * time.Minute),
				restartCount: 0,
				lastMetrics:  domainmetrics.ProcessMetrics{},
			}

			tracker.updateProcessMetrics(proc, tt.cpu, tt.mem)

			assert.Equal(t, tt.cpu, proc.lastMetrics.CPU)
			assert.Equal(t, tt.mem, proc.lastMetrics.Memory)
		})
	}
}
