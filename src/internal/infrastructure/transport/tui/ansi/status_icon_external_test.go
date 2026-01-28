// Package ansi_test provides external tests for the ansi package.
package ansi_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
)

func TestDefaultIcons(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"returns icons"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icons := ansi.DefaultIcons()
			if icons.Running == "" {
				t.Error("Running icon empty")
			}
		})
	}
}

func TestASCIIIcons(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"returns ASCII icons"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			icons := ansi.ASCIIIcons()
			if icons.Running == "" {
				t.Error("Running icon empty")
			}
		})
	}
}
