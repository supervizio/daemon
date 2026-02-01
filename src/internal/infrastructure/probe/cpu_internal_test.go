//go:build cgo

package probe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullPercentConstant verifies the fullPercent constant.
func TestFullPercentConstant(t *testing.T) {
	tests := []struct {
		name     string
		expected float64
	}{
		{
			name:     "fullPercent equals 100",
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, fullPercent)
		})
	}
}

// TestNewCPUCollector_Internal verifies constructor creates valid instance.
func TestNewCPUCollector_Internal(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewCPUCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestCPUCollector_StructType verifies the collector type.
func TestCPUCollector_StructType(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "struct type is not nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &CPUCollector{}
			assert.NotNil(t, collector)
		})
	}
}

// TestCollectLoadAverage verifies load average collection.
func TestCollectLoadAverage(t *testing.T) {
	tests := []struct {
		name         string
		initProbe    bool
		expectError  bool
		validateLoad bool
	}{
		{
			name:         "with initialized probe",
			initProbe:    true,
			expectError:  false,
			validateLoad: true,
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

			collector := NewCPUCollector()
			ctx := context.Background()

			load, err := collector.CollectLoadAverage(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				if tt.validateLoad {
					assert.GreaterOrEqual(t, load.Load1, 0.0)
					assert.GreaterOrEqual(t, load.Load5, 0.0)
					assert.GreaterOrEqual(t, load.Load15, 0.0)
					assert.False(t, load.Timestamp.IsZero())
				}
			}
		})
	}
}

// TestCPUCollector_CollectSystem verifies system CPU collection.
func TestCPUCollector_CollectSystem(t *testing.T) {
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

			collector := NewCPUCollector()
			ctx := context.Background()

			cpu, err := collector.CollectSystem(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, cpu.UsagePercent, 0.0)
				assert.LessOrEqual(t, cpu.UsagePercent, 100.0)
			}
		})
	}
}

// TestCPUCollector_CollectProcess verifies process CPU collection.
func TestCPUCollector_CollectProcess(t *testing.T) {
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

			collector := NewCPUCollector()
			ctx := context.Background()

			cpu, err := collector.CollectProcess(ctx, tt.pid)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.pid, cpu.PID)
			}
		})
	}
}

// TestCPUCollector_CollectAllProcesses verifies all processes CPU collection.
func TestCPUCollector_CollectAllProcesses(t *testing.T) {
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

			collector := NewCPUCollector()
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

// TestCPUCollector_CollectPressure verifies CPU pressure collection.
func TestCPUCollector_CollectPressure(t *testing.T) {
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

			collector := NewCPUCollector()
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
