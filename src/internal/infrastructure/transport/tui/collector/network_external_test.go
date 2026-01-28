// Package collector_test provides black-box tests for the collector package.
package collector_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/collector"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// TestNewNetworkCollector tests the NewNetworkCollector constructor.
// It verifies that a new NetworkCollector is properly initialized.
//
// Params:
//   - t: the testing context.
func TestNewNetworkCollector(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// wantNonNil indicates if result should be non-nil.
		wantNonNil bool
	}{
		{
			name:       "returns_non_nil_collector",
			wantNonNil: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call constructor.
			result := collector.NewNetworkCollector()

			// Verify result.
			if tt.wantNonNil {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestNetworkCollector_Gather tests the Gather method.
// It verifies that network information is properly collected.
//
// Params:
//   - t: the testing context.
func TestNetworkCollector_Gather(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "gathers_network_interfaces",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create collector.
			c := collector.NewNetworkCollector()

			// Create snapshot.
			snap := &model.Snapshot{}

			// Call Gather.
			err := c.Gather(snap)

			// Verify no error.
			assert.NoError(t, err)

			// Network slice should be non-nil.
			assert.NotNil(t, snap.Network)
		})
	}
}
