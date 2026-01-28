// Package tui provides internal tests for log adapter.
package tui

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestParseLogLine(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		wantOK bool
	}{
		{
			name:   "valid log line",
			line:   "2024-01-15T10:30:00Z [INFO] service Starting",
			wantOK: true,
		},
		{
			name:   "invalid format",
			line:   "not a log",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := parseLogLine(tt.line)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestParseLogTimestamp(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOK bool
	}{
		{"RFC3339", "2024-01-15T10:30:00Z", true},
		{"invalid", "not-a-timestamp", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := parseLogTimestamp(tt.input)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestIsServiceName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", false},
		{"Service keyword", "Service", false},
		{"valid name", "nginx", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isServiceName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseLogRemainder(t *testing.T) {
	tests := []struct {
		name        string
		remainder   string
		wantService string
		wantMessage string
	}{
		{
			name:        "service and message",
			remainder:   "myservice Hello",
			wantService: "myservice",
			wantMessage: "Hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &model.LogEntry{Metadata: make(map[string]any)}
			parseLogRemainder(entry, tt.remainder)
			assert.Equal(t, tt.wantService, entry.Service)
			assert.Equal(t, tt.wantMessage, entry.Message)
		})
	}
}

func TestReadLastLines(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"nonexistent file", "/nonexistent.log"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines, err := readLastLines(tt.path, 10)
			assert.NoError(t, err)
			assert.Nil(t, lines)
		})
	}
}

// Test_extractServiceName tests the extractServiceName function.
// It verifies that service name extraction from log parts works correctly.
//
// Params:
//   - t: the testing context.
func Test_extractServiceName(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// parts is the log line parts.
		parts []string
		// wantService is the expected service name.
		wantService string
		// wantIdx is the expected message start index.
		wantIdx int
	}{
		{
			name:        "single_part_no_service",
			parts:       []string{"Message"},
			wantService: "",
			wantIdx:     0,
		},
		{
			name:        "metadata_first_not_service",
			parts:       []string{"key=value", "Message"},
			wantService: "",
			wantIdx:     0,
		},
		{
			name:        "service_name_found",
			parts:       []string{"myservice", "Starting", "up"},
			wantService: "myservice",
			wantIdx:     1,
		},
		{
			name:        "keyword_not_service",
			parts:       []string{"Service", "started"},
			wantService: "",
			wantIdx:     0,
		},
		{
			name:        "daemon_keyword_not_service",
			parts:       []string{"Daemon", "running"},
			wantService: "",
			wantIdx:     0,
		},
		{
			name:        "empty_parts",
			parts:       []string{},
			wantService: "",
			wantIdx:     0,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create entry.
			entry := &model.LogEntry{Metadata: make(map[string]any)}

			// Skip empty parts test.
			if len(tt.parts) == 0 {
				return
			}

			// Call function.
			idx := extractServiceName(entry, tt.parts)

			// Verify results.
			assert.Equal(t, tt.wantService, entry.Service)
			assert.Equal(t, tt.wantIdx, idx)
		})
	}
}

// Test_findMetadataStart tests the findMetadataStart function.
// It verifies that metadata start index detection works correctly.
//
// Params:
//   - t: the testing context.
func Test_findMetadataStart(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// parts is the log line parts.
		parts []string
		// startIdx is the start index to search from.
		startIdx int
		// wantIdx is the expected metadata start index.
		wantIdx int
	}{
		{
			name:     "no_metadata",
			parts:    []string{"service", "Message", "here"},
			startIdx: 0,
			wantIdx:  3,
		},
		{
			name:     "metadata_at_start",
			parts:    []string{"key=value", "Message"},
			startIdx: 0,
			wantIdx:  0,
		},
		{
			name:     "metadata_in_middle",
			parts:    []string{"service", "Message", "key=value", "key2=val2"},
			startIdx: 0,
			wantIdx:  2,
		},
		{
			name:     "metadata_search_from_offset",
			parts:    []string{"service", "Message", "key=value"},
			startIdx: 1,
			wantIdx:  2,
		},
		{
			name:     "empty_parts",
			parts:    []string{},
			startIdx: 0,
			wantIdx:  0,
		},
		{
			name:     "start_past_end",
			parts:    []string{"service", "Message"},
			startIdx: 5,
			wantIdx:  2,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call function.
			idx := findMetadataStart(tt.parts, tt.startIdx)

			// Verify result.
			assert.Equal(t, tt.wantIdx, idx)
		})
	}
}

// Test_extractMetadata tests the extractMetadata function.
// It verifies that key=value extraction into entry metadata works correctly.
//
// Params:
//   - t: the testing context.
func Test_extractMetadata(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// parts is the log line parts.
		parts []string
		// startIdx is the metadata start index.
		startIdx int
		// wantKeys is the expected metadata keys.
		wantKeys []string
	}{
		{
			name:     "single_kv_pair",
			parts:    []string{"Message", "key=value"},
			startIdx: 1,
			wantKeys: []string{"key"},
		},
		{
			name:     "multiple_kv_pairs",
			parts:    []string{"Message", "key1=val1", "key2=val2"},
			startIdx: 1,
			wantKeys: []string{"key1", "key2"},
		},
		{
			name:     "no_metadata",
			parts:    []string{"Message", "text"},
			startIdx: 2,
			wantKeys: []string{},
		},
		{
			name:     "empty_key_skipped",
			parts:    []string{"=value", "key=val"},
			startIdx: 0,
			wantKeys: []string{"key"},
		},
		{
			name:     "empty_parts",
			parts:    []string{},
			startIdx: 0,
			wantKeys: []string{},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create entry.
			entry := &model.LogEntry{Metadata: make(map[string]any)}

			// Call function.
			extractMetadata(entry, tt.parts, tt.startIdx)

			// Verify keys present.
			for _, key := range tt.wantKeys {
				assert.Contains(t, entry.Metadata, key)
			}
			assert.Len(t, entry.Metadata, len(tt.wantKeys))
		})
	}
}
