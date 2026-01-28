// Package tui_test provides external tests.
package tui_test

import (
	"testing"
)

func TestTuiSnapshotDataCompiles(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "package compiles"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compilation test
		})
	}
}
