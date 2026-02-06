//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSystemCollector verifies system collector factory function.
func TestNewSystemCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates system collector with all sub-collectors"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewSystemCollector()
			assert.NotNil(t, collector)

			// Verify all sub-collectors are accessible
			assert.NotNil(t, collector.Cpu())
			assert.NotNil(t, collector.Memory())
			assert.NotNil(t, collector.Disk())
			assert.NotNil(t, collector.Network())
			assert.NotNil(t, collector.Io())
		})
	}
}

// TestNewAppProcessCollector verifies process collector factory function.
func TestNewAppProcessCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil process collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewAppProcessCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestDetectedPlatform verifies platform detection function.
func TestDetectedPlatform(t *testing.T) {
	tests := []struct {
		name           string
		validPlatforms []string
	}{
		{
			name:           "returns valid platform",
			validPlatforms: []string{"linux", "darwin", "freebsd", "openbsd", "netbsd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			platform := probe.DetectedPlatform()
			assert.NotEmpty(t, platform)

			// Should be one of the supported platforms
			assert.Contains(t, tt.validPlatforms, platform)
		})
	}
}
