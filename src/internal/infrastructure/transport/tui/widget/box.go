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
	// Using nil slice per VAR-NILSLICE (capacity unknown at creation).
	return &Box{
		Style:       RoundedBox,
		BorderColor: ansi.FgGray,
		Width:       width,
		Content:     nil,
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
	// evaluate condition.
	if b.Width < minBoxWidth {
		b.Width = minBoxWidth
	}

	var sb strings.Builder
	innerWidth := b.Width - borderWidth

	b.renderTopBorder(&sb, innerWidth)
	b.renderContentLines(&sb, innerWidth)
	b.renderBottomBorder(&sb, innerWidth)

	// return computed result.
	return sb.String()
}

// renderTopBorder renders the top border with optional title.
//
// Params:
//   - sb: string builder to write to
//   - innerWidth: width inside borders
func (b *Box) renderTopBorder(sb *strings.Builder, innerWidth int) {
	sb.WriteString(b.BorderColor)
	sb.WriteString(b.Style.TopLeft)

	// evaluate condition.
	if b.Title != "" && b.titleFits(innerWidth) {
		b.renderTitleInBorder(sb, innerWidth)
		// handle alternative case.
	} else {
		sb.WriteString(repeatHorizontal(b.Style.Horizontal, innerWidth))
	}

	sb.WriteString(b.Style.TopRight)
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")
}

// titleFits checks if the title fits within the available width.
//
// Params:
//   - innerWidth: available width for title
//
// Returns:
//   - bool: true if title fits with padding
func (b *Box) titleFits(innerWidth int) bool {
	// return computed result.
	return VisibleLen(b.Title)+titlePadding <= innerWidth
}

// renderTitleInBorder renders the title within the top border.
//
// Params:
//   - sb: string builder to write to
//   - innerWidth: width inside borders
func (b *Box) renderTitleInBorder(sb *strings.Builder, innerWidth int) {
	sb.WriteString(b.Style.TitleLeft)
	// evaluate condition.
	if b.TitleColor != "" {
		sb.WriteString(b.TitleColor)
	}
	sb.WriteString(b.Title)
	sb.WriteString(b.BorderColor)
	sb.WriteString(b.Style.TitleRight)
	remaining := innerWidth - VisibleLen(b.Title) - titlePadding
	sb.WriteString(repeatHorizontal(b.Style.Horizontal, remaining))
}

// renderContentLines renders all content lines with vertical borders.
//
// Params:
//   - sb: string builder to write to
//   - innerWidth: width inside borders
func (b *Box) renderContentLines(sb *strings.Builder, innerWidth int) {
	// iterate over collection.
	for _, line := range b.Content {
		b.renderContentLine(sb, line, innerWidth)
	}
}

// renderContentLine renders a single content line with borders.
//
// Params:
//   - sb: string builder to write to
//   - line: content line to render
//   - innerWidth: width inside borders
func (b *Box) renderContentLine(sb *strings.Builder, line string, innerWidth int) {
	sb.WriteString(b.BorderColor)
	sb.WriteString(b.Style.Vertical)
	sb.WriteString(ansi.Reset)

	lineLen := VisibleLen(line)
	// evaluate condition.
	if lineLen < innerWidth {
		sb.WriteString(line)
		sb.WriteString(Spaces(innerWidth - lineLen))
		// handle alternative case.
	} else {
		sb.WriteString(truncateVisible(line, innerWidth))
	}

	sb.WriteString(b.BorderColor)
	sb.WriteString(b.Style.Vertical)
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")
}

// renderBottomBorder renders the bottom border.
//
// Params:
//   - sb: string builder to write to
//   - innerWidth: width inside borders
func (b *Box) renderBottomBorder(sb *strings.Builder, innerWidth int) {
	sb.WriteString(b.BorderColor)
	sb.WriteString(b.Style.BottomLeft)
	sb.WriteString(repeatHorizontal(b.Style.Horizontal, innerWidth))
	sb.WriteString(b.Style.BottomRight)
	sb.WriteString(ansi.Reset)
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
