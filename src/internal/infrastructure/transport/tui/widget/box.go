// Package widget provides reusable TUI components.
package widget

import (
	"strings"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/mattn/go-runewidth"
)

// BoxStyle defines the border characters for a box.
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

// RoundedBox uses rounded Unicode corners.
var RoundedBox = BoxStyle{
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
var SquareBox = BoxStyle{
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
var ASCIIBox = BoxStyle{
	TopLeft:     "+",
	TopRight:    "+",
	BottomLeft:  "+",
	BottomRight: "+",
	Horizontal:  "-",
	Vertical:    "|",
	TitleLeft:   "- ",
	TitleRight:  " -",
}

// Box creates a bordered box with optional title.
type Box struct {
	Style       BoxStyle
	Title       string
	TitleColor  string
	BorderColor string
	Width       int
	Content     []string
}

// NewBox creates a new box with default rounded style.
func NewBox(width int) *Box {
	return &Box{
		Style:       RoundedBox,
		BorderColor: ansi.FgGray,
		Width:       width,
		Content:     make([]string, 0),
	}
}

// SetTitle sets the box title.
func (b *Box) SetTitle(title string) *Box {
	b.Title = title
	return b
}

// SetTitleColor sets the title color.
func (b *Box) SetTitleColor(color string) *Box {
	b.TitleColor = color
	return b
}

// SetStyle sets the box style.
func (b *Box) SetStyle(style BoxStyle) *Box {
	b.Style = style
	return b
}

// AddLine adds a content line.
func (b *Box) AddLine(line string) *Box {
	b.Content = append(b.Content, line)
	return b
}

// AddLines adds multiple content lines.
func (b *Box) AddLines(lines []string) *Box {
	b.Content = append(b.Content, lines...)
	return b
}

// Render returns the box as a string.
func (b *Box) Render() string {
	if b.Width < 4 {
		b.Width = 4
	}

	var sb strings.Builder
	innerWidth := b.Width - 2 // Account for borders

	// Top border with optional title.
	sb.WriteString(b.BorderColor)
	sb.WriteString(b.Style.TopLeft)

	if b.Title != "" {
		titleLen := VisibleLen(b.Title)
		if titleLen+4 <= innerWidth {
			sb.WriteString(b.Style.TitleLeft)
			if b.TitleColor != "" {
				sb.WriteString(b.TitleColor)
			}
			sb.WriteString(b.Title)
			sb.WriteString(b.BorderColor)
			sb.WriteString(b.Style.TitleRight)
			remaining := innerWidth - titleLen - 4
			sb.WriteString(strings.Repeat(b.Style.Horizontal, remaining))
		} else {
			sb.WriteString(strings.Repeat(b.Style.Horizontal, innerWidth))
		}
	} else {
		sb.WriteString(strings.Repeat(b.Style.Horizontal, innerWidth))
	}

	sb.WriteString(b.Style.TopRight)
	sb.WriteString(ansi.Reset)
	sb.WriteString("\n")

	// Content lines.
	for _, line := range b.Content {
		sb.WriteString(b.BorderColor)
		sb.WriteString(b.Style.Vertical)
		sb.WriteString(ansi.Reset)

		// Pad or truncate line to fit.
		lineLen := VisibleLen(line)
		if lineLen < innerWidth {
			sb.WriteString(line)
			sb.WriteString(strings.Repeat(" ", innerWidth-lineLen))
		} else {
			sb.WriteString(truncateVisible(line, innerWidth))
		}

		sb.WriteString(b.BorderColor)
		sb.WriteString(b.Style.Vertical)
		sb.WriteString(ansi.Reset)
		sb.WriteString("\n")
	}

	// Bottom border.
	sb.WriteString(b.BorderColor)
	sb.WriteString(b.Style.BottomLeft)
	sb.WriteString(strings.Repeat(b.Style.Horizontal, innerWidth))
	sb.WriteString(b.Style.BottomRight)
	sb.WriteString(ansi.Reset)

	return sb.String()
}

// VisibleLen returns the visible length of a string (excluding ANSI codes).
// Uses runewidth to correctly handle CJK characters and emoji which occupy
// 2 terminal columns.
func VisibleLen(s string) int {
	// Remove escape sequences and count terminal column width.
	inEscape := false
	length := 0

	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		length += runewidth.RuneWidth(r)
	}

	return length
}

// truncateVisible truncates a string to maxLen visible characters.
func truncateVisible(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	var result strings.Builder
	inEscape := false
	visible := 0

	for _, r := range s {
		if r == '\033' {
			inEscape = true
			result.WriteRune(r)
			continue
		}

		if inEscape {
			result.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		if visible >= maxLen {
			break
		}

		result.WriteRune(r)
		visible++
	}

	// Always reset at the end to prevent color bleed.
	result.WriteString(ansi.Reset)

	return result.String()
}

// TruncateVisible truncates a string to maxLen visible characters.
func TruncateVisible(s string, maxLen int) string {
	return truncateVisible(s, maxLen)
}
