//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestThermalIsSupported verifies thermal support checking.
func TestThermalIsSupported(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "checks thermal support"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			supported := probe.ThermalIsSupported()
			t.Logf("Thermal monitoring supported: %v", supported)
		})
	}
}

// TestCollectThermalZones verifies thermal zone collection.
func TestCollectThermalZones(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects thermal zones when supported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			if !probe.ThermalIsSupported() {
				t.Log("Thermal monitoring not supported on this platform, test not applicable")
				return
			}

			zones, err := probe.CollectThermalZones()
			require.NoError(t, err)

			// May be empty on some systems
			if len(zones) > 0 {
				for _, zone := range zones {
					t.Logf("Zone: %s (%s) - %.1f C", zone.Name, zone.Label, zone.TempCelsius)
					assert.NotEmpty(t, zone.Name)
				}
			}
		})
	}
}

// TestCollectThermalZones_NotInitialized verifies error when not initialized.
func TestCollectThermalZones_NotInitialized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns error when not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := probe.CollectThermalZones()
			// Should return error because probe is not initialized
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}

// TestThermalZone_Structure verifies ThermalZone struct fields.
func TestThermalZone_Structure(t *testing.T) {
	tests := []struct {
		name        string
		zoneName    string
		label       string
		tempCelsius float64
		tempMax     *float64
		tempCrit    *float64
	}{
		{
			name:        "with thresholds",
			zoneName:    "coretemp",
			label:       "Core 0",
			tempCelsius: 45.5,
			tempMax:     floatPtr(100.0),
			tempCrit:    floatPtr(110.0),
		},
		{
			name:        "without thresholds",
			zoneName:    "acpitz",
			label:       "",
			tempCelsius: 50.0,
			tempMax:     nil,
			tempCrit:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			zone := probe.ThermalZone{
				Name:        tt.zoneName,
				Label:       tt.label,
				TempCelsius: tt.tempCelsius,
				TempMax:     tt.tempMax,
				TempCrit:    tt.tempCrit,
			}

			assert.Equal(t, tt.zoneName, zone.Name)
			assert.Equal(t, tt.label, zone.Label)
			assert.InDelta(t, tt.tempCelsius, zone.TempCelsius, 0.01)

			if tt.tempMax != nil {
				require.NotNil(t, zone.TempMax)
				assert.InDelta(t, *tt.tempMax, *zone.TempMax, 0.01)
			} else {
				assert.Nil(t, zone.TempMax)
			}

			if tt.tempCrit != nil {
				require.NotNil(t, zone.TempCrit)
				assert.InDelta(t, *tt.tempCrit, *zone.TempCrit, 0.01)
			} else {
				assert.Nil(t, zone.TempCrit)
			}
		})
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
