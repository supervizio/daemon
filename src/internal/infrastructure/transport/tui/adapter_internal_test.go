// Package tui provides internal (white-box) tests for the TUI adapter.
// Internal tests can access unexported functions and fields.
package tui

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// TestIsServiceName verifies service name detection logic.
func TestIsServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		expect bool
	}{
		{name: "empty string", input: "", expect: false},
		{name: "common starter Service", input: "Service", expect: false},
		{name: "common starter Daemon", input: "Daemon", expect: false},
		{name: "common starter Failed", input: "Failed", expect: false},
		{name: "valid service name", input: "nginx", expect: true},
		{name: "valid service with dash", input: "my-service", expect: true},
		{name: "valid service with underscore", input: "my_service", expect: true},
	}

	// Run test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := isServiceName(tt.input)
			assert.Equal(t, tt.expect, result)
		})
	}
}

// TestParseLogTimestamp verifies timestamp parsing.
func TestParseLogTimestamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		wantOK bool
	}{
		{name: "RFC3339 format", input: "2024-01-15T10:30:00Z", wantOK: true},
		{name: "RFC3339 with timezone", input: "2024-01-15T10:30:00+01:00", wantOK: true},
		{name: "alternative format", input: "2024-01-15T10:30:00Z", wantOK: true},
		{name: "invalid format", input: "not-a-timestamp", wantOK: false},
		{name: "partial timestamp", input: "2024-01-15", wantOK: false},
	}

	// Run test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, ok := parseLogTimestamp(tt.input)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

// TestParseLogLine verifies full log line parsing.
func TestParseLogLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		line        string
		wantOK      bool
		wantLevel   string
		wantService string
		wantMessage string
	}{
		{
			name:        "full log line",
			line:        "2024-01-15T10:30:00Z [INFO] myservice Starting server",
			wantOK:      true,
			wantLevel:   "INFO",
			wantService: "myservice",
			wantMessage: "Starting server",
		},
		{
			name:        "log line without service",
			line:        "2024-01-15T10:30:00Z [ERROR] Service failed to start",
			wantOK:      true,
			wantLevel:   "ERROR",
			wantService: "",
			wantMessage: "Service failed to start",
		},
		{
			name:   "invalid format",
			line:   "not a valid log line",
			wantOK: false,
		},
		{
			name:      "log with metadata",
			line:      "2024-01-15T10:30:00Z [WARN] app Warning message key=value",
			wantOK:    true,
			wantLevel: "WARN",
		},
	}

	// Run test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entry, ok := parseLogLine(tt.line)
			assert.Equal(t, tt.wantOK, ok)

			// Verify parsed values if successful.
			if ok {
				assert.Equal(t, tt.wantLevel, entry.Level)

				// Check service if expected.
				if tt.wantService != "" {
					assert.Equal(t, tt.wantService, entry.Service)
				}

				// Check message if expected.
				if tt.wantMessage != "" {
					assert.Equal(t, tt.wantMessage, entry.Message)
				}
			}
		})
	}
}

// TestParseLogRemainder verifies remainder parsing.
func TestParseLogRemainder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		remainder   string
		wantService string
		wantMessage string
		wantMeta    map[string]any
	}{
		{
			name:        "service and message",
			remainder:   "myservice Hello world",
			wantService: "myservice",
			wantMessage: "Hello world",
		},
		{
			name:        "message only (starts with common word)",
			remainder:   "Service starting up",
			wantService: "",
			wantMessage: "Service starting up",
		},
		{
			name:        "with metadata",
			remainder:   "app Starting port=8080 host=localhost",
			wantService: "app",
			wantMessage: "Starting",
			wantMeta:    map[string]any{"port": "8080", "host": "localhost"},
		},
		{
			name:        "empty remainder",
			remainder:   "",
			wantService: "",
			wantMessage: "",
		},
	}

	// Run test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			entry := &model.LogEntry{
				Metadata: make(map[string]any),
			}
			parseLogRemainder(entry, tt.remainder)

			assert.Equal(t, tt.wantService, entry.Service)
			assert.Equal(t, tt.wantMessage, entry.Message)

			// Check metadata if expected.
			if tt.wantMeta != nil {
				for k, v := range tt.wantMeta {
					assert.Equal(t, v, entry.Metadata[k])
				}
			}
		})
	}
}

// TestLogBufferRingBehavior verifies ring buffer wrap-around.
func TestLogBufferRingBehavior(t *testing.T) {
	t.Parallel()

	buf := NewLogBuffer(3)

	// Add 5 entries to a buffer of size 3.
	for i := range 5 {
		buf.Add(model.LogEntry{Message: string(rune('A' + i))})
	}

	// Should only have last 3 entries (C, D, E).
	entries := buf.Entries()
	assert.Len(t, entries, 3)
	assert.Equal(t, "C", entries[0].Message)
	assert.Equal(t, "D", entries[1].Message)
	assert.Equal(t, "E", entries[2].Message)
}

// TestLogBufferLevelCounts verifies level counting.
func TestLogBufferLevelCounts(t *testing.T) {
	t.Parallel()

	buf := NewLogBuffer(10)

	// Add various levels.
	buf.Add(model.LogEntry{Level: "INFO"})
	buf.Add(model.LogEntry{Level: "INFO"})
	buf.Add(model.LogEntry{Level: "WARN"})
	buf.Add(model.LogEntry{Level: "WARNING"})
	buf.Add(model.LogEntry{Level: "ERROR"})
	buf.Add(model.LogEntry{Level: "ERR"})

	summary := buf.Summary()
	assert.Equal(t, 2, summary.InfoCount)
	assert.Equal(t, 2, summary.WarnCount)
	assert.Equal(t, 2, summary.ErrorCount)
}

// TestReadLastLinesNonExistent verifies handling of missing files.
func TestReadLastLinesNonExistent(t *testing.T) {
	t.Parallel()

	lines, err := readLastLines("/nonexistent/path/file.log", 10)
	assert.NoError(t, err)
	assert.Nil(t, lines)
}
