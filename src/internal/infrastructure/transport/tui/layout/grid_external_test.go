package layout_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/layout"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/stretchr/testify/assert"
)

func TestNewGrid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		size terminal.Size
	}{
		{"standard", terminal.Size{Cols: 100, Rows: 30}},
		{"compact", terminal.Size{Cols: 60, Rows: 20}},
		{"wide", terminal.Size{Cols: 160, Rows: 40}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(tt.size)
			grid := layout.NewGrid(l)
			assert.NotNil(t, grid)
		})
	}
}

func TestGrid_AddRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		size   terminal.Size
		height int
	}{
		{"fixed_height", terminal.Size{Cols: 100, Rows: 30}, 10},
		{"flexible_height", terminal.Size{Cols: 100, Rows: 30}, 0},
		{"large_height", terminal.Size{Cols: 100, Rows: 30}, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(tt.size)
			grid := layout.NewGrid(l)
			result := grid.AddRow(tt.height)
			assert.Equal(t, grid, result)
		})
	}
}

func TestGrid_AddCell(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		size    terminal.Size
		colspan int
	}{
		{"single_column", terminal.Size{Cols: 100, Rows: 30}, 1},
		{"double_column", terminal.Size{Cols: 100, Rows: 30}, 2},
		{"triple_column", terminal.Size{Cols: 100, Rows: 30}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(tt.size)
			grid := layout.NewGrid(l)
			result := grid.AddRow(10).AddCell(tt.colspan)
			assert.Equal(t, grid, result)
		})
	}
}

func TestGrid_Calculate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func(*layout.Grid)
		expected int
	}{
		{
			"empty_grid",
			func(g *layout.Grid) {},
			0,
		},
		{
			"single_row_single_cell",
			func(g *layout.Grid) {
				g.AddRow(10).AddCell(1)
			},
			1,
		},
		{
			"two_rows",
			func(g *layout.Grid) {
				g.AddRow(10).AddCell(1)
				g.AddRow(10).AddCell(1)
			},
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(terminal.Size{Cols: 100, Rows: 30})
			grid := layout.NewGrid(l)
			tt.setup(grid)
			regions := grid.Calculate()
			if tt.expected == 0 {
				assert.Nil(t, regions)
			} else {
				assert.Len(t, regions, tt.expected)
			}
		})
	}
}
