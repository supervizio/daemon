// Package ansi provides ANSI escape sequences for terminal styling.
// Pure Go implementation with no external dependencies.
package ansi

import "strconv"

// Reset and text attributes.
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Blink     = "\033[5m"
	Reverse   = "\033[7m"
	Hidden    = "\033[8m"
)

// Standard foreground colors.
const (
	FgBlack   = "\033[30m"
	FgRed     = "\033[31m"
	FgGreen   = "\033[32m"
	FgYellow  = "\033[33m"
	FgBlue    = "\033[34m"
	FgMagenta = "\033[35m"
	FgCyan    = "\033[36m"
	FgWhite   = "\033[37m"
	FgGray    = "\033[90m"
)

// Bright foreground colors.
const (
	FgBrightRed     = "\033[91m"
	FgBrightGreen   = "\033[92m"
	FgBrightYellow  = "\033[93m"
	FgBrightBlue    = "\033[94m"
	FgBrightMagenta = "\033[95m"
	FgBrightCyan    = "\033[96m"
	FgBrightWhite   = "\033[97m"
)

// Standard background colors.
const (
	BgBlack   = "\033[40m"
	BgRed     = "\033[41m"
	BgGreen   = "\033[42m"
	BgYellow  = "\033[43m"
	BgBlue    = "\033[44m"
	BgMagenta = "\033[45m"
	BgCyan    = "\033[46m"
	BgWhite   = "\033[47m"
	BgGray    = "\033[100m"
)

// Screen and cursor control.
const (
	ClearScreen      = "\033[2J"
	ClearLine        = "\033[2K"
	ClearToEnd       = "\033[K"
	CursorHome       = "\033[H"
	CursorHide       = "\033[?25l"
	CursorShow       = "\033[?25h"
	SaveCursor       = "\033[s"
	RestoreCursor    = "\033[u"
	EnableAltScreen  = "\033[?1049h"
	DisableAltScreen = "\033[?1049l"
	ScrollUp         = "\033[S"
	ScrollDown       = "\033[T"
)

// MoveTo returns escape sequence to position cursor at row, col (1-indexed).
// Uses byte buffer to avoid multiple string allocations.
func MoveTo(row, col int) string {
	var buf [24]byte
	b := buf[:0]
	b = append(b, "\033["...)
	b = strconv.AppendInt(b, int64(row), 10)
	b = append(b, ';')
	b = strconv.AppendInt(b, int64(col), 10)
	b = append(b, 'H')
	return string(b)
}

// MoveUp returns escape sequence to move cursor up n lines.
func MoveUp(n int) string {
	if n <= 0 {
		return ""
	}
	var buf [16]byte
	b := buf[:0]
	b = append(b, "\033["...)
	b = strconv.AppendInt(b, int64(n), 10)
	b = append(b, 'A')
	return string(b)
}

// MoveDown returns escape sequence to move cursor down n lines.
func MoveDown(n int) string {
	if n <= 0 {
		return ""
	}
	var buf [16]byte
	b := buf[:0]
	b = append(b, "\033["...)
	b = strconv.AppendInt(b, int64(n), 10)
	b = append(b, 'B')
	return string(b)
}

// MoveRight returns escape sequence to move cursor right n columns.
func MoveRight(n int) string {
	if n <= 0 {
		return ""
	}
	var buf [16]byte
	b := buf[:0]
	b = append(b, "\033["...)
	b = strconv.AppendInt(b, int64(n), 10)
	b = append(b, 'C')
	return string(b)
}

// MoveLeft returns escape sequence to move cursor left n columns.
func MoveLeft(n int) string {
	if n <= 0 {
		return ""
	}
	var buf [16]byte
	b := buf[:0]
	b = append(b, "\033["...)
	b = strconv.AppendInt(b, int64(n), 10)
	b = append(b, 'D')
	return string(b)
}

// RGB returns 24-bit true color foreground escape sequence.
// Uses byte buffer to avoid 7 string allocations per call.
func RGB(r, g, b uint8) string {
	var buf [24]byte
	n := copy(buf[:], "\033[38;2;")
	n += copy(buf[n:], strconv.FormatUint(uint64(r), 10))
	buf[n] = ';'
	n++
	n += copy(buf[n:], strconv.FormatUint(uint64(g), 10))
	buf[n] = ';'
	n++
	n += copy(buf[n:], strconv.FormatUint(uint64(b), 10))
	buf[n] = 'm'
	return string(buf[:n+1])
}

// BgRGB returns 24-bit true color background escape sequence.
// Uses byte buffer to avoid 7 string allocations per call.
func BgRGB(r, g, b uint8) string {
	var buf [24]byte
	n := copy(buf[:], "\033[48;2;")
	n += copy(buf[n:], strconv.FormatUint(uint64(r), 10))
	buf[n] = ';'
	n++
	n += copy(buf[n:], strconv.FormatUint(uint64(g), 10))
	buf[n] = ';'
	n++
	n += copy(buf[n:], strconv.FormatUint(uint64(b), 10))
	buf[n] = 'm'
	return string(buf[:n+1])
}

// Color256 returns 256-color foreground escape sequence.
func Color256(code uint8) string {
	var buf [16]byte
	b := buf[:0]
	b = append(b, "\033[38;5;"...)
	b = strconv.AppendUint(b, uint64(code), 10)
	b = append(b, 'm')
	return string(b)
}

// BgColor256 returns 256-color background escape sequence.
func BgColor256(code uint8) string {
	var buf [16]byte
	b := buf[:0]
	b = append(b, "\033[48;5;"...)
	b = strconv.AppendUint(b, uint64(code), 10)
	b = append(b, 'm')
	return string(b)
}
