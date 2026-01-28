// Package layout provides responsive layout management for TUI.
package layout

// rowRegionParams holds parameters for generating row regions.
// Grouping to comply with FUNC-MAXPARAM (max 5 params).
type rowRegionParams struct {
	startX, yPos, height int
	colWidth, cols, gap  int
}
