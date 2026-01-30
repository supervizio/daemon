//go:build cgo

package probe

/*
#include "probe.h"
*/
import "C"

// QuotaLimits represents detected resource limits for a process.
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

// Quota limit flags.
const (
	QuotaFlagCPU     uint32 = 1 << 0
	QuotaFlagMemory  uint32 = 1 << 1
	QuotaFlagPIDs    uint32 = 1 << 2
	QuotaFlagNofile  uint32 = 1 << 3
	QuotaFlagCPUTime uint32 = 1 << 4
	QuotaFlagData    uint32 = 1 << 5
	QuotaFlagIORead  uint32 = 1 << 6
	QuotaFlagIOWrite uint32 = 1 << 7
)

// HasCPULimit returns whether a CPU limit is set.
func (l *QuotaLimits) HasCPULimit() bool {
	return l.Flags&QuotaFlagCPU != 0 && l.CPUQuotaUS > 0 && l.CPUQuotaUS != ^uint64(0)
}

// HasMemoryLimit returns whether a memory limit is set.
func (l *QuotaLimits) HasMemoryLimit() bool {
	return l.Flags&QuotaFlagMemory != 0 && l.MemoryLimitBytes > 0 && l.MemoryLimitBytes != ^uint64(0)
}

// CPULimitPercent calculates the CPU limit as a percentage.
// Returns 0 if no limit is set.
func (l *QuotaLimits) CPULimitPercent() float64 {
	if !l.HasCPULimit() || l.CPUPeriodUS == 0 {
		return 0
	}
	return float64(l.CPUQuotaUS) / float64(l.CPUPeriodUS) * 100.0
}

// QuotaUsage represents current resource usage for a process.
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

// MemoryUsagePercent calculates memory usage as percentage of limit.
// Returns 0 if no limit is set.
func (u *QuotaUsage) MemoryUsagePercent() float64 {
	if u.MemoryLimitBytes == 0 || u.MemoryLimitBytes == ^uint64(0) {
		return 0
	}
	return float64(u.MemoryBytes) / float64(u.MemoryLimitBytes) * 100.0
}

// ContainerRuntime represents a container runtime type.
type ContainerRuntime int

// Container runtime constants.
const (
	ContainerRuntimeNone       ContainerRuntime = 0
	ContainerRuntimeDocker     ContainerRuntime = 1
	ContainerRuntimePodman     ContainerRuntime = 2
	ContainerRuntimeLXC        ContainerRuntime = 3
	ContainerRuntimeKubernetes ContainerRuntime = 4
	ContainerRuntimeJail       ContainerRuntime = 5
	ContainerRuntimeUnknown    ContainerRuntime = 255
)

// String returns the string representation of the container runtime.
func (r ContainerRuntime) String() string {
	switch r {
	case ContainerRuntimeNone:
		return "none"
	case ContainerRuntimeDocker:
		return "docker"
	case ContainerRuntimePodman:
		return "podman"
	case ContainerRuntimeLXC:
		return "lxc"
	case ContainerRuntimeKubernetes:
		return "kubernetes"
	case ContainerRuntimeJail:
		return "jail"
	case ContainerRuntimeUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// ContainerInfo represents container detection results.
type ContainerInfo struct {
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
//   - error: if the operation fails
func ReadQuotaLimits(pid int) (*QuotaLimits, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var cLimits C.QuotaLimits
	result := C.probe_quota_read_limits(C.int32_t(pid), &cLimits)
	if err := resultToError(result); err != nil {
		return nil, err
	}

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
//   - error: if the operation fails
func ReadQuotaUsage(pid int) (*QuotaUsage, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var cUsage C.QuotaUsage
	result := C.probe_quota_read_usage(C.int32_t(pid), &cUsage)
	if err := resultToError(result); err != nil {
		return nil, err
	}

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
//   - error: if the operation fails
func DetectContainer() (*ContainerInfo, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var cInfo C.ContainerInfo
	result := C.probe_detect_container(&cInfo)
	if err := resultToError(result); err != nil {
		return nil, err
	}

	// Convert container ID from C char array to Go string
	containerID := C.GoString(&cInfo.container_id[0])

	return &ContainerInfo{
		IsContainerized: bool(cInfo.is_containerized),
		Runtime:         ContainerRuntime(cInfo.runtime),
		ContainerID:     containerID,
	}, nil
}
