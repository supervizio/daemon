package layout

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/stretchr/testify/assert"
)

func TestGridRow_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		row        gridRow
		wantHeight int
		wantNil    bool
	}{
		{"with_nil_cells", gridRow{height: 10, cells: nil}, 10, true},
		{"with_empty_cells", gridRow{height: 20, cells: []gridCell{}}, 20, false},
		{"with_cells", gridRow{height: 15, cells: []gridCell{{colspan: 1}}}, 15, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantHeight, tt.row.height)
			if tt.wantNil {
				assert.Nil(t, tt.row.cells)
			} else {
				assert.NotNil(t, tt.row.cells)
			}
		})
	}
}

func TestGridCell_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cell        gridCell
		wantColspan int
	}{
		{"colspan_2", gridCell{colspan: 2}, 2},
		{"colspan_1", gridCell{colspan: 1}, 1},
		{"colspan_3", gridCell{colspan: 3}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantColspan, tt.cell.colspan)
		})
	}
}

// Test_Grid_calculateFlexibleHeight tests the private calculateFlexibleHeight method.
func Test_Grid_calculateFlexibleHeight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		rows           []gridRow
		totalHeight    int
		wantFlexHeight int
	}{
		{
			name:           "single flex row",
			rows:           []gridRow{{height: 0}},
			totalHeight:    100,
			wantFlexHeight: 100,
		},
		{
			name: "two flex rows",
			rows: []gridRow{
				{height: 0},
				{height: 0},
			},
			totalHeight:    100,
			wantFlexHeight: 50,
		},
		{
			name: "mixed fixed and flex",
			rows: []gridRow{
				{height: 20},
				{height: 0},
				{height: 10},
			},
			totalHeight:    100,
			wantFlexHeight: 70,
		},
		{
			name: "multiple flex with fixed",
			rows: []gridRow{
				{height: 15},
				{height: 0},
				{height: 0},
				{height: 5},
			},
			totalHeight:    100,
			wantFlexHeight: 40,
		},
		{
			name: "all fixed rows",
			rows: []gridRow{
				{height: 30},
				{height: 40},
			},
			totalHeight:    100,
			wantFlexHeight: 0,
		},
		{
			name:           "no rows",
			rows:           []gridRow{},
			totalHeight:    100,
			wantFlexHeight: 0,
		},
		{
			name: "flex rows with exact division",
			rows: []gridRow{
				{height: 0},
				{height: 0},
				{height: 0},
				{height: 0},
			},
			totalHeight:    80,
			wantFlexHeight: 20,
		},
		{
			name: "flex rows with remainder",
			rows: []gridRow{
				{height: 0},
				{height: 0},
				{height: 0},
			},
			totalHeight:    100,
			wantFlexHeight: 33,
		},
		{
			name: "negative height treated as flex",
			rows: []gridRow{
				{height: -1},
				{height: 10},
			},
			totalHeight:    100,
			wantFlexHeight: 90,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a minimal layout for testing.
			l := &Layout{
				Size:    terminal.Size{Cols: 80, Rows: 24},
				Mode:    terminal.LayoutNormal,
				Padding: 1,
				Gap:     2,
			}

			g := &Grid{
				layout: l,
				rows:   tt.rows,
			}

			got := g.calculateFlexibleHeight(tt.totalHeight)
			assert.Equal(t, tt.wantFlexHeight, got)
		})
	}
}

// Test_Grid_getRowHeight tests the private getRowHeight method.
func Test_Grid_getRowHeight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		row        gridRow
		flexHeight int
		wantHeight int
	}{
		{
			name:       "flexible row with zero height",
			row:        gridRow{height: 0},
			flexHeight: 50,
			wantHeight: 50,
		},
		{
			name:       "fixed row",
			row:        gridRow{height: 20},
			flexHeight: 50,
			wantHeight: 20,
		},
		{
			name:       "negative height treated as flexible",
			row:        gridRow{height: -5},
			flexHeight: 40,
			wantHeight: 40,
		},
		{
			name:       "large fixed height",
			row:        gridRow{height: 100},
			flexHeight: 20,
			wantHeight: 100,
		},
		{
			name:       "flexible with zero flexHeight",
			row:        gridRow{height: 0},
			flexHeight: 0,
			wantHeight: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create minimal grid for testing.
			g := &Grid{
				layout: &Layout{},
			}

			got := g.getRowHeight(tt.row, tt.flexHeight)
			assert.Equal(t, tt.wantHeight, got)
		})
	}
}

// Test_Grid_generateRowRegions tests the private generateRowRegions method.
func Test_Grid_generateRowRegions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		row         gridRow
		params      rowRegionParams
		wantRegions []Region
	}{
		{
			name: "single cell full width",
			row: gridRow{
				height: 10,
				cells:  []gridCell{{colspan: 1}},
			},
			params: rowRegionParams{
				startX:   0,
				yPos:     0,
				height:   10,
				colWidth: 80,
				cols:     1,
				gap:      2,
			},
			wantRegions: []Region{
				{X: 0, Y: 0, Width: 80, Height: 10},
			},
		},
		{
			name: "two cells with colspan 1",
			row: gridRow{
				height: 20,
				cells: []gridCell{
					{colspan: 1},
					{colspan: 1},
				},
			},
			params: rowRegionParams{
				startX:   0,
				yPos:     5,
				height:   20,
				colWidth: 40,
				cols:     2,
				gap:      2,
			},
			wantRegions: []Region{
				{X: 0, Y: 5, Width: 40, Height: 20},
				{X: 42, Y: 5, Width: 40, Height: 20}, // 40 + 2 gap
			},
		},
		{
			name: "cell with colspan 2",
			row: gridRow{
				height: 15,
				cells:  []gridCell{{colspan: 2}},
			},
			params: rowRegionParams{
				startX:   10,
				yPos:     10,
				height:   15,
				colWidth: 30,
				cols:     2,
				gap:      2,
			},
			wantRegions: []Region{
				{X: 10, Y: 10, Width: 62, Height: 15}, // 30*2 + 2 gap
			},
		},
		{
			name: "zero colspan defaults to 1",
			row: gridRow{
				height: 10,
				cells:  []gridCell{{colspan: 0}},
			},
			params: rowRegionParams{
				startX:   0,
				yPos:     0,
				height:   10,
				colWidth: 50,
				cols:     2,
				gap:      2,
			},
			wantRegions: []Region{
				{X: 0, Y: 0, Width: 50, Height: 10},
			},
		},
		{
			name: "negative colspan defaults to 1",
			row: gridRow{
				height: 10,
				cells:  []gridCell{{colspan: -1}},
			},
			params: rowRegionParams{
				startX:   0,
				yPos:     0,
				height:   10,
				colWidth: 50,
				cols:     2,
				gap:      2,
			},
			wantRegions: []Region{
				{X: 0, Y: 0, Width: 50, Height: 10},
			},
		},
		{
			name: "colspan exceeds columns",
			row: gridRow{
				height: 10,
				cells:  []gridCell{{colspan: 5}},
			},
			params: rowRegionParams{
				startX:   0,
				yPos:     0,
				height:   10,
				colWidth: 30,
				cols:     2,
				gap:      2,
			},
			wantRegions: []Region{
				{X: 0, Y: 0, Width: 62, Height: 10}, // Limited to 2 cols: 30*2 + 2
			},
		},
		{
			name: "three cells with gaps",
			row: gridRow{
				height: 25,
				cells: []gridCell{
					{colspan: 1},
					{colspan: 1},
					{colspan: 1},
				},
			},
			params: rowRegionParams{
				startX:   5,
				yPos:     15,
				height:   25,
				colWidth: 20,
				cols:     3,
				gap:      3,
			},
			wantRegions: []Region{
				{X: 5, Y: 15, Width: 20, Height: 25},
				{X: 28, Y: 15, Width: 20, Height: 25},  // 5 + 20 + 3
				{X: 51, Y: 15, Width: 20, Height: 25},  // 28 + 20 + 3
			},
		},
		{
			name: "empty row",
			row: gridRow{
				height: 10,
				cells:  []gridCell{},
			},
			params: rowRegionParams{
				startX:   0,
				yPos:     0,
				height:   10,
				colWidth: 50,
				cols:     2,
				gap:      2,
			},
			wantRegions: []Region{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create minimal grid for testing.
			g := &Grid{
				layout: &Layout{},
			}

			got := g.generateRowRegions(tt.row, tt.params)
			assert.Equal(t, tt.wantRegions, got)
		})
	}
}

// Test_Grid_generateAllRegions tests the private generateAllRegions method.
func Test_Grid_generateAllRegions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		rows        []gridRow
		content     Region
		flexHeight  int
		wantRegions [][]Region
	}{
		{
			name: "single row single cell",
			rows: []gridRow{
				{
					height: 10,
					cells:  []gridCell{{colspan: 1}},
				},
			},
			content: Region{
				X:      1,
				Y:      1,
				Width:  78,
				Height: 22,
			},
			flexHeight: 0,
			wantRegions: [][]Region{
				{
					{X: 1, Y: 1, Width: 78, Height: 10},
				},
			},
		},
		{
			name: "two rows fixed heights",
			rows: []gridRow{
				{
					height: 5,
					cells:  []gridCell{{colspan: 1}},
				},
				{
					height: 10,
					cells:  []gridCell{{colspan: 1}},
				},
			},
			content: Region{
				X:      0,
				Y:      0,
				Width:  78,
				Height: 24,
			},
			flexHeight: 0,
			wantRegions: [][]Region{
				{
					{X: 0, Y: 0, Width: 78, Height: 5},
				},
				{
					{X: 0, Y: 5, Width: 78, Height: 10},
				},
			},
		},
		{
			name: "flexible row",
			rows: []gridRow{
				{
					height: 0,
					cells:  []gridCell{{colspan: 1}},
				},
			},
			content: Region{
				X:      1,
				Y:      1,
				Width:  78,
				Height: 22,
			},
			flexHeight: 22,
			wantRegions: [][]Region{
				{
					{X: 1, Y: 1, Width: 78, Height: 22},
				},
			},
		},
		{
			name: "mixed fixed and flexible rows",
			rows: []gridRow{
				{
					height: 5,
					cells:  []gridCell{{colspan: 1}},
				},
				{
					height: 0,
					cells:  []gridCell{{colspan: 1}},
				},
			},
			content: Region{
				X:      0,
				Y:      0,
				Width:  78,
				Height: 25,
			},
			flexHeight: 20,
			wantRegions: [][]Region{
				{
					{X: 0, Y: 0, Width: 78, Height: 5},
				},
				{
					{X: 0, Y: 5, Width: 78, Height: 20},
				},
			},
		},
		{
			name: "two columns layout",
			rows: []gridRow{
				{
					height: 10,
					cells: []gridCell{
						{colspan: 1},
						{colspan: 1},
					},
				},
			},
			content: Region{
				X:      1,
				Y:      1,
				Width:  156,
				Height: 38,
			},
			flexHeight: 0,
			wantRegions: [][]Region{
				{
					{X: 1, Y: 1, Width: 78, Height: 10},
					{X: 81, Y: 1, Width: 78, Height: 10},
				},
			},
		},
		{
			name:        "empty grid",
			rows:        []gridRow{},
			content:     Region{X: 0, Y: 0, Width: 78, Height: 24},
			flexHeight:  0,
			wantRegions: [][]Region{},
		},
		{
			name: "three rows stacked",
			rows: []gridRow{
				{height: 3, cells: []gridCell{{colspan: 1}}},
				{height: 5, cells: []gridCell{{colspan: 1}}},
				{height: 7, cells: []gridCell{{colspan: 1}}},
			},
			content: Region{
				X:      2,
				Y:      2,
				Width:  78,
				Height: 20,
			},
			flexHeight: 0,
			wantRegions: [][]Region{
				{{X: 2, Y: 2, Width: 78, Height: 3}},
				{{X: 2, Y: 5, Width: 78, Height: 5}},
				{{X: 2, Y: 10, Width: 78, Height: 7}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create layout and grid.
			l := &Layout{
				Size:    terminal.Size{Cols: 160, Rows: 40},
				Mode:    terminal.LayoutWide,
				Padding: 1,
				Gap:     2,
			}

			g := &Grid{
				layout: l,
				rows:   tt.rows,
			}

			got := g.generateAllRegions(tt.content, tt.flexHeight)
			assert.Equal(t, tt.wantRegions, got)
		})
	}
}
