//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextSwitchesInternal(t *testing.T) {
	tests := []struct {
		name     string
		switches *ContextSwitches
	}{
		{
			name:     "EmptyContextSwitches",
			switches: &ContextSwitches{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.switches)
		})
	}
}
