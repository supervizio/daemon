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
	// Return configured logs renderer with defaults.
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

	// Summary line.
	period := logs.Period
	// Use default 5 minute period if not specified.
	if period == 0 {
		period = defaultLogPeriod
	}
	periodStr := widget.FormatDurationShort(period)

	// Build summary with strings.Builder to avoid allocations.
	var sb strings.Builder
	sb.Grow(summaryBufferSize)
	sb.WriteString("Last ")
	sb.WriteString(periodStr)
	sb.WriteString(":  INFO: ")
	sb.WriteString(strconv.Itoa(logs.InfoCount))
	sb.WriteString("  ")

	// Color warnings/errors.
	// Highlight warnings in yellow if no errors.
	if logs.WarnCount > 0 && logs.ErrorCount == 0 {
		sb.WriteString(l.theme.Warning)
		sb.WriteString("WARN: ")
		sb.WriteString(strconv.Itoa(logs.WarnCount))
		sb.WriteString(ansi.Reset)
		// Use default formatting otherwise.
	} else {
		sb.WriteString("WARN: ")
		sb.WriteString(strconv.Itoa(logs.WarnCount))
	}
	sb.WriteString("  ")

	// Highlight errors in red when present.
	if logs.ErrorCount > 0 {
		sb.WriteString(l.theme.Error)
		sb.WriteString("ERROR: ")
		sb.WriteString(strconv.Itoa(logs.ErrorCount))
		sb.WriteString(ansi.Reset)
		// Use default formatting otherwise.
	} else {
		sb.WriteString("ERROR: ")
		sb.WriteString(strconv.Itoa(logs.ErrorCount))
	}

	// Pre-allocate lines slice: summary + separator + entries + empty state.
	linesCap := linesCapBaseCount + len(logs.RecentEntries)
	lines := make([]string, 0, linesCap)
	lines = append(lines, "  "+sb.String())

	// Separator.
	// Add visual separator when log entries exist.
	if len(logs.RecentEntries) > 0 {
		sepWidth := l.width - separatorPadding
		// Ensure non-negative separator width.
		if sepWidth < 0 {
			sepWidth = 0
		}
		lines = append(lines, "  "+l.theme.Muted+strings.Repeat("─", sepWidth)+ansi.Reset)
	}

	// Recent entries using strings.Builder.
	maxWidth := l.width - logLinePadding
	// Iterate through recent log entries to format each line.
	for _, entry := range logs.RecentEntries {
		ts := entry.Timestamp.Format("15:04:05")
		level := l.status.LogLevel(entry.Level)
		service := entry.Service

		// Calculate prefix length for truncation.
		prefixLen := len(ts) + timestampBracketLen + len(entry.Level) + levelCloseBracketLen + len(service) + serviceColonLen

		// Truncate message if needed (rune-safe for UTF-8).
		msgWidth := maxWidth - prefixLen
		msg := entry.Message
		// Hide message when no space available.
		if msgWidth <= 0 {
			msg = ""
			// Process message when space is available.
		} else {
			msgRunes := []rune(msg)
			// Truncate with ellipsis if message exceeds available width.
			if len(msgRunes) > msgWidth {
				// Use only ellipsis for very narrow widths.
				if msgWidth <= 1 {
					msg = "…"
					// Truncate with partial message plus ellipsis.
				} else {
					msg = string(msgRunes[:msgWidth-1]) + "…"
				}
			}
		}

		// Build line with strings.Builder.
		var lineSb strings.Builder
		lineSb.Grow(prefixLen + len(msg) + linePrefixExtra)
		lineSb.WriteString("  ")
		lineSb.WriteString(ts)
		lineSb.WriteString(" [")
		lineSb.WriteString(level)
		lineSb.WriteString("] ")
		lineSb.WriteString(service)
		lineSb.WriteString(": ")
		lineSb.WriteString(msg)
		lines = append(lines, lineSb.String())
	}

	// Empty state.
	// Show placeholder message when no logs available.
	if len(logs.RecentEntries) == 0 {
		lines = append(lines, "  "+l.theme.Muted+"No recent logs"+ansi.Reset)
	}

	box := widget.NewBox(l.width).
		SetTitle("Logs Summary").
		SetTitleColor(l.theme.Header).
		AddLines(lines)

	// Return rendered logs summary box.
	return box.Render()
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

	// Show error count in red when errors exist.
	if logs.ErrorCount > 0 {
		// Return error badge with count.
		return l.theme.Error + "Errors: " + strconv.Itoa(logs.ErrorCount) + ansi.Reset
	}
	// Show warning count in yellow when warnings exist.
	if logs.WarnCount > 0 {
		// Return warning badge with count.
		return l.theme.Warning + "Warns: " + strconv.Itoa(logs.WarnCount) + ansi.Reset
	}
	// Return clean status when no issues.
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
	// Add error count if present.
	if logs.ErrorCount > 0 {
		parts = append(parts, l.theme.Error+"E:"+strconv.Itoa(logs.ErrorCount)+ansi.Reset)
	}
	// Add warning count if present.
	if logs.WarnCount > 0 {
		parts = append(parts, l.theme.Warning+"W:"+strconv.Itoa(logs.WarnCount)+ansi.Reset)
	}
	// Add info count if present.
	if logs.InfoCount > 0 {
		parts = append(parts, "I:"+strconv.Itoa(logs.InfoCount))
	}

	// Return empty state message when no logs.
	if len(parts) == 0 {
		// Return placeholder for no logs.
		return l.theme.Muted + "No logs" + ansi.Reset
	}

	// Return formatted log summary.
	return strings.Join(parts, " ")
}
