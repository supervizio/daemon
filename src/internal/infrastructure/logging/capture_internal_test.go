// Package logging provides internal tests for capture.go.
// It tests internal implementation details using white-box testing.
package logging

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// errCloseFailed is an error returned when a mock close fails.
var errCloseFailed = errors.New("close failed")

// mockWriteCloser is a mock io.WriteCloser for testing.
type mockWriteCloser struct {
	// failOnClose indicates whether close should fail.
	failOnClose bool
}

// Write implements io.Writer.
//
// Params:
//   - p: the byte slice to write.
//
// Returns:
//   - int: the number of bytes written.
//   - error: always nil.
func (m *mockWriteCloser) Write(p []byte) (int, error) {
	// Return success.
	return len(p), nil
}

// Close implements io.Closer.
//
// Returns:
//   - error: an error if failOnClose is true.
func (m *mockWriteCloser) Close() error {
	// Check if close should fail.
	if m.failOnClose {
		// Return error for failing close.
		return errCloseFailed
	}
	// Return success for non-failing close.
	return nil
}

// Ensure interface is satisfied.
var _ io.WriteCloser = (*mockWriteCloser)(nil)

// Test_Capture_Close_stdoutError tests Close when stdout close fails.
//
// Params:
//   - t: the testing context.
func Test_Capture_Close_stdoutError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "stdout_close_fails",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			c := &Capture{
				stdout: &mockWriteCloser{failOnClose: true},
				stderr: &mockWriteCloser{failOnClose: false},
			}

			err := c.Close()
			assert.Error(t, err)
		})
	}
}

// Test_Capture_Close_stderrError tests Close when stderr close fails.
//
// Params:
//   - t: the testing context.
func Test_Capture_Close_stderrError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "stderr_close_fails",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			c := &Capture{
				stdout: &mockWriteCloser{failOnClose: false},
				stderr: &mockWriteCloser{failOnClose: true},
			}

			err := c.Close()
			assert.Error(t, err)
		})
	}
}

// Test_Capture_Close_bothError tests Close when both stdout and stderr close fail.
//
// Params:
//   - t: the testing context.
func Test_Capture_Close_bothError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "both_close_fail",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			c := &Capture{
				stdout: &mockWriteCloser{failOnClose: true},
				stderr: &mockWriteCloser{failOnClose: true},
			}

			err := c.Close()
			// Should return the first error (from stdout).
			assert.Error(t, err)
		})
	}
}
