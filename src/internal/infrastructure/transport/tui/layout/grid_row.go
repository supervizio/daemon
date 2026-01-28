// Package layout provides responsive layout management for TUI.
package layout

// gridRow represents a single row in the grid.
// Height of 0 means flexible, positive values indicate fixed height.
type gridRow struct {
	// height is the row height (0 = flexible, >0 = fixed).
	height int
	// cells contains the cell definitions for this row.
	cells []gridCell
}
