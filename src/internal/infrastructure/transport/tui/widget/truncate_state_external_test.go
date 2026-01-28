package widget_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
	"github.com/stretchr/testify/assert"
)

func TestTruncateVisible(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		maxLen int
	}{
		{"short_string", "hello", 10},
		{"exact_length", "hello", 5},
		{"truncated", "hello world", 5},
		{"empty_string", "", 5},
		{"zero_max", "hello", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := widget.TruncateVisible(tt.input, tt.maxLen)
			if tt.maxLen <= 0 {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
			}
		})
	}
}
