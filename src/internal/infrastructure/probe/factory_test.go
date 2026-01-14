// Package probe_test provides tests for the probe factory.
package probe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
)

// TestNewSystemCollector tests the factory function.
func TestNewSystemCollector(t *testing.T) {
	t.Parallel()

	collector := probe.NewSystemCollector()
	require.NotNil(t, collector)

	// Verify all sub-collectors are available
	assert.NotNil(t, collector.CPU())
	assert.NotNil(t, collector.Memory())
	assert.NotNil(t, collector.Disk())
	assert.NotNil(t, collector.Network())
	assert.NotNil(t, collector.IO())
}

// TestDetectedPlatform tests platform detection.
func TestDetectedPlatform(t *testing.T) {
	t.Parallel()

	platform := probe.DetectedPlatform()
	assert.NotEmpty(t, platform)

	// Should be one of the known platform types
	validPrefixes := []string{"linux-proc", "bsd-", "darwin", "scratch-"}
	found := false
	for _, prefix := range validPrefixes {
		if len(platform) >= len(prefix) && platform[:len(prefix)] == prefix {
			found = true
			break
		}
	}
	assert.True(t, found, "platform %q should have a valid prefix", platform)
}
