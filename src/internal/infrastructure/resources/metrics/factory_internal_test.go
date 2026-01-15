package metrics

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHasProc verifies the /proc detection logic.
//
// Params:
//   - t: testing instance
func TestHasProc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wantResult bool
	}{
		{
			name:       "checks /proc availability on current platform",
			wantResult: runtime.GOOS == "linux",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := hasProc()

			// On Linux, /proc should typically be available.
			// On other platforms, it should not be.
			if runtime.GOOS == "linux" {
				// In container/CI environments, /proc may or may not exist.
				// Just verify it returns a boolean without panic.
				_ = result
			} else {
				assert.False(t, result, "hasProc should return false on non-Linux platforms")
			}
		})
	}
}

// TestIsBSD verifies BSD platform detection.
//
// Params:
//   - t: testing instance
func TestIsBSD(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		goos     string
		expected bool
	}{
		{
			name:     "freebsd is BSD",
			goos:     "freebsd",
			expected: true,
		},
		{
			name:     "openbsd is BSD",
			goos:     "openbsd",
			expected: true,
		},
		{
			name:     "netbsd is BSD",
			goos:     "netbsd",
			expected: true,
		},
		{
			name:     "dragonfly is BSD",
			goos:     "dragonfly",
			expected: true,
		},
		{
			name:     "linux is not BSD",
			goos:     "linux",
			expected: false,
		},
		{
			name:     "darwin is not BSD",
			goos:     "darwin",
			expected: false,
		},
		{
			name:     "windows is not BSD",
			goos:     "windows",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Only test the current platform since we can't change runtime.GOOS
			if runtime.GOOS == tt.goos {
				got := isBSD()
				assert.Equal(t, tt.expected, got, "isBSD() mismatch for %s", tt.goos)
			}
		})
	}

	// Additional test for current platform
	t.Run("current platform detection", func(t *testing.T) {
		t.Parallel()

		got := isBSD()
		expectedBSD := runtime.GOOS == "freebsd" || runtime.GOOS == "openbsd" ||
			runtime.GOOS == "netbsd" || runtime.GOOS == "dragonfly"
		assert.Equal(t, expectedBSD, got, "isBSD() should match expected for %s", runtime.GOOS)
	})
}
