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

// FocusPanel represents which panel has focus.
type FocusPanel int

const (
	// FocusServices focuses the services panel.
	FocusServices FocusPanel = iota
	// FocusLogs focuses the logs panel.
	FocusLogs
)

// Model is the Bubble Tea model.
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
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.tick(),
		tea.EnterAltScreen,
	)
}

// tick returns a command that ticks after the refresh interval.
func (m Model) tick() tea.Cmd {
	return tea.Tick(m.tui.config.RefreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "tab":
			// Toggle focus between panels.
			if m.focus == FocusServices {
				m.focus = FocusLogs
				m.servicesPanel.SetFocused(false)
				m.logsPanel.SetFocused(true)
			} else {
				m.focus = FocusServices
				m.logsPanel.SetFocused(false)
				m.servicesPanel.SetFocused(true)
			}
			return m, nil

		case "l":
			// Quick switch to logs.
			m.focus = FocusLogs
			m.servicesPanel.SetFocused(false)
			m.logsPanel.SetFocused(true)
			return m, nil

		case "s":
			// Quick switch to services.
			m.focus = FocusServices
			m.logsPanel.SetFocused(false)
			m.servicesPanel.SetFocused(true)
			return m, nil

		case "esc":
			// Return to services if in logs.
			if m.focus == FocusLogs {
				m.focus = FocusServices
				m.logsPanel.SetFocused(false)
				m.servicesPanel.SetFocused(true)
				return m, nil
			}
			// Otherwise quit.
			m.quitting = true
			return m, tea.Quit

		case "G":
			// Go to bottom of focused panel.
			switch m.focus {
			case FocusLogs:
				m.logsPanel.ScrollToBottom()
			case FocusServices:
				m.servicesPanel.ScrollToBottom()
			}
			return m, nil

		case "g":
			// Go to top of focused panel.
			switch m.focus {
			case FocusLogs:
				m.logsPanel.ScrollToTop()
			case FocusServices:
				m.servicesPanel.ScrollToTop()
			}
			return m, nil
		}

		// Forward to focused panel.
		switch m.focus {
		case FocusLogs:
			var cmd tea.Cmd
			m.logsPanel, cmd = m.logsPanel.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		case FocusServices:
			var cmd tea.Cmd
			m.servicesPanel, cmd = m.servicesPanel.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updatePanelSizes()
		return m, nil

	case tea.MouseMsg:
		// Forward mouse events to logs panel if focused.
		if m.focus == FocusLogs {
			var cmd tea.Cmd
			m.logsPanel, cmd = m.logsPanel.Update(msg)
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case tickMsg:
		// Refresh data.
		m.tui.collectData()
		// Update panels with new data.
		if m.tui.snapshot != nil {
			m.logsPanel.SetEntries(m.tui.snapshot.Logs.RecentEntries)
			m.servicesPanel.SetServices(m.tui.snapshot.Services)
		}
		cmds = append(cmds, m.tick())

	case logMsg:
		// Add new log entry.
		m.logsPanel.AddEntry(model.LogEntry(msg))
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

// updatePanelSizes recalculates panel sizes based on available space.
func (m *Model) updatePanelSizes() {
	// Layout: Header (11 lines) | Services + System/Network | Logs | Status (1 line)
	// Standard terminal: 80x24.
	headerHeight := 11
	statusHeight := 1
	systemHeight := 7 // System section (box with 5 content lines).

	// Available height for content.
	availableHeight := m.height - headerHeight - statusHeight - systemHeight

	// Services panel: adapts to number of services, max 10 visible (+3 for borders/header).
	servicesHeight := m.servicesPanel.OptimalHeight()

	// Remaining space goes to logs.
	logsHeight := availableHeight - servicesHeight
	if logsHeight < 4 {
		logsHeight = 4
	}

	m.logsPanel.SetSize(m.width, logsHeight)
	m.servicesPanel.SetSize(m.width, servicesHeight)
}

// View renders the UI.
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	snap := m.tui.snapshot
	if snap == nil {
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
	case terminal.LayoutCompact:
		sb.WriteString(m.renderCompact(snap))
	case terminal.LayoutNormal:
		sb.WriteString(m.renderNormal(snap))
	case terminal.LayoutWide, terminal.LayoutUltraWide:
		sb.WriteString(m.renderWide(snap))
	}

	// Status bar.
	sb.WriteString(m.renderStatusBar(snap))

	return sb.String()
}

// renderCompact renders for small terminals (80x24).
// Shows only: Services panel (scrollable) + Logs panel.
func (m Model) renderCompact(_ *model.Snapshot) string {
	var sb strings.Builder

	// Services panel (scrollable).
	sb.WriteString(m.servicesPanel.View())
	sb.WriteString("\n")

	// Logs panel (scrollable).
	sb.WriteString(m.logsPanel.View())

	return sb.String()
}

// renderNormal renders for normal terminals.
// Shows: System + Services panel (scrollable) + Logs panel.
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

	return sb.String()
}

// renderWide renders for wide terminals.
// Top: System + Network side by side
// Middle: Services (scrollable)
// Bottom: Logs (scrollable)
func (m Model) renderWide(snap *model.Snapshot) string {
	var sb strings.Builder

	// Calculate column width for system/network (half screen each).
	halfWidth := m.width / 2

	// Left: System, Right: Network (side by side).
	system := screen.NewSystemRenderer(halfWidth)
	network := screen.NewNetworkRenderer(halfWidth)
	systemContent := system.RenderForInteractive(snap)
	networkContent := network.Render(snap)

	// Merge side by side.
	systemLines := strings.Split(systemContent, "\n")
	networkLines := strings.Split(networkContent, "\n")

	// Remove trailing empty lines.
	for len(systemLines) > 0 && strings.TrimSpace(systemLines[len(systemLines)-1]) == "" {
		systemLines = systemLines[:len(systemLines)-1]
	}
	for len(networkLines) > 0 && strings.TrimSpace(networkLines[len(networkLines)-1]) == "" {
		networkLines = networkLines[:len(networkLines)-1]
	}

	maxLines := len(systemLines)
	if len(networkLines) > maxLines {
		maxLines = len(networkLines)
	}

	for i := range maxLines {
		left := ""
		right := ""
		if i < len(systemLines) {
			left = systemLines[i]
		}
		if i < len(networkLines) {
			right = networkLines[i]
		}

		// Pad left to halfWidth.
		leftVisible := widget.VisibleLen(left)
		if leftVisible < halfWidth {
			left += strings.Repeat(" ", halfWidth-leftVisible)
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

	return sb.String()
}

// renderStatusBar renders the bottom status bar.
func (m Model) renderStatusBar(snap *model.Snapshot) string {
	// Focus indicator.
	var focusIndicator string
	if m.focus == FocusLogs {
		focusIndicator = m.theme.Primary + "[LOGS]" + ansi.Reset
	} else {
		focusIndicator = m.theme.Primary + "[SERVICES]" + ansi.Reset
	}

	// Keybindings based on focus.
	var keys string
	if m.focus == FocusLogs {
		keys = m.theme.Muted + "[↑↓] Scroll  [g/G] Top/Bottom  [s] Services  [Tab] Switch  [q] Quit" + ansi.Reset
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
	padding := m.width - contentLen - badgeLen - 2

	if padding < 0 {
		padding = 0
	}

	return statusContent + strings.Repeat(" ", padding) + badge + "  "
}

// runBubbleTea starts the Bubble Tea program.
func (t *TUI) runBubbleTea(ctx context.Context) error {
	// Initial data collection.
	t.collectData()

	// Get initial size.
	size := terminal.GetSize()

	// Create services panel first to calculate optimal height.
	servicesPanel := component.NewServicesPanel(size.Cols, 6) // Temporary height.
	if t.snapshot != nil {
		servicesPanel.SetServices(t.snapshot.Services)
	}

	// Calculate initial panel sizes dynamically.
	headerHeight := 11
	statusHeight := 1
	systemHeight := 7
	availableHeight := size.Rows - headerHeight - statusHeight - systemHeight

	// Services panel adapts to number of services.
	servicesHeight := servicesPanel.OptimalHeight()

	// Remaining space goes to logs.
	logsHeight := availableHeight - servicesHeight
	if logsHeight < 4 {
		logsHeight = 4
	}

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

	p := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run in goroutine to support context cancellation.
	done := make(chan error, 1)
	go func() {
		_, err := p.Run()
		done <- err
	}()

	// Wait for context or program completion.
	select {
	case <-ctx.Done():
		p.Quit()
		return ctx.Err()
	case err := <-done:
		return err
	}
}
