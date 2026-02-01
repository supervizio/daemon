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

func TestProcessCollector_CollectFDs(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewProcessCollector()
	ctx := context.Background()

	// Test with current process
	pid := os.Getpid()
	fds, err := collector.CollectFDs(ctx, pid)
	require.NoError(t, err)

	// Verify the returned PID matches
	assert.Equal(t, pid, fds.PID)

	// Any process should have at least 3 FDs (stdin, stdout, stderr)
	assert.GreaterOrEqual(t, fds.Count, uint32(3), "process should have at least stdin/stdout/stderr")

	t.Logf("Process %d has %d open file descriptors", pid, fds.Count)
}

func TestProcessCollector_CollectFDs_InvalidPID(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewProcessCollector()
	ctx := context.Background()

	// Test with invalid PID (99999 is unlikely to exist)
	invalidPID := 99999
	_, err = collector.CollectFDs(ctx, invalidPID)

	// Should return an error for non-existent process
	assert.Error(t, err)
	t.Logf("Expected error for invalid PID: %v", err)
}

func TestProcessCollector_CollectFDs_NotInitialized(t *testing.T) {
	// Explicitly do NOT call probe.Init()
	collector := probe.NewProcessCollector()
	ctx := context.Background()

	pid := os.Getpid()
	_, err := collector.CollectFDs(ctx, pid)

	// Should return initialization error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestProcessCollector_CollectIO(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewProcessCollector()
	ctx := context.Background()

	// Test with current process
	pid := os.Getpid()

	// First collection
	io1, err := collector.CollectIO(ctx, pid)
	require.NoError(t, err)

	// Verify the returned PID matches
	assert.Equal(t, pid, io1.PID)

	t.Logf("Process %d I/O: read=%d B/s, write=%d B/s",
		pid, io1.ReadBytesPerSec, io1.WriteBytesPerSec)

	// Perform some I/O operations
	tmpFile, err := os.CreateTemp("", "probe-test-*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write some data
	data := make([]byte, 1024*1024) // 1 MB
	for i := 0; i < 10; i++ {
		_, _ = tmpFile.Write(data)
	}
	_ = tmpFile.Sync()
	_ = tmpFile.Close()

	// Wait a bit for the metrics to update
	time.Sleep(100 * time.Millisecond)

	// Second collection should show I/O activity
	io2, err := collector.CollectIO(ctx, pid)
	require.NoError(t, err)

	assert.Equal(t, pid, io2.PID)

	// Note: I/O metrics might be 0 depending on timing and OS buffering
	// Just verify the fields are accessible
	t.Logf("After I/O: read=%d B/s, write=%d B/s",
		io2.ReadBytesPerSec, io2.WriteBytesPerSec)
}

func TestProcessCollector_CollectIO_InvalidPID(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewProcessCollector()
	ctx := context.Background()

	// Test with invalid PID
	invalidPID := 99999
	_, err = collector.CollectIO(ctx, invalidPID)

	// Should return an error for non-existent process
	assert.Error(t, err)
	t.Logf("Expected error for invalid PID: %v", err)
}

func TestProcessCollector_CollectIO_NotInitialized(t *testing.T) {
	// Explicitly do NOT call probe.Init()
	collector := probe.NewProcessCollector()
	ctx := context.Background()

	pid := os.Getpid()
	_, err := collector.CollectIO(ctx, pid)

	// Should return initialization error
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestProcessCollector_CollectAll(t *testing.T) {
	err := probe.Init()
	require.NoError(t, err)
	defer probe.Shutdown()

	collector := probe.NewProcessCollector()
	ctx := context.Background()
	pid := os.Getpid()

	// Collect all metrics to ensure they work together
	cpu, err := collector.CollectCPU(ctx, pid)
	require.NoError(t, err)

	mem, err := collector.CollectMemory(ctx, pid)
	require.NoError(t, err)

	fds, err := collector.CollectFDs(ctx, pid)
	require.NoError(t, err)

	io, err := collector.CollectIO(ctx, pid)
	require.NoError(t, err)

	// Verify all metrics are for the same process
	assert.Equal(t, pid, cpu.PID)
	assert.Equal(t, pid, mem.PID)
	assert.Equal(t, pid, fds.PID)
	assert.Equal(t, pid, io.PID)

	// Log all metrics
	t.Logf("Process %d metrics:", pid)
	t.Logf("  CPU: %.2f%%", cpu.UsagePercent)
	t.Logf("  Memory: RSS=%d bytes (%.2f%%)", mem.RSS, mem.UsagePercent)
	t.Logf("  FDs: %d", fds.Count)
	t.Logf("  I/O: read=%d B/s, write=%d B/s", io.ReadBytesPerSec, io.WriteBytesPerSec)
}
