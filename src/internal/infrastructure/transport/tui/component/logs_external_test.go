// Package component_test provides black-box tests for the component package.
package component_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/component"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestNewLogsPanel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard_terminal", 80, 24},
		{"wide_terminal", 160, 50},
		{"narrow_terminal", 40, 20},
		{"small_terminal", 20, 10},
		{"minimum_size", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(tt.width, tt.height)
			assert.Equal(t, tt.width, panel.Width())
			assert.Equal(t, tt.height, panel.Height())
		})
	}
}

func TestLogsPanel_SetSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		initW int
		initH int
		newW  int
		newH  int
	}{
		{"increase_size", 80, 24, 120, 30},
		{"decrease_size", 120, 30, 80, 24},
		{"width_only", 80, 24, 100, 24},
		{"height_only", 80, 24, 80, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(tt.initW, tt.initH)
			panel.SetSize(tt.newW, tt.newH)
			assert.Equal(t, tt.newW, panel.Width())
			assert.Equal(t, tt.newH, panel.Height())
		})
	}
}

func TestLogsPanel_Focus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setFocused    bool
		wantFocused   bool
		toggleAgain   bool
		wantAfterToggle bool
	}{
		{"initial_unfocused_then_focus", true, true, false, true},
		{"initial_unfocused_then_focus_then_unfocus", true, true, true, false},
		{"stay_unfocused", false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)

			// Initial state should be unfocused.
			assert.False(t, panel.Focused())

			// Apply first focus state.
			panel.SetFocused(tt.setFocused)
			assert.Equal(t, tt.wantFocused, panel.Focused())

			// Toggle if requested.
			if tt.toggleAgain {
				panel.SetFocused(!tt.setFocused)
				assert.Equal(t, tt.wantAfterToggle, panel.Focused())
			}
		})
	}
}

func TestLogsPanel_SetEntries(t *testing.T) {
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
			name: "single_info_entry",
			entries: []model.LogEntry{
				{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test-service",
					Message:   "Service started",
				},
			},
		},
		{
			name: "multiple_levels",
			entries: []model.LogEntry{
				{Timestamp: time.Now(), Level: "INFO", Service: "svc1", Message: "Info message"},
				{Timestamp: time.Now(), Level: "WARN", Service: "svc2", Message: "Warning message"},
				{Timestamp: time.Now(), Level: "ERROR", Service: "svc3", Message: "Error message"},
				{Timestamp: time.Now(), Level: "DEBUG", Service: "svc4", Message: "Debug message"},
			},
		},
		{
			name: "with_metadata",
			entries: []model.LogEntry{
				{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "api",
					Message:   "Request received",
					Metadata: map[string]any{
						"method": "GET",
						"path":   "/api/users",
						"status": 200,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetEntries(tt.entries)
			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestLogsPanel_AddEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		maxSize    int
		entryCount int
	}{
		{"add_single_entry", 10, 1},
		{"add_multiple_entries", 10, 5},
		{"add_entries_at_limit", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetMaxSize(tt.maxSize)

			for i := range tt.entryCount {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "Entry " + string(rune('0'+i)),
				})
			}

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestLogsPanel_AddEntry_ExceedsBuffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		maxSize    int
		entryCount int
	}{
		{"double_buffer_size", 5, 10},
		{"triple_buffer_size", 5, 15},
		{"large_overflow", 10, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetMaxSize(tt.maxSize)

			// Add more entries than the buffer size.
			for range tt.entryCount {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "test message",
				})
			}

			// Panel should still render without panic.
			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestLogsPanel_Init(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard_size", 80, 24},
		{"wide_panel", 160, 50},
		{"narrow_panel", 40, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(tt.width, tt.height)
			cmd := panel.Init()
			assert.Nil(t, cmd)
		})
	}
}

func TestLogsPanel_Update_Unfocused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{"key_down", tea.KeyMsg{Type: tea.KeyDown}},
		{"key_up", tea.KeyMsg{Type: tea.KeyUp}},
		{"key_pgdn", tea.KeyMsg{Type: tea.KeyPgDown}},
		{"key_pgup", tea.KeyMsg{Type: tea.KeyPgUp}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetFocused(false)

			// When unfocused, update should not process key messages.
			updatedPanel, cmd := panel.Update(tt.msg)
			assert.NotNil(t, updatedPanel)
			assert.Nil(t, cmd)
		})
	}
}

func TestLogsPanel_Update_Focused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{"key_down", tea.KeyMsg{Type: tea.KeyDown}},
		{"key_up", tea.KeyMsg{Type: tea.KeyUp}},
		{"key_pgdn", tea.KeyMsg{Type: tea.KeyPgDown}},
		{"key_pgup", tea.KeyMsg{Type: tea.KeyPgUp}},
		{"key_home", tea.KeyMsg{Type: tea.KeyHome}},
		{"key_end", tea.KeyMsg{Type: tea.KeyEnd}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetFocused(true)

			// Add some entries for scrolling.
			for range 50 {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "test message",
				})
			}

			updatedPanel, cmd := panel.Update(tt.msg)
			assert.NotNil(t, updatedPanel)
			_ = cmd // May or may not be nil.
		})
	}
}

func TestLogsPanel_View_Empty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard_size", 80, 24},
		{"wide_panel", 160, 50},
		{"narrow_panel", 40, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(tt.width, tt.height)
			view := panel.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Logs")
		})
	}
}

func TestLogsPanel_View_WithEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		timestamp time.Time
		level     string
		service   string
		message   string
		wantTime  string
	}{
		{
			name:      "error_entry",
			timestamp: time.Date(2025, 1, 1, 12, 30, 45, 0, time.UTC),
			level:     "ERROR",
			service:   "api",
			message:   "Database connection failed",
			wantTime:  "12:30:45",
		},
		{
			name:      "info_entry",
			timestamp: time.Date(2025, 6, 15, 8, 0, 0, 0, time.UTC),
			level:     "INFO",
			service:   "worker",
			message:   "Processing started",
			wantTime:  "08:00:00",
		},
		{
			name:      "warn_entry",
			timestamp: time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			level:     "WARN",
			service:   "scheduler",
			message:   "High memory usage",
			wantTime:  "23:59:59",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.AddEntry(model.LogEntry{
				Timestamp: tt.timestamp,
				Level:     tt.level,
				Service:   tt.service,
				Message:   tt.message,
			})

			view := panel.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Logs")
			assert.Contains(t, view, tt.wantTime)
		})
	}
}

func TestLogsPanel_View_Focused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		focused bool
	}{
		{"focused", true},
		{"unfocused", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetFocused(tt.focused)
			panel.AddEntry(model.LogEntry{
				Timestamp: time.Now(),
				Level:     "INFO",
				Service:   "test",
				Message:   "test",
			})

			view := panel.View()
			assert.NotEmpty(t, view)
			// Border color should differ based on focus state.
		})
	}
}

func TestLogsPanel_ScrollToBottom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entryCount int
	}{
		{"few_entries", 20},
		{"many_entries", 100},
		{"large_buffer", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)

			// Add more entries than can fit on screen.
			for i := range tt.entryCount {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "Entry " + string(rune('0'+(i%10))),
				})
			}

			panel.ScrollToBottom()
			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestLogsPanel_ScrollToTop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		entryCount int
	}{
		{"few_entries", 20},
		{"many_entries", 50},
		{"large_buffer", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)

			// Add entries.
			for range tt.entryCount {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "test",
				})
			}

			panel.ScrollToTop()
			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestLogsPanel_SetMaxSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		maxSize int
	}{
		{"default", 100},
		{"small", 10},
		{"large", 1000},
		{"zero", 0},       // Should use default.
		{"negative", -10}, // Should use default.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetMaxSize(tt.maxSize)

			// Add entries.
			for range 20 {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "test",
				})
			}

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestLogsPanel_DifferentLogLevels(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level string
	}{
		{"info", "INFO"},
		{"warn", "WARN"},
		{"warning", "WARNING"},
		{"error", "ERROR"},
		{"err", "ERR"},
		{"debug", "DEBUG"},
		{"unknown", "CUSTOM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.AddEntry(model.LogEntry{
				Timestamp: time.Now(),
				Level:     tt.level,
				Service:   "test",
				Message:   "test message",
			})

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestLogsPanel_LongServiceNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
	}{
		{"short_name", "api"},
		{"medium_name", "user-service"},
		{"long_name", "very-long-service-name-that-exceeds-column-width"},
		{"empty_name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.AddEntry(model.LogEntry{
				Timestamp: time.Now(),
				Level:     "INFO",
				Service:   tt.serviceName,
				Message:   "test message",
			})

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestLogsPanel_Metadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata map[string]any
	}{
		{"no_metadata", nil},
		{"string_metadata", map[string]any{"key": "value"}},
		{"int_metadata", map[string]any{"count": 42}},
		{"float_metadata", map[string]any{"ratio": 3.14}},
		{"bool_metadata", map[string]any{"enabled": true}},
		{"mixed_metadata", map[string]any{
			"string": "value",
			"int":    123,
			"float":  2.5,
			"bool":   false,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.AddEntry(model.LogEntry{
				Timestamp: time.Now(),
				Level:     "INFO",
				Service:   "test",
				Message:   "test",
				Metadata:  tt.metadata,
			})

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestLogsPanel_SetFocused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		focused bool
	}{
		{"focus_panel", true},
		{"unfocus_panel", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetFocused(tt.focused)
			assert.Equal(t, tt.focused, panel.Focused())
		})
	}
}

func TestLogsPanel_Focused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		focused bool
	}{
		{"when_focused", true},
		{"when_unfocused", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetFocused(tt.focused)
			result := panel.Focused()
			assert.Equal(t, tt.focused, result)
		})
	}
}

func TestLogsPanel_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		focused bool
	}{
		{"update_when_focused", true},
		{"update_when_unfocused", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetFocused(tt.focused)

			// Add some entries.
			for range 10 {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "test",
				})
			}

			msg := tea.KeyMsg{Type: tea.KeyDown}
			updatedPanel, cmd := panel.Update(msg)
			assert.NotNil(t, updatedPanel)

			// When unfocused, cmd should be nil.
			if !tt.focused {
				assert.Nil(t, cmd)
			}
		})
	}
}

func TestLogsPanel_View(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		focused    bool
		entryCount int
	}{
		{"empty_unfocused", false, 0},
		{"empty_focused", true, 0},
		{"with_entries_unfocused", false, 5},
		{"with_entries_focused", true, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, 24)
			panel.SetFocused(tt.focused)

			// Add entries.
			for range tt.entryCount {
				panel.AddEntry(model.LogEntry{
					Timestamp: time.Now(),
					Level:     "INFO",
					Service:   "test",
					Message:   "test message",
				})
			}

			view := panel.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Logs")
			assert.Contains(t, view, "+")
		})
	}
}

func TestLogsPanel_Height(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		height int
	}{
		{"small", 10},
		{"standard", 24},
		{"large", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewLogsPanel(80, tt.height)
			assert.Equal(t, tt.height, panel.Height())
		})
	}
}

func TestLogsPanel_Width(t *testing.T) {
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
			panel := component.NewLogsPanel(tt.width, 24)
			assert.Equal(t, tt.width, panel.Width())
		})
	}
}
