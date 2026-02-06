// Package metrics provides domain types for system and process metrics collection.
package metrics

// IOPressureParams contains parameters for creating IOPressure instances.
// This struct embeds PressureParams to share common pressure fields.
type IOPressureParams struct {
	// PressureParams embeds common pressure fields.
	PressureParams
}
