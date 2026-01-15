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

func newTestStore(t *testing.T) *boltdb.Store {
	t.Helper()
	config := storage.StoreConfig{
		Path:          filepath.Join(t.TempDir(), "test.db"),
		Retention:     24 * time.Hour,
		PruneInterval: time.Hour,
	}
	store, err := boltdb.New(config)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestStore_New(t *testing.T) {
	t.Parallel()

	t.Run("creates store successfully", func(t *testing.T) {
		t.Parallel()
		store := newTestStore(t)
		require.NotNil(t, store)
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

func TestStore_WriteAndGetSystemCPU(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()

	now := time.Now()
	cpu := metrics.SystemCPU{
		User:      1000,
		System:    500,
		Idle:      5000,
		Timestamp: now,
	}

	err := store.WriteSystemCPU(ctx, &cpu)
	require.NoError(t, err)

	// Get latest
	latest, err := store.GetLatestSystemCPU(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(1000), latest.User)
	assert.Equal(t, uint64(500), latest.System)

	// Get by range
	results, err := store.GetSystemCPU(ctx, now.Add(-time.Hour), now.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, uint64(1000), results[0].User)
}

func TestStore_WriteAndGetSystemMemory(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()

	now := time.Now()
	mem := metrics.SystemMemory{
		Total:     16 * 1024 * 1024 * 1024,
		Available: 8 * 1024 * 1024 * 1024,
		Timestamp: now,
	}

	err := store.WriteSystemMemory(ctx, &mem)
	require.NoError(t, err)

	// Get latest
	latest, err := store.GetLatestSystemMemory(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint64(16*1024*1024*1024), latest.Total)

	// Get by range
	results, err := store.GetSystemMemory(ctx, now.Add(-time.Hour), now.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, uint64(16*1024*1024*1024), results[0].Total)
}

func TestStore_WriteAndGetProcessMetrics(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
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

	err := store.WriteProcessMetrics(ctx, &proc)
	require.NoError(t, err)

	// Get latest
	latest, err := store.GetLatestProcessMetrics(ctx, "test-service")
	require.NoError(t, err)
	assert.Equal(t, "test-service", latest.ServiceName)
	assert.Equal(t, 1234, latest.PID)

	// Get by range
	results, err := store.GetProcessMetrics(ctx, "test-service", now.Add(-time.Hour), now.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "test-service", results[0].ServiceName)
}

func TestStore_GetLatestReturnsErrorWhenEmpty(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()

	_, err := store.GetLatestSystemCPU(ctx)
	assert.Error(t, err)

	_, err = store.GetLatestSystemMemory(ctx)
	assert.Error(t, err)

	_, err = store.GetLatestProcessMetrics(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestStore_GetProcessMetricsForNonexistentService(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()

	results, err := store.GetProcessMetrics(ctx, "nonexistent", time.Now().Add(-time.Hour), time.Now())
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestStore_MultipleWrites(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()

	base := time.Now()
	for i := 0; i < 10; i++ {
		cpu := metrics.SystemCPU{
			User:      uint64(i * 100),
			Timestamp: base.Add(time.Duration(i) * time.Second),
		}
		err := store.WriteSystemCPU(ctx, &cpu)
		require.NoError(t, err)
	}

	results, err := store.GetSystemCPU(ctx, base, base.Add(10*time.Second))
	require.NoError(t, err)
	assert.Len(t, results, 10)

	// Verify ordering
	for i, r := range results {
		assert.Equal(t, uint64(i*100), r.User)
	}
}

func TestStore_Prune(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()

	// Write old and recent data
	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now()

	oldCPU := metrics.SystemCPU{User: 100, Timestamp: old}
	newCPU := metrics.SystemCPU{User: 200, Timestamp: recent}

	require.NoError(t, store.WriteSystemCPU(ctx, &oldCPU))
	require.NoError(t, store.WriteSystemCPU(ctx, &newCPU))

	// Prune data older than 1 hour
	deleted, err := store.Prune(ctx, time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	// Verify only recent data remains
	results, err := store.GetSystemCPU(ctx, old.Add(-time.Hour), recent.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, uint64(200), results[0].User)
}

func TestStore_PruneProcessMetrics(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	ctx := context.Background()

	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now()

	oldProc := metrics.ProcessMetrics{ServiceName: "svc", PID: 1, Timestamp: old}
	newProc := metrics.ProcessMetrics{ServiceName: "svc", PID: 2, Timestamp: recent}

	require.NoError(t, store.WriteProcessMetrics(ctx, &oldProc))
	require.NoError(t, store.WriteProcessMetrics(ctx, &newProc))

	deleted, err := store.Prune(ctx, time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	results, err := store.GetProcessMetrics(ctx, "svc", old.Add(-time.Hour), recent.Add(time.Hour))
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, 2, results[0].PID)
}

func TestStore_ContextCancellation(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := store.GetSystemCPU(ctx, time.Now(), time.Now())
	assert.ErrorIs(t, err, context.Canceled)

	err = store.WriteSystemCPU(ctx, &metrics.SystemCPU{})
	assert.ErrorIs(t, err, context.Canceled)
}
