// Package ansi_test provides black-box tests for the ansi package.
//
// These tests verify the public API behavior without accessing internal details,
// ensuring that escape sequences are generated correctly and consistently.
package ansi_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/stretchr/testify/assert"
)

// TestMoveTo verifies cursor positioning escape sequence generation.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func TestMoveTo(t *testing.T) {
	// Test table for various cursor positions
	tests := []struct {
		name     string
		row      int
		col      int
		expected string
	}{
		{
			name:     "origin position",
			row:      1,
			col:      1,
			expected: "\033[1;1H",
		},
		{
			name:     "middle of screen",
			row:      12,
			col:      40,
			expected: "\033[12;40H",
		},
		{
			name:     "large coordinates",
			row:      100,
			col:      200,
			expected: "\033[100;200H",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate escape sequence
			result := ansi.MoveTo(tt.row, tt.col)

			// Verify result matches expected
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMoveUp verifies upward cursor movement escape sequence generation.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func TestMoveUp(t *testing.T) {
	// Test table for various movement values
	tests := []struct {
		name     string
		n        int
		expected string
	}{
		{
			name:     "single line",
			n:        1,
			expected: "\033[1A",
		},
		{
			name:     "multiple lines",
			n:        10,
			expected: "\033[10A",
		},
		{
			name:     "zero returns empty",
			n:        0,
			expected: "",
		},
		{
			name:     "negative returns empty",
			n:        -5,
			expected: "",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate escape sequence
			result := ansi.MoveUp(tt.n)

			// Verify result matches expected
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMoveDown verifies downward cursor movement escape sequence generation.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func TestMoveDown(t *testing.T) {
	// Test table for various movement values
	tests := []struct {
		name     string
		n        int
		expected string
	}{
		{
			name:     "single line",
			n:        1,
			expected: "\033[1B",
		},
		{
			name:     "multiple lines",
			n:        10,
			expected: "\033[10B",
		},
		{
			name:     "zero returns empty",
			n:        0,
			expected: "",
		},
		{
			name:     "negative returns empty",
			n:        -5,
			expected: "",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate escape sequence
			result := ansi.MoveDown(tt.n)

			// Verify result matches expected
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMoveRight verifies rightward cursor movement escape sequence generation.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func TestMoveRight(t *testing.T) {
	// Test table for various movement values
	tests := []struct {
		name     string
		n        int
		expected string
	}{
		{
			name:     "single column",
			n:        1,
			expected: "\033[1C",
		},
		{
			name:     "multiple columns",
			n:        10,
			expected: "\033[10C",
		},
		{
			name:     "zero returns empty",
			n:        0,
			expected: "",
		},
		{
			name:     "negative returns empty",
			n:        -5,
			expected: "",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate escape sequence
			result := ansi.MoveRight(tt.n)

			// Verify result matches expected
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMoveLeft verifies leftward cursor movement escape sequence generation.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func TestMoveLeft(t *testing.T) {
	// Test table for various movement values
	tests := []struct {
		name     string
		n        int
		expected string
	}{
		{
			name:     "single column",
			n:        1,
			expected: "\033[1D",
		},
		{
			name:     "multiple columns",
			n:        10,
			expected: "\033[10D",
		},
		{
			name:     "zero returns empty",
			n:        0,
			expected: "",
		},
		{
			name:     "negative returns empty",
			n:        -5,
			expected: "",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate escape sequence
			result := ansi.MoveLeft(tt.n)

			// Verify result matches expected
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRGB verifies 24-bit true color foreground escape sequence generation.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func TestRGB(t *testing.T) {
	// Test table for various RGB values
	tests := []struct {
		name     string
		r        uint8
		g        uint8
		b        uint8
		expected string
	}{
		{
			name:     "black",
			r:        0,
			g:        0,
			b:        0,
			expected: "\033[38;2;0;0;0m",
		},
		{
			name:     "white",
			r:        255,
			g:        255,
			b:        255,
			expected: "\033[38;2;255;255;255m",
		},
		{
			name:     "red",
			r:        255,
			g:        0,
			b:        0,
			expected: "\033[38;2;255;0;0m",
		},
		{
			name:     "green",
			r:        0,
			g:        255,
			b:        0,
			expected: "\033[38;2;0;255;0m",
		},
		{
			name:     "blue",
			r:        0,
			g:        0,
			b:        255,
			expected: "\033[38;2;0;0;255m",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate escape sequence
			result := ansi.RGB(tt.r, tt.g, tt.b)

			// Verify result matches expected
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBgRGB verifies 24-bit true color background escape sequence generation.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func TestBgRGB(t *testing.T) {
	// Test table for various RGB values
	tests := []struct {
		name     string
		r        uint8
		g        uint8
		b        uint8
		expected string
	}{
		{
			name:     "black background",
			r:        0,
			g:        0,
			b:        0,
			expected: "\033[48;2;0;0;0m",
		},
		{
			name:     "white background",
			r:        255,
			g:        255,
			b:        255,
			expected: "\033[48;2;255;255;255m",
		},
		{
			name:     "custom color background",
			r:        128,
			g:        64,
			b:        32,
			expected: "\033[48;2;128;64;32m",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate escape sequence
			result := ansi.BgRGB(tt.r, tt.g, tt.b)

			// Verify result matches expected
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestColor256 verifies 256-color foreground escape sequence generation.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func TestColor256(t *testing.T) {
	// Test table for various color codes
	tests := []struct {
		name     string
		code     uint8
		expected string
	}{
		{
			name:     "standard black",
			code:     0,
			expected: "\033[38;5;0m",
		},
		{
			name:     "standard red",
			code:     1,
			expected: "\033[38;5;1m",
		},
		{
			name:     "color cube start",
			code:     16,
			expected: "\033[38;5;16m",
		},
		{
			name:     "grayscale start",
			code:     232,
			expected: "\033[38;5;232m",
		},
		{
			name:     "max value",
			code:     255,
			expected: "\033[38;5;255m",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate escape sequence
			result := ansi.Color256(tt.code)

			// Verify result matches expected
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBgColor256 verifies 256-color background escape sequence generation.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func TestBgColor256(t *testing.T) {
	// Test table for various color codes
	tests := []struct {
		name     string
		code     uint8
		expected string
	}{
		{
			name:     "standard black background",
			code:     0,
			expected: "\033[48;5;0m",
		},
		{
			name:     "mid grayscale background",
			code:     244,
			expected: "\033[48;5;244m",
		},
		{
			name:     "max value background",
			code:     255,
			expected: "\033[48;5;255m",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate escape sequence
			result := ansi.BgColor256(tt.code)

			// Verify result matches expected
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConstants verifies that exported constants have expected values.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func TestConstants(t *testing.T) {
	// Test table for all constant categories
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		// Text attributes
		{name: "Reset", got: ansi.Reset, expected: "\033[0m"},
		{name: "Bold", got: ansi.Bold, expected: "\033[1m"},
		{name: "Underline", got: ansi.Underline, expected: "\033[4m"},
		{name: "Dim", got: ansi.Dim, expected: "\033[2m"},
		{name: "Italic", got: ansi.Italic, expected: "\033[3m"},

		// Foreground colors
		{name: "FgBlack", got: ansi.FgBlack, expected: "\033[30m"},
		{name: "FgRed", got: ansi.FgRed, expected: "\033[31m"},
		{name: "FgGreen", got: ansi.FgGreen, expected: "\033[32m"},
		{name: "FgYellow", got: ansi.FgYellow, expected: "\033[33m"},
		{name: "FgBlue", got: ansi.FgBlue, expected: "\033[34m"},

		// Screen control
		{name: "ClearScreen", got: ansi.ClearScreen, expected: "\033[2J"},
		{name: "CursorHide", got: ansi.CursorHide, expected: "\033[?25l"},
		{name: "CursorShow", got: ansi.CursorShow, expected: "\033[?25h"},
		{name: "CursorHome", got: ansi.CursorHome, expected: "\033[H"},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify constant value matches expected
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}
