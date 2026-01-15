//go:build linux

// Package signals provides platform-specific implementations of kernel interfaces.
// This file contains internal (white-box) tests for Linux-specific signal functionality.
package signals

import (
	"testing"
)

// TestPrctlSubreaper tests the prctlSubreaper function internally.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestPrctlSubreaper(t *testing.T) {
	// Define test cases for prctlSubreaper.
	tests := []struct {
		name string
		flag int
	}{
		{name: "enable subreaper", flag: 1},
		{name: "disable subreaper", flag: 0},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			err := prctlSubreaper(tt.flag)
			// Check if no error occurred.
			if err != nil {
				t.Errorf("prctlSubreaper(%d) returned error: %v", tt.flag, err)
			}
		})
	}
}
