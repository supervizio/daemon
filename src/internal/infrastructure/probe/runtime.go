//go:build cgo

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

// String returns the string representation of the runtime type.
func (r RuntimeType) String() string {
	switch r {
	case RuntimeNone:
		return "none"
	case RuntimeDocker:
		return "docker"
	case RuntimePodman:
		return "podman"
	case RuntimeContainerd:
		return "containerd"
	case RuntimeCriO:
		return "cri-o"
	case RuntimeLXC:
		return "lxc"
	case RuntimeLXD:
		return "lxd"
	case RuntimeSystemdNspawn:
		return "systemd-nspawn"
	case RuntimeFirecracker:
		return "firecracker"
	case RuntimeFreeBSDJail:
		return "freebsd-jail"
	case RuntimeKubernetes:
		return "kubernetes"
	case RuntimeNomad:
		return "nomad"
	case RuntimeDockerSwarm:
		return "docker-swarm"
	case RuntimeOpenShift:
		return "openshift"
	case RuntimeAWSECS:
		return "aws-ecs"
	case RuntimeAWSFargate:
		return "aws-fargate"
	case RuntimeGoogleGKE:
		return "google-gke"
	case RuntimeAzureAKS:
		return "azure-aks"
	case RuntimeUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// IsOrchestrator returns whether this is an orchestrator (vs a container runtime).
func (r RuntimeType) IsOrchestrator() bool {
	switch r {
	case RuntimeKubernetes, RuntimeNomad, RuntimeDockerSwarm, RuntimeOpenShift,
		RuntimeAWSECS, RuntimeAWSFargate, RuntimeGoogleGKE, RuntimeAzureAKS:
		return true
	default:
		return false
	}
}

// AvailableRuntime describes a runtime available on the host.
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
//
// This detects:
// - Whether running inside a container
// - The container runtime and orchestrator
// - Container/workload IDs and names
// - Available runtimes on the host
//
// Returns:
//   - *RuntimeInfo: full runtime environment information
//   - error: if the operation fails
func DetectRuntime() (*RuntimeInfo, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var cInfo C.RuntimeInfo
	result := C.probe_detect_runtime(&cInfo)
	if err := resultToError(result); err != nil {
		return nil, err
	}

	info := &RuntimeInfo{
		IsContainerized:  bool(cInfo.is_containerized),
		ContainerRuntime: RuntimeType(cInfo.container_runtime),
		Orchestrator:     RuntimeType(cInfo.orchestrator),
		ContainerID:      C.GoString(&cInfo.container_id[0]),
		WorkloadID:       C.GoString(&cInfo.workload_id[0]),
		WorkloadName:     C.GoString(&cInfo.workload_name[0]),
		Namespace:        C.GoString(&cInfo.namespace[0]),
	}

	// Convert available runtimes
	count := int(cInfo.available_count)
	if count > 0 {
		info.AvailableRuntimes = make([]AvailableRuntime, count)
		for i := 0; i < count; i++ {
			crt := cInfo.available_runtimes[i]
			info.AvailableRuntimes[i] = AvailableRuntime{
				Runtime:    RuntimeType(crt.runtime),
				SocketPath: C.GoString(&crt.socket_path[0]),
				Version:    C.GoString(&crt.version[0]),
				IsRunning:  bool(crt.is_running),
			}
		}
	}

	return info, nil
}

// IsContainerized returns whether running inside a container (fast check).
//
// This only checks for containerization, not available runtimes.
func IsContainerized() bool {
	return bool(C.probe_is_containerized())
}

// GetRuntimeName returns the container runtime name as a string.
//
// Returns "none" if not containerized.
func GetRuntimeName() string {
	return C.GoString(C.probe_get_runtime_name())
}
