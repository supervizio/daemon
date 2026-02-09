//go:build !race

package boltdb_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/shared"
	"github.com/kodflow/daemon/internal/infrastructure/persistence/storage/boltdb"
)

// createTestMetrics creates sample metrics for benchmarking.
func createTestMetrics(serviceName string, pid int) *metrics.ProcessMetrics {
	return &metrics.ProcessMetrics{
		ServiceName: serviceName,
		PID:         pid,
		CPU: &metrics.ProcessCPU{
			PID:          pid,
			UsagePercent: 12.5,
			Timestamp:    time.Now(),
		},
		Memory: &metrics.ProcessMemory{
			PID:       pid,
			RSS:       1024 * 1024 * 50, // 50MB
			VMS:       1024 * 1024 * 100, // 100MB
			Timestamp: time.Now(),
		},
		State:     "running",
		Timestamp: time.Now(),
	}
}

// BenchmarkStoreWrite measures metrics write performance.
func BenchmarkStoreWrite(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "bench.db")
	store, err := boltdb.NewStore(dbPath, boltdb.DefaultStoreConfig())
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	metrics := createTestMetrics("test-service", 1234)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = store.Write(ctx, metrics)
	}
}

// BenchmarkStoreWriteBatch measures batch write performance.
func BenchmarkStoreWriteBatch(b *testing.B) {
	batchSizes := []int{1, 10, 50, 100}

	for _, size := range batchSizes {
		b.Run(string(rune('0'+size)), func(b *testing.B) {
			tmpDir, err := os.MkdirTemp("", "benchmark-*")
			if err != nil {
				b.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			dbPath := filepath.Join(tmpDir, "bench.db")
			store, err := boltdb.NewStore(dbPath, boltdb.DefaultStoreConfig())
			if err != nil {
				b.Fatalf("Failed to create store: %v", err)
			}
			defer store.Close()

			ctx := context.Background()

			// Create batch
			batch := make([]*metrics.ProcessMetrics, size)
			for i := 0; i < size; i++ {
				batch[i] = createTestMetrics("test-service", 1000+i)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				for _, m := range batch {
					_ = store.Write(ctx, m)
				}
			}
		})
	}
}

// BenchmarkStoreRead measures metrics read performance.
func BenchmarkStoreRead(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "bench.db")
	store, err := boltdb.NewStore(dbPath, boltdb.DefaultStoreConfig())
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Pre-populate with data
	for i := 0; i < 100; i++ {
		metrics := createTestMetrics("test-service", 1000+i)
		_ = store.Write(ctx, metrics)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = store.Read(ctx, "test-service", time.Minute)
	}
}

// BenchmarkStoreReadWithRetention measures read with different retention periods.
func BenchmarkStoreReadWithRetention(b *testing.B) {
	retentions := []struct {
		name     string
		duration shared.Duration
	}{
		{"1Min", shared.Duration(time.Minute)},
		{"5Min", shared.Duration(5 * time.Minute)},
		{"1Hour", shared.Duration(time.Hour)},
		{"24Hour", shared.Duration(24 * time.Hour)},
	}

	for _, retention := range retentions {
		b.Run(retention.name, func(b *testing.B) {
			tmpDir, err := os.MkdirTemp("", "benchmark-*")
			if err != nil {
				b.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			dbPath := filepath.Join(tmpDir, "bench.db")
			store, err := boltdb.NewStore(dbPath, boltdb.DefaultStoreConfig())
			if err != nil {
				b.Fatalf("Failed to create store: %v", err)
			}
			defer store.Close()

			ctx := context.Background()

			// Pre-populate
			for i := 0; i < 1000; i++ {
				metrics := createTestMetrics("test-service", 1000+i)
				_ = store.Write(ctx, metrics)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = store.Read(ctx, "test-service", time.Duration(retention.duration))
			}
		})
	}
}

// BenchmarkStoreCleanup measures cleanup performance.
func BenchmarkStoreCleanup(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "bench.db")
	store, err := boltdb.NewStore(dbPath, boltdb.DefaultStoreConfig())
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Pre-populate
		for j := 0; j < 100; j++ {
			metrics := createTestMetrics("test-service", 1000+j)
			_ = store.Write(ctx, metrics)
		}

		// Cleanup
		_ = store.Cleanup(ctx, shared.Duration(time.Millisecond))
	}
}

// BenchmarkStoreListServices measures service enumeration performance.
func BenchmarkStoreListServices(b *testing.B) {
	servicesCounts := []int{1, 10, 50, 100}

	for _, count := range servicesCounts {
		b.Run(string(rune('0'+count)), func(b *testing.B) {
			tmpDir, err := os.MkdirTemp("", "benchmark-*")
			if err != nil {
				b.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			dbPath := filepath.Join(tmpDir, "bench.db")
			store, err := boltdb.NewStore(dbPath, boltdb.DefaultStoreConfig())
			if err != nil {
				b.Fatalf("Failed to create store: %v", err)
			}
			defer store.Close()

			ctx := context.Background()

			// Pre-populate with N services
			for i := 0; i < count; i++ {
				serviceName := "service-" + string(rune('a'+i))
				metrics := createTestMetrics(serviceName, 1000+i)
				_ = store.Write(ctx, metrics)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = store.ListServices(ctx)
			}
		})
	}
}

// BenchmarkStoreConcurrentWrites measures concurrent write performance.
func BenchmarkStoreConcurrentWrites(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "benchmark-*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "bench.db")
	store, err := boltdb.NewStore(dbPath, boltdb.DefaultStoreConfig())
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			metrics := createTestMetrics("test-service", 1000+i)
			_ = store.Write(ctx, metrics)
			i++
		}
	})
}
