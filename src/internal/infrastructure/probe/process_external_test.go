//go:build cgo

package probe_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProcessCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "ReturnsNonNilCollector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewProcessCollector()
			assert.NotNil(t, collector)
		})
	}
}

func TestProcessCollector_CollectCPU(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		pid         int
		expectError bool
	}{
		{
			name:        "ValidPID",
			initProbe:   true,
			pid:         os.Getpid(),
			expectError: false,
		},
		{
			name:        "NotInitialized",
			initProbe:   false,
			pid:         os.Getpid(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := probe.Init()
				require.NoError(t, err)
				defer probe.Shutdown()
			}

			collector := probe.NewProcessCollector()
			ctx := context.Background()

			cpu, err := collector.CollectCPU(ctx, tt.pid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.pid, cpu.PID)
				t.Logf("Process %d CPU: %.2f%%", tt.pid, cpu.UsagePercent)
			}
		})
	}
}

func TestProcessCollector_CollectMemory(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		pid         int
		expectError bool
	}{
		{
			name:        "ValidPID",
			initProbe:   true,
			pid:         os.Getpid(),
			expectError: false,
		},
		{
			name:        "NotInitialized",
			initProbe:   false,
			pid:         os.Getpid(),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := probe.Init()
				require.NoError(t, err)
				defer probe.Shutdown()
			}

			collector := probe.NewProcessCollector()
			ctx := context.Background()

			mem, err := collector.CollectMemory(ctx, tt.pid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.pid, mem.PID)
				t.Logf("Process %d Memory: RSS=%d bytes (%.2f%%)", tt.pid, mem.RSS, mem.UsagePercent)
			}
		})
	}
}

func TestProcessCollector_CollectFDs(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		pid         int
		expectError bool
		errorMsg    string
		validate    func(t *testing.T, fds probe.ProcessFDs, pid int)
	}{
		{
			name:        "ValidPID",
			initProbe:   true,
			pid:         os.Getpid(),
			expectError: false,
			validate: func(t *testing.T, fds probe.ProcessFDs, pid int) {
				assert.Equal(t, pid, fds.PID)
				assert.GreaterOrEqual(t, fds.Count, uint32(3), "process should have at least stdin/stdout/stderr")
				t.Logf("Process %d has %d open file descriptors", pid, fds.Count)
			},
		},
		{
			name:        "InvalidPID",
			initProbe:   true,
			pid:         99999,
			expectError: true,
		},
		{
			name:        "NotInitialized",
			initProbe:   false,
			pid:         os.Getpid(),
			expectError: true,
			errorMsg:    "not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := probe.Init()
				require.NoError(t, err)
				defer probe.Shutdown()
			}

			collector := probe.NewProcessCollector()
			ctx := context.Background()

			fds, err := collector.CollectFDs(ctx, tt.pid)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				t.Logf("Expected error: %v", err)
			} else {
				require.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, fds, tt.pid)
				}
			}
		})
	}
}

func TestProcessCollector_CollectIO(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		pid         int
		expectError bool
		errorMsg    string
		doIOTest    bool
	}{
		{
			name:        "ValidPID",
			initProbe:   true,
			pid:         os.Getpid(),
			expectError: false,
			doIOTest:    true,
		},
		{
			name:        "InvalidPID",
			initProbe:   true,
			pid:         99999,
			expectError: true,
		},
		{
			name:        "NotInitialized",
			initProbe:   false,
			pid:         os.Getpid(),
			expectError: true,
			errorMsg:    "not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := probe.Init()
				require.NoError(t, err)
				defer probe.Shutdown()
			}

			collector := probe.NewProcessCollector()
			ctx := context.Background()

			io1, err := collector.CollectIO(ctx, tt.pid)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				t.Logf("Expected error: %v", err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.pid, io1.PID)
			t.Logf("Process %d I/O: read=%d B/s, write=%d B/s",
				tt.pid, io1.ReadBytesPerSec, io1.WriteBytesPerSec)

			if tt.doIOTest {
				// Perform some I/O operations using t.TempDir()
				tmpDir := t.TempDir()
				tmpFilePath := tmpDir + "/probe-test"
				tmpFile, err := os.Create(tmpFilePath)
				require.NoError(t, err)
				defer tmpFile.Close()

				// Write some data
				const ioTestDataSize int = 1024 * 1024 // 1 MB
				data := make([]byte, ioTestDataSize)
				// Write data multiple times to generate I/O activity.
				for range 10 {
					_, _ = tmpFile.Write(data)
				}
				_ = tmpFile.Sync()

				// Wait a bit for the metrics to update
				time.Sleep(100 * time.Millisecond)

				// Second collection should show I/O activity
				io2, err := collector.CollectIO(ctx, tt.pid)
				require.NoError(t, err)

				assert.Equal(t, tt.pid, io2.PID)

				// Note: I/O metrics might be 0 depending on timing and OS buffering
				// Just verify the fields are accessible
				t.Logf("After I/O: read=%d B/s, write=%d B/s",
					io2.ReadBytesPerSec, io2.WriteBytesPerSec)
			}
		})
	}
}

func TestProcessCollector_CollectAll(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		pid       int
		wantErr   bool
	}{
		{
			name:      "ValidPID_AllMetrics",
			initProbe: true,
			pid:       os.Getpid(),
			wantErr:   false,
		},
		{
			name:      "NotInitialized",
			initProbe: false,
			pid:       os.Getpid(),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := probe.Init()
				require.NoError(t, err)
				defer probe.Shutdown()
			}

			collector := probe.NewProcessCollector()
			ctx := context.Background()

			// Collect all metrics to ensure they work together
			cpu, err := collector.CollectCPU(ctx, tt.pid)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			mem, err := collector.CollectMemory(ctx, tt.pid)
			require.NoError(t, err)

			fds, err := collector.CollectFDs(ctx, tt.pid)
			require.NoError(t, err)

			io, err := collector.CollectIO(ctx, tt.pid)
			require.NoError(t, err)

			// Verify all metrics are for the same process
			assert.Equal(t, tt.pid, cpu.PID)
			assert.Equal(t, tt.pid, mem.PID)
			assert.Equal(t, tt.pid, fds.PID)
			assert.Equal(t, tt.pid, io.PID)

			// Log all metrics
			t.Logf("Process %d metrics:", tt.pid)
			t.Logf("  CPU: %.2f%%", cpu.UsagePercent)
			t.Logf("  Memory: RSS=%d bytes (%.2f%%)", mem.RSS, mem.UsagePercent)
			t.Logf("  FDs: %d", fds.Count)
			t.Logf("  I/O: read=%d B/s, write=%d B/s", io.ReadBytesPerSec, io.WriteBytesPerSec)
		})
	}
}
