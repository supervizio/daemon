// Package screen provides complete screen renderers.
package screen

import (
	"fmt"
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
			pid = fmt.Sprintf("%d", svc.PID)
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
			pid = fmt.Sprintf("%d", svc.PID)
		}

		uptime := "-"
		if svc.State == process.StateRunning || svc.State == process.StateStarting {
			uptime = widget.FormatDuration(svc.Uptime)
		}

		restarts := "-"
		if svc.RestartCount > 0 {
			restarts = fmt.Sprintf("%d", svc.RestartCount)
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

	parts := []string{fmt.Sprintf("%d services", total)}

	if running > 0 {
		parts = append(parts, fmt.Sprintf("%d running", running))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	if healthy > 0 && healthy != running {
		parts = append(parts, fmt.Sprintf("%d healthy", healthy))
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

// stateShort returns a short state string.
func (s *ServicesRenderer) stateShort(state process.State) string {
	return s.status.ProcessStateShort(state)
}

// formatPorts formats listener ports.
func (s *ServicesRenderer) formatPorts(listeners []model.ListenerSnapshot) string {
	if len(listeners) == 0 {
		return "-"
	}

	ports := make([]string, 0, len(listeners))
	for _, l := range listeners {
		ports = append(ports, fmt.Sprintf(":%d", l.Port))
	}

	result := strings.Join(ports, " ")
	if len(result) > 12 {
		return result[:11] + "…"
	}
	return result
}

// prefixLines adds a prefix to each line.
func prefixLines(lines []string, prefix string) []string {
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = prefix + line
	}
	return result
}

// RenderNamesOnly renders service names in columns (for raw mode startup banner).
// No dynamic data (state, PID, metrics) - just service names.
// Column width is dynamically calculated based on the longest service name.
func (s *ServicesRenderer) RenderNamesOnly(snap *model.Snapshot) string {
	if len(snap.Services) == 0 {
		box := widget.NewBox(s.width).
			SetTitle("Services (0 configured)").
			SetTitleColor(s.theme.Header).
						AddLine("  " + s.theme.Muted + "No services configured" + ansi.Reset)
		return box.Render()
	}

	// Find longest service name.
	maxLen := 0
	for _, svc := range snap.Services {
		if len(svc.Name) > maxLen {
			maxLen = len(svc.Name)
		}
	}

	// Column width = longest name + 2 chars padding (minimum 8).
	colWidth := maxLen + 2
	if colWidth < 8 {
		colWidth = 8
	}

	// Calculate number of columns that fit.
	usableWidth := s.width - 6 // Box borders + left padding
	cols := usableWidth / colWidth
	if cols < 1 {
		cols = 1
	}

	// Build rows.
	rows := (len(snap.Services) + cols - 1) / cols
	lines := make([]string, rows)

	for row := 0; row < rows; row++ {
		var rowParts []string
		for col := 0; col < cols; col++ {
			idx := row*cols + col
			if idx < len(snap.Services) {
				name := snap.Services[idx].Name
				// Truncate if somehow still too long.
				if len(name) > colWidth-1 {
					name = name[:colWidth-2] + "…"
				}
				// Pad to column width.
				rowParts = append(rowParts, widget.Pad(name, colWidth, widget.AlignLeft))
			}
		}
		lines[row] = "  " + strings.Join(rowParts, "")
	}

	title := fmt.Sprintf("Services (%d configured)", len(snap.Services))
	box := widget.NewBox(s.width).
		SetTitle(title).
		SetTitleColor(s.theme.Header).
				AddLines(lines)

	return box.Render()
}
