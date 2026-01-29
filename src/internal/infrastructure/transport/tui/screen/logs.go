// Package screen provides complete screen renderers.
package screen

import (
	"strconv"
	"strings"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

const (
	// defaultLogPeriod is the default log summary period (5 minutes).
	defaultLogPeriod time.Duration = 5 * time.Minute

	// summaryBufferSize is the pre-allocation size for summary string builder.
	summaryBufferSize int = 64

	// separatorPadding is the padding subtracted for log separators.
	separatorPadding int = 6

	// logLinePadding is the padding for log entry lines.
	logLinePadding int = 6

	// timestampBracketLen is the length of " [" after timestamp.
	timestampBracketLen int = 3

	// levelCloseBracketLen is the length of "] ".
	levelCloseBracketLen int = 2

	// serviceColonLen is the length of ": " after service name.
	serviceColonLen int = 2

	// linePrefixExtra is extra chars for ANSI codes and prefix in line buffer.
	linePrefixExtra int = 4

	// inlinePartsCapacity is the initial capacity for inline parts slice.
	inlinePartsCapacity int = 3

	// linesCapBaseCount is the base capacity for summary and separator lines.
	linesCapBaseCount int = 2
)

// LogsRenderer renders log summary.
// It provides formatted display of log counts and recent log entries.
type LogsRenderer struct {
	theme  ansi.Theme
	width  int
	status *widget.StatusIndicator
}

// NewLogsRenderer creates a logs renderer.
//
// Params:
//   - width: terminal width in columns
//
// Returns:
//   - *LogsRenderer: configured renderer instance
func NewLogsRenderer(width int) *LogsRenderer {
	// return computed result.
	return &LogsRenderer{
		theme:  ansi.DefaultTheme(),
		width:  width,
		status: widget.NewStatusIndicator(),
	}
}

// SetWidth updates the renderer width.
//
// Params:
//   - width: new terminal width in columns
func (l *LogsRenderer) SetWidth(width int) {
	l.width = width
}

// Render returns the logs summary section (raw mode).
// Uses pre-allocated slices and strings.Builder for efficiency.
//
// Params:
//   - snap: snapshot containing log data
//
// Returns:
//   - string: rendered logs summary section
func (l *LogsRenderer) Render(snap *model.Snapshot) string {
	logs := snap.Logs

	summaryLine := l.buildSummaryLine(logs)
	lines := l.buildLogLines(logs, summaryLine)

	box := widget.NewBox(l.width).
		SetTitle("Logs Summary").
		SetTitleColor(l.theme.Header).
		AddLines(lines)

	// return computed result.
	return box.Render()
}

// buildSummaryLine builds the summary line for log counts.
//
// Params:
//   - logs: log summary data.
//
// Returns:
//   - string: formatted summary line.
func (l *LogsRenderer) buildSummaryLine(logs model.LogSummary) string {
	period := logs.Period
	// check for empty value.
	if period == 0 {
		period = defaultLogPeriod
	}
	periodStr := widget.FormatDurationShort(period)

	var sb strings.Builder
	sb.Grow(summaryBufferSize)
	sb.WriteString("Last ")
	sb.WriteString(periodStr)
	sb.WriteString(":  INFO: ")
	sb.WriteString(strconv.Itoa(logs.InfoCount))
	sb.WriteString("  ")

	l.appendWarnCount(&sb, logs)
	sb.WriteString("  ")

	l.appendErrorCount(&sb, logs)

	// return computed result.
	return sb.String()
}

// appendWarnCount appends warning count to builder with optional color.
//
// Params:
//   - sb: string builder to append to.
//   - logs: log summary data.
func (l *LogsRenderer) appendWarnCount(sb *strings.Builder, logs model.LogSummary) {
	// check for empty value.
	if logs.WarnCount > 0 && logs.ErrorCount == 0 {
		sb.WriteString(l.theme.Warning)
		sb.WriteString("WARN: ")
		sb.WriteString(strconv.Itoa(logs.WarnCount))
		sb.WriteString(ansi.Reset)
	// handle alternative case.
	} else {
		sb.WriteString("WARN: ")
		sb.WriteString(strconv.Itoa(logs.WarnCount))
	}
}

// appendErrorCount appends error count to builder with optional color.
//
// Params:
//   - sb: string builder to append to.
//   - logs: log summary data.
func (l *LogsRenderer) appendErrorCount(sb *strings.Builder, logs model.LogSummary) {
	// check for positive value.
	if logs.ErrorCount > 0 {
		sb.WriteString(l.theme.Error)
		sb.WriteString("ERROR: ")
		sb.WriteString(strconv.Itoa(logs.ErrorCount))
		sb.WriteString(ansi.Reset)
	// handle alternative case.
	} else {
		sb.WriteString("ERROR: ")
		sb.WriteString(strconv.Itoa(logs.ErrorCount))
	}
}

// buildLogLines builds all content lines for the logs section.
//
// Params:
//   - logs: log summary data.
//   - summaryLine: pre-formatted summary line.
//
// Returns:
//   - []string: all content lines.
func (l *LogsRenderer) buildLogLines(logs model.LogSummary, summaryLine string) []string {
	linesCap := linesCapBaseCount + len(logs.RecentEntries)
	lines := make([]string, 0, linesCap)
	lines = append(lines, "  "+summaryLine)

	// check for positive value.
	if len(logs.RecentEntries) > 0 {
		lines = append(lines, l.buildSeparator())
	}

	lines = append(lines, l.buildEntryLines(logs.RecentEntries)...)

	// check for empty value.
	if len(logs.RecentEntries) == 0 {
		lines = append(lines, "  "+l.theme.Muted+"No recent logs"+ansi.Reset)
	}

	// return computed result.
	return lines
}

// buildSeparator builds a visual separator line.
//
// Returns:
//   - string: separator line.
func (l *LogsRenderer) buildSeparator() string {
	sepWidth := max(l.width-separatorPadding, 0)
	// return computed result.
	return "  " + l.theme.Muted + strings.Repeat("─", sepWidth) + ansi.Reset
}

// buildEntryLines builds lines for log entries.
//
// Params:
//   - entries: recent log entries.
//
// Returns:
//   - []string: formatted entry lines.
func (l *LogsRenderer) buildEntryLines(entries []model.LogEntry) []string {
	maxWidth := l.width - logLinePadding
	lines := make([]string, 0, len(entries))

	// iterate over collection.
	for _, entry := range entries {
		lines = append(lines, l.formatLogEntry(entry, maxWidth))
	}

	// return computed result.
	return lines
}

// formatLogEntry formats a single log entry.
//
// Params:
//   - entry: log entry to format.
//   - maxWidth: maximum line width.
//
// Returns:
//   - string: formatted log line.
func (l *LogsRenderer) formatLogEntry(entry model.LogEntry, maxWidth int) string {
	ts := entry.Timestamp.Format("15:04:05")
	level := l.status.LogLevel(entry.Level)

	prefixLen := len(ts) + timestampBracketLen + len(entry.Level) + levelCloseBracketLen + len(entry.Service) + serviceColonLen

	msg := l.truncateMessage(entry.Message, maxWidth-prefixLen)

	var sb strings.Builder
	sb.Grow(prefixLen + len(msg) + linePrefixExtra)
	sb.WriteString("  ")
	sb.WriteString(ts)
	sb.WriteString(" [")
	sb.WriteString(level)
	sb.WriteString("] ")
	sb.WriteString(entry.Service)
	sb.WriteString(": ")
	sb.WriteString(msg)

	// return computed result.
	return sb.String()
}

// truncateMessage truncates a message to fit width.
//
// Params:
//   - msg: message to truncate.
//   - width: available width.
//
// Returns:
//   - string: truncated message.
func (l *LogsRenderer) truncateMessage(msg string, width int) string {
	// evaluate condition.
	if width <= 0 {
		// return computed result.
		return ""
	}

	msgRunes := []rune(msg)
	// evaluate condition.
	if len(msgRunes) <= width {
		// return computed result.
		return msg
	}

	// evaluate condition.
	if width <= 1 {
		// return computed result.
		return "…"
	}

	// return computed result.
	return string(msgRunes[:width-1]) + "…"
}

// RenderBadge returns a compact badge for interactive mode.
//
// Params:
//   - snap: snapshot containing log data
//
// Returns:
//   - string: compact badge showing error/warning counts
func (l *LogsRenderer) RenderBadge(snap *model.Snapshot) string {
	logs := snap.Logs

	// check for positive value.
	if logs.ErrorCount > 0 {
		// return computed result.
		return l.theme.Error + "Errors: " + strconv.Itoa(logs.ErrorCount) + ansi.Reset
	}
	// check for positive value.
	if logs.WarnCount > 0 {
		// return computed result.
		return l.theme.Warning + "Warns: " + strconv.Itoa(logs.WarnCount) + ansi.Reset
	}
	// return computed result.
	return l.theme.Muted + "No errors" + ansi.Reset
}

// RenderInline returns a single-line summary.
//
// Params:
//   - snap: snapshot containing log data
//
// Returns:
//   - string: single-line log summary
func (l *LogsRenderer) RenderInline(snap *model.Snapshot) string {
	logs := snap.Logs

	parts := make([]string, 0, inlinePartsCapacity)
	// check for positive value.
	if logs.ErrorCount > 0 {
		parts = append(parts, l.theme.Error+"E:"+strconv.Itoa(logs.ErrorCount)+ansi.Reset)
	}
	// check for positive value.
	if logs.WarnCount > 0 {
		parts = append(parts, l.theme.Warning+"W:"+strconv.Itoa(logs.WarnCount)+ansi.Reset)
	}
	// check for positive value.
	if logs.InfoCount > 0 {
		parts = append(parts, "I:"+strconv.Itoa(logs.InfoCount))
	}

	// check for empty value.
	if len(parts) == 0 {
		// return computed result.
		return l.theme.Muted + "No logs" + ansi.Reset
	}

	// return computed result.
	return strings.Join(parts, " ")
}
