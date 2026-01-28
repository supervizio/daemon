package widget_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
	"github.com/stretchr/testify/assert"
)

func TestSpaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		n        int
		expected int
	}{
		{
			name:     "zero_spaces",
			n:        0,
			expected: 0,
		},
		{
			name:     "positive_spaces",
			n:        5,
			expected: 5,
		},
		{
			name:     "large_spaces",
			n:        300,
			expected: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.Spaces(tt.n)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestHorizontalBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		n        int
		expected int
	}{
		{
			name:     "zero_bars",
			n:        0,
			expected: 0,
		},
		{
			name:     "positive_bars",
			n:        5,
			expected: 5,
		},
		{
			name:     "large_bars",
			n:        300,
			expected: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.HorizontalBar(tt.n)
			assert.Len(t, []rune(result), tt.expected)
		})
	}
}

func TestPad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		text  string
		width int
		align widget.Align
	}{
		{
			name:  "pad_left",
			text:  "test",
			width: 10,
			align: widget.AlignLeft,
		},
		{
			name:  "pad_right",
			text:  "test",
			width: 10,
			align: widget.AlignRight,
		},
		{
			name:  "pad_center",
			text:  "test",
			width: 10,
			align: widget.AlignCenter,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.Pad(tt.text, tt.width, tt.align)
			assert.NotEmpty(t, result)
		})
	}
}

func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		text   string
		maxLen int
	}{
		{
			name:   "short_text",
			text:   "test",
			maxLen: 10,
		},
		{
			name:   "long_text",
			text:   "very long text that needs truncation",
			maxLen: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.Truncate(tt.text, tt.maxLen)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "seconds",
			duration: 30 * time.Second,
		},
		{
			name:     "minutes",
			duration: 5 * time.Minute,
		},
		{
			name:     "hours",
			duration: 2 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.FormatDuration(tt.duration)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormatDurationShort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "seconds",
			duration: 30 * time.Second,
		},
		{
			name:     "minutes",
			duration: 5 * time.Minute,
		},
		{
			name:     "hours",
			duration: 2 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.FormatDurationShort(tt.duration)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		bytes uint64
	}{
		{
			name:  "small_bytes",
			bytes: 512,
		},
		{
			name:  "kilobytes",
			bytes: 1024 * 10,
		},
		{
			name:  "megabytes",
			bytes: 1024 * 1024 * 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.FormatBytes(tt.bytes)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormatBytesShort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		bytes uint64
	}{
		{
			name:  "small_bytes",
			bytes: 512,
		},
		{
			name:  "kilobytes",
			bytes: 1024 * 10,
		},
		{
			name:  "megabytes",
			bytes: 1024 * 1024 * 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.FormatBytesShort(tt.bytes)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormatBytesPerSec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		bytes uint64
	}{
		{
			name:  "kilobytes_per_sec",
			bytes: 1024 * 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.FormatBytesPerSec(tt.bytes)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormatPercent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		percent float64
	}{
		{
			name:    "small_percent",
			percent: 5.5,
		},
		{
			name:    "large_percent",
			percent: 95.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.FormatPercent(tt.percent)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormatSpeed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		bitsPerSec uint64
	}{
		{
			name:       "kilobits",
			bitsPerSec: 1000 * 10,
		},
		{
			name:       "megabits",
			bitsPerSec: 1000 * 1000 * 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.FormatSpeed(tt.bitsPerSec)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormatSpeedShort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		bitsPerSec uint64
	}{
		{
			name:       "kilobits",
			bitsPerSec: 1000 * 10,
		},
		{
			name:       "megabits",
			bitsPerSec: 1000 * 1000 * 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.FormatSpeedShort(tt.bitsPerSec)
			assert.NotEmpty(t, result)
		})
	}
}

func TestRepeatString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		s      string
		n      int
		expect int
	}{
		{
			name:   "repeat_char",
			s:      "a",
			n:      5,
			expect: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.RepeatString(tt.s, tt.n)
			assert.Len(t, result, tt.expect)
		})
	}
}

func TestTruncateRunes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		s        string
		maxRunes int
		suffix   string
	}{
		{
			name:     "truncate_with_suffix",
			s:        "hello world",
			maxRunes: 8,
			suffix:   "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.TruncateRunes(tt.s, tt.maxRunes, tt.suffix)
			assert.NotEmpty(t, result)
		})
	}
}

func TestPadRight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		s     string
		width int
	}{
		{
			name:  "pad_needed",
			s:     "test",
			width: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.PadRight(tt.s, tt.width)
			assert.GreaterOrEqual(t, len(result), len(tt.s))
		})
	}
}

func TestPadLeft(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		s     string
		width int
	}{
		{
			name:  "pad_needed",
			s:     "test",
			width: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.PadLeft(tt.s, tt.width)
			assert.GreaterOrEqual(t, len(result), len(tt.s))
		})
	}
}

func TestPadRightAnsi(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		s     string
		width int
	}{
		{
			name:  "pad_ansi_string",
			s:     "\x1b[32mgreen\x1b[0m",
			width: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.PadRightAnsi(tt.s, tt.width)
			assert.NotEmpty(t, result)
		})
	}
}

func TestPadLeftAnsi(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		s     string
		width int
	}{
		{
			name:  "pad_ansi_string",
			s:     "\x1b[32mgreen\x1b[0m",
			width: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.PadLeftAnsi(tt.s, tt.width)
			assert.NotEmpty(t, result)
		})
	}
}

func TestJoinWithSep(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		sep   string
		parts []string
	}{
		{
			name:  "join_strings",
			sep:   ", ",
			parts: []string{"a", "b", "c"},
		},
		{
			name:  "skip_empty",
			sep:   ", ",
			parts: []string{"a", "", "c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.JoinWithSep(tt.sep, tt.parts...)
			assert.NotEmpty(t, result)
		})
	}
}
