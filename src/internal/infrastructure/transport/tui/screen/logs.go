// Package screen provides complete screen renderers.
package screen

import (
	"fmt"
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
func (l *LogsRenderer) Render(snap *model.Snapshot) string {
	logs := snap.Logs

	// Summary line.
	period := logs.Period
	if period == 0 {
		period = 5 * 60 * 1e9 // Default 5m.
	}
	periodStr := widget.FormatDurationShort(period)

	summary := fmt.Sprintf("Last %s:  INFO: %d  WARN: %d  ERROR: %d",
		periodStr, logs.InfoCount, logs.WarnCount, logs.ErrorCount)

	// Color warnings/errors.
	if logs.ErrorCount > 0 {
		summary = fmt.Sprintf("Last %s:  INFO: %d  WARN: %d  %sERROR: %d%s",
			periodStr, logs.InfoCount, logs.WarnCount,
			l.theme.Error, logs.ErrorCount, ansi.Reset)
	} else if logs.WarnCount > 0 {
		summary = fmt.Sprintf("Last %s:  INFO: %d  %sWARN: %d%s  ERROR: %d",
			periodStr, logs.InfoCount,
			l.theme.Warning, logs.WarnCount, ansi.Reset, logs.ErrorCount)
	}

	lines := []string{"  " + summary}

	// Separator.
	if len(logs.RecentEntries) > 0 {
		lines = append(lines, "  "+l.theme.Muted+strings.Repeat("─", l.width-6)+ansi.Reset)
	}

	// Recent entries.
	maxWidth := l.width - 6
	for _, entry := range logs.RecentEntries {
		ts := entry.Timestamp.Format("15:04:05")
		level := l.status.LogLevel(entry.Level)
		service := entry.Service

		// Format: timestamp [level] service: message
		prefix := fmt.Sprintf("%s [%s] %s: ", ts, level, service)
		prefixLen := len(ts) + 3 + len(entry.Level) + 2 + len(service) + 2

		// Truncate message if needed.
		msgWidth := maxWidth - prefixLen
		msg := entry.Message
		if len(msg) > msgWidth {
			msg = msg[:msgWidth-1] + "…"
		}

		line := fmt.Sprintf("  %s%s", prefix, msg)
		lines = append(lines, line)
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
		return fmt.Sprintf("%sErrors: %d%s", l.theme.Error, logs.ErrorCount, ansi.Reset)
	}
	if logs.WarnCount > 0 {
		return fmt.Sprintf("%sWarns: %d%s", l.theme.Warning, logs.WarnCount, ansi.Reset)
	}
	return l.theme.Muted + "No errors" + ansi.Reset
}

// RenderInline returns a single-line summary.
func (l *LogsRenderer) RenderInline(snap *model.Snapshot) string {
	logs := snap.Logs

	parts := []string{}
	if logs.ErrorCount > 0 {
		parts = append(parts, fmt.Sprintf("%sE:%d%s", l.theme.Error, logs.ErrorCount, ansi.Reset))
	}
	if logs.WarnCount > 0 {
		parts = append(parts, fmt.Sprintf("%sW:%d%s", l.theme.Warning, logs.WarnCount, ansi.Reset))
	}
	if logs.InfoCount > 0 {
		parts = append(parts, fmt.Sprintf("I:%d", logs.InfoCount))
	}

	if len(parts) == 0 {
		return l.theme.Muted + "No logs" + ansi.Reset
	}

	return strings.Join(parts, " ")
}
