// Package component provides internal white-box tests.
package component

import (
	"strings"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestLogsPanel_formatLogLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entry    model.LogEntry
		msgWidth int
	}{
		{
			name: "info_level",
			entry: model.LogEntry{
				Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "INFO",
				Service:   "test",
				Message:   "test message",
			},
			msgWidth: 50,
		},
		{
			name: "error_level",
			entry: model.LogEntry{
				Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "ERROR",
				Service:   "service1",
				Message:   "error occurred",
			},
			msgWidth: 50,
		},
		{
			name: "with_metadata",
			entry: model.LogEntry{
				Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "WARN",
				Service:   "test",
				Message:   "warning",
				Metadata: map[string]any{
					"key1": "value1",
					"key2": 42,
				},
			},
			msgWidth: 50,
		},
		{
			name: "long_service_name",
			entry: model.LogEntry{
				Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "INFO",
				Service:   "very-long-service-name-that-exceeds-limit",
				Message:   "test",
			},
			msgWidth: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, 24)
			line := panel.formatLogLine(tt.entry, tt.msgWidth)
			assert.NotEmpty(t, line)
			assert.Contains(t, line, "12:00:00")
		})
	}
}

func TestLogsPanel_formatServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service string
		want    string
	}{
		{"empty_service", "", "daemon"},
		{"short_name", "test", "test"},
		{"long_name", "very-long-service-name-exceeds-limit", "very-lon..."},
		{"exact_limit", "service12345", "service12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, 24)
			result := panel.formatServiceName(tt.service)
			if tt.service == "" {
				assert.Equal(t, "daemon", result)
			} else {
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestLogsPanel_getLevelInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		level     string
		wantLevel string
	}{
		{"error_upper", "ERROR", "ERROR"},
		{"error_lower", "error", "ERROR"},
		{"err_short", "ERR", "ERROR"},
		{"warn_upper", "WARN", "WARN"},
		{"warning_full", "WARNING", "WARN"},
		{"info", "INFO", "INFO"},
		{"debug", "DEBUG", "DEBUG"},
		{"unknown", "UNKNOWN", "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, 24)
			levelStr, color := panel.getLevelInfo(tt.level)
			assert.Equal(t, tt.wantLevel, levelStr)
			assert.NotEmpty(t, color)
		})
	}
}

func TestLogsPanel_formatMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		meta map[string]any
		want string
	}{
		{
			name: "empty_metadata",
			meta: map[string]any{},
			want: "",
		},
		{
			name: "nil_metadata",
			meta: nil,
			want: "",
		},
		{
			name: "string_value",
			meta: map[string]any{"key": "value"},
			want: "key=value",
		},
		{
			name: "int_value",
			meta: map[string]any{"count": 42},
			want: "count=42",
		},
		{
			name: "multiple_values",
			meta: map[string]any{"key1": "val1", "key2": 123},
			want: "key", // Should contain keys
		},
		{
			name: "bool_value",
			meta: map[string]any{"enabled": true},
			want: "enabled=true",
		},
		{
			name: "float_value",
			meta: map[string]any{"ratio": 3.14},
			want: "ratio=3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, 24)
			result := panel.formatMetadata(tt.meta)
			if tt.want == "" {
				assert.Empty(t, result)
			} else {
				assert.Contains(t, result, tt.want)
			}
		})
	}
}

func TestLogsPanel_scrollIndicator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entryCount int
		maxSize    int
	}{
		{"empty", 0, 100},
		{"partial", 50, 100},
		{"full", 100, 100},
		{"custom_max", 25, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, 24)
			panel.SetMaxSize(tt.maxSize)
			for range tt.entryCount {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "test",
				})
			}
			indicator := panel.scrollIndicator()
			assert.NotEmpty(t, indicator)
			assert.Contains(t, indicator, "[")
			assert.Contains(t, indicator, "]")
		})
	}
}

func TestLogsPanel_renderVerticalScrollbar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		height     int
		entryCount int
	}{
		{"no_scroll_needed", 20, 5},
		{"scroll_needed", 10, 50},
		{"exact_fit", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, tt.height)
			for range tt.entryCount {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "test",
				})
			}
			scrollbar := panel.renderVerticalScrollbar()
			assert.NotEmpty(t, scrollbar)
			// Scrollbar length should match viewport height.
			assert.LessOrEqual(t, len(scrollbar), tt.height)
		})
	}
}

// mockKeyMsg is a mock for tea.KeyMsg implementing Stringer.
type mockKeyMsg struct {
	str string
}

func (m mockKeyMsg) String() string {
	return m.str
}

func TestLogsPanel_handleKeyMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{"home", "home"},
		{"g_key", "g"},
		{"end", "end"},
		{"G_key", "G"},
		{"page_up", "pgup"},
		{"ctrl_u", "ctrl+u"},
		{"page_down", "pgdown"},
		{"ctrl_d", "ctrl+d"},
		{"up_arrow", "up"},
		{"k_key", "k"},
		{"down_arrow", "down"},
		{"j_key", "j"},
		{"unknown", "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, 24)
			msg := mockKeyMsg{str: tt.key}
			cmd := panel.handleKeyMsg(msg)
			_ = cmd // May or may not be nil.
		})
	}
}

func TestLogsPanel_buildMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		entry    model.LogEntry
		msgWidth int
		want     string
	}{
		{
			name: "normal_message",
			entry: model.LogEntry{
				Message: "test message",
			},
			msgWidth: 50,
			want:     "test message",
		},
		{
			name: "empty_message_with_event_type",
			entry: model.LogEntry{
				Message:   "",
				EventType: "service.started",
			},
			msgWidth: 50,
			want:     "service.started",
		},
		{
			name: "with_metadata",
			entry: model.LogEntry{
				Message:  "test",
				Metadata: map[string]any{"key": "value"},
			},
			msgWidth: 50,
			want:     "test",
		},
		{
			name: "long_message",
			entry: model.LogEntry{
				Message: strings.Repeat("a", 100),
			},
			msgWidth: 20,
			want:     "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, 24)
			result := panel.buildMessage(tt.entry, tt.msgWidth)
			assert.Contains(t, result, tt.want)
		})
	}
}

func TestLogsPanel_updateContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entryCount int
	}{
		{"empty_entries", 0},
		{"single_entry", 1},
		{"multiple_entries", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, 24)

			// Add entries.
			for i := range tt.entryCount {
				panel.entries = append(panel.entries, model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "message " + string(rune('0'+i)),
				})
			}

			// Call updateContent.
			panel.updateContent()

			// Verify viewport has content set.
			view := panel.viewport.View()
			if tt.entryCount > 0 {
				assert.NotEmpty(t, view)
			}
		})
	}
}

func TestLogsPanel_writeMetadataValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value any
		want  string
	}{
		{"string_value", "hello", "hello"},
		{"int_value", 42, "42"},
		{"int64_value", int64(12345678901), "12345678901"},
		{"uint64_value", uint64(9876543210), "9876543210"},
		{"float64_value", 3.14159, "3.14159"},
		{"bool_true", true, "true"},
		{"bool_false", false, "false"},
		{"complex_type", struct{ Name string }{Name: "test"}, "{test}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, 24)
			var sb strings.Builder
			panel.writeMetadataValue(&sb, tt.value)
			result := sb.String()
			assert.Contains(t, result, tt.want)
		})
	}
}

func TestLogsPanel_renderTopBorder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		entryCount int
	}{
		{"narrow_empty", 40, 0},
		{"standard_empty", 80, 0},
		{"wide_with_entries", 120, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(tt.width, 24)

			// Add entries.
			for range tt.entryCount {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "test",
				})
			}

			var sb strings.Builder
			borderColor := panel.theme.Muted
			innerWidth := tt.width - logBorderWidth
			panel.renderTopBorder(&sb, borderColor, innerWidth)

			result := sb.String()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "+")
			assert.Contains(t, result, "Logs")
		})
	}
}

func TestLogsPanel_renderContentLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entryCount int
		height     int
	}{
		{"empty_content", 0, 10},
		{"few_entries", 3, 10},
		{"many_entries", 20, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(80, tt.height)

			// Add entries.
			for range tt.entryCount {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "test message",
				})
			}

			var sb strings.Builder
			borderColor := panel.theme.Muted
			innerWidth := 80 - logBorderWidth
			panel.renderContentLines(&sb, borderColor, innerWidth)

			result := sb.String()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "|")
		})
	}
}

func TestLogsPanel_renderBottomBorder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{"narrow", 40},
		{"standard", 80},
		{"wide", 160},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewLogsPanel(tt.width, 24)

			var sb strings.Builder
			borderColor := panel.theme.Muted
			innerWidth := tt.width - logBorderWidth
			panel.renderBottomBorder(&sb, borderColor, innerWidth)

			result := sb.String()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "+")
		})
	}
}

