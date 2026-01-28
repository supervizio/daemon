// Package widget provides reusable TUI components.
package widget

import (
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
)

// truncateState holds state for truncation processing.
type truncateState struct {
	inEscape bool
	visible  int
	maxLen   int
}

// processRune processes a single rune during truncation.
// Returns true if truncation limit reached and processing should stop.
//
// Params:
//   - result: string builder to write to
//   - r: rune to process
//
// Returns:
//   - bool: true if limit reached
func (ts *truncateState) processRune(result *strings.Builder, r rune) bool {
	if r == escapeStart {
		ts.inEscape = true
		result.WriteRune(r)
		return false
	}

	if ts.inEscape {
		result.WriteRune(r)
		if isEscapeTerminator(r) {
			ts.inEscape = false
		}
		return false
	}

	if ts.visible >= ts.maxLen {
		return true
	}

	result.WriteRune(r)
	ts.visible++
	return false
}

// truncateVisible truncates a string to maxLen visible characters.
//
// Params:
//   - s: string to truncate containing text and optional ANSI codes.
//   - maxLen: maximum visible character count.
//
// Returns:
//   - string: truncated string with ANSI reset appended.
func truncateVisible(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	var result strings.Builder
	state := truncateState{inEscape: false, visible: 0, maxLen: maxLen}

	for _, r := range s {
		if state.processRune(&result, r) {
			break
		}
	}

	result.WriteString(ansi.Reset)
	return result.String()
}

// TruncateVisible truncates a string to maxLen visible characters.
//
// Params:
//   - s: string to truncate containing text and optional ANSI codes.
//   - maxLen: maximum visible character count.
//
// Returns:
//   - string: truncated string with ANSI reset appended.
func TruncateVisible(s string, maxLen int) string {
	// Delegate to internal function.
	return truncateVisible(s, maxLen)
}

// isEscapeTerminator checks if rune terminates an ANSI escape sequence.
//
// Params:
//   - r: rune to check
//
// Returns:
//   - bool: true if rune is a letter (escape terminator)
func isEscapeTerminator(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}
