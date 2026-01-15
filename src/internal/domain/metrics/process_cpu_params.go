// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// ProcessCPUParams contains parameters for creating ProcessCPU instances.
// This struct groups all process CPU metrics to avoid excessive constructor parameters.
type ProcessCPUParams struct {
	// PID is the process identifier.
	PID int
	// Name is the process command name (from comm field).
	Name string
	// User is the user mode CPU time (jiffies).
	User uint64
	// System is the kernel mode CPU time (jiffies).
	System uint64
	// ChildrenUser is the user mode CPU time of waited-for children (jiffies).
	ChildrenUser uint64
	// ChildrenSystem is the kernel mode CPU time of waited-for children (jiffies).
	ChildrenSystem uint64
	// StartTime is when the process started (jiffies after system boot).
	StartTime uint64
	// UsagePercent is the calculated CPU usage percentage (0-100).
	UsagePercent float64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}
