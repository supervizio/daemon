package widget_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
	"github.com/stretchr/testify/assert"
)

func TestNewTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{"small", 40},
		{"medium", 80},
		{"large", 160},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			table := widget.NewTable(tt.width)
			assert.NotNil(t, table)
			assert.Equal(t, tt.width, table.Width)
			assert.True(t, table.ShowHeader)
		})
	}
}

func TestTable_AddColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		header        string
		width         int
		align         widget.Align
		expectedCols  int
	}{
		{
			name:         "single_column",
			header:       "Name",
			width:        20,
			align:        widget.AlignLeft,
			expectedCols: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			table := widget.NewTable(80)
			result := table.AddColumn(tt.header, tt.width, tt.align)
			assert.Equal(t, table, result)
			assert.Len(t, table.Columns, tt.expectedCols)
		})
	}
}

func TestTable_AddFlexColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		header        string
		minWidth      int
		align         widget.Align
		expectedCols  int
		expectedFlex  bool
	}{
		{
			name:         "flex_column",
			header:       "Name",
			minWidth:     10,
			align:        widget.AlignLeft,
			expectedCols: 1,
			expectedFlex: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			table := widget.NewTable(80)
			result := table.AddFlexColumn(tt.header, tt.minWidth, tt.align)
			assert.Equal(t, table, result)
			assert.Len(t, table.Columns, tt.expectedCols)
			assert.Equal(t, tt.expectedFlex, table.Columns[0].Flex)
		})
	}
}

func TestTable_AddRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		values       []string
		expectedRows int
	}{
		{
			name:         "single_row",
			values:       []string{"Value"},
			expectedRows: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			table := widget.NewTable(80)
			table.AddColumn("Name", 20, widget.AlignLeft)
			result := table.AddRow(tt.values...)
			assert.Equal(t, table, result)
			assert.Len(t, table.Rows, tt.expectedRows)
		})
	}
}

func TestTable_SetHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		show     bool
		wantShow bool
	}{
		{
			name:     "enable header",
			show:     true,
			wantShow: true,
		},
		{
			name:     "disable header",
			show:     false,
			wantShow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table := widget.NewTable(80)
			result := table.SetHeader(tt.show)

			assert.Equal(t, tt.wantShow, table.ShowHeader)
			assert.Same(t, table, result, "SetHeader should return self for chaining")
		})
	}
}

func TestTable_SetHeaderColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		color     string
		wantColor string
	}{
		{
			name:      "set custom color",
			color:     "test-color",
			wantColor: "test-color",
		},
		{
			name:      "set empty color",
			color:     "",
			wantColor: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			table := widget.NewTable(80)
			result := table.SetHeaderColor(tt.color)

			assert.Equal(t, tt.wantColor, table.HeaderColor)
			assert.Same(t, table, result, "SetHeaderColor should return self for chaining")
		})
	}
}

func TestTable_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(*widget.Table)
		expected bool
	}{
		{
			"empty_table",
			func(table *widget.Table) {},
			false,
		},
		{
			"columns_only",
			func(table *widget.Table) {
				table.AddColumn("Name", 20, widget.AlignLeft)
			},
			true,
		},
		{
			"with_rows",
			func(table *widget.Table) {
				table.AddColumn("Name", 20, widget.AlignLeft)
				table.AddRow("Value")
			},
			true,
		},
		{
			"multiple_columns",
			func(table *widget.Table) {
				table.AddColumn("Name", 20, widget.AlignLeft)
				table.AddColumn("Value", 10, widget.AlignRight)
				table.AddRow("Test", "123")
			},
			true,
		},
		{
			"flex_column",
			func(table *widget.Table) {
				table.AddFlexColumn("Name", 10, widget.AlignLeft)
				table.AddRow("Value")
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			table := widget.NewTable(80)
			tt.setup(table)
			result := table.Render()
			if tt.expected {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestTable_RenderCompact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(*widget.Table)
		expected bool
	}{
		{
			"empty_table",
			func(table *widget.Table) {},
			false,
		},
		{
			"with_data",
			func(table *widget.Table) {
				table.AddColumn("Name", 20, widget.AlignLeft)
				table.AddRow("Value")
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			table := widget.NewTable(80)
			tt.setup(table)
			result := table.RenderCompact()
			if tt.expected {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestAlignConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    widget.Align
		expected widget.Align
	}{
		{
			name:     "align_left",
			value:    widget.Align(0),
			expected: widget.AlignLeft,
		},
		{
			name:     "align_right",
			value:    widget.Align(1),
			expected: widget.AlignRight,
		},
		{
			name:     "align_center",
			value:    widget.Align(2),
			expected: widget.AlignCenter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.value)
		})
	}
}
