//go:build linux

// Package proc_test provides external tests for the proc package.
package linux_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/linux"
)

// testProcStatContent contains sample /proc/stat for testing.
const testProcStatContent string = `cpu  10132153 290696 3084719 46828483 16683 0 25195 0 0 0
cpu0 5066076 145348 1542359 23414241 8341 0 12597 0 0 0
cpu1 5066077 145348 1542360 23414242 8342 0 12598 0 0 0
intr 200450538 122 9 0 0 0 0 3 0 1 79 0 0 156 0 0 0
ctxt 385016193
btime 1234567890
processes 39315
procs_running 1
procs_blocked 0
`

func TestNewCPUCollector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "creates collector with default /proc path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collector
			collector := linux.NewCPUCollector()

			// Verify collector is created and functional
			require.NotNil(t, collector)

			// Test that it can actually collect metrics
			ctx := context.Background()
			cpu, err := collector.CollectSystem(ctx)

			// Verify collection works
			require.NoError(t, err)
			assert.True(t, cpu.Total() > 0, "should collect non-zero CPU metrics")
		})
	}
}

func TestNewCPUCollectorWithPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setupFS  func(t *testing.T) string
		wantUser uint64
	}{
		{
			name: "creates collector with custom path",
			setupFS: func(t *testing.T) string {
				// Create mock /proc filesystem
				mockProc := t.TempDir()
				statContent := `cpu  12345 678 901 23456 789 0 123 0 0 0`
				require.NoError(t, os.WriteFile(filepath.Join(mockProc, "stat"), []byte(statContent), 0o644))
				// Return mock path
				return mockProc
			},
			wantUser: 12345,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup mock filesystem
			mockPath := tt.setupFS(t)

			// Create collector with custom path
			collector := linux.NewCPUCollectorWithPath(mockPath)

			// Verify collector is created and functional
			require.NotNil(t, collector)

			// Test that it can collect from custom path
			ctx := context.Background()
			cpu, err := collector.CollectSystem(ctx)

			// Verify collection works with custom path
			require.NoError(t, err)
			assert.Equal(t, tt.wantUser, cpu.User)
		})
	}
}

func TestCPUCollector_CollectSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFS   func(t *testing.T) string
		useRealFS bool
		cancelCtx bool
		wantErr   bool
		validate  func(t *testing.T, cpu metrics.SystemCPU, err error)
	}{
		{
			name:      "collects system CPU metrics from real /proc",
			useRealFS: true,
			validate: func(t *testing.T, cpu metrics.SystemCPU, err error) {
				require.NoError(t, err)
				assert.True(t, cpu.Total() > 0, "total CPU time should be positive")
				assert.NotZero(t, cpu.Timestamp, "timestamp should be set")
			},
		},
		{
			name:      "respects context cancellation",
			useRealFS: true,
			cancelCtx: true,
			wantErr:   true,
			validate: func(t *testing.T, cpu metrics.SystemCPU, err error) {
				assert.ErrorIs(t, err, context.Canceled)
			},
		},
		{
			name: "parses mock /proc/stat correctly",
			setupFS: func(t *testing.T) string {
				mockProc := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(mockProc, "stat"), []byte(testProcStatContent), 0o644))
				return mockProc
			},
			validate: func(t *testing.T, cpu metrics.SystemCPU, err error) {
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
			},
		},
		{
			name: "returns error for missing /proc/stat",
			setupFS: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, cpu metrics.SystemCPU, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "open")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var collector *linux.CPUCollector
			if tt.useRealFS {
				collector = linux.NewCPUCollector()
			} else {
				mockPath := tt.setupFS(t)
				collector = linux.NewCPUCollectorWithPath(mockPath)
			}

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			cpu, err := collector.CollectSystem(ctx)
			tt.validate(t, cpu, err)
		})
	}
}

func TestCPUCollector_CollectProcess(t *testing.T) {
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
			name:      "collects current process CPU metrics",
			useRealFS: true,
			pid:       os.Getpid(),
			validate: func(t *testing.T, cpu metrics.ProcessCPU, err error, pid int) {
				require.NoError(t, err)
				assert.Equal(t, pid, cpu.PID)
				assert.NotEmpty(t, cpu.Name, "process name should be set")
				assert.NotZero(t, cpu.Timestamp, "timestamp should be set")
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
		{
			name: "parses mock /proc/[pid]/stat correctly",
			setupFS: func(t *testing.T) string {
				mockProc := t.TempDir()
				pidDir := filepath.Join(mockProc, "1234")
				require.NoError(t, os.Mkdir(pidDir, 0o755))
				// Format: pid (comm) state ppid pgrp session tty_nr tpgid flags minflt cminflt majflt cmajflt utime stime cutime cstime ...
				statContent := `1234 (test-process) S 1 1234 1234 0 -1 4194304 1000 2000 10 20 1500 500 100 50 20 0 1 0 12345 10000000 500 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 17 0 0 0 0 0 0`
				require.NoError(t, os.WriteFile(filepath.Join(pidDir, "stat"), []byte(statContent), 0o644))
				return mockProc
			},
			pid: 1234,
			validate: func(t *testing.T, cpu metrics.ProcessCPU, err error, pid int) {
				require.NoError(t, err)
				assert.Equal(t, 1234, cpu.PID)
				assert.Equal(t, "test-process", cpu.Name)
				assert.Equal(t, uint64(1500), cpu.User)
				assert.Equal(t, uint64(500), cpu.System)
				assert.Equal(t, uint64(100), cpu.ChildrenUser)
				assert.Equal(t, uint64(50), cpu.ChildrenSystem)
				assert.Equal(t, uint64(12345), cpu.StartTime)
			},
		},
		{
			name: "handles process name with parentheses",
			setupFS: func(t *testing.T) string {
				mockProc := t.TempDir()
				pidDir := filepath.Join(mockProc, "5678")
				require.NoError(t, os.Mkdir(pidDir, 0o755))
				// Process name with parentheses: "(sd-pam)"
				statContent := `5678 ((sd-pam)) S 1 5678 5678 0 -1 4194304 100 200 1 2 150 50 10 5 20 0 1 0 54321 5000000 250 18446744073709551615 0 0 0 0 0 0 0 0 0 0 0 0 17 0 0 0 0 0 0`
				require.NoError(t, os.WriteFile(filepath.Join(pidDir, "stat"), []byte(statContent), 0o644))
				return mockProc
			},
			pid: 5678,
			validate: func(t *testing.T, cpu metrics.ProcessCPU, err error, pid int) {
				require.NoError(t, err)
				assert.Equal(t, "(sd-pam)", cpu.Name)
			},
		},
		{
			name: "returns error for non-existent process",
			setupFS: func(t *testing.T) string {
				return t.TempDir()
			},
			pid:     99999,
			wantErr: true,
			validate: func(t *testing.T, cpu metrics.ProcessCPU, err error, pid int) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var collector *linux.CPUCollector
			if tt.useRealFS {
				collector = linux.NewCPUCollector()
			} else {
				mockPath := tt.setupFS(t)
				collector = linux.NewCPUCollectorWithPath(mockPath)
			}

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			cpu, err := collector.CollectProcess(ctx, tt.pid)
			tt.validate(t, cpu, err, tt.pid)
		})
	}
}

func TestCPUCollector_CollectAllProcesses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFS   func(t *testing.T) string
		useRealFS bool
		cancelCtx bool
		wantErr   bool
		validate  func(t *testing.T, processes []metrics.ProcessCPU, err error)
	}{
		{
			name:      "collects all visible processes",
			useRealFS: true,
			validate: func(t *testing.T, processes []metrics.ProcessCPU, err error) {
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
			},
		},
		{
			name:      "respects context cancellation",
			useRealFS: true,
			cancelCtx: true,
			wantErr:   true,
			validate: func(t *testing.T, processes []metrics.ProcessCPU, err error) {
				assert.ErrorIs(t, err, context.Canceled)
			},
		},
		{
			name: "skips non-process directories",
			setupFS: func(t *testing.T) string {
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
				return mockProc
			},
			validate: func(t *testing.T, processes []metrics.ProcessCPU, err error) {
				require.NoError(t, err)
				// Should only find the one valid process
				assert.Len(t, processes, 1)
				assert.Equal(t, 1, processes[0].PID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var collector *linux.CPUCollector
			if tt.useRealFS {
				collector = linux.NewCPUCollector()
			} else {
				mockPath := tt.setupFS(t)
				collector = linux.NewCPUCollectorWithPath(mockPath)
			}

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			processes, err := collector.CollectAllProcesses(ctx)
			tt.validate(t, processes, err)
		})
	}
}

func TestCPUCollector_TimestampAccuracy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "timestamp is within expected bounds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := linux.NewCPUCollector()
			ctx := context.Background()

			before := time.Now()
			cpu, err := collector.CollectSystem(ctx)
			after := time.Now()

			require.NoError(t, err)
			assert.True(t, cpu.Timestamp.After(before) || cpu.Timestamp.Equal(before))
			assert.True(t, cpu.Timestamp.Before(after) || cpu.Timestamp.Equal(after))
		})
	}
}
