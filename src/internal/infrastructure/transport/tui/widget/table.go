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
)

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
		Columns:     make([]Column, 0),
		Rows:        make([][]string, 0),
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
	t.Columns = append(t.Columns, Column{
		Header:   header,
		Width:    width,
		MinWidth: len(header),
		Align:    align,
	})

	// Return table for chaining.
	return t
}

// AddFlexColumn adds a flexible-width column that expands to fill available space.
//
// Params:
//   - header: the column header text to display.
//   - minWidth: the minimum width for the column in characters.
//   - align: the text alignment for the column content.
//
// Returns:
//   - *Table: the table instance for method chaining.
func (t *Table) AddFlexColumn(header string, minWidth int, align Align) *Table {
	t.Columns = append(t.Columns, Column{
		Header:   header,
		MinWidth: minWidth,
		Align:    align,
		Flex:     true,
	})

	// Return table for chaining.
	return t
}

// AddRow adds a data row to the table.
//
// Params:
//   - cells: the cell values for the row, one per column.
//
// Returns:
//   - *Table: the table instance for method chaining.
func (t *Table) AddRow(cells ...string) *Table {
	t.Rows = append(t.Rows, cells)

	// Return table for chaining.
	return t
}

// Render returns the table as a formatted string with header and separator.
//
// Returns:
//   - string: the rendered table with ANSI formatting.
func (t *Table) Render() string {
	// Check if table has columns defined.
	if len(t.Columns) == 0 {
		// Return empty string for empty table.
		return ""
	}

	widths := t.calculateWidths()

	var sb strings.Builder
	estimatedSize := (t.Width + tableEstimatedRowSize) * (len(t.Rows) + tableHeaderRowCount)
	sb.Grow(estimatedSize)

	// Check if header should be rendered.
	if t.ShowHeader {
		t.renderHeader(&sb, widths)
	}

	t.renderRows(&sb, widths)

	// Return trimmed output without trailing newline.
	return strings.TrimSuffix(sb.String(), "\n")
}

// renderHeader writes the header row and separator to the builder.
//
// Params:
//   - sb: the string builder to write to.
//   - widths: the calculated column widths.
func (t *Table) renderHeader(sb *strings.Builder, widths []int) {
	sb.WriteString(t.HeaderColor)

	// Iterate over columns to render header cells.
	for i, col := range t.Columns {
		// Add separator between columns.
		if i > 0 {
			sb.WriteString(t.Separator)
		}
		sb.WriteString(Pad(col.Header, widths[i], col.Align))
	}
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")

	sb.WriteString(t.BorderColor)

	// Iterate over widths to render separator line.
	for i, w := range widths {
		// Add separator between columns.
		if i > 0 {
			sb.WriteString(t.Separator)
		}
		sb.WriteString(HorizontalBar(w))
	}
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")
}

// renderRows writes all data rows to the builder.
//
// Params:
//   - sb: the string builder to write to.
//   - widths: the calculated column widths.
func (t *Table) renderRows(sb *strings.Builder, widths []int) {
	// Iterate over each data row.
	for _, row := range t.Rows {
		// Iterate over columns in the row.
		for i := range t.Columns {
			// Add separator between columns.
			if i > 0 {
				sb.WriteString(t.Separator)
			}

			cell := ""

			// Check if cell value exists in row.
			if i < len(row) {
				cell = row[i]
			}

			sb.WriteString(Pad(cell, widths[i], t.Columns[i].Align))
		}
		sb.WriteString("\n")
	}
}

// RenderCompact returns the table without header separator line.
//
// Returns:
//   - string: the rendered table with ANSI formatting but no separator.
func (t *Table) RenderCompact() string {
	// Check if table has columns defined.
	if len(t.Columns) == 0 {
		// Return empty string for empty table.
		return ""
	}

	widths := t.calculateWidths()
	var sb strings.Builder
	sb.Grow((t.Width + tableCompactExtraSize) * (len(t.Rows) + 1))

	// Check if header should be rendered.
	if t.ShowHeader {
		t.renderCompactHeader(&sb, widths)
	}

	t.renderRows(&sb, widths)

	// Return trimmed output without trailing newline.
	return strings.TrimSuffix(sb.String(), "\n")
}

// renderCompactHeader writes only the header row without separator to the builder.
//
// Params:
//   - sb: the string builder to write to.
//   - widths: the calculated column widths.
func (t *Table) renderCompactHeader(sb *strings.Builder, widths []int) {
	sb.WriteString(t.HeaderColor)

	// Iterate over columns to render header cells.
	for i, col := range t.Columns {
		// Add separator between columns.
		if i > 0 {
			sb.WriteString(t.Separator)
		}
		sb.WriteString(Pad(col.Header, widths[i], col.Align))
	}
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")
}

// calculateWidths determines the width for each column based on content and constraints.
//
// Returns:
//   - []int: the calculated width for each column.
func (t *Table) calculateWidths() []int {
	widths := make([]int, len(t.Columns))
	sepWidth := len(t.Separator) * (len(t.Columns) - 1)
	availableWidth := t.Width - sepWidth

	flexCount, usedWidth := t.calculateInitialWidths(widths)
	t.distributeFlexWidth(widths, flexCount, usedWidth, availableWidth)

	// Return calculated widths.
	return widths
}

// calculateInitialWidths computes initial width for each column.
//
// Params:
//   - widths: slice to store calculated widths.
//
// Returns:
//   - int: number of flex columns.
//   - int: total width used by all columns.
func (t *Table) calculateInitialWidths(widths []int) (int, int) {
	flexCount := 0
	usedWidth := 0

	// Iterate over columns to calculate initial widths.
	for i, col := range t.Columns {
		widths[i] = t.calculateColumnWidth(i, col)

		// Track flex columns for distribution.
		if col.Flex {
			flexCount++
		}
		usedWidth += widths[i]
	}

	// Return flex count and used width.
	return flexCount, usedWidth
}

// calculateColumnWidth computes the width for a single column.
//
// Params:
//   - idx: column index in the table.
//   - col: column definition.
//
// Returns:
//   - int: calculated width for the column.
func (t *Table) calculateColumnWidth(idx int, col Column) int {
	// Handle fixed width columns.
	if col.Width > 0 {
		// Return fixed width.
		return col.Width
	}

	// Handle flexible width columns.
	if col.Flex {
		// Return minimum width for flex.
		return col.MinWidth
	}

	// Handle auto-width columns.
	return t.calculateAutoWidth(idx, col)
}

// calculateAutoWidth computes width based on content for auto columns.
//
// Params:
//   - idx: column index in the table.
//   - col: column definition.
//
// Returns:
//   - int: calculated width based on header and content.
func (t *Table) calculateAutoWidth(idx int, col Column) int {
	width := max(col.MinWidth, VisibleLen(col.Header))

	// Iterate over rows to find max content width.
	for _, row := range t.Rows {
		// Check if row has value for this column.
		if idx < len(row) {
			width = max(width, VisibleLen(row[idx]))
		}
	}

	// Apply max width constraint if specified.
	if col.MaxWidth > 0 && width > col.MaxWidth {
		// Return constrained width.
		return col.MaxWidth
	}

	// Return calculated width.
	return width
}

// distributeFlexWidth distributes remaining space to flex columns.
//
// Params:
//   - widths: slice of column widths to update.
//   - flexCount: number of flex columns.
//   - usedWidth: total width already used.
//   - availableWidth: total available width.
func (t *Table) distributeFlexWidth(widths []int, flexCount, usedWidth, availableWidth int) {
	// Check if flex columns need extra space.
	if flexCount == 0 || usedWidth >= availableWidth {
		// Exit early when no flex columns or no remaining space.
		return
	}

	extra := (availableWidth - usedWidth) / flexCount

	// Iterate over columns to distribute extra width.
	for i, col := range t.Columns {
		// Check if column is flexible.
		if col.Flex {
			widths[i] += extra

			// Apply max width constraint if specified.
			if col.MaxWidth > 0 && widths[i] > col.MaxWidth {
				widths[i] = col.MaxWidth
			}
		}
	}
}
