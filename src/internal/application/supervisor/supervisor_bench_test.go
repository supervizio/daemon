//go:build !race

package supervisor_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/application/supervisor"
	domainconfig "github.com/kodflow/daemon/internal/domain/config"
	domainprocess "github.com/kodflow/daemon/internal/domain/process"
)

// benchmarkExecutor implements domain.Executor for benchmarking.
type benchmarkExecutor struct{}

func (m *benchmarkExecutor) Start(_ context.Context, _ domainprocess.Spec) (int, <-chan domainprocess.ExitResult, error) {
	ch := make(chan domainprocess.ExitResult, 1)
	return 1234, ch, nil
}

func (m *benchmarkExecutor) Stop(_ int, _ time.Duration) error {
	return nil
}

func (m *benchmarkExecutor) Signal(_ int, _ os.Signal) error {
	return nil
}

// benchmarkLoader implements config.Loader for benchmarking.
type benchmarkLoader struct {
	cfg *domainconfig.Config
}

func (m *benchmarkLoader) Load(_ string) (*domainconfig.Config, error) {
	return m.cfg, nil
}

// benchmarkReaper implements lifecycle.Reaper for benchmarking.
type benchmarkReaper struct{}

func (m *benchmarkReaper) Start() {}

func (m *benchmarkReaper) Stop() {}

func (m *benchmarkReaper) ReapOnce() int {
	return 0
}

func (m *benchmarkReaper) IsPID1() bool {
	return false
}

// createTestConfig creates a test configuration with N services.
func createTestConfig(serviceCount int) *domainconfig.Config {
	services := make([]domainconfig.ServiceConfig, serviceCount)
	for i := range serviceCount {
		services[i] = domainconfig.ServiceConfig{
			Name:    "test-service-" + string(rune('a'+i)),
			Command: "/bin/sleep",
			Args:    []string{"infinity"},
			Restart: domainconfig.RestartConfig{
				Policy: domainconfig.RestartNever,
			},
		}
	}
	return &domainconfig.Config{
		Services: services,
	}
}

// BenchmarkSupervisorStart measures supervisor startup with N services.
func BenchmarkSupervisorStart(b *testing.B) {
	benchmarks := []struct {
		name         string
		serviceCount int
	}{
		{"1Service", 1},
		{"5Services", 5},
		{"10Services", 10},
		{"20Services", 20},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			cfg := createTestConfig(bm.serviceCount)
			loader := &benchmarkLoader{cfg: cfg}
			executor := &benchmarkExecutor{}
			reaper := &benchmarkReaper{}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				sup, _ := supervisor.NewSupervisor(cfg, loader, executor, reaper)
				ctx := context.Background()
				_ = sup.Start(ctx)
				_ = sup.Stop()
			}
		})
	}
}

// BenchmarkSupervisorReload measures configuration reload performance.
func BenchmarkSupervisorReload(b *testing.B) {
	cfg := createTestConfig(5)
	loader := &benchmarkLoader{cfg: cfg}
	executor := &benchmarkExecutor{}
	reaper := &benchmarkReaper{}

	sup, _ := supervisor.NewSupervisor(cfg, loader, executor, reaper)
	ctx := context.Background()
	_ = sup.Start(ctx)
	defer sup.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_ = sup.Reload()
	}
}

// BenchmarkSupervisorServices measures Services() list retrieval.
func BenchmarkSupervisorServices(b *testing.B) {
	benchmarks := []struct {
		name         string
		serviceCount int
	}{
		{"1Service", 1},
		{"10Services", 10},
		{"50Services", 50},
		{"100Services", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			cfg := createTestConfig(bm.serviceCount)
			loader := &benchmarkLoader{cfg: cfg}
			executor := &benchmarkExecutor{}
			reaper := &benchmarkReaper{}

			sup, _ := supervisor.NewSupervisor(cfg, loader, executor, reaper)
			ctx := context.Background()
			_ = sup.Start(ctx)
			defer sup.Stop()

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				_ = sup.Services()
			}
		})
	}
}

// BenchmarkSupervisorStats measures statistics retrieval performance.
func BenchmarkSupervisorStats(b *testing.B) {
	cfg := createTestConfig(10)
	loader := &benchmarkLoader{cfg: cfg}
	executor := &benchmarkExecutor{}
	reaper := &benchmarkReaper{}

	sup, _ := supervisor.NewSupervisor(cfg, loader, executor, reaper)
	ctx := context.Background()
	_ = sup.Start(ctx)
	defer sup.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_ = sup.AllStats()
	}
}

// BenchmarkSupervisorConcurrentAccess measures concurrent read performance.
func BenchmarkSupervisorConcurrentAccess(b *testing.B) {
	cfg := createTestConfig(10)
	loader := &benchmarkLoader{cfg: cfg}
	executor := &benchmarkExecutor{}
	reaper := &benchmarkReaper{}

	sup, _ := supervisor.NewSupervisor(cfg, loader, executor, reaper)
	ctx := context.Background()
	_ = sup.Start(ctx)
	defer sup.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = sup.Services()
			_ = sup.AllStats()
			_ = sup.State()
		}
	})
}
