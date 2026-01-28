// Package widget provides internal tests for box.go.
// It tests internal implementation details using white-box testing.
package widget

import (
	"strings"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/stretchr/testify/assert"
)

// Test_renderTopBorder tests the renderTopBorder method.
// It verifies that top border rendering works correctly with various configurations.
//
// Params:
//   - t: the testing context.
func Test_Box_renderTopBorder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		width        int
		title        string
		titleColor   string
		borderColor  string
		style        BoxStyle
		wantContains []string
		wantNotEmpty bool
	}{
		{
			name:         "border_without_title",
			width:        20,
			title:        "",
			titleColor:   "",
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"╭", "╮", "─", ansi.Reset, "\n"},
			wantNotEmpty: true,
		},
		{
			name:         "border_with_title",
			width:        30,
			title:        "Test",
			titleColor:   "",
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"╭", "╮", "Test", "─ ", " ─"},
			wantNotEmpty: true,
		},
		{
			name:         "border_with_colored_title",
			width:        30,
			title:        "Color",
			titleColor:   ansi.FgGreen,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"╭", "╮", "Color", ansi.FgGreen, ansi.FgGray},
			wantNotEmpty: true,
		},
		{
			name:         "narrow_box_no_title_space",
			width:        10,
			title:        "Very Long Title",
			titleColor:   "",
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"╭", "╮", "─"},
			wantNotEmpty: true,
		},
		{
			name:         "square_style",
			width:        20,
			title:        "Box",
			titleColor:   ansi.FgBlue,
			borderColor:  ansi.FgGray,
			style:        SquareBox,
			wantContains: []string{"┌", "┐", "Box"},
			wantNotEmpty: true,
		},
		{
			name:         "ascii_style",
			width:        20,
			title:        "ASCII",
			titleColor:   "",
			borderColor:  ansi.FgGray,
			style:        ASCIIBox,
			wantContains: []string{"+", "+", "ASCII", "- ", " -"},
			wantNotEmpty: true,
		},
		{
			name:         "minimum_width",
			width:        minBoxWidth,
			title:        "",
			titleColor:   "",
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"╭", "╮"},
			wantNotEmpty: true,
		},
		{
			name:         "empty_title_still_renders_border",
			width:        15,
			title:        "",
			titleColor:   ansi.FgRed,
			borderColor:  ansi.FgCyan,
			style:        RoundedBox,
			wantContains: []string{"╭", "╮", ansi.FgCyan, ansi.Reset},
			wantNotEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := NewBox(tt.width)
			b.Title = tt.title
			b.TitleColor = tt.titleColor
			b.BorderColor = tt.borderColor
			b.Style = tt.style

			var sb strings.Builder
			innerWidth := tt.width - borderWidth

			b.renderTopBorder(&sb, innerWidth)

			result := sb.String()

			if tt.wantNotEmpty {
				assert.NotEmpty(t, result)
			}

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}
		})
	}
}

// Test_titleFits tests the titleFits method.
// It verifies that title fitting logic works correctly with various title lengths.
//
// Params:
//   - t: the testing context.
func Test_Box_titleFits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		title      string
		innerWidth int
		want       bool
	}{
		{
			name:       "short_title_fits",
			title:      "OK",
			innerWidth: 20,
			want:       true,
		},
		{
			name:       "long_title_does_not_fit",
			title:      "This is a very long title",
			innerWidth: 10,
			want:       false,
		},
		{
			name:       "exact_fit_with_padding",
			title:      "ABCD",
			innerWidth: 8,
			want:       true,
		},
		{
			name:       "empty_title_fits",
			title:      "",
			innerWidth: 5,
			want:       true,
		},
		{
			name:       "single_char_title",
			title:      "X",
			innerWidth: 10,
			want:       true,
		},
		{
			name:       "title_exactly_too_long",
			title:      "ABCDE",
			innerWidth: 8,
			want:       false,
		},
		{
			name:       "zero_inner_width",
			title:      "A",
			innerWidth: 0,
			want:       false,
		},
		{
			name:       "negative_inner_width",
			title:      "Test",
			innerWidth: -5,
			want:       false,
		},
		{
			name:       "unicode_characters",
			title:      "日本",
			innerWidth: 20,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := NewBox(50)
			b.Title = tt.title

			result := b.titleFits(tt.innerWidth)

			assert.Equal(t, tt.want, result)
		})
	}
}

// Test_renderTitleInBorder tests the renderTitleInBorder method.
// It verifies that title rendering within border works correctly.
//
// Params:
//   - t: the testing context.
func Test_Box_renderTitleInBorder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		title        string
		titleColor   string
		borderColor  string
		style        BoxStyle
		innerWidth   int
		wantContains []string
		wantMinLen   int
	}{
		{
			name:         "simple_title",
			title:        "Test",
			titleColor:   "",
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			innerWidth:   20,
			wantContains: []string{"Test", "─ ", " ─"},
			wantMinLen:   4,
		},
		{
			name:         "colored_title",
			title:        "Color",
			titleColor:   ansi.FgRed,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			innerWidth:   20,
			wantContains: []string{"Color", ansi.FgRed, ansi.FgGray},
			wantMinLen:   5,
		},
		{
			name:         "long_title_narrow_width",
			title:        "LongTitle",
			titleColor:   ansi.FgBlue,
			borderColor:  ansi.FgCyan,
			style:        RoundedBox,
			innerWidth:   15,
			wantContains: []string{"LongTitle", ansi.FgBlue},
			wantMinLen:   9,
		},
		{
			name:         "single_char_title",
			title:        "X",
			titleColor:   ansi.FgYellow,
			borderColor:  ansi.FgGray,
			style:        SquareBox,
			innerWidth:   10,
			wantContains: []string{"X", "─ ", " ─"},
			wantMinLen:   1,
		},
		{
			name:         "ascii_style_title",
			title:        "ASCII",
			titleColor:   "",
			borderColor:  ansi.FgGray,
			style:        ASCIIBox,
			innerWidth:   20,
			wantContains: []string{"ASCII", "- ", " -"},
			wantMinLen:   5,
		},
		{
			name:         "no_title_color",
			title:        "Plain",
			titleColor:   "",
			borderColor:  ansi.FgGreen,
			style:        RoundedBox,
			innerWidth:   15,
			wantContains: []string{"Plain", ansi.FgGreen},
			wantMinLen:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := NewBox(50)
			b.Title = tt.title
			b.TitleColor = tt.titleColor
			b.BorderColor = tt.borderColor
			b.Style = tt.style

			var sb strings.Builder
			b.renderTitleInBorder(&sb, tt.innerWidth)

			result := sb.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}

			assert.GreaterOrEqual(t, VisibleLen(result), tt.wantMinLen)
		})
	}
}

// Test_renderContentLines tests the renderContentLines method.
// It verifies that content line rendering works correctly.
//
// Params:
//   - t: the testing context.
func Test_Box_renderContentLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		content       []string
		innerWidth    int
		wantLineCount int
	}{
		{
			name:          "empty_content",
			content:       nil,
			innerWidth:    20,
			wantLineCount: 0,
		},
		{
			name:          "single_line",
			content:       []string{"Line 1"},
			innerWidth:    20,
			wantLineCount: 1,
		},
		{
			name:          "multiple_lines",
			content:       []string{"Line 1", "Line 2", "Line 3"},
			innerWidth:    20,
			wantLineCount: 3,
		},
		{
			name:          "many_lines",
			content:       []string{"A", "B", "C", "D", "E", "F", "G"},
			innerWidth:    15,
			wantLineCount: 7,
		},
		{
			name:          "empty_string_lines",
			content:       []string{"", "", ""},
			innerWidth:    10,
			wantLineCount: 3,
		},
		{
			name:          "mixed_empty_and_content",
			content:       []string{"Line 1", "", "Line 3"},
			innerWidth:    20,
			wantLineCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := NewBox(50)
			b.Content = tt.content

			var sb strings.Builder
			b.renderContentLines(&sb, tt.innerWidth)

			result := sb.String()
			lineCount := strings.Count(result, "\n")

			assert.Equal(t, tt.wantLineCount, lineCount)
		})
	}
}

// Test_renderContentLine tests the renderContentLine method.
// It verifies that single content line rendering works correctly.
//
// Params:
//   - t: the testing context.
func Test_Box_renderContentLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		line         string
		innerWidth   int
		borderColor  string
		style        BoxStyle
		wantContains []string
		wantEndsNewl bool
	}{
		{
			name:         "short_line",
			line:         "Hello",
			innerWidth:   20,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"│", "Hello", ansi.Reset},
			wantEndsNewl: true,
		},
		{
			name:         "long_line_truncated",
			line:         "This is a very long line that should be truncated",
			innerWidth:   10,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"│", ansi.Reset},
			wantEndsNewl: true,
		},
		{
			name:         "empty_line",
			line:         "",
			innerWidth:   20,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"│"},
			wantEndsNewl: true,
		},
		{
			name:         "line_with_ansi_colors",
			line:         ansi.FgRed + "Error" + ansi.Reset,
			innerWidth:   20,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"│", "Error"},
			wantEndsNewl: true,
		},
		{
			name:         "exact_width_line",
			line:         "ExactlyTen",
			innerWidth:   10,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"│", "ExactlyTen"},
			wantEndsNewl: true,
		},
		{
			name:         "square_box_style",
			line:         "Square",
			innerWidth:   15,
			borderColor:  ansi.FgCyan,
			style:        SquareBox,
			wantContains: []string{"│", "Square"},
			wantEndsNewl: true,
		},
		{
			name:         "ascii_box_style",
			line:         "ASCII",
			innerWidth:   15,
			borderColor:  ansi.FgGray,
			style:        ASCIIBox,
			wantContains: []string{"|", "ASCII"},
			wantEndsNewl: true,
		},
		{
			name:         "narrow_width",
			line:         "Test",
			innerWidth:   2,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"│"},
			wantEndsNewl: true,
		},
		{
			name:         "unicode_content",
			line:         "日本語",
			innerWidth:   20,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"│"},
			wantEndsNewl: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := NewBox(50)
			b.BorderColor = tt.borderColor
			b.Style = tt.style

			var sb strings.Builder
			b.renderContentLine(&sb, tt.line, tt.innerWidth)

			result := sb.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}

			if tt.wantEndsNewl {
				assert.True(t, strings.HasSuffix(result, "\n"))
			}
		})
	}
}

// Test_renderBottomBorder tests the renderBottomBorder method.
// It verifies that bottom border rendering works correctly.
//
// Params:
//   - t: the testing context.
func Test_Box_renderBottomBorder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		width        int
		borderColor  string
		style        BoxStyle
		wantContains []string
		wantMinLen   int
	}{
		{
			name:         "standard_bottom_border",
			width:        20,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"╰", "╯", "─", ansi.Reset},
			wantMinLen:   2,
		},
		{
			name:         "narrow_bottom_border",
			width:        5,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"╰", "╯"},
			wantMinLen:   2,
		},
		{
			name:         "wide_bottom_border",
			width:        100,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"╰", "╯", "─"},
			wantMinLen:   2,
		},
		{
			name:         "square_style_bottom",
			width:        20,
			borderColor:  ansi.FgCyan,
			style:        SquareBox,
			wantContains: []string{"└", "┘", "─"},
			wantMinLen:   2,
		},
		{
			name:         "ascii_style_bottom",
			width:        20,
			borderColor:  ansi.FgGray,
			style:        ASCIIBox,
			wantContains: []string{"+", "+", "-"},
			wantMinLen:   2,
		},
		{
			name:         "minimum_width_border",
			width:        minBoxWidth,
			borderColor:  ansi.FgGray,
			style:        RoundedBox,
			wantContains: []string{"╰", "╯"},
			wantMinLen:   2,
		},
		{
			name:         "colored_border",
			width:        15,
			borderColor:  ansi.FgYellow,
			style:        RoundedBox,
			wantContains: []string{"╰", "╯", ansi.FgYellow, ansi.Reset},
			wantMinLen:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := NewBox(tt.width)
			b.BorderColor = tt.borderColor
			b.Style = tt.style

			var sb strings.Builder
			innerWidth := tt.width - borderWidth

			b.renderBottomBorder(&sb, innerWidth)

			result := sb.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}

			assert.GreaterOrEqual(t, len(result), tt.wantMinLen)
		})
	}
}

// Test_repeatHorizontal tests the repeatHorizontal function.
// It verifies that horizontal character repetition works correctly.
//
// Params:
//   - t: the testing context.
func Test_repeatHorizontal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		char    string
		n       int
		wantLen int
	}{
		{
			name:    "unicode_horizontal_bar",
			char:    "─",
			n:       5,
			wantLen: 5,
		},
		{
			name:    "ascii_dash",
			char:    "-",
			n:       5,
			wantLen: 5,
		},
		{
			name:    "zero_repetitions",
			char:    "─",
			n:       0,
			wantLen: 0,
		},
		{
			name:    "single_repetition",
			char:    "─",
			n:       1,
			wantLen: 1,
		},
		{
			name:    "large_repetition",
			char:    "─",
			n:       100,
			wantLen: 100,
		},
		{
			name:    "negative_repetition",
			char:    "─",
			n:       -5,
			wantLen: 0,
		},
		{
			name:    "ascii_plus",
			char:    "+",
			n:       10,
			wantLen: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := repeatHorizontal(tt.char, tt.n)

			if tt.n <= 0 {
				assert.Empty(t, result)
			} else {
				count := strings.Count(result, tt.char)
				assert.Equal(t, tt.wantLen, count)
			}
		})
	}
}

// Test_Box_constants tests the box-related constants.
// It verifies that constants have expected values.
//
// Params:
//   - t: the testing context.
func Test_Box_constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  int
		want int
	}{
		{
			name: "minBoxWidth",
			got:  minBoxWidth,
			want: 4,
		},
		{
			name: "borderWidth",
			got:  borderWidth,
			want: 2,
		},
		{
			name: "titlePadding",
			got:  titlePadding,
			want: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.got)
		})
	}
}

// Test_escapeStart tests the escapeStart constant.
// It verifies that the constant has the correct value.
//
// Params:
//   - t: the testing context.
func Test_escapeStart(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want rune
	}{
		{
			name: "escape_character",
			want: '\033',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, escapeStart)
		})
	}
}

// Test_BoxStyle tests the BoxStyle struct.
// It verifies that box styles have non-empty fields.
//
// Params:
//   - t: the testing context.
func Test_BoxStyle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		style BoxStyle
	}{
		{
			name:  "rounded_box",
			style: RoundedBox,
		},
		{
			name:  "square_box",
			style: SquareBox,
		},
		{
			name:  "ascii_box",
			style: ASCIIBox,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.NotEmpty(t, tt.style.TopLeft)
			assert.NotEmpty(t, tt.style.TopRight)
			assert.NotEmpty(t, tt.style.BottomLeft)
			assert.NotEmpty(t, tt.style.BottomRight)
			assert.NotEmpty(t, tt.style.Horizontal)
			assert.NotEmpty(t, tt.style.Vertical)
			assert.NotEmpty(t, tt.style.TitleLeft)
			assert.NotEmpty(t, tt.style.TitleRight)
		})
	}
}

// Test_Box_Render_minWidth tests that Render handles minimum width.
// It verifies that width below minBoxWidth is corrected.
//
// Params:
//   - t: the testing context.
func Test_Box_Render_minWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "width_below_minimum",
			width: 2,
		},
		{
			name:  "zero_width",
			width: 0,
		},
		{
			name:  "negative_width",
			width: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := NewBox(tt.width)
			b.Width = tt.width

			result := b.Render()

			assert.NotEmpty(t, result)
		})
	}
}

// Test_renderTopBorder_with_different_styles tests top border with all box styles.
// It verifies that each style renders its specific corner characters.
//
// Params:
//   - t: the testing context.
func Test_Box_renderTopBorder_with_different_styles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		style     BoxStyle
		wantLeft  string
		wantRight string
	}{
		{
			name:      "rounded_style",
			style:     RoundedBox,
			wantLeft:  "╭",
			wantRight: "╮",
		},
		{
			name:      "square_style",
			style:     SquareBox,
			wantLeft:  "┌",
			wantRight: "┐",
		},
		{
			name:      "ascii_style",
			style:     ASCIIBox,
			wantLeft:  "+",
			wantRight: "+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := NewBox(20)
			b.Style = tt.style

			var sb strings.Builder
			b.renderTopBorder(&sb, 18)

			result := sb.String()

			assert.Contains(t, result, tt.wantLeft)
			assert.Contains(t, result, tt.wantRight)
		})
	}
}

// Test_renderContentLine_padding tests content line padding behavior.
// It verifies that lines shorter than innerWidth are properly padded.
//
// Params:
//   - t: the testing context.
func Test_Box_renderContentLine_padding(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		line       string
		innerWidth int
		wantSpaces bool
	}{
		{
			name:       "line_needs_padding",
			line:       "Hi",
			innerWidth: 10,
			wantSpaces: true,
		},
		{
			name:       "line_exact_width",
			line:       "ExactlyTen",
			innerWidth: 10,
			wantSpaces: false,
		},
		{
			name:       "empty_line_all_padding",
			line:       "",
			innerWidth: 5,
			wantSpaces: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := NewBox(50)

			var sb strings.Builder
			b.renderContentLine(&sb, tt.line, tt.innerWidth)

			result := sb.String()

			// Count spaces between borders
			if tt.wantSpaces {
				assert.Contains(t, result, "  ")
			}
		})
	}
}
