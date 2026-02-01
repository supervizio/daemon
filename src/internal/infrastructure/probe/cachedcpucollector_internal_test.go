//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCachedCPUCollector_NewCachedCPUCollector_Internal(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsNonNilCollector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewCachedCPUCollector()
			assert.NotNil(t, collector)
		})
	}
}

func TestFullPercentCacheConstant(t *testing.T) {
	tests := []struct {
		name string
		want float64
	}{
		{
			name: "FullPercentCacheIs100",
			want: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, fullPercentCache)
		})
	}
}
