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

func TestSectorSizeConstant(t *testing.T) {
	tests := []struct {
		name string
		want uint64
	}{
		{
			name: "SectorSizeIs512",
			want: 512,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sectorSize)
		})
	}
}
