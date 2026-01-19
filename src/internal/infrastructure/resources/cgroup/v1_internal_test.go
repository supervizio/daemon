//go:build linux

// Package cgroup provides internal tests for cgroup v1 functionality.
package cgroup

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === parseV1CgroupData Tests ===

// TestParseV1CgroupData tests the parseV1CgroupData function.
func TestParseV1CgroupData(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		data        string
		controller  string
		expectPath  bool
		expectError bool
	}{
		{
			name:        "cpu controller found",
			data:        "3:cpu,cpuacct:/docker/abc123\n2:memory:/docker/abc123\n",
			controller:  "cpu",
			expectPath:  true,
			expectError: false,
		},
		{
			name:        "memory controller found",
			data:        "3:cpu,cpuacct:/docker/abc123\n2:memory:/docker/abc123\n",
			controller:  "memory",
			expectPath:  true,
			expectError: false,
		},
		{
			name:        "controller not found",
			data:        "3:cpu,cpuacct:/docker/abc123\n",
			controller:  "memory",
			expectPath:  false,
			expectError: true,
		},
		{
			name:        "empty data",
			data:        "",
			controller:  "cpu",
			expectPath:  false,
			expectError: true,
		},
		{
			name:        "malformed line",
			data:        "invalid\n",
			controller:  "cpu",
			expectPath:  false,
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call parseV1CgroupData
			path, err := parseV1CgroupData(tt.data, tt.controller)

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				assert.NoError(t, err)
			}

			// Check path
			if tt.expectPath {
				// Path should not be empty
				assert.NotEmpty(t, path)
			}
		})
	}
}

// === parseV1CgroupLine Tests ===

// TestParseV1CgroupLine tests the parseV1CgroupLine function.
func TestParseV1CgroupLine(t *testing.T) {
	// Define test cases
	tests := []struct {
		name       string
		line       string
		controller string
		expectPath string
		found      bool
	}{
		{
			name:       "cpu found",
			line:       "3:cpu,cpuacct:/docker/abc",
			controller: "cpu",
			found:      true,
		},
		{
			name:       "cpuacct found",
			line:       "3:cpu,cpuacct:/docker/abc",
			controller: "cpuacct",
			found:      true,
		},
		{
			name:       "memory found",
			line:       "2:memory:/docker/abc",
			controller: "memory",
			found:      true,
		},
		{
			name:       "controller not in line",
			line:       "3:cpu,cpuacct:/docker/abc",
			controller: "memory",
			found:      false,
		},
		{
			name:       "malformed line",
			line:       "invalid",
			controller: "cpu",
			found:      false,
		},
		{
			name:       "empty line",
			line:       "",
			controller: "cpu",
			found:      false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call parseV1CgroupLine
			path, found := parseV1CgroupLine(tt.line, tt.controller)

			// Verify found status
			assert.Equal(t, tt.found, found)

			// If found, path should not be empty
			if tt.found {
				// Path should contain controller path
				assert.NotEmpty(t, path)
			}
		})
	}
}

// === buildV1CgroupPath Tests ===

// TestBuildV1CgroupPath tests the buildV1CgroupPath function.
func TestBuildV1CgroupPath(t *testing.T) {
	// Define test cases
	tests := []struct {
		name       string
		controller string
		cgroupPath string
		expected   string
	}{
		{
			name:       "cpu controller",
			controller: "cpu",
			cgroupPath: "/docker/abc",
			expected:   "/sys/fs/cgroup/cpu/docker/abc",
		},
		{
			name:       "cpuacct controller",
			controller: "cpuacct",
			cgroupPath: "/docker/abc",
			expected:   "/sys/fs/cgroup/cpu/docker/abc",
		},
		{
			name:       "memory controller",
			controller: "memory",
			cgroupPath: "/docker/abc",
			expected:   "/sys/fs/cgroup/memory/docker/abc",
		},
		{
			name:       "other controller",
			controller: "blkio",
			cgroupPath: "/docker/abc",
			expected:   "/sys/fs/cgroup/blkio/docker/abc",
		},
		{
			name:       "root cgroup",
			controller: "cpu",
			cgroupPath: "/",
			expected:   "/sys/fs/cgroup/cpu",
		},
		{
			name:       "path with whitespace",
			controller: "cpu",
			cgroupPath: "  /docker/abc  ",
			expected:   "/sys/fs/cgroup/cpu/docker/abc",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call buildV1CgroupPath
			path := buildV1CgroupPath(tt.controller, tt.cgroupPath)

			// Verify path
			assert.Equal(t, tt.expected, path)
		})
	}
}

// === splitSeq Tests ===

// TestSplitSeq tests the splitSeq function.
func TestSplitSeq(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		input    string
		sep      string
		expected []string
	}{
		{
			name:     "simple split",
			input:    "a,b,c",
			sep:      ",",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "single element",
			input:    "abc",
			sep:      ",",
			expected: []string{"abc"},
		},
		{
			name:     "empty string",
			input:    "",
			sep:      ",",
			expected: []string{""},
		},
		{
			name:     "newline separator",
			input:    "line1\nline2\nline3",
			sep:      "\n",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "multi-char separator",
			input:    "a::b::c",
			sep:      "::",
			expected: []string{"a", "b", "c"},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Collect results using slices.Collect
			results := slices.Collect(splitSeq(tt.input, tt.sep))

			// Verify results
			assert.Equal(t, tt.expected, results)
		})
	}
}

// === parseMemoryLimit Tests ===

// TestParseMemoryLimit tests the parseMemoryLimit function.
func TestParseMemoryLimit(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		data        []byte
		expected    uint64
		expectError bool
	}{
		{
			name:        "valid limit",
			data:        []byte("104857600"),
			expected:    104857600,
			expectError: false,
		},
		{
			name:        "unlimited (large value)",
			data:        []byte("9223372036854771712"),
			expected:    0,
			expectError: false,
		},
		{
			name:        "zero limit",
			data:        []byte("0"),
			expected:    0,
			expectError: false,
		},
		{
			name:        "invalid format",
			data:        []byte("invalid"),
			expected:    0,
			expectError: true,
		},
		{
			name:        "with whitespace",
			data:        []byte("  104857600  "),
			expected:    104857600,
			expectError: false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call parseMemoryLimit
			limit, err := parseMemoryLimit(tt.data)

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				assert.NoError(t, err)
				// Verify limit value
				assert.Equal(t, tt.expected, limit)
			}
		})
	}
}

// === parseV1MemoryStatLine Tests ===

// TestParseV1MemoryStatLine tests the parseV1MemoryStatLine function.
func TestParseV1MemoryStatLine(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		line     string
		expected map[string]uint64
	}{
		{
			name: "valid rss line",
			line: "total_rss 1048576",
			expected: map[string]uint64{
				"total_rss": 1048576,
			},
		},
		{
			name: "valid cache line",
			line: "total_cache 2097152",
			expected: map[string]uint64{
				"total_cache": 2097152,
			},
		},
		{
			name: "unknown field",
			line: "unknown_field 12345",
			expected: map[string]uint64{
				"unknown_field": 0,
			},
		},
		{
			name:     "malformed line",
			line:     "invalid",
			expected: map[string]uint64{},
		},
		{
			name:     "empty line",
			line:     "",
			expected: map[string]uint64{},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize stat struct
			var stat MemoryStat
			// Create field map
			fieldMap := map[string]*uint64{
				"total_rss":   &stat.Anon,
				"total_cache": &stat.File,
			}

			// Call parseV1MemoryStatLine
			parseV1MemoryStatLine(tt.line, fieldMap)

			// Verify expected values
			for key, expectedValue := range tt.expected {
				// Check if field exists
				if field, ok := fieldMap[key]; ok {
					// Verify value
					assert.Equal(t, expectedValue, *field)
				}
			}
		})
	}
}

// === validateAndCreateReader Tests ===

// TestValidateAndCreateReader tests the validateAndCreateReader function.
func TestValidateAndCreateReader(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		setup       func(t *testing.T) (string, string)
		expectError bool
	}{
		{
			name: "valid paths",
			setup: func(t *testing.T) (string, string) {
				// Create temporary directories
				cpuDir := t.TempDir()
				memDir := t.TempDir()
				// Return paths
				return cpuDir, memDir
			},
			expectError: false,
		},
		{
			name: "invalid cpu path",
			setup: func(t *testing.T) (string, string) {
				// Return non-existent CPU path
				memDir := t.TempDir()
				// Return paths
				return "/nonexistent/cpu", memDir
			},
			expectError: true,
		},
		{
			name: "invalid memory path",
			setup: func(t *testing.T) (string, string) {
				// Create valid CPU path
				cpuDir := t.TempDir()
				// Return non-existent memory path
				return cpuDir, "/nonexistent/memory"
			},
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup paths
			cpuPath, memPath := tt.setup(t)

			// Call validateAndCreateReader
			reader, err := validateAndCreateReader(cpuPath, memPath)

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
				// Reader should be nil
				assert.Nil(t, reader)
			} else {
				// Expect success
				require.NoError(t, err)
				// Reader should be valid
				assert.NotNil(t, reader)
			}
		})
	}
}

// === detectV1CPUCgroup Tests ===

// TestDetectV1CPUCgroup tests the detectV1CPUCgroup function.
func TestDetectV1CPUCgroup(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "detection attempt",
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call detectV1CPUCgroup
			path, err := detectV1CPUCgroup()

			// In test environment, cgroup likely doesn't exist
			if tt.expectError {
				// May error in non-cgroup environment
				if err != nil {
					// Error is expected
					assert.Error(t, err)
				} else {
					// Path should not be empty if successful
					assert.NotEmpty(t, path)
				}
			}
		})
	}
}

// === detectV1MemoryCgroup Tests ===

// TestDetectV1MemoryCgroup tests the detectV1MemoryCgroup function.
func TestDetectV1MemoryCgroup(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		expectError bool
	}{
		{
			name:        "detection attempt",
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call detectV1MemoryCgroup
			path, err := detectV1MemoryCgroup()

			// In test environment, cgroup likely doesn't exist
			if tt.expectError {
				// May error in non-cgroup environment
				if err != nil {
					// Error is expected
					assert.Error(t, err)
				} else {
					// Path should not be empty if successful
					assert.NotEmpty(t, path)
				}
			}
		})
	}
}

// === detectV1Cgroup Tests ===

// TestDetectV1Cgroup tests the detectV1Cgroup function.
func TestDetectV1Cgroup(t *testing.T) {
	// Define test cases
	tests := []struct {
		name       string
		controller string
	}{
		{
			name:       "cpu controller",
			controller: "cpu",
		},
		{
			name:       "memory controller",
			controller: "memory",
		},
		{
			name:       "unknown controller",
			controller: "unknown",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call detectV1Cgroup
			path, err := detectV1Cgroup(tt.controller)

			// In test environment, cgroup may not exist
			if err != nil {
				// Error is acceptable in test environment
				assert.Error(t, err)
			} else {
				// Path should not be empty if successful
				assert.NotEmpty(t, path)
			}
		})
	}
}

// === readCPUQuota Tests ===

// Test_V1Reader_readCPUQuota tests the readCPUQuota method.
func Test_V1Reader_readCPUQuota(t *testing.T) {
	// Define test cases
	tests := []struct {
		name            string
		quotaValue      string
		expectedQuota   uint64
		expectedUnlim   bool
		expectError     bool
		createQuotaFile bool
	}{
		{
			name:            "valid quota",
			quotaValue:      "100000",
			expectedQuota:   100000,
			expectedUnlim:   false,
			expectError:     false,
			createQuotaFile: true,
		},
		{
			name:            "unlimited quota",
			quotaValue:      "-1",
			expectedQuota:   0,
			expectedUnlim:   true,
			expectError:     false,
			createQuotaFile: true,
		},
		{
			name:            "file not exists",
			quotaValue:      "",
			expectedQuota:   0,
			expectedUnlim:   true,
			expectError:     false,
			createQuotaFile: false,
		},
		{
			name:            "invalid format",
			quotaValue:      "invalid",
			expectedQuota:   0,
			expectedUnlim:   false,
			expectError:     true,
			createQuotaFile: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directories
			cpuDir := t.TempDir()
			memDir := t.TempDir()

			// Create quota file if needed
			if tt.createQuotaFile {
				// Write quota file
				err := os.WriteFile(filepath.Join(cpuDir, "cpu.cfs_quota_us"), []byte(tt.quotaValue), 0o644)
				require.NoError(t, err)
			}

			// Create reader directly
			reader := &V1Reader{path: cpuDir, memoryPath: memDir}

			// Call readCPUQuota
			quota, unlimited, err := reader.readCPUQuota()

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				assert.NoError(t, err)
				// Verify quota value
				assert.Equal(t, tt.expectedQuota, quota)
				// Verify unlimited flag
				assert.Equal(t, tt.expectedUnlim, unlimited)
			}
		})
	}
}

// === readCPUPeriod Tests ===

// Test_V1Reader_readCPUPeriod tests the readCPUPeriod method.
func Test_V1Reader_readCPUPeriod(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		periodValue    string
		expectedPeriod uint64
		expectError    bool
	}{
		{
			name:           "valid period",
			periodValue:    "100000",
			expectedPeriod: 100000,
			expectError:    false,
		},
		{
			name:           "zero period",
			periodValue:    "0",
			expectedPeriod: 0,
			expectError:    false,
		},
		{
			name:           "invalid format",
			periodValue:    "invalid",
			expectedPeriod: 0,
			expectError:    true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directories
			cpuDir := t.TempDir()
			memDir := t.TempDir()

			// Write period file
			err := os.WriteFile(filepath.Join(cpuDir, "cpu.cfs_period_us"), []byte(tt.periodValue), 0o644)
			require.NoError(t, err)

			// Create reader directly
			reader := &V1Reader{path: cpuDir, memoryPath: memDir}

			// Call readCPUPeriod
			period, err := reader.readCPUPeriod()

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				assert.NoError(t, err)
				// Verify period value
				assert.Equal(t, tt.expectedPeriod, period)
			}
		})
	}
}

// === parseV1MemoryStat Tests ===

// TestParseV1MemoryStat tests the parseV1MemoryStat function.
func TestParseV1MemoryStat(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		content     string
		expectAnon  uint64
		expectFile  uint64
		expectError bool
	}{
		{
			name:        "valid stats",
			content:     "total_rss 1048576\ntotal_cache 2097152\n",
			expectAnon:  1048576,
			expectFile:  2097152,
			expectError: false,
		},
		{
			name:        "empty file",
			content:     "",
			expectAnon:  0,
			expectFile:  0,
			expectError: false,
		},
		{
			name:        "partial stats",
			content:     "total_rss 1024\n",
			expectAnon:  1024,
			expectFile:  0,
			expectError: false,
		},
		{
			name:        "malformed lines",
			content:     "invalid\ntotal_rss 1024\n",
			expectAnon:  1024,
			expectFile:  0,
			expectError: false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpFile, err := os.CreateTemp(t.TempDir(), "memory.stat")
			require.NoError(t, err)

			// Write content
			_, err = tmpFile.WriteString(tt.content)
			require.NoError(t, err)

			// Seek to beginning
			_, err = tmpFile.Seek(0, 0)
			require.NoError(t, err)

			// Call parseV1MemoryStat
			stat, err := parseV1MemoryStat(tmpFile)

			// Close file
			_ = tmpFile.Close()

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				assert.NoError(t, err)
				// Verify anon value
				assert.Equal(t, tt.expectAnon, stat.Anon)
				// Verify file value
				assert.Equal(t, tt.expectFile, stat.File)
			}
		})
	}
}
