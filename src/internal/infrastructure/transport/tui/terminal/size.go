// Package terminal provides terminal handling utilities.
// Pure Go implementation using syscalls, no exec.Command.
package terminal

import (
	"os"
	"os/signal"
	"syscall"
)

// Size represents terminal dimensions.
type Size struct {
	// Cols is the number of columns (width).
	Cols int
	// Rows is the number of rows (height).
	Rows int
}

// MinSize is the minimum usable terminal size.
var MinSize = Size{Cols: 40, Rows: 10}

// DefaultSize is used when detection fails.
var DefaultSize = Size{Cols: 80, Rows: 24}

// Breakpoints for responsive layout.
const (
	// BreakpointSmall is for terminals < 80 cols.
	BreakpointSmall = 80
	// BreakpointMedium is for terminals 80-120 cols.
	BreakpointMedium = 120
	// BreakpointLarge is for terminals > 120 cols.
	BreakpointLarge = 160
)

// Layout represents the current layout mode.
type Layout int

// Layout constants.
const (
	LayoutCompact   Layout = iota // < 80 cols: single column, minimal
	LayoutNormal                  // 80-120 cols: single column, full info
	LayoutWide                    // 120-160 cols: two columns
	LayoutUltraWide               // > 160 cols: three columns
)

// GetSize returns the current terminal size.
// Falls back to DefaultSize if detection fails.
func GetSize() Size {
	size, err := getTerminalSize()
	if err != nil {
		return DefaultSize
	}
	return size
}

// GetLayout returns the appropriate layout for the given size.
func GetLayout(size Size) Layout {
	switch {
	case size.Cols < BreakpointSmall:
		return LayoutCompact
	case size.Cols < BreakpointMedium:
		return LayoutNormal
	case size.Cols < BreakpointLarge:
		return LayoutWide
	default:
		return LayoutUltraWide
	}
}

// String returns the layout name.
func (l Layout) String() string {
	switch l {
	case LayoutCompact:
		return "compact"
	case LayoutNormal:
		return "normal"
	case LayoutWide:
		return "wide"
	case LayoutUltraWide:
		return "ultrawide"
	default:
		return "unknown"
	}
}

// Columns returns the number of content columns for this layout.
func (l Layout) Columns() int {
	switch l {
	case LayoutCompact, LayoutNormal:
		return 1
	case LayoutWide:
		return 2
	case LayoutUltraWide:
		return 3
	default:
		return 1
	}
}

// WatchResize creates a channel that receives size updates on terminal resize.
// The channel is closed when the context is cancelled.
func WatchResize() <-chan Size {
	ch := make(chan Size, 1)

	// Send initial size.
	ch <- GetSize()

	// Watch for SIGWINCH.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)

	go func() {
		defer close(ch)
		defer signal.Stop(sigCh)

		for range sigCh {
			select {
			case ch <- GetSize():
			default:
				// Channel full, skip this update.
			}
		}
	}()

	return ch
}

// IsTerminal returns true if the given file descriptor is a terminal.
func IsTerminal(fd uintptr) bool {
	return isTerminal(fd)
}

// IsTTY returns true if stdout is a terminal.
func IsTTY() bool {
	return IsTerminal(os.Stdout.Fd())
}
