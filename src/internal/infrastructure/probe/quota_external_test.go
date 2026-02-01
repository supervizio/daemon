//go:build cgo

package probe_test

import (
	"os"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadQuotaLimits verifies quota limits reading.
func TestReadQuotaLimits(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "reads quota limits for current process"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			pid := os.Getpid()
			limits, err := probe.ReadQuotaLimits(pid)
			require.NoError(t, err)
			require.NotNil(t, limits)

			// Log the limits for debugging
			t.Logf("Limits: CPU=%d, Memory=%d, PIDs=%d, Nofile=%d",
				limits.CPUQuotaUS, limits.MemoryLimitBytes, limits.PIDsLimit, limits.NofileLimit)
		})
	}
}

// TestReadQuotaUsage verifies quota usage reading.
func TestReadQuotaUsage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "reads quota usage for current process"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			pid := os.Getpid()
			usage, err := probe.ReadQuotaUsage(pid)
			require.NoError(t, err)
			require.NotNil(t, usage)

			// Memory usage should be greater than zero for any process
			assert.Greater(t, usage.MemoryBytes, uint64(0))
		})
	}
}

// TestDetectContainer verifies container detection.
func TestDetectContainer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "detects container info"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			info, err := probe.DetectContainer()
			require.NoError(t, err)
			require.NotNil(t, info)

			// Log container info for debugging
			t.Logf("Container: IsContainerized=%v, Runtime=%v, ID=%s",
				info.IsContainerized, info.Runtime, info.ContainerID)
		})
	}
}

// TestQuotaLimits_HasCPULimit verifies CPU limit detection.
func TestQuotaLimits_HasCPULimit(t *testing.T) {
	tests := []struct {
		name     string
		limits   *probe.QuotaLimits
		expected bool
	}{
		{
			name: "has CPU limit",
			limits: &probe.QuotaLimits{
				Flags:       probe.QuotaFlagCPU,
				CPUQuotaUS:  100000,
				CPUPeriodUS: 100000,
			},
			expected: true,
		},
		{
			name: "no CPU flag",
			limits: &probe.QuotaLimits{
				CPUQuotaUS: 100000,
			},
			expected: false,
		},
		{
			name: "zero quota",
			limits: &probe.QuotaLimits{
				Flags:      probe.QuotaFlagCPU,
				CPUQuotaUS: 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.limits.HasCPULimit())
		})
	}
}

// TestQuotaLimits_HasMemoryLimit verifies memory limit detection.
func TestQuotaLimits_HasMemoryLimit(t *testing.T) {
	tests := []struct {
		name     string
		limits   *probe.QuotaLimits
		expected bool
	}{
		{
			name: "has memory limit",
			limits: &probe.QuotaLimits{
				Flags:            probe.QuotaFlagMemory,
				MemoryLimitBytes: 1024 * 1024 * 1024, // 1GB
			},
			expected: true,
		},
		{
			name: "no memory flag",
			limits: &probe.QuotaLimits{
				MemoryLimitBytes: 1024 * 1024 * 1024,
			},
			expected: false,
		},
		{
			name: "zero limit",
			limits: &probe.QuotaLimits{
				Flags:            probe.QuotaFlagMemory,
				MemoryLimitBytes: 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.limits.HasMemoryLimit())
		})
	}
}

// TestQuotaLimits_CPULimitPercent verifies CPU limit percentage calculation.
func TestQuotaLimits_CPULimitPercent(t *testing.T) {
	tests := []struct {
		name     string
		limits   *probe.QuotaLimits
		expected float64
	}{
		{
			name: "100% limit",
			limits: &probe.QuotaLimits{
				Flags:       probe.QuotaFlagCPU,
				CPUQuotaUS:  100000,
				CPUPeriodUS: 100000,
			},
			expected: 100.0,
		},
		{
			name: "50% limit",
			limits: &probe.QuotaLimits{
				Flags:       probe.QuotaFlagCPU,
				CPUQuotaUS:  50000,
				CPUPeriodUS: 100000,
			},
			expected: 50.0,
		},
		{
			name: "no limit",
			limits: &probe.QuotaLimits{
				CPUQuotaUS: 0,
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.InDelta(t, tt.expected, tt.limits.CPULimitPercent(), 0.01)
		})
	}
}

// TestQuotaUsage_MemoryUsagePercent verifies memory usage percentage calculation.
func TestQuotaUsage_MemoryUsagePercent(t *testing.T) {
	tests := []struct {
		name     string
		usage    *probe.QuotaUsage
		expected float64
	}{
		{
			name: "50% usage",
			usage: &probe.QuotaUsage{
				MemoryBytes:      512 * 1024 * 1024,
				MemoryLimitBytes: 1024 * 1024 * 1024,
			},
			expected: 50.0,
		},
		{
			name: "no limit",
			usage: &probe.QuotaUsage{
				MemoryBytes:      512 * 1024 * 1024,
				MemoryLimitBytes: 0,
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.InDelta(t, tt.expected, tt.usage.MemoryUsagePercent(), 0.01)
		})
	}
}

// TestContainerRuntime_String verifies container runtime string conversion.
func TestContainerRuntime_String(t *testing.T) {
	tests := []struct {
		name     string
		runtime  probe.ContainerRuntime
		expected string
	}{
		{name: "none", runtime: probe.ContainerRuntimeNone, expected: "none"},
		{name: "docker", runtime: probe.ContainerRuntimeDocker, expected: "docker"},
		{name: "podman", runtime: probe.ContainerRuntimePodman, expected: "podman"},
		{name: "lxc", runtime: probe.ContainerRuntimeLXC, expected: "lxc"},
		{name: "kubernetes", runtime: probe.ContainerRuntimeKubernetes, expected: "kubernetes"},
		{name: "jail", runtime: probe.ContainerRuntimeJail, expected: "jail"},
		{name: "unknown", runtime: probe.ContainerRuntimeUnknown, expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.runtime.String())
		})
	}
}

// TestNewQuotaLimits verifies QuotaLimits constructor.
func TestNewQuotaLimits(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil quota limits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			limits := probe.NewQuotaLimits()
			require.NotNil(t, limits)
			assert.Equal(t, uint64(0), limits.CPUQuotaUS)
		})
	}
}

// TestNewQuotaUsage verifies QuotaUsage constructor.
func TestNewQuotaUsage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil quota usage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			usage := probe.NewQuotaUsage()
			require.NotNil(t, usage)
			assert.Equal(t, uint64(0), usage.MemoryBytes)
		})
	}
}
