// Package widget provides reusable TUI components.
package widget

import "regexp"

// ansiEscapeRegex matches ANSI escape sequences.
// Compiled once at package init to avoid per-call regex compilation overhead.
var ansiEscapeRegex *regexp.Regexp = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// StripANSI removes ANSI escape sequences from a string.
// This prevents terminal escape injection attacks where malicious service names
// or log messages could manipulate the terminal display (clear screen, move cursor, etc.).
//
// Params:
//   - s: input string potentially containing ANSI escape sequences
//
// Returns:
//   - string: sanitized string with all ANSI escape sequences removed
func StripANSI(s string) string {
	// Remove all ANSI escape sequences from input.
	return ansiEscapeRegex.ReplaceAllString(s, "")
}
