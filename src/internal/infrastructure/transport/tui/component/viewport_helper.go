// Package component provides reusable Bubble Tea components.
package component

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
)

// handleViewportKeyMsg processes keyboard input for viewport navigation.
// This shared function handles common keyboard shortcuts for scrolling.
//
// Params:
//   - vp: pointer to viewport model (modified in place)
//   - msg: key message to process (uses Stringer interface)
//
// Returns:
//   - tea.Cmd: command to execute
func handleViewportKeyMsg(vp *viewport.Model, msg Stringer) tea.Cmd {
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
		*vp, cmd = vp.Update(msg)
		// Return viewport command.
		return cmd
	}
}

// renderContentLinesWithScrollbar renders content lines with a vertical scrollbar.
// This shared function handles the common rendering pattern for scrollable content.
//
// Params:
//   - sb: string builder to write to
//   - lines: content lines to render
//   - scrollbarChars: characters for the scrollbar
//   - height: number of lines to render
//   - innerWidth: width inside borders
//   - borderColor: ANSI color for border
//   - trackChar: character to use for scrollbar track
func renderContentLinesWithScrollbar(
	sb *strings.Builder,
	lines []string,
	scrollbarChars []string,
	height, innerWidth int,
	borderColor, trackChar string,
) {
	// Render each content line with scrollbar.
	for i := range height {
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
			sb.WriteString(trackChar)
		}
		sb.WriteString("|")
		sb.WriteString(ansi.Reset)
		sb.WriteString("\n")
	}
}
