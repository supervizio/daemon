// Package metrics provides domain types for system and process metrics collection.
package metrics

// PressureThreshold is the percentage threshold for determining if a system is under pressure.
// A value of 10% on the 10-second average is used as a general indicator.
const PressureThreshold float64 = 10.0
