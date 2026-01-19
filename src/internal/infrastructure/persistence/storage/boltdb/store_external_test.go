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

// Test iteration count for multiple writes test.
const testIterationCount int = 10

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

func newTestStore(t *testing.T) *boltdb.Store {
	t.Helper()
	config := storage.StoreConfig{
		Path:          filepath.Join(t.TempDir(), "test.db"),
		Retention:     24 * time.Hour,
		PruneInterval: time.Hour,
	}
	store, err := boltdb.NewStore(config)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	return store
}

// =============================================================================
// NEWSTORE AND INITSCHEMA TESTS
// =============================================================================

// TestNewStore tests the NewStore constructor and initSchema.
func TestNewStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupPath func(t *testing.T) string
		wantErr   bool
	}{
		{
			name: "creates store successfully",
			setupPath: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "test.db")
			},
			wantErr: false,
		},
		{
			name: "fails with invalid path",
			setupPath: func(_ *testing.T) string {
				return "/nonexistent/path/that/should/fail/test.db"
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := storage.StoreConfig{
				Path: tc.setupPath(t),
			}

			store, err := boltdb.NewStore(config)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, store)
			} else {
				require.NoError(t, err)
				require.NotNil(t, store)
				// Close also tests the Close function
				assert.NoError(t, store.Close())
			}
		})
	}
}

// =============================================================================
// WRITE TESTS
// =============================================================================

// TestStore_WriteSystemCPU tests writing system CPU metrics.
func TestStore_WriteSystemCPU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		cpuMetrics  metrics.SystemCPU
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "writes successfully",
			setupCtx: context.Background,
			cpuMetrics: metrics.SystemCPU{
				User:      1000,
				System:    500,
				Idle:      5000,
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "fails with cancelled context",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			cpuMetrics:  metrics.SystemCPU{},
			wantErr:     true,
			expectedErr: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			ctx := tc.setupCtx()

			err := store.WriteSystemCPU(ctx, &tc.cpuMetrics)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStore_WriteSystemMemory tests writing system memory metrics.
func TestStore_WriteSystemMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		memMetrics  metrics.SystemMemory
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "writes successfully",
			setupCtx: context.Background,
			memMetrics: metrics.SystemMemory{
				Total:     16 * 1024 * 1024 * 1024,
				Available: 8 * 1024 * 1024 * 1024,
				Timestamp: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "fails with cancelled context",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			memMetrics:  metrics.SystemMemory{},
			wantErr:     true,
			expectedErr: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			ctx := tc.setupCtx()

			err := store.WriteSystemMemory(ctx, &tc.memMetrics)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStore_WriteProcessMetrics tests writing process metrics.
func TestStore_WriteProcessMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		procMetrics metrics.ProcessMetrics
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "writes successfully",
			setupCtx: context.Background,
			procMetrics: metrics.ProcessMetrics{
				ServiceName:  "test-service",
				PID:          1234,
				State:        process.StateRunning,
				Healthy:      true,
				RestartCount: 0,
				Timestamp:    time.Now(),
			},
			wantErr: false,
		},
		{
			name: "fails with cancelled context",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			procMetrics: metrics.ProcessMetrics{ServiceName: "test"},
			wantErr:     true,
			expectedErr: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			ctx := tc.setupCtx()

			err := store.WriteProcessMetrics(ctx, &tc.procMetrics)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.expectedErr != nil {
					assert.ErrorIs(t, err, tc.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// GET RANGE TESTS
// =============================================================================

// TestStore_GetSystemCPU tests retrieving system CPU metrics by time range.
func TestStore_GetSystemCPU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupCtx  func() context.Context
		setupData func(t *testing.T, store *boltdb.Store, base time.Time)
		wantCount int
		wantErr   bool
		wantErrIs error
	}{
		{
			name:     "retrieves metrics in range",
			setupCtx: context.Background,
			setupData: func(t *testing.T, store *boltdb.Store, base time.Time) {
				cpu := metrics.SystemCPU{User: 1000, System: 500, Timestamp: base}
				require.NoError(t, store.WriteSystemCPU(context.Background(), &cpu))
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "returns empty for no data",
			setupCtx:  context.Background,
			setupData: func(_ *testing.T, _ *boltdb.Store, _ time.Time) {},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "fails with cancelled context",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupData: func(_ *testing.T, _ *boltdb.Store, _ time.Time) {},
			wantCount: 0,
			wantErr:   true,
			wantErrIs: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			base := time.Now()
			tc.setupData(t, store, base)
			ctx := tc.setupCtx()

			results, err := store.GetSystemCPU(ctx, base.Add(-time.Hour), base.Add(time.Hour))

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, results, tc.wantCount)
			}
		})
	}
}

// TestStore_GetSystemMemory tests retrieving system memory metrics by time range.
func TestStore_GetSystemMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupCtx  func() context.Context
		setupData func(t *testing.T, store *boltdb.Store, base time.Time)
		wantCount int
		wantErr   bool
		wantErrIs error
	}{
		{
			name:     "retrieves metrics in range",
			setupCtx: context.Background,
			setupData: func(t *testing.T, store *boltdb.Store, base time.Time) {
				mem := metrics.SystemMemory{Total: 16 * 1024 * 1024 * 1024, Timestamp: base}
				require.NoError(t, store.WriteSystemMemory(context.Background(), &mem))
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "returns empty for no data",
			setupCtx:  context.Background,
			setupData: func(_ *testing.T, _ *boltdb.Store, _ time.Time) {},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "fails with cancelled context",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupData: func(_ *testing.T, _ *boltdb.Store, _ time.Time) {},
			wantCount: 0,
			wantErr:   true,
			wantErrIs: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			base := time.Now()
			tc.setupData(t, store, base)
			ctx := tc.setupCtx()

			results, err := store.GetSystemMemory(ctx, base.Add(-time.Hour), base.Add(time.Hour))

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, results, tc.wantCount)
			}
		})
	}
}

// TestStore_GetProcessMetrics tests retrieving process metrics by time range.
func TestStore_GetProcessMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		setupCtx    func() context.Context
		setupData   func(t *testing.T, store *boltdb.Store, base time.Time)
		wantCount   int
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:        "retrieves metrics in range",
			serviceName: "test-service",
			setupCtx:    context.Background,
			setupData: func(t *testing.T, store *boltdb.Store, base time.Time) {
				proc := metrics.ProcessMetrics{
					ServiceName: "test-service",
					PID:         1234,
					Timestamp:   base,
				}
				require.NoError(t, store.WriteProcessMetrics(context.Background(), &proc))
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:        "returns empty for nonexistent service",
			serviceName: "nonexistent",
			setupCtx:    context.Background,
			setupData:   func(_ *testing.T, _ *boltdb.Store, _ time.Time) {},
			wantCount:   0,
			wantErr:     false,
		},
		{
			name:        "fails with cancelled context",
			serviceName: "test-service",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupData: func(_ *testing.T, _ *boltdb.Store, _ time.Time) {},
			wantCount: 0,
			wantErr:   true,
			wantErrIs: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			base := time.Now()
			tc.setupData(t, store, base)
			ctx := tc.setupCtx()

			results, err := store.GetProcessMetrics(ctx, tc.serviceName, base.Add(-time.Hour), base.Add(time.Hour))

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, results, tc.wantCount)
			}
		})
	}
}

// =============================================================================
// GET LATEST TESTS
// =============================================================================

// TestStore_GetLatestSystemCPU tests retrieving the most recent system CPU metrics.
func TestStore_GetLatestSystemCPU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupCtx  func() context.Context
		setupData func(t *testing.T, store *boltdb.Store)
		wantUser  uint64
		wantErr   bool
		wantErrIs error
	}{
		{
			name:     "retrieves latest successfully",
			setupCtx: context.Background,
			setupData: func(t *testing.T, store *boltdb.Store) {
				cpu := metrics.SystemCPU{User: 1000, System: 500, Timestamp: time.Now()}
				require.NoError(t, store.WriteSystemCPU(context.Background(), &cpu))
			},
			wantUser: 1000,
			wantErr:  false,
		},
		{
			name:      "returns error when empty",
			setupCtx:  context.Background,
			setupData: func(_ *testing.T, _ *boltdb.Store) {},
			wantErr:   true,
		},
		{
			name: "fails with cancelled context",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupData: func(_ *testing.T, _ *boltdb.Store) {},
			wantErr:   true,
			wantErrIs: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			tc.setupData(t, store)
			ctx := tc.setupCtx()

			result, err := store.GetLatestSystemCPU(ctx)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantUser, result.User)
			}
		})
	}
}

// TestStore_GetLatestSystemMemory tests retrieving the most recent system memory metrics.
func TestStore_GetLatestSystemMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupCtx  func() context.Context
		setupData func(t *testing.T, store *boltdb.Store)
		wantTotal uint64
		wantErr   bool
		wantErrIs error
	}{
		{
			name:     "retrieves latest successfully",
			setupCtx: context.Background,
			setupData: func(t *testing.T, store *boltdb.Store) {
				mem := metrics.SystemMemory{Total: 16 * 1024 * 1024 * 1024, Timestamp: time.Now()}
				require.NoError(t, store.WriteSystemMemory(context.Background(), &mem))
			},
			wantTotal: 16 * 1024 * 1024 * 1024,
			wantErr:   false,
		},
		{
			name:      "returns error when empty",
			setupCtx:  context.Background,
			setupData: func(_ *testing.T, _ *boltdb.Store) {},
			wantErr:   true,
		},
		{
			name: "fails with cancelled context",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupData: func(_ *testing.T, _ *boltdb.Store) {},
			wantErr:   true,
			wantErrIs: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			tc.setupData(t, store)
			ctx := tc.setupCtx()

			result, err := store.GetLatestSystemMemory(ctx)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantTotal, result.Total)
			}
		})
	}
}

// TestStore_GetLatestProcessMetrics tests retrieving the most recent process metrics.
func TestStore_GetLatestProcessMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		setupCtx    func() context.Context
		setupData   func(t *testing.T, store *boltdb.Store)
		wantPID     int
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:        "retrieves latest successfully",
			serviceName: "test-service",
			setupCtx:    context.Background,
			setupData: func(t *testing.T, store *boltdb.Store) {
				proc := metrics.ProcessMetrics{
					ServiceName: "test-service",
					PID:         1234,
					Timestamp:   time.Now(),
				}
				require.NoError(t, store.WriteProcessMetrics(context.Background(), &proc))
			},
			wantPID: 1234,
			wantErr: false,
		},
		{
			name:        "returns error for nonexistent service",
			serviceName: "nonexistent",
			setupCtx:    context.Background,
			setupData:   func(_ *testing.T, _ *boltdb.Store) {},
			wantErr:     true,
		},
		{
			name:        "fails with cancelled context",
			serviceName: "test-service",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupData: func(_ *testing.T, _ *boltdb.Store) {},
			wantErr:   true,
			wantErrIs: context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			tc.setupData(t, store)
			ctx := tc.setupCtx()

			result, err := store.GetLatestProcessMetrics(ctx, tc.serviceName)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantPID, result.PID)
			}
		})
	}
}

// =============================================================================
// PRUNE TESTS
// =============================================================================

// TestStore_Prune tests pruning old metrics data (covers pruneTransaction and pruneBucketHelper).
func TestStore_Prune(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		setupData   func(t *testing.T, store *boltdb.Store, old, recent time.Time)
		olderThan   time.Duration
		wantDeleted int
		wantErr     bool
		wantErrIs   error
	}{
		{
			name:     "prunes old system CPU metrics",
			setupCtx: context.Background,
			setupData: func(t *testing.T, store *boltdb.Store, old, recent time.Time) {
				oldCPU := metrics.SystemCPU{User: 100, Timestamp: old}
				newCPU := metrics.SystemCPU{User: 200, Timestamp: recent}
				require.NoError(t, store.WriteSystemCPU(context.Background(), &oldCPU))
				require.NoError(t, store.WriteSystemCPU(context.Background(), &newCPU))
			},
			olderThan:   time.Hour,
			wantDeleted: 1,
			wantErr:     false,
		},
		{
			name:     "prunes old system memory metrics",
			setupCtx: context.Background,
			setupData: func(t *testing.T, store *boltdb.Store, old, recent time.Time) {
				oldMem := metrics.SystemMemory{Total: 100, Timestamp: old}
				newMem := metrics.SystemMemory{Total: 200, Timestamp: recent}
				require.NoError(t, store.WriteSystemMemory(context.Background(), &oldMem))
				require.NoError(t, store.WriteSystemMemory(context.Background(), &newMem))
			},
			olderThan:   time.Hour,
			wantDeleted: 1,
			wantErr:     false,
		},
		{
			name:        "returns zero when no data to prune",
			setupCtx:    context.Background,
			setupData:   func(_ *testing.T, _ *boltdb.Store, _, _ time.Time) {},
			olderThan:   time.Hour,
			wantDeleted: 0,
			wantErr:     false,
		},
		{
			name: "fails with cancelled context",
			setupCtx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			setupData:   func(_ *testing.T, _ *boltdb.Store, _, _ time.Time) {},
			olderThan:   time.Hour,
			wantDeleted: 0,
			wantErr:     true,
			wantErrIs:   context.Canceled,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			old := time.Now().Add(-2 * time.Hour)
			recent := time.Now()
			tc.setupData(t, store, old, recent)
			ctx := tc.setupCtx()

			deleted, err := store.Prune(ctx, tc.olderThan)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.wantErrIs != nil {
					assert.ErrorIs(t, err, tc.wantErrIs)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantDeleted, deleted)
			}
		})
	}
}

// TestStore_PruneProcessMetrics tests pruning old process metrics (covers pruneProcessMetricsBuckets).
func TestStore_PruneProcessMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		setupData   func(t *testing.T, store *boltdb.Store, old, recent time.Time)
		olderThan   time.Duration
		wantDeleted int
		wantErr     bool
	}{
		{
			name:     "prunes old process metrics",
			setupCtx: context.Background,
			setupData: func(t *testing.T, store *boltdb.Store, old, recent time.Time) {
				oldProc := metrics.ProcessMetrics{ServiceName: "svc", PID: 1, Timestamp: old}
				newProc := metrics.ProcessMetrics{ServiceName: "svc", PID: 2, Timestamp: recent}
				require.NoError(t, store.WriteProcessMetrics(context.Background(), &oldProc))
				require.NoError(t, store.WriteProcessMetrics(context.Background(), &newProc))
			},
			olderThan:   time.Hour,
			wantDeleted: 1,
			wantErr:     false,
		},
		{
			name:     "prunes across multiple services",
			setupCtx: context.Background,
			setupData: func(t *testing.T, store *boltdb.Store, old, recent time.Time) {
				oldProc1 := metrics.ProcessMetrics{ServiceName: "svc1", PID: 1, Timestamp: old}
				oldProc2 := metrics.ProcessMetrics{ServiceName: "svc2", PID: 2, Timestamp: old}
				newProc1 := metrics.ProcessMetrics{ServiceName: "svc1", PID: 3, Timestamp: recent}
				require.NoError(t, store.WriteProcessMetrics(context.Background(), &oldProc1))
				require.NoError(t, store.WriteProcessMetrics(context.Background(), &oldProc2))
				require.NoError(t, store.WriteProcessMetrics(context.Background(), &newProc1))
			},
			olderThan:   time.Hour,
			wantDeleted: 2,
			wantErr:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			old := time.Now().Add(-2 * time.Hour)
			recent := time.Now()
			tc.setupData(t, store, old, recent)
			ctx := tc.setupCtx()

			deleted, err := store.Prune(ctx, tc.olderThan)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantDeleted, deleted)
			}
		})
	}
}

// =============================================================================
// CLOSE TESTS
// =============================================================================

// TestStore_Close tests closing the database connection.
func TestStore_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "closes successfully",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			config := storage.StoreConfig{
				Path: filepath.Join(t.TempDir(), "test.db"),
			}
			store, err := boltdb.NewStore(config)
			require.NoError(t, err)

			err = store.Close()

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

// TestStore_MultipleWrites tests writing and reading multiple metrics entries.
func TestStore_MultipleWrites(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		count     int
		wantCount int
	}{
		{
			name:      "writes and reads multiple entries",
			count:     testIterationCount,
			wantCount: testIterationCount,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			store := newTestStore(t)
			ctx := context.Background()

			base := time.Now()
			for i := range tc.count {
				cpu := metrics.SystemCPU{
					User:      uint64(i * 100),
					Timestamp: base.Add(time.Duration(i) * time.Second),
				}
				err := store.WriteSystemCPU(ctx, &cpu)
				require.NoError(t, err)
			}

			results, err := store.GetSystemCPU(ctx, base, base.Add(time.Duration(tc.count)*time.Second))
			require.NoError(t, err)
			assert.Len(t, results, tc.wantCount)

			// Verify ordering
			for i, r := range results {
				assert.Equal(t, uint64(i*100), r.User)
			}
		})
	}
}

// TestStore_Db verifies the Db accessor returns the underlying database.
//
// Params:
//   - t: testing context for assertions
func TestStore_Db(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		wantNotNil  bool
		description string
	}{
		{
			name:        "returns_valid_db",
			wantNotNil:  true,
			description: "Db should return a non-nil database instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := newTestStore(t)
			db := store.Db()

			if tt.wantNotNil {
				assert.NotNil(t, db, tt.description)
			}
		})
	}
}
