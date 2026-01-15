// Package scratch_test provides black-box tests for the scratch probe package.
package scratch_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/scratch"
)

// TestNewScratchProbe tests probe creation.
func TestNewScratchProbe(t *testing.T) {
	t.Parallel()

	probe := scratch.NewScratchProbe()
	require.NotNil(t, probe)
	assert.NotNil(t, probe.CPU())
	assert.NotNil(t, probe.Memory())
	assert.NotNil(t, probe.Disk())
	assert.NotNil(t, probe.Network())
	assert.NotNil(t, probe.IO())
}

// TestCPUCollector_ReturnsErrors tests that CPU collector returns ErrNotSupported.
func TestCPUCollector_ReturnsErrors(t *testing.T) {
	t.Parallel()

	probe := scratch.NewScratchProbe()
	cpu := probe.CPU()
	ctx := context.Background()

	t.Run("CollectSystem", func(t *testing.T) {
		t.Parallel()
		_, err := cpu.CollectSystem(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectProcess", func(t *testing.T) {
		t.Parallel()
		_, err := cpu.CollectProcess(ctx, 1234)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectAllProcesses", func(t *testing.T) {
		t.Parallel()
		_, err := cpu.CollectAllProcesses(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectLoadAverage", func(t *testing.T) {
		t.Parallel()
		_, err := cpu.CollectLoadAverage(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectPressure", func(t *testing.T) {
		t.Parallel()
		_, err := cpu.CollectPressure(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})
}

// TestMemoryCollector_ReturnsErrors tests that memory collector returns ErrNotSupported.
func TestMemoryCollector_ReturnsErrors(t *testing.T) {
	t.Parallel()

	probe := scratch.NewScratchProbe()
	mem := probe.Memory()
	ctx := context.Background()

	t.Run("CollectSystem", func(t *testing.T) {
		t.Parallel()
		_, err := mem.CollectSystem(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectProcess", func(t *testing.T) {
		t.Parallel()
		_, err := mem.CollectProcess(ctx, 1234)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectAllProcesses", func(t *testing.T) {
		t.Parallel()
		_, err := mem.CollectAllProcesses(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectPressure", func(t *testing.T) {
		t.Parallel()
		_, err := mem.CollectPressure(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})
}

// TestDiskCollector_ReturnsErrors tests that disk collector returns ErrNotSupported.
func TestDiskCollector_ReturnsErrors(t *testing.T) {
	t.Parallel()

	probe := scratch.NewScratchProbe()
	disk := probe.Disk()
	ctx := context.Background()

	t.Run("ListPartitions", func(t *testing.T) {
		t.Parallel()
		_, err := disk.ListPartitions(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectUsage", func(t *testing.T) {
		t.Parallel()
		_, err := disk.CollectUsage(ctx, "/")
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectAllUsage", func(t *testing.T) {
		t.Parallel()
		_, err := disk.CollectAllUsage(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectIO", func(t *testing.T) {
		t.Parallel()
		_, err := disk.CollectIO(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectDeviceIO", func(t *testing.T) {
		t.Parallel()
		_, err := disk.CollectDeviceIO(ctx, "sda")
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})
}

// TestNetworkCollector_ReturnsErrors tests that network collector returns ErrNotSupported.
func TestNetworkCollector_ReturnsErrors(t *testing.T) {
	t.Parallel()

	probe := scratch.NewScratchProbe()
	net := probe.Network()
	ctx := context.Background()

	t.Run("ListInterfaces", func(t *testing.T) {
		t.Parallel()
		_, err := net.ListInterfaces(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectStats", func(t *testing.T) {
		t.Parallel()
		_, err := net.CollectStats(ctx, "eth0")
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectAllStats", func(t *testing.T) {
		t.Parallel()
		_, err := net.CollectAllStats(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})
}

// TestIOCollector_ReturnsErrors tests that I/O collector returns ErrNotSupported.
func TestIOCollector_ReturnsErrors(t *testing.T) {
	t.Parallel()

	probe := scratch.NewScratchProbe()
	io := probe.IO()
	ctx := context.Background()

	t.Run("CollectStats", func(t *testing.T) {
		t.Parallel()
		_, err := io.CollectStats(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})

	t.Run("CollectPressure", func(t *testing.T) {
		t.Parallel()
		_, err := io.CollectPressure(ctx)
		assert.ErrorIs(t, err, scratch.ErrNotSupported)
	})
}
