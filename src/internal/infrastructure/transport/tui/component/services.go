// Package component provides reusable Bubble Tea components.
package component

import (
	"fmt"
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
const maxServicesVisible = 10

// ServicesPanel is a scrollable services list with vertical scrollbar.
type ServicesPanel struct {
	viewport viewport.Model
	theme    ansi.Theme
	width    int
	height   int
	services []model.ServiceSnapshot
	focused  bool
	title    string
}

// NewServicesPanel creates a new services panel.
func NewServicesPanel(width, height int) ServicesPanel {
	// -3 for borders (left border, right border, scrollbar).
	vp := viewport.New(width-3, height-2)

	return ServicesPanel{
		viewport: vp,
		theme:    ansi.DefaultTheme(),
		width:    width,
		height:   height,
		services: make([]model.ServiceSnapshot, 0),
		title:    "Services",
	}
}

// SetSize updates the panel dimensions.
func (s *ServicesPanel) SetSize(width, height int) {
	s.width = width
	s.height = height
	s.viewport.Width = width - 3   // -3 for left border, right border, scrollbar
	s.viewport.Height = height - 2 // -2 for top/bottom borders
	s.updateContent()
}

// SetFocused sets the focus state.
func (s *ServicesPanel) SetFocused(focused bool) {
	s.focused = focused
}

// IsFocused returns whether the panel is focused.
func (s ServicesPanel) IsFocused() bool {
	return s.focused
}

// SetServices updates the service list.
func (s *ServicesPanel) SetServices(services []model.ServiceSnapshot) {
	s.services = services
	s.updateContent()
}

// OptimalHeight returns the optimal height for the panel based on service count.
// Returns height including borders (2 lines) + header (1 line) + content lines.
// Content is capped at maxServicesVisible.
func (s ServicesPanel) OptimalHeight() int {
	contentLines := len(s.services)
	if contentLines > maxServicesVisible {
		contentLines = maxServicesVisible
	}
	if contentLines < 1 {
		contentLines = 1 // At least 1 line for "no services" message.
	}
	// +2 for top/bottom borders, +1 for header row.
	return contentLines + 3
}

// updateContent rebuilds the viewport content.
func (s *ServicesPanel) updateContent() {
	var sb strings.Builder

	if len(s.services) == 0 {
		sb.WriteString(s.theme.Muted + " No services configured" + ansi.Reset + "\n")
		s.viewport.SetContent(sb.String())
		return
	}

	for _, svc := range s.services {
		// State icon (2 chars visible: icon + space).
		stateIcon := s.getStateIcon(svc.State)

		// Name (truncate if needed, pad to 12).
		name := svc.Name
		if len(name) > 12 {
			name = name[:9] + "..."
		}

		// State text (pad to 9).
		stateText := s.getStateText(svc.State)

		// Health text (pad to 9).
		healthText := s.getHealthText(svc)

		// Uptime (right-align 7).
		uptime := "-"
		if svc.State == process.StateRunning && svc.Uptime > 0 {
			uptime = widget.FormatDurationShort(svc.Uptime)
		}

		// PID (right-align 6).
		pid := "-"
		if svc.PID > 0 {
			pid = fmt.Sprintf("%d", svc.PID)
		}

		// Restarts (right-align 3).
		restarts := "-"
		if svc.RestartCount > 0 {
			restarts = fmt.Sprintf("%d", svc.RestartCount)
		}

		// CPU% (right-align 5).
		cpu := "-"
		if svc.State == process.StateRunning {
			cpu = fmt.Sprintf("%.1f%%", svc.CPUPercent)
		}

		// Memory (right-align 8).
		mem := "-"
		if svc.State == process.StateRunning && svc.MemoryRSS > 0 {
			mem = widget.FormatBytes(svc.MemoryRSS)
		}

		// Ports with colors based on status (at the end).
		ports := s.formatPortsWithStatus(svc)

		// Build line with manual padding for ANSI-colored strings.
		// Column widths: icon(2) name(12) state(9) health(9) uptime(7) pid(6) rst(3) cpu(5) mem(8) ports(15)
		sb.WriteString(" ")
		sb.WriteString(stateIcon)
		sb.WriteString(" ")
		sb.WriteString(padRight(name, 12))
		sb.WriteString(" ")
		sb.WriteString(padRightAnsi(stateText, 9))
		sb.WriteString(" ")
		sb.WriteString(padRightAnsi(healthText, 9))
		sb.WriteString(" ")
		sb.WriteString(padLeft(uptime, 7))
		sb.WriteString(" ")
		sb.WriteString(padLeft(pid, 6))
		sb.WriteString(" ")
		sb.WriteString(padLeft(restarts, 3))
		sb.WriteString(" ")
		sb.WriteString(padLeft(cpu, 5))
		sb.WriteString(" ")
		sb.WriteString(padLeft(mem, 8))
		sb.WriteString(" ")
		sb.WriteString(ports) // Ports at the end, left-aligned with colors.
		sb.WriteString("\n")
	}

	s.viewport.SetContent(sb.String())
}

// padRight pads a string to the right with spaces.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// padRightAnsi pads an ANSI-colored string to the right based on visible length.
func padRightAnsi(s string, width int) string {
	visLen := visibleLen(s)
	if visLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visLen)
}

// padLeft pads a string to the left with spaces.
func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// formatPorts formats a list of ports as a comma-separated string.
// Truncates if too long.
func formatPorts(ports []int) string {
	if len(ports) == 0 {
		return "-"
	}

	var sb strings.Builder
	for i, port := range ports {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%d", port))
		// Stop if getting too long (max 11 chars to fit in 12 with truncation).
		if sb.Len() > 9 && i < len(ports)-1 {
			sb.WriteString("...")
			break
		}
	}
	return sb.String()
}

// formatPortsWithStatus formats ports with colors based on listener status.
// Colors: Green (OK), Yellow (Warning), Red (Error).
func (s *ServicesPanel) formatPortsWithStatus(svc model.ServiceSnapshot) string {
	// If no listeners configured, show detected ports in muted color.
	if len(svc.Listeners) == 0 {
		if len(svc.Ports) == 0 {
			return s.theme.Muted + "-" + ansi.Reset
		}
		// Show detected ports in muted (not configured but listening).
		return s.theme.Muted + formatPorts(svc.Ports) + ansi.Reset
	}

	var sb strings.Builder
	for i, l := range svc.Listeners {
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
		default:
			color = s.theme.Muted
		}

		// Format: :PORT or PORT (with color).
		sb.WriteString(color)
		sb.WriteString(fmt.Sprintf(":%d", l.Port))
		sb.WriteString(ansi.Reset)
	}

	return sb.String()
}

// getStateIcon returns the state icon with color.
func (s *ServicesPanel) getStateIcon(state process.State) string {
	switch state {
	case process.StateRunning:
		return s.theme.Success + "●" + ansi.Reset
	case process.StateStopped:
		return s.theme.Muted + "○" + ansi.Reset
	case process.StateFailed:
		return s.theme.Error + "●" + ansi.Reset
	case process.StateStarting:
		return s.theme.Warning + "◐" + ansi.Reset
	case process.StateStopping:
		return s.theme.Warning + "◑" + ansi.Reset
	default:
		return s.theme.Muted + "○" + ansi.Reset
	}
}

// getStateText returns the state text with color.
func (s *ServicesPanel) getStateText(state process.State) string {
	switch state {
	case process.StateRunning:
		return s.theme.Success + "running" + ansi.Reset
	case process.StateStopped:
		return s.theme.Muted + "stopped" + ansi.Reset
	case process.StateFailed:
		return s.theme.Error + "failed" + ansi.Reset
	case process.StateStarting:
		return s.theme.Warning + "starting" + ansi.Reset
	case process.StateStopping:
		return s.theme.Warning + "stopping" + ansi.Reset
	default:
		return s.theme.Muted + "unknown" + ansi.Reset
	}
}

// getHealthText returns the health status text with color.
func (s *ServicesPanel) getHealthText(svc model.ServiceSnapshot) string {
	// Only show health for running services with health checks configured.
	if svc.State != process.StateRunning {
		return s.theme.Muted + "-" + ansi.Reset
	}

	// No health checks configured - show dash.
	if !svc.HasHealthChecks {
		return s.theme.Muted + "-" + ansi.Reset
	}

	switch svc.Health {
	case health.StatusHealthy:
		return s.theme.Success + "healthy" + ansi.Reset
	case health.StatusUnhealthy:
		return s.theme.Error + "unhealthy" + ansi.Reset
	case health.StatusDegraded:
		return s.theme.Warning + "degraded" + ansi.Reset
	case health.StatusUnknown:
		return s.theme.Warning + "pending" + ansi.Reset
	default:
		return s.theme.Muted + "-" + ansi.Reset
	}
}

// Init initializes the component.
func (s ServicesPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (s ServicesPanel) Update(msg tea.Msg) (ServicesPanel, tea.Cmd) {
	if !s.focused {
		return s, nil
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "home", "g":
			s.viewport.GotoTop()
		case "end", "G":
			s.viewport.GotoBottom()
		case "pgup", "ctrl+u":
			s.viewport.HalfPageUp()
		case "pgdown", "ctrl+d":
			s.viewport.HalfPageDown()
		case "up", "k":
			s.viewport.ScrollUp(1)
		case "down", "j":
			s.viewport.ScrollDown(1)
		default:
			s.viewport, cmd = s.viewport.Update(msg)
		}
	case tea.MouseMsg:
		s.viewport, cmd = s.viewport.Update(msg)
	}

	return s, cmd
}

// View renders the services panel with border and vertical scrollbar.
func (s ServicesPanel) View() string {
	var sb strings.Builder

	// Border color based on focus.
	borderColor := s.theme.Muted
	if s.focused {
		borderColor = s.theme.Primary
	}

	// Inner width = total width - 2 borders - 1 scrollbar.
	innerWidth := s.width - 3

	// === Top border with title and count ===
	titlePart := fmt.Sprintf("─ %s%s%s ", s.theme.Header, s.title, borderColor)
	countPart := s.countIndicator()

	// Calculate dashes needed.
	titleVisLen := 3 + len(s.title) // "─ " + title + " "
	countVisLen := visibleLen(countPart)
	dashCount := innerWidth - titleVisLen - countVisLen - 1 // -1 for final "─"
	if dashCount < 0 {
		dashCount = 0
	}

	sb.WriteString(borderColor)
	sb.WriteString("╭")
	sb.WriteString(titlePart)
	sb.WriteString(strings.Repeat("─", dashCount))
	sb.WriteString(" ")
	sb.WriteString(countPart)
	sb.WriteString(borderColor) // Re-apply after countPart reset.
	sb.WriteString("─╮")
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")

	// === Header row ===
	headerLine := s.renderHeader(innerWidth)
	sb.WriteString(borderColor)
	sb.WriteString("│")
	sb.WriteString(ansi.Reset)
	sb.WriteString(headerLine)
	headerVisLen := visibleLen(headerLine)
	if headerVisLen < innerWidth {
		sb.WriteString(strings.Repeat(" ", innerWidth-headerVisLen))
	}
	sb.WriteString(borderColor)
	sb.WriteString(scrollTrack)
	sb.WriteString("│")
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")

	// === Content lines with vertical scrollbar ===
	content := s.viewport.View()
	lines := strings.Split(content, "\n")

	// Calculate scrollbar.
	scrollbarChars := s.renderVerticalScrollbar()

	// -1 for header row.
	contentHeight := s.viewport.Height - 1
	if contentHeight < 1 {
		contentHeight = 1
	}

	for i := 0; i < contentHeight; i++ {
		sb.WriteString(borderColor)
		sb.WriteString("│")
		sb.WriteString(ansi.Reset)

		// Content.
		if i < len(lines) {
			line := lines[i]
			visLen := visibleLen(line)
			sb.WriteString(line)
			if visLen < innerWidth {
				sb.WriteString(strings.Repeat(" ", innerWidth-visLen))
			}
		} else {
			sb.WriteString(strings.Repeat(" ", innerWidth))
		}

		// Scrollbar character.
		sb.WriteString(borderColor)
		if i < len(scrollbarChars) {
			sb.WriteString(scrollbarChars[i])
		} else {
			sb.WriteString(scrollTrack)
		}
		sb.WriteString("│")
		sb.WriteString(ansi.Reset)
		sb.WriteString("\n")
	}

	// === Bottom border ===
	sb.WriteString(borderColor)
	sb.WriteString("╰")
	sb.WriteString(strings.Repeat("─", innerWidth+1)) // +1 for scrollbar column
	sb.WriteString("╯")
	sb.WriteString(ansi.Reset)

	return sb.String()
}

// renderHeader renders the column headers.
func (s ServicesPanel) renderHeader(_ int) string {
	// Match column layout from updateContent.
	// Column widths: icon(2) name(12) state(9) health(9) uptime(7) pid(6) rst(3) cpu(5) mem(8) ports(at end)
	var sb strings.Builder
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + "S" + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + padRight("NAME", 12) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + padRight("STATE", 9) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + padRight("HEALTH", 9) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + padLeft("UPTIME", 7) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + padLeft("PID", 6) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + padLeft("RST", 3) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + padLeft("CPU", 5) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + padLeft("MEM", 8) + ansi.Reset)
	sb.WriteString(" ")
	sb.WriteString(s.theme.Muted + "PORTS" + ansi.Reset)

	return sb.String()
}

// countIndicator returns the service count as [ running / total ].
func (s ServicesPanel) countIndicator() string {
	total := len(s.services)
	running := 0
	for _, svc := range s.services {
		if svc.State == process.StateRunning {
			running++
		}
	}

	return fmt.Sprintf("%s[ %d / %d ]%s", s.theme.Muted, running, total, ansi.Reset)
}

// renderVerticalScrollbar returns the scrollbar characters for each row.
func (s ServicesPanel) renderVerticalScrollbar() []string {
	height := s.viewport.Height - 1 // -1 for header
	if height < 1 {
		height = 1
	}
	totalLines := len(s.services)

	if totalLines <= height {
		// No scrolling needed - no thumb.
		result := make([]string, height)
		for i := range result {
			result[i] = scrollTrack
		}
		return result
	}

	// Calculate thumb size (minimum 1).
	ratio := float64(height) / float64(totalLines)
	thumbSize := int(float64(height) * ratio)
	if thumbSize < 1 {
		thumbSize = 1
	}

	// Calculate thumb position.
	scrollableHeight := height - thumbSize
	scrollPercent := s.viewport.ScrollPercent()
	thumbPos := int(float64(scrollableHeight) * scrollPercent)

	// Build scrollbar.
	result := make([]string, height)
	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbSize {
			result[i] = scrollThumb
		} else {
			result[i] = scrollTrack
		}
	}
	return result
}

// Height returns the panel height.
func (s ServicesPanel) Height() int {
	return s.height
}

// Width returns the panel width.
func (s ServicesPanel) Width() int {
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
func (s ServicesPanel) ServiceCount() int {
	return len(s.services)
}
