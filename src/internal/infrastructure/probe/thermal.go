//go:build cgo

package probe

/*
#include "probe.h"
*/
import "C"

import "unsafe"

// ThermalZone contains thermal zone information and temperature reading.
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
func ThermalIsSupported() bool {
	return bool(C.probe_thermal_is_supported())
}

// CollectThermalZones collects all thermal zones and their temperatures.
//
//nolint:gocritic // dupSubExpr false positive from CGO list operations
func CollectThermalZones() ([]ThermalZone, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var list C.ThermalZoneList
	result := C.probe_collect_thermal_zones(&list)
	if err := resultToError(result); err != nil {
		return nil, err
	}

	if list.items == nil || list.count == 0 {
		return []ThermalZone{}, nil
	}

	defer C.probe_free_thermal_list(&list)

	count := int(list.count)
	zones := make([]ThermalZone, count)

	items := unsafe.Slice(list.items, count)
	for i, item := range items {
		zone := ThermalZone{
			Name:        cCharArrayToString(item.name[:]),
			Label:       cCharArrayToString(item.label[:]),
			TempCelsius: float64(item.temp_celsius),
		}

		if bool(item.has_max) {
			max := float64(item.temp_max)
			zone.TempMax = &max
		}

		if bool(item.has_crit) {
			crit := float64(item.temp_crit)
			zone.TempCrit = &crit
		}

		zones[i] = zone
	}

	return zones, nil
}
