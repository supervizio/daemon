// Package layout provides responsive layout management for TUI.
package layout

import (
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
)

// Region represents a rectangular area of the terminal.
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

// Layout calculates content regions based on terminal size.
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
func NewLayout(size terminal.Size) *Layout {
	return &Layout{
		Size:    size,
		Mode:    terminal.GetLayout(size),
		Padding: 1,
		Gap:     2,
	}
}

// ContentWidth returns the usable content width.
func (l *Layout) ContentWidth() int {
	return l.Size.Cols - (l.Padding * 2)
}

// ContentHeight returns the usable content height.
func (l *Layout) ContentHeight() int {
	return l.Size.Rows - (l.Padding * 2)
}

// Columns returns the number of content columns.
func (l *Layout) Columns() int {
	return l.Mode.Columns()
}

// ColumnWidth returns the width of each column.
func (l *Layout) ColumnWidth() int {
	cols := l.Columns()
	if cols <= 1 {
		return l.ContentWidth()
	}

	totalGap := (cols - 1) * l.Gap
	return (l.ContentWidth() - totalGap) / cols
}

// ColumnRegion returns the region for a specific column (0-indexed).
func (l *Layout) ColumnRegion(col int) Region {
	if col < 0 || col >= l.Columns() {
		col = 0
	}

	colWidth := l.ColumnWidth()
	x := l.Padding + col*(colWidth+l.Gap)

	return Region{
		X:      x,
		Y:      l.Padding,
		Width:  colWidth,
		Height: l.ContentHeight(),
	}
}

// FullWidthRegion returns a region spanning all columns.
func (l *Layout) FullWidthRegion() Region {
	return Region{
		X:      l.Padding,
		Y:      l.Padding,
		Width:  l.ContentWidth(),
		Height: l.ContentHeight(),
	}
}

// SplitHorizontal divides a region into n horizontal sections.
func SplitHorizontal(r Region, n int) []Region {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return []Region{r}
	}

	regions := make([]Region, n)
	height := r.Height / n

	for i := 0; i < n; i++ {
		regions[i] = Region{
			X:      r.X,
			Y:      r.Y + i*height,
			Width:  r.Width,
			Height: height,
		}
	}

	// Give remaining height to last region.
	if remaining := r.Height - (height * n); remaining > 0 {
		regions[n-1].Height += remaining
	}

	return regions
}

// SplitVertical divides a region into n vertical sections.
func SplitVertical(r Region, n int, gap int) []Region {
	if n <= 0 {
		return nil
	}
	if n == 1 {
		return []Region{r}
	}

	totalGap := (n - 1) * gap
	width := (r.Width - totalGap) / n
	regions := make([]Region, n)

	for i := 0; i < n; i++ {
		regions[i] = Region{
			X:      r.X + i*(width+gap),
			Y:      r.Y,
			Width:  width,
			Height: r.Height,
		}
	}

	// Give remaining width to last region.
	if remaining := r.Width - totalGap - (width * n); remaining > 0 {
		regions[n-1].Width += remaining
	}

	return regions
}

// Grid represents a flexible grid layout.
type Grid struct {
	layout *Layout
	rows   []gridRow
}

type gridRow struct {
	height int // 0 = flexible, >0 = fixed
	cells  []gridCell
}

type gridCell struct {
	colspan int
}

// NewGrid creates a grid for the given layout.
func NewGrid(layout *Layout) *Grid {
	return &Grid{
		layout: layout,
		rows:   make([]gridRow, 0),
	}
}

// AddRow adds a row with the specified height (0 for flexible).
func (g *Grid) AddRow(height int) *Grid {
	g.rows = append(g.rows, gridRow{
		height: height,
		cells:  make([]gridCell, 0),
	})
	return g
}

// AddCell adds a cell to the last row.
func (g *Grid) AddCell(colspan int) *Grid {
	if len(g.rows) == 0 {
		g.AddRow(0)
	}
	row := &g.rows[len(g.rows)-1]
	row.cells = append(row.cells, gridCell{colspan: colspan})
	return g
}

// Calculate returns regions for all cells.
func (g *Grid) Calculate() [][]Region {
	if len(g.rows) == 0 {
		return nil
	}

	content := g.layout.FullWidthRegion()
	cols := g.layout.Columns()
	gap := g.layout.Gap

	// Calculate row heights.
	fixedHeight := 0
	flexRows := 0
	for _, row := range g.rows {
		if row.height > 0 {
			fixedHeight += row.height
		} else {
			flexRows++
		}
	}

	flexHeight := 0
	if flexRows > 0 {
		flexHeight = (content.Height - fixedHeight) / flexRows
	}

	// Generate regions.
	result := make([][]Region, len(g.rows))
	y := content.Y

	for i, row := range g.rows {
		height := row.height
		if height <= 0 {
			height = flexHeight
		}

		result[i] = make([]Region, len(row.cells))
		x := content.X
		colWidth := g.layout.ColumnWidth()

		for j, cell := range row.cells {
			span := cell.colspan
			if span <= 0 {
				span = 1
			}
			if span > cols {
				span = cols
			}

			width := span*colWidth + (span-1)*gap

			result[i][j] = Region{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
			}

			x += width + gap
		}

		y += height
	}

	return result
}
