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

// TestOSVersion verifies OS version string retrieval.
func TestOSVersion(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns non-empty string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			version := probe.OSVersion()
			assert.NotEmpty(t, version)
			assert.NotEqual(t, "unknown", version, "OSVersion should return real data on test platform")
		})
	}
}

// TestKernelVersion verifies kernel version string retrieval.
func TestKernelVersion(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns non-empty string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			version := probe.KernelVersion()
			assert.NotEmpty(t, version)
			assert.NotEqual(t, "unknown", version, "KernelVersion should return real data on test platform")
		})
	}
}

// TestArch verifies architecture string retrieval.
func TestArch(t *testing.T) {
	tests := []struct {
		name       string
		validArchs []string
	}{
		{
			name:       "returns valid architecture",
			validArchs: []string{"x86_64", "amd64", "aarch64", "arm64", "armv7l", "i686", "i386", "riscv64"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			arch := probe.Arch()
			assert.NotEmpty(t, arch)
			assert.NotEqual(t, "unknown", arch, "Arch should return real data on test platform")
			assert.Contains(t, tt.validArchs, arch)
		})
	}
}

// TestOSVersionIdempotent verifies OSVersion returns the same value on repeated calls.
func TestOSVersionIdempotent(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns consistent value"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			first := probe.OSVersion()
			second := probe.OSVersion()
			assert.Equal(t, first, second, "OSVersion should return the same value on repeated calls")
		})
	}
}

// TestMetadataBeforeInit verifies metadata functions work even before explicit Init.
// These functions use libc::uname() directly and do not require probe_init().
func TestMetadataBeforeInit(t *testing.T) {
	tests := []struct {
		name string
		fn   func() string
	}{
		{name: "OSVersion before init", fn: probe.OSVersion},
		{name: "KernelVersion before init", fn: probe.KernelVersion},
		{name: "Arch before init", fn: probe.Arch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These should not panic or crash even without Init()
			result := tt.fn()
			assert.NotEmpty(t, result)
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
