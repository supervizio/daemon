// Package component provides reusable Bubble Tea components.
package component

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// viewportModel defines the interface for viewport operations.
// This interface matches the viewport.Model API for navigation and update.
type viewportModel interface {
	GotoTop() []string
	GotoBottom() []string
	HalfPageUp() []string
	HalfPageDown() []string
	ScrollUp(lines int) []string
	ScrollDown(lines int) []string
	Update(msg tea.Msg) (viewport.Model, tea.Cmd)
}

// handleViewportKeyMsg processes keyboard input for viewport navigation.
// This shared function handles common keyboard shortcuts for scrolling.
// This function uses a concrete viewport.Model pointer because the Update
// method returns a new viewport.Model that must be assigned back.
// An interface cannot be used here due to this mutating update pattern.
//
// Params:
//   - vp: viewport model interface for navigation and update
//   - vpPtr: pointer to viewport model for assignment after Update
//   - msg: key message to process (uses Stringer interface)
//
// Returns:
//   - tea.Cmd: command to execute
func handleViewportKeyMsg(vp viewportModel, vpPtr *viewport.Model, msg Stringer) tea.Cmd {
	// Process keyboard shortcuts.
	switch msg.String() {
	// Handle home/top navigation.
	case "home", "g":
		vp.GotoTop()
		// Return no command.
		return nil
	// Handle end/bottom navigation.
	case "end", "G":
		vp.GotoBottom()
		// Return no command.
		return nil
	// Handle page up navigation.
	case "pgup", "ctrl+u":
		vp.HalfPageUp()
		// Return no command.
		return nil
	// Handle page down navigation.
	case "pgdown", "ctrl+d":
		vp.HalfPageDown()
		// Return no command.
		return nil
	// Handle line up navigation.
	case "up", "k":
		vp.ScrollUp(1)
		// Return no command.
		return nil
	// Handle line down navigation.
	case "down", "j":
		vp.ScrollDown(1)
		// Return no command.
		return nil
	// Handle other keys via viewport.
	default:
		var cmd tea.Cmd
		*vpPtr, cmd = vp.Update(msg)
		// Return viewport command.
		return cmd
	}
}

// ScrollbarParams groups parameters for scrollbar rendering.
// It reduces the number of function parameters for renderContentLinesWithScrollbar.
type ScrollbarParams struct {
	// Lines contains the content lines to render.
	Lines []string
	// ScrollbarChars contains characters for the scrollbar position.
	ScrollbarChars []string
	// Height is the number of lines to render.
	Height int
	// InnerWidth is the width inside borders.
	InnerWidth int
	// BorderColor is the ANSI color for border.
	BorderColor string
	// TrackChar is the character for scrollbar track.
	TrackChar string
}

// renderContentLinesWithScrollbar renders content lines with a vertical scrollbar.
// This shared function handles the common rendering pattern for scrollable content.
//
// Params:
//   - sb: string builder to write to
//   - params: scrollbar rendering parameters
func renderContentLinesWithScrollbar(sb *strings.Builder, params ScrollbarParams) {
	// Render each content line with scrollbar.
	for i := range params.Height {
		sb.WriteString(params.BorderColor)
		sb.WriteString("|")
		sb.WriteString(ansi.Reset)

		// Write content line or blank space.
		if i < len(params.Lines) {
			line := params.Lines[i]
			visLen := widget.VisibleLen(line)
			sb.WriteString(line)

			// Pad line if needed.
			if visLen < params.InnerWidth {
				sb.WriteString(strings.Repeat(" ", params.InnerWidth-visLen))
			}
		} else {
			// Write blank line.
			sb.WriteString(strings.Repeat(" ", params.InnerWidth))
		}

		// Write scrollbar character.
		sb.WriteString(params.BorderColor)

		// Select appropriate scrollbar character.
		if i < len(params.ScrollbarChars) {
			sb.WriteString(params.ScrollbarChars[i])
		} else {
			// Use track character as fallback.
			sb.WriteString(params.TrackChar)
		}
		sb.WriteString("|")
		sb.WriteString(ansi.Reset)
		sb.WriteString("\n")
	}
}
