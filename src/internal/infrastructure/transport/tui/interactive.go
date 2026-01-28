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

// Stringer is the minimal interface for keyboard messages.
// Satisfied by tea.KeyMsg; concrete type preserved through interface assignment
// so Bubble Tea type switches in panel Update methods still match tea.KeyMsg.
type Stringer interface {
	String() string
}

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

// NewModel creates a new Model with the given configuration.
// Note: Init, Update, View use value receivers as required by tea.Model interface.
//
// Params:
//   - cfg: model configuration containing TUI, dimensions, theme, and panels.
//
// Returns:
//   - Model: configured model instance.
func NewModel(cfg ModelConfig) Model {
	return Model{
		tui:           cfg.TUI,
		width:         cfg.Width,
		height:        cfg.Height,
		theme:         cfg.Theme,
		focus:         FocusServices,
		logsPanel:     cfg.LogsPanel,
		servicesPanel: cfg.ServicesPanel,
	}
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
	return tea.Tick(m.tui.config.RefreshInterval, func(t time.Time) tea.Msg {
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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.updatePanelSizes()
		return m, nil

	case tea.MouseMsg:
		return m.handleMouseMsg(msg)

	case tickMsg:
		return m.handleTickMsg()

	case logMsg:
		m.logsPanel.AddEntry(model.LogEntry(msg))
	}

	return m, nil
}

// handleKeyMsg handles keyboard input messages.
//
// Params:
//   - msg: key message to process.
//
// Returns:
//   - tea.Model: updated model.
//   - tea.Cmd: command to execute.
func (m Model) handleKeyMsg(msg Stringer) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "tab":
		return m.toggleFocus(), nil
	case "l":
		return m.focusLogs(), nil
	case "s":
		return m.focusServices(), nil
	case "esc":
		return m.handleEscKey()
	case "G":
		return m.scrollToBottom(), nil
	case "g":
		return m.scrollToTop(), nil
	}
	return m.forwardKeyToPanel(msg)
}

// handleEscKey handles the escape key press.
//
// Returns:
//   - tea.Model: updated model.
//   - tea.Cmd: command to execute.
func (m Model) handleEscKey() (tea.Model, tea.Cmd) {
	if m.focus == FocusLogs {
		return m.focusServices(), nil
	}
	m.quitting = true
	return m, tea.Quit
}

// toggleFocus switches focus between panels.
//
// Returns:
//   - Model: updated model with toggled focus.
func (m Model) toggleFocus() Model {
	if m.focus == FocusServices {
		m.focus = FocusLogs
		m.servicesPanel.SetFocused(false)
		m.logsPanel.SetFocused(true)
	} else {
		m.focus = FocusServices
		m.logsPanel.SetFocused(false)
		m.servicesPanel.SetFocused(true)
	}
	return m
}

// focusLogs switches focus to logs panel.
//
// Returns:
//   - Model: updated model with logs focused.
func (m Model) focusLogs() Model {
	m.focus = FocusLogs
	m.servicesPanel.SetFocused(false)
	m.logsPanel.SetFocused(true)
	return m
}

// focusServices switches focus to services panel.
//
// Returns:
//   - Model: updated model with services focused.
func (m Model) focusServices() Model {
	m.focus = FocusServices
	m.logsPanel.SetFocused(false)
	m.servicesPanel.SetFocused(true)
	return m
}

// scrollToBottom scrolls focused panel to bottom.
//
// Returns:
//   - Model: updated model.
func (m Model) scrollToBottom() Model {
	switch m.focus {
	case FocusLogs:
		m.logsPanel.ScrollToBottom()
	case FocusServices:
		m.servicesPanel.ScrollToBottom()
	}
	return m
}

// scrollToTop scrolls focused panel to top.
//
// Returns:
//   - Model: updated model.
func (m Model) scrollToTop() Model {
	switch m.focus {
	case FocusLogs:
		m.logsPanel.ScrollToTop()
	case FocusServices:
		m.servicesPanel.ScrollToTop()
	}
	return m
}

// forwardKeyToPanel forwards key to focused panel.
//
// Params:
//   - msg: key message to forward.
//
// Returns:
//   - tea.Model: updated model.
//   - tea.Cmd: command from panel.
func (m Model) forwardKeyToPanel(msg Stringer) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch m.focus {
	case FocusLogs:
		lp, c := m.logsPanel.Update(msg)
		m.logsPanel = *lp
		cmd = c
	case FocusServices:
		sp, c := m.servicesPanel.Update(msg)
		m.servicesPanel = *sp
		cmd = c
	}

	return m, cmd
}

// handleMouseMsg handles mouse input messages.
//
// Params:
//   - msg: mouse message to process.
//
// Returns:
//   - tea.Model: updated model.
//   - tea.Cmd: command to execute.
func (m Model) handleMouseMsg(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.focus == FocusLogs {
		lp, cmd := m.logsPanel.Update(msg)
		m.logsPanel = *lp
		return m, cmd
	}
	return m, nil
}

// handleTickMsg handles periodic refresh tick.
//
// Returns:
//   - tea.Model: updated model.
//   - tea.Cmd: next tick command.
func (m Model) handleTickMsg() (tea.Model, tea.Cmd) {
	m.tui.collectData()
	if m.tui.snapshot != nil {
		m.logsPanel.SetEntries(m.tui.snapshot.Logs.RecentEntries)
		m.servicesPanel.SetServices(m.tui.snapshot.Services)
	}
	return m, m.tick()
}

// updatePanelSizes recalculates panel sizes based on available space.
// Returns updated model to maintain value receiver consistency with tea.Model interface.
//
// Returns:
//   - Model: updated model with recalculated panel sizes
func (m Model) updatePanelSizes() Model {
	// Layout: Header (11 lines) | Services + System/Network | Logs | Status (1 line)
	// Standard terminal: 80x24.
	headerHeight := layoutHeaderHeight
	statusHeight := layoutStatusBarHeight

	// Determine layout to check if system section is rendered.
	size := terminal.Size{Cols: m.width, Rows: m.height}
	layout := terminal.GetLayout(size)

	// System section only rendered in non-compact modes.
	var systemHeight int
	if layout != terminal.LayoutCompact {
		systemHeight = layoutSystemSectionLines
	}

	// Available height for content.
	availableHeight := m.height - headerHeight - statusHeight - systemHeight

	// Services panel: adapts to number of services, max 10 visible (+3 for borders/header).
	servicesHeight := m.servicesPanel.OptimalHeight()

	// Ensure services panel doesn't exceed available space minus minimum logs.
	maxServicesHeight := availableHeight - layoutMinLogHeight
	if maxServicesHeight > 0 && servicesHeight > maxServicesHeight {
		servicesHeight = maxServicesHeight
	}

	// Remaining space goes to logs, enforcing minimum.
	logsHeight := max(availableHeight-servicesHeight, layoutMinLogHeight)

	m.logsPanel.SetSize(m.width, logsHeight)
	m.servicesPanel.SetSize(m.width, servicesHeight)

	return m
}

// View renders the UI.
//
// Returns:
//   - string: rendered UI.
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
	sb.WriteString(header.Render(snap))
	sb.WriteString("\n")

	switch layout {
	case terminal.LayoutCompact:
		sb.WriteString(m.renderCompact())
	case terminal.LayoutNormal:
		sb.WriteString(m.renderNormal(snap))
	case terminal.LayoutWide, terminal.LayoutUltraWide:
		sb.WriteString(m.renderWide(snap))
	}

	sb.WriteString(m.renderStatusBar(snap))

	return sb.String()
}

// renderCompact renders for small terminals (80x24).
// Shows only: Services panel (scrollable) + Logs panel.
// Note: snap parameter exists for API consistency with other render methods
// but is unused in compact mode since panels contain their own data.
//
// Returns:
//   - string: rendered compact layout.
func (m Model) renderCompact() string {
	var sb strings.Builder

	sb.WriteString(m.servicesPanel.View())
	sb.WriteString("\n")

	sb.WriteString(m.logsPanel.View())

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

	system := screen.NewSystemRenderer(m.width)
	sb.WriteString(system.RenderForInteractive(snap))
	sb.WriteString("\n")

	sb.WriteString(m.servicesPanel.View())
	sb.WriteString("\n")

	sb.WriteString(m.logsPanel.View())

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

	sb.WriteString(m.renderSystemNetworkSideBySide(snap))
	sb.WriteString("\n")

	sb.WriteString(m.servicesPanel.View())
	sb.WriteString("\n")

	sb.WriteString(m.logsPanel.View())

	return sb.String()
}

// renderSystemNetworkSideBySide renders system and network panels side by side.
//
// Params:
//   - snap: current snapshot.
//
// Returns:
//   - string: merged side-by-side content.
func (m Model) renderSystemNetworkSideBySide(snap *model.Snapshot) string {
	halfWidth := m.width / panelHalfWidthDivisor

	system := screen.NewSystemRenderer(halfWidth)
	network := screen.NewNetworkRenderer(halfWidth)
	systemContent := system.RenderForInteractive(snap)
	networkContent := network.Render(snap)

	systemLines := trimTrailingEmptyLines(strings.Split(systemContent, "\n"))
	networkLines := trimTrailingEmptyLines(strings.Split(networkContent, "\n"))

	return mergeLinesSideBySide(systemLines, networkLines, halfWidth)
}

// trimTrailingEmptyLines removes empty lines from the end of a slice.
//
// Params:
//   - lines: input lines.
//
// Returns:
//   - []string: lines without trailing empty lines.
func trimTrailingEmptyLines(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// mergeLinesSideBySide merges two line slices side by side.
//
// Params:
//   - left: left column lines.
//   - right: right column lines.
//   - leftWidth: width to pad left column to.
//
// Returns:
//   - string: merged side-by-side content.
func mergeLinesSideBySide(left, right []string, leftWidth int) string {
	var sb strings.Builder
	maxLines := max(len(left), len(right))

	for i := range maxLines {
		leftLine := ""
		rightLine := ""
		if i < len(left) {
			leftLine = left[i]
		}
		if i < len(right) {
			rightLine = right[i]
		}

		leftLine = padToWidth(leftLine, leftWidth)

		sb.WriteString(leftLine)
		sb.WriteString(rightLine)
		sb.WriteString("\n")
	}

	return sb.String()
}

// padToWidth pads a string to the specified visible width.
//
// Params:
//   - s: string to pad.
//   - width: target visible width.
//
// Returns:
//   - string: padded string.
func padToWidth(s string, width int) string {
	visible := widget.VisibleLen(s)
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}

// renderStatusBar renders the bottom status bar.
//
// Params:
//   - snap: current snapshot.
//
// Returns:
//   - string: rendered status bar.
func (m Model) renderStatusBar(snap *model.Snapshot) string {
	var focusIndicator string
	if m.focus == FocusLogs {
		focusIndicator = m.theme.Primary + "[LOGS]" + ansi.Reset
	} else {
		focusIndicator = m.theme.Primary + "[SERVICES]" + ansi.Reset
	}

	var keys string
	if m.focus == FocusLogs {
		keys = m.theme.Muted + "[↑↓] Scroll  [g/G] Top/Bottom  [s] Services  [Tab] Switch  [q] Quit" + ansi.Reset
	} else {
		keys = m.theme.Muted + "[↑↓] Scroll  [g/G] Top/Bottom  [l] Logs  [Tab] Switch  [q] Quit" + ansi.Reset
	}

	logs := screen.NewLogsRenderer(m.width)
	badge := logs.RenderBadge(snap)

	statusContent := "  " + focusIndicator + "  " + keys
	contentLen := widget.VisibleLen(statusContent)
	badgeLen := widget.VisibleLen(badge)
	padding := max(0, m.width-contentLen-badgeLen-statusBarBadgeSpacing)

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
	t.collectData()

	m := t.createInitialModel()

	return t.runTeaProgram(ctx, m)
}

// createInitialModel creates the initial Bubble Tea model.
//
// Returns:
//   - Model: configured model with panels.
func (t *TUI) createInitialModel() Model {
	size := terminal.GetSize()

	servicesPanel, logsPanel := t.createInitialPanels(size)

	m := NewModel(ModelConfig{
		TUI:           t,
		Width:         size.Cols,
		Height:        size.Rows,
		Theme:         ansi.DefaultTheme(),
		LogsPanel:     logsPanel,
		ServicesPanel: servicesPanel,
	})

	m.servicesPanel.SetFocused(true)

	if t.snapshot != nil {
		m.logsPanel.SetEntries(t.snapshot.Logs.RecentEntries)
	}

	return m
}

// createInitialPanels creates services and logs panels with calculated sizes.
//
// Params:
//   - size: terminal size.
//
// Returns:
//   - component.ServicesPanel: configured services panel.
//   - component.LogsPanel: configured logs panel.
func (t *TUI) createInitialPanels(size terminal.Size) (component.ServicesPanel, component.LogsPanel) {
	servicesPanel := component.NewServicesPanel(size.Cols, panelInitialServicesHeight)
	if t.snapshot != nil {
		servicesPanel.SetServices(t.snapshot.Services)
	}

	availableHeight := size.Rows - layoutHeaderHeight - layoutStatusBarHeight - layoutSystemSectionLines

	servicesHeight := servicesPanel.OptimalHeight()

	logsHeight := max(availableHeight-servicesHeight, layoutMinLogHeight)

	servicesPanel.SetSize(size.Cols, servicesHeight)

	return servicesPanel, component.NewLogsPanel(size.Cols, logsHeight)
}

// runTeaProgram runs the Bubble Tea program with context support.
// Spawns goroutine for Bubble Tea, handles context cancellation.
//
// Params:
//   - ctx: context for cancellation
//   - m: initial model
//
// Returns:
//   - error: any error during execution
func (t *TUI) runTeaProgram(ctx context.Context, m Model) error {
	prg := tea.NewProgram(m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run in goroutine to support context cancellation.
	// Goroutine lifecycle:
	//   - Starts: When this function is called, goroutine is spawned immediately
	//   - Runs: Until prg.Run() returns (user quits or error)
	//   - Ends: Sends result to done channel, then exits
	// Cleanup: Select below handles context cancellation or completion
	done := make(chan error, 1)
	go func() {
		// Goroutine exits when prg.Run() completes.
		_, err := prg.Run()
		done <- err
	}()

	select {
	case <-ctx.Done():
		prg.Quit()
		return ctx.Err()
	case err := <-done:
		return err
	}
}
