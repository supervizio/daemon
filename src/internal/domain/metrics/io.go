// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// IOStats represents system-wide I/O statistics.
type IOStats struct {
	// ReadBytesTotal is the total bytes read across all devices.
	ReadBytesTotal uint64
	// WriteBytesTotal is the total bytes written across all devices.
	WriteBytesTotal uint64
	// ReadOpsTotal is the total read operations across all devices.
	ReadOpsTotal uint64
	// WriteOpsTotal is the total write operations across all devices.
	WriteOpsTotal uint64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// TotalBytes returns the total bytes transferred (read + write).
func (i IOStats) TotalBytes() uint64 {
	return i.ReadBytesTotal + i.WriteBytesTotal
}

// TotalOps returns the total operations (read + write).
func (i IOStats) TotalOps() uint64 {
	return i.ReadOpsTotal + i.WriteOpsTotal
}

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

// IsUnderPressure returns true if there is significant I/O pressure.
// A threshold of 10% on the 10-second average is used as a general indicator.
func (p IOPressure) IsUnderPressure() bool {
	return p.SomeAvg10 > 10 || p.FullAvg10 > 10
}

// MemoryPressure represents memory pressure metrics from PSI.
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

// IsUnderPressure returns true if there is significant memory pressure.
func (p MemoryPressure) IsUnderPressure() bool {
	return p.SomeAvg10 > 10 || p.FullAvg10 > 10
}

// CPUPressure represents CPU pressure metrics from PSI.
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

// IsUnderPressure returns true if there is significant CPU pressure.
func (p CPUPressure) IsUnderPressure() bool {
	return p.SomeAvg10 > 10
}

// LoadAverage represents system load average metrics.
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

// IsOverloaded returns true if the system is overloaded.
// This compares the 1-minute load average to the number of CPU cores.
func (l LoadAverage) IsOverloaded(numCPU int) bool {
	if numCPU <= 0 {
		numCPU = 1
	}
	return l.Load1 > float64(numCPU)
}
