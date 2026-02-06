//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// CPUMetricsJSON contains CPU-related metrics for JSON output.
// It includes usage percentage, core count, and optional pressure metrics.
type CPUMetricsJSON struct {
	UsagePercent float64          `json:"usage_percent"`
	Cores        uint32           `json:"cores"`
	FrequencyMHz uint64           `json:"frequency_mhz"`
	Pressure     *CPUPressureJSON `json:"pressure,omitempty"`
}
