package widget

import (
	"strings"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/stretchr/testify/assert"
)

func TestTableConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    int
		positive bool
	}{
		{
			name:     "tableEstimatedRowSize",
			value:    tableEstimatedRowSize,
			positive: true,
		},
		{
			name:     "tableHeaderRowCount",
			value:    tableHeaderRowCount,
			positive: true,
		},
		{
			name:     "tableCompactExtraSize",
			value:    tableCompactExtraSize,
			positive: true,
		},
		{
			name:     "defaultColumnCapacity",
			value:    defaultColumnCapacity,
			positive: true,
		},
		{
			name:     "defaultRowCapacity",
			value:    defaultRowCapacity,
			positive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.positive {
				assert.Greater(t, tt.value, 0)
			}
		})
	}
}

func Test_Table_renderHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func() (*Table, []int)
		want    string
		wantLen int
	}{
		{
			name: "single column",
			setup: func() (*Table, []int) {
				table := NewTable(20)
				table.AddColumn("Name", 10, AlignLeft)
				return table, []int{10}
			},
			wantLen: 10,
		},
		{
			name: "multiple columns with separator",
			setup: func() (*Table, []int) {
				table := NewTable(50)
				table.AddColumn("ID", 5, AlignLeft)
				table.AddColumn("Name", 15, AlignLeft)
				table.AddColumn("Status", 10, AlignRight)
				return table, []int{5, 15, 10}
			},
			wantLen: 34, // 5 + 2 + 15 + 2 + 10 = 34
		},
		{
			name: "empty columns",
			setup: func() (*Table, []int) {
				table := NewTable(20)
				return table, []int{}
			},
			want: "",
		},
		{
			name: "custom separator",
			setup: func() (*Table, []int) {
				table := NewTable(30)
				table.Separator = " | "
				table.AddColumn("A", 5, AlignLeft)
				table.AddColumn("B", 5, AlignLeft)
				return table, []int{5, 5}
			},
			wantLen: 13, // 5 + 3 + 5 = 13
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table, widths := tt.setup()
			result := table.renderHeader(widths)

			// Check for ANSI codes
			if len(table.Columns) > 0 {
				assert.Contains(t, result, ansi.Reset, "Header should contain reset code")
			}

			if tt.want != "" {
				assert.Equal(t, tt.want, result)
			}

			// Check length (accounting for ANSI codes)
			if tt.wantLen > 0 {
				stripped := stripAnsiCodes(result)
				assert.Equal(t, tt.wantLen, len(stripped), "Header length should match expected")
			}
		})
	}
}

func Test_Table_renderRows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func() (*Table, []int)
		wantRows  int
		wantEmpty bool
	}{
		{
			name: "no rows",
			setup: func() (*Table, []int) {
				table := NewTable(30)
				table.AddColumn("Name", 10, AlignLeft)
				return table, []int{10}
			},
			wantEmpty: true,
		},
		{
			name: "single row",
			setup: func() (*Table, []int) {
				table := NewTable(30)
				table.AddColumn("Name", 10, AlignLeft)
				table.AddRow("Alice")
				return table, []int{10}
			},
			wantRows: 1,
		},
		{
			name: "multiple rows",
			setup: func() (*Table, []int) {
				table := NewTable(50)
				table.AddColumn("ID", 5, AlignLeft)
				table.AddColumn("Name", 15, AlignLeft)
				table.AddRow("1", "Alice")
				table.AddRow("2", "Bob")
				table.AddRow("3", "Charlie")
				return table, []int{5, 15}
			},
			wantRows: 3,
		},
		{
			name: "rows with colors",
			setup: func() (*Table, []int) {
				table := NewTable(30)
				table.AddColumn("Name", 10, AlignLeft)
				table.RowColors = []string{ansi.FgWhite, ansi.FgGray}
				table.AddRow("Alice")
				table.AddRow("Bob")
				return table, []int{10}
			},
			wantRows: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table, widths := tt.setup()
			result := table.renderRows(widths)

			if tt.wantEmpty {
				assert.Empty(t, result, "Should return empty string for no rows")
			} else {
				lines := strings.Split(result, "\n")
				assert.Len(t, lines, tt.wantRows, "Should have expected number of rows")

				// Check for ANSI reset if row colors are used
				if len(table.RowColors) > 0 {
					assert.Contains(t, result, ansi.Reset, "Should contain reset codes")
				}
			}
		})
	}
}

func Test_Table_calculateWidths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Table
		want  []int
	}{
		{
			name: "single fixed column",
			setup: func() *Table {
				table := NewTable(80)
				table.AddColumn("Name", 20, AlignLeft)
				return table
			},
			want: []int{20},
		},
		{
			name: "multiple fixed columns",
			setup: func() *Table {
				table := NewTable(80)
				table.AddColumn("ID", 5, AlignLeft)
				table.AddColumn("Name", 20, AlignLeft)
				table.AddColumn("Status", 10, AlignLeft)
				return table
			},
			want: []int{5, 20, 10},
		},
		{
			name: "single flex column",
			setup: func() *Table {
				table := NewTable(80)
				table.AddFlexColumn("Description", 10, AlignLeft)
				return table
			},
			want: []int{80}, // Full width for flex column
		},
		{
			name: "fixed and flex columns",
			setup: func() *Table {
				table := NewTable(100)
				table.AddColumn("ID", 5, AlignLeft)
				table.AddColumn("Name", 20, AlignLeft)
				table.AddFlexColumn("Description", 10, AlignLeft)
				// Total: 5 + 20 + 10 = 35
				// Separator: 2 * 2 = 4
				// Remaining: 100 - 35 - 4 = 61
				// Flex gets: 10 + 61 = 71
				return table
			},
			want: []int{5, 20, 71},
		},
		{
			name: "multiple flex columns share remaining",
			setup: func() *Table {
				table := NewTable(100)
				table.AddColumn("ID", 10, AlignLeft)
				table.AddFlexColumn("Col1", 5, AlignLeft)
				table.AddFlexColumn("Col2", 5, AlignLeft)
				// Total: 10 + 5 + 5 = 20
				// Separator: 2 * 2 = 4
				// Remaining: 100 - 20 - 4 = 76
				// Each flex gets: 76 / 2 = 38
				// Final: 5 + 38 = 43 each
				return table
			},
			want: []int{10, 43, 43},
		},
		{
			name: "auto width column",
			setup: func() *Table {
				table := NewTable(80)
				table.AddColumn("Name", 0, AlignLeft) // Width 0 means auto
				table.AddRow("Alice")
				table.AddRow("Bob")
				return table
			},
			want: []int{5}, // Max of header (4) and content (5)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table := tt.setup()
			result := table.calculateWidths()

			assert.Equal(t, tt.want, result, "Column widths should match expected")
		})
	}
}

func Test_Table_calculateInitialWidths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Table
		want  []int
	}{
		{
			name: "empty table",
			setup: func() *Table {
				return NewTable(80)
			},
			want: []int{},
		},
		{
			name: "fixed width columns",
			setup: func() *Table {
				table := NewTable(80)
				table.AddColumn("A", 10, AlignLeft)
				table.AddColumn("B", 20, AlignLeft)
				return table
			},
			want: []int{10, 20},
		},
		{
			name: "minimum width columns",
			setup: func() *Table {
				table := NewTable(80)
				table.AddFlexColumn("Flex", 15, AlignLeft)
				return table
			},
			want: []int{15},
		},
		{
			name: "auto width columns",
			setup: func() *Table {
				table := NewTable(80)
				table.AddColumn("Short", 0, AlignLeft)
				table.AddRow("LongContent")
				return table
			},
			want: []int{11}, // Max of "Short"(5) and "LongContent"(11)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table := tt.setup()
			result := table.calculateInitialWidths()

			assert.Equal(t, tt.want, result, "Initial widths should match expected")
		})
	}
}

func Test_Table_calculateColumnWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() (*Table, Column)
		want  int
	}{
		{
			name: "fixed width",
			setup: func() (*Table, Column) {
				table := NewTable(80)
				col := Column{Header: "Name", Width: 25, Align: AlignLeft}
				return table, col
			},
			want: 25,
		},
		{
			name: "minimum width",
			setup: func() (*Table, Column) {
				table := NewTable(80)
				col := Column{Header: "Name", MinWidth: 15, Align: AlignLeft}
				return table, col
			},
			want: 15,
		},
		{
			name: "auto width - header longer",
			setup: func() (*Table, Column) {
				table := NewTable(80)
				col := Column{Header: "LongHeaderName", Width: 0, Align: AlignLeft}
				table.Columns = append(table.Columns, col)
				table.AddRow("X")
				return table, col
			},
			want: 14, // Length of "LongHeaderName"
		},
		{
			name: "auto width - content longer",
			setup: func() (*Table, Column) {
				table := NewTable(80)
				col := Column{Header: "ID", Width: 0, Align: AlignLeft}
				table.Columns = append(table.Columns, col)
				table.AddRow("VeryLongValue")
				return table, col
			},
			want: 13, // Length of "VeryLongValue"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table, col := tt.setup()
			result := table.calculateColumnWidth(col)

			assert.Equal(t, tt.want, result, "Column width should match expected")
		})
	}
}

func Test_Table_calculateAutoWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() (*Table, Column)
		want  int
	}{
		{
			name: "empty table - header only",
			setup: func() (*Table, Column) {
				table := NewTable(80)
				col := Column{Header: "Status", Align: AlignLeft}
				table.Columns = append(table.Columns, col)
				return table, col
			},
			want: 6, // Length of "Status"
		},
		{
			name: "header longer than content",
			setup: func() (*Table, Column) {
				table := NewTable(80)
				col := Column{Header: "VeryLongHeader", Align: AlignLeft}
				table.Columns = append(table.Columns, col)
				table.AddRow("A")
				return table, col
			},
			want: 14, // Length of "VeryLongHeader"
		},
		{
			name: "content longer than header",
			setup: func() (*Table, Column) {
				table := NewTable(80)
				col := Column{Header: "ID", Align: AlignLeft}
				table.Columns = append(table.Columns, col)
				table.AddRow("VeryLongContent")
				return table, col
			},
			want: 15, // Length of "VeryLongContent"
		},
		{
			name: "multiple rows - find max",
			setup: func() (*Table, Column) {
				table := NewTable(80)
				col := Column{Header: "Name", Align: AlignLeft}
				table.Columns = append(table.Columns, col)
				table.AddRow("Alice")
				table.AddRow("BobBobBobBob")
				table.AddRow("Charlie")
				return table, col
			},
			want: 12, // Length of "BobBobBobBob"
		},
		{
			name: "column not in table",
			setup: func() (*Table, Column) {
				table := NewTable(80)
				table.AddColumn("Other", 10, AlignLeft)
				col := Column{Header: "NotInTable", Align: AlignLeft}
				return table, col
			},
			want: 10, // Length of header "NotInTable"
		},
		{
			name: "second column in multi-column table",
			setup: func() (*Table, Column) {
				table := NewTable(80)
				table.AddColumn("ID", 5, AlignLeft)
				col := Column{Header: "Name", Align: AlignLeft}
				table.Columns = append(table.Columns, col)
				table.AddRow("1", "Alice")
				table.AddRow("2", "SuperLongName")
				return table, col
			},
			want: 13, // Length of "SuperLongName"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table, col := tt.setup()
			result := table.calculateAutoWidth(col)

			assert.Equal(t, tt.want, result, "Auto width should match expected")
		})
	}
}

func Test_Table_distributeFlexWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func() *Table
		widths    []int
		remaining int
		want      []int
	}{
		{
			name: "no flex columns",
			setup: func() *Table {
				table := NewTable(80)
				table.AddColumn("A", 10, AlignLeft)
				table.AddColumn("B", 20, AlignLeft)
				return table
			},
			widths:    []int{10, 20},
			remaining: 50,
			want:      []int{10, 20}, // No change
		},
		{
			name: "single flex column",
			setup: func() *Table {
				table := NewTable(100)
				table.AddColumn("ID", 10, AlignLeft)
				table.AddFlexColumn("Description", 20, AlignLeft)
				return table
			},
			widths:    []int{10, 20},
			remaining: 50,
			want:      []int{10, 70}, // 20 + 50 = 70
		},
		{
			name: "multiple flex columns - equal distribution",
			setup: func() *Table {
				table := NewTable(100)
				table.AddFlexColumn("Col1", 10, AlignLeft)
				table.AddFlexColumn("Col2", 10, AlignLeft)
				return table
			},
			widths:    []int{10, 10},
			remaining: 60,
			want:      []int{40, 40}, // Each gets 30 extra (60/2)
		},
		{
			name: "mixed fixed and flex columns",
			setup: func() *Table {
				table := NewTable(100)
				table.AddColumn("Fixed1", 15, AlignLeft)
				table.AddFlexColumn("Flex1", 10, AlignLeft)
				table.AddColumn("Fixed2", 15, AlignLeft)
				table.AddFlexColumn("Flex2", 10, AlignLeft)
				return table
			},
			widths:    []int{15, 10, 15, 10},
			remaining: 40,
			want:      []int{15, 30, 15, 30}, // Each flex gets 20 (40/2)
		},
		{
			name: "no remaining width",
			setup: func() *Table {
				table := NewTable(50)
				table.AddFlexColumn("Col", 10, AlignLeft)
				return table
			},
			widths:    []int{10},
			remaining: 0,
			want:      []int{10}, // No change
		},
		{
			name: "odd remaining width",
			setup: func() *Table {
				table := NewTable(100)
				table.AddFlexColumn("A", 10, AlignLeft)
				table.AddFlexColumn("B", 10, AlignLeft)
				table.AddFlexColumn("C", 10, AlignLeft)
				return table
			},
			widths:    []int{10, 10, 10},
			remaining: 10,
			want:      []int{13, 13, 13}, // 10/3 = 3 each (integer division)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table := tt.setup()
			result := table.distributeFlexWidth(tt.widths, tt.remaining)

			assert.Equal(t, tt.want, result, "Distributed widths should match expected")
		})
	}
}

func Test_Table_calculateWidths_edgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func() *Table
		want  []int
	}{
		{
			name: "empty table",
			setup: func() *Table {
				return NewTable(80)
			},
			want: []int{},
		},
		{
			name: "wide table exceeds total width",
			setup: func() *Table {
				table := NewTable(50)
				table.AddColumn("A", 30, AlignLeft)
				table.AddColumn("B", 30, AlignLeft)
				table.AddColumn("C", 30, AlignLeft)
				return table
			},
			want: []int{30, 30, 30}, // No distribution when already exceeding
		},
		{
			name: "table exactly fits width",
			setup: func() *Table {
				table := NewTable(38) // 10 + 2 + 15 + 2 + 10 = 39 (with separator "  ")
				table.AddColumn("A", 10, AlignLeft)
				table.AddColumn("B", 15, AlignLeft)
				table.AddColumn("C", 10, AlignLeft)
				return table
			},
			want: []int{10, 15, 10},
		},
		{
			name: "single column table",
			setup: func() *Table {
				table := NewTable(80)
				table.AddColumn("OnlyColumn", 20, AlignLeft)
				return table
			},
			want: []int{20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table := tt.setup()
			result := table.calculateWidths()

			assert.Equal(t, tt.want, result, "Edge case widths should match expected")
		})
	}
}

func Test_Table_renderSingleRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func() (*Table, []string, int, []int)
		wantColor bool
		wantReset bool
	}{
		{
			name: "row without colors",
			setup: func() (*Table, []string, int, []int) {
				table := NewTable(50)
				table.AddColumn("Name", 10, AlignLeft)
				table.AddColumn("Value", 10, AlignRight)
				row := []string{"Alice", "100"}
				widths := []int{10, 10}
				return table, row, 0, widths
			},
			wantColor: false,
			wantReset: false,
		},
		{
			name: "row with alternating colors - even index",
			setup: func() (*Table, []string, int, []int) {
				table := NewTable(50)
				table.AddColumn("Name", 10, AlignLeft)
				table.RowColors = []string{ansi.FgWhite, ansi.FgGray}
				row := []string{"Alice"}
				widths := []int{10}
				return table, row, 0, widths
			},
			wantColor: true,
			wantReset: true,
		},
		{
			name: "row with alternating colors - odd index",
			setup: func() (*Table, []string, int, []int) {
				table := NewTable(50)
				table.AddColumn("Name", 10, AlignLeft)
				table.RowColors = []string{ansi.FgWhite, ansi.FgGray}
				row := []string{"Bob"}
				widths := []int{10}
				return table, row, 1, widths
			},
			wantColor: true,
			wantReset: true,
		},
		{
			name: "empty row",
			setup: func() (*Table, []string, int, []int) {
				table := NewTable(50)
				table.AddColumn("Name", 10, AlignLeft)
				row := []string{}
				widths := []int{10}
				return table, row, 0, widths
			},
			wantColor: false,
			wantReset: false,
		},
		{
			name: "single color in RowColors",
			setup: func() (*Table, []string, int, []int) {
				table := NewTable(50)
				table.AddColumn("Name", 10, AlignLeft)
				table.RowColors = []string{ansi.FgCyan}
				row := []string{"Test"}
				widths := []int{10}
				return table, row, 5, widths
			},
			wantColor: true,
			wantReset: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table, row, rowIdx, widths := tt.setup()
			var sb strings.Builder
			table.renderSingleRow(&sb, row, rowIdx, widths)
			result := sb.String()

			if tt.wantColor {
				// Verify color is applied from RowColors
				colorIdx := rowIdx % len(table.RowColors)
				assert.Contains(t, result, table.RowColors[colorIdx], "Should contain row color")
			}

			if tt.wantReset {
				assert.Contains(t, result, ansi.Reset, "Should contain reset code")
			} else if len(table.RowColors) == 0 {
				assert.NotContains(t, result, ansi.Reset, "Should not contain reset code without RowColors")
			}
		})
	}
}

func Test_Table_renderRowCells(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setup        func() (*Table, []string, []int)
		wantContains []string
		wantLen      int
	}{
		{
			name: "single cell",
			setup: func() (*Table, []string, []int) {
				table := NewTable(50)
				table.AddColumn("Name", 10, AlignLeft)
				row := []string{"Alice"}
				widths := []int{10}
				return table, row, widths
			},
			wantContains: []string{"Alice"},
			wantLen:      10,
		},
		{
			name: "multiple cells with separator",
			setup: func() (*Table, []string, []int) {
				table := NewTable(50)
				table.AddColumn("ID", 5, AlignLeft)
				table.AddColumn("Name", 10, AlignLeft)
				table.AddColumn("Status", 8, AlignRight)
				row := []string{"1", "Alice", "Active"}
				widths := []int{5, 10, 8}
				return table, row, widths
			},
			wantContains: []string{"1", "Alice", "Active"},
			wantLen:      27, // 5 + 2 + 10 + 2 + 8 = 27
		},
		{
			name: "empty row",
			setup: func() (*Table, []string, []int) {
				table := NewTable(50)
				table.AddColumn("Name", 10, AlignLeft)
				row := []string{}
				widths := []int{10}
				return table, row, widths
			},
			wantContains: []string{},
			wantLen:      0,
		},
		{
			name: "row with fewer cells than columns",
			setup: func() (*Table, []string, []int) {
				table := NewTable(50)
				table.AddColumn("A", 5, AlignLeft)
				table.AddColumn("B", 5, AlignLeft)
				table.AddColumn("C", 5, AlignLeft)
				row := []string{"X", "Y"} // Only 2 cells for 3 columns
				widths := []int{5, 5, 5}
				return table, row, widths
			},
			wantContains: []string{"X", "Y"},
			wantLen:      12, // 5 + 2 + 5 = 12 (only 2 cells rendered)
		},
		{
			name: "custom separator",
			setup: func() (*Table, []string, []int) {
				table := NewTable(50)
				table.Separator = " | "
				table.AddColumn("A", 5, AlignLeft)
				table.AddColumn("B", 5, AlignLeft)
				row := []string{"X", "Y"}
				widths := []int{5, 5}
				return table, row, widths
			},
			wantContains: []string{"X", " | ", "Y"},
			wantLen:      13, // 5 + 3 + 5 = 13
		},
		{
			name: "right aligned cell",
			setup: func() (*Table, []string, []int) {
				table := NewTable(50)
				table.AddColumn("Value", 10, AlignRight)
				row := []string{"42"}
				widths := []int{10}
				return table, row, widths
			},
			wantContains: []string{"42"},
			wantLen:      10,
		},
		{
			name: "center aligned cell",
			setup: func() (*Table, []string, []int) {
				table := NewTable(50)
				table.AddColumn("Title", 10, AlignCenter)
				row := []string{"Hi"}
				widths := []int{10}
				return table, row, widths
			},
			wantContains: []string{"Hi"},
			wantLen:      10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table, row, widths := tt.setup()
			var sb strings.Builder
			table.renderRowCells(&sb, row, widths)
			result := sb.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want, "Result should contain expected content")
			}

			if tt.wantLen > 0 {
				assert.Len(t, result, tt.wantLen, "Result length should match expected")
			}
		})
	}
}

// stripAnsiCodes removes ANSI escape sequences from a string for length testing.
func stripAnsiCodes(s string) string {
	var result strings.Builder
	result.Grow(len(s))

	inEscape := false
	for i := range len(s) {
		if s[i] == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}

	return result.String()
}
