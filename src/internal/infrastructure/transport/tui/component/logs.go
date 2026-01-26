// Package component provides reusable Bubble Tea components.
package component

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// Scrollbar characters.
const (
	scrollTrack = "│"
	scrollThumb = "┃"
)

// LogsPanel is a scrollable logs viewport with vertical scrollbar.
type LogsPanel struct {
	viewport viewport.Model
	theme    ansi.Theme
	width    int
	height   int
	entries  []model.LogEntry
	maxSize  int // Maximum buffer size for display.
	focused  bool
	title    string
}

// DefaultLogBufferSize is the default maximum log entries to display.
const DefaultLogBufferSize = 100

// NewLogsPanel creates a new logs panel.
func NewLogsPanel(width, height int) LogsPanel {
	// -3 for borders (left border, right border, scrollbar)
	vp := viewport.New(width-3, height-2)

	return LogsPanel{
		viewport: vp,
		theme:    ansi.DefaultTheme(),
		width:    width,
		height:   height,
		entries:  make([]model.LogEntry, 0),
		maxSize:  DefaultLogBufferSize,
		title:    "Logs",
	}
}

// SetMaxSize sets the maximum buffer size for display indicator.
func (l *LogsPanel) SetMaxSize(maxSize int) {
	l.maxSize = maxSize
}

// SetSize updates the panel dimensions.
func (l *LogsPanel) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.viewport.Width = width - 3   // -3 for left border, right border, scrollbar
	l.viewport.Height = height - 2 // -2 for top/bottom borders
	l.updateContent()
}

// SetFocused sets the focus state.
func (l *LogsPanel) SetFocused(focused bool) {
	l.focused = focused
}

// IsFocused returns whether the panel is focused.
func (l LogsPanel) IsFocused() bool {
	return l.focused
}

// SetEntries updates the log entries.
func (l *LogsPanel) SetEntries(entries []model.LogEntry) {
	l.entries = entries
	l.updateContent()
}

// AddEntry adds a new log entry and scrolls to bottom.
func (l *LogsPanel) AddEntry(entry model.LogEntry) {
	l.entries = append(l.entries, entry)
	l.updateContent()
	l.viewport.GotoBottom()
}

// updateContent rebuilds the viewport content.
func (l *LogsPanel) updateContent() {
	var sb strings.Builder

	// Fixed column widths (no ANSI codes in width calculation).
	const (
		timeCol    = 9  // "HH:MM:SS "
		levelCol   = 8  // "[LEVEL] " (7 + space)
		serviceCol = 13 // "servicename  " (12 + space)
	)

	// Content area width (viewport width).
	contentWidth := l.viewport.Width
	msgWidth := contentWidth - timeCol - levelCol - serviceCol
	if msgWidth < 10 {
		msgWidth = 10
	}

	for _, entry := range l.entries {
		// Time - fixed 8 chars + space.
		ts := entry.Timestamp.Format("15:04:05")

		// Level - get display string and color separately.
		levelStr, levelColor := l.getLevelInfo(entry.Level)

		// Service - fixed 12 chars.
		service := entry.Service
		if service == "" {
			service = "daemon"
		}
		if len(service) > 12 {
			service = service[:11] + "…"
		}

		// Message with metadata.
		msg := entry.Message
		if msg == "" {
			msg = entry.EventType
		}
		if len(entry.Metadata) > 0 {
			msg += " " + l.formatMetadata(entry.Metadata)
		}
		if len(msg) > msgWidth {
			msg = msg[:msgWidth-1] + "…"
		}

		// Build line with proper padding (pad BEFORE adding colors).
		// Format: "HH:MM:SS [LEVEL] service      message"
		line := fmt.Sprintf("%s%s%s %s[%s%-5s%s]%s %-12s %s",
			l.theme.Muted, ts, ansi.Reset,
			levelColor, ansi.Reset, levelStr, levelColor, ansi.Reset,
			service,
			msg,
		)

		sb.WriteString(line)
		sb.WriteString("\n")
	}

	l.viewport.SetContent(sb.String())
}

// getLevelInfo returns the level string and its color separately.
func (l *LogsPanel) getLevelInfo(level string) (string, string) {
	switch strings.ToUpper(level) {
	case "ERROR", "ERR":
		return "ERROR", l.theme.Error
	case "WARN", "WARNING":
		return "WARN", l.theme.Warning
	case "INFO":
		return "INFO", l.theme.Primary
	case "DEBUG":
		return "DEBUG", l.theme.Muted
	default:
		return level, l.theme.Muted
	}
}

// formatMetadata formats metadata as key=value pairs.
func (l *LogsPanel) formatMetadata(meta map[string]any) string {
	if len(meta) == 0 {
		return ""
	}

	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, meta[k]))
	}
	return strings.Join(parts, " ")
}

// Init initializes the component.
func (l LogsPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (l LogsPanel) Update(msg tea.Msg) (LogsPanel, tea.Cmd) {
	if !l.focused {
		return l, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "home", "g":
			l.viewport.GotoTop()
		case "end", "G":
			l.viewport.GotoBottom()
		case "pgup", "ctrl+u":
			l.viewport.HalfPageUp()
		case "pgdown", "ctrl+d":
			l.viewport.HalfPageDown()
		case "up", "k":
			l.viewport.ScrollUp(1)
		case "down", "j":
			l.viewport.ScrollDown(1)
		default:
			l.viewport, cmd = l.viewport.Update(msg)
		}
	case tea.MouseMsg:
		l.viewport, cmd = l.viewport.Update(msg)
	}

	return l, cmd
}

// View renders the logs panel with border and vertical scrollbar.
func (l LogsPanel) View() string {
	var sb strings.Builder

	// Border color based on focus.
	borderColor := l.theme.Muted
	if l.focused {
		borderColor = l.theme.Primary
	}

	// Inner width = total width - 2 borders - 1 scrollbar.
	innerWidth := l.width - 3

	// === Top border ===
	// Format: ╭─ Logs ────────────────── 50% ─╮
	titlePart := fmt.Sprintf("─ %s%s%s ", l.theme.Header, l.title, borderColor)
	scrollPart := l.scrollIndicator()

	// Calculate dashes needed.
	titleVisLen := 3 + len(l.title) // "─ " + title + " "
	scrollVisLen := widget.VisibleLen(scrollPart)
	dashCount := innerWidth - titleVisLen - scrollVisLen - 1 // -1 for final "─"
	if dashCount < 0 {
		dashCount = 0
	}

	sb.WriteString(borderColor)
	sb.WriteString("╭")
	sb.WriteString(titlePart)
	sb.WriteString(strings.Repeat("─", dashCount))
	sb.WriteString(" ")
	sb.WriteString(scrollPart)
	sb.WriteString(borderColor) // Re-apply border color after scrollPart reset.
	sb.WriteString("─╮")
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")

	// === Content lines with vertical scrollbar ===
	content := l.viewport.View()
	lines := strings.Split(content, "\n")

	// Calculate scrollbar.
	scrollbarChars := l.renderVerticalScrollbar()

	for i := 0; i < l.viewport.Height; i++ {
		sb.WriteString(borderColor)
		sb.WriteString("│")
		sb.WriteString(ansi.Reset)

		// Content.
		if i < len(lines) {
			line := lines[i]
			visLen := widget.VisibleLen(line)
			sb.WriteString(line)
			if visLen < innerWidth {
				sb.WriteString(strings.Repeat(" ", innerWidth-visLen))
			}
		} else {
			sb.WriteString(strings.Repeat(" ", innerWidth))
		}

		// Scrollbar character.
		sb.WriteString(borderColor)
		if i < len(scrollbarChars) {
			sb.WriteString(scrollbarChars[i])
		} else {
			sb.WriteString(scrollTrack)
		}
		sb.WriteString("│")
		sb.WriteString(ansi.Reset)
		sb.WriteString("\n")
	}

	// === Bottom border ===
	sb.WriteString(borderColor)
	sb.WriteString("╰")
	sb.WriteString(strings.Repeat("─", innerWidth+1)) // +1 for scrollbar column
	sb.WriteString("╯")
	sb.WriteString(ansi.Reset)

	return sb.String()
}

// scrollIndicator returns the entry count as [ count / max ].
func (l LogsPanel) scrollIndicator() string {
	count := len(l.entries)
	max := l.maxSize
	if max <= 0 {
		max = DefaultLogBufferSize
	}

	return fmt.Sprintf("%s[ %d / %d ]%s", l.theme.Muted, count, max, ansi.Reset)
}

// renderVerticalScrollbar returns the scrollbar characters for each row.
func (l LogsPanel) renderVerticalScrollbar() []string {
	height := l.viewport.Height
	totalLines := len(l.entries)

	if totalLines <= height {
		// No scrolling needed - no thumb.
		result := make([]string, height)
		for i := range result {
			result[i] = scrollTrack
		}
		return result
	}

	// Calculate thumb size (minimum 1).
	ratio := float64(height) / float64(totalLines)
	thumbSize := int(float64(height) * ratio)
	if thumbSize < 1 {
		thumbSize = 1
	}

	// Calculate thumb position.
	scrollableHeight := height - thumbSize
	scrollPercent := l.viewport.ScrollPercent()
	thumbPos := int(float64(scrollableHeight) * scrollPercent)

	// Build scrollbar.
	result := make([]string, height)
	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			result[i] = scrollThumb
		} else {
			result[i] = scrollTrack
		}
	}
	return result
}

// Height returns the panel height.
func (l LogsPanel) Height() int {
	return l.height
}

// Width returns the panel width.
func (l LogsPanel) Width() int {
	return l.width
}

// ScrollToBottom scrolls to the bottom of the logs.
func (l *LogsPanel) ScrollToBottom() {
	l.viewport.GotoBottom()
}

// ScrollToTop scrolls to the top of the logs.
func (l *LogsPanel) ScrollToTop() {
	l.viewport.GotoTop()
}
