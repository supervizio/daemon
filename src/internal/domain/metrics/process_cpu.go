// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// ProcessCPU represents per-process CPU metrics collected from /proc/[pid]/stat.
//
// This value object captures the CPU time used by a specific process and its children.
// All time values are measured in jiffies from the process start.
//
// Fields are ordered by size for optimal memory alignment:
// time.Time (24B), string (16B), then 8-byte fields.
type ProcessCPU struct {
	// Timestamp is when this sample was taken.
	Timestamp time.Time
	// Name is the process command name (from comm field).
	Name string
	// PID is the process identifier.
	PID int
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
}

// NewProcessCPU creates a new ProcessCPU instance.
//
// Params:
//   - params: ProcessCPUParams containing all process CPU metrics
//
// Returns:
//   - *ProcessCPU: initialized process CPU metrics struct.
func NewProcessCPU(params *ProcessCPUParams) *ProcessCPU {
	// initialize with all process CPU fields
	return &ProcessCPU{
		PID:            params.PID,
		Name:           params.Name,
		User:           params.User,
		System:         params.System,
		ChildrenUser:   params.ChildrenUser,
		ChildrenSystem: params.ChildrenSystem,
		StartTime:      params.StartTime,
		UsagePercent:   params.UsagePercent,
		Timestamp:      params.Timestamp,
	}
}

// Total returns the total CPU time used by this process.
//
// Returns:
//   - uint64: sum of user and system time in jiffies.
func (p *ProcessCPU) Total() uint64 {
	// sum user and system time
	return p.User + p.System
}

// TotalWithChildren returns the total CPU time including waited-for children.
//
// Returns:
//   - uint64: sum of process and children CPU time in jiffies.
func (p *ProcessCPU) TotalWithChildren() uint64 {
	// sum process and children CPU times
	return p.User + p.System + p.ChildrenUser + p.ChildrenSystem
}
