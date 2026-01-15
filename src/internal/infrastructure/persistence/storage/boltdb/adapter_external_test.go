//go:build linux

// Package boltdb_test provides external tests for the boltdb package.
package boltdb_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/storage"
	"github.com/kodflow/daemon/internal/infrastructure/persistence/storage/boltdb"
)

func newTestAdapter(t *testing.T) *boltdb.Adapter {
	t.Helper()
	config := storage.StoreConfig{
		Path:          filepath.Join(t.TempDir(), "test.db"),
		Retention:     24 * time.Hour,
		PruneInterval: time.Hour,
	}
	adapter, err := boltdb.New(config)
	require.NoError(t, err)
	t.Cleanup(func() { _ = adapter.Close() })
	return adapter
}

func TestAdapter_New(t *testing.T) {
	t.Parallel()

	t.Run("creates adapter successfully", func(t *testing.T) {
		t.Parallel()
		adapter := newTestAdapter(t)
		require.NotNil(t, adapter)
	})

	t.Run("fails with invalid path", func(t *testing.T) {
		t.Parallel()
		config := storage.StoreConfig{
			Path: "/nonexistent/path/that/should/fail/test.db",
		}
		_, err := boltdb.New(config)
		assert.Error(t, err)
	})
}

func TestAdapter_WriteAndGetSystemCPU(t *testing.T) {
	t.Parallel()
	adapter := newTestAdapter(t)
	ctx := context.Background()

	now := time.Now()
	cpu := metrics.SystemCPU{
		User:      1000,
		System:    500,
		Idle:      5000,
		Timestamp: now,
	}

	err := adapter.WriteSystemCPU(ctx, &cpu)
	require.NoError(t, err)

	// Get latest
	latest, err := adapter.GetLatestSystemCPU(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), latest.User)
	assert.Equal(t, uint64(500), latest.System)

	// Get by range
	results, err := adapter.GetSystemCPU(ctx, now.Add(-time.Hour), now.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, uint64(1000), results[0].User)
}

func TestAdapter_WriteAndGetSystemMemory(t *testing.T) {
	t.Parallel()
	adapter := newTestAdapter(t)
	ctx := context.Background()

	now := time.Now()
	mem := metrics.SystemMemory{
		Total:     16 * 1024 * 1024 * 1024,
		Available: 8 * 1024 * 1024 * 1024,
		Timestamp: now,
	}

	err := adapter.WriteSystemMemory(ctx, &mem)
	require.NoError(t, err)

	// Get latest
	latest, err := adapter.GetLatestSystemMemory(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(16*1024*1024*1024), latest.Total)

	// Get by range
	results, err := adapter.GetSystemMemory(ctx, now.Add(-time.Hour), now.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, uint64(16*1024*1024*1024), results[0].Total)
}

func TestAdapter_WriteAndGetProcessMetrics(t *testing.T) {
	t.Parallel()
	adapter := newTestAdapter(t)
	ctx := context.Background()

	now := time.Now()
	proc := metrics.ProcessMetrics{
		ServiceName:  "test-service",
		PID:          1234,
		State:        process.StateRunning,
		Healthy:      true,
		RestartCount: 0,
		Timestamp:    now,
	}

	err := adapter.WriteProcessMetrics(ctx, &proc)
	require.NoError(t, err)

	// Get latest
	latest, err := adapter.GetLatestProcessMetrics(ctx, "test-service")
	require.NoError(t, err)
	assert.Equal(t, "test-service", latest.ServiceName)
	assert.Equal(t, 1234, latest.PID)

	// Get by range
	results, err := adapter.GetProcessMetrics(ctx, "test-service", now.Add(-time.Hour), now.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "test-service", results[0].ServiceName)
}

func TestAdapter_GetLatestReturnsErrorWhenEmpty(t *testing.T) {
	t.Parallel()
	adapter := newTestAdapter(t)
	ctx := context.Background()

	_, err := adapter.GetLatestSystemCPU(ctx)
	assert.Error(t, err)

	_, err = adapter.GetLatestSystemMemory(ctx)
	assert.Error(t, err)

	_, err = adapter.GetLatestProcessMetrics(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestAdapter_GetProcessMetricsForNonexistentService(t *testing.T) {
	t.Parallel()
	adapter := newTestAdapter(t)
	ctx := context.Background()

	results, err := adapter.GetProcessMetrics(ctx, "nonexistent", time.Now().Add(-time.Hour), time.Now())
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestAdapter_MultipleWrites(t *testing.T) {
	t.Parallel()
	adapter := newTestAdapter(t)
	ctx := context.Background()

	base := time.Now()
	for i := 0; i < 10; i++ {
		cpu := metrics.SystemCPU{
			User:      uint64(i * 100),
			Timestamp: base.Add(time.Duration(i) * time.Second),
		}
		err := adapter.WriteSystemCPU(ctx, &cpu)
		require.NoError(t, err)
	}

	results, err := adapter.GetSystemCPU(ctx, base, base.Add(10*time.Second))
	require.NoError(t, err)
	assert.Len(t, results, 10)

	// Verify ordering
	for i, r := range results {
		assert.Equal(t, uint64(i*100), r.User)
	}
}

func TestAdapter_Prune(t *testing.T) {
	t.Parallel()
	adapter := newTestAdapter(t)
	ctx := context.Background()

	// Write old and recent data
	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now()

	oldCPU := metrics.SystemCPU{User: 100, Timestamp: old}
	newCPU := metrics.SystemCPU{User: 200, Timestamp: recent}

	require.NoError(t, adapter.WriteSystemCPU(ctx, &oldCPU))
	require.NoError(t, adapter.WriteSystemCPU(ctx, &newCPU))

	// Prune data older than 1 hour
	deleted, err := adapter.Prune(ctx, time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	// Verify only recent data remains
	results, err := adapter.GetSystemCPU(ctx, old.Add(-time.Hour), recent.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, uint64(200), results[0].User)
}

func TestAdapter_PruneProcessMetrics(t *testing.T) {
	t.Parallel()
	adapter := newTestAdapter(t)
	ctx := context.Background()

	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now()

	oldProc := metrics.ProcessMetrics{ServiceName: "svc", PID: 1, Timestamp: old}
	newProc := metrics.ProcessMetrics{ServiceName: "svc", PID: 2, Timestamp: recent}

	require.NoError(t, adapter.WriteProcessMetrics(ctx, &oldProc))
	require.NoError(t, adapter.WriteProcessMetrics(ctx, &newProc))

	deleted, err := adapter.Prune(ctx, time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	results, err := adapter.GetProcessMetrics(ctx, "svc", old.Add(-time.Hour), recent.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 2, results[0].PID)
}

func TestAdapter_ContextCancellation(t *testing.T) {
	t.Parallel()
	adapter := newTestAdapter(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := adapter.GetSystemCPU(ctx, time.Now(), time.Now())
	assert.ErrorIs(t, err, context.Canceled)

	err = adapter.WriteSystemCPU(ctx, &metrics.SystemCPU{})
	assert.ErrorIs(t, err, context.Canceled)
}
