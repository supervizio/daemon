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
	// evaluate condition.
	if r == escapeStart {
		ts.inEscape = true
		result.WriteRune(r)
		// return false for failure.
		return false
	}

	// evaluate condition.
	if ts.inEscape {
		result.WriteRune(r)
		// evaluate condition.
		if isEscapeTerminator(r) {
			ts.inEscape = false
		}
		// return false for failure.
		return false
	}

	// evaluate condition.
	if ts.visible >= ts.maxLen {
		// return true for success.
		return true
	}

	result.WriteRune(r)
	ts.visible++
	// return false for failure.
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
	// evaluate condition.
	if maxLen <= 0 {
		// return computed result.
		return ""
	}

	var result strings.Builder
	state := truncateState{inEscape: false, visible: 0, maxLen: maxLen}

	// iterate over collection.
	for _, r := range s {
		// evaluate condition.
		if state.processRune(&result, r) {
			break
		}
	}

	result.WriteString(ansi.Reset)
	// return computed result.
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
	// return computed result.
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}
