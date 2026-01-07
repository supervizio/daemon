// Package logging_test provides external tests for linewriter.go.
// It tests the public API of the LineWriter type using black-box testing.
package logging_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/logging"
)

// errWriteFailed is an error returned when a mock writer fails.
var errWriteFailed = errors.New("write failed")

// failingWriter is a mock writer that fails on write.
type failingWriter struct {
	// failAfter specifies how many writes should succeed before failing.
	failAfter int
	// writeCount tracks the number of writes.
	writeCount int
}

// Write implements io.Writer and fails after failAfter writes.
//
// Params:
//   - p: the byte slice to write.
//
// Returns:
//   - int: the number of bytes written.
//   - error: an error after failAfter writes.
func (fw *failingWriter) Write(p []byte) (int, error) {
	fw.writeCount++
	// Check if we should fail now.
	if fw.writeCount > fw.failAfter {
		// Return error after failAfter successful writes.
		return 0, errWriteFailed
	}
	// Return success for writes before failAfter threshold.
	return len(p), nil
}

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

// TestLineWriter_Write_PrefixError tests Write when prefix writing fails.
//
// Params:
//   - t: the testing context.
func TestLineWriter_Write_PrefixError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prefix is the prefix to use.
		prefix string
		// input is the input to write.
		input string
		// failAfter is the number of writes to succeed before failing.
		failAfter int
	}{
		{
			name:      "prefix_write_fails",
			prefix:    "[PREFIX] ",
			input:     "hello world\n",
			failAfter: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			fw := &failingWriter{failAfter: tt.failAfter}
			lw := logging.NewLineWriter(fw, tt.prefix)

			n, err := lw.Write([]byte(tt.input))
			assert.Error(t, err)
			assert.Equal(t, 0, n)
		})
	}
}

// TestLineWriter_Write_LineError tests Write when line writing fails.
//
// Params:
//   - t: the testing context.
func TestLineWriter_Write_LineError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prefix is the prefix to use.
		prefix string
		// input is the input to write.
		input string
		// failAfter is the number of writes to succeed before failing.
		failAfter int
	}{
		{
			name:      "line_write_fails_with_prefix",
			prefix:    "[PREFIX] ",
			input:     "hello world\n",
			failAfter: 1,
		},
		{
			name:      "line_write_fails_no_prefix",
			prefix:    "",
			input:     "hello world\n",
			failAfter: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			fw := &failingWriter{failAfter: tt.failAfter}
			lw := logging.NewLineWriter(fw, tt.prefix)

			n, err := lw.Write([]byte(tt.input))
			assert.Error(t, err)
			assert.Equal(t, 0, n)
		})
	}
}

// TestLineWriter_Flush_PrefixError tests Flush when prefix writing fails.
//
// Params:
//   - t: the testing context.
func TestLineWriter_Flush_PrefixError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prefix is the prefix to use.
		prefix string
		// input is the input to write without newline.
		input string
		// failAfter is the number of writes to succeed before failing.
		failAfter int
	}{
		{
			name:      "flush_prefix_write_fails",
			prefix:    "[PREFIX] ",
			input:     "partial",
			failAfter: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// First use a real buffer to write the partial content.
			var buf strings.Builder
			lw := logging.NewLineWriter(&buf, tt.prefix)

			// Write partial content (no newline).
			_, err := lw.Write([]byte(tt.input))
			require.NoError(t, err)

			// Now replace writer with failing writer for flush.
			fw := &failingWriter{failAfter: tt.failAfter}
			lwFail := logging.NewLineWriter(fw, tt.prefix)

			// Write partial content to buffer.
			_, _ = lwFail.Write([]byte(tt.input))

			// Flush should fail on prefix write.
			err = lwFail.Flush()
			assert.Error(t, err)
		})
	}
}

// TestLineWriter_Flush_ContentError tests Flush when content writing fails.
//
// Params:
//   - t: the testing context.
func TestLineWriter_Flush_ContentError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prefix is the prefix to use.
		prefix string
		// input is the input to write without newline.
		input string
		// failAfter is the number of writes to succeed before failing.
		failAfter int
	}{
		{
			name:      "flush_content_write_fails",
			prefix:    "[PREFIX] ",
			input:     "partial",
			failAfter: 1,
		},
		{
			name:      "flush_content_write_fails_no_prefix",
			prefix:    "",
			input:     "partial",
			failAfter: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			fw := &failingWriter{failAfter: tt.failAfter}
			lw := logging.NewLineWriter(fw, tt.prefix)

			// Write partial content (no newline) - this buffers without writing.
			_, err := lw.Write([]byte(tt.input))
			require.NoError(t, err)

			// Flush should fail.
			err = lw.Flush()
			assert.Error(t, err)
		})
	}
}

// TestLineWriter_Flush_NewlineError tests Flush when newline writing fails.
//
// Params:
//   - t: the testing context.
func TestLineWriter_Flush_NewlineError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// prefix is the prefix to use.
		prefix string
		// input is the input to write without newline.
		input string
		// failAfter is the number of writes to succeed before failing.
		failAfter int
	}{
		{
			name:      "flush_newline_write_fails",
			prefix:    "[PREFIX] ",
			input:     "partial",
			failAfter: 2,
		},
		{
			name:      "flush_newline_write_fails_no_prefix",
			prefix:    "",
			input:     "partial",
			failAfter: 1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			fw := &failingWriter{failAfter: tt.failAfter}
			lw := logging.NewLineWriter(fw, tt.prefix)

			// Write partial content (no newline) - this buffers without writing.
			_, err := lw.Write([]byte(tt.input))
			require.NoError(t, err)

			// Flush should fail on newline write.
			err = lw.Flush()
			assert.Error(t, err)
		})
	}
}
