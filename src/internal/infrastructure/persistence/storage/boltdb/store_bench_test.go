//go:build !race

package boltdb_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/storage"
	"github.com/kodflow/daemon/internal/infrastructure/persistence/storage/boltdb"
)

// createTestMetrics creates sample metrics for benchmarking.
func createTestMetrics(serviceName string, pid int) *metrics.ProcessMetrics {
	return &metrics.ProcessMetrics{
		ServiceName: serviceName,
		PID:         pid,
		CPU: metrics.ProcessCPU{
			PID:       pid,
			Timestamp: time.Now(),
		},
		Memory: metrics.ProcessMemory{
			PID:       pid,
			RSS:       1024 * 1024 * 50,
			VMS:       1024 * 1024 * 100,
			Timestamp: time.Now(),
		},
		State:     process.StateRunning,
		Timestamp: time.Now(),
	}
}

// newBenchStore creates a temporary store for benchmarking.
func newBenchStore(b *testing.B) (*boltdb.Store, func()) {
	b.Helper()
	tmpDir := b.TempDir()
	cfg := storage.StoreConfig{
		Path:          filepath.Join(tmpDir, "bench.db"),
		Retention:     24 * time.Hour,
		PruneInterval: time.Hour,
	}
	store, err := boltdb.NewStore(cfg)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	cleanup := func() {
		store.Close()
	}
	return store, cleanup
}

// BenchmarkStoreWrite measures metrics write performance.
func BenchmarkStoreWrite(b *testing.B) {
	store, cleanup := newBenchStore(b)
	defer cleanup()

	ctx := context.Background()
	m := createTestMetrics("test-service", 1234)

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_ = store.WriteProcessMetrics(ctx, m)
	}
}

// BenchmarkStoreWriteBatch measures batch write performance.
func BenchmarkStoreWriteBatch(b *testing.B) {
	batchSizes := []struct {
		name string
		size int
	}{
		{"1", 1},
		{"10", 10},
		{"50", 50},
		{"100", 100},
	}

	for _, bs := range batchSizes {
		b.Run(bs.name, func(b *testing.B) {
			store, cleanup := newBenchStore(b)
			defer cleanup()

			ctx := context.Background()

			batch := make([]*metrics.ProcessMetrics, bs.size)
			for i := range bs.size {
				batch[i] = createTestMetrics("test-service", 1000+i)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				for _, m := range batch {
					_ = store.WriteProcessMetrics(ctx, m)
				}
			}
		})
	}
}

// BenchmarkStoreRead measures metrics read performance.
func BenchmarkStoreRead(b *testing.B) {
	store, cleanup := newBenchStore(b)
	defer cleanup()

	ctx := context.Background()

	for i := range 100 {
		m := createTestMetrics("test-service", 1000+i)
		_ = store.WriteProcessMetrics(ctx, m)
	}

	since := time.Now().Add(-time.Minute)
	until := time.Now()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_, _ = store.GetProcessMetrics(ctx, "test-service", since, until)
	}
}

// BenchmarkStorePrune measures prune performance.
func BenchmarkStorePrune(b *testing.B) {
	store, cleanup := newBenchStore(b)
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		for j := range 100 {
			m := createTestMetrics("test-service", 1000+j)
			_ = store.WriteProcessMetrics(ctx, m)
		}

		_, _ = store.Prune(ctx, time.Millisecond)
	}
}

// BenchmarkStoreConcurrentWrites measures concurrent write performance.
func BenchmarkStoreConcurrentWrites(b *testing.B) {
	store, cleanup := newBenchStore(b)
	defer cleanup()

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			m := createTestMetrics("test-service", 1000+i)
			_ = store.WriteProcessMetrics(ctx, m)
			i++
		}
	})
}
