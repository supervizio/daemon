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

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/linux"
)

func TestNewProcessCollector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "creates collector with default proc path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := linux.NewProcessCollector()
			require.NotNil(t, collector)
		})
	}
}

func TestNewProcessCollectorWithPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "creates collector with custom proc path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockProc := t.TempDir()
			collector := linux.NewProcessCollectorWithPath(mockProc)
			require.NotNil(t, collector)
		})
	}
}

func TestProcessCollector_CollectCPU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFS   func(t *testing.T) string
		pid       int
		useRealFS bool
		cancelCtx bool
		wantErr   bool
		validate  func(t *testing.T, cpu metrics.ProcessCPU, err error, pid int)
	}{
		{
			name:      "collects CPU metrics for current process",
			useRealFS: true,
			pid:       os.Getpid(),
			validate: func(t *testing.T, cpu metrics.ProcessCPU, err error, pid int) {
				require.NoError(t, err)
				assert.Equal(t, pid, cpu.PID)
				assert.NotEmpty(t, cpu.Name)
			},
		},
		{
			name:      "returns error for invalid PID",
			useRealFS: true,
			pid:       -1,
			wantErr:   true,
			validate: func(t *testing.T, cpu metrics.ProcessCPU, err error, pid int) {
				assert.Error(t, err)
			},
		},
		{
			name: "parses mock proc stat correctly",
			setupFS: func(t *testing.T) string {
				mockProc := t.TempDir()
				pidDir := filepath.Join(mockProc, "1234")
				require.NoError(t, os.MkdirAll(pidDir, 0o755))
				statContent := "1234 (test-proc) S 1 1234 1234 0 -1 4194304 100 0 0 0 50 25 10 5 20 0 1 0 12345 1000000 500 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 17 0 0 0 0 0 0"
				require.NoError(t, os.WriteFile(filepath.Join(pidDir, "stat"), []byte(statContent), 0o644))
				return mockProc
			},
			pid: 1234,
			validate: func(t *testing.T, cpu metrics.ProcessCPU, err error, pid int) {
				require.NoError(t, err)
				assert.Equal(t, pid, cpu.PID)
				assert.Equal(t, "test-proc", cpu.Name)
				assert.Equal(t, uint64(50), cpu.User)
				assert.Equal(t, uint64(25), cpu.System)
			},
		},
		{
			name:      "respects context cancellation",
			useRealFS: true,
			pid:       os.Getpid(),
			cancelCtx: true,
			wantErr:   true,
			validate: func(t *testing.T, cpu metrics.ProcessCPU, err error, pid int) {
				assert.ErrorIs(t, err, context.Canceled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var collector *linux.ProcessCollector
			if tt.useRealFS {
				collector = linux.NewProcessCollector()
			} else {
				mockPath := tt.setupFS(t)
				collector = linux.NewProcessCollectorWithPath(mockPath)
			}

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			cpu, err := collector.CollectCPU(ctx, tt.pid)
			tt.validate(t, cpu, err, tt.pid)
		})
	}
}

func TestProcessCollector_CollectMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFS   func(t *testing.T) string
		pid       int
		useRealFS bool
		cancelCtx bool
		wantErr   bool
		validate  func(t *testing.T, mem metrics.ProcessMemory, err error, pid int)
	}{
		{
			name:      "collects memory metrics for current process",
			useRealFS: true,
			pid:       os.Getpid(),
			validate: func(t *testing.T, mem metrics.ProcessMemory, err error, pid int) {
				require.NoError(t, err)
				assert.Equal(t, pid, mem.PID)
				assert.NotEmpty(t, mem.Name)
				assert.True(t, mem.RSS > 0, "RSS should be positive for running process")
			},
		},
		{
			name:      "returns error for invalid PID",
			useRealFS: true,
			pid:       -1,
			wantErr:   true,
			validate: func(t *testing.T, mem metrics.ProcessMemory, err error, pid int) {
				assert.Error(t, err)
			},
		},
		{
			name: "parses mock proc status correctly",
			setupFS: func(t *testing.T) string {
				mockProc := t.TempDir()
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
				return mockProc
			},
			pid: 5678,
			validate: func(t *testing.T, mem metrics.ProcessMemory, err error, pid int) {
				require.NoError(t, err)
				assert.Equal(t, pid, mem.PID)
				assert.Equal(t, "test-memory", mem.Name)
				assert.Equal(t, uint64(4096*1024), mem.RSS)
				assert.Equal(t, uint64(8000*1024), mem.VMS)
				assert.Equal(t, uint64(100*1024), mem.Swap)
			},
		},
		{
			name:      "respects context cancellation",
			useRealFS: true,
			pid:       os.Getpid(),
			cancelCtx: true,
			wantErr:   true,
			validate: func(t *testing.T, mem metrics.ProcessMemory, err error, pid int) {
				assert.ErrorIs(t, err, context.Canceled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var collector *linux.ProcessCollector
			if tt.useRealFS {
				collector = linux.NewProcessCollector()
			} else {
				mockPath := tt.setupFS(t)
				collector = linux.NewProcessCollectorWithPath(mockPath)
			}

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			mem, err := collector.CollectMemory(ctx, tt.pid)
			tt.validate(t, mem, err, tt.pid)
		})
	}
}
