// Package logging_test provides external tests for timestamp.go.
// It tests the public API of timestamp functions using black-box testing.
package logging_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/infrastructure/logging"
)

// TestFormatTimestamp tests the FormatTimestamp function with various formats.
//
// Params:
//   - t: the testing context.
func TestFormatTimestamp(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)

	tests := []struct {
		// name is the test case name.
		name string
		// format is the timestamp format to test.
		format string
		// contains is the expected substring in the result.
		contains string
	}{
		{
			name:     "iso8601_format",
			format:   logging.FormatISO8601,
			contains: "2024-01-15T10:30:45Z",
		},
		{
			name:     "rfc3339_format",
			format:   logging.FormatRFC3339,
			contains: "2024-01-15T10:30:45",
		},
		{
			name:     "unix_format",
			format:   logging.FormatUnix,
			contains: "1",
		},
		{
			name:     "unix_milli_format",
			format:   logging.FormatUnixMilli,
			contains: "1",
		},
		{
			name:     "unix_nano_format",
			format:   logging.FormatUnixNano,
			contains: "1",
		},
		{
			name:     "custom_date_only",
			format:   "2006-01-02",
			contains: "2024-01-15",
		},
		{
			name:     "custom_time_only",
			format:   "15:04:05",
			contains: "10:30:45",
		},
		{
			name:     "empty_defaults_to_iso8601",
			format:   "",
			contains: "2024-01-15T10:30:45Z",
		},
	}

	// Iterate through all timestamp format test cases.
	for _, tt := range tests {
		// Test case runs the specific timestamp format scenario.
		t.Run(tt.name, func(t *testing.T) {
			result := logging.FormatTimestamp(now, tt.format)
			assert.Contains(t, result, tt.contains)
		})
	}
}

// TestParseTimestampFormat tests the ParseTimestampFormat function.
//
// Params:
//   - t: the testing context.
func TestParseTimestampFormat(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// input is the format string to parse.
		input string
		// expected is the expected parsed format.
		expected string
	}{
		{
			name:     "empty_defaults_to_iso8601",
			input:    "",
			expected: logging.FormatISO8601,
		},
		{
			name:     "iso8601_returns_iso8601",
			input:    logging.FormatISO8601,
			expected: logging.FormatISO8601,
		},
		{
			name:     "rfc3339_returns_rfc3339",
			input:    logging.FormatRFC3339,
			expected: logging.FormatRFC3339,
		},
		{
			name:     "unix_returns_unix",
			input:    logging.FormatUnix,
			expected: logging.FormatUnix,
		},
		{
			name:     "unix_milli_returns_unix_milli",
			input:    logging.FormatUnixMilli,
			expected: logging.FormatUnixMilli,
		},
		{
			name:     "unix_nano_returns_unix_nano",
			input:    logging.FormatUnixNano,
			expected: logging.FormatUnixNano,
		},
		{
			name:     "custom_format_returned_as_is",
			input:    "2006-01-02",
			expected: "2006-01-02",
		},
	}

	// Iterate through all parse format test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			result := logging.ParseTimestampFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
