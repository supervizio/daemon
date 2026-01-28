// Package collector_test provides black-box tests for the collector package.
package collector_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/collector"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// TestNewContextCollector tests the NewContextCollector constructor.
// It verifies that a new ContextCollector is properly initialized.
//
// Params:
//   - t: the testing context.
func TestNewContextCollector(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// version is the daemon version.
		version string
		// wantNonNil indicates if result should be non-nil.
		wantNonNil bool
	}{
		{
			name:       "with_version",
			version:    "1.0.0",
			wantNonNil: true,
		},
		{
			name:       "with_empty_version",
			version:    "",
			wantNonNil: true,
		},
		{
			name:       "with_prerelease_version",
			version:    "2.0.0-beta.1",
			wantNonNil: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call constructor.
			result := collector.NewContextCollector(tt.version)

			// Verify result.
			if tt.wantNonNil {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestContextCollector_SetConfigPath tests the SetConfigPath method.
// It verifies that config path is set correctly.
//
// Params:
//   - t: the testing context.
func TestContextCollector_SetConfigPath(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// path is the config path to set.
		path string
	}{
		{
			name: "standard_path",
			path: "/etc/supervizio/config.yaml",
		},
		{
			name: "empty_path",
			path: "",
		},
		{
			name: "custom_path",
			path: "/custom/config.yml",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collector.
			c := collector.NewContextCollector("1.0.0")

			// Set config path (should not panic).
			c.SetConfigPath(tt.path)

			// Verify by gathering and checking snapshot.
			snap := &model.Snapshot{}
			err := c.Gather(snap)
			assert.NoError(t, err)
			assert.Equal(t, tt.path, snap.Context.ConfigPath)
		})
	}
}

// TestContextCollector_Gather tests the Gather method.
// It verifies that context information is properly collected.
//
// Params:
//   - t: the testing context.
func TestContextCollector_Gather(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// version is the daemon version.
		version string
		// configPath is the config path.
		configPath string
	}{
		{
			name:       "full_config",
			version:    "1.0.0",
			configPath: "/etc/supervizio/config.yaml",
		},
		{
			name:       "no_config_path",
			version:    "2.0.0",
			configPath: "",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collector.
			c := collector.NewContextCollector(tt.version)
			c.SetConfigPath(tt.configPath)

			// Create snapshot.
			snap := &model.Snapshot{}

			// Call Gather.
			err := c.Gather(snap)

			// Verify no error.
			assert.NoError(t, err)

			// Verify context is populated.
			assert.Equal(t, tt.version, snap.Context.Version)
			assert.Equal(t, tt.configPath, snap.Context.ConfigPath)
			assert.NotEmpty(t, snap.Context.OS)
			assert.NotEmpty(t, snap.Context.Arch)
			assert.NotEmpty(t, snap.Context.Hostname)
		})
	}
}

// TestContextCollector_uptime tests uptime calculation.
// It verifies that uptime is correctly calculated.
//
// Params:
//   - t: the testing context.
func TestContextCollector_uptime(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "uptime_increases_over_time",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collector.
			c := collector.NewContextCollector("1.0.0")

			// Create snapshot.
			snap1 := &model.Snapshot{}
			_ = c.Gather(snap1)

			// Wait a bit.
			time.Sleep(10 * time.Millisecond)

			// Gather again.
			snap2 := &model.Snapshot{}
			_ = c.Gather(snap2)

			// Second uptime should be greater.
			assert.GreaterOrEqual(t, snap2.Context.Uptime, snap1.Context.Uptime)
		})
	}
}
