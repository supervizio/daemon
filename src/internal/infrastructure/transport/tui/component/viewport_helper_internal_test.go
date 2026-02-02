// Package component provides internal tests for viewport helper functions.
package component

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/stretchr/testify/assert"
)

// TestHandleViewportKeyMsg tests the handleViewportKeyMsg function.
func TestHandleViewportKeyMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{name: "home key", key: "home"},
		{name: "end key", key: "end"},
		{name: "g key", key: "g"},
		{name: "G key", key: "G"},
		{name: "up key", key: "up"},
		{name: "down key", key: "down"},
		{name: "k key", key: "k"},
		{name: "j key", key: "j"},
		{name: "pgup key", key: "pgup"},
		{name: "pgdown key", key: "pgdown"},
		{name: "other key", key: "x"},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			vp := viewport.New(80, 24)
			vp.SetContent("line1\nline2\nline3\nline4\nline5")
			msg := mockKeyMsg{str: tc.key}
			cmd := handleViewportKeyMsg(&vp, &vp, msg)
			// Verify function completes without error.
			_ = cmd
		})
	}
}

// TestRenderContentLinesWithScrollbar tests the renderContentLinesWithScrollbar function.
func TestRenderContentLinesWithScrollbar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params ScrollbarParams
	}{
		{
			name: "few lines no scroll",
			params: ScrollbarParams{
				Lines:          []string{"line1", "line2", "line3"},
				ScrollbarChars: []string{"█", "█", "█"},
				Height:         5,
				InnerWidth:     10,
				BorderColor:    "",
				TrackChar:      "│",
			},
		},
		{
			name: "many lines with scroll",
			params: ScrollbarParams{
				Lines:          []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"},
				ScrollbarChars: []string{"█", "█", "█", "█", "█"},
				Height:         5,
				InnerWidth:     20,
				BorderColor:    "\033[36m",
				TrackChar:      "│",
			},
		},
		{
			name: "empty content",
			params: ScrollbarParams{
				Lines:          []string{},
				ScrollbarChars: []string{},
				Height:         3,
				InnerWidth:     10,
				BorderColor:    "",
				TrackChar:      "│",
			},
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var sb strings.Builder
			renderContentLinesWithScrollbar(&sb, tc.params)
			result := sb.String()
			assert.NotEmpty(t, result)
		})
	}
}
