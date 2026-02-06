//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestThermalZoneInternal(t *testing.T) {
	tests := []struct {
		name string
		zone *ThermalZone
	}{
		{
			name: "EmptyThermalZone",
			zone: &ThermalZone{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.zone)
		})
	}
}
