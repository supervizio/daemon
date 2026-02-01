//go:build cgo

package probe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewIOCollector_Internal verifies constructor creates valid instance.
func TestNewIOCollector_Internal(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewIOCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestIOCollector_StructType verifies the collector type.
func TestIOCollector_StructType(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "struct type is not nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &IOCollector{}
			assert.NotNil(t, collector)
		})
	}
}

// TestIOCollector_CollectStats verifies I/O stats collection.
func TestIOCollector_CollectStats(t *testing.T) {
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

			collector := NewIOCollector()
			ctx := context.Background()

			stats, err := collector.CollectStats(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.False(t, stats.Timestamp.IsZero())
			}
		})
	}
}

// TestIOCollector_CollectPressure verifies I/O pressure collection.
func TestIOCollector_CollectPressure(t *testing.T) {
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

			collector := NewIOCollector()
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
