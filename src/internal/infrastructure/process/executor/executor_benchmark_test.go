//go:build unix && !race

package executor_test

import (
	"context"
	"testing"
	"time"

	domainprocess "github.com/kodflow/daemon/internal/domain/process"
	domainshared "github.com/kodflow/daemon/internal/domain/shared"
	"github.com/kodflow/daemon/internal/infrastructure/process/executor"
)

// BenchmarkExecutorStart measures process startup overhead.
func BenchmarkExecutorStart(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	spec := &domainprocess.Spec{
		Command: "/bin/true",
		Args:    []string{},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		state, err := exec.Start(ctx, spec)
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		// Wait for process to complete
		timeout := domainshared.Duration(time.Second)
		_ = exec.Stop(ctx, state.PID, timeout)
	}
}

// BenchmarkExecutorStartLongRunning measures startup of long-running process.
func BenchmarkExecutorStartLongRunning(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	spec := &domainprocess.Spec{
		Command: "/bin/sleep",
		Args:    []string{"0.01"}, // 10ms sleep
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		state, err := exec.Start(ctx, spec)
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		timeout := domainshared.Duration(time.Second)
		_ = exec.Stop(ctx, state.PID, timeout)
	}
}

// BenchmarkExecutorStartWithEnv measures startup with environment variables.
func BenchmarkExecutorStartWithEnv(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	spec := &domainprocess.Spec{
		Command: "/bin/true",
		Args:    []string{},
		Env: []string{
			"FOO=bar",
			"BAZ=qux",
			"BENCHMARK=true",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		state, err := exec.Start(ctx, spec)
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		timeout := domainshared.Duration(time.Second)
		_ = exec.Stop(ctx, state.PID, timeout)
	}
}

// BenchmarkExecutorStartWithWorkDir measures startup with working directory.
func BenchmarkExecutorStartWithWorkDir(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	spec := &domainprocess.Spec{
		Command: "/bin/true",
		Args:    []string{},
		WorkDir: "/tmp",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		state, err := exec.Start(ctx, spec)
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		timeout := domainshared.Duration(time.Second)
		_ = exec.Stop(ctx, state.PID, timeout)
	}
}

// BenchmarkExecutorStop measures process stop overhead.
func BenchmarkExecutorStop(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	// Pre-start N processes
	pids := make([]int, b.N)
	spec := &domainprocess.Spec{
		Command: "/bin/sleep",
		Args:    []string{"60"}, // Long-running
	}

	for i := 0; i < b.N; i++ {
		state, err := exec.Start(ctx, spec)
		if err != nil {
			b.Fatalf("Start failed: %v", err)
		}
		pids[i] = state.PID
	}

	timeout := domainshared.Duration(time.Second)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = exec.Stop(ctx, pids[i], timeout)
	}
}

// BenchmarkExecutorSignal measures signal sending overhead.
func BenchmarkExecutorSignal(b *testing.B) {
	exec := executor.NewExecutor()
	ctx := context.Background()

	spec := &domainprocess.Spec{
		Command: "/bin/sleep",
		Args:    []string{"60"},
	}

	state, err := exec.Start(ctx, spec)
	if err != nil {
		b.Fatalf("Start failed: %v", err)
	}
	defer func() {
		timeout := domainshared.Duration(time.Second)
		_ = exec.Stop(ctx, state.PID, timeout)
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = exec.Signal(state.PID, "USR1")
	}
}

// BenchmarkExecutorStartStop measures complete lifecycle.
func BenchmarkExecutorStartStop(b *testing.B) {
	benchmarks := []struct {
		name    string
		command string
		args    []string
	}{
		{"True", "/bin/true", []string{}},
		{"Echo", "/bin/echo", []string{"benchmark"}},
		{"Sleep10ms", "/bin/sleep", []string{"0.01"}},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			exec := executor.NewExecutor()
			ctx := context.Background()

			spec := &domainprocess.Spec{
				Command: bm.command,
				Args:    bm.args,
			}

			timeout := domainshared.Duration(time.Second)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				state, err := exec.Start(ctx, spec)
				if err != nil {
					b.Fatalf("Start failed: %v", err)
				}
				_ = exec.Stop(ctx, state.PID, timeout)
			}
		})
	}
}
