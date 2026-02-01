//go:build cgo

package probe

// ThermalMetricsJSON contains thermal sensor information.
// It indicates support status and provides zone-specific temperature data.
type ThermalMetricsJSON struct {
	Supported bool              `json:"supported"`
	Zones     []ThermalZoneJSON `json:"zones,omitempty"`
}
