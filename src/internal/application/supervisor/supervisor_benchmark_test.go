//go:build !race

package supervisor_test

import (
	"context"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/application/supervisor"
	domainconfig "github.com/kodflow/daemon/internal/domain/config"
	domainlifecycle "github.com/kodflow/daemon/internal/domain/lifecycle"
	domainprocess "github.com/kodflow/daemon/internal/domain/process"
	domainshared "github.com/kodflow/daemon/internal/domain/shared"
)

// benchmarkExecutor implements domain.Executor for benchmarking.
type benchmarkExecutor struct{}

func (m *benchmarkExecutor) Start(ctx context.Context, spec *domainprocess.Spec) (*domainprocess.State, error) {
	return &domainprocess.State{
		PID: 1234,
	}, nil
}

func (m *benchmarkExecutor) Stop(ctx context.Context, pid int, timeout domainshared.Duration) error {
	return nil
}

func (m *benchmarkExecutor) Signal(pid int, sig string) error {
	return nil
}

// benchmarkLoader implements config.Loader for benchmarking.
type benchmarkLoader struct {
	cfg *domainconfig.Config
}

func (m *benchmarkLoader) Load(ctx context.Context, path string) (*domainconfig.Config, error) {
	return m.cfg, nil
}

func (m *benchmarkLoader) Validate(cfg *domainconfig.Config) error {
	return nil
}

// benchmarkReaper implements lifecycle.Reaper for benchmarking.
type benchmarkReaper struct{}

func (m *benchmarkReaper) Start(ctx context.Context) error {
	return nil
}

func (m *benchmarkReaper) Stop(ctx context.Context) error {
	return nil
}

func (m *benchmarkReaper) WaitOne(ctx context.Context) (int, error) {
	return -1, nil
}

// createTestConfig creates a test configuration with N services.
func createTestConfig(serviceCount int) *domainconfig.Config {
	services := make([]domainconfig.ServiceConfig, serviceCount)
	for i := 0; i < serviceCount; i++ {
		services[i] = domainconfig.ServiceConfig{
			Name:    "test-service-" + string(rune('a'+i)),
			Command: "/bin/sleep",
			Args:    []string{"infinity"},
			Restart: domainconfig.RestartConfig{
				Policy: domainconfig.RestartPolicyNever,
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

			for i := 0; i < b.N; i++ {
				sup := supervisor.NewSupervisor(cfg, loader, executor, reaper)
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

	sup := supervisor.NewSupervisor(cfg, loader, executor, reaper)
	ctx := context.Background()
	_ = sup.Start(ctx)
	defer sup.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
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

			sup := supervisor.NewSupervisor(cfg, loader, executor, reaper)
			ctx := context.Background()
			_ = sup.Start(ctx)
			defer sup.Stop()

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
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

	sup := supervisor.NewSupervisor(cfg, loader, executor, reaper)
	ctx := context.Background()
	_ = sup.Start(ctx)
	defer sup.Stop()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = sup.AllStats()
	}
}

// BenchmarkSupervisorConcurrentAccess measures concurrent read performance.
func BenchmarkSupervisorConcurrentAccess(b *testing.B) {
	cfg := createTestConfig(10)
	loader := &benchmarkLoader{cfg: cfg}
	executor := &benchmarkExecutor{}
	reaper := &benchmarkReaper{}

	sup := supervisor.NewSupervisor(cfg, loader, executor, reaper)
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
