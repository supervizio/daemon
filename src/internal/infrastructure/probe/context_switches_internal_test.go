//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContextSwitches_Structure verifies ContextSwitches struct fields.
func TestContextSwitches_Structure(t *testing.T) {
	tests := []struct {
		name        string
		voluntary   uint64
		involuntary uint64
		systemTotal uint64
	}{
		{
			name:        "with values",
			voluntary:   100,
			involuntary: 50,
			systemTotal: 1000,
		},
		{
			name:        "zero values",
			voluntary:   0,
			involuntary: 0,
			systemTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cs := ContextSwitches{
				Voluntary:   tt.voluntary,
				Involuntary: tt.involuntary,
				SystemTotal: tt.systemTotal,
			}

			assert.Equal(t, tt.voluntary, cs.Voluntary)
			assert.Equal(t, tt.involuntary, cs.Involuntary)
			assert.Equal(t, tt.systemTotal, cs.SystemTotal)
		})
	}
}
