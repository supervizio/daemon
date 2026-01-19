// Package metrics_test provides tests for the metrics factory.
package metrics_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics"
)

// TestNewSystemCollector tests the factory function.
//
// Params:
//   - t: testing instance
func TestNewSystemCollector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		verifyCollect bool
	}{
		{
			name:          "creates non-nil collector",
			verifyCollect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := metrics.NewSystemCollector()
			require.NotNil(t, collector)

			if tt.verifyCollect {
				// Verify all sub-collectors are available
				assert.NotNil(t, collector.CPU())
				assert.NotNil(t, collector.Memory())
				assert.NotNil(t, collector.Disk())
				assert.NotNil(t, collector.Network())
				assert.NotNil(t, collector.IO())
			}
		})
	}
}

// TestDetectedPlatform tests platform detection.
//
// Params:
//   - t: testing instance
func TestDetectedPlatform(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		validPrefixes []string
	}{
		{
			name:          "returns valid platform prefix",
			validPrefixes: []string{"linux-proc", "bsd-", "darwin", "scratch-"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			platform := metrics.DetectedPlatform()
			assert.NotEmpty(t, platform)

			// Should be one of the known platform types
			found := false
			for _, prefix := range tt.validPrefixes {
				if strings.HasPrefix(platform, prefix) {
					found = true
					break
				}
			}
			assert.True(t, found, "platform %q should have a valid prefix", platform)
		})
	}
}

// TestDetectedPlatform_MatchesRuntime verifies platform detection matches runtime.GOOS.
//
// Params:
//   - t: testing instance
func TestDetectedPlatform_MatchesRuntime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		goos         string
		expectPrefix string
	}{
		{
			name:         "linux returns linux-proc or scratch-linux",
			goos:         "linux",
			expectPrefix: "linux",
		},
		{
			name:         "darwin returns darwin",
			goos:         "darwin",
			expectPrefix: "darwin",
		},
		{
			name:         "freebsd returns bsd-freebsd",
			goos:         "freebsd",
			expectPrefix: "bsd-freebsd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Only test if we're on the matching platform
			if runtime.GOOS == tt.goos {
				platform := metrics.DetectedPlatform()
				assert.Contains(t, platform, tt.expectPrefix,
					"platform %q should contain %q for %s", platform, tt.expectPrefix, tt.goos)
			}
		})
	}
}
