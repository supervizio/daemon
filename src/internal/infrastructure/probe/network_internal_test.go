//go:build cgo

package probe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewNetworkCollector_Internal verifies constructor creates valid instance.
func TestNewNetworkCollector_Internal(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := NewNetworkCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestNetworkCollector_StructType verifies the collector type.
func TestNetworkCollector_StructType(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "struct type is not nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := &NetworkCollector{}
			assert.NotNil(t, collector)
		})
	}
}

// TestNetworkCollector_ListInterfaces verifies interface listing.
func TestNetworkCollector_ListInterfaces(t *testing.T) {
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

			collector := NewNetworkCollector()
			ctx := context.Background()

			ifaces, err := collector.ListInterfaces(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// On a running system, at least loopback should exist
				assert.NotEmpty(t, ifaces)
			}
		})
	}
}

// TestNetworkCollector_CollectStats verifies network stats collection.
func TestNetworkCollector_CollectStats(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		iface       string
		expectError bool
	}{
		{
			name:        "with initialized probe loopback",
			initProbe:   true,
			iface:       "lo",
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			iface:       "lo",
			expectError: true,
		},
		{
			name:        "nonexistent interface",
			initProbe:   true,
			iface:       "nonexistent12345",
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

			collector := NewNetworkCollector()
			ctx := context.Background()

			stats, err := collector.CollectStats(ctx, tt.iface)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.iface, stats.Interface)
			}
		})
	}
}

// TestNetworkCollector_CollectAllStats verifies all network stats collection.
func TestNetworkCollector_CollectAllStats(t *testing.T) {
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

			collector := NewNetworkCollector()
			ctx := context.Background()

			stats, err := collector.CollectAllStats(ctx)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				// On a running system, at least loopback should have stats
				assert.NotEmpty(t, stats)
			}
		})
	}
}
