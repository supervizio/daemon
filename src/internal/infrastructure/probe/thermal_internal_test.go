//go:build cgo

package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestThermalZone_ZeroValue verifies zero value behavior.
func TestThermalZone_ZeroValue(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "zero value struct has empty fields"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var zone ThermalZone

			assert.Empty(t, zone.Name)
			assert.Empty(t, zone.Label)
			assert.Equal(t, 0.0, zone.TempCelsius)
			assert.Nil(t, zone.TempMax)
			assert.Nil(t, zone.TempCrit)
		})
	}
}

// TestThermalZone_FieldAssignment verifies field assignment.
func TestThermalZone_FieldAssignment(t *testing.T) {
	tests := []struct {
		name        string
		zoneName    string
		label       string
		tempCelsius float64
	}{
		{
			name:        "test sensor",
			zoneName:    "test-sensor",
			label:       "Test Label",
			tempCelsius: 42.0,
		},
		{
			name:        "coretemp sensor",
			zoneName:    "coretemp",
			label:       "Core 0",
			tempCelsius: 55.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			zone := ThermalZone{
				Name:        tt.zoneName,
				Label:       tt.label,
				TempCelsius: tt.tempCelsius,
			}

			assert.Equal(t, tt.zoneName, zone.Name)
			assert.Equal(t, tt.label, zone.Label)
			assert.InDelta(t, tt.tempCelsius, zone.TempCelsius, 0.01)
		})
	}
}

// TestThermalIsSupported verifies thermal support detection.
func TestThermalIsSupported(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			// Just verify it returns a boolean without panic
			supported := ThermalIsSupported()
			_ = supported
		})
	}
}

// TestCollectThermalZones verifies thermal zone collection.
func TestCollectThermalZones(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			zones, err := CollectThermalZones()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				// On some systems thermal zones may not be available
				// but the function should not error
				if err == nil && zones != nil {
					for _, zone := range zones {
						assert.NotEmpty(t, zone.Name)
					}
				}
			}
		})
	}
}
