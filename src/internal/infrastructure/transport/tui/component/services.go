// Package component provides reusable Bubble Tea components.
package component

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// maxServicesVisible is the maximum number of services to show before scrolling.
const maxServicesVisible int = 10

// ServicesPanel is a scrollable services list with vertical scrollbar.
//
// It displays service status including state, health, uptime, PID, restarts,
// CPU, memory usage, and listening ports with color-coded status indicators.
type ServicesPanel struct {
	viewport viewport.Model
	theme    ansi.Theme
	width    int
	height   int
	services []model.ServiceSnapshot
	focused  bool
	title    string
}

const (
	// Border and scrollbar dimensions.
	borderWidth     int = 3 // left border + right border + scrollbar
	borderHeight    int = 2 // top + bottom borders
	headerRowHeight int = 1 // header row after title

	// Column widths for service display.
	nameColWidth int = 12
	// stateColWidth is the width of the state column in characters.
	stateColWidth int = 9
	// healthColWidth is the width of the health column in characters.
	healthColWidth int = 9
	// uptimeColWidth is the width of the uptime column in characters.
	uptimeColWidth int = 7
	// pidColWidth is the width of the PID column in characters.
	pidColWidth int = 6
	// restartsColWidth is the width of the restart count column in characters.
	restartsColWidth int = 3
	// cpuColWidth is the width of the CPU usage column in characters.
	cpuColWidth int = 5
	// memColWidth is the width of the memory usage column in characters.
	memColWidth int = 8

	// Text truncation limits.
	maxPortsTextLen int = 9 // truncate ports display after this length

	// Layout constants.
	optimalHeightExtra int = 3 // borders + header for optimal height calculation

	// serviceTitlePrefixLen is the length of "- " + " " surrounding the title.
	serviceTitlePrefixLen int = 3

	// dashCountOffset is the offset for final border character.
	serviceDashCountOffset int = 1

	// scrollbarColWidthSvc is the scrollbar column width for services panel.
	scrollbarColWidthSvc int = 1

	// minContentHeight is the minimum content area height.
	minContentHeight int = 1

	// minServiceLines is the minimum lines for service display.
	minServiceLines int = 1

	// minThumbSizeSvc is the minimum scrollbar thumb size.
	minThumbSizeSvc int = 1
)

// NewServicesPanel creates a new services panel.
//
// Params:
//   - width: panel width including borders
//   - height: panel height including borders
//
// Returns:
//   - ServicesPanel: initialized services panel
func NewServicesPanel(width, height int) ServicesPanel {
	// Calculate viewport size accounting for borders and scrollbar.
	vp := viewport.New(width-borderWidth, height-borderHeight)

	// Return initialized panel with default settings.
	return ServicesPanel{
		viewport: vp,
		theme:    ansi.DefaultTheme(),
		width:    width,
		height:   height,
		services: nil,
		title:    "Services",
	}
}

// SetSize updates the panel dimensions.
//
// Params:
//   - width: new panel width
//   - height: new panel height
func (s *ServicesPanel) SetSize(width, height int) {
	s.width = width
	s.height = height
	// Adjust viewport for borders and scrollbar.
	s.viewport.Width = width - borderWidth
	s.viewport.Height = height - borderHeight
	s.updateContent()
}

// SetFocused sets the focus state.
//
// Params:
//   - focused: true to focus the panel, false to unfocus
func (s *ServicesPanel) SetFocused(focused bool) {
	s.focused = focused
}

// Focused returns whether the panel is focused.
//
// Returns:
//   - bool: true if panel is focused
func (s *ServicesPanel) Focused() bool {
	// Return current focus state.
	return s.focused
}

// SetServices updates the service list.
//
// Params:
//   - services: slice of service snapshots to display
func (s *ServicesPanel) SetServices(services []model.ServiceSnapshot) {
	s.services = services
	s.updateContent()
}

// OptimalHeight returns the optimal height including borders and header.
//
// Returns:
//   - int: optimal panel height in lines
func (s *ServicesPanel) OptimalHeight() int {
	contentLines := len(s.services)

	// Cap at maximum visible services.
	contentLines = min(contentLines, maxServicesVisible)
	// Ensure at least one line for "no services" message.
	contentLines = max(contentLines, minServiceLines)

	// Add borders and header row.
	return contentLines + optimalHeightExtra
}

// updateContent rebuilds the viewport content.
func (s *ServicesPanel) updateContent() {
	var sb strings.Builder

	// Show message if no services configured.
	if len(s.services) == 0 {
		sb.WriteString(s.theme.Muted + " No services configured" + ansi.Reset + "\n")
		s.viewport.SetContent(sb.String())
		// Exit early when no services.
		return
	}

	// Build content for each service.
	for _, svc := range s.services {
		line := s.formatServiceLine(svc)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	s.viewport.SetContent(sb.String())
}

// formatServiceLine formats a single service entry as a display line.
//
// Params:
//   - svc: service snapshot to format
//
// Returns:
//   - string: formatted service line with ANSI colors
func (s *ServicesPanel) formatServiceLine(svc model.ServiceSnapshot) string {
	// Format service name with truncation.
	name := s.formatServiceName(svc.Name)

	// Collect all column values.
	cols := s.collectServiceColumns(svc)

	// Build the final line string.
	return s.buildServiceLineString(name, &cols)
}

// serviceColumns holds pre-formatted column values for a service line.
type serviceColumns struct {
	stateIcon  string
	stateText  string
	healthText string
	uptime     string
	pid        string
	restarts   string
	cpu        string
	mem        string
	ports      string
}

// formatServiceName truncates and formats a service name.
//
// Params:
//   - name: raw service name
//
// Returns:
//   - string: formatted service name
func (s *ServicesPanel) formatServiceName(name string) string {
	// Truncate long names.
	if len([]rune(name)) > nameColWidth {
		// Return truncated name.
		return widget.TruncateRunes(name, nameColWidth, "...")
	}
	// Return original name.
	return name
}

// collectServiceColumns gathers all formatted column values for a service.
//
// Params:
//   - svc: service snapshot
//
// Returns:
//   - serviceColumns: struct with all formatted column values
func (s *ServicesPanel) collectServiceColumns(svc model.ServiceSnapshot) serviceColumns {
	// Return all formatted columns.
	return serviceColumns{
		stateIcon:  s.getStateIcon(svc.State),
		stateText:  s.getStateText(svc.State),
		healthText: s.getHealthText(svc),
		uptime:     s.formatUptime(svc),
		pid:        s.formatPID(svc),
		restarts:   s.formatRestarts(svc),
		cpu:        s.formatCPU(svc),
		mem:        s.formatMemory(svc),
		ports:      s.formatPortsWithStatus(svc),
	}
}

// buildServiceLineString assembles the final service line string.
//
// Params:
//   - name: formatted service name
//   - cols: pointer to pre-formatted column values
//
// Returns:
//   - string: complete formatted line
func (s *ServicesPanel) buildServiceLineString(name string, cols *serviceColumns) string {
	var sb strings.Builder

	// Build line with manual padding for ANSI-colored strings.
	sb.WriteString(" ")
	sb.WriteString(cols.stateIcon)
	sb.WriteString(" ")
	sb.WriteString(widget.PadRight(name, nameColWidth))
	sb.WriteString(" ")
	sb.WriteString(widget.PadRightAnsi(cols.stateText, stateColWidth))
	sb.WriteString(" ")
	sb.WriteString(widget.PadRightAnsi(cols.healthText, healthColWidth))
	sb.WriteString(" ")
	sb.WriteString(widget.PadLeft(cols.uptime, uptimeColWidth))
	sb.WriteString(" ")
	sb.WriteString(widget.PadLeft(cols.pid, pidColWidth))
	sb.WriteString(" ")
	sb.WriteString(widget.PadLeft(cols.restarts, restartsColWidth))
	sb.WriteString(" ")
	sb.WriteString(widget.PadLeft(cols.cpu, cpuColWidth))
	sb.WriteString(" ")
	sb.WriteString(widget.PadLeft(cols.mem, memColWidth))
	sb.WriteString(" ")
	sb.WriteString(cols.ports)

	// Return formatted line.
	return sb.String()
}

// formatUptime formats the uptime for a service.
//
// Params:
//   - svc: service snapshot
//
// Returns:
//   - string: formatted uptime string
func (s *ServicesPanel) formatUptime(svc model.ServiceSnapshot) string {
	// Show dash for non-running services.
	if svc.State != process.StateRunning || svc.Uptime <= 0 {
		// Return placeholder.
		return "-"
	}
	// Return formatted duration.
	return widget.FormatDurationShort(svc.Uptime)
}

// formatPID formats the PID for a service.
//
// Params:
//   - svc: service snapshot
//
// Returns:
//   - string: formatted PID string
func (s *ServicesPanel) formatPID(svc model.ServiceSnapshot) string {
	// Show dash if no PID.
	if svc.PID <= 0 {
		// Return placeholder.
		return "-"
	}
	// Return PID as string.
	return strconv.Itoa(svc.PID)
}

// formatRestarts formats the restart count for a service.
//
// Params:
//   - svc: service snapshot
//
// Returns:
//   - string: formatted restart count
func (s *ServicesPanel) formatRestarts(svc model.ServiceSnapshot) string {
	// Show dash if no restarts.
	if svc.RestartCount <= 0 {
		// Return placeholder.
		return "-"
	}
	// Return restart count as string.
	return strconv.Itoa(svc.RestartCount)
}

// formatCPU formats the CPU percentage for a service.
//
// Params:
//   - svc: service snapshot
//
// Returns:
//   - string: formatted CPU percentage
func (s *ServicesPanel) formatCPU(svc model.ServiceSnapshot) string {
	// Show dash for non-running services.
	if svc.State != process.StateRunning {
		// Return placeholder.
		return "-"
	}
	// Return formatted percentage.
	return fmt.Sprintf("%.1f%%", svc.CPUPercent)
}

// formatMemory formats the memory usage for a service.
//
// Params:
//   - svc: service snapshot
//
// Returns:
//   - string: formatted memory usage
func (s *ServicesPanel) formatMemory(svc model.ServiceSnapshot) string {
	// Show dash for non-running or zero memory.
	if svc.State != process.StateRunning || svc.MemoryRSS <= 0 {
		// Return placeholder.
		return "-"
	}
	// Return formatted bytes.
	return widget.FormatBytes(svc.MemoryRSS)
}

// formatPorts formats a list of ports as comma-separated string with truncation.
//
// Params:
//   - ports: slice of port numbers to format
//
// Returns:
//   - string: formatted port string or dash if empty
func formatPorts(ports []int) string {
	// Return dash if no ports.
	if len(ports) == 0 {
		// Return placeholder.
		return "-"
	}

	var sb strings.Builder

	// Build comma-separated port list.
	for i, port := range ports {
		// Add comma separator.
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(strconv.Itoa(port))

		// Truncate if getting too long.
		if sb.Len() > maxPortsTextLen && i < len(ports)-1 {
			sb.WriteString("...")
			// Stop processing more ports.
			break
		}
	}

	// Return formatted port list.
	return sb.String()
}

// formatPortsWithStatus formats ports with color codes based on listener status.
//
// Params:
//   - svc: service snapshot with listener status and port information
//
// Returns:
//   - string: formatted ports with ANSI color codes
func (s *ServicesPanel) formatPortsWithStatus(svc model.ServiceSnapshot) string {
	// Show detected ports in muted color if no listeners configured.
	if len(svc.Listeners) == 0 {
		// Check for detected ports.
		if len(svc.Ports) == 0 {
			// Return muted dash.
			return s.theme.Muted + "-" + ansi.Reset
		}
		// Show detected ports (not configured but listening).
		return s.theme.Muted + formatPorts(svc.Ports) + ansi.Reset
	}

	var sb strings.Builder

	// Format each listener with status color.
	for i, l := range svc.Listeners {
		// Add separator between ports.
		if i > 0 {
			sb.WriteString(" ")
		}

		// Choose color based on status.
		color := s.getPortStatusColor(l.Status)

		// Format: :PORT (with color).
		sb.WriteString(color)
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(l.Port))
		sb.WriteString(ansi.Reset)
	}

	// Return formatted ports.
	return sb.String()
}

// getPortStatusColor returns the color for a port status.
//
// Params:
//   - status: port status value
//
// Returns:
//   - string: ANSI color code
func (s *ServicesPanel) getPortStatusColor(status model.PortStatus) string {
	// Map status to color.
	switch status {
	// OK status gets green.
	case model.PortStatusOK:
		// Return success color.
		return s.theme.Success
	// Warning status gets yellow.
	case model.PortStatusWarning:
		// Return warning color.
		return s.theme.Warning
	// Error status gets red.
	case model.PortStatusError:
		// Return error color.
		return s.theme.Error
	// Unknown status gets muted.
	case model.PortStatusUnknown:
		// Return muted color.
		return s.theme.Muted
	// Default to muted.
	default:
		// Return muted color.
		return s.theme.Muted
	}
}

// getStateIcon returns the state icon with color.
//
// Params:
//   - state: process state
//
// Returns:
//   - string: colored icon representing the state
func (s *ServicesPanel) getStateIcon(state process.State) string {
	// Map state to colored icon.
	switch state {
	// Running state gets green filled circle.
	case process.StateRunning:
		// Return success icon.
		return s.theme.Success + "o" + ansi.Reset
	// Stopped state gets muted empty circle.
	case process.StateStopped:
		// Return muted icon.
		return s.theme.Muted + "o" + ansi.Reset
	// Failed state gets red filled circle.
	case process.StateFailed:
		// Return error icon.
		return s.theme.Error + "o" + ansi.Reset
	// Starting state gets warning half-filled.
	case process.StateStarting:
		// Return warning icon.
		return s.theme.Warning + "o" + ansi.Reset
	// Stopping state gets warning half-filled.
	case process.StateStopping:
		// Return warning icon.
		return s.theme.Warning + "o" + ansi.Reset
	// Default to muted empty circle.
	default:
		// Return muted icon.
		return s.theme.Muted + "o" + ansi.Reset
	}
}

// getStateText returns the state text with color.
//
// Params:
//   - state: process state
//
// Returns:
//   - string: colored text representing the state
func (s *ServicesPanel) getStateText(state process.State) string {
	// Map state to colored text.
	switch state {
	// Running state.
	case process.StateRunning:
		// Return success text.
		return s.theme.Success + "running" + ansi.Reset
	// Stopped state.
	case process.StateStopped:
		// Return muted text.
		return s.theme.Muted + "stopped" + ansi.Reset
	// Failed state.
	case process.StateFailed:
		// Return error text.
		return s.theme.Error + "failed" + ansi.Reset
	// Starting state.
	case process.StateStarting:
		// Return warning text.
		return s.theme.Warning + "starting" + ansi.Reset
	// Stopping state.
	case process.StateStopping:
		// Return warning text.
		return s.theme.Warning + "stopping" + ansi.Reset
	// Unknown state.
	default:
		// Return muted text.
		return s.theme.Muted + "unknown" + ansi.Reset
	}
}

// getHealthText returns the health status text with color.
//
// Params:
//   - svc: service snapshot
//
// Returns:
//   - string: colored health status text
func (s *ServicesPanel) getHealthText(svc model.ServiceSnapshot) string {
	// Only show health for running services with health checks.
	if svc.State != process.StateRunning {
		// Return muted dash.
		return s.theme.Muted + "-" + ansi.Reset
	}

	// Show dash if no health checks configured.
	if !svc.HasHealthChecks {
		// Return muted dash.
		return s.theme.Muted + "-" + ansi.Reset
	}

	// Map health status to colored text.
	switch svc.Health {
	// Healthy status.
	case health.StatusHealthy:
		// Return success text.
		return s.theme.Success + "healthy" + ansi.Reset
	// Unhealthy status.
	case health.StatusUnhealthy:
		// Return error text.
		return s.theme.Error + "unhealthy" + ansi.Reset
	// Degraded status.
	case health.StatusDegraded:
		// Return warning text.
		return s.theme.Warning + "degraded" + ansi.Reset
	// Unknown/pending status.
	case health.StatusUnknown:
		// Return warning text.
		return s.theme.Warning + "pending" + ansi.Reset
	// Default to muted dash.
	default:
		// Return muted dash.
		return s.theme.Muted + "-" + ansi.Reset
	}
}

// Init initializes the component.
//
// Returns:
//   - tea.Cmd: initialization command (nil)
func (s *ServicesPanel) Init() tea.Cmd {
	// Return nil as no initialization needed.
	return nil
}

// Update handles messages.
//
// Params:
//   - msg: Bubble Tea message
//
// Returns:
//   - *ServicesPanel: updated panel
//   - tea.Cmd: command to execute
func (s *ServicesPanel) Update(msg tea.Msg) (*ServicesPanel, tea.Cmd) {
	// Ignore messages when not focused.
	if !s.focused {
		// Return unchanged panel.
		return s, nil
	}

	var cmd tea.Cmd

	// Handle keyboard and mouse input.
	switch msg := msg.(type) {
	// Handle keyboard messages.
	case tea.KeyMsg:
		cmd = s.handleKeyMsg(msg)
	// Handle mouse messages.
	case tea.MouseMsg:
		s.viewport, cmd = s.viewport.Update(msg)
	}

	// Return updated panel.
	return s, cmd
}

// handleKeyMsg processes keyboard input.
//
// Params:
//   - msg: key message to process
//
// Returns:
//   - tea.Cmd: command to execute
func (s *ServicesPanel) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	// Process keyboard shortcuts.
	switch msg.String() {
	// Handle home/top navigation.
	case "home", "g":
		s.viewport.GotoTop()
		// Return no command.
		return nil
	// Handle end/bottom navigation.
	case "end", "G":
		s.viewport.GotoBottom()
		// Return no command.
		return nil
	// Handle page up navigation.
	case "pgup", "ctrl+u":
		s.viewport.HalfPageUp()
		// Return no command.
		return nil
	// Handle page down navigation.
	case "pgdown", "ctrl+d":
		s.viewport.HalfPageDown()
		// Return no command.
		return nil
	// Handle line up navigation.
	case "up", "k":
		s.viewport.ScrollUp(1)
		// Return no command.
		return nil
	// Handle line down navigation.
	case "down", "j":
		s.viewport.ScrollDown(1)
		// Return no command.
		return nil
	// Handle other keys via viewport.
	default:
		var cmd tea.Cmd
		s.viewport, cmd = s.viewport.Update(msg)
		// Return viewport command.
		return cmd
	}
}

// View renders the services panel with border and vertical scrollbar.
//
// Returns:
//   - string: rendered panel
func (s *ServicesPanel) View() string {
	var sb strings.Builder

	// Set border color based on focus state.
	borderColor := s.theme.Muted
	// Use primary color when focused.
	if s.focused {
		borderColor = s.theme.Primary
	}

	// Calculate inner width excluding borders and scrollbar.
	innerWidth := s.width - borderWidth

	// Render top border with title and count.
	s.renderTopBorder(&sb, borderColor, innerWidth)

	// Render header row.
	s.renderHeaderRow(&sb, borderColor, innerWidth)

	// Render content lines with scrollbar.
	s.renderContentLines(&sb, borderColor, innerWidth)

	// Render bottom border.
	s.renderBottomBorder(&sb, borderColor, innerWidth)

	// Return complete view.
	return sb.String()
}

// renderTopBorder renders the top border with title and count indicator.
//
// Params:
//   - sb: string builder to write to
//   - borderColor: ANSI color for border
//   - innerWidth: width inside borders
func (s *ServicesPanel) renderTopBorder(sb *strings.Builder, borderColor string, innerWidth int) {
	titlePart := fmt.Sprintf("- %s%s%s ", s.theme.Header, s.title, borderColor)
	countPart := s.countIndicator()

	// Calculate dashes needed for spacing.
	titleVisLen := serviceTitlePrefixLen + len(s.title)
	countVisLen := widget.VisibleLen(countPart)
	dashCount := innerWidth - titleVisLen - countVisLen - serviceDashCountOffset

	// Ensure non-negative dash count.
	dashCount = max(dashCount, 0)

	sb.WriteString(borderColor)
	sb.WriteString("+")
	sb.WriteString(titlePart)
	sb.WriteString(strings.Repeat("-", dashCount))
	sb.WriteString(" ")
	sb.WriteString(countPart)
	sb.WriteString(borderColor)
	sb.WriteString("-+")
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")
}

// renderHeaderRow renders the column header row.
//
// Params:
//   - sb: string builder to write to
//   - borderColor: ANSI color for border
//   - innerWidth: width inside borders
func (s *ServicesPanel) renderHeaderRow(sb *strings.Builder, borderColor string, innerWidth int) {
	headerLine := s.renderHeader()
	sb.WriteString(borderColor)
	sb.WriteString("|")
	sb.WriteString(ansi.Reset)
	sb.WriteString(headerLine)
	headerVisLen := widget.VisibleLen(headerLine)

	// Pad header if needed.
	if headerVisLen < innerWidth {
		sb.WriteString(strings.Repeat(" ", innerWidth-headerVisLen))
	}
	sb.WriteString(borderColor)
	sb.WriteString(scrollTrack)
	sb.WriteString("|")
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")
}

// renderContentLines renders the content area with scrollbar.
//
// Params:
//   - sb: string builder to write to
//   - borderColor: ANSI color for border
//   - innerWidth: width inside borders
func (s *ServicesPanel) renderContentLines(sb *strings.Builder, borderColor string, innerWidth int) {
	// Get content lines from viewport.
	content := s.viewport.View()
	lines := strings.Split(content, "\n")

	// Calculate scrollbar characters.
	scrollbarChars := s.renderVerticalScrollbar()

	// Adjust content height for header row.
	contentHeight := s.viewport.Height - headerRowHeight

	// Ensure minimum content height.
	contentHeight = max(contentHeight, minContentHeight)

	// Render each content line with scrollbar.
	for i := range contentHeight {
		sb.WriteString(borderColor)
		sb.WriteString("|")
		sb.WriteString(ansi.Reset)

		// Write content line or blank space.
		if i < len(lines) {
			line := lines[i]
			visLen := widget.VisibleLen(line)
			sb.WriteString(line)

			// Pad line if needed.
			if visLen < innerWidth {
				sb.WriteString(strings.Repeat(" ", innerWidth-visLen))
			}
		} else {
			// Write blank line.
			sb.WriteString(strings.Repeat(" ", innerWidth))
		}

		// Write scrollbar character.
		sb.WriteString(borderColor)

		// Select appropriate scrollbar character.
		if i < len(scrollbarChars) {
			sb.WriteString(scrollbarChars[i])
		} else {
			// Use track character as fallback.
			sb.WriteString(scrollTrack)
		}
		sb.WriteString("|")
		sb.WriteString(ansi.Reset)
		sb.WriteString("\n")
	}
}

// renderBottomBorder renders the bottom border.
//
// Params:
//   - sb: string builder to write to
//   - borderColor: ANSI color for border
//   - innerWidth: width inside borders
func (s *ServicesPanel) renderBottomBorder(sb *strings.Builder, borderColor string, innerWidth int) {
	sb.WriteString(borderColor)
	sb.WriteString("+")
	sb.WriteString(strings.Repeat("-", innerWidth+scrollbarColWidthSvc))
	sb.WriteString("+")
	sb.WriteString(ansi.Reset)
}

// renderHeader renders the column headers.
//
// Returns:
//   - string: formatted header line
func (s *ServicesPanel) renderHeader() string {
	// Match column layout from formatServiceLine.
	var sb strings.Builder
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + "S" + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + widget.PadRight("NAME", nameColWidth) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + widget.PadRight("STATE", stateColWidth) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + widget.PadRight("HEALTH", healthColWidth) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + widget.PadLeft("UPTIME", uptimeColWidth) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + widget.PadLeft("PID", pidColWidth) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + widget.PadLeft("RST", restartsColWidth) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + widget.PadLeft("CPU", cpuColWidth) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + widget.PadLeft("MEM", memColWidth) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + "PORTS" + ansi.Reset)

	// Return formatted header.
	return sb.String()
}

// countIndicator returns the service count as [ running / total ].
//
// Returns:
//   - string: formatted count indicator
func (s *ServicesPanel) countIndicator() string {
	total := len(s.services)
	running := 0

	// Count running services.
	for _, svc := range s.services {
		// Increment for running services.
		if svc.State == process.StateRunning {
			running++
		}
	}

	// Return formatted indicator.
	return s.theme.Muted + "[ " + strconv.Itoa(running) + " / " + strconv.Itoa(total) + " ]" + ansi.Reset
}

// renderVerticalScrollbar returns the scrollbar characters for each row.
//
// Returns:
//   - []string: scrollbar characters for each row
func (s *ServicesPanel) renderVerticalScrollbar() []string {
	height := s.viewport.Height - headerRowHeight

	// Ensure minimum height.
	height = max(height, minContentHeight)
	totalLines := len(s.services)

	// No scrolling needed if content fits.
	if totalLines <= height {
		result := make([]string, height)

		// Fill with track characters.
		for i := range result {
			result[i] = scrollTrack
		}

		// Return track-only scrollbar.
		return result
	}

	// Calculate thumb size (minimum 1).
	ratio := float64(height) / float64(totalLines)
	thumbSize := int(float64(height) * ratio)

	// Ensure minimum thumb size.
	thumbSize = max(thumbSize, minThumbSizeSvc)

	// Calculate thumb position based on scroll percentage.
	scrollableHeight := height - thumbSize
	scrollPercent := s.viewport.ScrollPercent()
	thumbPos := int(float64(scrollableHeight) * scrollPercent)

	// Build scrollbar with thumb.
	result := make([]string, height)

	// Assign character to each position.
	for i := range height {
		// Use thumb character for thumb position, track elsewhere.
		if i >= thumbPos && i < thumbPos+thumbSize {
			result[i] = scrollThumb
		} else {
			// Use track character outside thumb.
			result[i] = scrollTrack
		}
	}

	// Return complete scrollbar.
	return result
}

// Height returns the panel height.
//
// Returns:
//   - int: panel height
func (s *ServicesPanel) Height() int {
	// Return current height.
	return s.height
}

// Width returns the panel width.
//
// Returns:
//   - int: panel width
func (s *ServicesPanel) Width() int {
	// Return current width.
	return s.width
}

// ScrollToTop scrolls to the top.
func (s *ServicesPanel) ScrollToTop() {
	s.viewport.GotoTop()
}

// ScrollToBottom scrolls to the bottom.
func (s *ServicesPanel) ScrollToBottom() {
	s.viewport.GotoBottom()
}

// ServiceCount returns the total number of services.
//
// Returns:
//   - int: number of services
func (s *ServicesPanel) ServiceCount() int {
	// Return service count.
	return len(s.services)
}
