// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// LoadAverage represents system load average metrics.
// This type captures load averages over 1, 5, and 15 minute intervals plus process counts.
type LoadAverage struct {
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

// NewLoadAverage creates a new LoadAverage instance.
//
// Params:
//   - params: LoadAverageParams containing all load metrics
//
// Returns:
//   - LoadAverage: the created LoadAverage instance
func NewLoadAverage(params LoadAverageParams) LoadAverage {
	// Convert parameter struct directly to LoadAverage (identical fields).
	return LoadAverage(params)
}

// IsOverloaded returns true if the system is overloaded.
// This compares the 1-minute load average to the number of CPU cores.
//
// Params:
//   - numCPU: number of CPU cores available
//
// Returns:
//   - bool: true if load exceeds available CPU cores
func (l LoadAverage) IsOverloaded(numCPU int) bool {
	// Ensure we have at least one CPU to avoid division by zero.
	if numCPU <= 0 {
		numCPU = 1
	}
	// Load exceeding CPU count indicates overload.
	return l.Load1 > float64(numCPU)
}
