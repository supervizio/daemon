// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// LoadAverageParams contains parameters for creating LoadAverage instances.
// This struct groups all load average metrics to avoid excessive constructor parameters.
type LoadAverageParams struct {
	// Load1 is the 1-minute load average.
	Load1 float64
	// Load5 is the 5-minute load average.
	Load5 float64
	// Load15 is the 15-minute load average.
	Load15 float64
	// RunningProcesses is the number of currently running processes.
	RunningProcesses int
	// TotalProcesses is the total number of processes in the system.
	TotalProcesses int
	// LastPID is the PID of the most recently created process.
	LastPID int
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}
