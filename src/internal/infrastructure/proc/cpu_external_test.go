//go:build linux

// Package proc_test provides external tests for the proc package.
package proc_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/proc"
)

func TestCPUCollector_CollectSystem(t *testing.T) {
	t.Parallel()

	t.Run("collects system CPU metrics from real /proc", func(t *testing.T) {
		t.Parallel()

		collector := proc.NewCPUCollector()
		ctx := context.Background()

		cpu, err := collector.CollectSystem(ctx)
		require.NoError(t, err)

		// Basic sanity checks - system should have some CPU time
		assert.True(t, cpu.Total() > 0, "total CPU time should be positive")
		assert.NotZero(t, cpu.Timestamp, "timestamp should be set")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		collector := proc.NewCPUCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectSystem(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("parses mock /proc/stat correctly", func(t *testing.T) {
		t.Parallel()

		mockProc := t.TempDir()
		statContent := `cpu  10132153 290696 3084719 46828483 16683 0 25195 0 0 0
cpu0 5066076 145348 1542359 23414241 8341 0 12597 0 0 0
cpu1 5066077 145348 1542360 23414242 8342 0 12598 0 0 0
intr 200450538 122 9 0 0 0 0 3 0 1 79 0 0 156 0 0 0
ctxt 385016193
btime 1234567890
processes 39315
procs_running 1
procs_blocked 0
`
		require.NoError(t, os.WriteFile(filepath.Join(mockProc, "stat"), []byte(statContent), 0o644))

		collector := proc.NewCPUCollectorWithPath(mockProc)
		ctx := context.Background()

		cpu, err := collector.CollectSystem(ctx)
		require.NoError(t, err)

		assert.Equal(t, uint64(10132153), cpu.User)
		assert.Equal(t, uint64(290696), cpu.Nice)
		assert.Equal(t, uint64(3084719), cpu.System)
		assert.Equal(t, uint64(46828483), cpu.Idle)
		assert.Equal(t, uint64(16683), cpu.IOWait)
		assert.Equal(t, uint64(0), cpu.IRQ)
		assert.Equal(t, uint64(25195), cpu.SoftIRQ)
		assert.Equal(t, uint64(0), cpu.Steal)
		assert.Equal(t, uint64(0), cpu.Guest)
		assert.Equal(t, uint64(0), cpu.GuestNice)
	})

	t.Run("returns error for missing /proc/stat", func(t *testing.T) {
		t.Parallel()

		mockProc := t.TempDir()
		collector := proc.NewCPUCollectorWithPath(mockProc)
		ctx := context.Background()

		_, err := collector.CollectSystem(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "open")
	})
}

func TestCPUCollector_CollectProcess(t *testing.T) {
	t.Parallel()

	t.Run("collects current process CPU metrics", func(t *testing.T) {
		t.Parallel()

		collector := proc.NewCPUCollector()
		ctx := context.Background()
		pid := os.Getpid()

		cpu, err := collector.CollectProcess(ctx, pid)
		require.NoError(t, err)

		assert.Equal(t, pid, cpu.PID)
		assert.NotEmpty(t, cpu.Name, "process name should be set")
		assert.NotZero(t, cpu.Timestamp, "timestamp should be set")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		collector := proc.NewCPUCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectProcess(ctx, os.Getpid())
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("parses mock /proc/[pid]/stat correctly", func(t *testing.T) {
		t.Parallel()

		mockProc := t.TempDir()
		pidDir := filepath.Join(mockProc, "1234")
		require.NoError(t, os.Mkdir(pidDir, 0o755))

		// Format: pid (comm) state ppid pgrp session tty_nr tpgid flags minflt cminflt majflt cmajflt utime stime cutime cstime ...
		statContent := `1234 (test-process) S 1 1234 1234 0 -1 4194304 1000 2000 10 20 1500 500 100 50 20 0 1 0 12345 10000000 500 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 17 0 0 0 0 0 0`
		require.NoError(t, os.WriteFile(filepath.Join(pidDir, "stat"), []byte(statContent), 0o644))

		collector := proc.NewCPUCollectorWithPath(mockProc)
		ctx := context.Background()

		cpu, err := collector.CollectProcess(ctx, 1234)
		require.NoError(t, err)

		assert.Equal(t, 1234, cpu.PID)
		assert.Equal(t, "test-process", cpu.Name)
		assert.Equal(t, uint64(1500), cpu.User)
		assert.Equal(t, uint64(500), cpu.System)
		assert.Equal(t, uint64(100), cpu.ChildrenUser)
		assert.Equal(t, uint64(50), cpu.ChildrenSystem)
		assert.Equal(t, uint64(12345), cpu.StartTime)
	})

	t.Run("handles process name with parentheses", func(t *testing.T) {
		t.Parallel()

		mockProc := t.TempDir()
		pidDir := filepath.Join(mockProc, "5678")
		require.NoError(t, os.Mkdir(pidDir, 0o755))

		// Process name with parentheses: "(sd-pam)"
		statContent := `5678 ((sd-pam)) S 1 5678 5678 0 -1 4194304 100 200 1 2 150 50 10 5 20 0 1 0 54321 5000000 250 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 17 0 0 0 0 0 0`
		require.NoError(t, os.WriteFile(filepath.Join(pidDir, "stat"), []byte(statContent), 0o644))

		collector := proc.NewCPUCollectorWithPath(mockProc)
		ctx := context.Background()

		cpu, err := collector.CollectProcess(ctx, 5678)
		require.NoError(t, err)

		assert.Equal(t, "(sd-pam)", cpu.Name)
	})

	t.Run("returns error for non-existent process", func(t *testing.T) {
		t.Parallel()

		mockProc := t.TempDir()
		collector := proc.NewCPUCollectorWithPath(mockProc)
		ctx := context.Background()

		_, err := collector.CollectProcess(ctx, 99999)
		assert.Error(t, err)
	})
}

func TestCPUCollector_CollectAllProcesses(t *testing.T) {
	t.Parallel()

	t.Run("collects all visible processes", func(t *testing.T) {
		t.Parallel()

		collector := proc.NewCPUCollector()
		ctx := context.Background()

		processes, err := collector.CollectAllProcesses(ctx)
		require.NoError(t, err)

		// Should have at least one process (this test process)
		assert.NotEmpty(t, processes)

		// Find current process in results
		pid := os.Getpid()
		var found bool
		for _, p := range processes {
			if p.PID == pid {
				found = true
				break
			}
		}
		assert.True(t, found, "current process should be in results")
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		collector := proc.NewCPUCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectAllProcesses(ctx)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("skips non-process directories", func(t *testing.T) {
		t.Parallel()

		mockProc := t.TempDir()

		// Create system stat file
		statContent := `cpu  100 50 30 1000 10 0 5 0 0 0`
		require.NoError(t, os.WriteFile(filepath.Join(mockProc, "stat"), []byte(statContent), 0o644))

		// Create a valid process directory
		pidDir := filepath.Join(mockProc, "1")
		require.NoError(t, os.Mkdir(pidDir, 0o755))
		procStatContent := `1 (init) S 0 1 1 0 -1 4194304 100 200 1 2 150 50 10 5 20 0 1 0 1 5000000 250 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 17 0 0 0 0 0 0`
		require.NoError(t, os.WriteFile(filepath.Join(pidDir, "stat"), []byte(procStatContent), 0o644))

		// Create non-process directories (should be skipped)
		require.NoError(t, os.Mkdir(filepath.Join(mockProc, "self"), 0o755))
		require.NoError(t, os.Mkdir(filepath.Join(mockProc, "sys"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(mockProc, "meminfo"), []byte(""), 0o644))

		collector := proc.NewCPUCollectorWithPath(mockProc)
		ctx := context.Background()

		processes, err := collector.CollectAllProcesses(ctx)
		require.NoError(t, err)

		// Should only find the one valid process
		assert.Len(t, processes, 1)
		assert.Equal(t, 1, processes[0].PID)
	})
}

func TestCPUCollector_TimestampAccuracy(t *testing.T) {
	t.Parallel()

	collector := proc.NewCPUCollector()
	ctx := context.Background()

	before := time.Now()
	cpu, err := collector.CollectSystem(ctx)
	after := time.Now()

	require.NoError(t, err)
	assert.True(t, cpu.Timestamp.After(before) || cpu.Timestamp.Equal(before))
	assert.True(t, cpu.Timestamp.Before(after) || cpu.Timestamp.Equal(after))
}
