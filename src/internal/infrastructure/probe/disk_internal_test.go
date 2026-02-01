//go:build cgo

package probe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSectorSizeConstant verifies the sector size constant.
func TestSectorSizeConstant(t *testing.T) {
	tests := []struct {
		name     string
		expected uint64
	}{
		{
			name:     "sectorSize equals 512",
			expected: 512,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, sectorSize)
		})
	}
}

// TestNewDiskCollector_Internal verifies constructor creates valid instance.
func TestNewDiskCollector_Internal(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewDiskCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestDiskCollector_StructType verifies the collector type.
func TestDiskCollector_StructType(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "struct type is not nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &DiskCollector{}
			assert.NotNil(t, collector)
		})
	}
}

// TestDiskCollector_ListPartitions verifies partition listing.
func TestDiskCollector_ListPartitions(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewDiskCollector()
			ctx := context.Background()

			partitions, err := collector.ListPartitions(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// On a running system, at least the root partition should exist
				assert.NotEmpty(t, partitions)
			}
		})
	}
}

// TestDiskCollector_CollectUsage verifies disk usage collection.
func TestDiskCollector_CollectUsage(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		path        string
		expectError bool
	}{
		{
			name:        "with initialized probe root path",
			initProbe:   true,
			path:        "/",
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			path:        "/",
			expectError: true,
		},
		{
			name:        "invalid path",
			initProbe:   true,
			path:        "/nonexistent/path/12345",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewDiskCollector()
			ctx := context.Background()

			usage, err := collector.CollectUsage(ctx, tt.path)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.path, usage.Path)
				assert.Greater(t, usage.Total, uint64(0))
			}
		})
	}
}

// TestDiskCollector_CollectAllUsage verifies all disk usage collection.
func TestDiskCollector_CollectAllUsage(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewDiskCollector()
			ctx := context.Background()

			usages, err := collector.CollectAllUsage(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, usages)
			}
		})
	}
}

// TestDiskCollector_CollectIO verifies disk I/O collection.
func TestDiskCollector_CollectIO(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			collector := NewDiskCollector()
			ctx := context.Background()

			ioStats, err := collector.CollectIO(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// I/O stats may be empty on some systems
				_ = ioStats
			}
		})
	}
}

// TestCCharArrayToString verifies C char array to string conversion.
func TestCCharArrayToString(t *testing.T) {
	tests := []struct {
		name     string
		input    []C.char
		expected string
	}{
		{
			name:     "empty array",
			input:    make([]C.char, 10),
			expected: "",
		},
		{
			name:     "array with content",
			input:    []C.char{'t', 'e', 's', 't', 0, 0, 0},
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := cCharArrayToString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
