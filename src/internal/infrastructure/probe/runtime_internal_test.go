//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuntimeInfoInternal(t *testing.T) {
	tests := []struct {
		name string
		info *RuntimeInfo
	}{
		{
			name: "EmptyRuntimeInfo",
			info: &RuntimeInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.info)
		})
	}
}
