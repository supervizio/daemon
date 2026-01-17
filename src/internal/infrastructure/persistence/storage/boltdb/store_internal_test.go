//go:build linux

// Package boltdb provides internal tests for unexported functions.
package boltdb

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/storage"
)

// =============================================================================
// TEST HELPERS
// =============================================================================

// newInternalTestStore creates a store for internal tests.
func newInternalTestStore(t *testing.T) *Store {
	t.Helper()
	config := storage.StoreConfig{
		Path:          filepath.Join(t.TempDir(), "internal_test.db"),
		Retention:     24 * time.Hour,
		PruneInterval: time.Hour,
	}
	store, err := NewStore(config)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	return store
}

// =============================================================================
// NEWSTORE ERROR TESTS
// =============================================================================

// TestNewStore_errors tests NewStore error conditions.
func TestNewStore_errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupPath func(t *testing.T) string
		wantErr   bool
	}{
		{
			name: "fails with directory path",
			setupPath: func(t *testing.T) string {
				// Return a directory path (cannot open as bolt db)
				return t.TempDir()
			},
			wantErr: true,
		},
		{
			name: "fails with non-existent directory",
			setupPath: func(_ *testing.T) string {
				// Return path in non-existent directory
				return "/nonexistent/path/to/db.bolt"
			},
			wantErr: true,
		},
		{
			name: "fails with read-only file",
			setupPath: func(t *testing.T) string {
				// Create read-only file
				path := filepath.Join(t.TempDir(), "readonly.db")
				err := os.WriteFile(path, []byte("not a db"), 0o000)
				require.NoError(t, err)
				t.Cleanup(func() {
					// Restore permissions for cleanup
					_ = os.Chmod(path, 0o644)
				})
				return path
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.setupPath(t)
			config := storage.StoreConfig{
				Path:          path,
				Retention:     24 * time.Hour,
				PruneInterval: time.Hour,
			}

			store, err := NewStore(config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, store)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, store)
				if store != nil {
					_ = store.Close()
				}
			}
		})
	}
}

// =============================================================================
// INITSCHEMA TESTS (PRIVATE FUNCTION)
// =============================================================================

// TestStore_initSchema tests database schema initialization (private method).
func TestStore_initSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		verify func(t *testing.T, store *Store)
	}{
		{
			name: "creates all required buckets",
			verify: func(t *testing.T, store *Store) {
				err := store.db.View(func(tx *bolt.Tx) error {
					cpuBucket := tx.Bucket(bucketSystemCPU)
					assert.NotNil(t, cpuBucket)
					memBucket := tx.Bucket(bucketSystemMemory)
					assert.NotNil(t, memBucket)
					procBucket := tx.Bucket(bucketProcessMetrics)
					assert.NotNil(t, procBucket)
					metaBucket := tx.Bucket(bucketMetadata)
					assert.NotNil(t, metaBucket)
					return nil
				})
				require.NoError(t, err)
			},
		},
		{
			name: "sets metadata on fresh database",
			verify: func(t *testing.T, store *Store) {
				err := store.db.View(func(tx *bolt.Tx) error {
					meta := tx.Bucket(bucketMetadata)
					require.NotNil(t, meta)
					created := meta.Get(keyCreated)
					assert.NotNil(t, created)
					version := meta.Get(keyVersion)
					assert.NotNil(t, version)
					return nil
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newInternalTestStore(t)
			tt.verify(t, store)
		})
	}
}

// =============================================================================
// PRUNETRANSACTION TESTS (PRIVATE FUNCTION)
// =============================================================================

// TestStore_pruneTransaction tests the internal pruning transaction logic (private method).
func TestStore_pruneTransaction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupStore  func(t *testing.T, store *Store) (old, recent time.Time)
		wantDeleted int
		verifyMeta  bool
	}{
		{
			name: "prunes old entries",
			setupStore: func(t *testing.T, store *Store) (time.Time, time.Time) {
				ctx := t.Context()
				old := time.Now().Add(-2 * time.Hour)
				recent := time.Now()
				require.NoError(t, store.WriteSystemCPU(ctx, &metrics.SystemCPU{User: 100, Timestamp: old}))
				require.NoError(t, store.WriteSystemCPU(ctx, &metrics.SystemCPU{User: 200, Timestamp: recent}))
				require.NoError(t, store.WriteSystemMemory(ctx, &metrics.SystemMemory{Total: 100, Timestamp: old}))
				require.NoError(t, store.WriteSystemMemory(ctx, &metrics.SystemMemory{Total: 200, Timestamp: recent}))
				return old, recent
			},
			wantDeleted: 2,
		},
		{
			name: "updates last prune timestamp",
			setupStore: func(_ *testing.T, _ *Store) (time.Time, time.Time) {
				return time.Now().Add(-2 * time.Hour), time.Now()
			},
			wantDeleted: 0,
			verifyMeta:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newInternalTestStore(t)
			old, recent := tt.setupStore(t, store)
			cutoffKey := timeToKey(time.Now().Add(-time.Hour))
			deleted, err := store.pruneTransaction(cutoffKey)
			require.NoError(t, err)
			assert.Equal(t, tt.wantDeleted, deleted)

			if tt.wantDeleted > 0 {
				ctx := t.Context()
				cpuResults, err := store.GetSystemCPU(ctx, old.Add(-time.Hour), recent.Add(time.Hour))
				require.NoError(t, err)
				assert.Len(t, cpuResults, 1)
				assert.Equal(t, uint64(200), cpuResults[0].User)
			}

			if tt.verifyMeta {
				verifyErr := store.db.View(func(tx *bolt.Tx) error {
					meta := tx.Bucket(bucketMetadata)
					lastPrune := meta.Get(keyLastPrune)
					assert.NotNil(t, lastPrune)
					return nil
				})
				require.NoError(t, verifyErr)
			}
		})
	}
}

// =============================================================================
// PRUNEPROCESSMETRICSBUCKETS TESTS (PRIVATE FUNCTION)
// =============================================================================

// TestStore_pruneProcessMetricsBuckets tests process metrics bucket pruning (private method).
func TestStore_pruneProcessMetricsBuckets(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		services    []string
		setupData   func(t *testing.T, store *Store, old, recent time.Time)
		wantDeleted int
		verifyData  func(t *testing.T, store *Store, services []string, old, recent time.Time)
	}{
		{
			name:     "prunes old entries across services",
			services: []string{"svc1", "svc2"},
			setupData: func(t *testing.T, store *Store, old, recent time.Time) {
				ctx := t.Context()
				for _, svc := range []string{"svc1", "svc2"} {
					require.NoError(t, store.WriteProcessMetrics(ctx, &metrics.ProcessMetrics{
						ServiceName: svc, PID: 1, Timestamp: old,
					}))
					require.NoError(t, store.WriteProcessMetrics(ctx, &metrics.ProcessMetrics{
						ServiceName: svc, PID: 2, Timestamp: recent,
					}))
				}
			},
			wantDeleted: 2,
			verifyData: func(t *testing.T, store *Store, services []string, old, recent time.Time) {
				ctx := t.Context()
				for _, svc := range services {
					results, err := store.GetProcessMetrics(ctx, svc, old.Add(-time.Hour), recent.Add(time.Hour))
					require.NoError(t, err)
					assert.Len(t, results, 1)
					assert.Equal(t, 2, results[0].PID)
				}
			},
		},
		{
			name:     "returns zero when no data to prune",
			services: []string{"svc1"},
			setupData: func(t *testing.T, store *Store, _, recent time.Time) {
				ctx := t.Context()
				require.NoError(t, store.WriteProcessMetrics(ctx, &metrics.ProcessMetrics{
					ServiceName: "svc1", PID: 1, Timestamp: recent,
				}))
			},
			wantDeleted: 0,
			verifyData: func(t *testing.T, store *Store, _ []string, old, recent time.Time) {
				ctx := t.Context()
				results, err := store.GetProcessMetrics(ctx, "svc1", old.Add(-time.Hour), recent.Add(time.Hour))
				require.NoError(t, err)
				assert.Len(t, results, 1)
			},
		},
		{
			name:        "handles empty bucket",
			services:    []string{},
			setupData:   func(_ *testing.T, _ *Store, _, _ time.Time) {},
			wantDeleted: 0,
			verifyData:  func(_ *testing.T, _ *Store, _ []string, _, _ time.Time) {},
		},
		{
			name:     "handles single service with multiple entries",
			services: []string{"single-svc"},
			setupData: func(t *testing.T, store *Store, old, recent time.Time) {
				ctx := t.Context()
				for i := range 3 {
					require.NoError(t, store.WriteProcessMetrics(ctx, &metrics.ProcessMetrics{
						ServiceName: "single-svc", PID: i + 1, Timestamp: old.Add(time.Duration(i) * time.Minute),
					}))
				}
				require.NoError(t, store.WriteProcessMetrics(ctx, &metrics.ProcessMetrics{
					ServiceName: "single-svc", PID: 100, Timestamp: recent,
				}))
			},
			wantDeleted: 3,
			verifyData: func(t *testing.T, store *Store, _ []string, old, recent time.Time) {
				ctx := t.Context()
				results, err := store.GetProcessMetrics(ctx, "single-svc", old.Add(-time.Hour), recent.Add(time.Hour))
				require.NoError(t, err)
				assert.Len(t, results, 1)
				assert.Equal(t, 100, results[0].PID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newInternalTestStore(t)
			ctx := t.Context()

			old := time.Now().Add(-2 * time.Hour)
			recent := time.Now()

			tt.setupData(t, store, old, recent)

			deleted, err := store.Prune(ctx, time.Hour)
			require.NoError(t, err)
			assert.Equal(t, tt.wantDeleted, deleted)

			tt.verifyData(t, store, tt.services, old, recent)
		})
	}
}

// =============================================================================
// PRUNEBUCKETHELPER TESTS (PRIVATE FUNCTION)
// =============================================================================

// TestStore_pruneBucketHelper tests the bucket pruning helper (private method).
func TestStore_pruneBucketHelper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		writeCount  int
		oldCount    int
		wantDeleted int
		wantRemain  int
	}{
		{
			name:        "deletes entries before cutoff",
			writeCount:  3,
			oldCount:    2,
			wantDeleted: 2,
			wantRemain:  1,
		},
		{
			name:        "returns zero when nothing to delete",
			writeCount:  1,
			oldCount:    0,
			wantDeleted: 0,
			wantRemain:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := newInternalTestStore(t)
			ctx := t.Context()

			old := time.Now().Add(-2 * time.Hour)
			recent := time.Now()

			for i := range tt.oldCount {
				require.NoError(t, store.WriteSystemCPU(ctx, &metrics.SystemCPU{
					User:      uint64((i + 1) * 50),
					Timestamp: old.Add(time.Duration(i) * time.Minute),
				}))
			}
			for i := range tt.writeCount - tt.oldCount {
				require.NoError(t, store.WriteSystemCPU(ctx, &metrics.SystemCPU{
					User:      uint64((i + 1) * 100),
					Timestamp: recent.Add(time.Duration(i) * time.Minute),
				}))
			}

			deleted, err := store.Prune(ctx, time.Hour)
			require.NoError(t, err)
			assert.Equal(t, tt.wantDeleted, deleted)

			results, err := store.GetSystemCPU(ctx, old.Add(-time.Hour), recent.Add(time.Hour))
			require.NoError(t, err)
			assert.Len(t, results, tt.wantRemain)
		})
	}
}

// =============================================================================
// TIMETOKEY TESTS (PRIVATE FUNCTION)
// =============================================================================

// Test_timeToKey tests time to key conversion (private function).
func Test_timeToKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func() (time.Time, time.Time)
		wantOrdered bool
		wantLen     int
	}{
		{
			name: "converts time to sortable bytes",
			setup: func() (time.Time, time.Time) {
				now := time.Now()
				return now, now.Add(time.Hour)
			},
			wantOrdered: true,
			wantLen:     int64ByteLength,
		},
		{
			name: "produces fixed length key",
			setup: func() (time.Time, time.Time) {
				return time.Now(), time.Time{}
			},
			wantOrdered: false,
			wantLen:     int64ByteLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t1, t2 := tt.setup()
			key1 := timeToKey(t1)
			assert.Len(t, key1, tt.wantLen)

			if tt.wantOrdered {
				key2 := timeToKey(t2)
				assert.Len(t, key2, tt.wantLen)
				assert.Greater(t, string(key2), string(key1))
			}
		})
	}
}

// =============================================================================
// INT64TOBYTES TESTS (PRIVATE FUNCTION)
// =============================================================================

// Test_int64ToBytes tests int64 to bytes conversion (private function).
func Test_int64ToBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       int64
		compareWith int64
		wantLen     int
		wantSmaller bool
	}{
		{
			name:    "converts positive int64 to bytes",
			input:   12345,
			wantLen: int64ByteLength,
		},
		{
			name:        "smaller values produce lexicographically smaller keys",
			input:       100,
			compareWith: 1000,
			wantLen:     int64ByteLength,
			wantSmaller: true,
		},
		{
			name:        "preserves ordering",
			input:       100,
			compareWith: 200,
			wantLen:     int64ByteLength,
			wantSmaller: true,
		},
		{
			name:    "handles zero",
			input:   0,
			wantLen: int64ByteLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := int64ToBytes(tt.input)
			assert.Len(t, data, tt.wantLen)

			if tt.wantSmaller {
				larger := int64ToBytes(tt.compareWith)
				assert.True(t, bytes.Compare(data, larger) < 0)
			}
		})
	}
}

// =============================================================================
// ENCODESYSTEMCPU TESTS (PRIVATE FUNCTION)
// =============================================================================

// Test_encodeSystemCPU tests system CPU encoding (private function).
func Test_encodeSystemCPU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cpu        metrics.SystemCPU
		verifyTrip bool
	}{
		{
			name: "encodes struct successfully",
			cpu: metrics.SystemCPU{
				User:   1000,
				System: 500,
			},
		},
		{
			name: "round trip",
			cpu: metrics.SystemCPU{
				User:   1000,
				System: 500,
				Idle:   8500,
			},
			verifyTrip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := encodeSystemCPU(&tt.cpu)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			if tt.verifyTrip {
				var decoded metrics.SystemCPU
				err = decodeSystemCPU(data, &decoded)
				require.NoError(t, err)
				assert.Equal(t, tt.cpu.User, decoded.User)
				assert.Equal(t, tt.cpu.System, decoded.System)
				assert.Equal(t, tt.cpu.Idle, decoded.Idle)
			}
		})
	}
}

// =============================================================================
// ENCODESYSTEMMEMORY TESTS (PRIVATE FUNCTION)
// =============================================================================

// Test_encodeSystemMemory tests system memory encoding (private function).
func Test_encodeSystemMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mem        metrics.SystemMemory
		verifyTrip bool
	}{
		{
			name: "encodes struct successfully",
			mem: metrics.SystemMemory{
				Total:     16 * 1024 * 1024 * 1024,
				Available: 8 * 1024 * 1024 * 1024,
			},
		},
		{
			name: "round trip",
			mem: metrics.SystemMemory{
				Total:     16 * 1024 * 1024 * 1024,
				Available: 8 * 1024 * 1024 * 1024,
			},
			verifyTrip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := encodeSystemMemory(&tt.mem)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			if tt.verifyTrip {
				var decoded metrics.SystemMemory
				err = decodeSystemMemory(data, &decoded)
				require.NoError(t, err)
				assert.Equal(t, tt.mem.Total, decoded.Total)
				assert.Equal(t, tt.mem.Available, decoded.Available)
			}
		})
	}
}

// =============================================================================
// ENCODEPROCESSMETRICS TESTS (PRIVATE FUNCTION)
// =============================================================================

// Test_encodeProcessMetrics tests process metrics encoding (private function).
func Test_encodeProcessMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		proc       metrics.ProcessMetrics
		verifyTrip bool
	}{
		{
			name: "encodes struct successfully",
			proc: metrics.ProcessMetrics{
				ServiceName: "test-service",
				PID:         1234,
			},
		},
		{
			name: "round trip",
			proc: metrics.ProcessMetrics{
				ServiceName: "test-service",
				PID:         1234,
			},
			verifyTrip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := encodeProcessMetrics(&tt.proc)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			if tt.verifyTrip {
				var decoded metrics.ProcessMetrics
				err = decodeProcessMetrics(data, &decoded)
				require.NoError(t, err)
				assert.Equal(t, tt.proc.ServiceName, decoded.ServiceName)
				assert.Equal(t, tt.proc.PID, decoded.PID)
			}
		})
	}
}

// =============================================================================
// DECODESYSTEMCPU TESTS (PRIVATE FUNCTION)
// =============================================================================

// Test_decodeSystemCPU tests system CPU decoding (private function).
func Test_decodeSystemCPU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupData   func() []byte
		wantErr     bool
		verifyValue bool
		wantUser    uint64
	}{
		{
			name: "decodes encoded value",
			setupData: func() []byte {
				original := metrics.SystemCPU{
					User:   1000,
					System: 500,
					Idle:   8500,
				}
				data, _ := encodeSystemCPU(&original)
				return data
			},
			wantErr:     false,
			verifyValue: true,
			wantUser:    1000,
		},
		{
			name: "returns error for invalid data",
			setupData: func() []byte {
				return []byte("not valid gob data")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := tt.setupData()
			var cpu metrics.SystemCPU
			err := decodeSystemCPU(data, &cpu)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.verifyValue {
					assert.Equal(t, tt.wantUser, cpu.User)
				}
			}
		})
	}
}

// =============================================================================
// DECODESYSTEMMEMORY TESTS (PRIVATE FUNCTION)
// =============================================================================

// Test_decodeSystemMemory tests system memory decoding (private function).
func Test_decodeSystemMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupData func() []byte
		wantErr   bool
	}{
		{
			name: "returns error for invalid data",
			setupData: func() []byte {
				return []byte("not valid gob data")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := tt.setupData()
			var mem metrics.SystemMemory
			err := decodeSystemMemory(data, &mem)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// =============================================================================
// DECODEPROCESSMETRICS TESTS (PRIVATE FUNCTION)
// =============================================================================

// Test_decodeProcessMetrics tests process metrics decoding (private function).
func Test_decodeProcessMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupData func() []byte
		wantErr   bool
	}{
		{
			name: "returns error for invalid data",
			setupData: func() []byte {
				return []byte("not valid gob data")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := tt.setupData()
			var proc metrics.ProcessMetrics
			err := decodeProcessMetrics(data, &proc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
