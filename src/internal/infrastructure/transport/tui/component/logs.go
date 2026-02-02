// Package component provides reusable Bubble Tea components.
package component

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// Scrollbar characters.
const (
	// scrollTrack is the character for the scrollbar track.
	scrollTrack string = "│"
	// scrollThumb is the character for the scrollbar thumb.
	scrollThumb string = "┃"

	// formatFloatPrecision is the default precision for float formatting.
	formatFloatPrecision int = -1

	// formatFloatBitSize is the bit size for float formatting.
	formatFloatBitSize int = 64

	// metadataBufferSize is the estimated buffer size for metadata formatting.
	metadataBufferSize int = 32

	// decimalBase is the base for decimal number parsing.
	decimalBase int = 10
)

// Border and scrollbar dimensions.
const (
	// logBorderWidth is the total horizontal border width (left + right + scrollbar).
	logBorderWidth int = 3
	// logBorderHeight is the total vertical border height (top + bottom).
	logBorderHeight int = 2

	// timeColWidth is the width of the timestamp column.
	timeColWidth int = 9
	// levelColWidth is the width of the log level column.
	levelColWidth int = 8
	// serviceColWidth is the width of the service name column.
	serviceColWidth int = 13

	// minMsgWidth is the minimum message column width.
	minMsgWidth int = 10
	// levelTextMaxWidth is the maximum width for level text.
	levelTextMaxWidth int = 5
	// serviceNameMaxWidth is the maximum width for service name.
	serviceNameMaxWidth int = 12
	// metadataEstimatedSize is the estimated chars per metadata key-value pair.
	metadataEstimatedSize int = 16
	// lineBufferGrow is the estimated extra chars for ANSI codes in line.
	lineBufferGrow int = 60

	// titleBufferGrow is the pre-allocation size for title buffer.
	titleBufferGrow int = 32

	// titlePrefixLen is the length of "- " + " " surrounding the title.
	titlePrefixLen int = 3

	// scrollbarColumnWidth is the width of the scrollbar column.
	scrollbarColumnWidth int = 1

	// dashCountOffset is the offset for final border character.
	dashCountOffset int = 1

	// minThumbSize is the minimum scrollbar thumb size.
	minThumbSize int = 1
)

// DefaultLogBufferSize is the default maximum log entries to display.
const DefaultLogBufferSize int = 100

// Stringer defines the minimal interface for key message handling (KTN-API-MINIF).
type Stringer interface {
	String() string
}

// LogsPanel is a scrollable logs viewport with vertical scrollbar.
//
// It displays log entries with timestamp, level, service name, and message.
// Supports automatic truncation, color coding by level, and metadata formatting.
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

// NewLogsPanel creates a new logs panel.
//
// Params:
//   - width: panel width including borders
//   - height: panel height including borders
//
// Returns:
//   - LogsPanel: initialized logs panel
func NewLogsPanel(width, height int) LogsPanel {
	// Calculate viewport size accounting for borders and scrollbar.
	vw := width - logBorderWidth
	vh := height - logBorderHeight

	// Ensure minimum viewport width.
	vw = max(vw, 1)
	// Ensure minimum viewport height.
	vh = max(vh, 1)
	vp := viewport.New(vw, vh)

	// Return initialized panel with default settings.
	return LogsPanel{
		viewport: vp,
		theme:    ansi.DefaultTheme(),
		width:    width,
		height:   height,
		entries:  nil,
		maxSize:  DefaultLogBufferSize,
		title:    "Logs",
	}
}

// SetMaxSize sets the maximum buffer size for display indicator.
//
// Params:
//   - maxSize: maximum number of log entries to display
func (l *LogsPanel) SetMaxSize(maxSize int) {
	// Use default if invalid size provided.
	if maxSize <= 0 {
		maxSize = DefaultLogBufferSize
	}
	l.maxSize = maxSize
}

// SetSize updates the panel dimensions.
//
// Params:
//   - width: new panel width
//   - height: new panel height
func (l *LogsPanel) SetSize(width, height int) {
	l.width = width
	l.height = height

	// Calculate viewport size accounting for borders and scrollbar.
	vw := width - logBorderWidth
	vh := height - logBorderHeight

	// Ensure minimum viewport width.
	vw = max(vw, 1)
	// Ensure minimum viewport height.
	vh = max(vh, 1)
	l.viewport.Width = vw
	l.viewport.Height = vh
	l.updateContent()
}

// SetFocused sets the focus state.
//
// Params:
//   - focused: true to focus the panel, false to unfocus
func (l *LogsPanel) SetFocused(focused bool) {
	l.focused = focused
}

// Focused returns whether the panel is focused.
//
// Returns:
//   - bool: true if panel is focused
func (l *LogsPanel) Focused() bool {
	// Return current focus state.
	return l.focused
}

// SetEntries updates the log entries.
//
// Params:
//   - entries: slice of log entries to display
func (l *LogsPanel) SetEntries(entries []model.LogEntry) {
	l.entries = entries
	l.updateContent()
}

// AddEntry adds a new log entry and scrolls to bottom.
//
// Params:
//   - entry: log entry to add
func (l *LogsPanel) AddEntry(entry model.LogEntry) {
	l.entries = append(l.entries, entry)
	// Enforce buffer size limit to prevent unbounded memory growth.
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}
	l.updateContent()
	l.viewport.GotoBottom()
}

// updateContent rebuilds the viewport content.
func (l *LogsPanel) updateContent() {
	var sb strings.Builder

	// Calculate message width based on content area.
	contentWidth := l.viewport.Width
	msgWidth := contentWidth - timeColWidth - levelColWidth - serviceColWidth

	// Ensure minimum message width.
	msgWidth = max(msgWidth, minMsgWidth)

	// Build content for each log entry.
	for _, entry := range l.entries {
		line := l.formatLogLine(entry, msgWidth)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	l.viewport.SetContent(sb.String())
}

// formatLogLine formats a single log entry as a display line.
//
// Params:
//   - entry: log entry to format
//   - msgWidth: maximum width for message column
//
// Returns:
//   - string: formatted log line with ANSI colors
func (l *LogsPanel) formatLogLine(entry model.LogEntry, msgWidth int) string {
	// Format timestamp (fixed 8 chars + space).
	ts := entry.Timestamp.Format("15:04:05")

	// Get level display string and color separately.
	levelStr, levelColor := l.getLevelInfo(entry.Level)

	// Format service name (fixed 12 chars).
	service := l.formatServiceName(entry.Service)

	// Build message with metadata.
	msg := l.buildMessage(entry, msgWidth)

	// Build line with proper padding (pad BEFORE adding colors).
	var lineBuf strings.Builder
	lineBuf.Grow(len(ts) + len(levelStr) + len(service) + len(msg) + lineBufferGrow)
	lineBuf.WriteString(l.theme.Muted)
	lineBuf.WriteString(ts)
	lineBuf.WriteString(ansi.Reset)
	lineBuf.WriteByte(' ')
	lineBuf.WriteString(levelColor)
	lineBuf.WriteByte('[')
	lineBuf.WriteString(ansi.Reset)
	lineBuf.WriteString(levelStr)

	// Pad level text to fixed width.
	for i := len(levelStr); i < levelTextMaxWidth; i++ {
		lineBuf.WriteByte(' ')
	}
	lineBuf.WriteString(levelColor)
	lineBuf.WriteByte(']')
	lineBuf.WriteString(ansi.Reset)
	lineBuf.WriteByte(' ')
	lineBuf.WriteString(service)

	// Pad service name to fixed width.
	for i := len([]rune(service)); i < serviceNameMaxWidth; i++ {
		lineBuf.WriteByte(' ')
	}
	lineBuf.WriteByte(' ')
	lineBuf.WriteString(msg)

	// Return formatted line.
	return lineBuf.String()
}

// formatServiceName formats and truncates the service name.
// Sanitizes input to prevent terminal escape injection attacks.
//
// Params:
//   - service: raw service name
//
// Returns:
//   - string: formatted service name
func (l *LogsPanel) formatServiceName(service string) string {
	// Use daemon as default service name.
	if service == "" {
		// Return default service name.
		return "daemon"
	}

	// Sanitize input to prevent ANSI escape injection.
	sanitized := widget.StripANSI(service)

	// Truncate service name if too long.
	if len([]rune(sanitized)) > serviceNameMaxWidth {
		// Return truncated name with ellipsis.
		return widget.TruncateRunes(sanitized, serviceNameMaxWidth, "...")
	}

	// Return original service name.
	return sanitized
}

// buildMessage builds the message string with metadata.
// Sanitizes input to prevent terminal escape injection attacks.
//
// Params:
//   - entry: log entry containing message and metadata
//   - msgWidth: maximum width for message
//
// Returns:
//   - string: formatted message with metadata
func (l *LogsPanel) buildMessage(entry model.LogEntry, msgWidth int) string {
	// Sanitize message to prevent ANSI escape injection.
	msg := widget.StripANSI(entry.Message)

	// Use event type if message is empty.
	if msg == "" {
		msg = widget.StripANSI(entry.EventType)
	}

	// Append metadata if present.
	if len(entry.Metadata) > 0 {
		var sb strings.Builder
		sb.WriteString(msg)
		sb.WriteString(" ")
		sb.WriteString(l.formatMetadata(entry.Metadata))
		msg = sb.String()
	}

	// Truncate message if too long.
	if len([]rune(msg)) > msgWidth {
		// Return truncated message with ellipsis.
		return widget.TruncateRunes(msg, msgWidth, "...")
	}

	// Return complete message.
	return msg
}

// getLevelInfo returns the level string and its color separately.
//
// Params:
//   - level: log level string
//
// Returns:
//   - string: normalized level string
//   - string: ANSI color code for the level
func (l *LogsPanel) getLevelInfo(level string) (levelStr, color string) {
	// Map log level to display string and color.
	switch strings.ToUpper(level) {
	// Handle error level variants.
	case "ERROR", "ERR":
		// Return error color.
		return "ERROR", l.theme.Error
	// Handle warning level variants.
	case "WARN", "WARNING":
		// Return warning color.
		return "WARN", l.theme.Warning
	// Handle info level.
	case "INFO":
		// Return primary color.
		return "INFO", l.theme.Primary
	// Handle debug level.
	case "DEBUG":
		// Return muted color.
		return "DEBUG", l.theme.Muted
	// Handle unknown levels.
	default:
		// Return muted color with original level.
		return level, l.theme.Muted
	}
}

// formatMetadata formats metadata as key=value pairs using strings.Builder.
//
// Params:
//   - meta: metadata map to format (key to any value)
//
// Returns:
//   - string: formatted metadata string
//
// formatMetadata formats metadata as key=value pairs using strings.Builder.
//
// Params:
//   - meta: metadata map to format (key to any value)
//
// Returns:
//   - string: formatted metadata string
func (l *LogsPanel) formatMetadata(meta map[string]any) string {
	// Return early if no metadata.
	if len(meta) == 0 {
		// Return empty string for nil or empty map.
		return ""
	}

	// Collect and sort keys for consistent output.
	keys := slices.Sorted(maps.Keys(meta))

	var sb strings.Builder
	sb.Grow(len(keys) * metadataEstimatedSize)

	// Build key=value pairs for each metadata entry.
	for i, k := range keys {
		// Add space separator between pairs.
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		l.appendMetadataValue(&sb, meta[k])
	}

	// Return formatted string.
	return sb.String()
}

// appendMetadataValue writes a typed metadata value to the builder.
//
// Params:
//   - sb: string builder to write to
//   - val: metadata value to format (any type from log metadata)
func (l *LogsPanel) appendMetadataValue(sb *strings.Builder, val any) {
	// Type switch for efficient formatting of common types.
	switch typed := val.(type) {
	// Format string values directly.
	case string:
		sb.WriteString(typed)
	// Format int values.
	case int:
		sb.WriteString(strconv.Itoa(typed))
	// Format int64 values.
	case int64:
		sb.WriteString(strconv.FormatInt(typed, decimalBase))
	// Format uint64 values.
	case uint64:
		sb.WriteString(strconv.FormatUint(typed, decimalBase))
	// Format float64 values.
	case float64:
		sb.WriteString(strconv.FormatFloat(typed, 'f', formatFloatPrecision, formatFloatBitSize))
	// Format bool values.
	case bool:
		sb.WriteString(strconv.FormatBool(typed))
	// Fallback to fmt for complex types.
	default:
		fmt.Fprint(sb, typed)
	}
}

// Init initializes the component.
//
// Returns:
//   - tea.Cmd: initialization command (nil)
func (l *LogsPanel) Init() tea.Cmd {
	// Return nil as no initialization needed.
	return nil
}

// Update handles messages.
//
// Params:
//   - msg: Bubble Tea message
//
// Returns:
//   - *LogsPanel: updated panel
//   - tea.Cmd: command to execute
func (l *LogsPanel) Update(msg tea.Msg) (*LogsPanel, tea.Cmd) {
	// Ignore messages when not focused.
	if !l.focused {
		// Return unchanged panel.
		return l, nil
	}

	var cmd tea.Cmd

	// Handle keyboard and mouse input.
	switch msg := msg.(type) {
	// Handle keyboard messages.
	case tea.KeyMsg:
		cmd = l.handleKeyMsg(msg)
	// Handle mouse messages.
	case tea.MouseMsg:
		l.viewport, cmd = l.viewport.Update(msg)
	}

	// Return updated panel.
	return l, cmd
}

// handleKeyMsg processes keyboard input.
//
// Params:
//   - msg: key message to process (uses Stringer interface)
//
// Returns:
//   - tea.Cmd: command to execute
func (l *LogsPanel) handleKeyMsg(msg Stringer) tea.Cmd {
	// Delegate to shared viewport key handler.
	return handleViewportKeyMsg(&l.viewport, msg)
}

// View renders the logs panel with border and vertical scrollbar.
//
// Returns:
//   - string: rendered panel
func (l *LogsPanel) View() string {
	var sb strings.Builder

	// Set border color based on focus state.
	borderColor := l.theme.Muted
	// Use primary color when focused.
	if l.focused {
		borderColor = l.theme.Primary
	}

	// Calculate inner width excluding borders and scrollbar.
	innerWidth := l.width - logBorderWidth

	// Render top border with title and indicator.
	l.renderTopBorder(&sb, borderColor, innerWidth)

	// Render content lines with scrollbar.
	l.renderContentLines(&sb, borderColor, innerWidth)

	// Render bottom border.
	l.renderBottomBorder(&sb, borderColor, innerWidth)

	// Return complete view.
	return sb.String()
}

// renderTopBorder renders the top border with title and scroll indicator.
//
// Params:
//   - sb: string builder to write to
//   - borderColor: ANSI color for border
//   - innerWidth: width inside borders
func (l *LogsPanel) renderTopBorder(sb *strings.Builder, borderColor string, innerWidth int) {
	// Build titlePart with strings.Builder to avoid fmt.Sprintf.
	var titleBuf strings.Builder
	titleBuf.Grow(len(l.title) + titleBufferGrow)
	titleBuf.WriteString("- ")
	titleBuf.WriteString(l.theme.Header)
	titleBuf.WriteString(l.title)
	titleBuf.WriteString(borderColor)
	titleBuf.WriteByte(' ')
	titlePart := titleBuf.String()
	scrollPart := l.scrollIndicator()

	// Calculate dashes needed for spacing.
	titleVisLen := titlePrefixLen + len(l.title)
	scrollVisLen := widget.VisibleLen(scrollPart)
	dashCount := innerWidth - titleVisLen - scrollVisLen - dashCountOffset

	// Ensure non-negative dash count.
	dashCount = max(dashCount, 0)

	sb.WriteString(borderColor)
	sb.WriteString("+")
	sb.WriteString(titlePart)
	sb.WriteString(strings.Repeat("-", dashCount))
	sb.WriteString(" ")
	sb.WriteString(scrollPart)
	sb.WriteString(borderColor)
	sb.WriteString("-+")
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")
}

// renderContentLines renders the content area with scrollbar.
//
// Params:
//   - sb: string builder to write to
//   - borderColor: ANSI color for border
//   - innerWidth: width inside borders
func (l *LogsPanel) renderContentLines(sb *strings.Builder, borderColor string, innerWidth int) {
	// Get content lines from viewport.
	content := l.viewport.View()
	lines := strings.Split(content, "\n")

	// Calculate scrollbar characters.
	scrollbarChars := l.renderVerticalScrollbar()

	// Delegate to shared helper for rendering.
	renderContentLinesWithScrollbar(sb, lines, scrollbarChars, l.viewport.Height, innerWidth, borderColor, scrollTrack)
}

// renderBottomBorder renders the bottom border.
//
// Params:
//   - sb: string builder to write to
//   - borderColor: ANSI color for border
//   - innerWidth: width inside borders
func (l *LogsPanel) renderBottomBorder(sb *strings.Builder, borderColor string, innerWidth int) {
	sb.WriteString(borderColor)
	sb.WriteString("+")
	sb.WriteString(strings.Repeat("-", innerWidth+scrollbarColumnWidth))
	sb.WriteString("+")
	sb.WriteString(ansi.Reset)
}

// scrollIndicator returns the entry count as [ count / max ] with ANSI colors.
//
// Returns:
//   - string: formatted scroll indicator
func (l *LogsPanel) scrollIndicator() string {
	count := len(l.entries)
	maxVal := l.maxSize

	// Use default if max not set.
	if maxVal <= 0 {
		maxVal = DefaultLogBufferSize
	}

	var sb strings.Builder
	sb.Grow(metadataBufferSize)
	sb.WriteString(l.theme.Muted)
	sb.WriteString("[ ")
	sb.WriteString(strconv.Itoa(count))
	sb.WriteString(" / ")
	sb.WriteString(strconv.Itoa(maxVal))
	sb.WriteString(" ]")
	sb.WriteString(ansi.Reset)

	// Return formatted indicator.
	return sb.String()
}

// renderVerticalScrollbar returns the scrollbar characters for each row.
//
// Returns:
//   - []string: scrollbar characters for each row
func (l *LogsPanel) renderVerticalScrollbar() []string {
	height := l.viewport.Height
	totalLines := len(l.entries)

	// No scrolling needed if content fits.
	if totalLines <= height {
		// Pre-allocate with capacity using append pattern per VAR-MAKEAPPEND.
		result := make([]string, 0, height)

		// Fill with track characters.
		for range height {
			result = append(result, scrollTrack)
		}

		// Return track-only scrollbar.
		return result
	}

	// Calculate thumb size (minimum 1).
	ratio := float64(height) / float64(totalLines)
	thumbSize := int(float64(height) * ratio)

	// Ensure minimum thumb size.
	thumbSize = max(thumbSize, minThumbSize)

	// Calculate thumb position based on scroll percentage.
	scrollableHeight := height - thumbSize
	scrollPercent := l.viewport.ScrollPercent()
	thumbPos := int(float64(scrollableHeight) * scrollPercent)

	// Build scrollbar with thumb.
	// Pre-allocate with capacity using append pattern per VAR-MAKEAPPEND.
	result := make([]string, 0, height)

	// Assign character to each position.
	for i := range height {
		// Use thumb character for thumb position, track elsewhere.
		if i >= thumbPos && i < thumbPos+thumbSize {
			result = append(result, scrollThumb)
		} else {
			// Use track character outside thumb.
			result = append(result, scrollTrack)
		}
	}

	// Return complete scrollbar.
	return result
}

// Height returns the panel height.
//
// Returns:
//   - int: panel height
func (l *LogsPanel) Height() int {
	// Return current height.
	return l.height
}

// Width returns the panel width.
//
// Returns:
//   - int: panel width
func (l *LogsPanel) Width() int {
	// Return current width.
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
