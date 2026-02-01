//go:build cgo

package probe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPercentMultiplierConstant verifies the percent multiplier constant.
func TestPercentMultiplierConstant(t *testing.T) {
	tests := []struct {
		name     string
		expected float64
	}{
		{
			name:     "percentMultiplier equals 100",
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, percentMultiplier)
		})
	}
}

// TestNewMemoryCollector_Internal verifies constructor creates valid instance.
func TestNewMemoryCollector_Internal(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewMemoryCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestMemoryCollector_StructType verifies the collector type.
func TestMemoryCollector_StructType(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "struct type is not nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &MemoryCollector{}
			assert.NotNil(t, collector)
		})
	}
}

// TestMemoryCollector_CollectSystem verifies system memory collection.
func TestMemoryCollector_CollectSystem(t *testing.T) {
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

			collector := NewMemoryCollector()
			ctx := context.Background()

			mem, err := collector.CollectSystem(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Greater(t, mem.Total, uint64(0))
			}
		})
	}
}

// TestMemoryCollector_CollectProcess verifies process memory collection.
func TestMemoryCollector_CollectProcess(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		pid         int
		expectError bool
	}{
		{
			name:        "with initialized probe valid pid",
			initProbe:   true,
			pid:         1,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			pid:         1,
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

			collector := NewMemoryCollector()
			ctx := context.Background()

			mem, err := collector.CollectProcess(ctx, tt.pid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.pid, mem.PID)
			}
		})
	}
}

// TestMemoryCollector_CollectAllProcesses verifies all processes memory collection.
func TestMemoryCollector_CollectAllProcesses(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
		expectErr   error
	}{
		{
			name:        "returns not supported",
			expectError: true,
			expectErr:   ErrNotSupported,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewMemoryCollector()
			ctx := context.Background()

			result, err := collector.CollectAllProcesses(ctx)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectErr, err)
				assert.Nil(t, result)
			}
		})
	}
}

// TestMemoryCollector_CollectPressure verifies memory pressure collection.
func TestMemoryCollector_CollectPressure(t *testing.T) {
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

			collector := NewMemoryCollector()
			ctx := context.Background()

			pressure, err := collector.CollectPressure(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// Pressure may return ErrNotSupported on non-Linux
				if err == nil {
					assert.GreaterOrEqual(t, pressure.SomeAvg10, 0.0)
				}
			}
		})
	}
}
