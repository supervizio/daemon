// Package shared provides common domain types used across multiple domain packages.
package shared_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/shared"
)

// TestParseSize verifies that ParseSize correctly parses valid and invalid size strings.
//
// Params:
//   - t: testing context for assertions
func TestParseSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expected    int64
		expectErr   bool
		expectedErr error
		errContains string
	}{
		{name: "plain number", input: "100", expected: 100, expectErr: false},
		{name: "bytes lowercase", input: "100b", expected: 100, expectErr: false},
		{name: "bytes uppercase", input: "100B", expected: 100, expectErr: false},
		{name: "kilobytes", input: "10KB", expected: 10 * shared.Kilobyte, expectErr: false},
		{name: "megabytes", input: "5MB", expected: 5 * shared.Megabyte, expectErr: false},
		{name: "gigabytes", input: "2GB", expected: 2 * shared.Gigabyte, expectErr: false},
		{name: "with spaces", input: "  50 KB  ", expected: 50 * shared.Kilobyte, expectErr: false},
		{name: "mixed case", input: "10kb", expected: 10 * shared.Kilobyte, expectErr: false},
		{name: "empty string", input: "", expected: 0, expectErr: true, expectedErr: shared.ErrEmptySize},
		{name: "invalid number", input: "abcKB", expected: 0, expectErr: true, errContains: "invalid size number"},
		{name: "negative value", input: "-10MB", expected: 0, expectErr: true, expectedErr: shared.ErrNegativeSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := shared.ParseSize(tt.input)

			if tt.expectErr {
				require.Error(t, err)

				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}

				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatSize verifies that FormatSize correctly formats sizes to strings.
//
// Params:
//   - t: testing context for assertions
func TestFormatSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{name: "bytes", input: 100, expected: "100B"},
		{name: "kilobytes", input: shared.Kilobyte, expected: "1KB"},
		{name: "megabytes", input: shared.Megabyte, expected: "1MB"},
		{name: "gigabytes", input: shared.Gigabyte, expected: "1GB"},
		{name: "multiple kb", input: 5 * shared.Kilobyte, expected: "5KB"},
		{name: "multiple mb", input: 10 * shared.Megabyte, expected: "10MB"},
		{name: "multiple gb", input: 3 * shared.Gigabyte, expected: "3GB"},
		{name: "zero bytes", input: 0, expected: "0B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := shared.FormatSize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSizeConstants verifies that size constants have correct values.
//
// Params:
//   - t: testing context for assertions
func TestSizeConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant int64
		expected int64
	}{
		{name: "Byte", constant: shared.Byte, expected: 1},
		{name: "Kilobyte", constant: shared.Kilobyte, expected: 1024},
		{name: "Megabyte", constant: shared.Megabyte, expected: 1024 * 1024},
		{name: "Gigabyte", constant: shared.Gigabyte, expected: 1024 * 1024 * 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}
