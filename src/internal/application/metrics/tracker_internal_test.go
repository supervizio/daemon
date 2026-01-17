package metrics

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
)

// === Test Constants ===

// internalTestInterval is the interval used in internal tests.
const internalTestInterval time.Duration = 50 * time.Millisecond

// internalTestPID is a sample PID for internal testing.
const internalTestPID int = 1234

// === Mock Types ===

// internalMockCollector implements Collector for internal testing.
type internalMockCollector struct {
	mu       sync.Mutex
	cpuCalls int
	memCalls int
	cpuErr   error
	memErr   error
	cpu      domainmetrics.ProcessCPU
	mem      domainmetrics.ProcessMemory
}

// CollectCPU collects CPU metrics for testing.
//
// Params:
//   - ctx: context for cancellation
//   - pid: process ID to collect for
//
// Returns:
//   - ProcessCPU: collected CPU metrics
//   - error: collection error if any
func (m *internalMockCollector) CollectCPU(_ context.Context, pid int) (domainmetrics.ProcessCPU, error) {
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

// CollectMemory collects memory metrics for testing.
//
// Params:
//   - ctx: context for cancellation
//   - pid: process ID to collect for
//
// Returns:
//   - ProcessMemory: collected memory metrics
//   - error: collection error if any
func (m *internalMockCollector) CollectMemory(_ context.Context, pid int) (domainmetrics.ProcessMemory, error) {
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

// Test internal buildMetrics function.
func TestTracker_buildMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		proc     *trackedProcess
		now      time.Time
		expected domainmetrics.ProcessMetrics
	}{
		{
			name: "running process with uptime",
			proc: &trackedProcess{
				serviceName:  "test-service",
				pid:          123,
				state:        process.StateRunning,
				healthy:      true,
				startTime:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				restartCount: 2,
				lastError:    "",
				lastMetrics: domainmetrics.ProcessMetrics{
					CPU:    domainmetrics.ProcessCPU{User: 100, System: 50},
					Memory: domainmetrics.ProcessMemory{RSS: 1024},
				},
			},
			now: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
			expected: domainmetrics.ProcessMetrics{
				ServiceName:  "test-service",
				PID:          123,
				State:        process.StateRunning,
				Healthy:      true,
				CPU:          domainmetrics.ProcessCPU{User: 100, System: 50},
				Memory:       domainmetrics.ProcessMemory{RSS: 1024},
				StartTime:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Uptime:       time.Hour,
				RestartCount: 2,
				LastError:    "",
				Timestamp:    time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "stopped process without uptime",
			proc: &trackedProcess{
				serviceName:  "stopped-service",
				pid:          0,
				state:        process.StateStopped,
				healthy:      false,
				startTime:    time.Time{},
				restartCount: 0,
				lastError:    "exit code 1",
				lastMetrics:  domainmetrics.ProcessMetrics{},
			},
			now: time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
			expected: domainmetrics.ProcessMetrics{
				ServiceName:  "stopped-service",
				PID:          0,
				State:        process.StateStopped,
				Healthy:      false,
				CPU:          domainmetrics.ProcessCPU{},
				Memory:       domainmetrics.ProcessMemory{},
				StartTime:    time.Time{},
				Uptime:       0,
				RestartCount: 0,
				LastError:    "exit code 1",
				Timestamp:    time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tracker := &Tracker{}
			result := tracker.buildMetrics(tt.proc, tt.now)

			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTracker_collectLoop tests the collectLoop method.
func TestTracker_collectLoop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupProcs  int
		waitCycles  int
		expectCalls bool
	}{
		{
			name:        "runs collection cycle",
			setupProcs:  1,
			waitCycles:  2,
			expectCalls: true,
		},
		{
			name:        "stops on context cancellation",
			setupProcs:  1,
			waitCycles:  1,
			expectCalls: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &internalMockCollector{
				cpu: domainmetrics.ProcessCPU{User: 100, System: 50},
				mem: domainmetrics.ProcessMemory{RSS: 1024},
			}
			tracker := NewTracker(collector, WithCollectionInterval(internalTestInterval))

			ctx, cancel := context.WithCancel(t.Context())

			err := tracker.Start(ctx)
			require.NoError(t, err)

			for i := range tt.setupProcs {
				err := tracker.Track("service-"+string(rune('a'+i)), internalTestPID+i)
				require.NoError(t, err)
			}

			// Wait for collection cycles
			time.Sleep(internalTestInterval * time.Duration(tt.waitCycles+1))

			cancel()
			tracker.Stop()

			collector.mu.Lock()
			cpuCalls := collector.cpuCalls
			collector.mu.Unlock()

			if tt.expectCalls {
				assert.Greater(t, cpuCalls, 0, "expected collection calls")
			}
		})
	}
}

// TestTracker_collectAll tests the collectAll method.
func TestTracker_collectAll(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		processCount int
		expectCalls  int
	}{
		{
			name:         "collects for single process",
			processCount: 1,
			expectCalls:  1,
		},
		{
			name:         "collects for multiple processes",
			processCount: 3,
			expectCalls:  3,
		},
		{
			name:         "no-op with no processes",
			processCount: 0,
			expectCalls:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &internalMockCollector{
				cpu: domainmetrics.ProcessCPU{User: 100},
				mem: domainmetrics.ProcessMemory{RSS: 1024},
			}
			tracker := NewTracker(collector, WithCollectionInterval(internalTestInterval))
			tracker.ctx, tracker.cancel = context.WithCancel(t.Context())

			for i := range tt.processCount {
				err := tracker.Track("service-"+string(rune('a'+i)), internalTestPID+i)
				require.NoError(t, err)
			}

			tracker.collectAll()

			collector.mu.Lock()
			cpuCalls := collector.cpuCalls
			memCalls := collector.memCalls
			collector.mu.Unlock()

			assert.Equal(t, tt.expectCalls, cpuCalls, "CPU collection calls")
			assert.Equal(t, tt.expectCalls, memCalls, "Memory collection calls")
		})
	}
}

// TestTracker_collectProcess tests the collectProcess method.
func TestTracker_collectProcess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		pid         int
		cpuErr      error
		memErr      error
		expectState process.State
		expectCalls bool
	}{
		{
			name:        "collects successfully",
			pid:         internalTestPID,
			cpuErr:      nil,
			memErr:      nil,
			expectState: process.StateRunning,
			expectCalls: true,
		},
		{
			name:        "skips collection for zero PID",
			pid:         0,
			cpuErr:      nil,
			memErr:      nil,
			expectState: process.StateRunning,
			expectCalls: false,
		},
		{
			name:        "marks failed when both errors",
			pid:         internalTestPID,
			cpuErr:      errors.New("cpu error"),
			memErr:      errors.New("mem error"),
			expectState: process.StateFailed,
			expectCalls: true,
		},
		{
			name:        "continues when only CPU fails",
			pid:         internalTestPID,
			cpuErr:      errors.New("cpu error"),
			memErr:      nil,
			expectState: process.StateRunning,
			expectCalls: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &internalMockCollector{
				cpuErr: tt.cpuErr,
				memErr: tt.memErr,
				cpu:    domainmetrics.ProcessCPU{User: 100},
				mem:    domainmetrics.ProcessMemory{RSS: 1024},
			}
			tracker := NewTracker(collector, WithCollectionInterval(internalTestInterval))
			tracker.ctx, tracker.cancel = context.WithCancel(t.Context())

			proc := &trackedProcess{
				serviceName: "test-service",
				pid:         tt.pid,
				state:       process.StateRunning,
				healthy:     true,
				startTime:   time.Now(),
			}
			tracker.processes["test-service"] = proc

			tracker.collectProcess(proc)

			collector.mu.Lock()
			cpuCalls := collector.cpuCalls
			collector.mu.Unlock()

			if tt.expectCalls {
				assert.Greater(t, cpuCalls, 0, "expected collection calls")
			} else {
				assert.Equal(t, 0, cpuCalls, "expected no collection calls")
			}

			m, ok := tracker.Get("test-service")
			require.True(t, ok)
			assert.Equal(t, tt.expectState, m.State)
		})
	}
}

// TestTracker_updateProcessMetrics tests the updateProcessMetrics method.
func TestTracker_updateProcessMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		cpu           domainmetrics.ProcessCPU
		mem           domainmetrics.ProcessMemory
		expectCPU     domainmetrics.ProcessCPU
		expectMem     domainmetrics.ProcessMemory
		subscriberCnt int
	}{
		{
			name:          "updates metrics without subscribers",
			cpu:           domainmetrics.ProcessCPU{User: 100, System: 50},
			mem:           domainmetrics.ProcessMemory{RSS: 1024},
			expectCPU:     domainmetrics.ProcessCPU{User: 100, System: 50},
			expectMem:     domainmetrics.ProcessMemory{RSS: 1024},
			subscriberCnt: 0,
		},
		{
			name:          "updates and publishes with subscriber",
			cpu:           domainmetrics.ProcessCPU{User: 200, System: 100},
			mem:           domainmetrics.ProcessMemory{RSS: 2048},
			expectCPU:     domainmetrics.ProcessCPU{User: 200, System: 100},
			expectMem:     domainmetrics.ProcessMemory{RSS: 2048},
			subscriberCnt: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &internalMockCollector{}
			tracker := NewTracker(collector)

			proc := &trackedProcess{
				serviceName: "test-service",
				pid:         internalTestPID,
				state:       process.StateRunning,
				healthy:     true,
				startTime:   time.Now(),
			}
			tracker.processes["test-service"] = proc

			var subscribers []<-chan domainmetrics.ProcessMetrics
			for range tt.subscriberCnt {
				subscribers = append(subscribers, tracker.Subscribe())
			}

			tracker.updateProcessMetrics(proc, tt.cpu, tt.mem)

			m, ok := tracker.Get("test-service")
			require.True(t, ok)
			assert.Equal(t, tt.expectCPU, m.CPU)
			assert.Equal(t, tt.expectMem, m.Memory)

			// Cleanup subscribers
			for _, sub := range subscribers {
				tracker.Unsubscribe(sub)
			}
		})
	}
}

// TestTracker_Unsubscribe_nil tests Unsubscribe with nil channel.
// This verifies the defensive handling when a nil channel is passed.
//
// Note: With the reflection-based implementation, nil channels are handled
// safely by reflect.ValueOf().Pointer() which returns 0 for nil channels,
// so the channel is simply not found in the subscribers map and nothing happens.
func TestTracker_Unsubscribe_nil(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "handles_nil_channel_without_panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &internalMockCollector{}
			tracker := NewTracker(collector)

			// Pass nil - should not panic or cause errors.
			// Note: nil channels are handled safely by reflection,
			// as reflect.ValueOf(nil).Pointer() returns 0 and won't match any channel.
			assert.NotPanics(t, func() {
				tracker.Unsubscribe(nil)
			})
		})
	}
}

// TestTracker_Unsubscribe_verifyChannelClosed tests that Unsubscribe
// properly closes the channel and removes it from subscribers.
func TestTracker_Unsubscribe_verifyChannelClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "channel_is_closed_and_removed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &internalMockCollector{}
			tracker := NewTracker(collector)

			// Subscribe to get a channel
			ch := tracker.Subscribe()

			// Verify channel is open by trying non-blocking receive
			select {
			case _, ok := <-ch:
				if !ok {
					t.Fatal("channel should be open after subscribe")
				}
			default:
				// Channel is open and empty (expected)
			}

			// Unsubscribe
			tracker.Unsubscribe(ch)

			// Verify channel is closed
			_, ok := <-ch
			assert.False(t, ok, "channel should be closed after unsubscribe")

			// Verify channel is removed from subscribers by publishing
			// If channel was not removed, this would panic trying to send to closed channel
			metrics := &domainmetrics.ProcessMetrics{
				ServiceName: "test",
				PID:         1234,
			}
			assert.NotPanics(t, func() {
				tracker.publish(metrics)
			}, "publishing should not panic after unsubscribe")
		})
	}
}

// TestTracker_publish tests the publish method.
func TestTracker_publish(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		subscriberCnt int
		expectMsgs    int
	}{
		{
			name:          "no subscribers",
			subscriberCnt: 0,
			expectMsgs:    0,
		},
		{
			name:          "single subscriber",
			subscriberCnt: 1,
			expectMsgs:    1,
		},
		{
			name:          "multiple subscribers",
			subscriberCnt: 3,
			expectMsgs:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &internalMockCollector{}
			tracker := NewTracker(collector)

			var subscribers []<-chan domainmetrics.ProcessMetrics
			for range tt.subscriberCnt {
				subscribers = append(subscribers, tracker.Subscribe())
			}

			metrics := &domainmetrics.ProcessMetrics{
				ServiceName: "test-service",
				PID:         internalTestPID,
				State:       process.StateRunning,
				Timestamp:   time.Now(),
			}

			tracker.publish(metrics)

			received := 0
			for _, sub := range subscribers {
				select {
				case m := <-sub:
					assert.Equal(t, "test-service", m.ServiceName)
					received++
				case <-time.After(100 * time.Millisecond):
					// Timeout
				}
			}

			assert.Equal(t, tt.expectMsgs, received)

			// Cleanup
			for _, sub := range subscribers {
				tracker.Unsubscribe(sub)
			}
		})
	}
}
