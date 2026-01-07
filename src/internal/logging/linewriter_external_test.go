// Package logging_test provides external tests for linewriter.go.
// It tests the public API of the LineWriter type using black-box testing.
package logging_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/logging"
)

// TestNewLineWriter tests the NewLineWriter constructor.
//
// Params:
//   - t: the testing context.
func TestNewLineWriter(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prefix is the prefix to use.
		prefix string
	}{
		{
			name:   "empty_prefix",
			prefix: "",
		},
		{
			name:   "simple_prefix",
			prefix: "[LOG] ",
		},
		{
			name:   "timestamp_style_prefix",
			prefix: "2024-01-15T10:30:45Z ",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			lw := logging.NewLineWriter(&buf, tt.prefix)
			assert.NotNil(t, lw)
		})
	}
}

// TestLineWriter_Write tests the Write method with various inputs.
//
// Params:
//   - t: the testing context.
func TestLineWriter_Write(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prefix is the prefix to use.
		prefix string
		// input is the input to write.
		input string
		// expected is the expected output.
		expected string
	}{
		{
			name:     "complete_line_no_prefix",
			prefix:   "",
			input:    "hello world\n",
			expected: "hello world\n",
		},
		{
			name:     "complete_line_with_prefix",
			prefix:   "[PREFIX] ",
			input:    "hello world\n",
			expected: "[PREFIX] hello world\n",
		},
		{
			name:     "partial_line_buffered",
			prefix:   "[PREFIX] ",
			input:    "hello",
			expected: "",
		},
		{
			name:     "multiple_lines",
			prefix:   ">> ",
			input:    "line1\nline2\n",
			expected: ">> line1\n>> line2\n",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			lw := logging.NewLineWriter(&buf, tt.prefix)

			n, err := lw.Write([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, len(tt.input), n)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

// TestLineWriter_Flush tests the Flush method.
//
// Params:
//   - t: the testing context.
func TestLineWriter_Flush(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prefix is the prefix to use.
		prefix string
		// input is the input to write.
		input string
		// expectedAfterWrite is the expected output after write.
		expectedAfterWrite string
		// expectedAfterFlush is the expected output after flush.
		expectedAfterFlush string
	}{
		{
			name:               "flush_empty_buffer",
			prefix:             "",
			input:              "",
			expectedAfterWrite: "",
			expectedAfterFlush: "",
		},
		{
			name:               "flush_partial_line_no_prefix",
			prefix:             "",
			input:              "partial",
			expectedAfterWrite: "",
			expectedAfterFlush: "partial\n",
		},
		{
			name:               "flush_partial_line_with_prefix",
			prefix:             "[LOG] ",
			input:              "partial",
			expectedAfterWrite: "",
			expectedAfterFlush: "[LOG] partial\n",
		},
		{
			name:               "flush_after_complete_line",
			prefix:             "",
			input:              "complete\n",
			expectedAfterWrite: "complete\n",
			expectedAfterFlush: "complete\n",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			var buf strings.Builder
			lw := logging.NewLineWriter(&buf, tt.prefix)

			// Write input if provided.
			if tt.input != "" {
				_, err := lw.Write([]byte(tt.input))
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedAfterWrite, buf.String())

			err := lw.Flush()
			require.NoError(t, err)
			assert.Equal(t, tt.expectedAfterFlush, buf.String())
		})
	}
}
