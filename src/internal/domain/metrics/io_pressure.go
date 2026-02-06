// Package metrics provides domain types for system and process metrics collection.
package metrics

// IOPressure represents I/O pressure metrics from PSI (Pressure Stall Information).
// PSI is available on Linux 4.20+ and provides information about resource contention.
type IOPressure struct {
	// Pressure embeds common PSI fields and the IsUnderPressure method.
	Pressure
}

// NewIOPressure creates a new IOPressure instance.
//
// Params:
//   - params: IOPressureParams containing all pressure metrics
//
// Returns:
//   - IOPressure: the created IOPressure instance
func NewIOPressure(params *IOPressureParams) IOPressure {
	// initialize with all I/O pressure metrics via embedded Pressure
	return IOPressure{
		Pressure: params.ToPressure(),
	}
}
