// Package logging provides internal tests for multiwriter.go.
// It tests internal implementation details using white-box testing.
package logging

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiWriter_internals tests internal state of MultiWriter.
//
// Params:
//   - t: the testing context.
func TestMultiWriter_internals(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// writerCount is the number of writers to create.
		writerCount int
	}{
		{
			name:        "no_writers",
			writerCount: 0,
		},
		{
			name:        "single_writer",
			writerCount: 1,
		},
		{
			name:        "multiple_writers",
			writerCount: 3,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create mock writers using existing mockWriteCloser.
			writers := make([]io.WriteCloser, tt.writerCount)
			for i := range tt.writerCount {
				writers[i] = &mockWriteCloser{}
			}

			// Create multiwriter.
			mw := NewMultiWriter(writers...)
			require.NotNil(t, mw)

			// Verify internal writers slice length.
			assert.Equal(t, tt.writerCount, len(mw.writers))
		})
	}
}

// TestMultiWriter_writers_field tests the writers field is correctly set.
//
// Params:
//   - t: the testing context.
func TestMultiWriter_writers_field(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// writerCount is the number of writers to create.
		writerCount int
	}{
		{
			name:        "empty_writers_slice",
			writerCount: 0,
		},
		{
			name:        "two_writers",
			writerCount: 2,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create mock writers using existing mockWriteCloser.
			writers := make([]io.WriteCloser, tt.writerCount)
			for i := range tt.writerCount {
				writers[i] = &mockWriteCloser{}
			}

			// Create multiwriter.
			mw := NewMultiWriter(writers...)
			require.NotNil(t, mw)

			// Verify the writers field is properly initialized.
			if tt.writerCount == 0 {
				assert.Empty(t, mw.writers)
			} else {
				assert.NotEmpty(t, mw.writers)
				assert.Equal(t, tt.writerCount, len(mw.writers))
			}
		})
	}
}

// TestMultiWriter_Write_internal tests Write updates internal state.
//
// Params:
//   - t: the testing context.
func TestMultiWriter_Write_internal(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// message is the message to write.
		message string
		// expectedLen is the expected bytes written.
		expectedLen int
	}{
		{
			name:        "write_updates_all_writers",
			message:     "test message",
			expectedLen: 12,
		},
		{
			name:        "empty_write",
			message:     "",
			expectedLen: 0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			w1 := &mockWriteCloser{}
			w2 := &mockWriteCloser{}

			// Create multiwriter with two writers.
			mw := NewMultiWriter(w1, w2)
			require.NotNil(t, mw)

			// Verify writers are stored.
			assert.Equal(t, 2, len(mw.writers))

			// Write message.
			n, err := mw.Write([]byte(tt.message))
			require.NoError(t, err)
			assert.Equal(t, tt.expectedLen, n)
		})
	}
}
