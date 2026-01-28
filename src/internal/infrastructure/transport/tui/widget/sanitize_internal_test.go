package widget

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAnsiEscapeRegex verifies the regex is properly initialized and matches patterns.
func TestAnsiEscapeRegex(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         string
		expectedCount int
	}{
		{
			name:          "initialized not nil",
			input:         "",
			expectedCount: 0,
		},
		{
			name:          "matches basic sequence",
			input:         "\x1b[31m\x1b[0m",
			expectedCount: 2,
		},
		{
			name:          "no match plain text",
			input:         "hello world",
			expectedCount: 0,
		},
		{
			name:          "matches cursor movement",
			input:         "\x1b[H\x1b[2J",
			expectedCount: 2,
		},
		{
			name:          "matches with parameters",
			input:         "\x1b[38;5;196m",
			expectedCount: 1,
		},
	}

	// Verify regex is initialized.
	assert.NotNil(t, ansiEscapeRegex)

	// Run each test case.
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			matches := ansiEscapeRegex.FindAllString(tc.input, -1)

			assert.Len(t, matches, tc.expectedCount)
		})
	}
}
