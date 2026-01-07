// Package logging provides internal tests for nopcloser.go.
// It tests the nopCloser type using white-box testing.
package logging

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_nopCloser_Close tests the Close method of nopCloser.
//
// Params:
//   - t: the testing context.
func Test_nopCloser_Close(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// content is the content to write before closing.
		content string
	}{
		{
			name:    "close_returns_nil",
			content: "test content",
		},
		{
			name:    "close_empty_buffer",
			content: "",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			nc := &nopCloser{Writer: &buf}

			// Write content if provided.
			if tt.content != "" {
				_, err := nc.Write([]byte(tt.content))
				require.NoError(t, err)
			}

			err := nc.Close()
			assert.NoError(t, err)

			// Verify content was written.
			assert.Equal(t, tt.content, buf.String())
		})
	}
}
