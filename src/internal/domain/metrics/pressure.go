// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// PressureThreshold is the percentage threshold for determining if a system is under pressure.
// A value of 10% on the 10-second average is used as a general indicator.
const PressureThreshold float64 = 10.0

// Pressure represents common PSI (Pressure Stall Information) metrics.
// PSI is available on Linux 4.20+ and provides information about resource contention.
// This base type is embedded by IOPressure and MemoryPressure.
type Pressure struct {
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

// IsUnderPressure returns true if there is significant pressure.
// A threshold of 10% on the 10-second average is used as a general indicator.
//
// Returns:
//   - bool: true if pressure exceeds the threshold
func (p *Pressure) IsUnderPressure() bool {
	// check if either some or full pressure exceeds threshold
	return p.SomeAvg10 > PressureThreshold || p.FullAvg10 > PressureThreshold
}

// PressureParams contains common parameters for creating pressure instances.
// This struct groups all pressure metrics to avoid excessive constructor parameters.
type PressureParams struct {
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

// ToPressure converts params to a Pressure instance.
//
// Returns:
//   - Pressure: the created Pressure instance
func (p *PressureParams) ToPressure() Pressure {
	// initialize with all pressure metrics
	return Pressure{
		SomeAvg10:  p.SomeAvg10,
		SomeAvg60:  p.SomeAvg60,
		SomeAvg300: p.SomeAvg300,
		SomeTotal:  p.SomeTotal,
		FullAvg10:  p.FullAvg10,
		FullAvg60:  p.FullAvg60,
		FullAvg300: p.FullAvg300,
		FullTotal:  p.FullTotal,
		Timestamp:  p.Timestamp,
	}
}
