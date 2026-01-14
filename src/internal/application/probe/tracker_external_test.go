// Package metrics_test provides external tests for the metrics package.
package probe_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appprobe "github.com/kodflow/daemon/internal/application/probe"
	"github.com/kodflow/daemon/internal/domain/probe"
)

// mockCollector implements MetricsCollector for testing.
type mockCollector struct {
	mu       sync.Mutex
	cpuCalls int
	memCalls int
	cpuErr   error
	memErr   error
	cpu      probe.ProcessCPU
	mem      probe.ProcessMemory
}

func (m *mockCollector) CollectCPU(_ context.Context, pid int) (probe.ProcessCPU, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cpuCalls++
	if m.cpuErr != nil {
		return probe.ProcessCPU{}, m.cpuErr
	}
	cpu := m.cpu
	cpu.PID = pid
	return cpu, nil
}

func (m *mockCollector) CollectMemory(_ context.Context, pid int) (probe.ProcessMemory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.memCalls++
	if m.memErr != nil {
		return probe.ProcessMemory{}, m.memErr
	}
	mem := m.mem
	mem.PID = pid
	return mem, nil
}

func TestTracker_Track(t *testing.T) {
	t.Parallel()

	t.Run("tracks new service", func(t *testing.T) {
		t.Parallel()

		collector := &mockCollector{}
		tracker := appprobe.NewTracker(collector)

		err := tracker.Track(context.Background(), "test-service", 1234)
		require.NoError(t, err)

		m, ok := tracker.Get("test-service")
		require.True(t, ok)
		assert.Equal(t, "test-service", m.ServiceName)
		assert.Equal(t, 0, m.RestartCount)
	})

	t.Run("increments restart count on same service new PID", func(t *testing.T) {
		t.Parallel()

		collector := &mockCollector{}
		tracker := appprobe.NewTracker(collector)

		err := tracker.Track(context.Background(), "test-service", 1234)
		require.NoError(t, err)

		err = tracker.Track(context.Background(), "test-service", 5678)
		require.NoError(t, err)

		m, ok := tracker.Get("test-service")
		require.True(t, ok)
		assert.Equal(t, 1, m.RestartCount)
	})
}

func TestTracker_Untrack(t *testing.T) {
	t.Parallel()

	collector := &mockCollector{}
	tracker := appprobe.NewTracker(collector)

	err := tracker.Track(context.Background(), "test-service", 1234)
	require.NoError(t, err)

	_, ok := tracker.Get("test-service")
	require.True(t, ok)

	tracker.Untrack("test-service")

	_, ok = tracker.Get("test-service")
	assert.False(t, ok)
}

func TestTracker_GetAll(t *testing.T) {
	t.Parallel()

	collector := &mockCollector{}
	tracker := appprobe.NewTracker(collector)

	err := tracker.Track(context.Background(), "service-1", 1001)
	require.NoError(t, err)
	err = tracker.Track(context.Background(), "service-2", 1002)
	require.NoError(t, err)

	all := tracker.GetAll()
	assert.Len(t, all, 2)
}

func TestTracker_Subscribe(t *testing.T) {
	t.Parallel()

	collector := &mockCollector{
		cpu: probe.ProcessCPU{User: 100, System: 50},
		mem: probe.ProcessMemory{RSS: 1024 * 1024},
	}
	tracker := appprobe.NewTracker(collector, appprobe.WithCollectionInterval(50*time.Millisecond))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := tracker.Start(ctx)
	require.NoError(t, err)

	sub := tracker.Subscribe()
	defer tracker.Unsubscribe(sub)

	err = tracker.Track(ctx, "test-service", 1234)
	require.NoError(t, err)

	// Wait for at least one collection
	select {
	case m := <-sub:
		assert.Equal(t, "test-service", m.ServiceName)
		assert.Equal(t, 1234, m.PID)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for metrics")
	}

	tracker.Stop()
}

func TestTracker_UpdateState(t *testing.T) {
	t.Parallel()

	collector := &mockCollector{}
	tracker := appprobe.NewTracker(collector)

	err := tracker.Track(context.Background(), "test-service", 1234)
	require.NoError(t, err)

	tracker.UpdateState("test-service", probe.ProcessStateFailed, "exit code 1")

	m, ok := tracker.Get("test-service")
	require.True(t, ok)
	assert.Equal(t, probe.ProcessStateFailed, m.State)
	assert.Equal(t, "exit code 1", m.LastError)
	assert.Equal(t, 0, m.PID) // PID should be cleared on failure
}

func TestTracker_UpdateHealth(t *testing.T) {
	t.Parallel()

	collector := &mockCollector{}
	tracker := appprobe.NewTracker(collector)

	err := tracker.Track(context.Background(), "test-service", 1234)
	require.NoError(t, err)

	// Default should be healthy
	m, _ := tracker.Get("test-service")
	assert.True(t, m.Healthy)

	tracker.UpdateHealth("test-service", false)

	m, _ = tracker.Get("test-service")
	assert.False(t, m.Healthy)
}
