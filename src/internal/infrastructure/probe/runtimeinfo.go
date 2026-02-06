//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RuntimeInfo contains full runtime environment detection results.
// It describes containerization state and available runtimes.
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
