package widget_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
	"github.com/stretchr/testify/assert"
)

// TestStripANSI verifies ANSI escape sequence removal.
func TestStripANSI(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "no escape sequences",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "clear screen sequence",
			input:    "\x1b[2Jmalicious",
			expected: "malicious",
		},
		{
			name:     "cursor move sequence",
			input:    "\x1b[Htext",
			expected: "text",
		},
		{
			name:     "color codes",
			input:    "\x1b[31mred\x1b[0m",
			expected: "red",
		},
		{
			name:     "complex sequence with multiple codes",
			input:    "\x1b[2J\x1b[H\x1b[31mmalicious\x1b[0m",
			expected: "malicious",
		},
		{
			name:     "mixed content",
			input:    "before\x1b[31mcolored\x1b[0mafter",
			expected: "beforecoloredafter",
		},
		{
			name:     "unicode preserved",
			input:    "hello 世界 🌍",
			expected: "hello 世界 🌍",
		},
		{
			name:     "256 color code",
			input:    "\x1b[38;5;196mred\x1b[0m",
			expected: "red",
		},
		{
			name:     "cursor position",
			input:    "\x1b[10;20H",
			expected: "",
		},
		{
			name:     "bold and underline",
			input:    "\x1b[1m\x1b[4mbold underline\x1b[0m",
			expected: "bold underline",
		},
	}

	// Run each test case.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := widget.StripANSI(tc.input)

			assert.Equal(t, tc.expected, result)
		})
	}
}
