// Package metrics provides domain types for system and process metrics collection.
package probe

import "time"

// SystemMemory represents system-wide memory metrics collected from /proc/meminfo.
type SystemMemory struct {
	// Total is the total physical RAM in bytes.
	Total uint64
	// Available is the memory available for starting new applications in bytes.
	Available uint64
	// Used is the used memory in bytes (Total - Available).
	Used uint64
	// Free is the free memory in bytes (MemFree from /proc/meminfo).
	Free uint64
	// Cached is the page cache memory in bytes.
	Cached uint64
	// Buffers is the buffer memory in bytes.
	Buffers uint64
	// SwapTotal is the total swap space in bytes.
	SwapTotal uint64
	// SwapUsed is the used swap space in bytes.
	SwapUsed uint64
	// SwapFree is the free swap space in bytes.
	SwapFree uint64
	// Shared is the shared memory in bytes (Shmem from /proc/meminfo).
	Shared uint64
	// UsagePercent is the calculated memory usage percentage (0-100).
	UsagePercent float64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// SwapUsagePercent returns the swap usage percentage (0-100).
// Returns 0 if SwapTotal is 0.
func (m SystemMemory) SwapUsagePercent() float64 {
	if m.SwapTotal == 0 {
		return 0
	}
	return float64(m.SwapUsed) / float64(m.SwapTotal) * 100
}

// ProcessMemory represents per-process memory metrics collected from /proc/[pid]/status.
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

// TotalResident returns the total resident memory (RSS + Swap).
func (p ProcessMemory) TotalResident() uint64 {
	return p.RSS + p.Swap
}
