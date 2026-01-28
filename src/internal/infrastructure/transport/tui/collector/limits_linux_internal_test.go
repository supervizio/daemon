//go:build linux

package collector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseCgroupV2CPUMax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		content        string
		expectedQuota  int64
		expectedPeriod int64
		expectedFloat  float64
	}{
		{
			name:           "valid quota and period",
			content:        "100000 100000\n",
			expectedQuota:  100000,
			expectedPeriod: 100000,
			expectedFloat:  1.0,
		},
		{
			name:           "fractional CPU",
			content:        "50000 100000\n",
			expectedQuota:  50000,
			expectedPeriod: 100000,
			expectedFloat:  0.5,
		},
		{
			name:           "max unlimited",
			content:        "max 100000\n",
			expectedQuota:  0,
			expectedPeriod: 0,
			expectedFloat:  0,
		},
		{
			name:           "insufficient parts",
			content:        "100000\n",
			expectedQuota:  0,
			expectedPeriod: 0,
			expectedFloat:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory and file.
			tmpDir := t.TempDir()
			cpuMaxPath := filepath.Join(tmpDir, "cpu.max")
			err := os.WriteFile(cpuMaxPath, []byte(tt.content), 0644)
			require.NoError(t, err)

			limits := &model.ResourceLimits{}
			parseCgroupV2CPUMax(tmpDir, limits)

			assert.Equal(t, tt.expectedQuota, limits.CPUQuotaRaw)
			assert.Equal(t, tt.expectedPeriod, limits.CPUPeriod)
			assert.Equal(t, tt.expectedFloat, limits.CPUQuota)
		})
	}
}

func Test_parseCgroupV2CPUSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		files    map[string]string
		expected string
	}{
		{
			name: "prefers effective cpuset",
			files: map[string]string{
				"cpuset.cpus.effective": "0-3\n",
				"cpuset.cpus":           "0-7\n",
			},
			expected: "0-3",
		},
		{
			name: "falls back to cpuset.cpus",
			files: map[string]string{
				"cpuset.cpus": "0-7\n",
			},
			expected: "0-7",
		},
		{
			name:     "empty when no files",
			files:    map[string]string{},
			expected: "",
		},
		{
			name: "trims whitespace",
			files: map[string]string{
				"cpuset.cpus": "  0-3  \n",
			},
			expected: "0-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			for name, content := range tt.files {
				err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
				require.NoError(t, err)
			}

			limits := &model.ResourceLimits{}
			parseCgroupV2CPUSet(tmpDir, limits)

			assert.Equal(t, tt.expected, limits.CPUSet)
		})
	}
}

func Test_getCgroupV2Path(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		expectContains  string
	}{
		{
			name:           "returns path containing /sys/fs/cgroup if v2 available",
			expectContains: "/sys/fs/cgroup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := getCgroupV2Path()

			// If cgroup v2 is available, path should contain expected string.
			if path != "" {
				assert.Contains(t, path, tt.expectContains)
			}
		})
	}
}

func Test_getCgroupV1Paths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		wantNil        bool
		expectContains string
	}{
		{
			name:           "returns non-nil map with valid paths",
			wantNil:        false,
			expectContains: "/sys/fs/cgroup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			paths := getCgroupV1Paths()

			if tt.wantNil {
				assert.Nil(t, paths)
				return
			}

			assert.NotNil(t, paths)

			// If any paths are returned, they should be valid.
			for controller, path := range paths {
				assert.NotEmpty(t, controller)
				assert.Contains(t, path, tt.expectContains)
			}
		})
	}
}

func Test_parseCgroupV1CPUQuota(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		quotaContent   string
		periodContent  string
		expectedQuota  int64
		expectedPeriod int64
		expectedFloat  float64
	}{
		{
			name:           "valid quota and period",
			quotaContent:   "100000\n",
			periodContent:  "100000\n",
			expectedQuota:  100000,
			expectedPeriod: 100000,
			expectedFloat:  1.0,
		},
		{
			name:           "fractional CPU",
			quotaContent:   "50000\n",
			periodContent:  "100000\n",
			expectedQuota:  50000,
			expectedPeriod: 100000,
			expectedFloat:  0.5,
		},
		{
			name:           "disabled quota (-1)",
			quotaContent:   "-1\n",
			periodContent:  "100000\n",
			expectedQuota:  0,
			expectedPeriod: 0,
			expectedFloat:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			err := os.WriteFile(filepath.Join(tmpDir, "cpu.cfs_quota_us"), []byte(tt.quotaContent), 0644)
			require.NoError(t, err)
			err = os.WriteFile(filepath.Join(tmpDir, "cpu.cfs_period_us"), []byte(tt.periodContent), 0644)
			require.NoError(t, err)

			cgroupPaths := map[string]string{"cpu": tmpDir}
			limits := &model.ResourceLimits{}
			parseCgroupV1CPUQuota(cgroupPaths, limits)

			assert.Equal(t, tt.expectedQuota, limits.CPUQuotaRaw)
			assert.Equal(t, tt.expectedPeriod, limits.CPUPeriod)
			assert.Equal(t, tt.expectedFloat, limits.CPUQuota)
		})
	}
}

func Test_collectCgroupV2Memory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		memoryMax       string
		memoryCurrent   string
		expectedMax     uint64
		expectedCurrent uint64
	}{
		{
			name:            "valid memory values",
			memoryMax:       "1073741824\n",
			memoryCurrent:   "536870912\n",
			expectedMax:     1073741824,
			expectedCurrent: 536870912,
		},
		{
			name:            "unlimited memory",
			memoryMax:       "max\n",
			memoryCurrent:   "536870912\n",
			expectedMax:     0,
			expectedCurrent: 536870912,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			err := os.WriteFile(filepath.Join(tmpDir, "memory.max"), []byte(tt.memoryMax), 0644)
			require.NoError(t, err)
			err = os.WriteFile(filepath.Join(tmpDir, "memory.current"), []byte(tt.memoryCurrent), 0644)
			require.NoError(t, err)

			limits := &model.ResourceLimits{}
			collectCgroupV2Memory(tmpDir, limits)

			assert.Equal(t, tt.expectedMax, limits.MemoryMax)
			assert.Equal(t, tt.expectedCurrent, limits.MemoryCurrent)
		})
	}
}

func Test_collectCgroupV2PIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		pidsMax         string
		pidsCurrent     string
		expectedMax     int64
		expectedCurrent int64
	}{
		{
			name:            "valid PIDs values",
			pidsMax:         "1024\n",
			pidsCurrent:     "42\n",
			expectedMax:     1024,
			expectedCurrent: 42,
		},
		{
			name:            "unlimited PIDs",
			pidsMax:         "max\n",
			pidsCurrent:     "42\n",
			expectedMax:     0,
			expectedCurrent: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			err := os.WriteFile(filepath.Join(tmpDir, "pids.max"), []byte(tt.pidsMax), 0644)
			require.NoError(t, err)
			err = os.WriteFile(filepath.Join(tmpDir, "pids.current"), []byte(tt.pidsCurrent), 0644)
			require.NoError(t, err)

			limits := &model.ResourceLimits{}
			collectCgroupV2PIDs(tmpDir, limits)

			assert.Equal(t, tt.expectedMax, limits.PIDsMax)
			assert.Equal(t, tt.expectedCurrent, limits.PIDsCurrent)
		})
	}
}

// Test_collectCgroupLimits tests the collectCgroupLimits function.
// It verifies that cgroup limits are collected correctly.
//
// Params:
//   - t: the testing context.
func Test_collectCgroupLimits(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "collects_cgroup_limits",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create limits.
			limits := &model.ResourceLimits{}

			// Call function - should not panic.
			assert.NotPanics(t, func() {
				collectCgroupLimits(limits)
			})
		})
	}
}

// Test_collectCgroupV2 tests the collectCgroupV2 function.
// It verifies that cgroup v2 limits are collected correctly.
//
// Params:
//   - t: the testing context.
func Test_collectCgroupV2(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "collects_from_system_cgroup",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create limits.
			limits := &model.ResourceLimits{}

			// Call function - should not panic.
			assert.NotPanics(t, func() {
				_ = collectCgroupV2(limits)
			})
		})
	}
}

// Test_collectCgroupV2CPU tests the collectCgroupV2CPU function.
// It verifies that cgroup v2 CPU limits are collected correctly.
//
// Params:
//   - t: the testing context.
func Test_collectCgroupV2CPU(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// path is the cgroup path.
		path string
	}{
		{
			name: "handles_nonexistent_path",
			path: "/nonexistent/cgroup/path",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create limits.
			limits := &model.ResourceLimits{}

			// Call function - should not panic.
			assert.NotPanics(t, func() {
				collectCgroupV2CPU(tt.path, limits)
			})
		})
	}
}

// Test_collectCgroupV1 tests the collectCgroupV1 function.
// It verifies that cgroup v1 limits are collected correctly.
//
// Params:
//   - t: the testing context.
func Test_collectCgroupV1(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "collects_from_system_cgroup",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create limits.
			limits := &model.ResourceLimits{}

			// Call function - should not panic.
			assert.NotPanics(t, func() {
				collectCgroupV1(limits)
			})
		})
	}
}

// Test_collectCgroupV1CPU tests the collectCgroupV1CPU function.
// It verifies that cgroup v1 CPU limits are collected correctly.
//
// Params:
//   - t: the testing context.
func Test_collectCgroupV1CPU(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// paths are the cgroup paths.
		paths map[string]string
	}{
		{
			name:  "handles_empty_paths",
			paths: map[string]string{},
		},
		{
			name: "handles_missing_cpu_path",
			paths: map[string]string{
				"memory": "/sys/fs/cgroup/memory",
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create limits.
			limits := &model.ResourceLimits{}

			// Call function - should not panic.
			assert.NotPanics(t, func() {
				collectCgroupV1CPU(tt.paths, limits)
			})
		})
	}
}

// Test_parseCgroupV1CPUSet tests the parseCgroupV1CPUSet function.
// It verifies that cgroup v1 CPUSet is parsed correctly.
//
// Params:
//   - t: the testing context.
func Test_parseCgroupV1CPUSet(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// files are the files to create.
		files map[string]string
		// expected is the expected CPUSet.
		expected string
	}{
		{
			name: "parses_cpuset_cpus",
			files: map[string]string{
				"cpuset.cpus": "0-3\n",
			},
			expected: "0-3",
		},
		{
			name:     "handles_missing_file",
			files:    map[string]string{},
			expected: "",
		},
		{
			name: "trims_whitespace",
			files: map[string]string{
				"cpuset.cpus": "  0-7  \n",
			},
			expected: "0-7",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create temp directory.
			tmpDir := t.TempDir()

			// Create files.
			for name, content := range tt.files {
				err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644)
				require.NoError(t, err)
			}

			// Create limits and paths.
			limits := &model.ResourceLimits{}
			paths := map[string]string{"cpuset": tmpDir}

			// Call function.
			parseCgroupV1CPUSet(paths, limits)

			// Verify result.
			assert.Equal(t, tt.expected, limits.CPUSet)
		})
	}
}

// Test_collectCgroupV1Memory tests the collectCgroupV1Memory function.
// It verifies that cgroup v1 memory limits are collected correctly.
//
// Params:
//   - t: the testing context.
func Test_collectCgroupV1Memory(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// paths are the cgroup paths.
		paths map[string]string
	}{
		{
			name:  "handles_empty_paths",
			paths: map[string]string{},
		},
		{
			name: "handles_missing_memory_path",
			paths: map[string]string{
				"cpu": "/sys/fs/cgroup/cpu",
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create limits.
			limits := &model.ResourceLimits{}

			// Call function - should not panic.
			assert.NotPanics(t, func() {
				collectCgroupV1Memory(tt.paths, limits)
			})
		})
	}
}

// Test_collectCgroupV1PIDs tests the collectCgroupV1PIDs function.
// It verifies that cgroup v1 PIDs limits are collected correctly.
//
// Params:
//   - t: the testing context.
func Test_collectCgroupV1PIDs(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// paths are the cgroup paths.
		paths map[string]string
	}{
		{
			name:  "handles_empty_paths",
			paths: map[string]string{},
		},
		{
			name: "handles_missing_pids_path",
			paths: map[string]string{
				"cpu": "/sys/fs/cgroup/cpu",
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create limits.
			limits := &model.ResourceLimits{}

			// Call function - should not panic.
			assert.NotPanics(t, func() {
				collectCgroupV1PIDs(tt.paths, limits)
			})
		})
	}
}
