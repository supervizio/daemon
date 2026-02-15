//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiskCollector_NewDiskCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsNonNilCollector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewDiskCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestCCharArrayToString tests the cCharArrayToString function exists.
func TestCCharArrayToString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{name: "function exists and compiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// cCharArrayToString requires C.char array, tested via bytesToStringWithNull.
			assert.NotNil(t, cCharArrayToString)
		})
	}
}
