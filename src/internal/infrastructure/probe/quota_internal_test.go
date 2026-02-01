//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPercentMultiplierQuotaConstant verifies the percentage multiplier constant.
func TestPercentMultiplierQuotaConstant(t *testing.T) {
	tests := []struct {
		name     string
		expected float64
	}{
		{
			name:     "percentMultiplierQuota equals 100",
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, percentMultiplierQuota)
		})
	}
}

// TestUnlimitedValueConstant verifies the unlimited value constant.
func TestUnlimitedValueConstant(t *testing.T) {
	tests := []struct {
		name     string
		expected uint64
	}{
		{
			name:     "unlimitedValue equals max uint64",
			expected: ^uint64(0),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, unlimitedValue)
		})
	}
}

// TestQuotaFlagConstants verifies quota flag constants.
func TestQuotaFlagConstants(t *testing.T) {
	tests := []struct {
		name string
		flag uint32
	}{
		{name: "QuotaFlagCPU", flag: QuotaFlagCPU},
		{name: "QuotaFlagMemory", flag: QuotaFlagMemory},
		{name: "QuotaFlagPIDs", flag: QuotaFlagPIDs},
		{name: "QuotaFlagNofile", flag: QuotaFlagNofile},
		{name: "QuotaFlagCPUTime", flag: QuotaFlagCPUTime},
		{name: "QuotaFlagData", flag: QuotaFlagData},
		{name: "QuotaFlagIORead", flag: QuotaFlagIORead},
		{name: "QuotaFlagIOWrite", flag: QuotaFlagIOWrite},
	}

	seen := make(map[uint32]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.False(t, seen[tt.flag], "duplicate flag value")
			// Verify flag is a power of 2
			assert.True(t, tt.flag > 0 && tt.flag&(tt.flag-1) == 0, "flag should be power of 2")
		})
		seen[tt.flag] = true
	}
}

// TestContainerRuntime_StringUnknown verifies unknown runtime handling.
func TestContainerRuntime_StringUnknown(t *testing.T) {
	tests := []struct {
		name     string
		runtime  ContainerRuntime
		expected string
	}{
		{
			name:     "unknown runtime returns unknown string",
			runtime:  ContainerRuntime(200),
			expected: containerRuntimeUnknownStr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.runtime.String())
		})
	}
}

// TestContainerRuntimeUnknownStrConstant verifies the unknown string constant.
func TestContainerRuntimeUnknownStrConstant(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "containerRuntimeUnknownStr equals unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, containerRuntimeUnknownStr)
		})
	}
}
