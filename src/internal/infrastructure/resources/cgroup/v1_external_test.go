//go:build linux

// Package cgroup_test provides external tests for the cgroup package.
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

// testMemoryStatDetailed is a detailed memory.stat file content for v1 testing.
const testMemoryStatDetailed string = `total_rss 1048576
total_cache 2097152
total_shmem 4096
total_mapped_file 8192
total_dirty 16384
total_pgfault 1000
total_pgmajfault 10
`

// === Test Constants ===

// testCPUUsageValue is the test CPU usage value in nanoseconds.
const testCPUUsageValue string = "1234567890"

// testCPUUsageExpected is the expected CPU usage in microseconds.
const testCPUUsageExpected uint64 = 1234567

// testCPUQuotaValue is the test CPU quota value.
const testCPUQuotaValue string = "100000"

// testCPUPeriodValue is the test CPU period value.
const testCPUPeriodValue string = "100000"

// testMemoryUsageValue is the test memory usage value.
const testMemoryUsageValue string = "104857600"

// testMemoryUsageExpected is the expected memory usage in bytes.
const testMemoryUsageExpected uint64 = 104857600

// testMemoryLimitValue is the test memory limit value.
const testMemoryLimitValue string = "209715200"

// testMemoryLimitExpected is the expected memory limit in bytes.
const testMemoryLimitExpected uint64 = 209715200

// === NewV1Reader Tests ===

// TestNewV1Reader tests the NewV1Reader constructor with various inputs.
func TestNewV1Reader(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		cpuPath     string
		memoryPath  string
		setup       func(t *testing.T) (string, string)
		expectError bool
	}{
		{
			name: "valid paths",
			setup: func(t *testing.T) (string, string) {
				// Create temporary directories
				cpuDir := t.TempDir()
				memDir := t.TempDir()
				// Create required files
				createCgroupFiles(t, cpuDir, memDir)
				// Return paths
				return cpuDir, memDir
			},
			expectError: false,
		},
		{
			name: "invalid cpu path",
			setup: func(t *testing.T) (string, string) {
				// Return non-existent CPU path
				return "/nonexistent/cpu", t.TempDir()
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
			// Setup test environment
			cpuPath, memPath := tt.setup(t)

			// Create reader
			reader, err := cgroup.NewV1Reader(cpuPath, memPath)

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

// TestV1Reader_CPUUsage tests the CPUUsage method.
func TestV1Reader_CPUUsage(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		setup       func(t *testing.T) *cgroup.V1Reader
		ctx         context.Context
		expected    uint64
		expectError bool
	}{
		{
			name: "valid cpu usage",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
			ctx:         context.Background(),
			expected:    testCPUUsageExpected,
			expectError: false,
		},
		{
			name: "cancelled context",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
			ctx: func() context.Context {
				// Create cancelled context
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				// Return cancelled context
				return ctx
			}(),
			expected:    0,
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup reader
			reader := tt.setup(t)

			// Call CPUUsage
			usage, err := reader.CPUUsage(tt.ctx)

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				require.NoError(t, err)
				// Verify usage value
				assert.Equal(t, tt.expected, usage)
			}
		})
	}
}

// === CPULimit Tests ===

// TestV1Reader_CPULimit tests the CPULimit method.
func TestV1Reader_CPULimit(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		setup          func(t *testing.T) *cgroup.V1Reader
		ctx            context.Context
		expectedQuota  uint64
		expectedPeriod uint64
		expectError    bool
	}{
		{
			name: "valid cpu limit",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
			ctx:            context.Background(),
			expectedQuota:  100000,
			expectedPeriod: 100000,
			expectError:    false,
		},
		{
			name: "unlimited cpu",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment with unlimited quota
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  "-1",
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
			ctx:            context.Background(),
			expectedQuota:  0,
			expectedPeriod: 0,
			expectError:    false,
		},
		{
			name: "cancelled context",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
			ctx: func() context.Context {
				// Create cancelled context
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				// Return cancelled context
				return ctx
			}(),
			expectedQuota:  0,
			expectedPeriod: 0,
			expectError:    true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup reader
			reader := tt.setup(t)

			// Call CPULimit
			quota, period, err := reader.CPULimit(tt.ctx)

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				require.NoError(t, err)
				// Verify quota value
				assert.Equal(t, tt.expectedQuota, quota)
				// Verify period value
				assert.Equal(t, tt.expectedPeriod, period)
			}
		})
	}
}

// === MemoryUsage Tests ===

// TestV1Reader_MemoryUsage tests the MemoryUsage method.
func TestV1Reader_MemoryUsage(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		setup       func(t *testing.T) *cgroup.V1Reader
		ctx         context.Context
		expected    uint64
		expectError bool
	}{
		{
			name: "valid memory usage",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
			ctx:         context.Background(),
			expected:    testMemoryUsageExpected,
			expectError: false,
		},
		{
			name: "cancelled context",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
			ctx: func() context.Context {
				// Create cancelled context
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				// Return cancelled context
				return ctx
			}(),
			expected:    0,
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup reader
			reader := tt.setup(t)

			// Call MemoryUsage
			usage, err := reader.MemoryUsage(tt.ctx)

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				require.NoError(t, err)
				// Verify usage value
				assert.Equal(t, tt.expected, usage)
			}
		})
	}
}

// === MemoryLimit Tests ===

// TestV1Reader_MemoryLimit tests the MemoryLimit method.
func TestV1Reader_MemoryLimit(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		setup       func(t *testing.T) *cgroup.V1Reader
		ctx         context.Context
		expected    uint64
		expectError bool
	}{
		{
			name: "valid memory limit",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
			ctx:         context.Background(),
			expected:    testMemoryLimitExpected,
			expectError: false,
		},
		{
			name: "unlimited memory",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment with effectively unlimited memory
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  "9223372036854771712",
				})
			},
			ctx:         context.Background(),
			expected:    0,
			expectError: false,
		},
		{
			name: "cancelled context",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
			ctx: func() context.Context {
				// Create cancelled context
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				// Return cancelled context
				return ctx
			}(),
			expected:    0,
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup reader
			reader := tt.setup(t)

			// Call MemoryLimit
			limit, err := reader.MemoryLimit(tt.ctx)

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				require.NoError(t, err)
				// Verify limit value
				assert.Equal(t, tt.expected, limit)
			}
		})
	}
}

// === ReadMemoryStat Tests ===

// TestV1Reader_ReadMemoryStat tests the ReadMemoryStat method.
func TestV1Reader_ReadMemoryStat(t *testing.T) {
	// Define test cases
	tests := []struct {
		name        string
		setup       func(t *testing.T) *cgroup.V1Reader
		ctx         context.Context
		expectError bool
	}{
		{
			name:        "valid memory stat",
			setup:       createTestV1ReaderWithMemoryStat,
			ctx:         context.Background(),
			expectError: false,
		},
		{
			name:  "cancelled context",
			setup: createTestV1ReaderWithMemoryStat,
			ctx: func() context.Context {
				// Create cancelled context
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				// Return cancelled context
				return ctx
			}(),
			expectError: true,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup reader
			reader := tt.setup(t)

			// Call ReadMemoryStat
			stat, err := reader.ReadMemoryStat(tt.ctx)

			// Verify expectations
			if tt.expectError {
				// Expect error
				assert.Error(t, err)
			} else {
				// Expect success
				require.NoError(t, err)
				// Verify stat has values
				assert.NotZero(t, stat.Anon)
				// Verify file stat
				assert.NotZero(t, stat.File)
			}
		})
	}
}

// === Accessor Tests ===

// TestV1Reader_Path tests the Path accessor method.
func TestV1Reader_Path(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		setup    func(t *testing.T) *cgroup.V1Reader
		expected string
	}{
		{
			name: "returns cpu path",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup reader
			reader := tt.setup(t)

			// Call Path
			path := reader.Path()

			// Verify path is not empty
			assert.NotEmpty(t, path)
		})
	}
}

// TestV1Reader_MemoryPath tests the MemoryPath accessor method.
func TestV1Reader_MemoryPath(t *testing.T) {
	// Define test cases
	tests := []struct {
		name  string
		setup func(t *testing.T) *cgroup.V1Reader
	}{
		{
			name: "returns memory path",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup reader
			reader := tt.setup(t)

			// Call MemoryPath
			path := reader.MemoryPath()

			// Verify path is not empty
			assert.NotEmpty(t, path)
		})
	}
}

// TestV1Reader_Version tests the Version accessor method.
func TestV1Reader_Version(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		setup    func(t *testing.T) *cgroup.V1Reader
		expected int
	}{
		{
			name: "returns version 1",
			setup: func(t *testing.T) *cgroup.V1Reader {
				// Create test environment
				return createTestV1Reader(t, testCgroupConfig{
					cpuUsage:  testCPUUsageValue,
					cpuQuota:  testCPUQuotaValue,
					cpuPeriod: testCPUPeriodValue,
					memUsage:  testMemoryUsageValue,
					memLimit:  testMemoryLimitValue,
				})
			},
			expected: 1,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup reader
			reader := tt.setup(t)

			// Call Version
			version := reader.Version()

			// Verify version
			assert.Equal(t, tt.expected, version)
		})
	}
}

// === Helper Types ===

// testCgroupConfig holds configuration for creating test cgroup files.
type testCgroupConfig struct {
	cpuUsage  string
	cpuQuota  string
	cpuPeriod string
	memUsage  string
	memLimit  string
}

// === Helper Functions ===

// createTestV1Reader creates a V1Reader with mock cgroup files.
func createTestV1Reader(t *testing.T, cfg testCgroupConfig) *cgroup.V1Reader {
	// Create temporary directories
	cpuDir := t.TempDir()
	memDir := t.TempDir()

	// Create CPU files
	err := os.WriteFile(filepath.Join(cpuDir, "cpuacct.usage"), []byte(cfg.cpuUsage), 0o644)
	require.NoError(t, err)

	// Create CPU quota file
	err = os.WriteFile(filepath.Join(cpuDir, "cpu.cfs_quota_us"), []byte(cfg.cpuQuota), 0o644)
	require.NoError(t, err)

	// Create CPU period file
	err = os.WriteFile(filepath.Join(cpuDir, "cpu.cfs_period_us"), []byte(cfg.cpuPeriod), 0o644)
	require.NoError(t, err)

	// Create memory usage file
	err = os.WriteFile(filepath.Join(memDir, "memory.usage_in_bytes"), []byte(cfg.memUsage), 0o644)
	require.NoError(t, err)

	// Create memory limit file
	err = os.WriteFile(filepath.Join(memDir, "memory.limit_in_bytes"), []byte(cfg.memLimit), 0o644)
	require.NoError(t, err)

	// Create memory.stat file
	statContent := "total_rss 1024\ntotal_cache 2048\n"
	err = os.WriteFile(filepath.Join(memDir, "memory.stat"), []byte(statContent), 0o644)
	require.NoError(t, err)

	// Create reader
	reader, err := cgroup.NewV1Reader(cpuDir, memDir)
	require.NoError(t, err)

	// Return reader
	return reader
}

// createTestV1ReaderWithMemoryStat creates a V1Reader with detailed memory.stat.
func createTestV1ReaderWithMemoryStat(t *testing.T) *cgroup.V1Reader {
	// Create temporary directories
	cpuDir := t.TempDir()
	memDir := t.TempDir()

	// Create CPU files
	err := os.WriteFile(filepath.Join(cpuDir, "cpuacct.usage"), []byte(testCPUUsageValue), 0o644)
	require.NoError(t, err)

	// Create CPU quota file
	err = os.WriteFile(filepath.Join(cpuDir, "cpu.cfs_quota_us"), []byte(testCPUQuotaValue), 0o644)
	require.NoError(t, err)

	// Create CPU period file
	err = os.WriteFile(filepath.Join(cpuDir, "cpu.cfs_period_us"), []byte(testCPUPeriodValue), 0o644)
	require.NoError(t, err)

	// Create memory usage file
	err = os.WriteFile(filepath.Join(memDir, "memory.usage_in_bytes"), []byte(testMemoryUsageValue), 0o644)
	require.NoError(t, err)

	// Create memory limit file
	err = os.WriteFile(filepath.Join(memDir, "memory.limit_in_bytes"), []byte(testMemoryLimitValue), 0o644)
	require.NoError(t, err)

	// Create detailed memory.stat file
	err = os.WriteFile(filepath.Join(memDir, "memory.stat"), []byte(testMemoryStatDetailed), 0o644)
	require.NoError(t, err)

	// Create reader
	reader, err := cgroup.NewV1Reader(cpuDir, memDir)
	require.NoError(t, err)

	// Return reader
	return reader
}

// createCgroupFiles creates the required cgroup files for testing.
func createCgroupFiles(t *testing.T, cpuDir, memDir string) {
	// Create CPU files
	err := os.WriteFile(filepath.Join(cpuDir, "cpuacct.usage"), []byte("0"), 0o644)
	require.NoError(t, err)

	// Create CPU quota file
	err = os.WriteFile(filepath.Join(cpuDir, "cpu.cfs_quota_us"), []byte("-1"), 0o644)
	require.NoError(t, err)

	// Create CPU period file
	err = os.WriteFile(filepath.Join(cpuDir, "cpu.cfs_period_us"), []byte("100000"), 0o644)
	require.NoError(t, err)

	// Create memory usage file
	err = os.WriteFile(filepath.Join(memDir, "memory.usage_in_bytes"), []byte("0"), 0o644)
	require.NoError(t, err)

	// Create memory limit file
	err = os.WriteFile(filepath.Join(memDir, "memory.limit_in_bytes"), []byte("0"), 0o644)
	require.NoError(t, err)

	// Create memory.stat file
	err = os.WriteFile(filepath.Join(memDir, "memory.stat"), []byte(""), 0o644)
	require.NoError(t, err)
}
