//go:build cgo

package probe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCachePolicy verifies cache policy constants.
func TestCachePolicy(t *testing.T) {
	tests := []struct {
		name   string
		policy CachePolicy
	}{
		{name: "CachePolicyDefault", policy: CachePolicyDefault},
		{name: "CachePolicyHighFreq", policy: CachePolicyHighFreq},
		{name: "CachePolicyLowFreq", policy: CachePolicyLowFreq},
		{name: "CachePolicyNoCache", policy: CachePolicyNoCache},
	}

	seen := make(map[CachePolicy]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.False(t, seen[tt.policy], "duplicate policy value")
		})
		seen[tt.policy] = true
	}
}

// TestMetricType verifies metric type constants.
func TestMetricType(t *testing.T) {
	tests := []struct {
		name       string
		metricType MetricType
	}{
		{name: "MetricCPUSystem", metricType: MetricCPUSystem},
		{name: "MetricCPUPressure", metricType: MetricCPUPressure},
		{name: "MetricMemorySystem", metricType: MetricMemorySystem},
		{name: "MetricMemoryPressure", metricType: MetricMemoryPressure},
		{name: "MetricLoad", metricType: MetricLoad},
		{name: "MetricDiskPartitions", metricType: MetricDiskPartitions},
		{name: "MetricDiskUsage", metricType: MetricDiskUsage},
		{name: "MetricDiskIO", metricType: MetricDiskIO},
		{name: "MetricNetInterfaces", metricType: MetricNetInterfaces},
		{name: "MetricNetStats", metricType: MetricNetStats},
		{name: "MetricIOStats", metricType: MetricIOStats},
		{name: "MetricIOPressure", metricType: MetricIOPressure},
	}

	seen := make(map[MetricType]bool)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.False(t, seen[tt.metricType], "duplicate metric type value")
		})
		seen[tt.metricType] = true
	}
}

// TestFullPercentCacheConstant verifies the percentage constant.
func TestFullPercentCacheConstant(t *testing.T) {
	tests := []struct {
		name     string
		expected float64
	}{
		{
			name:     "fullPercentCache equals 100",
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, fullPercentCache)
		})
	}
}

// TestCacheIsEnabled verifies cache enabled state checking.
func TestCacheIsEnabled(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			// Just verify it returns a boolean without panic
			enabled := CacheIsEnabled()
			_ = enabled
		})
	}
}

// TestNewCachedCPUCollector verifies cached CPU collector construction.
func TestNewCachedCPUCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewCachedCPUCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestNewCachedMemoryCollector verifies cached memory collector construction.
func TestNewCachedMemoryCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewCachedMemoryCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestCachedCPUCollector_CollectSystem verifies cached CPU system collection.
func TestCachedCPUCollector_CollectSystem(t *testing.T) {
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

			collector := NewCachedCPUCollector()
			ctx := context.Background()

			cpu, err := collector.CollectSystem(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, cpu.UsagePercent, 0.0)
			}
		})
	}
}

// TestCachedCPUCollector_CollectAllProcesses verifies cached CPU all processes collection.
func TestCachedCPUCollector_CollectAllProcesses(t *testing.T) {
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

			collector := NewCachedCPUCollector()
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

// TestCachedMemoryCollector_CollectSystem verifies cached memory system collection.
func TestCachedMemoryCollector_CollectSystem(t *testing.T) {
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

			collector := NewCachedMemoryCollector()
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

// TestCachedMemoryCollector_CollectAllProcesses verifies cached memory all processes collection.
func TestCachedMemoryCollector_CollectAllProcesses(t *testing.T) {
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

			collector := NewCachedMemoryCollector()
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
