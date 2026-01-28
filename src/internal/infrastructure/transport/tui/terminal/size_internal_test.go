package terminal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTerminalConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"minCols", minCols, 40},
		{"minRows", minRows, 10},
		{"defaultCols", defaultCols, 80},
		{"defaultRows", defaultRows, 24},
		{"columnsWide", columnsWide, 2},
		{"columnsUltraWide", columnsUltraWide, 3},
		{"BreakpointSmall", BreakpointSmall, 80},
		{"BreakpointMedium", BreakpointMedium, 120},
		{"BreakpointLarge", BreakpointLarge, 160},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestLayoutValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		layout   Layout
		expected int
	}{
		{"compact", LayoutCompact, 0},
		{"normal", LayoutNormal, 1},
		{"wide", LayoutWide, 2},
		{"ultrawide", LayoutUltraWide, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, int(tt.layout))
		})
	}
}

func TestSize_Structure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cols     int
		rows     int
		wantCols int
		wantRows int
	}{
		{
			name:     "standard_size",
			cols:     100,
			rows:     30,
			wantCols: 100,
			wantRows: 30,
		},
		{
			name:     "minimum_size",
			cols:     40,
			rows:     10,
			wantCols: 40,
			wantRows: 10,
		},
		{
			name:     "large_size",
			cols:     200,
			rows:     60,
			wantCols: 200,
			wantRows: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := Size{Cols: tt.cols, Rows: tt.rows}
			assert.Equal(t, tt.wantCols, s.Cols)
			assert.Equal(t, tt.wantRows, s.Rows)
		})
	}
}

func TestGetLayout_AllCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cols     int
		expected Layout
	}{
		{"below_small_breakpoint", 79, LayoutCompact},
		{"at_small_breakpoint", 80, LayoutNormal},
		{"between_small_and_medium", 100, LayoutNormal},
		{"at_medium_breakpoint", 120, LayoutWide},
		{"between_medium_and_large", 140, LayoutWide},
		{"at_large_breakpoint", 160, LayoutUltraWide},
		{"above_large_breakpoint", 200, LayoutUltraWide},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			size := Size{Cols: tt.cols, Rows: 24}
			layout := GetLayout(size)
			assert.Equal(t, tt.expected, layout)
		})
	}
}

func TestLayout_String_Coverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		layout   Layout
		expected string
	}{
		{"compact", LayoutCompact, "compact"},
		{"normal", LayoutNormal, "normal"},
		{"wide", LayoutWide, "wide"},
		{"ultrawide", LayoutUltraWide, "ultrawide"},
		{"negative", Layout(-1), "unknown"},
		{"large_invalid", Layout(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.layout.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLayout_Columns_Coverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		layout   Layout
		expected int
	}{
		{"compact", LayoutCompact, 1},
		{"normal", LayoutNormal, 1},
		{"wide", LayoutWide, columnsWide},
		{"ultrawide", LayoutUltraWide, columnsUltraWide},
		{"invalid", Layout(999), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.layout.Columns()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMinSizeValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    string
		got      int
		expected int
	}{
		{
			name:     "min_cols",
			field:    "Cols",
			got:      MinSize.Cols,
			expected: minCols,
		},
		{
			name:     "min_rows",
			field:    "Rows",
			got:      MinSize.Rows,
			expected: minRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}

func TestDefaultSizeValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    string
		got      int
		expected int
	}{
		{
			name:     "default_cols",
			field:    "Cols",
			got:      DefaultSize.Cols,
			expected: defaultCols,
		},
		{
			name:     "default_rows",
			field:    "Rows",
			got:      DefaultSize.Rows,
			expected: defaultRows,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}
