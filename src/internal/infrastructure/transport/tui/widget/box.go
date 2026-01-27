// Package widget provides reusable TUI components.
package widget

import (
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/mattn/go-runewidth"
)

const (
	// minBoxWidth is the minimum allowed box width to ensure borders fit.
	minBoxWidth int = 4
	// borderWidth is the width consumed by left and right borders.
	borderWidth int = 2
	// titlePadding is the space around title text.
	titlePadding int = 4
	// escapeStart is the ASCII code for ANSI escape sequences.
	escapeStart rune = '\033'
)

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

// Box creates a bordered box with optional title.
// It provides customizable borders, colors, and automatic content padding/truncation.
type Box struct {
	Style       BoxStyle
	Title       string
	TitleColor  string
	BorderColor string
	Width       int
	Content     []string
}

// NewBox creates a new box with default rounded style.
//
// Params:
//   - width: total width of the box including borders.
//
// Returns:
//   - *Box: configured box with rounded style and gray border.
func NewBox(width int) *Box {
	// Return configured box with defaults.
	return &Box{
		Style:       RoundedBox,
		BorderColor: ansi.FgGray,
		Width:       width,
		Content:     make([]string, 0),
	}
}

// SetTitle sets the box title.
//
// Params:
//   - title: text to display in the top border.
//
// Returns:
//   - *Box: self for method chaining.
func (b *Box) SetTitle(title string) *Box {
	b.Title = title
	// Return self for method chaining.
	return b
}

// SetTitleColor sets the title color.
//
// Params:
//   - color: ANSI color code for the title text.
//
// Returns:
//   - *Box: self for method chaining.
func (b *Box) SetTitleColor(color string) *Box {
	b.TitleColor = color
	// Return self for method chaining.
	return b
}

// SetStyle sets the box style.
//
// Params:
//   - style: BoxStyle defining border characters.
//
// Returns:
//   - *Box: self for method chaining.
func (b *Box) SetStyle(style BoxStyle) *Box {
	b.Style = style
	// Return self for method chaining.
	return b
}

// AddLine adds a content line.
//
// Params:
//   - line: text to add to the box content.
//
// Returns:
//   - *Box: self for method chaining.
func (b *Box) AddLine(line string) *Box {
	b.Content = append(b.Content, line)
	// Return self for method chaining.
	return b
}

// AddLines adds multiple content lines.
//
// Params:
//   - lines: slice of text lines to add to the box content.
//
// Returns:
//   - *Box: self for method chaining.
func (b *Box) AddLines(lines []string) *Box {
	b.Content = append(b.Content, lines...)
	// Return self for method chaining.
	return b
}

// Render returns the box as a string.
//
// Returns:
//   - string: rendered box with borders, title, and content.
func (b *Box) Render() string {
	// Enforce minimum width to ensure borders fit.
	if b.Width < minBoxWidth {
		b.Width = minBoxWidth
	}

	var sb strings.Builder
	// Account for left and right borders.
	innerWidth := b.Width - borderWidth

	// Top border with optional title.
	sb.WriteString(b.BorderColor)
	sb.WriteString(b.Style.TopLeft)

	// Add title if present and fits within width.
	if b.Title != "" {
		titleLen := VisibleLen(b.Title)
		// Check if title with padding fits in available width.
		if titleLen+titlePadding <= innerWidth {
			sb.WriteString(b.Style.TitleLeft)
			// Apply title color if specified.
			if b.TitleColor != "" {
				sb.WriteString(b.TitleColor)
			}
			sb.WriteString(b.Title)
			sb.WriteString(b.BorderColor)
			sb.WriteString(b.Style.TitleRight)
			// Fill remaining width with horizontal border.
			remaining := innerWidth - titleLen - titlePadding
			sb.WriteString(repeatHorizontal(b.Style.Horizontal, remaining))
		} else {
			// Title too long, use full horizontal border.
			sb.WriteString(repeatHorizontal(b.Style.Horizontal, innerWidth))
		}
	} else {
		// No title, use full horizontal border.
		sb.WriteString(repeatHorizontal(b.Style.Horizontal, innerWidth))
	}

	sb.WriteString(b.Style.TopRight)
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")

	// Content lines with vertical borders.
	for _, line := range b.Content {
		sb.WriteString(b.BorderColor)
		sb.WriteString(b.Style.Vertical)
		sb.WriteString(ansi.Reset)

		// Pad or truncate line to fit inner width.
		lineLen := VisibleLen(line)
		// Check if line needs padding or truncation.
		if lineLen < innerWidth {
			// Line shorter than width, add padding.
			sb.WriteString(line)
			sb.WriteString(Spaces(innerWidth - lineLen))
		} else {
			// Line longer than width, truncate.
			sb.WriteString(truncateVisible(line, innerWidth))
		}

		sb.WriteString(b.BorderColor)
		sb.WriteString(b.Style.Vertical)
		sb.WriteString(ansi.Reset)
		sb.WriteString("\n")
	}

	// Bottom border with corners.
	sb.WriteString(b.BorderColor)
	sb.WriteString(b.Style.BottomLeft)
	sb.WriteString(repeatHorizontal(b.Style.Horizontal, innerWidth))
	sb.WriteString(b.Style.BottomRight)
	sb.WriteString(ansi.Reset)

	// Return the complete rendered box.
	return sb.String()
}

// VisibleLen returns the visible length of a string (excluding ANSI codes).
// Uses runewidth to correctly handle CJK characters and emoji which occupy
// 2 terminal columns.
//
// Params:
//   - s: string containing text and optional ANSI codes.
//
// Returns:
//   - int: visible character width excluding escape sequences.
func VisibleLen(s string) int {
	// Remove escape sequences and count terminal column width.
	inEscape := false
	length := 0

	// Iterate through runes to handle ANSI codes.
	for _, r := range s {
		// Detect start of ANSI escape sequence.
		if r == escapeStart {
			inEscape = true
			continue
		}
		// Skip characters within escape sequence.
		if inEscape {
			// Check for end of escape sequence (letter).
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		// Count visible width of character.
		length += runewidth.RuneWidth(r)
	}

	// Return total visible length.
	return length
}

// truncateVisible truncates a string to maxLen visible characters.
//
// Params:
//   - s: string to truncate containing text and optional ANSI codes.
//   - maxLen: maximum visible character count.
//
// Returns:
//   - string: truncated string with ANSI reset appended.
func truncateVisible(s string, maxLen int) string {
	// Handle zero or negative length.
	if maxLen <= 0 {
		// Return empty for invalid length.
		return ""
	}

	var result strings.Builder
	inEscape := false
	visible := 0

	// Process each rune, preserving ANSI codes.
	for _, r := range s {
		// Detect start of ANSI escape sequence.
		if r == escapeStart {
			inEscape = true
			result.WriteRune(r)
			continue
		}

		// Handle characters within escape sequence.
		if inEscape {
			result.WriteRune(r)
			// Check for end of escape sequence.
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		// Stop if max visible length reached.
		if visible >= maxLen {
			break
		}

		// Add visible character and increment count.
		result.WriteRune(r)
		visible++
	}

	// Always reset at the end to prevent color bleed.
	result.WriteString(ansi.Reset)

	// Return truncated string.
	return result.String()
}

// TruncateVisible truncates a string to maxLen visible characters.
//
// Params:
//   - s: string to truncate containing text and optional ANSI codes.
//   - maxLen: maximum visible character count.
//
// Returns:
//   - string: truncated string with ANSI reset appended.
func TruncateVisible(s string, maxLen int) string {
	// Delegate to internal function.
	return truncateVisible(s, maxLen)
}

// repeatHorizontal returns n repetitions of the horizontal character.
// Uses cache for the common "─" character to avoid allocations.
//
// Params:
//   - char: character to repeat.
//   - n: number of repetitions.
//
// Returns:
//   - string: repeated character string.
func repeatHorizontal(char string, n int) string {
	// Use cached version for Unicode horizontal bar.
	if char == "─" {
		// Return from cache for efficiency.
		return HorizontalBar(n)
	}
	// Generate string for other characters.
	return strings.Repeat(char, n)
}
