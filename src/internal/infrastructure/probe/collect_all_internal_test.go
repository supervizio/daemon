//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullPercentageConstant(t *testing.T) {
	tests := []struct {
		name string
		want float64
	}{
		{
			name: "FullPercentageIs100",
			want: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fullPercentage)
		})
	}
}
