// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// IOPressureParams contains parameters for creating IOPressure instances.
// This struct groups all I/O pressure metrics to avoid excessive constructor parameters.
type IOPressureParams struct {
	// SomeAvg10 is the percentage of time some tasks were stalled (10s average).
	SomeAvg10 float64
	// SomeAvg60 is the percentage of time some tasks were stalled (60s average).
	SomeAvg60 float64
	// SomeAvg300 is the percentage of time some tasks were stalled (300s average).
	SomeAvg300 float64
	// SomeTotal is the total stall time for some tasks in microseconds.
	SomeTotal uint64
	// FullAvg10 is the percentage of time all tasks were stalled (10s average).
	FullAvg10 float64
	// FullAvg60 is the percentage of time all tasks were stalled (60s average).
	FullAvg60 float64
	// FullAvg300 is the percentage of time all tasks were stalled (300s average).
	FullAvg300 float64
	// FullTotal is the total stall time for all tasks in microseconds.
	FullTotal uint64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}
