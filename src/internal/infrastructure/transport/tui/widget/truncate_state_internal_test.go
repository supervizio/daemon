package widget

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_truncateState_processRune(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		maxLen     int
		input      string
		wantStop   bool
		wantOutput string
	}{
		{"normal_within_limit", 5, "abc", false, "abc"},
		{"reaches_limit", 2, "abc", true, "ab"},
		{"exact_limit", 3, "abc", false, "abc"},
		{"single_char", 1, "x", false, "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var result strings.Builder
			state := truncateState{maxLen: tt.maxLen}
			var stopped bool
			for _, r := range tt.input {
				if state.processRune(&result, r) {
					stopped = true
					break
				}
			}
			assert.Equal(t, tt.wantStop, stopped)
			assert.Equal(t, tt.wantOutput, result.String())
		})
	}
}

func Test_truncateVisible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		maxLen      int
		wantContain string
	}{
		{"short_string", "hello", 10, "hello"},
		{"exact_length", "hello", 5, "hello"},
		{"truncated", "hello world", 5, "hello"},
		{"zero_length", "hello", 0, ""},
		{"negative_length", "hello", -1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := truncateVisible(tt.input, tt.maxLen)
			if tt.maxLen > 0 {
				assert.Contains(t, result, tt.wantContain)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func Test_isEscapeTerminator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		r        rune
		expected bool
	}{
		{"lowercase_m", 'm', true},
		{"uppercase_A", 'A', true},
		{"lowercase_z", 'z', true},
		{"digit", '0', false},
		{"semicolon", ';', false},
		{"bracket", '[', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, isEscapeTerminator(tt.r))
		})
	}
}
