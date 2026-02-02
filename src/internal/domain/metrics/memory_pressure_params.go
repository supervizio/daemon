// Package metrics provides domain types for system and process metrics collection.
package metrics

// MemoryPressureParams contains parameters for creating MemoryPressure instances.
// This struct embeds PressureParams to share common pressure fields.
type MemoryPressureParams struct {
	// PressureParams embeds common pressure fields.
	PressureParams
}
