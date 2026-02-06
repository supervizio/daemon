//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
*/
import "C"

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
