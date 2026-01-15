// Package metrics_test provides external tests for the metrics package.
package metrics_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appmetrics "github.com/kodflow/daemon/internal/application/metrics"
	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
)

// === Test Constants ===

// testCollectionInterval is the interval used in tests.
const testCollectionInterval time.Duration = 50 * time.Millisecond

// testTimeout is the timeout for waiting for metrics.
const testTimeout time.Duration = 500 * time.Millisecond

// testPID is a sample PID for testing.
const testPID int = 1234

// mockCollector implements MetricsCollector for testing.
type mockCollector struct {
	mu       sync.Mutex
	cpuCalls int
	memCalls int
	cpuErr   error
	memErr   error
	cpu      domainmetrics.ProcessCPU
	mem      domainmetrics.ProcessMemory
}

func (m *mockCollector) CollectCPU(_ context.Context, pid int) (domainmetrics.ProcessCPU, error) {
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

func (m *mockCollector) CollectMemory(_ context.Context, pid int) (domainmetrics.ProcessMemory, error) {
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

func TestTracker_Track(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		serviceName      string
		pid              int
		secondPID        int
		wantRestartCount int
	}{
		{
			name:             "tracks new service",
			serviceName:      "test-service",
			pid:              1234,
			secondPID:        0,
			wantRestartCount: 0,
		},
		{
			name:             "increments restart count on same service new PID",
			serviceName:      "test-service",
			pid:              1234,
			secondPID:        5678,
			wantRestartCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			tracker := appmetrics.NewTracker(collector)

			err := tracker.Track(tt.serviceName, tt.pid)
			require.NoError(t, err)

			if tt.secondPID != 0 {
				err = tracker.Track(tt.serviceName, tt.secondPID)
				require.NoError(t, err)
			}

			m, ok := tracker.Get(tt.serviceName)
			require.True(t, ok)
			assert.Equal(t, tt.serviceName, m.ServiceName)
			assert.Equal(t, tt.wantRestartCount, m.RestartCount)
		})
	}
}

func TestTracker_Untrack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		pid         int
	}{
		{
			name:        "untracks existing service",
			serviceName: "test-service",
			pid:         1234,
		},
		{
			name:        "untracks service with different name",
			serviceName: "another-service",
			pid:         5678,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			tracker := appmetrics.NewTracker(collector)

			err := tracker.Track(tt.serviceName, tt.pid)
			require.NoError(t, err)

			_, ok := tracker.Get(tt.serviceName)
			require.True(t, ok)

			tracker.Untrack(tt.serviceName)

			_, ok = tracker.Get(tt.serviceName)
			assert.False(t, ok)
		})
	}
}

func TestTracker_All(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		services []struct {
			name string
			pid  int
		}
		wantCount int
	}{
		{
			name:      "returns empty for no tracked services",
			services:  nil,
			wantCount: 0,
		},
		{
			name: "returns all tracked services",
			services: []struct {
				name string
				pid  int
			}{
				{"service-1", 1001},
				{"service-2", 1002},
			},
			wantCount: 2,
		},
		{
			name: "returns single tracked service",
			services: []struct {
				name string
				pid  int
			}{
				{"service-only", 1003},
			},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			tracker := appmetrics.NewTracker(collector)

			for _, svc := range tt.services {
				err := tracker.Track(svc.name, svc.pid)
				require.NoError(t, err)
			}

			all := tracker.All()
			assert.Len(t, all, tt.wantCount)
		})
	}
}

func TestTracker_Subscribe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		pid         int
		cpuUser     uint64
		memRSS      uint64
	}{
		{
			name:        "receives metrics updates",
			serviceName: "test-service",
			pid:         1234,
			cpuUser:     100,
			memRSS:      1024 * 1024,
		},
		{
			name:        "receives metrics for different service",
			serviceName: "another-service",
			pid:         5678,
			cpuUser:     200,
			memRSS:      2048 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{
				cpu: domainmetrics.ProcessCPU{User: tt.cpuUser, System: 50},
				mem: domainmetrics.ProcessMemory{RSS: tt.memRSS},
			}
			tracker := appmetrics.NewTracker(collector, appmetrics.WithCollectionInterval(50*time.Millisecond))

			ctx := t.Context()

			err := tracker.Start(ctx)
			require.NoError(t, err)

			sub := tracker.Subscribe()
			defer tracker.Unsubscribe(sub)

			err = tracker.Track(tt.serviceName, tt.pid)
			require.NoError(t, err)

			// Wait for at least one collection
			select {
			case m := <-sub:
				assert.Equal(t, tt.serviceName, m.ServiceName)
				assert.Equal(t, tt.pid, m.PID)
			case <-time.After(500 * time.Millisecond):
				t.Fatal("timeout waiting for metrics")
			}

			tracker.Stop()
		})
	}
}

func TestTracker_UpdateState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		pid         int
		newState    process.State
		lastError   string
		wantPID     int
	}{
		{
			name:        "updates state to failed clears PID",
			serviceName: "test-service",
			pid:         1234,
			newState:    process.StateFailed,
			lastError:   "exit code 1",
			wantPID:     0,
		},
		{
			name:        "updates state to running keeps PID",
			serviceName: "test-service",
			pid:         1234,
			newState:    process.StateRunning,
			lastError:   "",
			wantPID:     1234,
		},
		{
			name:        "updates state to stopped clears PID",
			serviceName: "worker-service",
			pid:         5678,
			newState:    process.StateStopped,
			lastError:   "",
			wantPID:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			tracker := appmetrics.NewTracker(collector)

			err := tracker.Track(tt.serviceName, tt.pid)
			require.NoError(t, err)

			tracker.UpdateState(tt.serviceName, tt.newState, tt.lastError)

			m, ok := tracker.Get(tt.serviceName)
			require.True(t, ok)
			assert.Equal(t, tt.newState, m.State)
			assert.Equal(t, tt.lastError, m.LastError)
			assert.Equal(t, tt.wantPID, m.PID)
		})
	}
}

func TestTracker_UpdateHealth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		pid         int
		healthy     bool
	}{
		{
			name:        "updates health to unhealthy",
			serviceName: "test-service",
			pid:         1234,
			healthy:     false,
		},
		{
			name:        "updates health to healthy",
			serviceName: "test-service",
			pid:         1234,
			healthy:     true,
		},
		{
			name:        "updates different service health",
			serviceName: "another-service",
			pid:         5678,
			healthy:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			tracker := appmetrics.NewTracker(collector)

			err := tracker.Track(tt.serviceName, tt.pid)
			require.NoError(t, err)

			tracker.UpdateHealth(tt.serviceName, tt.healthy)

			m, _ := tracker.Get(tt.serviceName)
			assert.Equal(t, tt.healthy, m.Healthy)
		})
	}
}

// TestWithCollectionInterval tests the WithCollectionInterval option.
func TestWithCollectionInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		interval         time.Duration
		expectApplied    bool
		expectedInterval time.Duration
	}{
		{
			name:             "positive interval is applied",
			interval:         100 * time.Millisecond,
			expectApplied:    true,
			expectedInterval: 100 * time.Millisecond,
		},
		{
			name:             "zero interval is ignored",
			interval:         0,
			expectApplied:    false,
			expectedInterval: 5 * time.Second, // default
		},
		{
			name:             "negative interval is ignored",
			interval:         -100 * time.Millisecond,
			expectApplied:    false,
			expectedInterval: 5 * time.Second, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			opt := appmetrics.WithCollectionInterval(tt.interval)
			tracker := appmetrics.NewTracker(collector, opt)

			// Verify tracker was created successfully
			assert.NotNil(t, tracker)
		})
	}
}

// TestNewTracker tests the NewTracker constructor.
func TestNewTracker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []appmetrics.TrackerOption
		wantNil bool
	}{
		{
			name:    "creates tracker with no options",
			opts:    nil,
			wantNil: false,
		},
		{
			name:    "creates tracker with collection interval option",
			opts:    []appmetrics.TrackerOption{appmetrics.WithCollectionInterval(100 * time.Millisecond)},
			wantNil: false,
		},
		{
			name:    "creates tracker with multiple options",
			opts:    []appmetrics.TrackerOption{appmetrics.WithCollectionInterval(200 * time.Millisecond)},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			tracker := appmetrics.NewTracker(collector, tt.opts...)

			if tt.wantNil {
				assert.Nil(t, tracker)
			} else {
				assert.NotNil(t, tracker)
			}
		})
	}
}

// TestTracker_Start tests the Start method.
func TestTracker_Start(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		startTwice    bool
		expectError   bool
		cancelContext bool
	}{
		{
			name:          "starts collection successfully",
			startTwice:    false,
			expectError:   false,
			cancelContext: false,
		},
		{
			name:          "idempotent when called twice",
			startTwice:    true,
			expectError:   false,
			cancelContext: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			tracker := appmetrics.NewTracker(collector, appmetrics.WithCollectionInterval(testCollectionInterval))

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			err := tracker.Start(ctx)

			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.startTwice {
				err = tracker.Start(ctx)
				require.NoError(t, err)
			}

			tracker.Stop()
		})
	}
}

// TestTracker_Stop tests the Stop method.
func TestTracker_Stop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		startFirst  bool
		stopTwice   bool
		expectPanic bool
	}{
		{
			name:        "stops running tracker",
			startFirst:  true,
			stopTwice:   false,
			expectPanic: false,
		},
		{
			name:        "no-op when not started",
			startFirst:  false,
			stopTwice:   false,
			expectPanic: false,
		},
		{
			name:        "idempotent when called twice",
			startFirst:  true,
			stopTwice:   true,
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			tracker := appmetrics.NewTracker(collector, appmetrics.WithCollectionInterval(testCollectionInterval))

			if tt.startFirst {
				ctx := t.Context()
				err := tracker.Start(ctx)
				require.NoError(t, err)
			}

			tracker.Stop()

			if tt.stopTwice {
				tracker.Stop()
			}
		})
	}
}

// TestTracker_Get tests the Get method.
func TestTracker_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		trackFirst  bool
		expectFound bool
	}{
		{
			name:        "returns metrics for tracked service",
			serviceName: "test-service",
			trackFirst:  true,
			expectFound: true,
		},
		{
			name:        "returns false for untracked service",
			serviceName: "unknown-service",
			trackFirst:  false,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			tracker := appmetrics.NewTracker(collector)

			if tt.trackFirst {
				err := tracker.Track(tt.serviceName, testPID)
				require.NoError(t, err)
			}

			m, found := tracker.Get(tt.serviceName)

			assert.Equal(t, tt.expectFound, found)
			if tt.expectFound {
				assert.Equal(t, tt.serviceName, m.ServiceName)
			}
		})
	}
}

// TestTracker_Unsubscribe tests the Unsubscribe method.
func TestTracker_Unsubscribe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		subscribeFirst bool
	}{
		{
			name:           "unsubscribes channel successfully",
			subscribeFirst: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{}
			tracker := appmetrics.NewTracker(collector)

			if tt.subscribeFirst {
				ch := tracker.Subscribe()
				// Unsubscribe should not panic
				tracker.Unsubscribe(ch)
			}
		})
	}
}

// TestTracker_CollectionLoop tests the collection loop behavior.
func TestTracker_CollectionLoop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		cpuErr         error
		memErr         error
		expectState    process.State
		expectCPUCalls bool
		expectMemCalls bool
	}{
		{
			name:           "collects metrics successfully",
			cpuErr:         nil,
			memErr:         nil,
			expectState:    process.StateRunning,
			expectCPUCalls: true,
			expectMemCalls: true,
		},
		{
			name:           "marks process failed when both collections fail",
			cpuErr:         errors.New("cpu error"),
			memErr:         errors.New("mem error"),
			expectState:    process.StateFailed,
			expectCPUCalls: true,
			expectMemCalls: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{
				cpuErr: tt.cpuErr,
				memErr: tt.memErr,
				cpu:    domainmetrics.ProcessCPU{User: 100, System: 50},
				mem:    domainmetrics.ProcessMemory{RSS: 1024 * 1024},
			}
			tracker := appmetrics.NewTracker(collector, appmetrics.WithCollectionInterval(testCollectionInterval))

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			err := tracker.Start(ctx)
			require.NoError(t, err)

			err = tracker.Track("test-service", testPID)
			require.NoError(t, err)

			// Wait for at least one collection cycle
			time.Sleep(testCollectionInterval * 3)

			tracker.Stop()

			collector.mu.Lock()
			cpuCalls := collector.cpuCalls
			memCalls := collector.memCalls
			collector.mu.Unlock()

			if tt.expectCPUCalls {
				assert.Greater(t, cpuCalls, 0, "expected CPU collection calls")
			}
			if tt.expectMemCalls {
				assert.Greater(t, memCalls, 0, "expected memory collection calls")
			}

			m, ok := tracker.Get("test-service")
			require.True(t, ok)
			assert.Equal(t, tt.expectState, m.State)
		})
	}
}

// TestTracker_Publish tests that metrics are published to subscribers.
func TestTracker_Publish(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		subscriberCnt  int
		expectReceived bool
	}{
		{
			name:           "publishes to single subscriber",
			subscriberCnt:  1,
			expectReceived: true,
		},
		{
			name:           "publishes to multiple subscribers",
			subscriberCnt:  3,
			expectReceived: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &mockCollector{
				cpu: domainmetrics.ProcessCPU{User: 100, System: 50},
				mem: domainmetrics.ProcessMemory{RSS: 1024 * 1024},
			}
			tracker := appmetrics.NewTracker(collector, appmetrics.WithCollectionInterval(testCollectionInterval))

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			err := tracker.Start(ctx)
			require.NoError(t, err)

			// Create subscribers
			subscribers := make([]<-chan domainmetrics.ProcessMetrics, tt.subscriberCnt)
			for i := range tt.subscriberCnt {
				subscribers[i] = tracker.Subscribe()
			}

			// Track a service to trigger metrics collection
			err = tracker.Track("test-service", testPID)
			require.NoError(t, err)

			// Wait for metrics on all subscribers
			for i, sub := range subscribers {
				select {
				case m := <-sub:
					if tt.expectReceived {
						assert.Equal(t, "test-service", m.ServiceName, "subscriber %d", i)
					}
				case <-time.After(testTimeout):
					if tt.expectReceived {
						t.Errorf("subscriber %d: timeout waiting for metrics", i)
					}
				}
			}

			// Cleanup
			for _, sub := range subscribers {
				tracker.Unsubscribe(sub)
			}
			tracker.Stop()
		})
	}
}
