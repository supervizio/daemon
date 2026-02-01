//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
//
//nolint:ktn-struct-onefile // Runtime structs (AvailableRuntime, RuntimeInfo) are logically grouped
package probe

/*
#include "probe.h"
*/
import "C"

// RuntimeType represents a container/orchestrator runtime type.
type RuntimeType int

// Runtime type constants.
const (
	// No runtime / not containerized.
	RuntimeNone RuntimeType = 0

	// Container runtimes (1-19)
	RuntimeDocker        RuntimeType = 1
	RuntimePodman        RuntimeType = 2
	RuntimeContainerd    RuntimeType = 3
	RuntimeCriO          RuntimeType = 4
	RuntimeLXC           RuntimeType = 5
	RuntimeLXD           RuntimeType = 6
	RuntimeSystemdNspawn RuntimeType = 7
	RuntimeFirecracker   RuntimeType = 8
	RuntimeFreeBSDJail   RuntimeType = 9

	// Orchestrators (20-39)
	RuntimeKubernetes  RuntimeType = 20
	RuntimeNomad       RuntimeType = 21
	RuntimeDockerSwarm RuntimeType = 22
	RuntimeOpenShift   RuntimeType = 23

	// Cloud-specific (40-59)
	RuntimeAWSECS     RuntimeType = 40
	RuntimeAWSFargate RuntimeType = 41
	RuntimeGoogleGKE  RuntimeType = 42
	RuntimeAzureAKS   RuntimeType = 43

	// Unknown
	RuntimeUnknown RuntimeType = 254
)

// runtimeNames maps runtime types to their string representations.
var runtimeNames map[RuntimeType]string = map[RuntimeType]string{
	RuntimeNone:          "none",
	RuntimeDocker:        "docker",
	RuntimePodman:        "podman",
	RuntimeContainerd:    "containerd",
	RuntimeCriO:          "cri-o",
	RuntimeLXC:           "lxc",
	RuntimeLXD:           "lxd",
	RuntimeSystemdNspawn: "systemd-nspawn",
	RuntimeFirecracker:   "firecracker",
	RuntimeFreeBSDJail:   "freebsd-jail",
	RuntimeKubernetes:    "kubernetes",
	RuntimeNomad:         "nomad",
	RuntimeDockerSwarm:   "docker-swarm",
	RuntimeOpenShift:     "openshift",
	RuntimeAWSECS:        "aws-ecs",
	RuntimeAWSFargate:    "aws-fargate",
	RuntimeGoogleGKE:     "google-gke",
	RuntimeAzureAKS:      "azure-aks",
	RuntimeUnknown:       "unknown",
}

// String returns the string representation of the runtime type.
//
// Returns:
//   - string: human-readable name of the runtime type
func (r RuntimeType) String() string {
	// Look up runtime name in the map
	if name, ok := runtimeNames[r]; ok {
		// Return the mapped name
		return name
	}
	// Return unknown for any unrecognized value
	return "unknown"
}

// IsOrchestrator returns whether this is an orchestrator (vs a container runtime).
//
// Returns:
//   - bool: true if runtime type is an orchestrator, false otherwise
//
//nolint:exhaustive // Only orchestrator cases return true; all other RuntimeType values return false.
func (r RuntimeType) IsOrchestrator() bool {
	// Check if runtime type is in orchestrator range
	switch r {
	// Handle orchestrator runtime types
	case RuntimeKubernetes, RuntimeNomad, RuntimeDockerSwarm, RuntimeOpenShift,
		RuntimeAWSECS, RuntimeAWSFargate, RuntimeGoogleGKE, RuntimeAzureAKS:
		// Return true for orchestrator types
		return true
	// Handle all other runtime types
	default:
		// Return false for container runtimes and unknown types
		return false
	}
}

// AvailableRuntime describes a runtime available on the host.
// It includes connection details and version information.
type AvailableRuntime struct {
	// Runtime type.
	Runtime RuntimeType

	// SocketPath is the Unix socket path (empty if not available).
	SocketPath string

	// Version is the version string (empty if not available).
	Version string

	// IsRunning indicates whether the runtime is currently responsive.
	IsRunning bool
}

// RuntimeInfo contains full runtime environment detection results.
// It describes containerization state and available runtimes.
//
//nolint:ktn-struct-onefile // grouped with runtime types
type RuntimeInfo struct {
	// IsContainerized indicates whether running inside a container.
	IsContainerized bool

	// ContainerRuntime is the container runtime (if containerized).
	ContainerRuntime RuntimeType

	// Orchestrator is the orchestrator (may differ from runtime).
	Orchestrator RuntimeType

	// ContainerID is the container ID (64-char hex for Docker, varies by runtime).
	ContainerID string

	// WorkloadID is the workload/allocation ID (Nomad alloc ID, K8s pod UID, etc.).
	WorkloadID string

	// WorkloadName is the workload/pod name.
	WorkloadName string

	// Namespace is the namespace (K8s namespace, Nomad namespace, etc.).
	Namespace string

	// AvailableRuntimes lists runtimes available on the host.
	AvailableRuntimes []AvailableRuntime
}

// DetectRuntime performs full runtime environment detection.
// This detects whether running inside a container, the container runtime
// and orchestrator, container/workload IDs and names, and available runtimes.
//
// Returns:
//   - *RuntimeInfo: full runtime environment information
//   - error: nil on success, error if probe not initialized or detection fails
//
//nolint:ktn-comment-func // Returns section is present, no params needed
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
//
//nolint:ktn-comment-func // Returns section is present, no params needed
func IsContainerized() bool {
	return bool(C.probe_is_containerized())
}

// GetRuntimeName returns the container runtime name as a string.
// Returns "none" if not containerized.
//
// Returns:
//   - string: name of the current container runtime or "none"
//
//nolint:ktn-comment-func // Returns section is present, no params needed
func GetRuntimeName() string {
	return C.GoString(C.probe_get_runtime_name())
}
