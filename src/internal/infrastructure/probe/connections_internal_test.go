//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectionCollector_NewConnectionCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsNonNilCollector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewConnectionCollector()
			assert.NotNil(t, collector)
		})
	}
}

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
