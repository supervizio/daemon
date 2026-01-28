// Package tui_test provides external tests.
package tui_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// TestNewSystemMetricsAdapter tests the NewSystemMetricsAdapter constructor.
// It verifies that a new SystemMetricsAdapter is properly initialized.
//
// Params:
//   - t: the testing context.
func TestNewSystemMetricsAdapter(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "returns_non_nil_adapter",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call constructor.
			adapter := tui.NewSystemMetricsAdapter()

			// Verify result.
			assert.NotNil(t, adapter)
		})
	}
}

// TestSystemMetricsAdapter_Metrics tests the Metrics method.
// It verifies that the adapter returns empty metrics.
//
// Params:
//   - t: the testing context.
func TestSystemMetricsAdapter_Metrics(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "returns_empty_metrics",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create adapter.
			adapter := tui.NewSystemMetricsAdapter()

			// Call Metrics.
			metrics := adapter.Metrics()

			// Verify result is empty SystemMetrics.
			assert.Equal(t, model.SystemMetrics{}, metrics)
		})
	}
}
