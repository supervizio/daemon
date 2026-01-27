// Package screen provides complete screen renderers.
package screen

import (
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// ServicesRenderer renders the services section.
// Displays service list with state, metrics, and health information.
type ServicesRenderer struct {
	theme  ansi.Theme
	icons  ansi.StatusIcon
	width  int
	status *widget.StatusIndicator
}

// NewServicesRenderer creates a services renderer.
//
// Params:
//   - width: terminal width for rendering
//
// Returns:
//   - *ServicesRenderer: configured renderer instance
func NewServicesRenderer(width int) *ServicesRenderer {
	// Initialize with default theme and status indicators.
	return &ServicesRenderer{
		theme:  ansi.DefaultTheme(),
		icons:  ansi.DefaultIcons(),
		width:  width,
		status: widget.NewStatusIndicator(),
	}
}

// SetWidth updates the renderer width.
//
// Params:
//   - width: new terminal width
func (s *ServicesRenderer) SetWidth(width int) {
	s.width = width
}

// Render returns the services section.
//
// Params:
//   - snap: system snapshot with service data
//
// Returns:
//   - string: rendered services section
func (s *ServicesRenderer) Render(snap *model.Snapshot) string {
	// defaultRows is the default number of terminal rows for layout calculations.
	const defaultRows int = 24

	// Show empty state if no services.
	if len(snap.Services) == 0 {
		// Return empty services box.
		return s.renderEmpty()
	}

	layout := terminal.GetLayout(terminal.Size{Cols: s.width, Rows: defaultRows})

	// Select rendering mode based on terminal layout.
	switch layout {
	// Compact mode for small terminals.
	case terminal.LayoutCompact:
		// Return minimal service listing.
		return s.renderCompact(snap)
	// Standard mode for normal terminals.
	case terminal.LayoutNormal:
		// Return standard service table.
		return s.renderNormal(snap)
	// Wide mode with additional columns.
	case terminal.LayoutWide, terminal.LayoutUltraWide:
		// Return expanded service table.
		return s.renderWide(snap)
	}
	// Default to normal rendering.
	return s.renderNormal(snap)
}

// renderEmpty renders when there are no services.
//
// Returns:
//   - string: empty state message
func (s *ServicesRenderer) renderEmpty() string {
	box := widget.NewBox(s.width).
		SetTitle("Services").
		SetTitleColor(s.theme.Header).
		AddLine("  " + s.theme.Muted + "No services configured" + ansi.Reset)

	// Return rendered empty services box.
	return box.Render()
}

// renderCompact renders a minimal service list.
//
// Params:
//   - snap: system snapshot with service data
//
// Returns:
//   - string: compact service listing
func (s *ServicesRenderer) renderCompact(snap *model.Snapshot) string {
	const (
		// compactBuilderSize is the pre-allocated buffer size for compact rendering.
		compactBuilderSize int = 40
		// nameWidth is the width allocated for service names.
		nameWidth int = 10
	)

	lines := make([]string, 0, len(snap.Services)+1)

	// Format each service on a single line.
	for _, svc := range snap.Services {
		icon := s.status.ProcessState(svc.State)
		name := widget.Truncate(svc.Name, nameWidth)
		state := s.stateShort(svc.State)

		var extra string
		// Show CPU usage for running services.
		if svc.State == process.StateRunning {
			extra = widget.FormatPercent(svc.CPUPercent)
		}

		// Build line with strings.Builder to avoid fmt.Sprintf allocation.
		var sb strings.Builder
		sb.Grow(compactBuilderSize)
		sb.WriteString("  ")
		sb.WriteString(icon)
		sb.WriteByte(' ')
		sb.WriteString(name)
		// Pad name to fixed width.
		for i := len([]rune(name)); i < nameWidth; i++ {
			sb.WriteByte(' ')
		}
		sb.WriteByte(' ')
		sb.WriteString(state)
		sb.WriteByte(' ')
		sb.WriteString(extra)
		lines = append(lines, sb.String())
	}

	// Add summary line.
	summary := s.renderSummary(snap)
	lines = append(lines, "  "+s.theme.Muted+summary+ansi.Reset)

	box := widget.NewBox(s.width).
		SetTitle("Services").
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	// Return rendered compact services box.
	return box.Render()
}

// renderNormal renders a standard service table.
//
// Params:
//   - snap: system snapshot with service data
//
// Returns:
//   - string: standard service table
func (s *ServicesRenderer) renderNormal(snap *model.Snapshot) string {
	const (
		// boxBorderWidth is the total width consumed by box borders.
		boxBorderWidth int = 4
		// minTableWidth is the minimum width for the service table.
		minTableWidth int = 10
		// iconWidth is the width allocated for status icons.
		iconWidth int = 2
		// nameMinWidth is the minimum width for service names.
		nameMinWidth int = 8
		// stateWidth is the width allocated for process state display.
		stateWidth int = 8
		// pidWidth is the width allocated for PID display.
		pidWidth int = 6
		// uptimeWidth is the width allocated for uptime display.
		uptimeWidth int = 8
		// cpuWidth is the width allocated for CPU usage display.
		cpuWidth int = 5
		// memWidth is the width allocated for memory usage display.
		memWidth int = 6
	)

	// Build table (clamp width to prevent negative on tiny terminals).
	tableWidth := max(s.width-boxBorderWidth, minTableWidth)
	table := widget.NewTable(tableWidth).
		AddColumn("", iconWidth, widget.AlignLeft).            // Icon
		AddFlexColumn("NAME", nameMinWidth, widget.AlignLeft). // Name
		AddColumn("STATE", stateWidth, widget.AlignLeft).      // State
		AddColumn("PID", pidWidth, widget.AlignRight).         // PID
		AddColumn("UPTIME", uptimeWidth, widget.AlignRight).   // Uptime
		AddColumn("CPU", cpuWidth, widget.AlignRight).         // CPU
		AddColumn("MEM", memWidth, widget.AlignRight)          // Memory

	// Add row for each service.
	for _, svc := range snap.Services {
		icon := s.status.ProcessState(svc.State)
		state := s.status.ProcessStateText(svc.State)

		// Format PID or placeholder.
		pid := "-"
		// Convert PID to string when valid.
		if svc.PID > 0 {
			pid = strconv.Itoa(svc.PID)
		}

		// Format uptime for running/starting services.
		uptime := "-"
		// Show uptime for active services.
		if svc.State == process.StateRunning || svc.State == process.StateStarting {
			uptime = widget.FormatDurationShort(svc.Uptime)
		}

		// Format metrics for running services.
		cpu := "-"
		mem := "-"
		// Show metrics only for running services.
		if svc.State == process.StateRunning {
			cpu = widget.FormatPercent(svc.CPUPercent)
			mem = widget.FormatBytesShort(svc.MemoryRSS)
		}

		table.AddRow(icon, svc.Name, state, pid, uptime, cpu, mem)
	}

	// Split table into lines and add summary.
	lines := strings.Split(table.Render(), "\n")
	lines = append(lines, "")
	lines = append(lines, "  "+s.theme.Muted+s.renderSummary(snap)+ansi.Reset)

	box := widget.NewBox(s.width).
		SetTitle("Services").
		SetTitleColor(s.theme.Header).
		AddLines(prefixLines(lines, "  "))

	// Return rendered normal services box.
	return box.Render()
}

// renderWide renders an expanded service table.
//
// Params:
//   - snap: system snapshot with service data
//
// Returns:
//   - string: expanded service table with additional columns
func (s *ServicesRenderer) renderWide(snap *model.Snapshot) string {
	const (
		// boxBorderWidth is the total width consumed by box borders.
		boxBorderWidth int = 4
		// minTableWidth is the minimum width for the service table.
		minTableWidth int = 10
		// iconWidth is the width allocated for status icons.
		iconWidth int = 2
		// nameMinWidth is the minimum width for service names.
		nameMinWidth int = 10
		// stateWidth is the width allocated for process state display.
		stateWidth int = 8
		// pidWidth is the width allocated for PID display.
		pidWidth int = 7
		// uptimeWidth is the width allocated for uptime display.
		uptimeWidth int = 10
		// restartsWidth is the width allocated for restart count display.
		restartsWidth int = 8
		// healthWidth is the width allocated for health status display.
		healthWidth int = 8
		// cpuWidth is the width allocated for CPU usage display.
		cpuWidth int = 6
		// memWidth is the width allocated for memory usage display.
		memWidth int = 7
		// portsWidth is the width allocated for port listener display.
		portsWidth int = 12
	)

	// Build table with more columns (clamp width to prevent negative).
	tableWidth := max(s.width-boxBorderWidth, minTableWidth)
	table := widget.NewTable(tableWidth).
		AddColumn("", iconWidth, widget.AlignLeft).              // Icon
		AddFlexColumn("NAME", nameMinWidth, widget.AlignLeft).   // Name
		AddColumn("STATE", stateWidth, widget.AlignLeft).        // State
		AddColumn("PID", pidWidth, widget.AlignRight).           // PID
		AddColumn("UPTIME", uptimeWidth, widget.AlignRight).     // Uptime
		AddColumn("RESTARTS", restartsWidth, widget.AlignRight). // Restarts
		AddColumn("HEALTH", healthWidth, widget.AlignLeft).      // Health
		AddColumn("CPU", cpuWidth, widget.AlignRight).           // CPU
		AddColumn("MEM", memWidth, widget.AlignRight).           // Memory
		AddColumn("PORTS", portsWidth, widget.AlignLeft)         // Ports

	// Add row for each service.
	for _, svc := range snap.Services {
		icon := s.status.ProcessState(svc.State)
		state := s.status.ProcessStateText(svc.State)

		// Format PID or placeholder.
		pid := "-"
		// Convert PID to string when valid.
		if svc.PID > 0 {
			pid = strconv.Itoa(svc.PID)
		}

		// Format uptime for running/starting services.
		uptime := "-"
		// Show uptime for active services.
		if svc.State == process.StateRunning || svc.State == process.StateStarting {
			uptime = widget.FormatDuration(svc.Uptime)
		}

		// Format restart count.
		restarts := "-"
		// Show restart count when service has restarted.
		if svc.RestartCount > 0 {
			restarts = strconv.Itoa(svc.RestartCount)
		}

		// Format health status.
		healthStr := s.status.HealthStatusText(svc.Health)

		// Format metrics for running services.
		cpu := "-"
		mem := "-"
		// Show metrics only for running services.
		if svc.State == process.StateRunning {
			cpu = widget.FormatPercent(svc.CPUPercent)
			mem = widget.FormatBytesShort(svc.MemoryRSS)
		}

		// Format port listeners.
		ports := s.formatPorts(svc.Listeners)

		table.AddRow(icon, svc.Name, state, pid, uptime, restarts, healthStr, cpu, mem, ports)
	}

	// Split table into lines and add summary.
	lines := strings.Split(table.Render(), "\n")
	lines = append(lines, "")
	lines = append(lines, "  "+s.theme.Muted+s.renderSummary(snap)+ansi.Reset)

	box := widget.NewBox(s.width).
		SetTitle("Services").
		SetTitleColor(s.theme.Header).
		AddLines(prefixLines(lines, "  "))

	// Return rendered wide services box.
	return box.Render()
}

// renderSummary returns a summary line.
//
// Params:
//   - snap: system snapshot with service data
//
// Returns:
//   - string: summary statistics
func (s *ServicesRenderer) renderSummary(snap *model.Snapshot) string {
	total := len(snap.Services)
	running := snap.RunningCount()
	failed := snap.FailedCount()
	healthy := snap.HealthyCount()

	parts := []string{strconv.Itoa(total) + " services"}

	// Add running count if non-zero.
	if running > 0 {
		parts = append(parts, strconv.Itoa(running)+" running")
	}
	// Add failed count if non-zero.
	if failed > 0 {
		parts = append(parts, strconv.Itoa(failed)+" failed")
	}
	// Add healthy count if different from running.
	if healthy > 0 && healthy != running {
		parts = append(parts, strconv.Itoa(healthy)+" healthy")
	}

	// Return formatted summary string.
	return "[" + strings.Join(parts, ", ") + "]"
}

// stateShort returns a short state string.
//
// Params:
//   - state: process state
//
// Returns:
//   - string: abbreviated state string
func (s *ServicesRenderer) stateShort(state process.State) string {
	// Return abbreviated state from status indicator.
	return s.status.ProcessStateShort(state)
}

// formatPorts formats listener ports with colors based on status.
// Colors: Green (OK), Yellow (Warning), Red (Error).
//
// Params:
//   - listeners: port listeners for a service
//
// Returns:
//   - string: formatted port list with status colors
func (s *ServicesRenderer) formatPorts(listeners []model.ListenerSnapshot) string {
	// Return placeholder when no listeners.
	if len(listeners) == 0 {
		// Return dash for empty listeners.
		return "-"
	}

	var sb strings.Builder
	// Add each listener with status color.
	for i, l := range listeners {
		// Add space separator between ports.
		if i > 0 {
			sb.WriteString(" ")
		}

		// Choose color based on port status.
		var color string
		// Select color based on port status.
		switch l.Status {
		// Green for OK ports.
		case model.PortStatusOK:
			color = s.theme.Success
		// Yellow for warning ports.
		case model.PortStatusWarning:
			color = s.theme.Warning
		// Red for error ports.
		case model.PortStatusError:
			color = s.theme.Error
		// Muted for unknown status.
		case model.PortStatusUnknown:
			color = s.theme.Muted
		}

		sb.WriteString(color)
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(l.Port))
		sb.WriteString(ansi.Reset)
	}

	// Return formatted port string.
	return sb.String()
}

// prefixLines adds a prefix to each line.
//
// Params:
//   - lines: input lines
//   - prefix: prefix to add
//
// Returns:
//   - []string: prefixed lines
func prefixLines(lines []string, prefix string) []string {
	result := make([]string, len(lines))
	// Add prefix to each line.
	// Iterate through lines to add prefix.
	for i, line := range lines {
		result[i] = prefix + line
	}
	// Return prefixed lines array.
	return result
}

// RenderNamesOnly renders service names with ports in columns (for raw mode startup banner).
// Shows service name followed by port numbers from config (no colors, no status checking).
// Column width is dynamically calculated based on content.
//
// Params:
//   - snap: system snapshot with service data
//
// Returns:
//   - string: service names with ports in column layout
func (s *ServicesRenderer) RenderNamesOnly(snap *model.Snapshot) string {
	const (
		// boxBorderWidth is the total width consumed by box borders.
		boxBorderWidth int = 6
		// minColumnWidth is the minimum width for service columns.
		minColumnWidth int = 10
		// columnPadding is the padding between columns.
		columnPadding int = 2
	)

	// Show empty state if no services.
	if len(snap.Services) == 0 {
		box := widget.NewBox(s.width).
			SetTitle("Services (0 configured)").
			SetTitleColor(s.theme.Header).
			AddLine("  " + s.theme.Muted + "No services configured" + ansi.Reset)
		// Return empty services box for raw mode.
		return box.Render()
	}

	// Build service entries with name and ports (plain text, no colors).
	type serviceEntry struct {
		display    string // Full display string.
		visibleLen int    // Length of display string.
	}
	entries := make([]serviceEntry, 0, len(snap.Services))

	// Format each service entry.
	for _, svc := range snap.Services {
		var sb strings.Builder
		sb.WriteString(svc.Name)

		// Add ports if available (plain text from config, no status colors).
		// Append port list when listeners exist.
		if len(svc.Listeners) > 0 {
			sb.WriteByte(' ')
			// Iterate through listeners to format ports.
			for i, l := range svc.Listeners {
				// Add comma separator between ports.
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
	// Iterate to find maximum entry length.
	for _, e := range entries {
		// Update max if entry is longer.
		if e.visibleLen > maxLen {
			maxLen = e.visibleLen
		}
	}

	// Calculate column width with padding.
	colWidth := max(maxLen+columnPadding, minColumnWidth)

	// Calculate number of columns that fit.
	usableWidth := s.width - boxBorderWidth
	cols := max(usableWidth/colWidth, 1)

	// Build rows of services.
	rows := (len(entries) + cols - 1) / cols
	lines := make([]string, rows)

	// Format each row.
	// Iterate through rows to build column layout.
	for row := range rows {
		var rowParts []string
		// Add entries for this row.
		// Iterate through columns.
		for col := range cols {
			idx := row*cols + col
			// Add entry if index is valid.
			if idx < len(entries) {
				entry := entries[idx]
				// Pad to column width using visible length.
				padding := max(colWidth-entry.visibleLen, 0)
				rowParts = append(rowParts, entry.display+strings.Repeat(" ", padding))
			}
		}
		lines[row] = "  " + strings.Join(rowParts, "")
	}

	title := "Services (" + strconv.Itoa(len(snap.Services)) + " configured)"
	box := widget.NewBox(s.width).
		SetTitle(title).
		SetTitleColor(s.theme.Header).
		AddLines(lines)

	// Return rendered services box for raw mode.
	return box.Render()
}

// NOTE: Tests needed - create services_external_test.go and services_internal_test.go
