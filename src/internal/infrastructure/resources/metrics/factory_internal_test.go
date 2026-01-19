package metrics

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHasProc verifies hasProc returns correct values based on /proc existence.
func TestHasProc(t *testing.T) {
	tests := []struct {
		name         string
		mockStatProc func() error
		expected     bool
	}{
		{
			name: "returns_true_when_proc_exists",
			mockStatProc: func() error {
				return nil
			},
			expected: true,
		},
		{
			name: "returns_false_when_proc_missing",
			mockStatProc: func() error {
				return errors.New("file not found")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original and restore after test.
			originalStatProc := statProc
			t.Cleanup(func() {
				statProc = originalStatProc
			})

			// Set mock.
			statProc = tt.mockStatProc

			// Call function and verify result.
			result := hasProc()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsBSD_AllPlatforms verifies BSD detection for all platforms.
//
// Params:
//   - t: testing instance
func TestIsBSD_AllPlatforms(t *testing.T) {
	// Save original and restore after test
	originalCurrentGOOS := currentGOOS
	t.Cleanup(func() {
		currentGOOS = originalCurrentGOOS
	})

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
		{
			name:     "plan9 is not BSD",
			goos:     "plan9",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock GOOS
			currentGOOS = func() string {
				return tt.goos
			}

			got := isBSD()
			assert.Equal(t, tt.expected, got, "isBSD() mismatch for %s", tt.goos)
		})
	}
}

// TestNewSystemCollector_AllPlatforms tests collector creation for all platforms.
//
// Params:
//   - t: testing instance
func TestNewSystemCollector_AllPlatforms(t *testing.T) {
	// Save originals and restore after test
	originalCurrentGOOS := currentGOOS
	originalStatProc := statProc
	t.Cleanup(func() {
		currentGOOS = originalCurrentGOOS
		statProc = originalStatProc
	})

	tests := []struct {
		name    string
		goos    string
		hasProc bool
	}{
		{
			name:    "linux with /proc",
			goos:    "linux",
			hasProc: true,
		},
		{
			name:    "linux without /proc",
			goos:    "linux",
			hasProc: false,
		},
		{
			name:    "freebsd",
			goos:    "freebsd",
			hasProc: false,
		},
		{
			name:    "openbsd",
			goos:    "openbsd",
			hasProc: false,
		},
		{
			name:    "netbsd",
			goos:    "netbsd",
			hasProc: false,
		},
		{
			name:    "dragonfly",
			goos:    "dragonfly",
			hasProc: false,
		},
		{
			name:    "darwin",
			goos:    "darwin",
			hasProc: false,
		},
		{
			name:    "windows",
			goos:    "windows",
			hasProc: false,
		},
		{
			name:    "plan9",
			goos:    "plan9",
			hasProc: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock platform
			currentGOOS = func() string {
				return tt.goos
			}
			if tt.hasProc {
				statProc = func() error {
					return nil
				}
			} else {
				statProc = func() error {
					return errors.New("no /proc")
				}
			}

			collector := NewSystemCollector()
			require.NotNil(t, collector)
			assert.NotNil(t, collector.CPU())
			assert.NotNil(t, collector.Memory())
			assert.NotNil(t, collector.Disk())
			assert.NotNil(t, collector.Network())
			assert.NotNil(t, collector.IO())
		})
	}
}

// TestDetectedPlatform_AllPlatforms tests platform detection strings.
//
// Params:
//   - t: testing instance
func TestDetectedPlatform_AllPlatforms(t *testing.T) {
	// Save originals and restore after test
	originalCurrentGOOS := currentGOOS
	originalStatProc := statProc
	t.Cleanup(func() {
		currentGOOS = originalCurrentGOOS
		statProc = originalStatProc
	})

	tests := []struct {
		name     string
		goos     string
		hasProc  bool
		expected string
	}{
		{
			name:     "linux with /proc",
			goos:     "linux",
			hasProc:  true,
			expected: "linux-proc",
		},
		{
			name:     "linux without /proc",
			goos:     "linux",
			hasProc:  false,
			expected: "scratch-linux",
		},
		{
			name:     "freebsd",
			goos:     "freebsd",
			hasProc:  false,
			expected: "bsd-freebsd",
		},
		{
			name:     "openbsd",
			goos:     "openbsd",
			hasProc:  false,
			expected: "bsd-openbsd",
		},
		{
			name:     "netbsd",
			goos:     "netbsd",
			hasProc:  false,
			expected: "bsd-netbsd",
		},
		{
			name:     "dragonfly",
			goos:     "dragonfly",
			hasProc:  false,
			expected: "bsd-dragonfly",
		},
		{
			name:     "darwin",
			goos:     "darwin",
			hasProc:  false,
			expected: "darwin",
		},
		{
			name:     "windows",
			goos:     "windows",
			hasProc:  false,
			expected: "scratch-windows",
		},
		{
			name:     "plan9",
			goos:     "plan9",
			hasProc:  false,
			expected: "scratch-plan9",
		},
		{
			name:     "solaris",
			goos:     "solaris",
			hasProc:  false,
			expected: "scratch-solaris",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock platform
			currentGOOS = func() string {
				return tt.goos
			}
			if tt.hasProc {
				statProc = func() error {
					return nil
				}
			} else {
				statProc = func() error {
					return errors.New("no /proc")
				}
			}

			platform := DetectedPlatform()
			assert.Equal(t, tt.expected, platform, "DetectedPlatform() mismatch")
		})
	}
}

// Test_isBSD verifies isBSD returns true for BSD variants.
//
// Params:
//   - t: testing context for assertions
func Test_isBSD(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		expected bool
	}{
		{name: "freebsd_is_bsd", goos: "freebsd", expected: true},
		{name: "openbsd_is_bsd", goos: "openbsd", expected: true},
		{name: "netbsd_is_bsd", goos: "netbsd", expected: true},
		{name: "dragonfly_is_bsd", goos: "dragonfly", expected: true},
		{name: "linux_is_not_bsd", goos: "linux", expected: false},
		{name: "darwin_is_not_bsd", goos: "darwin", expected: false},
		{name: "windows_is_not_bsd", goos: "windows", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore
			oldGOOS := currentGOOS
			defer func() { currentGOOS = oldGOOS }()

			currentGOOS = func() string { return tt.goos }

			result := isBSD()
			assert.Equal(t, tt.expected, result, "isBSD() mismatch for %s", tt.goos)
		})
	}
}
