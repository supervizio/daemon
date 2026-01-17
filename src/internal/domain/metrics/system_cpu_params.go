// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// SystemCPUParams contains parameters for creating SystemCPU instances.
// This struct groups all CPU metrics to avoid excessive constructor parameters.
type SystemCPUParams struct {
	// User is the time spent in user mode (jiffies).
	User uint64
	// Nice is the time spent in user mode with low priority (jiffies).
	Nice uint64
	// System is the time spent in kernel mode (jiffies).
	System uint64
	// Idle is the time spent in idle state (jiffies).
	Idle uint64
	// IOWait is the time waiting for I/O completion (jiffies).
	IOWait uint64
	// IRQ is the time servicing hardware interrupts (jiffies).
	IRQ uint64
	// SoftIRQ is the time servicing software interrupts (jiffies).
	SoftIRQ uint64
	// Steal is the time stolen by other operating systems in virtualized environments (jiffies).
	Steal uint64
	// Guest is the time spent running a virtual CPU (jiffies).
	Guest uint64
	// GuestNice is the time spent running a niced guest (jiffies).
	GuestNice uint64
	// UsagePercent is the calculated CPU usage percentage (0-100).
	UsagePercent float64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}
