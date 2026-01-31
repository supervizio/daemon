//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
*/
import "C"

import "unsafe"

// ThermalZone contains thermal zone information and temperature reading.
// It represents a single thermal sensor with its current reading and thresholds.
type ThermalZone struct {
	// Name is the device name (e.g., "coretemp", "acpitz", "nvme").
	Name string
	// Label is the zone label (e.g., "Core 0", "Package id 0").
	Label string
	// TempCelsius is the current temperature in Celsius.
	TempCelsius float64
	// TempMax is the maximum safe temperature (nil if not available).
	TempMax *float64
	// TempCrit is the critical temperature (nil if not available).
	TempCrit *float64
}

// ThermalIsSupported checks if thermal monitoring is supported on this platform.
//
// Returns:
//   - bool: true if thermal monitoring is available
func ThermalIsSupported() bool {
	// Delegate to the Rust probe library for platform detection.
	return bool(C.probe_thermal_is_supported())
}

// CollectThermalZones collects all thermal zones and their temperatures.
//
// Returns:
//   - []ThermalZone: list of thermal zones with temperature readings
//   - error: nil on success, error if probe not initialized or collection fails
//
//nolint:gocritic // dupSubExpr false positive from CGO list operations
func CollectThermalZones() ([]ThermalZone, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return nil slice with initialization error.
		return nil, err
	}

	var list C.ThermalZoneList
	result := C.probe_collect_thermal_zones(&list)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return nil slice with collection error.
		return nil, err
	}

	// Handle empty thermal zone list.
	if list.items == nil || list.count == 0 {
		return nil, nil //nolint:nilnil // Nil slice is valid for empty result
	}

	defer C.probe_free_thermal_list(&list)

	count := int(list.count)
	zones := make([]ThermalZone, 0, count)

	items := unsafe.Slice(list.items, count)
	// Iterate through each thermal zone from the Rust library.
	for _, item := range items {
		zone := ThermalZone{
			Name:        cCharArrayToString(item.name[:]),
			Label:       cCharArrayToString(item.label[:]),
			TempCelsius: float64(item.temp_celsius),
		}

		// Check if maximum temperature threshold is available.
		if bool(item.has_max) {
			max := float64(item.temp_max)
			zone.TempMax = &max
		}

		// Check if critical temperature threshold is available.
		if bool(item.has_crit) {
			crit := float64(item.temp_crit)
			zone.TempCrit = &crit
		}

		zones = append(zones, zone)
	}

	// Return collected thermal zones.
	return zones, nil
}
