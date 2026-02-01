//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

// percentMultiplierQuota is used to convert ratios to percentages in quota calculations.
const percentMultiplierQuota float64 = 100.0

// unlimitedValue is the sentinel value for unlimited resources.
const unlimitedValue uint64 = ^uint64(0)

// Quota limit flags indicate which resource limits are set.
// These flags are used with QuotaLimits.Flags field.
const (
	// QuotaFlagCPU indicates CPU quota is set.
	QuotaFlagCPU uint32 = 1 << 0
	// QuotaFlagMemory indicates memory limit is set.
	QuotaFlagMemory uint32 = 1 << 1
	// QuotaFlagPIDs indicates PIDs limit is set.
	QuotaFlagPIDs uint32 = 1 << 2
	// QuotaFlagNofile indicates file descriptor limit is set.
	QuotaFlagNofile uint32 = 1 << 3
	// QuotaFlagCPUTime indicates CPU time limit is set.
	QuotaFlagCPUTime uint32 = 1 << 4
	// QuotaFlagData indicates data segment limit is set.
	QuotaFlagData uint32 = 1 << 5
	// QuotaFlagIORead indicates I/O read bandwidth limit is set.
	QuotaFlagIORead uint32 = 1 << 6
	// QuotaFlagIOWrite indicates I/O write bandwidth limit is set.
	QuotaFlagIOWrite uint32 = 1 << 7
)

// QuotaLimits represents detected resource limits for a process.
// It provides cross-platform access to CPU, memory, and I/O constraints.
type QuotaLimits struct {
	// CPUQuotaUS is the CPU quota in microseconds per period.
	// Zero means not set or unlimited.
	CPUQuotaUS uint64

	// CPUPeriodUS is the CPU period in microseconds (typically 100000).
	CPUPeriodUS uint64

	// MemoryLimitBytes is the memory limit in bytes.
	// Zero means not set, MaxUint64 means unlimited.
	MemoryLimitBytes uint64

	// PIDsLimit is the maximum number of processes/threads.
	PIDsLimit uint64

	// NofileLimit is the maximum file descriptors.
	NofileLimit uint64

	// CPUTimeLimitSecs is the maximum CPU time in seconds.
	CPUTimeLimitSecs uint64

	// DataLimitBytes is the maximum heap/data size in bytes.
	DataLimitBytes uint64

	// IOReadBPS is the I/O read bandwidth limit in bytes/sec.
	IOReadBPS uint64

	// IOWriteBPS is the I/O write bandwidth limit in bytes/sec.
	IOWriteBPS uint64

	// Flags indicates which fields are valid.
	Flags uint32
}

// NewQuotaLimits creates a new QuotaLimits with all limits initialized to zero.
//
// Returns:
//   - *QuotaLimits: new quota limits instance with no limits set
func NewQuotaLimits() *QuotaLimits {
	// Return empty limits struct with zero values.
	return &QuotaLimits{}
}

// HasCPULimit returns whether a CPU limit is set.
//
// Returns:
//   - bool: true if a valid CPU limit is configured
func (l *QuotaLimits) HasCPULimit() bool {
	// Check flag, non-zero quota, and not unlimited.
	return l.Flags&QuotaFlagCPU != 0 && l.CPUQuotaUS > 0 && l.CPUQuotaUS != unlimitedValue
}

// HasMemoryLimit returns whether a memory limit is set.
//
// Returns:
//   - bool: true if a valid memory limit is configured
func (l *QuotaLimits) HasMemoryLimit() bool {
	// Check flag, non-zero limit, and not unlimited.
	return l.Flags&QuotaFlagMemory != 0 && l.MemoryLimitBytes > 0 && l.MemoryLimitBytes != unlimitedValue
}

// CPULimitPercent calculates the CPU limit as a percentage.
// Returns 0 if no limit is set.
//
// Returns:
//   - float64: CPU limit percentage (0 if unlimited)
func (l *QuotaLimits) CPULimitPercent() float64 {
	// Check if CPU limit is not set or period is zero.
	if !l.HasCPULimit() || l.CPUPeriodUS == 0 {
		// Return zero for no limit.
		return 0
	}
	// Calculate percentage from quota and period.
	return float64(l.CPUQuotaUS) / float64(l.CPUPeriodUS) * percentMultiplierQuota
}
