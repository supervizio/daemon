// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// ProcessMemory represents per-process memory metrics collected from /proc/[pid]/status.
//
// This value object captures the memory usage of a specific process including
// resident, virtual, shared, and swap memory.
type ProcessMemory struct {
	// PID is the process identifier.
	PID int
	// Name is the process command name.
	Name string
	// RSS is the Resident Set Size in bytes (physical memory used).
	RSS uint64
	// VMS is the Virtual Memory Size in bytes (total virtual memory).
	VMS uint64
	// Shared is the shared memory in bytes (RssShmem + RssFile).
	Shared uint64
	// Swap is the swap memory used by this process in bytes.
	Swap uint64
	// Data is the private data segment size in bytes.
	Data uint64
	// Stack is the stack size in bytes.
	Stack uint64
	// UsagePercent is the percentage of total system RAM used by this process (0-100).
	UsagePercent float64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// NewProcessMemory creates a new ProcessMemory instance with calculated fields.
//
// Params:
//   - input: pointer to ProcessMemoryInput containing all process memory parameters.
//
// Returns:
//   - *ProcessMemory: initialized process memory metrics with calculated UsagePercent.
func NewProcessMemory(input *ProcessMemoryInput) *ProcessMemory {
	// Calculate usage percentage based on total system memory.
	var usagePercent float64
	// Check if totalSystemMemory is non-zero to avoid division by zero.
	if input.TotalSystemMemory > 0 {
		usagePercent = float64(input.RSS) / float64(input.TotalSystemMemory) * percentMultiplier
	}
	// Return initialized process memory metrics struct.
	return &ProcessMemory{
		PID:          input.PID,
		Name:         input.Name,
		RSS:          input.RSS,
		VMS:          input.VMS,
		Shared:       input.Shared,
		Swap:         input.Swap,
		Data:         input.Data,
		Stack:        input.Stack,
		UsagePercent: usagePercent,
		Timestamp:    time.Now(),
	}
}

// TotalResident returns the total resident memory (RSS + Swap).
//
// Returns:
//   - uint64: total resident memory including swapped pages.
func (p *ProcessMemory) TotalResident() uint64 {
	// Combine physical memory and swapped memory for total resident size.
	return p.RSS + p.Swap
}
