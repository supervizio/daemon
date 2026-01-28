// Package layout provides responsive layout management for TUI.
package layout

import (
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
)

// Layout calculation constants.
const (
	// defaultPadding is the default padding around content.
	defaultPadding int = 1

	// defaultGap is the default gap between columns.
	defaultGap int = 2

	// paddingSides is the number of sides with padding (left and right).
	paddingSides int = 2

	// minContentDimension is the minimum content width/height.
	minContentDimension int = 1

	// singleSection indicates a single section (no splitting needed).
	singleSection int = 1

	// noSections indicates zero sections requested.
	noSections int = 0
)

// Layout calculates content regions based on terminal size.
// It supports multiple layout modes (compact, normal, wide) with responsive breakpoints.
type Layout struct {
	// Size is the current terminal size.
	Size terminal.Size
	// Mode is the layout mode.
	Mode terminal.Layout
	// Padding around content.
	Padding int
	// Gap between columns.
	Gap int
}

// NewLayout creates a layout for the given terminal size.
//
// Params:
//   - size: terminal dimensions to create layout for
//
// Returns:
//   - *Layout: new layout instance configured for the terminal size
func NewLayout(size terminal.Size) *Layout {
	// Return a new layout with default padding and gap.
	return &Layout{
		Size:    size,
		Mode:    terminal.GetLayout(size),
		Padding: defaultPadding,
		Gap:     defaultGap,
	}
}

// ContentWidth returns the usable content width.
// Returns at least 1 to prevent negative widths on tiny terminals.
//
// Returns:
//   - int: usable content width in columns
func (l *Layout) ContentWidth() int {
	// Return content width with minimum of 1.
	return max(l.Size.Cols-(l.Padding*paddingSides), minContentDimension)
}

// ContentHeight returns the usable content height.
// Returns at least 1 to prevent negative heights on tiny terminals.
//
// Returns:
//   - int: usable content height in rows
func (l *Layout) ContentHeight() int {
	// Return content height with minimum of 1.
	return max(l.Size.Rows-(l.Padding*paddingSides), minContentDimension)
}

// Columns returns the number of content columns.
//
// Returns:
//   - int: number of columns for the current layout mode
func (l *Layout) Columns() int {
	// Return the column count from the layout mode.
	return l.Mode.Columns()
}

// ColumnWidth returns the width of each column.
// Returns at least 1 to prevent zero/negative widths.
//
// Returns:
//   - int: width of each column in characters
func (l *Layout) ColumnWidth() int {
	cols := l.Columns()

	// Handle single-column layout by returning full content width.
	if cols <= singleSection {
		// Return full content width for single column.
		return l.ContentWidth()
	}

	totalGap := (cols - singleSection) * l.Gap

	// Return calculated column width with minimum of 1.
	return max((l.ContentWidth()-totalGap)/cols, minContentDimension)
}

// ColumnRegion returns the region for a specific column (0-indexed).
//
// Params:
//   - col: column index (0-indexed)
//
// Returns:
//   - Region: rectangular area for the specified column
func (l *Layout) ColumnRegion(col int) Region {
	// Validate column index and reset to 0 if invalid.
	if col < noSections || col >= l.Columns() {
		col = noSections
	}

	colWidth := l.ColumnWidth()
	xPos := l.Padding + col*(colWidth+l.Gap)

	// Return the calculated region for the column.
	return Region{
		X:      xPos,
		Y:      l.Padding,
		Width:  colWidth,
		Height: l.ContentHeight(),
	}
}

// FullWidthRegion returns a region spanning all columns.
//
// Returns:
//   - Region: rectangular area spanning the full content width
func (l *Layout) FullWidthRegion() Region {
	// Return a region covering the entire content area.
	return Region{
		X:      l.Padding,
		Y:      l.Padding,
		Width:  l.ContentWidth(),
		Height: l.ContentHeight(),
	}
}

// SplitHorizontal divides a region into n horizontal sections.
//
// Params:
//   - r: region to split
//   - n: number of sections to create
//
// Returns:
//   - []Region: slice of regions stacked vertically
func SplitHorizontal(r Region, n int) []Region {
	// Handle invalid section count by returning nil.
	if n <= noSections {
		// Return nil for zero or negative sections.
		return nil
	}

	// Handle single section by returning the original region.
	if n == singleSection {
		// Return original region as single-element slice.
		return []Region{r}
	}

	// Pre-allocate with capacity using append pattern per VAR-MAKEAPPEND.
	regions := make([]Region, 0, n)
	height := r.Height / n

	// Create regions for each horizontal section.
	for i := range n {
		regions = append(regions, Region{
			X:      r.X,
			Y:      r.Y + i*height,
			Width:  r.Width,
			Height: height,
		})
	}

	// Give remaining height to last region.
	if remaining := r.Height - (height * n); remaining > noSections {
		regions[n-singleSection].Height += remaining
	}

	// Return the calculated regions.
	return regions
}

// SplitVertical divides a region into n vertical sections.
//
// Params:
//   - r: region to split
//   - n: number of sections to create
//   - gap: spacing between sections
//
// Returns:
//   - []Region: slice of regions arranged horizontally
func SplitVertical(r Region, n, gap int) []Region {
	// Handle invalid section count by returning nil.
	if n <= noSections {
		// Return nil for zero or negative sections.
		return nil
	}

	// Handle single section by returning the original region.
	if n == singleSection {
		// Return original region as single-element slice.
		return []Region{r}
	}

	totalGap := (n - singleSection) * gap
	width := (r.Width - totalGap) / n

	// Pre-allocate with capacity using append pattern per VAR-MAKEAPPEND.
	regions := make([]Region, 0, n)

	// Create regions for each vertical section.
	for i := range n {
		regions = append(regions, Region{
			X:      r.X + i*(width+gap),
			Y:      r.Y,
			Width:  width,
			Height: r.Height,
		})
	}

	// Give remaining width to last region.
	if remaining := r.Width - totalGap - (width * n); remaining > noSections {
		regions[n-singleSection].Width += remaining
	}

	// Return the calculated regions.
	return regions
}
