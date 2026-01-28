// Package terminal provides terminal handling utilities.
// Pure Go implementation using syscalls, no exec.Command.
package terminal

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

const (
	// minCols is the minimum number of columns for a usable terminal.
	minCols int = 40
	// minRows is the minimum number of rows for a usable terminal.
	minRows int = 10
	// defaultCols is the default number of columns when detection fails.
	defaultCols int = 80
	// defaultRows is the default number of rows when detection fails.
	defaultRows int = 24
	// columnsWide is the number of columns for wide layout (two columns).
	columnsWide int = 2
	// columnsUltraWide is the number of columns for ultrawide layout (three columns).
	columnsUltraWide int = 3

	// BreakpointSmall is for terminals < 80 cols.
	BreakpointSmall int = 80
	// BreakpointMedium is for terminals 80-120 cols.
	BreakpointMedium int = 120
	// BreakpointLarge is for terminals > 120 cols.
	BreakpointLarge int = 160
)

// Size represents terminal dimensions.
// It holds the width (columns) and height (rows) of the terminal.
type Size struct {
	// Cols is the number of columns (width).
	Cols int
	// Rows is the number of rows (height).
	Rows int
}

// Layout represents the current layout mode.
type Layout int

// Layout constants.
const (
	LayoutCompact   Layout = iota // < 80 cols: single column, minimal
	LayoutNormal                  // 80-120 cols: single column, full info
	LayoutWide                    // 120-160 cols: two columns
	LayoutUltraWide               // > 160 cols: three columns
)

var (
	// MinSize is the minimum usable terminal size.
	MinSize Size = Size{Cols: minCols, Rows: minRows}

	// DefaultSize is used when detection fails.
	DefaultSize Size = Size{Cols: defaultCols, Rows: defaultRows}
)

// GetSize returns the current terminal size.
// Falls back to DefaultSize if detection fails.
//
// Returns:
//   - Size: detected terminal dimensions or DefaultSize on error
func GetSize() Size {
	size, err := getTerminalSize()
	// Fall back to default size if detection fails.
	if err != nil {
		// Return default dimensions.
		return DefaultSize
	}
	// Return detected size.
	return size
}

// GetLayout returns the appropriate layout for the given size.
//
// Params:
//   - size: terminal dimensions to determine layout
//
// Returns:
//   - Layout: appropriate layout mode for the given terminal size
func GetLayout(size Size) Layout {
	// Determine layout based on column width.
	switch {
	// Compact layout for narrow terminals.
	case size.Cols < BreakpointSmall:
		// Single column, minimal info.
		return LayoutCompact
	// Normal layout for standard terminals.
	case size.Cols < BreakpointMedium:
		// Single column, full info.
		return LayoutNormal
	// Wide layout for larger terminals.
	case size.Cols < BreakpointLarge:
		// Two columns.
		return LayoutWide
	// Ultrawide layout for very large terminals.
	default:
		// Three columns.
		return LayoutUltraWide
	}
}

// String returns the layout name.
//
// Returns:
//   - string: human-readable layout name
func (l Layout) String() string {
	// Convert layout enum to string.
	switch l {
	// Compact layout name.
	case LayoutCompact:
		// Narrow terminals.
		return "compact"
	// Normal layout name.
	case LayoutNormal:
		// Standard terminals.
		return "normal"
	// Wide layout name.
	case LayoutWide:
		// Two-column terminals.
		return "wide"
	// Ultrawide layout name.
	case LayoutUltraWide:
		// Three-column terminals.
		return "ultrawide"
	// Unknown layout name.
	case Layout(-1):
		// Invalid layout value.
		return "unknown"
	// Default fallback.
	default:
		// Unexpected layout value.
		return "unknown"
	}
}

// Columns returns the number of content columns for this layout.
//
// Returns:
//   - int: number of columns for this layout (1, 2, or 3)
func (l Layout) Columns() int {
	// Determine column count based on layout.
	switch l {
	// Single column layouts.
	case LayoutCompact, LayoutNormal:
		// One column for compact and normal.
		return 1
	// Two column layout.
	case LayoutWide:
		// Two columns for wide.
		return columnsWide
	// Three column layout.
	case LayoutUltraWide:
		// Three columns for ultrawide.
		return columnsUltraWide
	// Default single column.
	default:
		// Fallback to single column.
		return 1
	}
}

// WatchResize creates a channel that receives size updates on terminal resize.
// The channel is closed when the context is cancelled.
//
// Params:
//   - ctx: context to control the lifecycle of the watcher
//
// Returns:
//   - <-chan Size: channel receiving terminal size updates on SIGWINCH
//
// Lifecycle:
//   - Spawns a goroutine that listens for SIGWINCH signals.
//   - The goroutine exits when ctx is cancelled (ctx.Done()).
//   - Cleanup: signal.Stop() and close(ch) are called via defer.
func WatchResize(ctx context.Context) <-chan Size {
	ch := make(chan Size, 1)

	// Send initial size.
	ch <- GetSize()

	// Watch for SIGWINCH.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)

	go func() {
		defer close(ch)
		defer signal.Stop(sigCh)

		// Listen for resize signals or context cancellation.
		for {
			select {
			// Context cancelled, stop watching.
			case <-ctx.Done():
				// Exit goroutine.
				return
			// Received resize signal.
			case <-sigCh:
				select {
				case ch <- GetSize():
				default:
					// Channel full, skip this update.
				}
			}
		}
	}()

	// Return read-only channel.
	return ch
}

// IsTerminal returns true if the given file descriptor is a terminal.
//
// Params:
//   - fd: file descriptor to check
//
// Returns:
//   - bool: true if fd is a terminal, false otherwise
func IsTerminal(fd uintptr) bool {
	// Delegate to platform-specific implementation.
	return isTerminal(fd)
}

// IsTTY returns true if stdout is a terminal.
//
// Returns:
//   - bool: true if stdout is a terminal, false otherwise
func IsTTY() bool {
	// Check if stdout is a terminal.
	return IsTerminal(os.Stdout.Fd())
}
