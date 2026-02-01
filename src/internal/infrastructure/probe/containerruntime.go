//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

// containerRuntimeUnknownStr is the string for unknown runtimes.
const containerRuntimeUnknownStr string = "unknown"

// ContainerRuntime represents a container runtime type.
// It identifies the orchestration platform or isolation mechanism.
type ContainerRuntime int

// Container runtime constants.
const (
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
