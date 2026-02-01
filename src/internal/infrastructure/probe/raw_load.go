// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawLoadData holds raw load data for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawLoadData struct {
	// Load1Min represents 1-minute load average.
	Load1Min float64
	// Load5Min represents 5-minute load average.
	Load5Min float64
	// Load15Min represents 15-minute load average.
	Load15Min float64
}
