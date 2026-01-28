// Package ansi provides ANSI escape sequences for terminal styling.
// Pure Go implementation with no external dependencies.
//
// Tests: codes_internal_test.go (TODO: create tests for ANSI escape sequences).
package ansi

import "strconv"

// Buffer size constants for ANSI escape sequence formatting.
const (
	// bufSize24 is the buffer size for MoveTo and RGB operations.
	bufSize24 int = 24
	// bufSize16 is the buffer size for cursor movement and 256-color operations.
	bufSize16 int = 16
	// base10 is the decimal base for integer formatting.
	base10 int = 10
)

// Reset and text attributes.
const (
	Reset     string = "\033[0m"
	Bold      string = "\033[1m"
	Dim       string = "\033[2m"
	Italic    string = "\033[3m"
	Underline string = "\033[4m"
	Blink     string = "\033[5m"
	Reverse   string = "\033[7m"
	Hidden    string = "\033[8m"
)

// Standard foreground colors.
const (
	FgBlack   string = "\033[30m"
	FgRed     string = "\033[31m"
	FgGreen   string = "\033[32m"
	FgYellow  string = "\033[33m"
	FgBlue    string = "\033[34m"
	FgMagenta string = "\033[35m"
	FgCyan    string = "\033[36m"
	FgWhite   string = "\033[37m"
	FgGray    string = "\033[90m"
)

// Bright foreground colors.
const (
	FgBrightRed     string = "\033[91m"
	FgBrightGreen   string = "\033[92m"
	FgBrightYellow  string = "\033[93m"
	FgBrightBlue    string = "\033[94m"
	FgBrightMagenta string = "\033[95m"
	FgBrightCyan    string = "\033[96m"
	FgBrightWhite   string = "\033[97m"
)

// Standard background colors.
const (
	BgBlack   string = "\033[40m"
	BgRed     string = "\033[41m"
	BgGreen   string = "\033[42m"
	BgYellow  string = "\033[43m"
	BgBlue    string = "\033[44m"
	BgMagenta string = "\033[45m"
	BgCyan    string = "\033[46m"
	BgWhite   string = "\033[47m"
	BgGray    string = "\033[100m"
)

// Screen and cursor control.
const (
	ClearScreen      string = "\033[2J"
	ClearLine        string = "\033[2K"
	ClearToEnd       string = "\033[K"
	CursorHome       string = "\033[H"
	CursorHide       string = "\033[?25l"
	CursorShow       string = "\033[?25h"
	SaveCursor       string = "\033[s"
	RestoreCursor    string = "\033[u"
	EnableAltScreen  string = "\033[?1049h"
	DisableAltScreen string = "\033[?1049l"
	ScrollUp         string = "\033[S"
	ScrollDown       string = "\033[T"
)

// MoveTo returns escape sequence to position cursor at row, col (1-indexed).
// Uses byte buffer to avoid multiple string allocations.
//
// Params:
//   - row: target row position (1-indexed)
//   - col: target column position (1-indexed)
//
// Returns:
//   - string: ANSI escape sequence for cursor positioning
func MoveTo(row, col int) string {
	// Allocate fixed-size buffer for escape sequence.
	var buf [bufSize24]byte
	b := buf[:0]

	// Build escape sequence: ESC[row;colH
	b = append(b, "\033["...)
	b = strconv.AppendInt(b, int64(row), base10)
	b = append(b, ';')
	b = strconv.AppendInt(b, int64(col), base10)
	b = append(b, 'H')

	// Convert buffer to string.
	return string(b)
}

// MoveUp returns escape sequence to move cursor up n lines.
//
// Params:
//   - n: number of lines to move up
//
// Returns:
//   - string: ANSI escape sequence, or empty string if n <= 0
func MoveUp(n int) string {
	// Validate input parameter.
	if n <= 0 {
		// Return empty string for invalid input.
		return ""
	}

	// Allocate fixed-size buffer for escape sequence.
	var buf [bufSize16]byte
	b := buf[:0]

	// Build escape sequence: ESC[nA
	b = append(b, "\033["...)
	b = strconv.AppendInt(b, int64(n), base10)
	b = append(b, 'A')

	// Convert buffer to string.
	return string(b)
}

// MoveDown returns escape sequence to move cursor down n lines.
//
// Params:
//   - n: number of lines to move down
//
// Returns:
//   - string: ANSI escape sequence, or empty string if n <= 0
func MoveDown(n int) string {
	// Validate input parameter.
	if n <= 0 {
		// Return empty string for invalid input.
		return ""
	}

	// Allocate fixed-size buffer for escape sequence.
	var buf [bufSize16]byte
	b := buf[:0]

	// Build escape sequence: ESC[nB
	b = append(b, "\033["...)
	b = strconv.AppendInt(b, int64(n), base10)
	b = append(b, 'B')

	// Convert buffer to string.
	return string(b)
}

// MoveRight returns escape sequence to move cursor right n columns.
//
// Params:
//   - n: number of columns to move right
//
// Returns:
//   - string: ANSI escape sequence, or empty string if n <= 0
func MoveRight(n int) string {
	// Validate input parameter.
	if n <= 0 {
		// Return empty string for invalid input.
		return ""
	}

	// Allocate fixed-size buffer for escape sequence.
	var buf [bufSize16]byte
	b := buf[:0]

	// Build escape sequence: ESC[nC
	b = append(b, "\033["...)
	b = strconv.AppendInt(b, int64(n), base10)
	b = append(b, 'C')

	// Convert buffer to string.
	return string(b)
}

// MoveLeft returns escape sequence to move cursor left n columns.
//
// Params:
//   - n: number of columns to move left
//
// Returns:
//   - string: ANSI escape sequence, or empty string if n <= 0
func MoveLeft(n int) string {
	// Validate input parameter.
	if n <= 0 {
		// Return empty string for invalid input.
		return ""
	}

	// Allocate fixed-size buffer for escape sequence.
	var buf [bufSize16]byte
	b := buf[:0]

	// Build escape sequence: ESC[nD
	b = append(b, "\033["...)
	b = strconv.AppendInt(b, int64(n), base10)
	b = append(b, 'D')

	// Convert buffer to string.
	return string(b)
}

// RGB returns 24-bit true color foreground escape sequence.
// Uses byte buffer to avoid 7 string allocations per call.
//
// Params:
//   - r: red component (0-255)
//   - g: green component (0-255)
//   - b: blue component (0-255)
//
// Returns:
//   - string: ANSI escape sequence for 24-bit foreground color
func RGB(r, g, b uint8) string {
	// Allocate fixed-size buffer for escape sequence.
	var buf [bufSize24]byte

	// Build escape sequence: ESC[38;2;r;g;bm
	n := copy(buf[:], "\033[38;2;")
	n += copy(buf[n:], strconv.FormatUint(uint64(r), base10))
	buf[n] = ';'
	n++
	n += copy(buf[n:], strconv.FormatUint(uint64(g), base10))
	buf[n] = ';'
	n++
	n += copy(buf[n:], strconv.FormatUint(uint64(b), base10))
	buf[n] = 'm'

	// Convert buffer to string.
	return string(buf[:n+1])
}

// BgRGB returns 24-bit true color background escape sequence.
// Uses byte buffer to avoid 7 string allocations per call.
//
// Params:
//   - r: red component (0-255)
//   - g: green component (0-255)
//   - b: blue component (0-255)
//
// Returns:
//   - string: ANSI escape sequence for 24-bit background color
func BgRGB(r, g, b uint8) string {
	// Allocate fixed-size buffer for escape sequence.
	var buf [bufSize24]byte

	// Build escape sequence: ESC[48;2;r;g;bm
	n := copy(buf[:], "\033[48;2;")
	n += copy(buf[n:], strconv.FormatUint(uint64(r), base10))
	buf[n] = ';'
	n++
	n += copy(buf[n:], strconv.FormatUint(uint64(g), base10))
	buf[n] = ';'
	n++
	n += copy(buf[n:], strconv.FormatUint(uint64(b), base10))
	buf[n] = 'm'

	// Convert buffer to string.
	return string(buf[:n+1])
}

// Color256 returns 256-color foreground escape sequence.
//
// Params:
//   - code: color code (0-255)
//
// Returns:
//   - string: ANSI escape sequence for 256-color foreground
func Color256(code uint8) string {
	// Allocate fixed-size buffer for escape sequence.
	var buf [bufSize16]byte
	b := buf[:0]

	// Build escape sequence: ESC[38;5;codem
	b = append(b, "\033[38;5;"...)
	b = strconv.AppendUint(b, uint64(code), base10)
	b = append(b, 'm')

	// Convert buffer to string.
	return string(b)
}

// BgColor256 returns 256-color background escape sequence.
//
// Params:
//   - code: color code (0-255)
//
// Returns:
//   - string: ANSI escape sequence for 256-color background
func BgColor256(code uint8) string {
	// Allocate fixed-size buffer for escape sequence.
	var buf [bufSize16]byte
	b := buf[:0]

	// Build escape sequence: ESC[48;5;codem
	b = append(b, "\033[48;5;"...)
	b = strconv.AppendUint(b, uint64(code), base10)
	b = append(b, 'm')

	// Convert buffer to string.
	return string(b)
}
