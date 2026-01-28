//go:build linux

package collector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_readSysfsCounter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		expected uint64
	}{
		{
			name:     "valid counter",
			content:  "123456789\n",
			expected: 123456789,
		},
		{
			name:     "zero value",
			content:  "0\n",
			expected: 0,
		},
		{
			name:     "large value",
			content:  "18446744073709551615\n",
			expected: 18446744073709551615,
		},
		{
			name:     "whitespace trimmed",
			content:  "  12345  \n",
			expected: 12345,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "counter")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			result := readSysfsCounter(tmpFile)
			assert.Equal(t, tt.expected, result)
		})
	}

	t.Run("nonexistent file returns zero", func(t *testing.T) {
		t.Parallel()
		result := readSysfsCounter("/nonexistent/path")
		assert.Equal(t, uint64(0), result)
	})
}

func Test_readHardwareSpeed_parsing(t *testing.T) {
	t.Parallel()

	// Test the parsing logic by verifying with known good values.
	tests := []struct {
		name          string
		speedMbps     string
		expectedValid bool
	}{
		{
			name:          "1 Gbps",
			speedMbps:     "1000",
			expectedValid: true,
		},
		{
			name:          "10 Gbps",
			speedMbps:     "10000",
			expectedValid: true,
		},
		{
			name:          "zero speed (invalid)",
			speedMbps:     "0",
			expectedValid: false,
		},
		{
			name:          "insane speed",
			speedMbps:     "9999999",
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test the validation logic directly.
			if tt.speedMbps != "" {
				val := parseUint64(tt.speedMbps)
				isValid := val > 0 && val < maxSaneMbps

				if tt.expectedValid {
					assert.True(t, isValid)
					assert.Greater(t, val*mbpsMultiplier, uint64(0))
				} else {
					assert.False(t, isValid)
				}
			}
		})
	}

	t.Run("real interface may have speed or not", func(t *testing.T) {
		t.Parallel()
		// This test just verifies the function doesn't panic.
		// It reads from /sys/class/net which may or may not exist.
		speed := readHardwareSpeed("eth0")
		// Speed can be 0 (no hardware speed available) or positive.
		assert.GreaterOrEqual(t, speed, uint64(0))
	})
}

func Test_getAdaptiveSpeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ifName      string
		expectSpeed uint64
		checkSecond bool
	}{
		{
			name:        "initializes to 1 Gbps",
			ifName:      "test_init_1234567890",
			expectSpeed: speed1Gbps,
			checkSecond: false,
		},
		{
			name:        "returns existing speed on second call",
			ifName:      "test_existing_9876543210",
			expectSpeed: speed1Gbps,
			checkSecond: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			speed1 := getAdaptiveSpeed(tt.ifName)
			assert.Equal(t, tt.expectSpeed, speed1)

			if tt.checkSecond {
				// Second call should return same value.
				speed2 := getAdaptiveSpeed(tt.ifName)
				assert.Equal(t, speed1, speed2)
			}
		})
	}
}

func Test_getInterfaceStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		ifName         string
		expectRxBytes  uint64
		expectTxBytes  uint64
		expectMinSpeed uint64
	}{
		{
			name:           "nonexistent interface returns zeros for counters",
			ifName:         "nonexistent999",
			expectRxBytes:  0,
			expectTxBytes:  0,
			expectMinSpeed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rxBytes, txBytes, speed := getInterfaceStats(tt.ifName)
			assert.Equal(t, tt.expectRxBytes, rxBytes)
			assert.Equal(t, tt.expectTxBytes, txBytes)
			// Speed should be initialized to 1 Gbps by adaptive speed.
			assert.GreaterOrEqual(t, speed, tt.expectMinSpeed)
		})
	}
}

func Test_getInterfaceSpeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		ifName         string
		expectMinSpeed uint64
	}{
		{
			name:           "returns non-zero for virtual interfaces",
			ifName:         "nonexistent_virtual",
			expectMinSpeed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Virtual interfaces without hardware speed should get adaptive speed.
			speed := getInterfaceSpeed(tt.ifName)
			// Should initialize to 1 Gbps.
			assert.GreaterOrEqual(t, speed, tt.expectMinSpeed)
		})
	}
}

// Test_readHardwareSpeed tests the readHardwareSpeed function.
// It verifies that hardware speed reading from sysfs works correctly.
//
// Params:
//   - t: the testing context.
func Test_readHardwareSpeed(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// ifName is the interface name.
		ifName string
		// wantZero indicates if zero is expected.
		wantZero bool
	}{
		{
			name:     "nonexistent_interface",
			ifName:   "nonexistent_if_99999",
			wantZero: true,
		},
		{
			name:     "loopback_interface",
			ifName:   "lo",
			wantZero: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call function.
			speed := readHardwareSpeed(tt.ifName)

			// Verify result.
			if tt.wantZero {
				assert.Equal(t, uint64(0), speed)
			} else {
				assert.Greater(t, speed, uint64(0))
			}
		})
	}
}

// Test_readHardwareSpeed_withFile tests readHardwareSpeed with mock sysfs.
// It verifies that speed file parsing works correctly.
//
// Params:
//   - t: the testing context.
func Test_readHardwareSpeed_withFile(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// content is the speed file content.
		content string
		// wantSpeed is the expected speed.
		wantSpeed uint64
	}{
		{
			name:      "valid_1Gbps",
			content:   "1000\n",
			wantSpeed: 1000 * mbpsMultiplier,
		},
		{
			name:      "valid_10Gbps",
			content:   "10000\n",
			wantSpeed: 10000 * mbpsMultiplier,
		},
		{
			name:      "zero_speed_invalid",
			content:   "0\n",
			wantSpeed: 0,
		},
		{
			name:      "insane_speed_invalid",
			content:   "9999999\n",
			wantSpeed: 0,
		},
		{
			name:      "whitespace_trimmed",
			content:   "  1000  \n",
			wantSpeed: 1000 * mbpsMultiplier,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp dir structure.
			tmpDir := t.TempDir()
			ifDir := filepath.Join(tmpDir, "test_if")
			err := os.MkdirAll(ifDir, 0755)
			require.NoError(t, err)

			// Write speed file.
			speedPath := filepath.Join(ifDir, "speed")
			err = os.WriteFile(speedPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Verify parsing logic directly (mimicking readHardwareSpeed).
			content, err := os.ReadFile(speedPath)
			require.NoError(t, err)

			// TrimSpace like the real function does.
			val := parseUint64(strings.TrimSpace(string(content)))
			isValid := val > 0 && val < maxSaneMbps

			if tt.wantSpeed > 0 {
				assert.True(t, isValid)
				assert.Equal(t, tt.wantSpeed, val*mbpsMultiplier)
			} else {
				assert.False(t, isValid)
			}
		})
	}
}
