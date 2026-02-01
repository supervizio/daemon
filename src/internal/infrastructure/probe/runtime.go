//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
*/
import "C"

// DetectRuntime performs full runtime environment detection.
// This detects whether running inside a container, the container runtime
// and orchestrator, container/workload IDs and names, and available runtimes.
//
// Returns:
//   - *RuntimeInfo: full runtime environment information
//   - error: nil on success, error if probe not initialized or detection fails
func DetectRuntime() (*RuntimeInfo, error) {
	// Check if probe is initialized
	if err := checkInitialized(); err != nil {
		// Return early if not initialized
		return nil, err
	}

	var cInfo C.RuntimeInfo
	result := C.probe_detect_runtime(&cInfo)
	// Check detection result
	if err := resultToError(result); err != nil {
		// Return early on detection failure
		return nil, err
	}

	// Build RuntimeInfo from C struct
	info := &RuntimeInfo{
		IsContainerized:  bool(cInfo.is_containerized),
		ContainerRuntime: RuntimeType(cInfo.container_runtime),
		Orchestrator:     RuntimeType(cInfo.orchestrator),
		ContainerID:      C.GoString(&cInfo.container_id[0]),
		WorkloadID:       C.GoString(&cInfo.workload_id[0]),
		WorkloadName:     C.GoString(&cInfo.workload_name[0]),
		Namespace:        C.GoString(&cInfo.namespace[0]),
	}

	// Convert available runtimes from C array
	count := int(cInfo.available_count)
	// Check if any runtimes are available
	if count > 0 {
		info.AvailableRuntimes = make([]AvailableRuntime, 0, count)
		// Iterate over available runtimes
		for idx := range count {
			crt := cInfo.available_runtimes[idx]
			info.AvailableRuntimes = append(info.AvailableRuntimes, AvailableRuntime{
				Runtime:    RuntimeType(crt.runtime),
				SocketPath: C.GoString(&crt.socket_path[0]),
				Version:    C.GoString(&crt.version[0]),
				IsRunning:  bool(crt.is_running),
			})
		}
	}

	// Return the populated runtime info
	return info, nil
}

// IsContainerized returns whether running inside a container (fast check).
// This only checks for containerization, not available runtimes.
//
// Returns:
//   - bool: true if running inside a container, false otherwise
func IsContainerized() bool {
	return bool(C.probe_is_containerized())
}

// GetRuntimeName returns the container runtime name as a string.
// Returns "none" if not containerized.
//
// Returns:
//   - string: name of the current container runtime or "none"
func GetRuntimeName() string {
	return C.GoString(C.probe_get_runtime_name())
}
