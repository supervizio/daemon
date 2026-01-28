// Package layout provides responsive layout management for TUI.
package layout

// Grid calculation constants.
const (
	// defaultColspan is the default column span for cells.
	defaultColspan int = 1

	// flexibleRowHeight indicates a row with flexible height.
	flexibleRowHeight int = 0
)

// Grid represents a flexible grid layout.
// It allows defining rows with fixed or flexible heights and cells with column spanning.
type Grid struct {
	// layout is the parent layout for dimension calculations.
	layout *Layout
	// rows contains the grid row definitions.
	rows []gridRow
}

// NewGrid creates a grid for the given layout.
//
// Params:
//   - layout: parent layout for dimension calculations
//
// Returns:
//   - *Grid: new grid instance with empty rows
func NewGrid(layout *Layout) *Grid {
	// Return a new grid with the specified layout.
	// Using nil slice per VAR-NILSLICE (rows added dynamically).
	return &Grid{
		layout: layout,
		rows:   nil,
	}
}

// AddRow adds a row with the specified height (0 for flexible).
//
// Params:
//   - height: row height in rows (0 for flexible)
//
// Returns:
//   - *Grid: the grid for method chaining
func (g *Grid) AddRow(height int) *Grid {
	// Append a new row with the specified height.
	// Using nil slice per VAR-NILSLICE (cells added dynamically).
	g.rows = append(g.rows, gridRow{
		height: height,
		cells:  nil,
	})

	// Return the grid for chaining.
	return g
}

// AddCell adds a cell to the last row.
//
// Params:
//   - colspan: number of columns this cell spans
//
// Returns:
//   - *Grid: the grid for method chaining
func (g *Grid) AddCell(colspan int) *Grid {
	// Create a row if none exist.
	if len(g.rows) == noSections {
		// Add a flexible row to hold the cell.
		g.AddRow(flexibleRowHeight)
	}

	row := &g.rows[len(g.rows)-singleSection]
	row.cells = append(row.cells, gridCell{colspan: colspan})

	// Return the grid for chaining.
	return g
}

// Calculate returns regions for all cells.
//
// Returns:
//   - [][]Region: 2D slice of regions indexed by [row][cell]
func (g *Grid) Calculate() [][]Region {
	// Handle empty grid by returning nil.
	if len(g.rows) == noSections {
		// Return nil for empty grid.
		return nil
	}

	content := g.layout.FullWidthRegion()
	flexHeight := g.calculateFlexibleHeight(content.Height)

	// Generate regions for all cells.
	return g.generateAllRegions(content, flexHeight)
}

// calculateFlexibleHeight calculates the height for flexible rows.
//
// Params:
//   - totalHeight: total available height.
//
// Returns:
//   - int: height for each flexible row.
func (g *Grid) calculateFlexibleHeight(totalHeight int) int {
	// Calculate row heights by summing fixed heights and counting flexible rows.
	fixedHeight := 0
	flexRows := 0

	// Iterate through rows to calculate fixed and flexible heights.
	for _, row := range g.rows {
		// Check if this row has a fixed height.
		if row.height > flexibleRowHeight {
			fixedHeight += row.height
		} else {
			// Count flexible rows for later distribution.
			flexRows++
		}
	}

	// Calculate flexible row height if there are flexible rows.
	if flexRows > noSections {
		// Return calculated flexible height.
		return (totalHeight - fixedHeight) / flexRows
	}

	// Return 0 for no flexible rows.
	return 0
}

// generateAllRegions generates regions for all rows and cells.
//
// Params:
//   - content: content region for positioning.
//   - flexHeight: height for flexible rows.
//
// Returns:
//   - [][]Region: 2D slice of regions indexed by [row][cell].
func (g *Grid) generateAllRegions(content Region, flexHeight int) [][]Region {
	// Pre-allocate with capacity per VAR-MAKEAPPEND.
	result := make([][]Region, 0, len(g.rows))
	yPos := content.Y
	colWidth := g.layout.ColumnWidth()
	cols := g.layout.Columns()
	gap := g.layout.Gap

	// Iterate through each row to calculate cell regions.
	for _, row := range g.rows {
		height := g.getRowHeight(row, flexHeight)
		params := rowRegionParams{
			startX:   content.X,
			yPos:     yPos,
			height:   height,
			colWidth: colWidth,
			cols:     cols,
			gap:      gap,
		}
		rowRegions := g.generateRowRegions(row, params)
		result = append(result, rowRegions)
		yPos += height
	}

	// Return the calculated regions.
	return result
}

// getRowHeight returns the height for a row.
//
// Params:
//   - row: the row to get height for.
//   - flexHeight: height for flexible rows.
//
// Returns:
//   - int: row height.
func (g *Grid) getRowHeight(row gridRow, flexHeight int) int {
	// Use flexible height for rows with zero or negative height.
	if row.height <= flexibleRowHeight {
		// Return flexible height.
		return flexHeight
	}
	// Return fixed height.
	return row.height
}

// generateRowRegions generates regions for cells in a row.
//
// Params:
//   - row: row containing cells.
//   - params: position and size parameters for region generation.
//
// Returns:
//   - []Region: regions for all cells in the row.
func (g *Grid) generateRowRegions(row gridRow, params rowRegionParams) []Region {
	// Pre-allocate row regions with capacity per VAR-MAKEAPPEND.
	rowRegions := make([]Region, 0, len(row.cells))
	xPos := params.startX

	// Iterate through cells to calculate their regions.
	for _, cell := range row.cells {
		span := cell.colspan

		// Ensure minimum colspan of 1.
		if span <= noSections {
			span = defaultColspan
		}

		// Limit colspan to maximum available columns.
		if span > params.cols {
			span = params.cols
		}

		width := span*params.colWidth + (span-singleSection)*params.gap

		rowRegions = append(rowRegions, Region{
			X:      xPos,
			Y:      params.yPos,
			Width:  width,
			Height: params.height,
		})

		xPos += width + params.gap
	}

	// Return calculated regions.
	return rowRegions
}
