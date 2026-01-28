// Package collector provides internal tests for collector.go.
// It tests internal implementation details using white-box testing.
package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_defaultCollectorsCap tests the defaultCollectorsCap constant.
// It verifies that the constant has the expected value.
//
// Params:
//   - t: the testing context.
func Test_defaultCollectorsCap(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// want is the expected value.
		want int
	}{
		{
			name: "has_expected_value",
			want: 8,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify constant value.
			assert.Equal(t, tt.want, defaultCollectorsCap)
		})
	}
}
