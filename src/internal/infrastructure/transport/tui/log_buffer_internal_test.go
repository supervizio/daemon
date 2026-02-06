// Package tui provides internal tests for log buffer.
package tui

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestLogBufferRing(t *testing.T) {
	tests := []struct {
		name      string
		bufSize   int
		addCount  int
		wantCount int
	}{
		{"ring wraps", 3, 5, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewLogBuffer(tt.bufSize)
			for i := range tt.addCount {
				buf.Add(model.LogEntry{Message: string(rune('A' + i))})
			}
			entries := buf.Entries()
			assert.Len(t, entries, tt.wantCount)
		})
	}
}

func TestLogBufferLevelCounts(t *testing.T) {
	tests := []struct {
		name      string
		levels    []string
		wantInfo  int
		wantWarn  int
		wantError int
	}{
		{
			name:      "various levels",
			levels:    []string{"INFO", "INFO", "WARN", "ERROR"},
			wantInfo:  2,
			wantWarn:  1,
			wantError: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := NewLogBuffer(10)
			for _, level := range tt.levels {
				buf.Add(model.LogEntry{Level: level})
			}
			summary := buf.Summary()
			assert.Equal(t, tt.wantInfo, summary.InfoCount)
			assert.Equal(t, tt.wantWarn, summary.WarnCount)
			assert.Equal(t, tt.wantError, summary.ErrorCount)
		})
	}
}

// Test_LogBuffer_entriesLocked tests the entriesLocked method.
// It verifies that entries retrieval without lock works correctly.
//
// Params:
//   - t: the testing context.
func Test_LogBuffer_entriesLocked(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// entries is the number of entries to add.
		entries int
		// bufSize is the buffer size.
		bufSize int
		// wantCount is the expected entry count.
		wantCount int
	}{
		{
			name:      "empty_buffer",
			entries:   0,
			bufSize:   10,
			wantCount: 0,
		},
		{
			name:      "partial_buffer",
			entries:   3,
			bufSize:   10,
			wantCount: 3,
		},
		{
			name:      "full_buffer",
			entries:   10,
			bufSize:   10,
			wantCount: 10,
		},
		{
			name:      "wrapped_buffer",
			entries:   15,
			bufSize:   10,
			wantCount: 10,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create buffer.
			buf := NewLogBuffer(tt.bufSize)

			// Add entries.
			for i := range tt.entries {
				buf.Add(model.LogEntry{Message: string(rune('A' + i%26))})
			}

			// Lock and call internal method.
			buf.mu.RLock()
			result := buf.entriesLocked()
			buf.mu.RUnlock()

			// Verify count.
			if tt.wantCount == 0 {
				assert.Nil(t, result)
			} else {
				assert.Len(t, result, tt.wantCount)
			}
		})
	}
}
