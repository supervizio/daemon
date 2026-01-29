// Package screen provides complete screen renderers.
package screen

import (
	"strconv"
	"strings"
	"time"

	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// AddRower defines the minimal interface for adding rows to a table (KTN-API-MINIF).
type AddRower interface {
	AddRow(cells ...string) *widget.Table
}

// Renderer defines the minimal interface for rendering a table (KTN-API-MINIF).
type Renderer interface {
	Render() string
}

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
	// handle default case.
	default:
		// Default to normal rendering.
		return s.renderNormal(snap)
	}
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
	table := s.createNormalTable()
	s.populateNormalRows(table, snap.Services)
	// return computed result.
	return s.renderTableInBox(table, snap)
}

// createNormalTable creates a table with normal mode columns.
//
// Returns:
//   - *widget.Table: configured table with standard columns
func (s *ServicesRenderer) createNormalTable() *widget.Table {
	const (
		// boxBorderWidth is the total width of box borders.
		boxBorderWidth int = 4
		// minTableWidth is the minimum table width.
		minTableWidth int = 10
		// iconWidth is the width of the status icon column.
		iconWidth int = 2
		// nameMinWidth is the minimum width for service name column.
		nameMinWidth int = 8
		// stateWidth is the width of the state column.
		stateWidth int = 8
		// pidWidth is the width of the PID column.
		pidWidth int = 6
		// uptimeWidth is the width of the uptime column.
		uptimeWidth int = 8
		// cpuWidth is the width of the CPU column.
		cpuWidth int = 5
		// memWidth is the width of the memory column.
		memWidth int = 6
	)

	tableWidth := max(s.width-boxBorderWidth, minTableWidth)
	// return computed result.
	return widget.NewTable(tableWidth).
		AddColumn("", iconWidth, widget.AlignLeft).
		AddFlexColumn("NAME", nameMinWidth, widget.AlignLeft).
		AddColumn("STATE", stateWidth, widget.AlignLeft).
		AddColumn("PID", pidWidth, widget.AlignRight).
		AddColumn("UPTIME", uptimeWidth, widget.AlignRight).
		AddColumn("CPU", cpuWidth, widget.AlignRight).
		AddColumn("MEM", memWidth, widget.AlignRight)
}

// populateNormalRows adds service rows to a normal mode table.
//
// Params:
//   - table: target table to populate (uses AddRower interface)
//   - services: services to add as rows
func (s *ServicesRenderer) populateNormalRows(table AddRower, services []model.ServiceSnapshot) {
	// iterate over collection.
	for _, svc := range services {
		icon := s.status.ProcessState(svc.State)
		state := s.status.ProcessStateText(svc.State)
		pid := s.formatPID(svc.PID)
		uptime := s.formatUptimeShort(svc.State, svc.Uptime)
		cpu, mem := s.formatMetrics(svc.State, svc.CPUPercent, svc.MemoryRSS)
		table.AddRow(icon, svc.Name, state, pid, uptime, cpu, mem)
	}
}

// formatPID formats a PID value or returns placeholder.
//
// Params:
//   - pid: process ID to format
//
// Returns:
//   - string: formatted PID or "-"
func (s *ServicesRenderer) formatPID(pid int) string {
	// check for positive value.
	if pid > 0 {
		// return computed result.
		return strconv.Itoa(pid)
	}
	// return computed result.
	return "-"
}

// formatUptimeShort formats uptime for running/starting services (short format).
//
// Params:
//   - state: current process state
//   - uptime: duration since start
//
// Returns:
//   - string: formatted uptime or "-"
func (s *ServicesRenderer) formatUptimeShort(state process.State, uptime time.Duration) string {
	// evaluate condition.
	if state == process.StateRunning || state == process.StateStarting {
		// return computed result.
		return widget.FormatDurationShort(uptime)
	}
	// return computed result.
	return "-"
}

// formatMetrics formats CPU and memory for running services.
//
// Params:
//   - state: current process state
//   - cpuPercent: CPU usage percentage
//   - memoryRSS: memory usage in bytes
//
// Returns:
//   - string: formatted CPU percentage or "-"
//   - string: formatted memory or "-"
func (s *ServicesRenderer) formatMetrics(state process.State, cpuPercent float64, memoryRSS uint64) (cpu, mem string) {
	// evaluate condition.
	if state == process.StateRunning {
		// return computed result.
		return widget.FormatPercent(cpuPercent), widget.FormatBytesShort(memoryRSS)
	}
	// return computed result.
	return "-", "-"
}

// renderTableInBox wraps table output in a services box with summary.
//
// Params:
//   - table: rendered table (uses Renderer interface)
//   - snap: snapshot for summary
//
// Returns:
//   - string: rendered box containing table
func (s *ServicesRenderer) renderTableInBox(table Renderer, snap *model.Snapshot) string {
	lines := strings.Split(table.Render(), "\n")
	lines = append(lines, "")
	lines = append(lines, "  "+s.theme.Muted+s.renderSummary(snap)+ansi.Reset)

	box := widget.NewBox(s.width).
		SetTitle("Services").
		SetTitleColor(s.theme.Header).
		AddLines(prefixLines(lines, "  "))

	// return computed result.
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
	table := s.createWideTable()
	s.populateWideRows(table, snap.Services)
	// return computed result.
	return s.renderTableInBox(table, snap)
}

// createWideTable creates a table with wide mode columns.
//
// Returns:
//   - *widget.Table: configured table with extended columns
func (s *ServicesRenderer) createWideTable() *widget.Table {
	const (
		// boxBorderWidth is the total width of box borders.
		boxBorderWidth int = 4
		// minTableWidth is the minimum table width.
		minTableWidth int = 10
		// iconWidth is the width of the status icon column.
		iconWidth int = 2
		// nameMinWidth is the minimum width for service name column.
		nameMinWidth int = 10
		// stateWidth is the width of the state column.
		stateWidth int = 8
		// pidWidth is the width of the PID column.
		pidWidth int = 7
		// uptimeWidth is the width of the uptime column.
		uptimeWidth int = 10
		// restartsWidth is the width of the restarts column.
		restartsWidth int = 8
		// healthWidth is the width of the health column.
		healthWidth int = 8
		// cpuWidth is the width of the CPU column.
		cpuWidth int = 6
		// memWidth is the width of the memory column.
		memWidth int = 7
		// portsWidth is the width of the ports column.
		portsWidth int = 12
	)

	tableWidth := max(s.width-boxBorderWidth, minTableWidth)
	// return computed result.
	return widget.NewTable(tableWidth).
		AddColumn("", iconWidth, widget.AlignLeft).
		AddFlexColumn("NAME", nameMinWidth, widget.AlignLeft).
		AddColumn("STATE", stateWidth, widget.AlignLeft).
		AddColumn("PID", pidWidth, widget.AlignRight).
		AddColumn("UPTIME", uptimeWidth, widget.AlignRight).
		AddColumn("RESTARTS", restartsWidth, widget.AlignRight).
		AddColumn("HEALTH", healthWidth, widget.AlignLeft).
		AddColumn("CPU", cpuWidth, widget.AlignRight).
		AddColumn("MEM", memWidth, widget.AlignRight).
		AddColumn("PORTS", portsWidth, widget.AlignLeft)
}

// populateWideRows adds service rows to a wide mode table.
//
// Params:
//   - table: target table to populate (uses AddRower interface)
//   - services: services to add as rows
func (s *ServicesRenderer) populateWideRows(table AddRower, services []model.ServiceSnapshot) {
	// iterate over collection.
	for _, svc := range services {
		icon := s.status.ProcessState(svc.State)
		state := s.status.ProcessStateText(svc.State)
		pid := s.formatPID(svc.PID)
		uptime := s.formatUptimeLong(svc.State, svc.Uptime)
		restarts := s.formatRestarts(svc.RestartCount)
		healthStr := s.status.HealthStatusText(svc.Health)
		cpu, mem := s.formatMetrics(svc.State, svc.CPUPercent, svc.MemoryRSS)
		ports := s.formatPorts(svc.Listeners)
		table.AddRow(icon, svc.Name, state, pid, uptime, restarts, healthStr, cpu, mem, ports)
	}
}

// formatUptimeLong formats uptime for running/starting services (long format).
//
// Params:
//   - state: current process state
//   - uptime: duration since start
//
// Returns:
//   - string: formatted uptime or "-"
func (s *ServicesRenderer) formatUptimeLong(state process.State, uptime time.Duration) string {
	// evaluate condition.
	if state == process.StateRunning || state == process.StateStarting {
		// return computed result.
		return widget.FormatDuration(uptime)
	}
	// return computed result.
	return "-"
}

// formatRestarts formats restart count or returns placeholder.
//
// Params:
//   - count: number of restarts
//
// Returns:
//   - string: formatted count or "-"
func (s *ServicesRenderer) formatRestarts(count int) string {
	// check for positive value.
	if count > 0 {
		// return computed result.
		return strconv.Itoa(count)
	}
	// return computed result.
	return "-"
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
		// handle default case.
		default:
			// Handle unknown port status values.
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
	// Pre-allocate with capacity using append pattern per VAR-MAKEAPPEND.
	result := make([]string, 0, len(lines))
	// Add prefix to each line using append.
	for _, line := range lines {
		result = append(result, prefix+line)
	}
	// Return prefixed lines array.
	return result
}

// RenderNamesOnly renders service names with ports in columns (for raw mode startup banner).
// Shows service name followed by port numbers from config (no colors, no status checking).
//
// Params:
//   - snap: system snapshot with service data
//
// Returns:
//   - string: service names with ports in column layout
func (s *ServicesRenderer) RenderNamesOnly(snap *model.Snapshot) string {
	// check for empty value.
	if len(snap.Services) == 0 {
		// return computed result.
		return s.renderEmptyServices()
	}
	entries := s.buildServiceEntries(snap.Services)
	lines := s.layoutEntriesInColumns(entries)
	// return computed result.
	return s.renderNamesOnlyBox(snap, lines)
}

// renderEmptyServices renders the empty state box.
//
// Returns:
//   - string: rendered empty services box
func (s *ServicesRenderer) renderEmptyServices() string {
	box := widget.NewBox(s.width).
		SetTitle("Services (0 configured)").
		SetTitleColor(s.theme.Header).
		AddLine("  " + s.theme.Muted + "No services configured" + ansi.Reset)
	// return computed result.
	return box.Render()
}

// buildServiceEntries creates display entries for all services.
//
// Params:
//   - services: services to format
//
// Returns:
//   - []serviceEntry: formatted entries with visible lengths
func (s *ServicesRenderer) buildServiceEntries(services []model.ServiceSnapshot) []serviceEntry {
	entries := make([]serviceEntry, 0, len(services))
	// iterate over collection.
	for _, svc := range services {
		display := s.formatServiceEntry(svc)
		entries = append(entries, serviceEntry{display: display, visibleLen: len(display)})
	}
	// return computed result.
	return entries
}

// formatServiceEntry formats a single service with its ports.
//
// Params:
//   - svc: service snapshot to format
//
// Returns:
//   - string: formatted service name with ports
func (s *ServicesRenderer) formatServiceEntry(svc model.ServiceSnapshot) string {
	var sb strings.Builder
	sb.WriteString(svc.Name)
	// check for positive value.
	if len(svc.Listeners) > 0 {
		sb.WriteByte(' ')
		// iterate over collection.
		for i, l := range svc.Listeners {
			// check for positive value.
			if i > 0 {
				sb.WriteByte(',')
			}
			sb.WriteByte(':')
			sb.WriteString(strconv.Itoa(l.Port))
		}
	}
	// return computed result.
	return sb.String()
}

// layoutEntriesInColumns arranges entries in a column layout.
//
// Params:
//   - entries: entries to layout
//
// Returns:
//   - []string: lines with entries arranged in columns
func (s *ServicesRenderer) layoutEntriesInColumns(entries []serviceEntry) []string {
	const (
		// boxBorderWidth is the total width of box borders.
		boxBorderWidth int = 6
		// minColumnWidth is the minimum width for a column.
		minColumnWidth int = 10
		// columnPadding is the padding between columns.
		columnPadding int = 2
	)

	colWidth, cols := s.calculateColumnLayout(entries, boxBorderWidth, minColumnWidth, columnPadding)
	rows := (len(entries) + cols - 1) / cols
	lines := make([]string, 0, rows)
	rowParts := make([]string, 0, cols)

	// iterate over collection.
	for row := range rows {
		rowParts = rowParts[:0]
		// iterate over collection.
		for col := range cols {
			idx := row*cols + col
			// evaluate condition.
			if idx < len(entries) {
				entry := entries[idx]
				padding := max(colWidth-entry.visibleLen, 0)
				rowParts = append(rowParts, entry.display+strings.Repeat(" ", padding))
			}
		}
		lines = append(lines, "  "+strings.Join(rowParts, ""))
	}
	// return computed result.
	return lines
}

// calculateColumnLayout computes column width and count.
//
// Params:
//   - entries: entries to calculate layout for
//   - borderWidth: box border width
//   - minWidth: minimum column width
//   - padding: padding between columns
//
// Returns:
//   - int: calculated column width
//   - int: number of columns
func (s *ServicesRenderer) calculateColumnLayout(entries []serviceEntry, borderWidth, minWidth, padding int) (colWidth, cols int) {
	maxLen := 0
	// iterate over collection.
	for _, e := range entries {
		// evaluate condition.
		if e.visibleLen > maxLen {
			maxLen = e.visibleLen
		}
	}
	colWidth = max(maxLen+padding, minWidth)
	usableWidth := s.width - borderWidth
	cols = max(usableWidth/colWidth, 1)
	// return computed result.
	return colWidth, cols
}

// renderNamesOnlyBox creates the final box for names-only rendering.
//
// Params:
//   - snap: snapshot for title
//   - lines: content lines
//
// Returns:
//   - string: rendered box
func (s *ServicesRenderer) renderNamesOnlyBox(snap *model.Snapshot, lines []string) string {
	title := "Services (" + strconv.Itoa(len(snap.Services)) + " configured)"
	box := widget.NewBox(s.width).
		SetTitle(title).
		SetTitleColor(s.theme.Header).
		AddLines(lines)
	// return computed result.
	return box.Render()
}

// NOTE: Tests needed - create services_external_test.go and services_internal_test.go
