// Package metrics provides domain types for system and process metrics collection.
package metrics

// MemoryPressure represents memory pressure metrics from PSI.
// PSI is available on Linux 4.20+ and provides information about memory contention.
type MemoryPressure struct {
	// Pressure embeds common PSI fields and the IsUnderPressure method.
	Pressure
}

// NewMemoryPressure creates a new MemoryPressure instance.
//
// Params:
//   - params: MemoryPressureParams containing all pressure metrics
//
// Returns:
//   - MemoryPressure: the created MemoryPressure instance
func NewMemoryPressure(params *MemoryPressureParams) MemoryPressure {
	// initialize with all memory pressure metrics via embedded Pressure
	return MemoryPressure{
		Pressure: params.ToPressure(),
	}
}
