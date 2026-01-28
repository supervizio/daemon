package layout_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/layout"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/stretchr/testify/assert"
)

func TestNewLayout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		size terminal.Size
	}{
		{"small_terminal", terminal.Size{Cols: 60, Rows: 20}},
		{"standard_terminal", terminal.Size{Cols: 80, Rows: 24}},
		{"wide_terminal", terminal.Size{Cols: 140, Rows: 40}},
		{"ultrawide_terminal", terminal.Size{Cols: 200, Rows: 50}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(tt.size)
			assert.NotNil(t, l)
			assert.Equal(t, tt.size, l.Size)
		})
	}
}

func TestLayout_ContentWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     terminal.Size
		minWidth int
	}{
		{"small_terminal", terminal.Size{Cols: 60, Rows: 20}, 1},
		{"standard_terminal", terminal.Size{Cols: 80, Rows: 24}, 1},
		{"tiny_terminal", terminal.Size{Cols: 2, Rows: 2}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(tt.size)
			width := l.ContentWidth()
			assert.GreaterOrEqual(t, width, tt.minWidth)
		})
	}
}

func TestLayout_ContentHeight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		size      terminal.Size
		minHeight int
	}{
		{"small_terminal", terminal.Size{Cols: 60, Rows: 20}, 1},
		{"standard_terminal", terminal.Size{Cols: 80, Rows: 24}, 1},
		{"tiny_terminal", terminal.Size{Cols: 2, Rows: 2}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(tt.size)
			height := l.ContentHeight()
			assert.GreaterOrEqual(t, height, tt.minHeight)
		})
	}
}

func TestLayout_Columns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		size         terminal.Size
		expectedCols int
	}{
		{"compact", terminal.Size{Cols: 60, Rows: 20}, 1},
		{"normal", terminal.Size{Cols: 100, Rows: 24}, 1},
		{"wide", terminal.Size{Cols: 140, Rows: 40}, 2},
		{"ultrawide", terminal.Size{Cols: 200, Rows: 50}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(tt.size)
			cols := l.Columns()
			assert.Equal(t, tt.expectedCols, cols)
		})
	}
}

func TestLayout_ColumnWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     terminal.Size
		minWidth int
	}{
		{"compact", terminal.Size{Cols: 60, Rows: 20}, 1},
		{"normal", terminal.Size{Cols: 100, Rows: 24}, 1},
		{"wide", terminal.Size{Cols: 140, Rows: 40}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(tt.size)
			width := l.ColumnWidth()
			assert.GreaterOrEqual(t, width, tt.minWidth)
		})
	}
}

func TestLayout_ColumnRegion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		size   terminal.Size
		column int
	}{
		{"first_column", terminal.Size{Cols: 140, Rows: 40}, 0},
		{"second_column", terminal.Size{Cols: 140, Rows: 40}, 1},
		{"invalid_negative", terminal.Size{Cols: 140, Rows: 40}, -1},
		{"invalid_too_large", terminal.Size{Cols: 140, Rows: 40}, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(tt.size)
			region := l.ColumnRegion(tt.column)
			assert.GreaterOrEqual(t, region.Width, 1)
			assert.GreaterOrEqual(t, region.Height, 1)
		})
	}
}

func TestLayout_FullWidthRegion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		size terminal.Size
	}{
		{"small", terminal.Size{Cols: 60, Rows: 20}},
		{"standard", terminal.Size{Cols: 80, Rows: 24}},
		{"wide", terminal.Size{Cols: 140, Rows: 40}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := layout.NewLayout(tt.size)
			region := l.FullWidthRegion()
			assert.GreaterOrEqual(t, region.Width, 1)
			assert.GreaterOrEqual(t, region.Height, 1)
		})
	}
}

func TestSplitHorizontal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		region   layout.Region
		n        int
		expected int
	}{
		{"two_sections", layout.Region{X: 0, Y: 0, Width: 80, Height: 40}, 2, 2},
		{"three_sections", layout.Region{X: 0, Y: 0, Width: 80, Height: 30}, 3, 3},
		{"single_section", layout.Region{X: 0, Y: 0, Width: 80, Height: 20}, 1, 1},
		{"zero_sections", layout.Region{X: 0, Y: 0, Width: 80, Height: 20}, 0, 0},
		{"negative_sections", layout.Region{X: 0, Y: 0, Width: 80, Height: 20}, -1, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			regions := layout.SplitHorizontal(tt.region, tt.n)
			if tt.expected == 0 {
				assert.Nil(t, regions)
			} else {
				assert.Len(t, regions, tt.expected)
			}
		})
	}
}

func TestSplitVertical(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		region   layout.Region
		n        int
		gap      int
		expected int
	}{
		{"two_sections", layout.Region{X: 0, Y: 0, Width: 80, Height: 40}, 2, 2, 2},
		{"three_sections", layout.Region{X: 0, Y: 0, Width: 120, Height: 30}, 3, 2, 3},
		{"single_section", layout.Region{X: 0, Y: 0, Width: 80, Height: 20}, 1, 2, 1},
		{"zero_sections", layout.Region{X: 0, Y: 0, Width: 80, Height: 20}, 0, 2, 0},
		{"negative_sections", layout.Region{X: 0, Y: 0, Width: 80, Height: 20}, -1, 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			regions := layout.SplitVertical(tt.region, tt.n, tt.gap)
			if tt.expected == 0 {
				assert.Nil(t, regions)
			} else {
				assert.Len(t, regions, tt.expected)
			}
		})
	}
}
