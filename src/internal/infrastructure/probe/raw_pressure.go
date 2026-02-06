// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawPressureMetrics holds raw pressure data for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawPressureMetrics struct {
	// Available indicates if pressure metrics are available.
	Available bool
	// CPU holds CPU pressure data.
	CPU RawCPUPressure
	// Memory holds memory pressure data.
	Memory RawMemoryPressure
	// IO holds I/O pressure data.
	IO RawIOPressure
}
