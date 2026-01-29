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

// testMeminfoContent contains sample /proc/meminfo for testing.
const testMeminfoContent string = `MemTotal:       16384000 kB
MemFree:         4096000 kB
MemAvailable:    8192000 kB
Buffers:          512000 kB
Cached:          2048000 kB
SwapCached:            0 kB
Active:          6000000 kB
Inactive:        4000000 kB
Active(anon):    4000000 kB
Inactive(anon):   500000 kB
Active(file):    2000000 kB
Inactive(file):  3500000 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:       4096000 kB
SwapFree:        3072000 kB
Dirty:               100 kB
Writeback:             0 kB
AnonPages:       4500000 kB
Mapped:           500000 kB
Shmem:            256000 kB
`

// testPidStatusContent contains sample /proc/[pid]/status for testing.
const testPidStatusContent string = `Name:	test-daemon
Umask:	0022
State:	S (sleeping)
Tgid:	1234
Ngid:	0
Pid:	1234
PPid:	1
TracerPid:	0
Uid:	1000	1000	1000	1000
Gid:	1000	1000	1000	1000
FDSize:	256
Groups:	4 24 27 30 46 120 131 132 1000
NStgid:	1234
NSpid:	1234
NSpgid:	1234
NSsid:	1234
VmPeak:	   500000 kB
VmSize:	   450000 kB
VmLck:	        0 kB
VmPin:	        0 kB
VmHWM:	   100000 kB
VmRSS:	    80000 kB
RssAnon:	    60000 kB
RssFile:	    15000 kB
RssShmem:	     5000 kB
VmData:	   200000 kB
VmStk:	      136 kB
VmExe:	       10 kB
VmLib:	    10000 kB
VmPTE:	      500 kB
VmSwap:	     1000 kB
`

func TestMemoryCollector_CollectSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFS   func(t *testing.T) string
		useRealFS bool
		cancelCtx bool
		wantErr   bool
		validate  func(t *testing.T, mem metrics.SystemMemory, err error)
	}{
		{
			name:      "collects system memory metrics from real /proc",
			useRealFS: true,
			validate: func(t *testing.T, mem metrics.SystemMemory, err error) {
				require.NoError(t, err)
				// Basic sanity checks
				assert.True(t, mem.Total > 0, "total memory should be positive")
				assert.True(t, mem.Available > 0, "available memory should be positive")
				assert.True(t, mem.Available <= mem.Total, "available should not exceed total")
				assert.NotZero(t, mem.Timestamp, "timestamp should be set")
			},
		},
		{
			name:      "respects context cancellation",
			useRealFS: true,
			cancelCtx: true,
			wantErr:   true,
			validate: func(t *testing.T, mem metrics.SystemMemory, err error) {
				assert.ErrorIs(t, err, context.Canceled)
			},
		},
		{
			name: "parses mock /proc/meminfo correctly",
			setupFS: func(t *testing.T) string {
				mockProc := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(mockProc, "meminfo"), []byte(testMeminfoContent), 0o644))
				return mockProc
			},
			validate: func(t *testing.T, mem metrics.SystemMemory, err error) {
				require.NoError(t, err)
				// Values are in bytes (kB * 1024)
				assert.Equal(t, uint64(16384000*1024), mem.Total)
				assert.Equal(t, uint64(4096000*1024), mem.Free)
				assert.Equal(t, uint64(8192000*1024), mem.Available)
				assert.Equal(t, uint64(512000*1024), mem.Buffers)
				assert.Equal(t, uint64(2048000*1024), mem.Cached)
				assert.Equal(t, uint64(4096000*1024), mem.SwapTotal)
				assert.Equal(t, uint64(3072000*1024), mem.SwapFree)
				assert.Equal(t, uint64(256000*1024), mem.Shared)
				// Derived values
				assert.Equal(t, uint64(1024000*1024), mem.SwapUsed) // 4096000 - 3072000
				assert.Equal(t, uint64(8192000*1024), mem.Used)     // Total - Available
			},
		},
		{
			name: "calculates usage percentage correctly",
			setupFS: func(t *testing.T) string {
				mockProc := t.TempDir()
				meminfoContent := `MemTotal:       10000000 kB
MemFree:         2000000 kB
MemAvailable:    5000000 kB
`
				require.NoError(t, os.WriteFile(filepath.Join(mockProc, "meminfo"), []byte(meminfoContent), 0o644))
				return mockProc
			},
			validate: func(t *testing.T, mem metrics.SystemMemory, err error) {
				require.NoError(t, err)
				// Used = Total - Available = 5000000 kB
				// UsagePercent = 5000000 / 10000000 * 100 = 50%
				assert.InDelta(t, 50.0, mem.UsagePercent, 0.01)
			},
		},
		{
			name: "returns error for missing /proc/meminfo",
			setupFS: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: true,
			validate: func(t *testing.T, mem metrics.SystemMemory, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "open")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var collector *linux.MemoryCollector
			if tt.useRealFS {
				collector = linux.NewMemoryCollector()
			} else {
				mockPath := tt.setupFS(t)
				collector = linux.NewMemoryCollectorWithPath(mockPath)
			}

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			mem, err := collector.CollectSystem(ctx)
			tt.validate(t, mem, err)
		})
	}
}

func TestMemoryCollector_CollectProcess(t *testing.T) {
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
			name:      "collects current process memory metrics",
			useRealFS: true,
			pid:       os.Getpid(),
			validate: func(t *testing.T, mem metrics.ProcessMemory, err error, pid int) {
				require.NoError(t, err)
				assert.Equal(t, pid, mem.PID)
				assert.NotEmpty(t, mem.Name, "process name should be set")
				assert.True(t, mem.RSS > 0, "RSS should be positive for running process")
				assert.True(t, mem.VMS > 0, "VMS should be positive for running process")
				assert.NotZero(t, mem.Timestamp, "timestamp should be set")
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
		{
			name: "parses mock /proc/[pid]/status correctly",
			setupFS: func(t *testing.T) string {
				mockProc := t.TempDir()
				pidDir := filepath.Join(mockProc, "1234")
				require.NoError(t, os.Mkdir(pidDir, 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(pidDir, "status"), []byte(testPidStatusContent), 0o644))
				return mockProc
			},
			pid: 1234,
			validate: func(t *testing.T, mem metrics.ProcessMemory, err error, pid int) {
				require.NoError(t, err)
				assert.Equal(t, 1234, mem.PID)
				assert.Equal(t, "test-daemon", mem.Name)
				assert.Equal(t, uint64(80000*1024), mem.RSS)
				assert.Equal(t, uint64(450000*1024), mem.VMS)
				assert.Equal(t, uint64(1000*1024), mem.Swap)
				assert.Equal(t, uint64(200000*1024), mem.Data)
				assert.Equal(t, uint64(136*1024), mem.Stack)
				assert.Equal(t, uint64(20000*1024), mem.Shared) // Sum of RssShmem (5000) and RssFile (15000)
			},
		},
		{
			name: "returns error for non-existent process",
			setupFS: func(t *testing.T) string {
				return t.TempDir()
			},
			pid:     99999,
			wantErr: true,
			validate: func(t *testing.T, mem metrics.ProcessMemory, err error, pid int) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var collector *linux.MemoryCollector
			if tt.useRealFS {
				collector = linux.NewMemoryCollector()
			} else {
				mockPath := tt.setupFS(t)
				collector = linux.NewMemoryCollectorWithPath(mockPath)
			}

			ctx := context.Background()
			if tt.cancelCtx {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(context.Background())
				cancel()
			}

			mem, err := collector.CollectProcess(ctx, tt.pid)
			tt.validate(t, mem, err, tt.pid)
		})
	}
}

func TestMemoryCollector_CollectAllProcesses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setupFS   func(t *testing.T) string
		useRealFS bool
		cancelCtx bool
		wantErr   bool
		validate  func(t *testing.T, processes []metrics.ProcessMemory, err error)
	}{
		{
			name:      "collects all visible processes",
			useRealFS: true,
			validate: func(t *testing.T, processes []metrics.ProcessMemory, err error) {
				require.NoError(t, err)
				// Should have at least one process
				assert.NotEmpty(t, processes)
				// Find current process in results
				pid := os.Getpid()
				var found bool
				for _, p := range processes {
					if p.PID == pid {
						found = true
						assert.True(t, p.RSS > 0, "current process should have positive RSS")
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
			validate: func(t *testing.T, processes []metrics.ProcessMemory, err error) {
				assert.ErrorIs(t, err, context.Canceled)
			},
		},
		{
			name: "calculates usage percentage from system memory",
			setupFS: func(t *testing.T) string {
				mockProc := t.TempDir()
				// System meminfo
				meminfoContent := `MemTotal:       10000000 kB
MemFree:         5000000 kB
MemAvailable:    6000000 kB
`
				require.NoError(t, os.WriteFile(filepath.Join(mockProc, "meminfo"), []byte(meminfoContent), 0o644))
				// Process 1
				pidDir1 := filepath.Join(mockProc, "1")
				require.NoError(t, os.Mkdir(pidDir1, 0o755))
				statusContent1 := `Name:	init
VmRSS:	   100000 kB
VmSize:	   200000 kB
`
				require.NoError(t, os.WriteFile(filepath.Join(pidDir1, "status"), []byte(statusContent1), 0o644))
				return mockProc
			},
			validate: func(t *testing.T, processes []metrics.ProcessMemory, err error) {
				require.NoError(t, err)
				require.Len(t, processes, 1)
				// UsagePercent = RSS / Total * 100 = 100000 / 10000000 * 100 = 1%
				assert.InDelta(t, 1.0, processes[0].UsagePercent, 0.01)
			},
		},
		{
			name: "skips non-process directories",
			setupFS: func(t *testing.T) string {
				mockProc := t.TempDir()
				// System meminfo (required for CollectAllProcesses)
				meminfoContent := `MemTotal:       10000000 kB
MemAvailable:    5000000 kB
`
				require.NoError(t, os.WriteFile(filepath.Join(mockProc, "meminfo"), []byte(meminfoContent), 0o644))
				// Valid process directory
				pidDir := filepath.Join(mockProc, "1")
				require.NoError(t, os.Mkdir(pidDir, 0o755))
				statusContent := `Name:	init
VmRSS:	   50000 kB
VmSize:	  100000 kB
`
				require.NoError(t, os.WriteFile(filepath.Join(pidDir, "status"), []byte(statusContent), 0o644))
				// Non-process directories
				require.NoError(t, os.Mkdir(filepath.Join(mockProc, "self"), 0o755))
				require.NoError(t, os.Mkdir(filepath.Join(mockProc, "sys"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(mockProc, "version"), []byte("Linux version"), 0o644))
				return mockProc
			},
			validate: func(t *testing.T, processes []metrics.ProcessMemory, err error) {
				require.NoError(t, err)
				assert.Len(t, processes, 1)
				assert.Equal(t, 1, processes[0].PID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var collector *linux.MemoryCollector
			if tt.useRealFS {
				collector = linux.NewMemoryCollector()
			} else {
				mockPath := tt.setupFS(t)
				collector = linux.NewMemoryCollectorWithPath(mockPath)
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

func TestMemoryCollector_TimestampAccuracy(t *testing.T) {
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

			collector := linux.NewMemoryCollector()
			ctx := context.Background()

			before := time.Now()
			mem, err := collector.CollectSystem(ctx)
			after := time.Now()

			require.NoError(t, err)
			assert.True(t, mem.Timestamp.After(before) || mem.Timestamp.Equal(before))
			assert.True(t, mem.Timestamp.Before(after) || mem.Timestamp.Equal(after))
		})
	}
}

func TestNewMemoryCollector(t *testing.T) {
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

			// Create collector with default constructor.
			collector := linux.NewMemoryCollector()

			// Verify collector is not nil.
			require.NotNil(t, collector, "expected non-nil collector")
		})
	}
}

func TestNewMemoryCollectorWithPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		customPath string
	}{
		{
			name:       "creates collector with custom path",
			customPath: "/tmp/mock/proc",
		},
		{
			name:       "creates collector with another custom path",
			customPath: "/var/test/proc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collector with custom path.
			collector := linux.NewMemoryCollectorWithPath(tt.customPath)

			// Verify collector is not nil.
			require.NotNil(t, collector, "expected non-nil collector")
		})
	}
}
