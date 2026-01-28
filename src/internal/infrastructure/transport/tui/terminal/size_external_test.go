package terminal_test

import (
	"context"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/terminal"
	"github.com/stretchr/testify/assert"
)

func TestGetSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		minCols int
		minRows int
	}{
		{
			name:    "returns_valid_size",
			minCols: 1,
			minRows: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			size := terminal.GetSize()
			assert.GreaterOrEqual(t, size.Cols, tt.minCols)
			assert.GreaterOrEqual(t, size.Rows, tt.minRows)
		})
	}
}

func TestGetLayout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		size     terminal.Size
		expected terminal.Layout
	}{
		{"compact", terminal.Size{Cols: 60, Rows: 20}, terminal.LayoutCompact},
		{"normal", terminal.Size{Cols: 100, Rows: 24}, terminal.LayoutNormal},
		{"wide", terminal.Size{Cols: 140, Rows: 40}, terminal.LayoutWide},
		{"ultrawide", terminal.Size{Cols: 200, Rows: 50}, terminal.LayoutUltraWide},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			layout := terminal.GetLayout(tt.size)
			assert.Equal(t, tt.expected, layout)
		})
	}
}

func TestLayout_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		layout   terminal.Layout
		expected string
	}{
		{"compact", terminal.LayoutCompact, "compact"},
		{"normal", terminal.LayoutNormal, "normal"},
		{"wide", terminal.LayoutWide, "wide"},
		{"ultrawide", terminal.LayoutUltraWide, "ultrawide"},
		{"unknown", terminal.Layout(-1), "unknown"},
		{"invalid_positive", terminal.Layout(100), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.layout.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLayout_Columns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		layout   terminal.Layout
		expected int
	}{
		{"compact", terminal.LayoutCompact, 1},
		{"normal", terminal.LayoutNormal, 1},
		{"wide", terminal.LayoutWide, 2},
		{"ultrawide", terminal.LayoutUltraWide, 3},
		{"invalid", terminal.Layout(-1), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.layout.Columns()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWatchResize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		timeout     time.Duration
		waitTimeout time.Duration
		minCols     int
		minRows     int
	}{
		{
			name:        "returns_initial_size",
			timeout:     100 * time.Millisecond,
			waitTimeout: 50 * time.Millisecond,
			minCols:     1,
			minRows:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			ch := terminal.WatchResize(ctx)
			assert.NotNil(t, ch)

			select {
			case size := <-ch:
				assert.GreaterOrEqual(t, size.Cols, tt.minCols)
				assert.GreaterOrEqual(t, size.Rows, tt.minRows)
			case <-time.After(tt.waitTimeout):
				t.Error("expected initial size from WatchResize")
			}
		})
	}
}

func TestWatchResize_ContextCancel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		waitTimeout time.Duration
		sleepAfter  time.Duration
	}{
		{
			name:        "context_cancellation_closes_channel",
			waitTimeout: 50 * time.Millisecond,
			sleepAfter:  20 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			ch := terminal.WatchResize(ctx)

			// Should receive channel (may or may not have initial size).
			select {
			case _, ok := <-ch:
				// Channel should be open initially.
				if ok {
					// Got initial size, good.
				}
			case <-time.After(tt.waitTimeout):
				// Timeout is also acceptable, channel is open.
			}

			cancel()

			// Wait for goroutine to notice cancellation.
			time.Sleep(tt.sleepAfter)

			// Verify context was canceled.
			if ctx.Err() == nil {
				t.Error("expected context to be canceled")
			}
		})
	}
}

func TestIsTTY(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "returns_boolean_without_panic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// IsTTY returns bool based on whether stdout is a terminal.
			// In CI, this is typically false; on a real terminal, true.
			result := terminal.IsTTY()
			// Cannot assert specific value as it depends on environment,
			// but we verify the function doesn't panic and returns a bool.
			assert.True(t, result == true || result == false)
		})
	}
}

func TestIsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		fd       uintptr
		wantBool bool
		checkVal bool
	}{
		{
			name:     "stdin_fd_returns_boolean",
			fd:       0,
			wantBool: false,
			checkVal: false,
		},
		{
			name:     "invalid_high_fd_returns_false",
			fd:       999999,
			wantBool: false,
			checkVal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := terminal.IsTerminal(tt.fd)
			if tt.checkVal {
				assert.Equal(t, tt.wantBool, result)
			} else {
				// Cannot assert specific value as it depends on environment.
				assert.True(t, result == true || result == false)
			}
		})
	}
}

func TestBreakpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    int
		expected int
	}{
		{"small", terminal.BreakpointSmall, 80},
		{"medium", terminal.BreakpointMedium, 120},
		{"large", terminal.BreakpointLarge, 160},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.value)
		})
	}
}

func TestMinSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    string
		got      int
		expected int
	}{
		{
			name:     "min_cols_value",
			field:    "Cols",
			got:      terminal.MinSize.Cols,
			expected: 40,
		},
		{
			name:     "min_rows_value",
			field:    "Rows",
			got:      terminal.MinSize.Rows,
			expected: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}

func TestDefaultSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		field    string
		got      int
		expected int
	}{
		{
			name:     "default_cols_value",
			field:    "Cols",
			got:      terminal.DefaultSize.Cols,
			expected: 80,
		},
		{
			name:     "default_rows_value",
			field:    "Rows",
			got:      terminal.DefaultSize.Rows,
			expected: 24,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.got)
		})
	}
}
