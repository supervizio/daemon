// Package widget provides reusable TUI components.
package widget

// BoxStyle defines the border characters for a box.
// It specifies corners, edges, and title decorations for bordered containers.
type BoxStyle struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	TitleLeft   string
	TitleRight  string
}

var (
	// RoundedBox uses rounded Unicode corners.
	RoundedBox BoxStyle = BoxStyle{
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
		Horizontal:  "─",
		Vertical:    "│",
		TitleLeft:   "─ ",
		TitleRight:  " ─",
	}

	// SquareBox uses square Unicode corners.
	SquareBox BoxStyle = BoxStyle{
		TopLeft:     "┌",
		TopRight:    "┐",
		BottomLeft:  "└",
		BottomRight: "┘",
		Horizontal:  "─",
		Vertical:    "│",
		TitleLeft:   "─ ",
		TitleRight:  " ─",
	}

	// ASCIIBox uses ASCII characters only.
	ASCIIBox BoxStyle = BoxStyle{
		TopLeft:     "+",
		TopRight:    "+",
		BottomLeft:  "+",
		BottomRight: "+",
		Horizontal:  "-",
		Vertical:    "|",
		TitleLeft:   "- ",
		TitleRight:  " -",
	}
)
