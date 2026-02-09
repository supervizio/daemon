//go:build unix && !race

package executor_test

import (
	"context"
	"syscall"
	"testing"
	"time"

	domainprocess "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/process/executor"
)

// BenchmarkExecutorStart measures process startup overhead.
func BenchmarkExecutorStart(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	spec := domainprocess.Spec{
		Command: "/bin/true",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		pid, _, err := exec.Start(ctx, spec)
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		_ = exec.Stop(pid, time.Second)
	}
}

// BenchmarkExecutorStartLongRunning measures startup of long-running process.
func BenchmarkExecutorStartLongRunning(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	spec := domainprocess.Spec{
		Command: "/bin/sleep",
		Args:    []string{"0.01"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		pid, _, err := exec.Start(ctx, spec)
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		_ = exec.Stop(pid, time.Second)
	}
}

// BenchmarkExecutorStartWithEnv measures startup with environment variables.
func BenchmarkExecutorStartWithEnv(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	spec := domainprocess.Spec{
		Command: "/bin/true",
		Env: map[string]string{
			"FOO":       "bar",
			"BAZ":       "qux",
			"BENCHMARK": "true",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		pid, _, err := exec.Start(ctx, spec)
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		_ = exec.Stop(pid, time.Second)
	}
}

// BenchmarkExecutorStartWithWorkDir measures startup with working directory.
func BenchmarkExecutorStartWithWorkDir(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	spec := domainprocess.Spec{
		Command: "/bin/true",
		Dir:     "/tmp",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		pid, _, err := exec.Start(ctx, spec)
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		_ = exec.Stop(pid, time.Second)
	}
}

// BenchmarkExecutorStop measures process stop overhead.
func BenchmarkExecutorStop(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	pids := make([]int, b.N)
	spec := domainprocess.Spec{
		Command: "/bin/sleep",
		Args:    []string{"60"},
	}

	for i := range b.N {
		pid, _, err := exec.Start(ctx, spec)
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		pids[i] = pid
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := range b.N {
		_ = exec.Stop(pids[i], time.Second)
	}
}

// BenchmarkExecutorSignal measures signal sending overhead.
func BenchmarkExecutorSignal(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	spec := domainprocess.Spec{
		Command: "/bin/sleep",
		Args:    []string{"60"},
	}

	pid, _, err := exec.Start(ctx, spec)
	if err != nil {
		b.Fatalf("Start failed: %v", err)
	}
	defer func() {
		_ = exec.Stop(pid, time.Second)
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_ = exec.Signal(pid, syscall.SIGUSR1)
	}
}

// BenchmarkExecutorStartStop measures complete lifecycle.
func BenchmarkExecutorStartStop(b *testing.B) {
	benchmarks := []struct {
		name    string
		command string
		args    []string
	}{
		{"True", "/bin/true", nil},
		{"Echo", "/bin/echo", []string{"benchmark"}},
		{"Sleep10ms", "/bin/sleep", []string{"0.01"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			exec := executor.NewExecutor()
			ctx := context.Background()

			spec := domainprocess.Spec{
				Command: bm.command,
				Args:    bm.args,
			}

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				pid, _, err := exec.Start(ctx, spec)
				if err != nil {
					b.Fatalf("Start failed: %v", err)
				}
				_ = exec.Stop(pid, time.Second)
			}
		})
	}
}
