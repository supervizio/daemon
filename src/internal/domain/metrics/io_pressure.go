// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// IOPressure represents I/O pressure metrics from PSI (Pressure Stall Information).
// PSI is available on Linux 4.20+ and provides information about resource contention.
type IOPressure struct {
	// SomeAvg10 is the percentage of time some tasks were stalled on I/O (10s average).
	SomeAvg10 float64
	// SomeAvg60 is the percentage of time some tasks were stalled on I/O (60s average).
	SomeAvg60 float64
	// SomeAvg300 is the percentage of time some tasks were stalled on I/O (300s average).
	SomeAvg300 float64
	// SomeTotal is the total stall time for some tasks in microseconds.
	SomeTotal uint64
	// FullAvg10 is the percentage of time all tasks were stalled on I/O (10s average).
	FullAvg10 float64
	// FullAvg60 is the percentage of time all tasks were stalled on I/O (60s average).
	FullAvg60 float64
	// FullAvg300 is the percentage of time all tasks were stalled on I/O (300s average).
	FullAvg300 float64
	// FullTotal is the total stall time for all tasks in microseconds.
	FullTotal uint64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// NewIOPressure creates a new IOPressure instance.
//
// Params:
//   - params: IOPressureParams containing all pressure metrics
//
// Returns:
//   - IOPressure: the created IOPressure instance
func NewIOPressure(params *IOPressureParams) IOPressure {
	// Create IOPressure from parameter struct.
	return IOPressure{
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

// IsUnderPressure returns true if there is significant I/O pressure.
// A threshold of 10% on the 10-second average is used as a general indicator.
//
// Returns:
//   - bool: true if I/O pressure exceeds the threshold
func (p *IOPressure) IsUnderPressure() bool {
	// Check if either partial or full stall exceeds pressure threshold.
	return p.SomeAvg10 > PressureThreshold || p.FullAvg10 > PressureThreshold
}
