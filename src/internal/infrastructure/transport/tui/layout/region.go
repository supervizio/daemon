// Package layout provides responsive layout management for TUI.
package layout

// Region represents a rectangular area of the terminal.
// It defines the position (X, Y) and dimensions (Width, Height) for rendering content.
type Region struct {
	// X is the starting column (0-indexed).
	X int
	// Y is the starting row (0-indexed).
	Y int
	// Width in columns.
	Width int
	// Height in rows.
	Height int
}
