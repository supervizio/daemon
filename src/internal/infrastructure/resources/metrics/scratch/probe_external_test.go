// Package scratch_test provides black-box tests for the scratch probe package.
package scratch_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/scratch"
)

// TestNewProbe tests probe creation.
//
// Params:
//   - t: testing instance
func TestNewProbe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		checkCPU     bool
		checkMemory  bool
		checkDisk    bool
		checkNetwork bool
		checkIO      bool
	}{
		{
			name:         "creates probe with all collectors",
			checkCPU:     true,
			checkMemory:  true,
			checkDisk:    true,
			checkNetwork: true,
			checkIO:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			probe := scratch.NewProbe()
			require.NotNil(t, probe)

			if tt.checkCPU {
				assert.NotNil(t, probe.CPU())
			}
			if tt.checkMemory {
				assert.NotNil(t, probe.Memory())
			}
			if tt.checkDisk {
				assert.NotNil(t, probe.Disk())
			}
			if tt.checkNetwork {
				assert.NotNil(t, probe.Network())
			}
			if tt.checkIO {
				assert.NotNil(t, probe.IO())
			}
		})
	}
}

// TestCPUCollector_ReturnsErrors tests that CPU collector returns ErrNotSupported.
//
// Params:
//   - t: testing instance
func TestCPUCollector_ReturnsErrors(t *testing.T) {
	t.Parallel()

	probe := scratch.NewProbe()
	cpu := probe.CPU()
	ctx := context.Background()

	tests := []struct {
		name        string
		collectFunc func() error
	}{
		{
			name: "CollectSystem returns ErrNotSupported",
			collectFunc: func() error {
				_, err := cpu.CollectSystem(ctx)
				return err
			},
		},
		{
			name: "CollectProcess returns ErrNotSupported",
			collectFunc: func() error {
				_, err := cpu.CollectProcess(ctx, 1234)
				return err
			},
		},
		{
			name: "CollectAllProcesses returns ErrNotSupported",
			collectFunc: func() error {
				_, err := cpu.CollectAllProcesses(ctx)
				return err
			},
		},
		{
			name: "CollectLoadAverage returns ErrNotSupported",
			collectFunc: func() error {
				_, err := cpu.CollectLoadAverage(ctx)
				return err
			},
		},
		{
			name: "CollectPressure returns ErrNotSupported",
			collectFunc: func() error {
				_, err := cpu.CollectPressure(ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.collectFunc()
			assert.ErrorIs(t, err, scratch.ErrNotSupported)
		})
	}
}

// TestMemoryCollector_ReturnsErrors tests that memory collector returns ErrNotSupported.
//
// Params:
//   - t: testing instance
func TestMemoryCollector_ReturnsErrors(t *testing.T) {
	t.Parallel()

	probe := scratch.NewProbe()
	mem := probe.Memory()
	ctx := context.Background()

	tests := []struct {
		name        string
		collectFunc func() error
	}{
		{
			name: "CollectSystem returns ErrNotSupported",
			collectFunc: func() error {
				_, err := mem.CollectSystem(ctx)
				return err
			},
		},
		{
			name: "CollectProcess returns ErrNotSupported",
			collectFunc: func() error {
				_, err := mem.CollectProcess(ctx, 1234)
				return err
			},
		},
		{
			name: "CollectAllProcesses returns ErrNotSupported",
			collectFunc: func() error {
				_, err := mem.CollectAllProcesses(ctx)
				return err
			},
		},
		{
			name: "CollectPressure returns ErrNotSupported",
			collectFunc: func() error {
				_, err := mem.CollectPressure(ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.collectFunc()
			assert.ErrorIs(t, err, scratch.ErrNotSupported)
		})
	}
}

// TestDiskCollector_ReturnsErrors tests that disk collector returns ErrNotSupported.
//
// Params:
//   - t: testing instance
func TestDiskCollector_ReturnsErrors(t *testing.T) {
	t.Parallel()

	probe := scratch.NewProbe()
	disk := probe.Disk()
	ctx := context.Background()

	tests := []struct {
		name        string
		collectFunc func() error
	}{
		{
			name: "ListPartitions returns ErrNotSupported",
			collectFunc: func() error {
				_, err := disk.ListPartitions(ctx)
				return err
			},
		},
		{
			name: "CollectUsage returns ErrNotSupported",
			collectFunc: func() error {
				_, err := disk.CollectUsage(ctx, "/")
				return err
			},
		},
		{
			name: "CollectAllUsage returns ErrNotSupported",
			collectFunc: func() error {
				_, err := disk.CollectAllUsage(ctx)
				return err
			},
		},
		{
			name: "CollectIO returns ErrNotSupported",
			collectFunc: func() error {
				_, err := disk.CollectIO(ctx)
				return err
			},
		},
		{
			name: "CollectDeviceIO returns ErrNotSupported",
			collectFunc: func() error {
				_, err := disk.CollectDeviceIO(ctx, "sda")
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.collectFunc()
			assert.ErrorIs(t, err, scratch.ErrNotSupported)
		})
	}
}

// TestNetworkCollector_ReturnsErrors tests that network collector returns ErrNotSupported.
//
// Params:
//   - t: testing instance
func TestNetworkCollector_ReturnsErrors(t *testing.T) {
	t.Parallel()

	probe := scratch.NewProbe()
	net := probe.Network()
	ctx := context.Background()

	tests := []struct {
		name        string
		collectFunc func() error
	}{
		{
			name: "ListInterfaces returns ErrNotSupported",
			collectFunc: func() error {
				_, err := net.ListInterfaces(ctx)
				return err
			},
		},
		{
			name: "CollectStats returns ErrNotSupported",
			collectFunc: func() error {
				_, err := net.CollectStats(ctx, "eth0")
				return err
			},
		},
		{
			name: "CollectAllStats returns ErrNotSupported",
			collectFunc: func() error {
				_, err := net.CollectAllStats(ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.collectFunc()
			assert.ErrorIs(t, err, scratch.ErrNotSupported)
		})
	}
}

// TestIOCollector_ReturnsErrors tests that I/O collector returns ErrNotSupported.
//
// Params:
//   - t: testing instance
func TestIOCollector_ReturnsErrors(t *testing.T) {
	t.Parallel()

	probe := scratch.NewProbe()
	io := probe.IO()
	ctx := context.Background()

	tests := []struct {
		name        string
		collectFunc func() error
	}{
		{
			name: "CollectStats returns ErrNotSupported",
			collectFunc: func() error {
				_, err := io.CollectStats(ctx)
				return err
			},
		},
		{
			name: "CollectPressure returns ErrNotSupported",
			collectFunc: func() error {
				_, err := io.CollectPressure(ctx)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.collectFunc()
			assert.ErrorIs(t, err, scratch.ErrNotSupported)
		})
	}
}
