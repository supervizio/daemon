// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// SystemMemory represents system-wide memory metrics collected from /proc/meminfo.
//
// This value object captures the current memory state of the system including RAM and swap.
// The Used field is calculated as Total - Available.
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

// NewSystemMemory creates a new SystemMemory instance with calculated fields.
//
// Params:
//   - input: pointer to SystemMemoryInput containing all memory parameters.
//
// Returns:
//   - *SystemMemory: initialized memory metrics with calculated Used and UsagePercent.
func NewSystemMemory(input *SystemMemoryInput) *SystemMemory {
	// calculate used memory from total and available
	used := input.Total - input.Available
	var usagePercent float64
	// calculate usage percentage if total is non-zero
	if input.Total > 0 {
		usagePercent = float64(used) / float64(input.Total) * percentMultiplier
	}
	// initialize with all system memory fields
	return &SystemMemory{
		Total:        input.Total,
		Available:    input.Available,
		Used:         used,
		Free:         input.Free,
		Cached:       input.Cached,
		Buffers:      input.Buffers,
		SwapTotal:    input.SwapTotal,
		SwapUsed:     input.SwapUsed,
		SwapFree:     input.SwapFree,
		Shared:       input.Shared,
		UsagePercent: usagePercent,
		Timestamp:    time.Now(),
	}
}

// SwapUsagePercent returns the swap usage percentage (0-100).
//
// Returns:
//   - float64: swap usage percentage, or 0 if SwapTotal is 0.
func (m *SystemMemory) SwapUsagePercent() float64 {
	// return zero if no swap is configured
	if m.SwapTotal == 0 {
		// return zero percentage
		return 0
	}
	// calculate swap usage percentage
	return float64(m.SwapUsed) / float64(m.SwapTotal) * percentMultiplier
}
