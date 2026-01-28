// Package ansi provides ANSI escape sequences for terminal styling.
//
// Tests: theme_internal_test.go (TODO: create tests for Theme and StatusIcon).
package ansi

// RGB color component constants for TrueColorTheme.
const (
	// colorMax is the maximum value for RGB components (white).
	colorMax uint8 = 255
	// colorMid is the middle value for RGB components (50% gray).
	colorMid uint8 = 128
	// colorOrange is the green/blue component for orange (255, 170, 0).
	colorOrange uint8 = 170
	// colorDarkGray is the RGB value for dark gray borders (85, 85, 85).
	colorDarkGray uint8 = 85
	// colorZero is the minimum value for RGB components (black).
	colorZero uint8 = 0
)

// Theme contains the superviz.io color scheme.
// Uses ANSI 256-color palette for broad terminal support.
//
// The theme defines semantic colors for different UI elements and states,
// ensuring consistent visual presentation across the TUI.
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
// Uses standard ANSI colors for maximum terminal compatibility.
//
// Returns:
//   - Theme: color scheme with standard ANSI colors
func DefaultTheme() Theme {
	// Return theme with standard ANSI color codes.
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
// Provides richer colors using RGB values for terminals with true color support.
//
// Returns:
//   - Theme: color scheme with 24-bit RGB colors
func TrueColorTheme() Theme {
	// Return theme with true color RGB values.
	return Theme{
		Primary: RGB(colorZero, colorMax, colorMax),    // Cyan
		Accent:  RGB(colorZero, colorMax, colorZero),   // Green
		Success: RGB(colorZero, colorMax, colorZero),   // Green
		Warning: RGB(colorMax, colorOrange, colorZero), // Orange
		Error:   RGB(colorMax, colorZero, colorZero),   // Red
		Muted:   RGB(colorMid, colorMid, colorMid),     // Gray
		Text:    Reset,
		Border:  RGB(colorDarkGray, colorDarkGray, colorDarkGray), // Dark gray
		Header:  Bold + RGB(colorMax, colorMax, colorMax),
	}
}

// Colorize applies a color to text and resets after.
// Wraps text with the specified color code and Reset sequence.
//
// Params:
//   - color: ANSI color escape sequence
//   - text: text to colorize
//
// Returns:
//   - string: colorized text with reset, or original text if color is empty
func Colorize(color, text string) string {
	// Check if color is provided.
	if color == "" {
		// Return unmodified text if no color specified.
		return text
	}

	// Wrap text with color and reset codes.
	return color + text + Reset
}

// BoldText applies bold styling to text.
// Wraps text with Bold and Reset sequences.
//
// Params:
//   - text: text to make bold
//
// Returns:
//   - string: bold text with reset
func BoldText(text string) string {
	// Wrap text with bold and reset codes.
	return Bold + text + Reset
}

// DimText applies dim styling to text.
// Wraps text with Dim and Reset sequences.
//
// Params:
//   - text: text to dim
//
// Returns:
//   - string: dimmed text with reset
func DimText(text string) string {
	// Wrap text with dim and reset codes.
	return Dim + text + Reset
}
