//go:build linux

// Package cgroup_test provides external tests for the cgroup v2 package.
package cgroup_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/resources/cgroup"
)
// testCPUStatWithThrottling is a complete cpu.stat file content for v2 testing.
const testCPUStatWithThrottling string = `usage_usec 1234567
user_usec 800000
system_usec 434567
nr_periods 100
nr_throttled 5
throttled_usec 50000
`

// testMemoryStatComplete is a complete memory.stat file content for v2 testing.
const testMemoryStatComplete string = `anon 52428800
file 41943040
kernel 8388608
slab 2097152
sock 1048576
shmem 4194304
mapped 10485760
dirty 524288
pgfault 12345
pgmajfault 67
`

// === NewV2Reader Tests ===

// TestNewV2Reader tests the NewV2Reader constructor with various inputs.
func TestNewV2Reader(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectError bool
	}{
		{
			name: "valid path with all files",
			setup: func(t *testing.T) string {
				// Create temporary directory
				dir := t.TempDir()
				// Create all required cgroup v2 files
				createAllV2CgroupFiles(t, dir)
				// Return path
				return dir
			},
			expectError: false,
		},
		{
			name: "valid path empty directory",
			setup: func(t *testing.T) string {
				// Create temporary directory
				return t.TempDir()
			},
			expectError: false,
		},
		{
			name: "invalid path",
			setup: func(t *testing.T) string {
				// Return non-existent path
				return "/nonexistent/path/that/should/not/exist"
			},
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test environment
			path := tt.setup(t)

			// Create reader
			reader, err := cgroup.NewV2Reader(path)

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

// === CPUUsage Tests ===

// TestV2Reader_CPUUsage tests the CPUUsage method.
func TestV2Reader_CPUUsage(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name      string
		cpuStat   string
		wantUsage uint64
		wantErr   bool
	}{
		{
			name: "parses usage from cpu.stat",
			cpuStat: testCPUStatWithThrottling,
			wantUsage: 1234567,
			wantErr:   false,
		},
		{
			name: "handles zero usage",
			cpuStat: `usage_usec 0
user_usec 0
system_usec 0
`,
			wantUsage: 0,
			wantErr:   false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock cgroup directory
			mockCgroup := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "cpu.stat"), []byte(tt.cpuStat), 0o644))

			// Create reader
			reader, err := cgroup.NewV2Reader(mockCgroup)
			require.NoError(t, err)

			// Call CPUUsage
			usage, err := reader.CPUUsage(context.Background())

			// Verify expectations
			if tt.wantErr {
				// Expect error
				assert.Error(t, err)
				return
			}
			// Expect success
			require.NoError(t, err)
			// Verify usage value
			assert.Equal(t, tt.wantUsage, usage)
		})
	}
}

// === CPULimit Tests ===

// TestV2Reader_CPULimit tests the CPULimit method.
func TestV2Reader_CPULimit(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name       string
		cpuMax     string
		hasFile    bool
		wantQuota  uint64
		wantPeriod uint64
		wantErr    bool
	}{
		{
			name:       "parses quota and period",
			cpuMax:     "100000 100000\n",
			hasFile:    true,
			wantQuota:  100000,
			wantPeriod: 100000,
			wantErr:    false,
		},
		{
			name:       "handles max (unlimited)",
			cpuMax:     "max 100000\n",
			hasFile:    true,
			wantQuota:  0,
			wantPeriod: 100000,
			wantErr:    false,
		},
		{
			name:       "handles missing file",
			hasFile:    false,
			wantQuota:  0,
			wantPeriod: 0,
			wantErr:    false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock cgroup directory
			mockCgroup := t.TempDir()
			if tt.hasFile {
				// Write cpu.max file
				require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "cpu.max"), []byte(tt.cpuMax), 0o644))
			}

			// Create reader
			reader, err := cgroup.NewV2Reader(mockCgroup)
			require.NoError(t, err)

			// Call CPULimit
			quota, period, err := reader.CPULimit(context.Background())

			// Verify expectations
			if tt.wantErr {
				// Expect error
				assert.Error(t, err)
				return
			}
			// Expect success
			require.NoError(t, err)
			// Verify quota value
			assert.Equal(t, tt.wantQuota, quota)
			// Verify period value
			assert.Equal(t, tt.wantPeriod, period)
		})
	}
}

// === MemoryUsage Tests ===

// TestV2Reader_MemoryUsage tests the MemoryUsage method.
func TestV2Reader_MemoryUsage(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name      string
		content   string
		wantUsage uint64
		wantErr   bool
	}{
		{
			name:      "parses memory usage",
			content:   "104857600\n",
			wantUsage: 104857600,
			wantErr:   false,
		},
		{
			name:      "handles zero usage",
			content:   "0\n",
			wantUsage: 0,
			wantErr:   false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock cgroup directory
			mockCgroup := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "memory.current"), []byte(tt.content), 0o644))

			// Create reader
			reader, err := cgroup.NewV2Reader(mockCgroup)
			require.NoError(t, err)

			// Call MemoryUsage
			usage, err := reader.MemoryUsage(context.Background())

			// Verify expectations
			if tt.wantErr {
				// Expect error
				assert.Error(t, err)
				return
			}
			// Expect success
			require.NoError(t, err)
			// Verify usage value
			assert.Equal(t, tt.wantUsage, usage)
		})
	}
}

// === MemoryLimit Tests ===

// TestV2Reader_MemoryLimit tests the MemoryLimit method.
func TestV2Reader_MemoryLimit(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name      string
		content   string
		hasFile   bool
		wantLimit uint64
		wantErr   bool
	}{
		{
			name:      "parses limit",
			content:   "1073741824\n",
			hasFile:   true,
			wantLimit: 1073741824,
			wantErr:   false,
		},
		{
			name:      "handles max (unlimited)",
			content:   "max\n",
			hasFile:   true,
			wantLimit: 0,
			wantErr:   false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock cgroup directory
			mockCgroup := t.TempDir()
			if tt.hasFile {
				// Write memory.max file
				require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "memory.max"), []byte(tt.content), 0o644))
			}

			// Create reader
			reader, err := cgroup.NewV2Reader(mockCgroup)
			require.NoError(t, err)

			// Call MemoryLimit
			limit, err := reader.MemoryLimit(context.Background())

			// Verify expectations
			if tt.wantErr {
				// Expect error
				assert.Error(t, err)
				return
			}
			// Expect success
			require.NoError(t, err)
			// Verify limit value
			assert.Equal(t, tt.wantLimit, limit)
		})
	}
}

// === ReadMemoryStat Tests ===

// TestV2Reader_ReadMemoryStat tests the ReadMemoryStat method.
func TestV2Reader_ReadMemoryStat(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name        string
		memoryStat  string
		wantAnon    uint64
		wantFile    uint64
		wantKernel  uint64
		wantSlab    uint64
		wantShmem   uint64
		wantPgfault uint64
		wantPgmaj   uint64
		wantErr     bool
	}{
		{
			name: "parses all memory stat fields",
			memoryStat: testMemoryStatComplete,
			wantAnon:    52428800,
			wantFile:    41943040,
			wantKernel:  8388608,
			wantSlab:    2097152,
			wantShmem:   4194304,
			wantPgfault: 12345,
			wantPgmaj:   67,
			wantErr:     false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock cgroup directory
			mockCgroup := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "memory.stat"), []byte(tt.memoryStat), 0o644))

			// Create reader
			reader, err := cgroup.NewV2Reader(mockCgroup)
			require.NoError(t, err)

			// Call ReadMemoryStat
			stat, err := reader.ReadMemoryStat(context.Background())

			// Verify expectations
			if tt.wantErr {
				// Expect error
				assert.Error(t, err)
				return
			}
			// Expect success
			require.NoError(t, err)

			// Verify stat values
			assert.Equal(t, tt.wantAnon, stat.Anon)
			assert.Equal(t, tt.wantFile, stat.File)
			assert.Equal(t, tt.wantKernel, stat.Kernel)
			assert.Equal(t, tt.wantSlab, stat.Slab)
			assert.Equal(t, tt.wantShmem, stat.Shmem)
			assert.Equal(t, tt.wantPgfault, stat.Pgfault)
			assert.Equal(t, tt.wantPgmaj, stat.Pgmajfault)
		})
	}
}

// === Version Tests ===

// TestV2Reader_Version tests the Version accessor method.
func TestV2Reader_Version(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name     string
		expected int
	}{
		{
			name:     "returns version 2",
			expected: 2,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory
			dir := t.TempDir()

			// Create reader
			reader, err := cgroup.NewV2Reader(dir)
			require.NoError(t, err)

			// Call Version
			version := reader.Version()

			// Verify version is 2
			assert.Equal(t, tt.expected, version)
		})
	}
}

// === Context Cancellation Tests ===

// TestV2Reader_ContextCancellation tests context cancellation handling.
func TestV2Reader_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name   string
		method string
	}{
		{name: "CPUUsage returns context error", method: "CPUUsage"},
		{name: "CPULimit returns context error", method: "CPULimit"},
		{name: "MemoryUsage returns context error", method: "MemoryUsage"},
		{name: "MemoryLimit returns context error", method: "MemoryLimit"},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock cgroup directory
			mockCgroup := t.TempDir()
			reader, err := cgroup.NewV2Reader(mockCgroup)
			require.NoError(t, err)

			// Create cancelled context
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// Call method and verify error
			switch tt.method {
			// Test CPUUsage
			case "CPUUsage":
				_, err = reader.CPUUsage(ctx)
			// Test CPULimit
			case "CPULimit":
				_, _, err = reader.CPULimit(ctx)
			// Test MemoryUsage
			case "MemoryUsage":
				_, err = reader.MemoryUsage(ctx)
			// Test MemoryLimit
			case "MemoryLimit":
				_, err = reader.MemoryLimit(ctx)
			}

			// Verify context error
			assert.ErrorIs(t, err, context.Canceled)
		})
	}
}

// === Path Tests ===

// TestV2Reader_Path tests the Path accessor method.
func TestV2Reader_Path(t *testing.T) {
	t.Parallel()

	// Define test cases
	tests := []struct {
		name string
	}{
		{name: "returns configured path"},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock cgroup directory
			mockCgroup := t.TempDir()
			reader, err := cgroup.NewV2Reader(mockCgroup)
			require.NoError(t, err)

			// Verify path
			assert.Equal(t, mockCgroup, reader.Path())
		})
	}
}

// === Helper Functions ===

// createAllV2CgroupFiles creates all cgroup v2 files for testing.
func createAllV2CgroupFiles(t *testing.T, dir string) {
	t.Helper()

	// Create cpu.stat file
	cpuStatContent := "usage_usec 1234567\nuser_usec 500000\nsystem_usec 200000\n"
	err := os.WriteFile(filepath.Join(dir, "cpu.stat"), []byte(cpuStatContent), 0o644)
	require.NoError(t, err)

	// Create cpu.max file
	err = os.WriteFile(filepath.Join(dir, "cpu.max"), []byte("100000 100000"), 0o644)
	require.NoError(t, err)

	// Create memory.current file
	err = os.WriteFile(filepath.Join(dir, "memory.current"), []byte("104857600"), 0o644)
	require.NoError(t, err)

	// Create memory.max file
	err = os.WriteFile(filepath.Join(dir, "memory.max"), []byte("209715200"), 0o644)
	require.NoError(t, err)

	// Create memory.stat file
	statContent := "anon 1048576\nfile 2097152\n"
	err = os.WriteFile(filepath.Join(dir, "memory.stat"), []byte(statContent), 0o644)
	require.NoError(t, err)
}
