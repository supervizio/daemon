package widget_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
	"github.com/stretchr/testify/assert"
)

func TestNewBox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{"small", 10},
		{"medium", 40},
		{"large", 100},
		{"tiny", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			box := widget.NewBox(tt.width)
			assert.NotNil(t, box)
			assert.Equal(t, tt.width, box.Width)
		})
	}
}

func TestBox_SetTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"normal_title", "Test Title", "Test Title"},
		{"empty_title", "", ""},
		{"long_title", "This is a very long title that spans multiple words", "This is a very long title that spans multiple words"},
		{"unicode_title", "CPU Stats", "CPU Stats"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			box := widget.NewBox(40)
			result := box.SetTitle(tt.title)
			assert.Equal(t, box, result)
			assert.Equal(t, tt.want, box.Title)
		})
	}
}

func TestBox_SetTitleColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		color string
		want  string
	}{
		{"green_color", "\033[32m", "\033[32m"},
		{"red_color", "\033[31m", "\033[31m"},
		{"empty_color", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			box := widget.NewBox(40)
			result := box.SetTitleColor(tt.color)
			assert.Equal(t, box, result)
			assert.Equal(t, tt.want, box.TitleColor)
		})
	}
}

func TestBox_SetStyle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		style widget.BoxStyle
	}{
		{"rounded", widget.RoundedBox},
		{"square", widget.SquareBox},
		{"ascii", widget.ASCIIBox},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			box := widget.NewBox(40)
			result := box.SetStyle(tt.style)
			assert.Equal(t, box, result)
			assert.Equal(t, tt.style, box.Style)
		})
	}
}

func TestBox_AddLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		line        string
		wantLen     int
		wantContent string
	}{
		{"single_line", "Line 1", 1, "Line 1"},
		{"empty_line", "", 1, ""},
		{"long_line", "This is a very long line of content", 1, "This is a very long line of content"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			box := widget.NewBox(40)
			result := box.AddLine(tt.line)
			assert.Equal(t, box, result)
			assert.Len(t, box.Content, tt.wantLen)
			assert.Equal(t, tt.wantContent, box.Content[0])
		})
	}
}

func TestBox_AddLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		lines   []string
		wantLen int
	}{
		{"three_lines", []string{"Line 1", "Line 2", "Line 3"}, 3},
		{"single_line", []string{"Only one"}, 1},
		{"empty_lines", []string{}, 0},
		{"many_lines", []string{"A", "B", "C", "D", "E"}, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			box := widget.NewBox(40)
			result := box.AddLines(tt.lines)
			assert.Equal(t, box, result)
			assert.Len(t, box.Content, tt.wantLen)
		})
	}
}

func TestBox_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		width   int
		title   string
		content []string
	}{
		{"empty", 40, "", nil},
		{"with_title", 40, "Title", nil},
		{"with_content", 40, "", []string{"Line 1", "Line 2"}},
		{"full", 40, "Title", []string{"Line 1", "Line 2"}},
		{"tiny", 4, "Title", []string{"A"}},
		{"long_title", 20, "This is a very long title", []string{"A"}},
		{"long_content", 20, "", []string{"This is a very long line"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			box := widget.NewBox(tt.width)
			if tt.title != "" {
				box.SetTitle(tt.title)
			}
			if tt.content != nil {
				box.AddLines(tt.content)
			}
			result := box.Render()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "\n")
		})
	}
}

func TestVisibleLen(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"plain", "hello", 5},
		{"empty", "", 0},
		{"with_ansi", "\033[31mhello\033[0m", 5},
		{"only_ansi", "\033[31m\033[0m", 0},
		{"unicode", "hello world", 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.VisibleLen(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBoxStyles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		style widget.BoxStyle
	}{
		{"rounded", widget.RoundedBox},
		{"square", widget.SquareBox},
		{"ascii", widget.ASCIIBox},
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
		})
	}
}
