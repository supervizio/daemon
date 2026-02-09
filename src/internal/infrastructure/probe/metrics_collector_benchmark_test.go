//go:build cgo

package probe_test

/*
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/infrastructure/probe"
)

// BenchmarkCollectAllMetrics_Standard measures allocation/performance with standard config.
func BenchmarkCollectAllMetrics_Standard(b *testing.B) {
	// Initialize probe
	err := probe.Init()
	if err != nil {
		b.Fatalf("Failed to initialize probe: %v", err)
	}
	defer probe.Shutdown()

	ctx := context.Background()
	cfg := config.StandardMetricsConfig()

	// Reset timer to exclude setup
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = probe.CollectAllMetrics(ctx, &cfg)
	}
}

// BenchmarkCollectAllMetrics_Minimal measures allocation/performance with minimal config.
func BenchmarkCollectAllMetrics_Minimal(b *testing.B) {
	// Initialize probe
	err := probe.Init()
	if err != nil {
		b.Fatalf("Failed to initialize probe: %v", err)
	}
	defer probe.Shutdown()

	ctx := context.Background()
	cfg := config.MinimalMetricsConfig()

	// Reset timer to exclude setup
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = probe.CollectAllMetrics(ctx, &cfg)
	}
}

// BenchmarkCollectAllMetrics_NoConnections measures performance with connections disabled.
func BenchmarkCollectAllMetrics_NoConnections(b *testing.B) {
	// Initialize probe
	err := probe.Init()
	if err != nil {
		b.Fatalf("Failed to initialize probe: %v", err)
	}
	defer probe.Shutdown()

	ctx := context.Background()
	cfg := config.StandardMetricsConfig()
	cfg.Connections.Enabled = false

	// Reset timer to exclude setup
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = probe.CollectAllMetrics(ctx, &cfg)
	}
}

// BenchmarkCollectAllMetricsJSON_Standard measures JSON encoding performance with standard config.
func BenchmarkCollectAllMetricsJSON_Standard(b *testing.B) {
	// Initialize probe
	err := probe.Init()
	if err != nil {
		b.Fatalf("Failed to initialize probe: %v", err)
	}
	defer probe.Shutdown()

	ctx := context.Background()
	cfg := config.StandardMetricsConfig()

	// Reset timer to exclude setup
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = probe.CollectAllMetricsJSON(ctx, &cfg)
	}
}

// BenchmarkCollectAllMetricsJSON_Minimal measures JSON encoding performance with minimal config.
func BenchmarkCollectAllMetricsJSON_Minimal(b *testing.B) {
	// Initialize probe
	err := probe.Init()
	if err != nil {
		b.Fatalf("Failed to initialize probe: %v", err)
	}
	defer probe.Shutdown()

	ctx := context.Background()
	cfg := config.MinimalMetricsConfig()

	// Reset timer to exclude setup
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = probe.CollectAllMetricsJSON(ctx, &cfg)
	}
}

// BenchmarkCollectTCPConnections_WithPooling measures TCP connection collection with pooling.
func BenchmarkCollectTCPConnections_WithPooling(b *testing.B) {
	// Initialize probe
	err := probe.Init()
	if err != nil {
		b.Fatalf("Failed to initialize probe: %v", err)
	}
	defer probe.Shutdown()

	ctx := context.Background()

	// Reset timer to exclude setup
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		collector := probe.NewConnectionCollector()
		_, _ = collector.CollectTCP(ctx)
	}
}

// BenchmarkStringCaching_Stable measures C string caching for stable strings.
func BenchmarkStringCaching_Stable(b *testing.B) {
	// Clear cache before benchmark
	probe.ClearCStringCacheForTest()

	// Create test input (simulating device name)
	input := []C.char{
		C.char('s'), C.char('d'), C.char('a'), C.char('1'),
	}

	// Reset timer
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = probe.CCharArrayToStringCachedForTest(input, true)
	}
}

// BenchmarkStringCaching_Unstable measures C string conversion for unstable strings.
func BenchmarkStringCaching_Unstable(b *testing.B) {
	// Create test input
	input := []C.char{
		C.char('1'), C.char('2'), C.char('7'), C.char('.'),
		C.char('0'), C.char('.'), C.char('0'), C.char('.'),
		C.char('1'),
	}

	// Reset timer
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = probe.CCharArrayToStringCachedForTest(input, false)
	}
}

// BenchmarkJSONBufferPool measures JSON buffer pooling performance.
func BenchmarkJSONBufferPool(b *testing.B) {
	// Reset timer
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := probe.GetJSONBufferForTest()
		buf.WriteString("test data for benchmarking")
		probe.PutJSONBufferForTest(buf)
	}
}

// BenchmarkTCPConnPool measures TCP connection slice pooling performance.
func BenchmarkTCPConnPool(b *testing.B) {
	// Reset timer
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		slice := probe.GetTCPConnSliceForTest()
		*slice = append(*slice, probe.TcpConnJSON{
			Family:    "ipv4",
			LocalAddr: "127.0.0.1",
			LocalPort: 8080,
		})
		probe.PutTCPConnSliceForTest(slice)
	}
}
