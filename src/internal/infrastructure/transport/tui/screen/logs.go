// Package screen provides complete screen renderers.
package screen

import (
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// LogsRenderer renders log summary.
type LogsRenderer struct {
	theme  ansi.Theme
	width  int
	status *widget.StatusIndicator
}

// NewLogsRenderer creates a logs renderer.
func NewLogsRenderer(width int) *LogsRenderer {
	return &LogsRenderer{
		theme:  ansi.DefaultTheme(),
		width:  width,
		status: widget.NewStatusIndicator(),
	}
}

// SetWidth updates the renderer width.
func (l *LogsRenderer) SetWidth(width int) {
	l.width = width
}

// Render returns the logs summary section (raw mode).
// Uses pre-allocated slices and strings.Builder for efficiency.
func (l *LogsRenderer) Render(snap *model.Snapshot) string {
	logs := snap.Logs

	// Summary line.
	period := logs.Period
	if period == 0 {
		period = 5 * 60 * 1e9 // Default 5m.
	}
	periodStr := widget.FormatDurationShort(period)

	// Build summary with strings.Builder to avoid allocations.
	var sb strings.Builder
	sb.Grow(64)
	sb.WriteString("Last ")
	sb.WriteString(periodStr)
	sb.WriteString(":  INFO: ")
	sb.WriteString(strconv.Itoa(logs.InfoCount))
	sb.WriteString("  ")

	// Color warnings/errors.
	if logs.WarnCount > 0 && logs.ErrorCount == 0 {
		sb.WriteString(l.theme.Warning)
		sb.WriteString("WARN: ")
		sb.WriteString(strconv.Itoa(logs.WarnCount))
		sb.WriteString(ansi.Reset)
	} else {
		sb.WriteString("WARN: ")
		sb.WriteString(strconv.Itoa(logs.WarnCount))
	}
	sb.WriteString("  ")

	if logs.ErrorCount > 0 {
		sb.WriteString(l.theme.Error)
		sb.WriteString("ERROR: ")
		sb.WriteString(strconv.Itoa(logs.ErrorCount))
		sb.WriteString(ansi.Reset)
	} else {
		sb.WriteString("ERROR: ")
		sb.WriteString(strconv.Itoa(logs.ErrorCount))
	}

	// Pre-allocate lines slice: summary + separator + entries + empty state.
	linesCap := 2 + len(logs.RecentEntries)
	lines := make([]string, 0, linesCap)
	lines = append(lines, "  "+sb.String())

	// Separator.
	if len(logs.RecentEntries) > 0 {
		sepWidth := l.width - 6
		if sepWidth < 0 {
			sepWidth = 0
		}
		lines = append(lines, "  "+l.theme.Muted+strings.Repeat("─", sepWidth)+ansi.Reset)
	}

	// Recent entries using strings.Builder.
	maxWidth := l.width - 6
	for _, entry := range logs.RecentEntries {
		ts := entry.Timestamp.Format("15:04:05")
		level := l.status.LogLevel(entry.Level)
		service := entry.Service

		// Calculate prefix length for truncation.
		prefixLen := len(ts) + 3 + len(entry.Level) + 2 + len(service) + 2

		// Truncate message if needed (rune-safe for UTF-8).
		msgWidth := maxWidth - prefixLen
		msg := entry.Message
		if msgWidth <= 0 {
			msg = ""
		} else {
			msgRunes := []rune(msg)
			if len(msgRunes) > msgWidth {
				if msgWidth <= 1 {
					msg = "…"
				} else {
					msg = string(msgRunes[:msgWidth-1]) + "…"
				}
			}
		}

		// Build line with strings.Builder.
		var lineSb strings.Builder
		lineSb.Grow(prefixLen + len(msg) + 4)
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
	if len(logs.RecentEntries) == 0 {
		lines = append(lines, "  "+l.theme.Muted+"No recent logs"+ansi.Reset)
	}

	box := widget.NewBox(l.width).
		SetTitle("Logs Summary").
		SetTitleColor(l.theme.Header).
		AddLines(lines)

	return box.Render()
}

// RenderBadge returns a compact badge for interactive mode.
func (l *LogsRenderer) RenderBadge(snap *model.Snapshot) string {
	logs := snap.Logs

	if logs.ErrorCount > 0 {
		return l.theme.Error + "Errors: " + strconv.Itoa(logs.ErrorCount) + ansi.Reset
	}
	if logs.WarnCount > 0 {
		return l.theme.Warning + "Warns: " + strconv.Itoa(logs.WarnCount) + ansi.Reset
	}
	return l.theme.Muted + "No errors" + ansi.Reset
}

// RenderInline returns a single-line summary.
func (l *LogsRenderer) RenderInline(snap *model.Snapshot) string {
	logs := snap.Logs

	parts := make([]string, 0, 3)
	if logs.ErrorCount > 0 {
		parts = append(parts, l.theme.Error+"E:"+strconv.Itoa(logs.ErrorCount)+ansi.Reset)
	}
	if logs.WarnCount > 0 {
		parts = append(parts, l.theme.Warning+"W:"+strconv.Itoa(logs.WarnCount)+ansi.Reset)
	}
	if logs.InfoCount > 0 {
		parts = append(parts, "I:"+strconv.Itoa(logs.InfoCount))
	}

	if len(parts) == 0 {
		return l.theme.Muted + "No logs" + ansi.Reset
	}

	return strings.Join(parts, " ")
}
