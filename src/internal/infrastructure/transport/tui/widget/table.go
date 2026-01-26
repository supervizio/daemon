// Package widget provides reusable TUI components.
package widget

import (
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
)

// Column defines a table column.
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

// NewTable creates a new table.
func NewTable(width int) *Table {
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

// AddColumn adds a column definition.
func (t *Table) AddColumn(header string, width int, align Align) *Table {
	t.Columns = append(t.Columns, Column{
		Header:   header,
		Width:    width,
		MinWidth: len(header),
		Align:    align,
	})
	return t
}

// AddFlexColumn adds a flexible-width column.
func (t *Table) AddFlexColumn(header string, minWidth int, align Align) *Table {
	t.Columns = append(t.Columns, Column{
		Header:   header,
		MinWidth: minWidth,
		Align:    align,
		Flex:     true,
	})
	return t
}

// AddRow adds a data row.
func (t *Table) AddRow(cells ...string) *Table {
	t.Rows = append(t.Rows, cells)
	return t
}

// Render returns the table as a string.
func (t *Table) Render() string {
	if len(t.Columns) == 0 {
		return ""
	}

	// Calculate column widths.
	widths := t.calculateWidths()

	var sb strings.Builder
	// Pre-allocate for estimated output size.
	estimatedSize := (t.Width + 20) * (len(t.Rows) + 2) // rows + header + separator
	sb.Grow(estimatedSize)

	// Header.
	if t.ShowHeader {
		sb.WriteString(t.HeaderColor)
		for i, col := range t.Columns {
			if i > 0 {
				sb.WriteString(t.Separator)
			}
			sb.WriteString(Pad(col.Header, widths[i], col.Align))
		}
		sb.WriteString(ansi.Reset)
		sb.WriteString("\n")

		// Header separator.
		sb.WriteString(t.BorderColor)
		for i, w := range widths {
			if i > 0 {
				sb.WriteString(t.Separator)
			}
			sb.WriteString(HorizontalBar(w)) // Use cached horizontal bar.
		}
		sb.WriteString(ansi.Reset)
		sb.WriteString("\n")
	}

	// Data rows.
	for _, row := range t.Rows {
		for i := range t.Columns {
			if i > 0 {
				sb.WriteString(t.Separator)
			}

			cell := ""
			if i < len(row) {
				cell = row[i]
			}

			sb.WriteString(Pad(cell, widths[i], t.Columns[i].Align))
		}
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// RenderCompact returns the table without header separator.
func (t *Table) RenderCompact() string {
	if len(t.Columns) == 0 {
		return ""
	}

	widths := t.calculateWidths()
	var sb strings.Builder
	// Pre-allocate for estimated output size.
	sb.Grow((t.Width + 10) * (len(t.Rows) + 1))

	// Header.
	if t.ShowHeader {
		sb.WriteString(t.HeaderColor)
		for i, col := range t.Columns {
			if i > 0 {
				sb.WriteString(t.Separator)
			}
			sb.WriteString(Pad(col.Header, widths[i], col.Align))
		}
		sb.WriteString(ansi.Reset)
		sb.WriteString("\n")
	}

	// Data rows.
	for _, row := range t.Rows {
		for i := range t.Columns {
			if i > 0 {
				sb.WriteString(t.Separator)
			}

			cell := ""
			if i < len(row) {
				cell = row[i]
			}

			sb.WriteString(Pad(cell, widths[i], t.Columns[i].Align))
		}
		sb.WriteString("\n")
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// calculateWidths determines column widths.
func (t *Table) calculateWidths() []int {
	widths := make([]int, len(t.Columns))
	sepWidth := len(t.Separator) * (len(t.Columns) - 1)
	availableWidth := t.Width - sepWidth

	// First pass: set fixed widths and minimums.
	flexCount := 0
	usedWidth := 0

	for i, col := range t.Columns {
		switch {
		case col.Width > 0:
			widths[i] = col.Width
		case col.Flex:
			widths[i] = col.MinWidth
			flexCount++
		default:
			// Auto: use header width or content max.
			widths[i] = max(col.MinWidth, VisibleLen(col.Header))
			for _, row := range t.Rows {
				if i < len(row) {
					widths[i] = max(widths[i], VisibleLen(row[i]))
				}
			}
			if col.MaxWidth > 0 && widths[i] > col.MaxWidth {
				widths[i] = col.MaxWidth
			}
		}
		usedWidth += widths[i]
	}

	// Second pass: distribute remaining space to flex columns.
	if flexCount > 0 && usedWidth < availableWidth {
		extra := (availableWidth - usedWidth) / flexCount
		for i, col := range t.Columns {
			if col.Flex {
				widths[i] += extra
				if col.MaxWidth > 0 && widths[i] > col.MaxWidth {
					widths[i] = col.MaxWidth
				}
			}
		}
	}

	return widths
}
