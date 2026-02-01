//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// ThermalMetricsJSON contains thermal sensor information.
// It indicates support status and provides zone-specific temperature data.
type ThermalMetricsJSON struct {
	Supported bool              `json:"supported"`
	Zones     []ThermalZoneJSON `json:"zones,omitempty"`
}
