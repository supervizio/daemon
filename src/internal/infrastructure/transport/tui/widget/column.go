// Package widget provides reusable TUI components.
package widget

// Column defines a table column.
// It specifies header text, width constraints, alignment, and flexibility.
type Column struct {
	// Header is the column header text.
	Header string
	// Width is the column width (0 = auto).
	Width int
	// MinWidth is the minimum width.
	MinWidth int
	// MaxWidth is the maximum width (0 = no limit).
	MaxWidth int
	// Align is the text alignment.
	Align Align
	// Flex allows the column to expand to fill space.
	Flex bool
}
