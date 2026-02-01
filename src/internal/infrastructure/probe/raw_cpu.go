// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawCPUData holds raw CPU data for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawCPUData struct {
	// IdlePercent represents the CPU idle percentage.
	IdlePercent float64
}
