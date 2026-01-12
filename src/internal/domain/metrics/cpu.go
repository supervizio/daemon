// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// SystemCPU represents system-wide CPU metrics collected from /proc/stat.
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

// Total returns the sum of all CPU time fields.
func (c SystemCPU) Total() uint64 {
	return c.User + c.Nice + c.System + c.Idle + c.IOWait + c.IRQ + c.SoftIRQ + c.Steal + c.Guest + c.GuestNice
}

// Active returns the sum of all non-idle CPU time fields.
func (c SystemCPU) Active() uint64 {
	return c.Total() - c.Idle - c.IOWait
}

// ProcessCPU represents per-process CPU metrics collected from /proc/[pid]/stat.
type ProcessCPU struct {
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

// Total returns the total CPU time used by this process.
func (p ProcessCPU) Total() uint64 {
	return p.User + p.System
}

// TotalWithChildren returns the total CPU time including waited-for children.
func (p ProcessCPU) TotalWithChildren() uint64 {
	return p.User + p.System + p.ChildrenUser + p.ChildrenSystem
}
