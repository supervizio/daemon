// Package scratch provides internal tests for all scratch collectors.
package scratch

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_NewCPUCollector verifies the CPU collector constructor.
//
// Params:
//   - t: testing instance
func Test_NewCPUCollector(t *testing.T) {
	t.Parallel()

	collector := NewCPUCollector()

	require.NotNil(t, collector, "NewCPUCollector should return non-nil collector")
}

// Test_CPUCollector_CollectSystem verifies system CPU collection.
//
// Params:
//   - t: testing instance
func Test_CPUCollector_CollectSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ctx       context.Context
		wantErr   error
		checkFunc func(*testing.T, error)
	}{
		{
			name:    "returns ErrNotSupported with valid context",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			checkFunc: func(t *testing.T, err error) {
				assert.True(t, errors.Is(err, context.Canceled), "should return context.Canceled")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewCPUCollector()
			_, err := collector.CollectSystem(tt.ctx)

			if tt.checkFunc != nil {
				tt.checkFunc(t, err)
			} else {
				assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

// Test_CPUCollector_CollectProcess verifies process CPU collection.
//
// Params:
//   - t: testing instance
func Test_CPUCollector_CollectProcess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pid     int
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported with valid PID",
			pid:     1,
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name:    "returns ErrInvalidPID with zero PID",
			pid:     0,
			ctx:     context.Background(),
			wantErr: ErrInvalidPID,
		},
		{
			name:    "returns ErrInvalidPID with negative PID",
			pid:     -1,
			ctx:     context.Background(),
			wantErr: ErrInvalidPID,
		},
		{
			name: "returns context error when cancelled",
			pid:  1,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewCPUCollector()
			result, err := collector.CollectProcess(tt.ctx, tt.pid)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
			if tt.wantErr == ErrNotSupported {
				assert.Equal(t, tt.pid, result.PID, "PID should be set in result")
			}
		})
	}
}

// Test_CPUCollector_CollectAllProcesses verifies all processes CPU collection.
//
// Params:
//   - t: testing instance
func Test_CPUCollector_CollectAllProcesses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewCPUCollector()
			_, err := collector.CollectAllProcesses(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_CPUCollector_CollectLoadAverage verifies load average collection.
//
// Params:
//   - t: testing instance
func Test_CPUCollector_CollectLoadAverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewCPUCollector()
			_, err := collector.CollectLoadAverage(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_CPUCollector_CollectPressure verifies CPU pressure collection.
//
// Params:
//   - t: testing instance
func Test_CPUCollector_CollectPressure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewCPUCollector()
			_, err := collector.CollectPressure(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_NewMemoryCollector verifies the memory collector constructor.
//
// Params:
//   - t: testing instance
func Test_NewMemoryCollector(t *testing.T) {
	t.Parallel()

	collector := NewMemoryCollector()

	require.NotNil(t, collector, "NewMemoryCollector should return non-nil collector")
}

// Test_MemoryCollector_CollectSystem verifies system memory collection.
//
// Params:
//   - t: testing instance
func Test_MemoryCollector_CollectSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewMemoryCollector()
			_, err := collector.CollectSystem(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_MemoryCollector_CollectProcess verifies process memory collection.
//
// Params:
//   - t: testing instance
func Test_MemoryCollector_CollectProcess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pid     int
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported with valid PID",
			pid:     1,
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name:    "returns ErrInvalidPID with zero PID",
			pid:     0,
			ctx:     context.Background(),
			wantErr: ErrInvalidPID,
		},
		{
			name:    "returns ErrInvalidPID with negative PID",
			pid:     -1,
			ctx:     context.Background(),
			wantErr: ErrInvalidPID,
		},
		{
			name: "returns context error when cancelled",
			pid:  1,
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewMemoryCollector()
			result, err := collector.CollectProcess(tt.ctx, tt.pid)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
			if tt.wantErr == ErrNotSupported {
				assert.Equal(t, tt.pid, result.PID, "PID should be set in result")
			}
		})
	}
}

// Test_MemoryCollector_CollectAllProcesses verifies all processes memory collection.
//
// Params:
//   - t: testing instance
func Test_MemoryCollector_CollectAllProcesses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewMemoryCollector()
			_, err := collector.CollectAllProcesses(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_MemoryCollector_CollectPressure verifies memory pressure collection.
//
// Params:
//   - t: testing instance
func Test_MemoryCollector_CollectPressure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewMemoryCollector()
			_, err := collector.CollectPressure(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_NewDiskCollector verifies the disk collector constructor.
//
// Params:
//   - t: testing instance
func Test_NewDiskCollector(t *testing.T) {
	t.Parallel()

	collector := NewDiskCollector()

	require.NotNil(t, collector, "NewDiskCollector should return non-nil collector")
}

// Test_DiskCollector_ListPartitions verifies partition listing.
//
// Params:
//   - t: testing instance
func Test_DiskCollector_ListPartitions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewDiskCollector()
			_, err := collector.ListPartitions(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_DiskCollector_CollectUsage verifies disk usage collection.
//
// Params:
//   - t: testing instance
func Test_DiskCollector_CollectUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported with valid path",
			path:    "/",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name:    "returns ErrEmptyPath with empty path",
			path:    "",
			ctx:     context.Background(),
			wantErr: ErrEmptyPath,
		},
		{
			name: "returns context error when cancelled",
			path: "/",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewDiskCollector()
			result, err := collector.CollectUsage(tt.ctx, tt.path)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
			if tt.wantErr == ErrNotSupported {
				assert.Equal(t, tt.path, result.Path, "Path should be set in result")
			}
		})
	}
}

// Test_DiskCollector_CollectAllUsage verifies all disk usage collection.
//
// Params:
//   - t: testing instance
func Test_DiskCollector_CollectAllUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewDiskCollector()
			_, err := collector.CollectAllUsage(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_DiskCollector_CollectIO verifies disk I/O collection.
//
// Params:
//   - t: testing instance
func Test_DiskCollector_CollectIO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewDiskCollector()
			_, err := collector.CollectIO(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_DiskCollector_CollectDeviceIO verifies device I/O collection.
//
// Params:
//   - t: testing instance
func Test_DiskCollector_CollectDeviceIO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		device  string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported with valid device",
			device:  "sda",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name:    "returns ErrEmptyDevice with empty device",
			device:  "",
			ctx:     context.Background(),
			wantErr: ErrEmptyDevice,
		},
		{
			name:   "returns context error when cancelled",
			device: "sda",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewDiskCollector()
			result, err := collector.CollectDeviceIO(tt.ctx, tt.device)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
			if tt.wantErr == ErrNotSupported {
				assert.Equal(t, tt.device, result.Device, "Device should be set in result")
			}
		})
	}
}

// Test_NewNetworkCollector verifies the network collector constructor.
//
// Params:
//   - t: testing instance
func Test_NewNetworkCollector(t *testing.T) {
	t.Parallel()

	collector := NewNetworkCollector()

	require.NotNil(t, collector, "NewNetworkCollector should return non-nil collector")
}

// Test_NetworkCollector_ListInterfaces verifies interface listing.
//
// Params:
//   - t: testing instance
func Test_NetworkCollector_ListInterfaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewNetworkCollector()
			_, err := collector.ListInterfaces(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_NetworkCollector_CollectStats verifies network stats collection.
//
// Params:
//   - t: testing instance
func Test_NetworkCollector_CollectStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		iface   string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported with valid interface",
			iface:   "eth0",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name:    "returns ErrEmptyInterface with empty interface",
			iface:   "",
			ctx:     context.Background(),
			wantErr: ErrEmptyInterface,
		},
		{
			name:  "returns context error when cancelled",
			iface: "eth0",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewNetworkCollector()
			result, err := collector.CollectStats(tt.ctx, tt.iface)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
			if tt.wantErr == ErrNotSupported {
				assert.Equal(t, tt.iface, result.Interface, "Interface should be set in result")
			}
		})
	}
}

// Test_NetworkCollector_CollectAllStats verifies all network stats collection.
//
// Params:
//   - t: testing instance
func Test_NetworkCollector_CollectAllStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewNetworkCollector()
			_, err := collector.CollectAllStats(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_NewIOCollector verifies the I/O collector constructor.
//
// Params:
//   - t: testing instance
func Test_NewIOCollector(t *testing.T) {
	t.Parallel()

	collector := NewIOCollector()

	require.NotNil(t, collector, "NewIOCollector should return non-nil collector")
}

// Test_IOCollector_CollectStats verifies I/O stats collection.
//
// Params:
//   - t: testing instance
func Test_IOCollector_CollectStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewIOCollector()
			_, err := collector.CollectStats(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_IOCollector_CollectPressure verifies I/O pressure collection.
//
// Params:
//   - t: testing instance
func Test_IOCollector_CollectPressure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr error
	}{
		{
			name:    "returns ErrNotSupported",
			ctx:     context.Background(),
			wantErr: ErrNotSupported,
		},
		{
			name: "returns context error when cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewIOCollector()
			_, err := collector.CollectPressure(tt.ctx)

			assert.True(t, errors.Is(err, tt.wantErr), "expected %v, got %v", tt.wantErr, err)
		})
	}
}

// Test_NewProbe verifies the probe constructor.
//
// Params:
//   - t: testing instance
func Test_NewProbe(t *testing.T) {
	t.Parallel()

	probe := NewProbe()

	require.NotNil(t, probe, "NewProbe should return non-nil probe")
	assert.NotNil(t, probe.CPU(), "CPU collector should not be nil")
	assert.NotNil(t, probe.Memory(), "Memory collector should not be nil")
	assert.NotNil(t, probe.Disk(), "Disk collector should not be nil")
	assert.NotNil(t, probe.Network(), "Network collector should not be nil")
	assert.NotNil(t, probe.IO(), "IO collector should not be nil")
}
