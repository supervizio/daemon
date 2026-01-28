// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// CPUPressure represents CPU pressure metrics from PSI.
// PSI is available on Linux 4.20+ and provides information about CPU contention.
type CPUPressure struct {
	// SomeAvg10 is the percentage of time some tasks were stalled on CPU (10s average).
	SomeAvg10 float64
	// SomeAvg60 is the percentage of time some tasks were stalled on CPU (60s average).
	SomeAvg60 float64
	// SomeAvg300 is the percentage of time some tasks were stalled on CPU (300s average).
	SomeAvg300 float64
	// SomeTotal is the total stall time for some tasks in microseconds.
	SomeTotal uint64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// NewCPUPressure creates a new CPUPressure instance.
//
// Params:
//   - someAvg10: percentage of time some tasks were stalled (10s average)
//   - someAvg60: percentage of time some tasks were stalled (60s average)
//   - someAvg300: percentage of time some tasks were stalled (300s average)
//   - someTotal: total stall time for some tasks in microseconds
//   - timestamp: when this sample was taken
//
// Returns:
//   - CPUPressure: the created CPUPressure instance
func NewCPUPressure(someAvg10, someAvg60, someAvg300 float64, someTotal uint64, timestamp time.Time) CPUPressure {
	return CPUPressure{
		SomeAvg10:  someAvg10,
		SomeAvg60:  someAvg60,
		SomeAvg300: someAvg300,
		SomeTotal:  someTotal,
		Timestamp:  timestamp,
	}
}

// IsUnderPressure returns true if there is significant CPU pressure.
//
// Returns:
//   - bool: true if CPU pressure exceeds the threshold
func (p CPUPressure) IsUnderPressure() bool {
	return p.SomeAvg10 > PressureThreshold
}
