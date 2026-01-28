// Package ansi provides ANSI escape sequences for terminal styling.
package ansi

import "testing"

// Test_codes_constants verifies that ANSI constants are properly defined.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none: test function
func Test_codes_constants(t *testing.T) {
	t.Parallel()

	// Test table for constant validation
	tests := []struct {
		name  string
		value string
	}{
		{name: "Reset constant defined", value: Reset},
		{name: "Bold constant defined", value: Bold},
		{name: "ClearScreen constant defined", value: ClearScreen},
		{name: "Dim constant defined", value: Dim},
		{name: "Italic constant defined", value: Italic},
		{name: "Underline constant defined", value: Underline},
		{name: "FgBlack constant defined", value: FgBlack},
		{name: "FgRed constant defined", value: FgRed},
		{name: "FgGreen constant defined", value: FgGreen},
		{name: "CursorHide constant defined", value: CursorHide},
		{name: "CursorShow constant defined", value: CursorShow},
		{name: "CursorHome constant defined", value: CursorHome},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify constant is not empty
			if tt.value == "" {
				t.Errorf("%s should not be empty", tt.name)
			}
		})
	}
}
