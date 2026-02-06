// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// SystemCPU represents system-wide CPU metrics collected from /proc/stat.
//
// This value object captures the raw CPU time counters (in jiffies) from the Linux kernel.
// Use Total() and Active() to calculate aggregate values, or access individual fields for detailed analysis.
type SystemCPU struct {
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

// NewSystemCPU creates a new SystemCPU instance.
//
// Params:
//   - params: SystemCPUParams containing all CPU metrics
//
// Returns:
//   - *SystemCPU: initialized system CPU metrics struct.
func NewSystemCPU(params *SystemCPUParams) *SystemCPU {
	// initialize with all system CPU fields
	return &SystemCPU{
		User:         params.User,
		Nice:         params.Nice,
		System:       params.System,
		Idle:         params.Idle,
		IOWait:       params.IOWait,
		IRQ:          params.IRQ,
		SoftIRQ:      params.SoftIRQ,
		Steal:        params.Steal,
		Guest:        params.Guest,
		GuestNice:    params.GuestNice,
		UsagePercent: params.UsagePercent,
		Timestamp:    params.Timestamp,
	}
}

// Total returns the sum of all CPU time fields.
//
// Returns:
//   - uint64: total CPU time across all states in jiffies.
func (c *SystemCPU) Total() uint64 {
	// sum all CPU time components
	return c.User + c.Nice + c.System + c.Idle + c.IOWait + c.IRQ + c.SoftIRQ + c.Steal + c.Guest + c.GuestNice
}

// Active returns the sum of all non-idle CPU time fields.
//
// Returns:
//   - uint64: active CPU time excluding idle and iowait states.
func (c *SystemCPU) Active() uint64 {
	// subtract idle time from total
	return c.Total() - c.Idle - c.IOWait
}
