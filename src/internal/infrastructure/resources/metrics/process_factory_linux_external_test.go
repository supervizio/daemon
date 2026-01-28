//go:build linux

// Package metrics_test provides black-box tests for the metrics package.
// It tests the public API of metrics factory functions.
package metrics_test

import (
	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics"
	"testing"

	"github.com/stretchr/testify/assert"

)

// TestNewProcessCollector tests metrics.NewProcessCollector function.
// It verifies that a process collector is correctly created for the current platform.
//
// Params:
//   - t: testing context
func TestNewProcessCollector(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		name string
	}{
		{
			name: "creates linux process collector",
		},
	}

	// Run all test cases.
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := metrics.NewProcessCollector()

			// Verify collector is not nil.
			assert.NotNil(t, result)
		})
	}
}
