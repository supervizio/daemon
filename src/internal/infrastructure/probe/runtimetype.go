//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RuntimeType represents a container/orchestrator runtime type.
type RuntimeType int

// Runtime type constants.
const (
	// RuntimeNone indicates no runtime / not containerized.
	RuntimeNone RuntimeType = 0

	// Container runtimes (1-19)
	// RuntimeDocker indicates Docker runtime.
	RuntimeDocker RuntimeType = 1
	// RuntimePodman indicates Podman runtime.
	RuntimePodman RuntimeType = 2
	// RuntimeContainerd indicates containerd runtime.
	RuntimeContainerd RuntimeType = 3
	// RuntimeCriO indicates CRI-O runtime.
	RuntimeCriO RuntimeType = 4
	// RuntimeLXC indicates LXC runtime.
	RuntimeLXC RuntimeType = 5
	// RuntimeLXD indicates LXD runtime.
	RuntimeLXD RuntimeType = 6
	// RuntimeSystemdNspawn indicates systemd-nspawn runtime.
	RuntimeSystemdNspawn RuntimeType = 7
	// RuntimeFirecracker indicates Firecracker runtime.
	RuntimeFirecracker RuntimeType = 8
	// RuntimeFreeBSDJail indicates FreeBSD jail runtime.
	RuntimeFreeBSDJail RuntimeType = 9

	// Orchestrators (20-39)
	// RuntimeKubernetes indicates Kubernetes orchestrator.
	RuntimeKubernetes RuntimeType = 20
	// RuntimeNomad indicates Nomad orchestrator.
	RuntimeNomad RuntimeType = 21
	// RuntimeDockerSwarm indicates Docker Swarm orchestrator.
	RuntimeDockerSwarm RuntimeType = 22
	// RuntimeOpenShift indicates OpenShift orchestrator.
	RuntimeOpenShift RuntimeType = 23

	// Cloud-specific (40-59)
	// RuntimeAWSECS indicates AWS ECS runtime.
	RuntimeAWSECS RuntimeType = 40
	// RuntimeAWSFargate indicates AWS Fargate runtime.
	RuntimeAWSFargate RuntimeType = 41
	// RuntimeGoogleGKE indicates Google GKE runtime.
	RuntimeGoogleGKE RuntimeType = 42
	// RuntimeAzureAKS indicates Azure AKS runtime.
	RuntimeAzureAKS RuntimeType = 43

	// RuntimeUnknown indicates unknown runtime.
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
