package widget_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
	"github.com/stretchr/testify/assert"
)

func TestNewProgressBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		width   int
		percent float64
	}{
		{"zero_percent", 20, 0},
		{"half_percent", 20, 50},
		{"full_percent", 20, 100},
		{"over_percent", 20, 150},
		{"negative_percent", 20, -10},
		{"small_width", 5, 50},
		{"large_width", 100, 75},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := widget.NewProgressBar(tt.width, tt.percent)
			assert.NotNil(t, bar)
			assert.Equal(t, tt.width, bar.Width)
		})
	}
}

func TestProgressBar_SetLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		label string
		want  string
	}{
		{"cpu_label", "CPU", "CPU"},
		{"memory_label", "MEM", "MEM"},
		{"empty_label", "", ""},
		{"long_label", "Network I/O", "Network I/O"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := widget.NewProgressBar(20, 50)
			result := bar.SetLabel(tt.label)
			assert.Equal(t, bar, result)
			assert.Equal(t, tt.want, bar.Label)
		})
	}
}

func TestProgressBar_SetStyle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		style widget.BarStyle
		want  widget.BarStyle
	}{
		{"ascii_style", widget.ASCIIBar, widget.ASCIIBar},
		{"block_style", widget.BlockBar, widget.BlockBar},
		{"bracket_style", widget.BracketBar, widget.BracketBar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := widget.NewProgressBar(20, 50)
			result := bar.SetStyle(tt.style)
			assert.Equal(t, bar, result)
			assert.Equal(t, tt.want, bar.Style)
		})
	}
}

func TestProgressBar_SetColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		color string
		want  string
	}{
		{"red_color", "\033[31m", "\033[31m"},
		{"green_color", "\033[32m", "\033[32m"},
		{"empty_color", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := widget.NewProgressBar(20, 50)
			result := bar.SetColor(tt.color)
			assert.Equal(t, bar, result)
			assert.Equal(t, tt.want, bar.Color)
		})
	}
}

func TestProgressBar_SetColorByPercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		percent float64
	}{
		{"normal", 50},
		{"warning", 75},
		{"critical", 95},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := widget.NewProgressBar(20, tt.percent)
			result := bar.SetColorByPercent()
			assert.Equal(t, bar, result)
		})
	}
}

func TestProgressBar_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		width    int
		percent  float64
		label    string
		hasValue bool
	}{
		{"with_label", 20, 50, "CPU", true},
		{"without_label", 20, 50, "", true},
		{"zero_percent", 20, 0, "", true},
		{"full_percent", 20, 100, "", true},
		{"small_width", 5, 50, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := widget.NewProgressBar(tt.width, tt.percent)
			if tt.label != "" {
				bar.SetLabel(tt.label)
			}
			bar.ShowValue = tt.hasValue
			result := bar.Render()
			assert.NotEmpty(t, result)
			if tt.label != "" {
				assert.Contains(t, result, tt.label)
			}
		})
	}
}

func TestBarStyles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		style widget.BarStyle
	}{
		{"block_bar", widget.BlockBar},
		{"bracket_bar", widget.BracketBar},
		{"ascii_bar", widget.ASCIIBar},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.NotEmpty(t, tt.style.Full)
		})
	}
}

func TestSubBlockChars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		chars   []string
		wantLen int
	}{
		{"sub_block_chars", widget.SubBlockChars, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Len(t, tt.chars, tt.wantLen)
		})
	}
}
