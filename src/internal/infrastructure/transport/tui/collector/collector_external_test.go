// Package collector_test provides black-box tests for the collector package.
package collector_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/collector"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCollectors tests the NewCollectors constructor.
// It verifies that a new Collectors instance is properly initialized.
//
// Params:
//   - t: the testing context.
func TestNewCollectors(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// wantNonNil indicates if result should be non-nil.
		wantNonNil bool
	}{
		{
			name:       "returns_non_nil_collectors",
			wantNonNil: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the constructor.
			result := collector.NewCollectors()

			// Verify result is non-nil.
			if tt.wantNonNil {
				require.NotNil(t, result)
			}
		})
	}
}

// TestCollectors_Add tests the Add method.
// It verifies that collectors can be added and chained.
//
// Params:
//   - t: the testing context.
func TestCollectors_Add(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// addCount is the number of collectors to add.
		addCount int
	}{
		{
			name:     "add_single_collector",
			addCount: 1,
		},
		{
			name:     "add_multiple_collectors",
			addCount: 3,
		},
		{
			name:     "add_zero_collectors",
			addCount: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collectors.
			c := collector.NewCollectors()

			// Add collectors (using context collector as example).
			for i := range tt.addCount {
				result := c.Add(collector.NewContextCollector("1.0.0"))
				// Verify fluent interface returns self.
				assert.Same(t, c, result, "Add should return self for chaining at iteration %d", i)
			}
		})
	}
}

// TestCollectors_CollectAll tests the CollectAll method.
// It verifies that all collectors are invoked.
//
// Params:
//   - t: the testing context.
func TestCollectors_CollectAll(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// addCollectors is the number of collectors to add.
		addCollectors int
	}{
		{
			name:          "calls_all_gatherers",
			addCollectors: 3,
		},
		{
			name:          "empty_collectors",
			addCollectors: 0,
		},
		{
			name:          "single_gatherer",
			addCollectors: 1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collectors and add gatherers.
			c := collector.NewCollectors()
			for range tt.addCollectors {
				c.Add(collector.NewContextCollector("1.0.0"))
			}

			// Call CollectAll.
			snap := model.NewSnapshot()
			err := c.CollectAll(snap)

			// Verify no error returned.
			assert.NoError(t, err)
		})
	}
}

// TestDefaultCollectors tests the DefaultCollectors function.
// It verifies that the default collector set is properly configured.
//
// Params:
//   - t: the testing context.
func TestDefaultCollectors(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// version is the daemon version.
		version string
	}{
		{
			name:    "returns_configured_collectors",
			version: "1.0.0",
		},
		{
			name:    "with_empty_version",
			version: "",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call DefaultCollectors.
			result := collector.DefaultCollectors(tt.version)

			// Verify result is non-nil.
			require.NotNil(t, result)

			// Verify it can collect data.
			snap := model.NewSnapshot()
			err := result.CollectAll(snap)
			assert.NoError(t, err)
		})
	}
}

// TestCollectors_SetConfigPath tests the SetConfigPath method.
// It verifies that config path is set on the context collector.
//
// Params:
//   - t: the testing context.
func TestCollectors_SetConfigPath(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// path is the config path to set.
		path string
	}{
		{
			name: "sets_config_path",
			path: "/etc/supervizio/config.yaml",
		},
		{
			name: "sets_empty_path",
			path: "",
		},
		{
			name: "sets_custom_path",
			path: "/custom/path/config.yaml",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create default collectors.
			c := collector.DefaultCollectors("1.0.0")

			// Set config path (should not panic).
			c.SetConfigPath(tt.path)

			// Verify by collecting and checking snapshot.
			snap := model.NewSnapshot()
			err := c.CollectAll(snap)
			assert.NoError(t, err)

			// If path was set, it should be in the snapshot.
			if tt.path != "" {
				assert.Equal(t, tt.path, snap.Context.ConfigPath)
			}
		})
	}
}

// TestCollectors_SetConfigPath_without_context_collector tests SetConfigPath
// when no context collector is present.
//
// Params:
//   - t: the testing context.
func TestCollectors_SetConfigPath_without_context_collector(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "does_not_panic_without_context_collector",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collectors without context collector.
			c := collector.NewCollectors()
			c.Add(collector.NewNetworkCollector())

			// Should not panic.
			assert.NotPanics(t, func() {
				c.SetConfigPath("/some/path")
			})
		})
	}
}
