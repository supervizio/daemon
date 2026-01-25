// Package ansi provides ANSI escape sequences for terminal styling.
package ansi

// Theme contains the superviz.io color scheme.
// Uses ANSI 256-color palette for broad terminal support.
type Theme struct {
	// Primary is the main brand color (cyan).
	Primary string
	// Accent is the secondary brand color (green for .io).
	Accent string
	// Success indicates healthy/running state.
	Success string
	// Warning indicates starting/degraded state.
	Warning string
	// Error indicates failed/unhealthy state.
	Error string
	// Muted is for secondary/dim text.
	Muted string
	// Text is the default text color.
	Text string
	// Border is for box borders.
	Border string
	// Header is for section headers.
	Header string
}

// DefaultTheme returns the superviz.io color scheme.
func DefaultTheme() Theme {
	return Theme{
		Primary: FgCyan,
		Accent:  FgGreen,
		Success: FgGreen,
		Warning: FgYellow,
		Error:   FgRed,
		Muted:   FgGray,
		Text:    Reset,
		Border:  FgGray,
		Header:  Bold + FgWhite,
	}
}

// TrueColorTheme returns a 24-bit color theme for modern terminals.
func TrueColorTheme() Theme {
	return Theme{
		Primary: RGB(0, 255, 255),   // Cyan
		Accent:  RGB(0, 255, 0),     // Green
		Success: RGB(0, 255, 0),     // Green
		Warning: RGB(255, 170, 0),   // Orange
		Error:   RGB(255, 0, 0),     // Red
		Muted:   RGB(128, 128, 128), // Gray
		Text:    Reset,
		Border:  RGB(85, 85, 85), // Dark gray
		Header:  Bold + RGB(255, 255, 255),
	}
}

// StatusIcon returns the appropriate icon for a state.
type StatusIcon struct {
	Running  string
	Starting string
	Stopped  string
	Failed   string
	Healthy  string
	Unknown  string
}

// DefaultIcons returns Unicode status icons.
func DefaultIcons() StatusIcon {
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
func ASCIIIcons() StatusIcon {
	return StatusIcon{
		Running:  "*",
		Starting: "~",
		Stopped:  "o",
		Failed:   "x",
		Healthy:  "+",
		Unknown:  "?",
	}
}

// Colorize applies a color to text and resets after.
func Colorize(color, text string) string {
	if color == "" {
		return text
	}
	return color + text + Reset
}

// Bold makes text bold.
func BoldText(text string) string {
	return Bold + text + Reset
}

// Dim makes text dim.
func DimText(text string) string {
	return Dim + text + Reset
}
