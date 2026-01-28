package layout

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/stretchr/testify/assert"
)

func TestLayoutConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"defaultPadding", defaultPadding, 1},
		{"defaultGap", defaultGap, 2},
		{"paddingSides", paddingSides, 2},
		{"minContentDimension", minContentDimension, 1},
		{"singleSection", singleSection, 1},
		{"noSections", noSections, 0},
		{"defaultColspan", defaultColspan, 1},
		{"flexibleRowHeight", flexibleRowHeight, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.constant)
		})
	}
}

func TestLayout_Modes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		size            terminal.Size
		mode            terminal.Layout
		expectedColumns int
	}{
		{"defaults_normal", terminal.Size{Cols: 80, Rows: 24}, terminal.LayoutNormal, 1},
		{"compact", terminal.Size{Cols: 60, Rows: 20}, terminal.LayoutCompact, 1},
		{"wide", terminal.Size{Cols: 140, Rows: 40}, terminal.LayoutWide, 2},
		{"ultra_wide", terminal.Size{Cols: 200, Rows: 50}, terminal.LayoutUltraWide, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := &Layout{
				Size:    tt.size,
				Mode:    tt.mode,
				Padding: defaultPadding,
				Gap:     defaultGap,
			}

			assert.Equal(t, defaultPadding, l.Padding)
			assert.Equal(t, defaultGap, l.Gap)
			assert.Equal(t, tt.mode, l.Mode)
			assert.Equal(t, tt.expectedColumns, l.Mode.Columns())
		})
	}
}

func TestRegion_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		region Region
		wantX  int
		wantY  int
		wantW  int
		wantH  int
	}{
		{"standard", Region{X: 10, Y: 5, Width: 80, Height: 24}, 10, 5, 80, 24},
		{"origin", Region{X: 0, Y: 0, Width: 100, Height: 50}, 0, 0, 100, 50},
		{"offset", Region{X: 20, Y: 15, Width: 60, Height: 30}, 20, 15, 60, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantX, tt.region.X)
			assert.Equal(t, tt.wantY, tt.region.Y)
			assert.Equal(t, tt.wantW, tt.region.Width)
			assert.Equal(t, tt.wantH, tt.region.Height)
		})
	}
}
