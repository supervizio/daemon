//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuntimeTypeInternal(t *testing.T) {
	tests := []struct {
		name     string
		rt       RuntimeType
		wantName string
	}{
		{
			name:     "RuntimeNoneValue",
			rt:       RuntimeNone,
			wantName: "none",
		},
		{
			name:     "RuntimeUnknownValue",
			rt:       RuntimeUnknown,
			wantName: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runtimeNames[tt.rt]
			assert.Equal(t, tt.wantName, got)
		})
	}
}
