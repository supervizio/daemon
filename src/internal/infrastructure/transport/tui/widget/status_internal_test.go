// Package widget provides internal tests for the status indicator.
package widget

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/stretchr/testify/assert"
)

// TestStateColorText tests the stateColorText helper function.
func TestStateColorText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		state         process.State
		expectedText  string
		expectedShort string
	}{
		{
			name:          "running state returns success color and text",
			state:         process.StateRunning,
			expectedText:  "running",
			expectedShort: "run",
		},
		{
			name:          "starting state returns warning color and text",
			state:         process.StateStarting,
			expectedText:  "starting",
			expectedShort: "start",
		},
		{
			name:          "stopped state returns muted color and text",
			state:         process.StateStopped,
			expectedText:  "stopped",
			expectedShort: "stop",
		},
		{
			name:          "stopping state returns warning color and text",
			state:         process.StateStopping,
			expectedText:  "stopping",
			expectedShort: "stopping",
		},
		{
			name:          "failed state returns error color and text",
			state:         process.StateFailed,
			expectedText:  "failed",
			expectedShort: "fail",
		},
		{
			name:          "unknown state returns muted color and text",
			state:         process.State(99),
			expectedText:  "unknown",
			expectedShort: "?",
		},
	}

	theme := ansi.DefaultTheme()

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			colorFn, text, short := stateColorText(tc.state)

			assert.Equal(t, tc.expectedText, text)
			assert.Equal(t, tc.expectedShort, short)
			assert.NotNil(t, colorFn)

			// Verify color function returns valid theme color.
			color := colorFn(&theme)
			assert.NotEmpty(t, color)
		})
	}
}
