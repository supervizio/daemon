//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
//
//nolint:ktn-struct-onefile,ktn-const-order // Quota structs logically grouped; ContainerRuntime constants follow type
package probe

/*
#include "probe.h"
*/
import "C"

// percentMultiplierQuota is used to convert ratios to percentages in quota calculations.
const percentMultiplierQuota float64 = 100.0

// unlimitedValue is the sentinel value for unlimited resources.
const unlimitedValue uint64 = ^uint64(0)

// Quota limit flags indicate which resource limits are set.
// These flags are used with QuotaLimits.Flags field.
//
//nolint:ktn-const-order // Flags are grouped with QuotaLimits type for readability
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

// QuotaUsage represents current resource usage for a process.
// It provides a snapshot of memory, PIDs, and CPU utilization.
type QuotaUsage struct { //nolint:ktn-struct-onefile // grouped with quota types
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

// containerRuntimeUnknownStr is the string representation for unknown runtimes.
const containerRuntimeUnknownStr string = "unknown" //nolint:ktn-const-order // Placed near ContainerRuntime type for readability

// ContainerRuntime represents a container runtime type.
// It identifies the orchestration platform or isolation mechanism.
type ContainerRuntime int

// Container runtime constants for identifying the execution environment.
const ( //nolint:ktn-const-order // Typed constants must follow their type definition
	// ContainerRuntimeNone indicates no containerization.
	ContainerRuntimeNone ContainerRuntime = 0
	// ContainerRuntimeDocker indicates Docker runtime.
	ContainerRuntimeDocker ContainerRuntime = 1
	// ContainerRuntimePodman indicates Podman runtime.
	ContainerRuntimePodman ContainerRuntime = 2
	// ContainerRuntimeLXC indicates LXC runtime.
	ContainerRuntimeLXC ContainerRuntime = 3
	// ContainerRuntimeKubernetes indicates Kubernetes runtime.
	ContainerRuntimeKubernetes ContainerRuntime = 4
	// ContainerRuntimeJail indicates FreeBSD jail.
	ContainerRuntimeJail ContainerRuntime = 5
	// ContainerRuntimeUnknown indicates unknown container runtime.
	ContainerRuntimeUnknown ContainerRuntime = 255
)

// String returns the string representation of the container runtime.
//
// Returns:
//   - string: human-readable runtime name
//
//nolint:cyclop // Switch-based enum stringer requires multiple branches
func (r ContainerRuntime) String() string {
	// Map runtime enum to string representation.
	switch r {
	// No container runtime detected.
	case ContainerRuntimeNone:
		// Return string for no containerization.
		return "none"
	// Docker container runtime.
	case ContainerRuntimeDocker:
		// Return string for Docker.
		return "docker"
	// Podman container runtime.
	case ContainerRuntimePodman:
		// Return string for Podman.
		return "podman"
	// LXC container runtime.
	case ContainerRuntimeLXC:
		// Return string for LXC.
		return "lxc"
	// Kubernetes container orchestrator.
	case ContainerRuntimeKubernetes:
		// Return string for Kubernetes.
		return "kubernetes"
	// FreeBSD jail isolation.
	case ContainerRuntimeJail:
		// Return string for FreeBSD jail.
		return "jail"
	// Unknown container runtime.
	case ContainerRuntimeUnknown:
		// Return string for unknown runtime.
		return containerRuntimeUnknownStr
	// Default case for future runtime values.
	default:
		// Return unknown for unrecognized values.
		return containerRuntimeUnknownStr
	}
}

// ContainerInfo represents container detection results.
// It provides information about the containerized execution environment.
type ContainerInfo struct { //nolint:ktn-struct-onefile // grouped with quota types
	// IsContainerized indicates whether running in a container.
	IsContainerized bool

	// Runtime is the detected container runtime.
	Runtime ContainerRuntime

	// ContainerID is the container ID if available.
	ContainerID string
}

// ReadQuotaLimits reads resource limits for a process.
//
// Params:
//   - pid: process ID (use os.Getpid() for current process)
//
// Returns:
//   - *QuotaLimits: detected limits
//   - error: nil on success, error if probe not initialized or operation fails
func ReadQuotaLimits(pid int) (*QuotaLimits, error) {
	// Verify probe library is initialized before reading.
	if err := checkInitialized(); err != nil {
		// Return nil with initialization error.
		return nil, err
	}

	var cLimits C.QuotaLimits
	result := C.probe_quota_read_limits(C.int32_t(pid), &cLimits)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return nil with operation error.
		return nil, err
	}

	// Return converted quota limits.
	return &QuotaLimits{
		CPUQuotaUS:       uint64(cLimits.cpu_quota_us),
		CPUPeriodUS:      uint64(cLimits.cpu_period_us),
		MemoryLimitBytes: uint64(cLimits.memory_limit_bytes),
		PIDsLimit:        uint64(cLimits.pids_limit),
		NofileLimit:      uint64(cLimits.nofile_limit),
		CPUTimeLimitSecs: uint64(cLimits.cpu_time_limit_secs),
		DataLimitBytes:   uint64(cLimits.data_limit_bytes),
		IOReadBPS:        uint64(cLimits.io_read_bps),
		IOWriteBPS:       uint64(cLimits.io_write_bps),
		Flags:            uint32(cLimits.flags),
	}, nil
}

// ReadQuotaUsage reads current resource usage for a process.
//
// Params:
//   - pid: process ID (use os.Getpid() for current process)
//
// Returns:
//   - *QuotaUsage: current usage
//   - error: nil on success, error if probe not initialized or operation fails
func ReadQuotaUsage(pid int) (*QuotaUsage, error) {
	// Verify probe library is initialized before reading.
	if err := checkInitialized(); err != nil {
		// Return nil with initialization error.
		return nil, err
	}

	var cUsage C.QuotaUsage
	result := C.probe_quota_read_usage(C.int32_t(pid), &cUsage)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return nil with operation error.
		return nil, err
	}

	// Return converted quota usage.
	return &QuotaUsage{
		MemoryBytes:      uint64(cUsage.memory_bytes),
		MemoryLimitBytes: uint64(cUsage.memory_limit_bytes),
		PIDsCurrent:      uint64(cUsage.pids_current),
		PIDsLimit:        uint64(cUsage.pids_limit),
		CPUPercent:       float64(cUsage.cpu_percent),
		CPULimitPercent:  float64(cUsage.cpu_limit_percent),
	}, nil
}

// DetectContainer detects the container runtime.
//
// Returns:
//   - *ContainerInfo: container information
//   - error: nil on success, error if probe not initialized or operation fails
func DetectContainer() (*ContainerInfo, error) {
	// Verify probe library is initialized before detecting.
	if err := checkInitialized(); err != nil {
		// Return nil with initialization error.
		return nil, err
	}

	var cInfo C.ContainerInfo
	result := C.probe_detect_container(&cInfo)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return nil with operation error.
		return nil, err
	}

	// Convert container ID from C char array to Go string.
	containerID := C.GoString(&cInfo.container_id[0])

	// Return detected container information.
	return &ContainerInfo{
		IsContainerized: bool(cInfo.is_containerized),
		Runtime:         ContainerRuntime(cInfo.runtime),
		ContainerID:     containerID,
	}, nil
}
