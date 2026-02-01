//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInit verifies probe library initialization.
func TestInit(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "initializes successfully"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			assert.True(t, probe.IsInitialized())
		})
	}
}

// TestInitIdempotent verifies that Init can be called multiple times.
func TestInitIdempotent(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "second call also succeeds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			// Second call should also succeed (idempotent)
			err = probe.Init()
			require.NoError(t, err)
		})
	}
}

// TestShutdown verifies probe library shutdown.
func TestShutdown(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "shutdown sets initialized to false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)

			probe.Shutdown()
			assert.False(t, probe.IsInitialized())
		})
	}
}

// TestShutdownIdempotent verifies shutdown is safe when not initialized.
func TestShutdownIdempotent(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "does not panic when not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic even if not initialized
			probe.Shutdown()
			probe.Shutdown()
			assert.False(t, probe.IsInitialized())
		})
	}
}

// TestIsInitialized verifies initialization state checking.
func TestIsInitialized(t *testing.T) {
	tests := []struct {
		name       string
		doInit     bool
		doShutdown bool
		expected   bool
	}{
		{
			name:       "false after shutdown",
			doInit:     false,
			doShutdown: true,
			expected:   false,
		},
		{
			name:       "true after init",
			doInit:     true,
			doShutdown: false,
			expected:   true,
		},
		{
			name:       "false after init then shutdown",
			doInit:     true,
			doShutdown: true,
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Start clean
			probe.Shutdown()

			if tt.doInit {
				err := probe.Init()
				require.NoError(t, err)
			}

			if tt.doShutdown {
				probe.Shutdown()
			} else if tt.doInit {
				defer probe.Shutdown()
			}

			assert.Equal(t, tt.expected, probe.IsInitialized())
		})
	}
}

// TestPlatform verifies platform detection.
func TestPlatform(t *testing.T) {
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

			platform := probe.Platform()
			assert.NotEmpty(t, platform)
			assert.Contains(t, tt.validPlatforms, platform)
		})
	}
}

// TestQuotaSupported verifies quota support checking.
func TestQuotaSupported(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "can be called without error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			// Just verify the function can be called without error
			_ = probe.QuotaSupported()
			assert.True(t, true) // Test passes if we get here
		})
	}
}
