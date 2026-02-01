// Package target provides domain entities for external monitoring targets.
// External targets are services, containers, or hosts that supervizio
// monitors but does not manage (no lifecycle control).
package target

// Type represents the kind of external target being monitored.
type Type string

// Target type constants define the supported external target types.
const (
	// TypeSystemd represents a systemd service unit.
	TypeSystemd Type = "systemd"

	// TypeOpenRC represents an OpenRC service (Alpine/Gentoo).
	TypeOpenRC Type = "openrc"

	// TypeBSDRC represents a BSD rc.d service.
	TypeBSDRC Type = "bsd-rc"

	// TypeDocker represents a Docker container.
	TypeDocker Type = "docker"

	// TypePodman represents a Podman container.
	TypePodman Type = "podman"

	// TypeKubernetes represents a Kubernetes pod or service.
	TypeKubernetes Type = "kubernetes"

	// TypeNomad represents a Nomad allocation or task.
	TypeNomad Type = "nomad"

	// TypeRemote represents a remote host or endpoint.
	// Used for ICMP ping or TCP/HTTP probes to external hosts.
	TypeRemote Type = "remote"

	// TypeCustom represents a custom target with exec probe.
	TypeCustom Type = "custom"
)

// String returns the string representation of the target type.
//
// Returns:
//   - string: the target type as a string.
func (t Type) String() string {
	// Convert Type enum to string for display and serialization.
	return string(t)
}

// IsValid checks if the target type is a known valid type.
//
// Returns:
//   - bool: true if the type is valid.
func (t Type) IsValid() bool {
	// Use switch to validate against all known Type constants.
	switch t {
	// Match all known target types across init systems, containers, and orchestrators.
	case TypeSystemd, TypeOpenRC, TypeBSDRC,
		TypeDocker, TypePodman,
		TypeKubernetes, TypeNomad,
		TypeRemote, TypeCustom:
		// Type is one of the known valid types.
		return true
	// Handle any unrecognized types as invalid.
	default:
		// Type is not recognized.
		return false
	}
}

// IsContainerRuntime checks if the target type is a container runtime.
//
// Returns:
//   - bool: true if the type is docker or podman.
func (t Type) IsContainerRuntime() bool {
	// Check if type matches any container runtime constant.
	return t == TypeDocker || t == TypePodman
}

// IsOrchestrator checks if the target type is an orchestrator.
//
// Returns:
//   - bool: true if the type is kubernetes or nomad.
func (t Type) IsOrchestrator() bool {
	// Check if type matches any orchestrator constant.
	return t == TypeKubernetes || t == TypeNomad
}

// IsInitSystem checks if the target type is an init system.
//
// Returns:
//   - bool: true if the type is systemd, openrc, or bsd-rc.
func (t Type) IsInitSystem() bool {
	// Check if type matches any init system constant.
	return t == TypeSystemd || t == TypeOpenRC || t == TypeBSDRC
}
