//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

// QuotaUsage represents current resource usage for a process.
// It provides a snapshot of memory, PIDs, and CPU utilization.
type QuotaUsage struct {
	// MemoryBytes is the current memory usage in bytes.
	MemoryBytes uint64

	// MemoryLimitBytes is the memory limit (0 = no limit).
	MemoryLimitBytes uint64

	// PIDsCurrent is the current number of processes/threads.
	PIDsCurrent uint64

	// PIDsLimit is the PIDs limit (0 = no limit).
	PIDsLimit uint64

	// CPUPercent is the current CPU usage percentage.
	CPUPercent float64

	// CPULimitPercent is the CPU limit percentage (0 = no limit).
	CPULimitPercent float64
}

// NewQuotaUsage creates a new QuotaUsage with all usage values initialized to zero.
//
// Returns:
//   - *QuotaUsage: new quota usage instance
func NewQuotaUsage() *QuotaUsage {
	// Return empty usage struct with zero values.
	return &QuotaUsage{}
}

// MemoryUsagePercent calculates memory usage as percentage of limit.
// Returns 0 if no limit is set.
//
// Returns:
//   - float64: memory usage percentage (0 if unlimited)
func (u *QuotaUsage) MemoryUsagePercent() float64 {
	// Check if memory limit is not set or unlimited.
	if u.MemoryLimitBytes == 0 || u.MemoryLimitBytes == unlimitedValue {
		// Return zero for no limit.
		return 0
	}
	// Calculate percentage from usage and limit.
	return float64(u.MemoryBytes) / float64(u.MemoryLimitBytes) * percentMultiplierQuota
}
