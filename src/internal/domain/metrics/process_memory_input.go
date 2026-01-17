// Package metrics provides domain types for system and process metrics collection.
package metrics

// ProcessMemoryInput contains the input parameters for creating ProcessMemory.
//
// This struct groups the parameters needed to construct a ProcessMemory value object.
type ProcessMemoryInput struct {
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
	// TotalSystemMemory is the total system RAM for percentage calculation.
	TotalSystemMemory uint64
}
