//go:build linux

// Package proc_test provides external tests for the proc package.
package linux_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/linux"
)

func TestProcessCollector_NewProcessCollector(t *testing.T) {
	t.Parallel()

	t.Run("creates collector with default proc path", func(t *testing.T) {
		t.Parallel()

		collector := linux.NewProcessCollector()
		require.NotNil(t, collector)
	})
}

func TestProcessCollector_NewProcessCollectorWithPath(t *testing.T) {
	t.Parallel()

	t.Run("creates collector with custom proc path", func(t *testing.T) {
		t.Parallel()

		mockProc := t.TempDir()
		collector := linux.NewProcessCollectorWithPath(mockProc)
		require.NotNil(t, collector)
	})
}

func TestProcessCollector_CollectCPU(t *testing.T) {
	t.Parallel()

	t.Run("collects CPU metrics for current process", func(t *testing.T) {
		t.Parallel()

		collector := linux.NewProcessCollector()
		ctx := context.Background()
		pid := os.Getpid()

		cpu, err := collector.CollectCPU(ctx, pid)
		require.NoError(t, err)

		assert.Equal(t, pid, cpu.PID)
		assert.NotEmpty(t, cpu.Name)
	})

	t.Run("returns error for invalid PID", func(t *testing.T) {
		t.Parallel()

		collector := linux.NewProcessCollector()
		ctx := context.Background()

		_, err := collector.CollectCPU(ctx, -1)
		assert.Error(t, err)
	})

	t.Run("parses mock proc stat correctly", func(t *testing.T) {
		t.Parallel()

		mockProc := t.TempDir()
		pid := 1234

		// Create mock /proc/[pid]/stat
		pidDir := filepath.Join(mockProc, "1234")
		require.NoError(t, os.MkdirAll(pidDir, 0o755))

		statContent := "1234 (test-proc) S 1 1234 1234 0 -1 4194304 100 0 0 0 50 25 10 5 20 0 1 0 12345 1000000 500 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 17 0 0 0 0 0 0"
		require.NoError(t, os.WriteFile(filepath.Join(pidDir, "stat"), []byte(statContent), 0o644))

		collector := linux.NewProcessCollectorWithPath(mockProc)
		ctx := context.Background()

		cpu, err := collector.CollectCPU(ctx, pid)
		require.NoError(t, err)

		assert.Equal(t, pid, cpu.PID)
		assert.Equal(t, "test-proc", cpu.Name)
		assert.Equal(t, uint64(50), cpu.User)
		assert.Equal(t, uint64(25), cpu.System)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		collector := linux.NewProcessCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectCPU(ctx, os.Getpid())
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestProcessCollector_CollectMemory(t *testing.T) {
	t.Parallel()

	t.Run("collects memory metrics for current process", func(t *testing.T) {
		t.Parallel()

		collector := linux.NewProcessCollector()
		ctx := context.Background()
		pid := os.Getpid()

		mem, err := collector.CollectMemory(ctx, pid)
		require.NoError(t, err)

		assert.Equal(t, pid, mem.PID)
		assert.NotEmpty(t, mem.Name)
		assert.True(t, mem.RSS > 0, "RSS should be positive for running process")
	})

	t.Run("returns error for invalid PID", func(t *testing.T) {
		t.Parallel()

		collector := linux.NewProcessCollector()
		ctx := context.Background()

		_, err := collector.CollectMemory(ctx, -1)
		assert.Error(t, err)
	})

	t.Run("parses mock proc status correctly", func(t *testing.T) {
		t.Parallel()

		mockProc := t.TempDir()
		pid := 5678

		// Create mock /proc/[pid]/status
		pidDir := filepath.Join(mockProc, "5678")
		require.NoError(t, os.MkdirAll(pidDir, 0o755))

		statusContent := `Name:	test-memory
State:	S (sleeping)
Tgid:	5678
Ngid:	0
Pid:	5678
PPid:	1
VmPeak:	10000 kB
VmSize:	8000 kB
VmRSS:	4096 kB
VmData:	2000 kB
VmStk:	136 kB
VmSwap:	100 kB
RssShmem:	50 kB
RssFile:	100 kB
`
		require.NoError(t, os.WriteFile(filepath.Join(pidDir, "status"), []byte(statusContent), 0o644))

		collector := linux.NewProcessCollectorWithPath(mockProc)
		ctx := context.Background()

		mem, err := collector.CollectMemory(ctx, pid)
		require.NoError(t, err)

		assert.Equal(t, pid, mem.PID)
		assert.Equal(t, "test-memory", mem.Name)
		assert.Equal(t, uint64(4096*1024), mem.RSS)
		assert.Equal(t, uint64(8000*1024), mem.VMS)
		assert.Equal(t, uint64(100*1024), mem.Swap)
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()

		collector := linux.NewProcessCollector()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := collector.CollectMemory(ctx, os.Getpid())
		assert.ErrorIs(t, err, context.Canceled)
	})
}
