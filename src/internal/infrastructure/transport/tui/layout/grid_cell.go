// Package layout provides responsive layout management for TUI.
package layout

// gridCell represents a single cell in a grid row.
// It defines how many columns the cell spans.
type gridCell struct {
	// colspan is the number of columns this cell spans.
	colspan int
}
