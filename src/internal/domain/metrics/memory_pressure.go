// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// MemoryPressure represents memory pressure metrics from PSI.
// PSI is available on Linux 4.20+ and provides information about memory contention.
type MemoryPressure struct {
	// SomeAvg10 is the percentage of time some tasks were stalled on memory (10s average).
	SomeAvg10 float64
	// SomeAvg60 is the percentage of time some tasks were stalled on memory (60s average).
	SomeAvg60 float64
	// SomeAvg300 is the percentage of time some tasks were stalled on memory (300s average).
	SomeAvg300 float64
	// SomeTotal is the total stall time for some tasks in microseconds.
	SomeTotal uint64
	// FullAvg10 is the percentage of time all tasks were stalled on memory (10s average).
	FullAvg10 float64
	// FullAvg60 is the percentage of time all tasks were stalled on memory (60s average).
	FullAvg60 float64
	// FullAvg300 is the percentage of time all tasks were stalled on memory (300s average).
	FullAvg300 float64
	// FullTotal is the total stall time for all tasks in microseconds.
	FullTotal uint64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// NewMemoryPressure creates a new MemoryPressure instance.
//
// Params:
//   - params: MemoryPressureParams containing all pressure metrics
//
// Returns:
//   - MemoryPressure: the created MemoryPressure instance
func NewMemoryPressure(params *MemoryPressureParams) MemoryPressure {
	return MemoryPressure{
		SomeAvg10:  params.SomeAvg10,
		SomeAvg60:  params.SomeAvg60,
		SomeAvg300: params.SomeAvg300,
		SomeTotal:  params.SomeTotal,
		FullAvg10:  params.FullAvg10,
		FullAvg60:  params.FullAvg60,
		FullAvg300: params.FullAvg300,
		FullTotal:  params.FullTotal,
		Timestamp:  params.Timestamp,
	}
}

// IsUnderPressure returns true if there is significant memory pressure.
//
// Returns:
//   - bool: true if memory pressure exceeds the threshold
func (p *MemoryPressure) IsUnderPressure() bool {
	return p.SomeAvg10 > PressureThreshold || p.FullAvg10 > PressureThreshold
}
