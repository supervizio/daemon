// Package widget provides reusable TUI components.
package widget

import (
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
)

const (
	// tableEstimatedRowSize is the estimated size per row for pre-allocation.
	tableEstimatedRowSize int = 20
	// tableHeaderRowCount is the number of extra rows for header and separator.
	tableHeaderRowCount int = 2
	// tableCompactExtraSize is extra size for compact table pre-allocation.
	tableCompactExtraSize int = 10
	// defaultColumnCapacity is the default capacity for table columns.
	defaultColumnCapacity int = 8
	// defaultRowCapacity is the default capacity for table rows.
	defaultRowCapacity int = 16
)

// Table renders a data table.
// It supports headers, column alignment, flexible widths, and optional separators.
type Table struct {
	Columns     []Column
	Rows        [][]string
	Width       int
	ShowHeader  bool
	HeaderColor string
	BorderColor string
	RowColors   []string // Alternating row colors.
	Separator   string
}

// NewTable creates a new table with default configuration.
//
// Params:
//   - width: the total width available for the table in characters.
//
// Returns:
//   - *Table: a new table instance with empty columns and rows.
func NewTable(width int) *Table {
	// Build and return table with default colors and separator.
	return &Table{
		Columns:     make([]Column, 0, defaultColumnCapacity),
		Rows:        make([][]string, 0, defaultRowCapacity),
		Width:       width,
		ShowHeader:  true,
		HeaderColor: ansi.Bold + ansi.FgWhite,
		BorderColor: ansi.FgGray,
		Separator:   "  ",
	}
}

// AddColumn adds a fixed-width column definition to the table.
//
// Params:
//   - header: the column header text to display.
//   - width: the fixed width for the column in characters.
//   - align: the text alignment for the column content.
//
// Returns:
//   - *Table: the table instance for method chaining.
func (t *Table) AddColumn(header string, width int, align Align) *Table {
	// Append column with alignment to columns slice.
	t.Columns = append(t.Columns, Column{
		Header: header,
		Width:  width,
		Align:  align,
	})
	// Return self for fluent interface.
	return t
}

// AddFlexColumn adds a flexible column that expands to fill available space.
//
// Params:
//   - header: the column header text to display.
//   - minWidth: the minimum width for the column in characters.
//   - align: the text alignment for the column content.
//
// Returns:
//   - *Table: the table instance for method chaining.
func (t *Table) AddFlexColumn(header string, minWidth int, align Align) *Table {
	// Append flexible column to columns slice.
	t.Columns = append(t.Columns, Column{
		Header:   header,
		MinWidth: minWidth,
		Align:    align,
		Flex:     true,
	})
	// Return self for fluent interface.
	return t
}

// AddRow adds a data row to the table.
//
// Params:
//   - cells: the cell values for the row.
//
// Returns:
//   - *Table: the table instance for method chaining.
func (t *Table) AddRow(cells ...string) *Table {
	// Append row to rows slice.
	t.Rows = append(t.Rows, cells)
	// Return self for fluent interface.
	return t
}

// SetHeader enables or disables the header row.
//
// Params:
//   - show: whether to display the header.
//
// Returns:
//   - *Table: the table instance for method chaining.
func (t *Table) SetHeader(show bool) *Table {
	// Set header visibility flag.
	t.ShowHeader = show
	// Return self for fluent interface.
	return t
}

// SetHeaderColor sets the color for header text.
//
// Params:
//   - color: ANSI color code for the header.
//
// Returns:
//   - *Table: the table instance for method chaining.
func (t *Table) SetHeaderColor(color string) *Table {
	// Set header color.
	t.HeaderColor = color
	// Return self for fluent interface.
	return t
}

// Render returns the complete table as a multi-line string.
//
// Returns:
//   - string: the rendered table with header, rows, and borders.
func (t *Table) Render() string {
	// Handle empty table case.
	if len(t.Columns) == 0 {
		// Return empty string for no columns.
		return ""
	}

	// Calculate column widths.
	widths := t.calculateWidths()

	// Pre-allocate string builder with estimated capacity.
	estimatedSize := len(t.Rows)*tableEstimatedRowSize + tableHeaderRowCount
	var sb strings.Builder
	sb.Grow(estimatedSize)

	// Render header if enabled.
	if t.ShowHeader {
		sb.WriteString(t.renderHeader(widths))
		sb.WriteByte('\n')
	}

	// Render all data rows.
	sb.WriteString(t.renderRows(widths))

	// Convert to string and return.
	return sb.String()
}

// renderHeader formats the header row with column names.
//
// Params:
//   - widths: calculated width for each column.
//
// Returns:
//   - string: the formatted header row.
func (t *Table) renderHeader(widths []int) string {
	// Pre-allocate string builder with estimated capacity.
	var sb strings.Builder
	sb.Grow(t.Width)

	// Iterate through columns to build header.
	for i, col := range t.Columns {
		// Add separator before all columns except the first.
		if i > 0 {
			sb.WriteString(t.Separator)
		}

		// Format header with alignment and color.
		sb.WriteString(t.HeaderColor)
		sb.WriteString(Pad(col.Header, widths[i], col.Align))
		sb.WriteString(ansi.Reset)
	}

	// Convert to string and return.
	return sb.String()
}

// renderRows formats all data rows.
//
// Params:
//   - widths: calculated width for each column.
//
// Returns:
//   - string: all formatted rows joined with newlines.
func (t *Table) renderRows(widths []int) string {
	// Handle empty rows case.
	if len(t.Rows) == 0 {
		// Return empty string for no rows.
		return ""
	}

	// Pre-allocate string builder with estimated capacity.
	estimatedSize := len(t.Rows) * tableEstimatedRowSize
	var sb strings.Builder
	sb.Grow(estimatedSize)

	// Iterate through rows and format each.
	for rowIdx, row := range t.Rows {
		// Add newline before all rows except the first.
		if rowIdx > 0 {
			sb.WriteByte('\n')
		}

		// Render single row with colors and cells.
		t.renderSingleRow(&sb, row, rowIdx, widths)
	}

	// Convert to string and return.
	return sb.String()
}

// renderSingleRow formats a single data row.
//
// Params:
//   - sb: string builder to write to.
//   - row: cell values for this row.
//   - rowIdx: row index for color alternation.
//   - widths: calculated column widths.
func (t *Table) renderSingleRow(sb *strings.Builder, row []string, rowIdx int, widths []int) {
	// Apply row color if configured.
	if len(t.RowColors) > 0 {
		color := t.RowColors[rowIdx%len(t.RowColors)]
		sb.WriteString(color)
	}

	// Format each cell in the row.
	t.renderRowCells(sb, row, widths)

	// Reset color if row colors are used.
	if len(t.RowColors) > 0 {
		sb.WriteString(ansi.Reset)
	}
}

// renderRowCells formats all cells in a row.
//
// Params:
//   - sb: string builder to write to.
//   - row: cell values for this row.
//   - widths: calculated column widths.
func (t *Table) renderRowCells(sb *strings.Builder, row []string, widths []int) {
	// Format each cell in the row.
	for colIdx, cell := range row {
		// Add separator before all columns except the first.
		if colIdx > 0 {
			sb.WriteString(t.Separator)
		}

		// Ensure column index is valid.
		if colIdx < len(t.Columns) {
			col := t.Columns[colIdx]
			sb.WriteString(Pad(cell, widths[colIdx], col.Align))
		}
	}
}

// RenderCompact returns a compact table without header.
//
// Returns:
//   - string: the rendered table rows only.
func (t *Table) RenderCompact() string {
	// Handle empty table case.
	if len(t.Columns) == 0 {
		// Return empty string for no columns.
		return ""
	}

	// Calculate column widths.
	widths := t.calculateWidths()

	// Pre-allocate string builder with estimated capacity.
	estimatedSize := len(t.Rows)*tableEstimatedRowSize + tableCompactExtraSize
	var sb strings.Builder
	sb.Grow(estimatedSize)

	// Render rows only (skip header).
	sb.WriteString(t.renderRows(widths))

	// Convert to string and return.
	return sb.String()
}

// calculateWidths determines the width of each column.
//
// Returns:
//   - []int: width for each column.
func (t *Table) calculateWidths() []int {
	// Initialize widths slice.
	widths := t.calculateInitialWidths()

	// Calculate total used width.
	totalUsed := 0
	// iterate over collection.
	for _, w := range widths {
		totalUsed += w
	}

	// Add separator widths to total.
	separatorWidth := len(t.Separator) * (len(t.Columns) - 1)
	totalUsed += separatorWidth

	// Distribute remaining space to flex columns.
	if totalUsed < t.Width {
		remaining := t.Width - totalUsed
		widths = t.distributeFlexWidth(widths, remaining)
	}

	// Return calculated widths.
	return widths
}

// calculateInitialWidths calculates initial column widths.
//
// Returns:
//   - []int: initial width for each column.
func (t *Table) calculateInitialWidths() []int {
	// Pre-allocate widths slice with capacity.
	widths := make([]int, 0, len(t.Columns))

	// Calculate width for each column.
	for _, col := range t.Columns {
		widths = append(widths, t.calculateColumnWidth(col))
	}

	// Return initial widths.
	return widths
}

// calculateColumnWidth calculates the width for a single column.
//
// Params:
//   - col: the column to calculate width for.
//
// Returns:
//   - int: the calculated width.
func (t *Table) calculateColumnWidth(col Column) int {
	// Use fixed width if specified.
	if col.Width > 0 {
		// Return fixed width.
		return col.Width
	}

	// Use minimum width if specified.
	if col.MinWidth > 0 {
		// Return minimum width.
		return col.MinWidth
	}

	// Calculate automatic width.
	return t.calculateAutoWidth(col)
}

// calculateAutoWidth calculates automatic width for a column.
//
// Params:
//   - col: the column to calculate width for.
//
// Returns:
//   - int: the calculated automatic width.
func (t *Table) calculateAutoWidth(col Column) int {
	// Start with header width.
	maxWidth := len(col.Header)

	// Find the index of this column.
	colIdx := -1
	// iterate over collection.
	for i, c := range t.Columns {
		// Compare column pointers.
		if c.Header == col.Header {
			colIdx = i
			// Stop after finding the column.
			break
		}
	}

	// If column found, check row cell widths.
	if colIdx >= 0 {
		// iterate over collection.
		for _, row := range t.Rows {
			// Ensure row has this column.
			if colIdx < len(row) {
				cellWidth := len(row[colIdx])
				// Update max width if larger.
				if cellWidth > maxWidth {
					maxWidth = cellWidth
				}
			}
		}
	}

	// Return maximum width found.
	return maxWidth
}

// distributeFlexWidth distributes remaining width among flex columns.
//
// Params:
//   - widths: current column widths.
//   - remaining: remaining width to distribute.
//
// Returns:
//   - []int: updated widths with flex distribution.
func (t *Table) distributeFlexWidth(widths []int, remaining int) []int {
	// Count flex columns.
	flexCount := 0
	// iterate over collection.
	for _, col := range t.Columns {
		// Count flex columns.
		if col.Flex {
			flexCount++
		}
	}

	// If no flex columns, return original widths.
	if flexCount == 0 {
		// No flex columns to distribute to.
		return widths
	}

	// Calculate extra width per flex column.
	extraPerFlex := remaining / flexCount

	// Distribute extra width to flex columns.
	for i, col := range t.Columns {
		// Only distribute to flex columns.
		if col.Flex {
			widths[i] += extraPerFlex
		}
	}

	// Return updated widths.
	return widths
}
