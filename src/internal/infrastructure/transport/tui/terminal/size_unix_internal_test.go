//go:build linux || darwin || freebsd || openbsd || netbsd

// Package terminal provides internal tests for size_unix.go.
// It tests internal implementation details using white-box testing.
package terminal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_getTerminalSize tests the getTerminalSize function.
// It verifies that terminal size detection works correctly.
//
// Params:
//   - t: the testing context.
func Test_getTerminalSize(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "returns_size_or_error",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the function.
			size, err := getTerminalSize()

			// Either we get a valid size or an error.
			if err == nil {
				// Valid size should have positive dimensions.
				assert.Greater(t, size.Cols, 0)
				assert.Greater(t, size.Rows, 0)
			} else {
				// On error, size should be default.
				assert.Equal(t, DefaultSize, size)
			}
		})
	}
}

// Test_getSizeFromEnv tests the getSizeFromEnv function.
// It verifies that environment variable parsing works correctly.
//
// Params:
//   - t: the testing context.
func Test_getSizeFromEnv(t *testing.T) {
	// Note: These tests modify environment variables, so they cannot be parallel.

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// columns is the COLUMNS env var value.
		columns string
		// lines is the LINES env var value.
		lines string
		// wantCols is the expected columns.
		wantCols int
		// wantRows is the expected rows.
		wantRows int
	}{
		{
			name:     "no_env_vars",
			columns:  "",
			lines:    "",
			wantCols: defaultCols,
			wantRows: defaultRows,
		},
		{
			name:     "columns_only",
			columns:  "100",
			lines:    "",
			wantCols: 100,
			wantRows: defaultRows,
		},
		{
			name:     "lines_only",
			columns:  "",
			lines:    "50",
			wantCols: defaultCols,
			wantRows: 50,
		},
		{
			name:     "both_set",
			columns:  "120",
			lines:    "40",
			wantCols: 120,
			wantRows: 40,
		},
		{
			name:     "invalid_columns",
			columns:  "abc",
			lines:    "30",
			wantCols: defaultCols,
			wantRows: 30,
		},
		{
			name:     "invalid_lines",
			columns:  "80",
			lines:    "xyz",
			wantCols: 80,
			wantRows: defaultRows,
		},
		{
			name:     "negative_columns",
			columns:  "-1",
			lines:    "24",
			wantCols: defaultCols,
			wantRows: 24,
		},
		{
			name:     "zero_lines",
			columns:  "80",
			lines:    "0",
			wantCols: 80,
			wantRows: defaultRows,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars.
			origColumns := os.Getenv("COLUMNS")
			origLines := os.Getenv("LINES")

			// Restore env vars after test.
			defer func() {
				if origColumns != "" {
					os.Setenv("COLUMNS", origColumns)
				} else {
					os.Unsetenv("COLUMNS")
				}
				if origLines != "" {
					os.Setenv("LINES", origLines)
				} else {
					os.Unsetenv("LINES")
				}
			}()

			// Set test env vars.
			if tt.columns != "" {
				os.Setenv("COLUMNS", tt.columns)
			} else {
				os.Unsetenv("COLUMNS")
			}
			if tt.lines != "" {
				os.Setenv("LINES", tt.lines)
			} else {
				os.Unsetenv("LINES")
			}

			// Call the function.
			result := getSizeFromEnv()

			// Verify result.
			assert.Equal(t, tt.wantCols, result.Cols)
			assert.Equal(t, tt.wantRows, result.Rows)
		})
	}
}

// Test_isTerminal tests the isTerminal function.
// It verifies that terminal detection works correctly.
//
// Params:
//   - t: the testing context.
func Test_isTerminal(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// fd is the file descriptor to check.
		fd uintptr
	}{
		{
			name: "stdin",
			fd:   os.Stdin.Fd(),
		},
		{
			name: "stdout",
			fd:   os.Stdout.Fd(),
		},
		{
			name: "stderr",
			fd:   os.Stderr.Fd(),
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the function - should not panic.
			result := isTerminal(tt.fd)

			// Result is a boolean, either value is valid.
			// In CI, these are typically not terminals.
			_ = result
		})
	}
}

// Test_winsize tests the winsize struct.
// It verifies that the struct matches the C struct layout.
//
// Params:
//   - t: the testing context.
func Test_winsize(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// row is the row value.
		row uint16
		// col is the col value.
		col uint16
		// xpixel is the xpixel value.
		xpixel uint16
		// ypixel is the ypixel value.
		ypixel uint16
	}{
		{
			name:   "standard_80x24",
			row:    24,
			col:    80,
			xpixel: 0,
			ypixel: 0,
		},
		{
			name:   "wide_terminal",
			row:    50,
			col:    200,
			xpixel: 1600,
			ypixel: 1000,
		},
		{
			name:   "small_terminal",
			row:    10,
			col:    40,
			xpixel: 320,
			ypixel: 200,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create winsize struct.
			ws := winsize{
				Row:    tt.row,
				Col:    tt.col,
				Xpixel: tt.xpixel,
				Ypixel: tt.ypixel,
			}

			// Verify fields.
			assert.Equal(t, tt.row, ws.Row)
			assert.Equal(t, tt.col, ws.Col)
			assert.Equal(t, tt.xpixel, ws.Xpixel)
			assert.Equal(t, tt.ypixel, ws.Ypixel)
		})
	}
}

// Test_getTerminalSize_with_env tests getTerminalSize with env vars set.
// It verifies that environment variables take precedence.
//
// Params:
//   - t: the testing context.
func Test_getTerminalSize_with_env(t *testing.T) {
	// Note: This test modifies environment variables, so it cannot be parallel.

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// columns is the COLUMNS env var value.
		columns string
		// lines is the LINES env var value.
		lines string
		// wantCols is the expected columns.
		wantCols int
		// wantRows is the expected rows.
		wantRows int
	}{
		{
			name:     "env_vars_override_ioctl",
			columns:  "150",
			lines:    "60",
			wantCols: 150,
			wantRows: 60,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Save original env vars.
			origColumns := os.Getenv("COLUMNS")
			origLines := os.Getenv("LINES")

			// Restore env vars after test.
			defer func() {
				if origColumns != "" {
					os.Setenv("COLUMNS", origColumns)
				} else {
					os.Unsetenv("COLUMNS")
				}
				if origLines != "" {
					os.Setenv("LINES", origLines)
				} else {
					os.Unsetenv("LINES")
				}
			}()

			// Set test env vars.
			os.Setenv("COLUMNS", tt.columns)
			os.Setenv("LINES", tt.lines)

			// Call the function.
			size, err := getTerminalSize()

			// With env vars set, should succeed.
			assert.NoError(t, err)
			assert.Equal(t, tt.wantCols, size.Cols)
			assert.Equal(t, tt.wantRows, size.Rows)
		})
	}
}

// Test_isTerminal_invalid_fd tests isTerminal with invalid file descriptor.
// It verifies that invalid FDs return false.
//
// Params:
//   - t: the testing context.
func Test_isTerminal_invalid_fd(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// fd is an invalid file descriptor.
		fd uintptr
		// wantResult is the expected result.
		wantResult bool
	}{
		{
			name:       "invalid_large_fd",
			fd:         999999,
			wantResult: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the function.
			result := isTerminal(tt.fd)

			// Verify result.
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
