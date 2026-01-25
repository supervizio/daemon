// Package tui provides terminal user interface for superviz.io.
package tui

import (
	"context"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/screen"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
)

// Model is the Bubble Tea model.
type Model struct {
	tui      *TUI
	width    int
	height   int
	quitting bool
}

// tickMsg is sent on each refresh interval.
type tickMsg time.Time

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
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		// Refresh data.
		m.tui.collectData()
		return m, m.tick()
	}

	return m, nil
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

	// Header (with time in interactive mode).
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
	sb.WriteString("\n")
	sb.WriteString(m.renderStatusBar(snap))

	return sb.String()
}

// renderCompact renders for small terminals.
func (m Model) renderCompact(snap *model.Snapshot) string {
	var sb strings.Builder

	// Services only.
	services := screen.NewServicesRenderer(m.width)
	sb.WriteString(services.Render(snap))
	sb.WriteString("\n")

	// System (compact).
	system := screen.NewSystemRenderer(m.width)
	sb.WriteString(system.Render(snap))

	return sb.String()
}

// renderNormal renders for normal terminals (single column).
func (m Model) renderNormal(snap *model.Snapshot) string {
	var sb strings.Builder

	// Services.
	services := screen.NewServicesRenderer(m.width)
	sb.WriteString(services.Render(snap))
	sb.WriteString("\n")

	// System.
	system := screen.NewSystemRenderer(m.width)
	sb.WriteString(system.Render(snap))
	sb.WriteString("\n")

	// Network (if space).
	if m.height > 25 {
		network := screen.NewNetworkRenderer(m.width)
		sb.WriteString(network.Render(snap))
	}

	return sb.String()
}

// renderWide renders for wide terminals (two columns).
func (m Model) renderWide(snap *model.Snapshot) string {
	var sb strings.Builder

	// Calculate column widths.
	gap := 2
	leftWidth := (m.width - gap) * 2 / 3
	rightWidth := m.width - gap - leftWidth

	// Left column: Services.
	services := screen.NewServicesRenderer(leftWidth)
	servicesContent := services.Render(snap)

	// Right column: System + Network.
	system := screen.NewSystemRenderer(rightWidth)
	network := screen.NewNetworkRenderer(rightWidth)
	systemContent := system.Render(snap)
	networkContent := network.Render(snap)

	// Combine columns.
	leftLines := strings.Split(servicesContent, "\n")
	rightLines := strings.Split(systemContent+"\n"+networkContent, "\n")

	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	for i := 0; i < maxLines; i++ {
		left := ""
		right := ""

		if i < len(leftLines) {
			left = leftLines[i]
		}
		if i < len(rightLines) {
			right = rightLines[i]
		}

		// Pad left column.
		leftVisible := visibleLen(left)
		if leftVisible < leftWidth {
			left += strings.Repeat(" ", leftWidth-leftVisible)
		}

		sb.WriteString(left)
		sb.WriteString(strings.Repeat(" ", gap))
		sb.WriteString(right)
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderStatusBar renders the bottom status bar.
func (m Model) renderStatusBar(snap *model.Snapshot) string {
	theme := ansi.DefaultTheme()

	// Keybindings.
	keys := theme.Muted + "[q] Quit  [r] Restart  [s] Stop  [l] Logs  [↑↓] Select" + ansi.Reset

	// Error badge.
	logs := screen.NewLogsRenderer(m.width)
	badge := logs.RenderBadge(snap)

	// Combine.
	keysLen := 55 // Approximate visible length.
	badgeLen := 15
	padding := m.width - keysLen - badgeLen - 4

	if padding < 0 {
		padding = 0
	}

	return "  " + keys + strings.Repeat(" ", padding) + badge
}

// runBubbleTea starts the Bubble Tea program.
func (t *TUI) runBubbleTea(ctx context.Context) error {
	// Initial data collection.
	t.collectData()

	// Get initial size.
	size := terminal.GetSize()

	m := Model{
		tui:    t,
		width:  size.Cols,
		height: size.Rows,
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

// visibleLen calculates visible string length.
func visibleLen(s string) int {
	inEscape := false
	length := 0

	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		length++
	}

	return length
}
