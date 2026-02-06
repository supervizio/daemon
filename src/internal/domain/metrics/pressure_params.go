// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

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

// NewPressureParams creates a new PressureParams with zero values.
//
// Returns:
//   - *PressureParams: a new pressure params instance.
func NewPressureParams() *PressureParams {
	// Return zero-value params for caller to populate.
	return &PressureParams{}
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
