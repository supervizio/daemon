// Package ansi provides ANSI escape sequences for terminal styling.
package ansi

// StatusIcon represents Unicode or ASCII icons for different process states.
//
// Provides visual indicators for service and health status in the TUI,
// with fallback ASCII variants for limited terminal environments.
type StatusIcon struct {
	// Running indicates an active process.
	Running string
	// Starting indicates a process in startup phase.
	Starting string
	// Stopped indicates an inactive process.
	Stopped string
	// Failed indicates a process that encountered an error.
	Failed string
	// Healthy indicates a passing health check.
	Healthy string
	// Unknown indicates an indeterminate state.
	Unknown string
}

// DefaultIcons returns Unicode status icons.
// Provides rich visual indicators for modern terminals with Unicode support.
//
// Returns:
//   - StatusIcon: set of Unicode status icons
func DefaultIcons() StatusIcon {
	// Return Unicode icon set.
	return StatusIcon{
		Running:  "●",
		Starting: "◐",
		Stopped:  "○",
		Failed:   "✗",
		Healthy:  "✓",
		Unknown:  "?",
	}
}

// ASCIIIcons returns ASCII-only status icons for limited terminals.
// Provides basic visual indicators that work in all terminal environments.
//
// Returns:
//   - StatusIcon: set of ASCII-only status icons
func ASCIIIcons() StatusIcon {
	// Return ASCII-only icon set.
	return StatusIcon{
		Running:  "*",
		Starting: "~",
		Stopped:  "o",
		Failed:   "x",
		Healthy:  "+",
		Unknown:  "?",
	}
}
