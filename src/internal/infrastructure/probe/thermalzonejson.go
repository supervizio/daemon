//go:build cgo

package probe

// ThermalZoneJSON contains information about a thermal zone.
// It includes name, current temperature, and optional threshold values.
type ThermalZoneJSON struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	TempCelsius float64  `json:"temp_celsius"`
	TempMax     *float64 `json:"temp_max,omitempty"`
	TempCrit    *float64 `json:"temp_crit,omitempty"`
}
