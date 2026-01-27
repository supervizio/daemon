// Package tui provides terminal user interface for superviz.io.
package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/component"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/screen"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

const (
	// Layout dimensions for terminal sections.
	layoutHeaderHeight       int = 11 // Header section height in lines.
	layoutStatusBarHeight    int = 1  // Status bar height in lines.
	layoutSystemSectionLines int = 7  // System section height (box with content).
	layoutMinLogHeight       int = 4  // Minimum log panel height.

	// Panel sizing constants.
	panelInitialServicesHeight int = 6 // Initial temporary height for services panel.
	panelHalfWidthDivisor      int = 2 // Divisor for half-width panels.

	// Status bar spacing.
	statusBarBadgeSpacing int = 2 // Badge spacing for status bar.
)

// FocusPanel represents which panel has focus.
type FocusPanel int

const (
	// FocusServices focuses the services panel.
	FocusServices FocusPanel = iota
	// FocusLogs focuses the logs panel.
	FocusLogs
)

// Model is the Bubble Tea model.
// TODO: Add missing tests - create interactive_external_test.go and interactive_internal_test.go.
type Model struct {
	tui           *TUI
	width         int
	height        int
	quitting      bool
	focus         FocusPanel
	logsPanel     component.LogsPanel
	servicesPanel component.ServicesPanel
	theme         ansi.Theme
}

// tickMsg is sent on each refresh interval.
type tickMsg time.Time

// logMsg is sent when a new log entry arrives.
type logMsg model.LogEntry

// Init initializes the model.
//
// Returns:
//   - tea.Cmd: batch of initialization commands.
func (m Model) Init() tea.Cmd {
	// Initialize with tick command and enter alt screen.
	return tea.Batch(
		m.tick(),
		tea.EnterAltScreen,
	)
}

// tick returns a command that ticks after the refresh interval.
//
// Returns:
//   - tea.Cmd: tick command.
func (m Model) tick() tea.Cmd {
	// Schedule next refresh tick.
	return tea.Tick(m.tui.config.RefreshInterval, func(t time.Time) tea.Msg {
		// Convert time to tick message.
		return tickMsg(t)
	})
}

// Update handles messages.
//
// Params:
//   - msg: message to process.
//
// Returns:
//   - tea.Model: updated model.
//   - tea.Cmd: command to execute.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Process message by type.
	switch msg := msg.(type) {
	// Handle keyboard input.
	case tea.KeyMsg:
		// Process keyboard shortcuts.
		switch msg.String() {
		// Quit shortcuts.
		case "q", "ctrl+c":
			m.quitting = true
			// Exit immediately.
			return m, tea.Quit

		// Toggle focus between panels.
		case "tab":
			// Toggle focus between panels.
			if m.focus == FocusServices {
				m.focus = FocusLogs
				m.servicesPanel.SetFocused(false)
				m.logsPanel.SetFocused(true)
				// Switch to services.
			} else {
				m.focus = FocusServices
				m.logsPanel.SetFocused(false)
				m.servicesPanel.SetFocused(true)
			}
			// Return with no command.
			return m, nil

		// Quick switch to logs.
		case "l":
			// Quick switch to logs.
			m.focus = FocusLogs
			m.servicesPanel.SetFocused(false)
			m.logsPanel.SetFocused(true)
			// Return with no command.
			return m, nil

		// Quick switch to services.
		case "s":
			// Quick switch to services.
			m.focus = FocusServices
			m.logsPanel.SetFocused(false)
			m.servicesPanel.SetFocused(true)
			// Return with no command.
			return m, nil

		// Handle escape key (context-dependent).
		case "esc":
			// Return to services if in logs.
			if m.focus == FocusLogs {
				m.focus = FocusServices
				m.logsPanel.SetFocused(false)
				m.servicesPanel.SetFocused(true)
				// Return with no command.
				return m, nil
			}
			// Otherwise quit.
			m.quitting = true
			// Exit immediately.
			return m, tea.Quit

		// Jump to bottom (Vim-style).
		case "G":
			// Go to bottom of focused panel.
			switch m.focus {
			// Scroll logs to bottom.
			case FocusLogs:
				m.logsPanel.ScrollToBottom()
			// Scroll services to bottom.
			case FocusServices:
				m.servicesPanel.ScrollToBottom()
			}
			// Return with no command.
			return m, nil

		// Jump to top (Vim-style).
		case "g":
			// Go to top of focused panel.
			switch m.focus {
			// Scroll logs to top.
			case FocusLogs:
				m.logsPanel.ScrollToTop()
			// Scroll services to top.
			case FocusServices:
				m.servicesPanel.ScrollToTop()
			}
			// Return with no command.
			return m, nil
		}

		// Forward to focused panel.
		switch m.focus {
		// Forward key to logs panel.
		case FocusLogs:
			lp, cmd := m.logsPanel.Update(msg)
			m.logsPanel = *lp
			// Append command if not nil.
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		// Forward key to services panel.
		case FocusServices:
			sp, cmd := m.servicesPanel.Update(msg)
			m.servicesPanel = *sp
			// Append command if not nil.
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	// Handle window resize.
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePanelSizes()
		// Return with no command.
		return m, nil

	// Handle mouse events.
	case tea.MouseMsg:
		// Forward mouse events to logs panel if focused.
		if m.focus == FocusLogs {
			lp, cmd := m.logsPanel.Update(msg)
			m.logsPanel = *lp
			// Append command if not nil.
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	// Handle periodic refresh tick.
	case tickMsg:
		// Refresh data.
		m.tui.collectData()
		// Update panels with new data.
		if m.tui.snapshot != nil {
			m.logsPanel.SetEntries(m.tui.snapshot.Logs.RecentEntries)
			m.servicesPanel.SetServices(m.tui.snapshot.Services)
		}
		cmds = append(cmds, m.tick())

	// Handle new log entry.
	case logMsg:
		// Add new log entry.
		m.logsPanel.AddEntry(model.LogEntry(msg))
	}

	// Execute all batched commands.
	if len(cmds) > 0 {
		// Return with batched commands.
		return m, tea.Batch(cmds...)
	}
	// Return with no commands.
	return m, nil
}

// updatePanelSizes recalculates panel sizes based on available space.
func (m *Model) updatePanelSizes() {
	// Layout: Header (11 lines) | Services + System/Network | Logs | Status (1 line)
	// Standard terminal: 80x24.
	headerHeight := layoutHeaderHeight
	statusHeight := layoutStatusBarHeight
	systemHeight := layoutSystemSectionLines // System section (box with 5 content lines).

	// Available height for content.
	availableHeight := m.height - headerHeight - statusHeight - systemHeight

	// Services panel: adapts to number of services, max 10 visible (+3 for borders/header).
	servicesHeight := m.servicesPanel.OptimalHeight()

	// Remaining space goes to logs, enforcing minimum.
	logsHeight := max(availableHeight-servicesHeight, layoutMinLogHeight)

	m.logsPanel.SetSize(m.width, logsHeight)
	m.servicesPanel.SetSize(m.width, servicesHeight)
}

// View renders the UI.
//
// Returns:
//   - string: rendered UI.
func (m Model) View() string {
	// Return empty view if quitting.
	if m.quitting {
		// Return empty string when quitting.
		return ""
	}

	snap := m.tui.snapshot
	// Show loading message if no snapshot.
	if snap == nil {
		// Return loading message.
		return "Loading..."
	}

	var sb strings.Builder

	// Clear screen.
	sb.WriteString(ansi.ClearScreen)
	sb.WriteString(ansi.CursorHome)

	// Determine layout.
	size := terminal.Size{Cols: m.width, Rows: m.height}
	layout := terminal.GetLayout(size)

	// Header.
	header := screen.NewHeaderRenderer(m.width)
	sb.WriteString(header.Render(snap, true))
	sb.WriteString("\n")

	// Content based on layout.
	switch layout {
	// Compact layout for small terminals.
	case terminal.LayoutCompact:
		sb.WriteString(m.renderCompact(snap))
	// Normal layout for standard terminals.
	case terminal.LayoutNormal:
		sb.WriteString(m.renderNormal(snap))
	// Wide/ultra-wide layout for large terminals.
	case terminal.LayoutWide, terminal.LayoutUltraWide:
		sb.WriteString(m.renderWide(snap))
	}

	// Status bar.
	sb.WriteString(m.renderStatusBar(snap))

	// Return rendered UI.
	return sb.String()
}

// renderCompact renders for small terminals (80x24).
// Shows only: Services panel (scrollable) + Logs panel.
//
// Params:
//   - _: snapshot (unused for compact layout).
//
// Returns:
//   - string: rendered compact layout.
func (m Model) renderCompact(_ *model.Snapshot) string {
	var sb strings.Builder

	// Services panel (scrollable).
	sb.WriteString(m.servicesPanel.View())
	sb.WriteString("\n")

	// Logs panel (scrollable).
	sb.WriteString(m.logsPanel.View())

	// Return rendered compact layout.
	return sb.String()
}

// renderNormal renders for normal terminals.
// Shows: System + Services panel (scrollable) + Logs panel.
//
// Params:
//   - snap: current snapshot.
//
// Returns:
//   - string: rendered normal layout.
func (m Model) renderNormal(snap *model.Snapshot) string {
	var sb strings.Builder

	// System section with progress bars.
	system := screen.NewSystemRenderer(m.width)
	sb.WriteString(system.RenderForInteractive(snap))
	sb.WriteString("\n")

	// Services panel (scrollable).
	sb.WriteString(m.servicesPanel.View())
	sb.WriteString("\n")

	// Logs panel (scrollable).
	sb.WriteString(m.logsPanel.View())

	// Return rendered normal layout.
	return sb.String()
}

// renderWide renders for wide terminals.
// Top: System + Network side by side
// Middle: Services (scrollable)
// Bottom: Logs (scrollable)
//
// Params:
//   - snap: current snapshot.
//
// Returns:
//   - string: rendered wide layout.
func (m Model) renderWide(snap *model.Snapshot) string {
	var sb strings.Builder

	// Calculate column width for system/network (half screen each).
	halfWidth := m.width / panelHalfWidthDivisor

	// Left: System, Right: Network (side by side).
	system := screen.NewSystemRenderer(halfWidth)
	network := screen.NewNetworkRenderer(halfWidth)
	systemContent := system.RenderForInteractive(snap)
	networkContent := network.Render(snap)

	// Merge side by side.
	systemLines := strings.Split(systemContent, "\n")
	networkLines := strings.Split(networkContent, "\n")

	// Remove trailing empty lines from system.
	for len(systemLines) > 0 && strings.TrimSpace(systemLines[len(systemLines)-1]) == "" {
		systemLines = systemLines[:len(systemLines)-1]
	}
	// Remove trailing empty lines from network.
	for len(networkLines) > 0 && strings.TrimSpace(networkLines[len(networkLines)-1]) == "" {
		networkLines = networkLines[:len(networkLines)-1]
	}

	// Find maximum number of lines between panels.
	maxLines := max(len(systemLines), len(networkLines))

	// Combine lines side by side.
	for i := range maxLines {
		left := ""
		right := ""
		// Get left line if available.
		if i < len(systemLines) {
			left = systemLines[i]
		}
		// Get right line if available.
		if i < len(networkLines) {
			right = networkLines[i]
		}

		leftVisible := widget.VisibleLen(left)
		// Pad left to halfWidth using builder if needed.
		if leftVisible < halfWidth {
			var padded strings.Builder
			padded.WriteString(left)
			padded.WriteString(strings.Repeat(" ", halfWidth-leftVisible))
			left = padded.String()
		}

		sb.WriteString(left)
		sb.WriteString(right)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Services panel (full width, scrollable).
	sb.WriteString(m.servicesPanel.View())
	sb.WriteString("\n")

	// Logs panel (full width, scrollable).
	sb.WriteString(m.logsPanel.View())

	// Return rendered wide layout.
	return sb.String()
}

// renderStatusBar renders the bottom status bar.
//
// Params:
//   - snap: current snapshot.
//
// Returns:
//   - string: rendered status bar.
func (m Model) renderStatusBar(snap *model.Snapshot) string {
	// Focus indicator.
	var focusIndicator string
	// Show logs focus indicator.
	if m.focus == FocusLogs {
		focusIndicator = m.theme.Primary + "[LOGS]" + ansi.Reset
		// Show services focus indicator.
	} else {
		focusIndicator = m.theme.Primary + "[SERVICES]" + ansi.Reset
	}

	// Keybindings based on focus.
	var keys string
	// Show logs keybindings.
	if m.focus == FocusLogs {
		keys = m.theme.Muted + "[↑↓] Scroll  [g/G] Top/Bottom  [s] Services  [Tab] Switch  [q] Quit" + ansi.Reset
		// Show services keybindings.
	} else {
		keys = m.theme.Muted + "[↑↓] Scroll  [g/G] Top/Bottom  [l] Logs  [Tab] Switch  [q] Quit" + ansi.Reset
	}

	// Error badge.
	logs := screen.NewLogsRenderer(m.width)
	badge := logs.RenderBadge(snap)

	// Combine.
	statusContent := "  " + focusIndicator + "  " + keys
	contentLen := widget.VisibleLen(statusContent)
	badgeLen := widget.VisibleLen(badge)
	// Ensure non-negative padding for badge placement.
	padding := max(0, m.width-contentLen-badgeLen-statusBarBadgeSpacing)

	// Return status bar with badge.
	return statusContent + strings.Repeat(" ", padding) + badge + "  "
}

// runBubbleTea starts the Bubble Tea program.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - error: any error during execution.
func (t *TUI) runBubbleTea(ctx context.Context) error {
	// Initial data collection.
	t.collectData()

	// Get initial size.
	size := terminal.GetSize()

	// Create services panel first to calculate optimal height.
	servicesPanel := component.NewServicesPanel(size.Cols, panelInitialServicesHeight) // Temporary height.
	if t.snapshot != nil {
		servicesPanel.SetServices(t.snapshot.Services)
	}

	// Calculate initial panel sizes dynamically.
	headerHeight := layoutHeaderHeight
	statusHeight := layoutStatusBarHeight
	systemHeight := layoutSystemSectionLines
	availableHeight := size.Rows - headerHeight - statusHeight - systemHeight

	// Services panel adapts to number of services.
	servicesHeight := servicesPanel.OptimalHeight()

	// Remaining space goes to logs, enforcing minimum.
	logsHeight := max(availableHeight-servicesHeight, layoutMinLogHeight)

	// Update services panel with correct height.
	servicesPanel.SetSize(size.Cols, servicesHeight)

	m := Model{
		tui:           t,
		width:         size.Cols,
		height:        size.Rows,
		theme:         ansi.DefaultTheme(),
		focus:         FocusServices, // Start with services focused.
		logsPanel:     component.NewLogsPanel(size.Cols, logsHeight),
		servicesPanel: servicesPanel,
	}

	// Services panel starts focused.
	m.servicesPanel.SetFocused(true)

	// Initialize panels with current data.
	if t.snapshot != nil {
		m.logsPanel.SetEntries(t.snapshot.Logs.RecentEntries)
	}

	prg := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run in goroutine to support context cancellation.
	done := make(chan error, 1)
	go func() {
		_, err := prg.Run()
		done <- err
	}()

	// Wait for context or program completion.
	select {
	case <-ctx.Done():
		prg.Quit()
		// Return context error.
		return ctx.Err()
	case err := <-done:
		// Return program error.
		return err
	}
}
