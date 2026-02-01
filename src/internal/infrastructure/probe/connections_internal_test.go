//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFoundPIDConstant(t *testing.T) {
	tests := []struct {
		name string
		want int32
	}{
		{
			name: "NotFoundPIDIsNegativeOne",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, notFoundPID)
		})
	}
}
