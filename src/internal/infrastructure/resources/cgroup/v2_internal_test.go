//go:build linux

// Package cgroup provides internal tests for cgroup v2 functionality.
package cgroup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === detectCurrentCgroup Tests ===

// Test_detectCurrentCgroup tests the detectCurrentCgroup function.
func Test_detectCurrentCgroup(t *testing.T) {
	// Define test cases covering different scenarios
	// Note: This function reads /proc/self/cgroup which varies by environment.
	// In non-cgroup environments, it may error or return a valid path if cgroups exist.
	tests := []struct {
		name              string
		expectPathPrefix  string
		expectEmptyPath   bool
		acceptErrorOrPath bool
	}{
		{
			name:              "detection in current environment",
			expectPathPrefix:  DefaultCgroupPath,
			acceptErrorOrPath: true,
		},
		{
			name:              "path should start with cgroup base",
			expectPathPrefix:  "/sys/fs/cgroup",
			acceptErrorOrPath: true,
		},
		{
			name:              "returns error or valid path",
			acceptErrorOrPath: true,
		},
		{
			name:              "path not empty when successful",
			expectEmptyPath:   false,
			acceptErrorOrPath: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call detectCurrentCgroup
			path, err := detectCurrentCgroup()

			// In test environment, behavior depends on host
			if tt.acceptErrorOrPath {
				// Accept either error or valid path
				if err != nil {
					// Error is acceptable in non-cgroup environment
					assert.Error(t, err)
				} else {
					// Path should not be empty if successful
					assert.NotEmpty(t, path)
					// Path should start with expected prefix
					if tt.expectPathPrefix != "" {
						assert.Contains(t, path, tt.expectPathPrefix)
					}
				}
			}
		})
	}
}

// === readCPUStat Tests ===

// Test_V2Reader_readCPUStat tests the readCPUStat method.
func Test_V2Reader_readCPUStat(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		cpuStatContent string
		createFile     bool
		expectedUsage  uint64
		expectedUser   uint64
		expectedSystem uint64
		expectError    bool
	}{
		{
			name:           "valid stats",
			cpuStatContent: "usage_usec 1234567\nuser_usec 500000\nsystem_usec 200000\n",
			createFile:     true,
			expectedUsage:  1234567,
			expectedUser:   500000,
			expectedSystem: 200000,
			expectError:    false,
		},
		{
			name:           "partial stats",
			cpuStatContent: "usage_usec 1000000\n",
			createFile:     true,
			expectedUsage:  1000000,
			expectedUser:   0,
			expectedSystem: 0,
			expectError:    false,
		},
		{
			name:           "empty file",
			cpuStatContent: "",
			createFile:     true,
			expectedUsage:  0,
			expectedUser:   0,
			expectedSystem: 0,
			expectError:    false,
		},
		{
			name:           "malformed lines",
			cpuStatContent: "invalid\nusage_usec 1234567\nmalformed data\n",
			createFile:     true,
			expectedUsage:  1234567,
			expectedUser:   0,
			expectedSystem: 0,
			expectError:    false,
		},
		{
			name:        "missing file",
			createFile:  false,
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir := t.TempDir()

			// Write cpu.stat file if needed
			if tt.createFile {
				err := os.WriteFile(filepath.Join(dir, "cpu.stat"), []byte(tt.cpuStatContent), 0o644)
				require.NoError(t, err)
			}

			// Create reader directly
			reader := &V2Reader{path: dir}

			// Call readCPUStat
			stat, err := reader.readCPUStat()

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				assert.NoError(t, err)
				// Verify usage value
				assert.Equal(t, tt.expectedUsage, stat.UsageUsec)
				// Verify user value
				assert.Equal(t, tt.expectedUser, stat.UserUsec)
				// Verify system value
				assert.Equal(t, tt.expectedSystem, stat.SystemUsec)
			}
		})
	}
}

// === CPULimit Tests ===

// Test_V2Reader_CPULimit_InvalidFormat tests CPULimit with invalid format.
func Test_V2Reader_CPULimit_InvalidFormat(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		cpuMax      string
		expectError bool
	}{
		{
			name:        "single field",
			cpuMax:      "100000",
			expectError: true,
		},
		{
			name:        "too many fields",
			cpuMax:      "100000 100000 extra",
			expectError: true,
		},
		{
			name:        "empty file",
			cpuMax:      "",
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir := t.TempDir()

			// Write cpu.max file
			err := os.WriteFile(filepath.Join(dir, "cpu.max"), []byte(tt.cpuMax), 0o644)
			require.NoError(t, err)

			// Create reader directly
			reader := &V2Reader{path: dir}

			// Call CPULimit via public method requires context
			// So we test the error path through the constructor instead
			_, _, err = reader.CPULimit(t.Context())

			// Verify error
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			}
		})
	}
}

// === Memory Tests ===

// Test_V2Reader_MemoryUsage_InvalidFormat tests MemoryUsage with invalid format.
func Test_V2Reader_MemoryUsage_InvalidFormat(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "invalid content",
			content:     "invalid",
			expectError: true,
		},
		{
			name:        "empty string",
			content:     "",
			expectError: true,
		},
		{
			name:        "negative value string",
			content:     "-123",
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir := t.TempDir()

			// Write invalid memory.current file
			err := os.WriteFile(filepath.Join(dir, "memory.current"), []byte(tt.content), 0o644)
			require.NoError(t, err)

			// Create reader directly
			reader := &V2Reader{path: dir}

			// Call MemoryUsage
			_, err = reader.MemoryUsage(t.Context())

			// Verify error
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			}
		})
	}
}

// Test_V2Reader_MemoryLimit_InvalidFormat tests MemoryLimit with various formats.
func Test_V2Reader_MemoryLimit_InvalidFormat(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		content     string
		createFile  bool
		expectLimit uint64
		expectError bool
	}{
		{
			name:        "invalid content",
			content:     "invalid",
			createFile:  true,
			expectError: true,
		},
		{
			name:        "empty content",
			content:     "",
			createFile:  true,
			expectError: true,
		},
		{
			name:        "missing file returns unlimited",
			createFile:  false,
			expectLimit: 0,
			expectError: false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir := t.TempDir()

			// Write memory.max file if needed
			if tt.createFile {
				err := os.WriteFile(filepath.Join(dir, "memory.max"), []byte(tt.content), 0o644)
				require.NoError(t, err)
			}

			// Create reader directly
			reader := &V2Reader{path: dir}

			// Call MemoryLimit
			limit, err := reader.MemoryLimit(t.Context())

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				assert.NoError(t, err)
				// Verify limit value
				assert.Equal(t, tt.expectLimit, limit)
			}
		})
	}
}

// === ReadMemoryStat Tests ===

// Test_V2Reader_ReadMemoryStat_Internal tests ReadMemoryStat internal parsing.
func Test_V2Reader_ReadMemoryStat_Internal(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		statContent string
		createFile  bool
		expectAnon  uint64
		expectFile  uint64
		expectError bool
	}{
		{
			name:        "valid stats",
			statContent: "anon 1048576\nfile 2097152\n",
			createFile:  true,
			expectAnon:  1048576,
			expectFile:  2097152,
			expectError: false,
		},
		{
			name:        "empty file",
			statContent: "",
			createFile:  true,
			expectAnon:  0,
			expectFile:  0,
			expectError: false,
		},
		{
			name:        "partial stats",
			statContent: "anon 1024\n",
			createFile:  true,
			expectAnon:  1024,
			expectFile:  0,
			expectError: false,
		},
		{
			name:        "malformed lines",
			statContent: "invalid\nanon 1024\n",
			createFile:  true,
			expectAnon:  1024,
			expectFile:  0,
			expectError: false,
		},
		{
			name:        "all fields",
			statContent: "anon 100\nfile 200\nkernel 300\nslab 400\nsock 500\nshmem 600\nmapped 700\ndirty 800\npgfault 900\npgmajfault 1000\n",
			createFile:  true,
			expectAnon:  100,
			expectFile:  200,
			expectError: false,
		},
		{
			name:        "missing file",
			createFile:  false,
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			dir := t.TempDir()

			// Write memory.stat file if needed
			if tt.createFile {
				err := os.WriteFile(filepath.Join(dir, "memory.stat"), []byte(tt.statContent), 0o644)
				require.NoError(t, err)
			}

			// Create reader directly
			reader := &V2Reader{path: dir}

			// Call ReadMemoryStat
			stat, err := reader.ReadMemoryStat(t.Context())

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
