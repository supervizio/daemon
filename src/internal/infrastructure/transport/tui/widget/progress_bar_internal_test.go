package widget

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_clamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{"within_bounds", 50, 0, 100, 50},
		{"below_min", -10, 0, 100, 0},
		{"above_max", 150, 0, 100, 100},
		{"at_min", 0, 0, 100, 0},
		{"at_max", 100, 0, 100, 100},
		{"negative_range", -50, -100, -10, -50},
		{"decimal_values", 0.5, 0.1, 0.9, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := clamp(tt.value, tt.min, tt.max)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBarConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"percent_threshold_critical", percentThresholdCritical, 90.0},
		{"percent_threshold_warning", percentThresholdWarning, 70.0},
		{"percent_min", percentMin, 0.0},
		{"percent_max", percentMax, 100.0},
		{"sub_block_levels", subBlockLevels, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.got)
		})
	}

	// Verify slice lengths in separate table
	sliceTests := []struct {
		name    string
		slice   any
		wantLen int
	}{
		{"sub_block_chars_len", SubBlockChars, 9},
		{"sparks_len", Sparks, 8},
	}

	for _, tt := range sliceTests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Len(t, tt.slice, tt.wantLen)
		})
	}
}

func Test_ProgressBar_writeLabel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		label    string
		expected string
	}{
		{"with_label", "CPU", "CPU "},
		{"empty_label", "", ""},
		{"long_label", "Memory Usage", "Memory Usage "},
		{"single_char", "M", "M "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := &ProgressBar{Label: tt.label}
			var sb strings.Builder
			bar.writeLabel(&sb)
			assert.Equal(t, tt.expected, sb.String())
		})
	}
}

func Test_ProgressBar_calculateBarWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		width    int
		style    BarStyle
		expected int
	}{
		{"with_brackets", 20, BracketBar, 18},
		{"no_brackets", 20, BlockBar, 20},
		{"left_only", 20, BarStyle{Full: "#", Empty: " ", Left: "[", Right: ""}, 19},
		{"right_only", 20, BarStyle{Full: "#", Empty: " ", Left: "", Right: "]"}, 19},
		{"minimum_width", 2, BracketBar, 1},
		{"zero_width", 0, BracketBar, 1},
		{"negative_width", -5, BracketBar, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := &ProgressBar{Width: tt.width, Style: tt.style}
			result := bar.calculateBarWidth()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_ProgressBar_writeBarContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		percent     float64
		barWidth    int
		style       BarStyle
		color       string
		containsFul bool
	}{
		{"zero_percent", 0, 10, BlockBar, "\033[32m", false},
		{"full_percent", 100, 10, BlockBar, "\033[32m", true},
		{"half_percent", 50, 10, BlockBar, "\033[32m", true},
		{"small_width", 50, 3, BlockBar, "\033[31m", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := &ProgressBar{
				Percent: tt.percent,
				Style:   tt.style,
				Color:   tt.color,
			}
			var sb strings.Builder
			bar.writeBarContent(&sb, tt.barWidth)
			result := sb.String()
			assert.NotEmpty(t, result)
			if tt.containsFul {
				assert.Contains(t, result, tt.style.Full)
			}
		})
	}
}

func Test_ProgressBar_calculateFillUnits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		percent          float64
		barWidth         int
		wantFull         int
		wantPartial      int
		wantEmptyAtLeast int
	}{
		{"zero_percent", 0, 10, 0, 0, 10},
		{"full_percent", 100, 10, 10, 0, 0},
		{"half_percent", 50, 10, 5, 0, 5},
		{"quarter_percent", 25, 8, 2, 0, 6},
		{"partial_fill", 12.5, 10, 1, 0, 9},
		{"small_width", 50, 2, 1, 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := &ProgressBar{Percent: tt.percent}
			fullChars, partialEighths, emptyChars := bar.calculateFillUnits(tt.barWidth)
			assert.Equal(t, tt.wantFull, fullChars, "fullChars mismatch")
			assert.GreaterOrEqual(t, partialEighths, 0, "partialEighths should be >= 0")
			assert.Less(t, partialEighths, 8, "partialEighths should be < 8")
			assert.GreaterOrEqual(t, emptyChars, 0, "emptyChars should be >= 0")
		})
	}
}

func Test_ProgressBar_writePercentValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		percent   float64
		showValue bool
		contains  string
	}{
		{"single_digit", 5, true, "5%"},
		{"double_digit", 50, true, "50%"},
		{"triple_digit", 100, true, "100%"},
		{"zero_percent", 0, true, "0%"},
		{"hidden", 50, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bar := &ProgressBar{Percent: tt.percent, ShowValue: tt.showValue}
			var sb strings.Builder
			bar.writePercentValue(&sb)
			result := sb.String()
			if tt.showValue {
				assert.Contains(t, result, tt.contains)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}
