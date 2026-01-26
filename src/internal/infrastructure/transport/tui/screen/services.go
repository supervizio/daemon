// Package screen provides complete screen renderers.
package screen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// ServicesRenderer renders the services section.
type ServicesRenderer struct {
	theme  ansi.Theme
	icons  ansi.StatusIcon
	width  int
	status *widget.StatusIndicator
}

// NewServicesRenderer creates a services renderer.
func NewServicesRenderer(width int) *ServicesRenderer {
	return &ServicesRenderer{
		theme:  ansi.DefaultTheme(),
		icons:  ansi.DefaultIcons(),
		width:  width,
		status: widget.NewStatusIndicator(),
	}
}

// SetWidth updates the renderer width.
func (s *ServicesRenderer) SetWidth(width int) {
	s.width = width
}

// Render returns the services section.
func (s *ServicesRenderer) Render(snap *model.Snapshot) string {
	if len(snap.Services) == 0 {
		return s.renderEmpty()
	}

	layout := terminal.GetLayout(terminal.Size{Cols: s.width, Rows: 24})

	switch layout {
	case terminal.LayoutCompact:
		return s.renderCompact(snap)
	case terminal.LayoutNormal:
		return s.renderNormal(snap)
	case terminal.LayoutWide, terminal.LayoutUltraWide:
		return s.renderWide(snap)
	}
	return s.renderNormal(snap)
}

// renderEmpty renders when there are no services.
func (s *ServicesRenderer) renderEmpty() string {
	box := widget.NewBox(s.width).
		SetTitle("Services").
		SetTitleColor(s.theme.Header).
		AddLine("  " + s.theme.Muted + "No services configured" + ansi.Reset)

	return box.Render()
}

// renderCompact renders a minimal service list.
func (s *ServicesRenderer) renderCompact(snap *model.Snapshot) string {
	lines := make([]string, 0, len(snap.Services)+1)

	for _, svc := range snap.Services {
		icon := s.status.ProcessState(svc.State)
		name := widget.Truncate(svc.Name, 10)
		state := s.stateShort(svc.State)

		var extra string
		if svc.State == process.StateRunning {
			extra = widget.FormatPercent(svc.CPUPercent)
		}

		line := fmt.Sprintf("  %s %-10s %s %s",
			icon, name, state, extra)
		lines = append(lines, line)
	}

	// Summary.
	summary := s.renderSummary(snap)
	lines = append(lines, "  "+s.theme.Muted+summary+ansi.Reset)

	box := widget.NewBox(s.width).
		SetTitle("Services").
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	return box.Render()
}

// renderNormal renders a standard service table.
func (s *ServicesRenderer) renderNormal(snap *model.Snapshot) string {
	// Build table.
	table := widget.NewTable(s.width-4).
		AddColumn("", 2, widget.AlignLeft).         // Icon
		AddFlexColumn("NAME", 8, widget.AlignLeft). // Name
		AddColumn("STATE", 8, widget.AlignLeft).    // State
		AddColumn("PID", 6, widget.AlignRight).     // PID
		AddColumn("UPTIME", 8, widget.AlignRight).  // Uptime
		AddColumn("CPU", 5, widget.AlignRight).     // CPU
		AddColumn("MEM", 6, widget.AlignRight)      // Memory

	for _, svc := range snap.Services {
		icon := s.status.ProcessState(svc.State)
		state := s.status.ProcessStateText(svc.State)

		pid := "-"
		if svc.PID > 0 {
			pid = strconv.Itoa(svc.PID)
		}

		uptime := "-"
		if svc.State == process.StateRunning || svc.State == process.StateStarting {
			uptime = widget.FormatDurationShort(svc.Uptime)
		}

		cpu := "-"
		mem := "-"
		if svc.State == process.StateRunning {
			cpu = widget.FormatPercent(svc.CPUPercent)
			mem = widget.FormatBytesShort(svc.MemoryRSS)
		}

		table.AddRow(icon, svc.Name, state, pid, uptime, cpu, mem)
	}

	lines := strings.Split(table.Render(), "\n")
	lines = append(lines, "")
	lines = append(lines, "  "+s.theme.Muted+s.renderSummary(snap)+ansi.Reset)

	box := widget.NewBox(s.width).
		SetTitle("Services").
		SetTitleColor(s.theme.Header).
		AddLines(prefixLines(lines, "  "))

	return box.Render()
}

// renderWide renders an expanded service table.
func (s *ServicesRenderer) renderWide(snap *model.Snapshot) string {
	// Build table with more columns.
	table := widget.NewTable(s.width-4).
		AddColumn("", 2, widget.AlignLeft).          // Icon
		AddFlexColumn("NAME", 10, widget.AlignLeft). // Name
		AddColumn("STATE", 8, widget.AlignLeft).     // State
		AddColumn("PID", 7, widget.AlignRight).      // PID
		AddColumn("UPTIME", 10, widget.AlignRight).  // Uptime
		AddColumn("RESTARTS", 8, widget.AlignRight). // Restarts
		AddColumn("HEALTH", 8, widget.AlignLeft).    // Health
		AddColumn("CPU", 6, widget.AlignRight).      // CPU
		AddColumn("MEM", 7, widget.AlignRight).      // Memory
		AddColumn("PORTS", 12, widget.AlignLeft)     // Ports

	for _, svc := range snap.Services {
		icon := s.status.ProcessState(svc.State)
		state := s.status.ProcessStateText(svc.State)

		pid := "-"
		if svc.PID > 0 {
			pid = strconv.Itoa(svc.PID)
		}

		uptime := "-"
		if svc.State == process.StateRunning || svc.State == process.StateStarting {
			uptime = widget.FormatDuration(svc.Uptime)
		}

		restarts := "-"
		if svc.RestartCount > 0 {
			restarts = strconv.Itoa(svc.RestartCount)
		}

		healthStr := s.status.HealthStatusText(svc.Health)

		cpu := "-"
		mem := "-"
		if svc.State == process.StateRunning {
			cpu = widget.FormatPercent(svc.CPUPercent)
			mem = widget.FormatBytesShort(svc.MemoryRSS)
		}

		ports := s.formatPorts(svc.Listeners)

		table.AddRow(icon, svc.Name, state, pid, uptime, restarts, healthStr, cpu, mem, ports)
	}

	lines := strings.Split(table.Render(), "\n")
	lines = append(lines, "")
	lines = append(lines, "  "+s.theme.Muted+s.renderSummary(snap)+ansi.Reset)

	box := widget.NewBox(s.width).
		SetTitle("Services").
		SetTitleColor(s.theme.Header).
		AddLines(prefixLines(lines, "  "))

	return box.Render()
}

// renderSummary returns a summary line.
func (s *ServicesRenderer) renderSummary(snap *model.Snapshot) string {
	total := len(snap.Services)
	running := snap.RunningCount()
	failed := snap.FailedCount()
	healthy := snap.HealthyCount()

	parts := []string{strconv.Itoa(total) + " services"}

	if running > 0 {
		parts = append(parts, strconv.Itoa(running)+" running")
	}
	if failed > 0 {
		parts = append(parts, strconv.Itoa(failed)+" failed")
	}
	if healthy > 0 && healthy != running {
		parts = append(parts, strconv.Itoa(healthy)+" healthy")
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

// stateShort returns a short state string.
func (s *ServicesRenderer) stateShort(state process.State) string {
	return s.status.ProcessStateShort(state)
}

// formatPorts formats listener ports with colors based on status.
// Colors: Green (OK), Yellow (Warning), Red (Error).
func (s *ServicesRenderer) formatPorts(listeners []model.ListenerSnapshot) string {
	if len(listeners) == 0 {
		return "-"
	}

	var sb strings.Builder
	for i, l := range listeners {
		if i > 0 {
			sb.WriteString(" ")
		}

		// Choose color based on status.
		var color string
		switch l.Status {
		case model.PortStatusOK:
			color = s.theme.Success // Green.
		case model.PortStatusWarning:
			color = s.theme.Warning // Yellow.
		case model.PortStatusError:
			color = s.theme.Error // Red.
		case model.PortStatusUnknown:
			color = s.theme.Muted
		}

		sb.WriteString(color)
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(l.Port))
		sb.WriteString(ansi.Reset)
	}

	return sb.String()
}

// prefixLines adds a prefix to each line.
func prefixLines(lines []string, prefix string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = prefix + line
	}
	return result
}

// RenderNamesOnly renders service names with ports in columns (for raw mode startup banner).
// Shows service name followed by port numbers from config (no colors, no status checking).
// Column width is dynamically calculated based on content.
func (s *ServicesRenderer) RenderNamesOnly(snap *model.Snapshot) string {
	if len(snap.Services) == 0 {
		box := widget.NewBox(s.width).
			SetTitle("Services (0 configured)").
			SetTitleColor(s.theme.Header).
			AddLine("  " + s.theme.Muted + "No services configured" + ansi.Reset)
		return box.Render()
	}

	// Build service entries with name and ports (plain text, no colors).
	type serviceEntry struct {
		display    string // Full display string.
		visibleLen int    // Length of display string.
	}
	entries := make([]serviceEntry, 0, len(snap.Services))

	for _, svc := range snap.Services {
		var sb strings.Builder
		sb.WriteString(svc.Name)

		// Add ports if available (plain text from config, no status colors).
		if len(svc.Listeners) > 0 {
			sb.WriteByte(' ')
			for i, l := range svc.Listeners {
				if i > 0 {
					sb.WriteByte(',')
				}
				sb.WriteByte(':')
				sb.WriteString(strconv.Itoa(l.Port))
			}
		}

		display := sb.String()
		// In raw mode, display has no ANSI codes, so visibleLen = len(display).
		entries = append(entries, serviceEntry{display: display, visibleLen: len(display)})
	}

	// Find longest entry for column width.
	maxLen := 0
	for _, e := range entries {
		if e.visibleLen > maxLen {
			maxLen = e.visibleLen
		}
	}

	// Column width = longest entry + 2 chars padding (minimum 10).
	colWidth := maxLen + 2
	if colWidth < 10 {
		colWidth = 10
	}

	// Calculate number of columns that fit.
	usableWidth := s.width - 6 // Box borders + left padding.
	cols := usableWidth / colWidth
	if cols < 1 {
		cols = 1
	}

	// Build rows.
	rows := (len(entries) + cols - 1) / cols
	lines := make([]string, rows)

	for row := 0; row < rows; row++ {
		var rowParts []string
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if idx < len(entries) {
				e := entries[idx]
				// Pad to column width using visible length.
				padding := colWidth - e.visibleLen
				if padding < 0 {
					padding = 0
				}
				rowParts = append(rowParts, e.display+strings.Repeat(" ", padding))
			}
		}
		lines[row] = "  " + strings.Join(rowParts, "")
	}

	title := "Services (" + strconv.Itoa(len(snap.Services)) + " configured)"
	box := widget.NewBox(s.width).
		SetTitle(title).
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	return box.Render()
}
