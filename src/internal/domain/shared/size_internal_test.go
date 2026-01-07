// Package shared provides common domain types used across multiple domain packages.
package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_extractSizeComponents verifies that extractSizeComponents correctly extracts
// the multiplier and numeric part from size strings.
//
// Params:
//   - t: testing context for assertions
func Test_extractSizeComponents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		input              string
		expectedMultiplier int64
		expectedNumeric    string
	}{
		{name: "gigabytes", input: "10GB", expectedMultiplier: Gigabyte, expectedNumeric: "10"},
		{name: "megabytes", input: "5MB", expectedMultiplier: Megabyte, expectedNumeric: "5"},
		{name: "kilobytes", input: "100KB", expectedMultiplier: Kilobyte, expectedNumeric: "100"},
		{name: "bytes", input: "500B", expectedMultiplier: Byte, expectedNumeric: "500"},
		{name: "no suffix", input: "1000", expectedMultiplier: Byte, expectedNumeric: "1000"},
		{name: "zero gb", input: "0GB", expectedMultiplier: Gigabyte, expectedNumeric: "0"},
		{name: "zero mb", input: "0MB", expectedMultiplier: Megabyte, expectedNumeric: "0"},
		{name: "zero kb", input: "0KB", expectedMultiplier: Kilobyte, expectedNumeric: "0"},
		{name: "zero b", input: "0B", expectedMultiplier: Byte, expectedNumeric: "0"},
		{name: "empty string", input: "", expectedMultiplier: Byte, expectedNumeric: ""},
		{name: "only B", input: "B", expectedMultiplier: Byte, expectedNumeric: ""},
		{name: "only GB", input: "GB", expectedMultiplier: Gigabyte, expectedNumeric: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			multiplier, numericPart := extractSizeComponents(tt.input)
			assert.Equal(t, tt.expectedMultiplier, multiplier)
			assert.Equal(t, tt.expectedNumeric, numericPart)
		})
	}
}
