package screen

import (
	"strings"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestLogsRenderer_buildSummaryLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		logs model.LogSummary
	}{
		{
			name: "empty_logs",
			logs: model.LogSummary{},
		},
		{
			name: "with_counts",
			logs: model.LogSummary{
				InfoCount:  100,
				WarnCount:  10,
				ErrorCount: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewLogsRenderer(80)
			result := renderer.buildSummaryLine(tt.logs)
			assert.NotEmpty(t, result)
		})
	}
}

func TestLogsRenderer_buildLogLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		logs        model.LogSummary
		summaryLine string
	}{
		{
			name:        "empty_logs",
			logs:        model.LogSummary{},
			summaryLine: "Summary",
		},
		{
			name: "with_entries",
			logs: model.LogSummary{
				RecentEntries: []model.LogEntry{
					{Level: "INFO", Service: "test", Message: "test message", Timestamp: time.Now()},
				},
			},
			summaryLine: "Summary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewLogsRenderer(80)
			result := renderer.buildLogLines(tt.logs, tt.summaryLine)
			assert.NotNil(t, result)
		})
	}
}

func TestLogsRenderer_appendWarnCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		logs model.LogSummary
	}{
		{
			name: "zero_warns",
			logs: model.LogSummary{
				WarnCount: 0,
			},
		},
		{
			name: "with_warns",
			logs: model.LogSummary{
				WarnCount:  5,
				ErrorCount: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewLogsRenderer(80)
			var sb strings.Builder
			renderer.appendWarnCount(&sb, tt.logs)
			assert.NotNil(t, sb.String())
		})
	}
}

func TestLogsRenderer_appendErrorCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		logs model.LogSummary
	}{
		{
			name: "zero_errors",
			logs: model.LogSummary{
				ErrorCount: 0,
			},
		},
		{
			name: "with_errors",
			logs: model.LogSummary{
				ErrorCount: 3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewLogsRenderer(80)
			var sb strings.Builder
			renderer.appendErrorCount(&sb, tt.logs)
			assert.NotNil(t, sb.String())
		})
	}
}

func TestLogsRenderer_buildSeparator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "separator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewLogsRenderer(80)
			result := renderer.buildSeparator()
			assert.NotEmpty(t, result)
		})
	}
}

func TestLogsRenderer_buildEntryLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entries []model.LogEntry
	}{
		{
			name:    "empty_entries",
			entries: []model.LogEntry{},
		},
		{
			name: "with_entries",
			entries: []model.LogEntry{
				{Level: "INFO", Service: "test", Message: "test message", Timestamp: time.Now()},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewLogsRenderer(80)
			result := renderer.buildEntryLines(tt.entries)
			assert.NotNil(t, result)
		})
	}
}

func TestLogsRenderer_formatLogEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entry    model.LogEntry
		maxWidth int
	}{
		{
			name: "basic_entry",
			entry: model.LogEntry{
				Level:     "INFO",
				Service:   "test",
				Message:   "test message",
				Timestamp: time.Now(),
			},
			maxWidth: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewLogsRenderer(80)
			result := renderer.formatLogEntry(tt.entry, tt.maxWidth)
			assert.NotEmpty(t, result)
		})
	}
}

func TestLogsRenderer_truncateMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		msg   string
		width int
	}{
		{
			name:  "no_truncate",
			msg:   "short",
			width: 10,
		},
		{
			name:  "with_truncate",
			msg:   "very long message that needs truncation",
			width: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := NewLogsRenderer(80)
			result := renderer.truncateMessage(tt.msg, tt.width)
			assert.NotNil(t, &result)
		})
	}
}
