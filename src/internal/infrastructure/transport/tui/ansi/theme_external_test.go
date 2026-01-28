// Package ansi_test provides black-box tests for ANSI escape sequences.
package ansi_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
)

func TestDefaultTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{"returns theme"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			theme := ansi.DefaultTheme()
			if theme.Primary == "" {
				t.Error("Primary empty")
			}
		})
	}
}

func TestTrueColorTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{"returns RGB theme"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			theme := ansi.TrueColorTheme()
			if theme.Primary == "" {
				t.Error("Primary empty")
			}
		})
	}
}

func TestColorize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		color string
		text  string
		want  string
	}{
		{"empty color", "", "hello", "hello"},
		{"with color", ansi.FgRed, "error", ansi.FgRed + "error" + ansi.Reset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ansi.Colorize(tt.color, tt.text); got != tt.want {
				t.Errorf("Colorize(%q, %q) = %q; want %q", tt.color, tt.text, got, tt.want)
			}
		})
	}
}

func TestBoldText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want string
	}{
		{"basic", "hello", ansi.Bold + "hello" + ansi.Reset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ansi.BoldText(tt.text); got != tt.want {
				t.Errorf("BoldText(%q) = %q; want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestDimText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
		want string
	}{
		{"basic", "hello", ansi.Dim + "hello" + ansi.Reset},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := ansi.DimText(tt.text); got != tt.want {
				t.Errorf("DimText(%q) = %q; want %q", tt.text, got, tt.want)
			}
		})
	}
}
