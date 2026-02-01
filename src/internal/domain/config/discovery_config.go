// Package config provides domain value objects for service configuration.
package config

// DiscoveryConfig configures auto-discovery for different platforms.
// Each field enables discovery for a specific platform or container runtime.
type DiscoveryConfig struct {
	// Systemd configures systemd service discovery (Linux only).
	Systemd *SystemdDiscoveryConfig

	// OpenRC configures OpenRC service discovery (Alpine/Gentoo).
	OpenRC *OpenRCDiscoveryConfig

	// BSDRC configures BSD rc.d service discovery.
	BSDRC *BSDRCDiscoveryConfig

	// Docker configures Docker container discovery.
	Docker *DockerDiscoveryConfig

	// Podman configures Podman container discovery.
	Podman *PodmanDiscoveryConfig

	// Kubernetes configures Kubernetes pod/service discovery.
	Kubernetes *KubernetesDiscoveryConfig

	// Nomad configures Nomad allocation discovery.
	Nomad *NomadDiscoveryConfig
}

// hasInitSystemDiscovery checks if init system discovery is enabled.
//
// Returns:
//   - bool: true if systemd, openrc, or bsdrc discovery is enabled.
func (d *DiscoveryConfig) hasInitSystemDiscovery() bool {
	// check init systems: systemd, openrc, bsdrc
	return (d.Systemd != nil && d.Systemd.Enabled) ||
		(d.OpenRC != nil && d.OpenRC.Enabled) ||
		(d.BSDRC != nil && d.BSDRC.Enabled)
}

// hasContainerDiscovery checks if container runtime discovery is enabled.
//
// Returns:
//   - bool: true if docker or podman discovery is enabled.
func (d *DiscoveryConfig) hasContainerDiscovery() bool {
	// check container runtimes: docker, podman
	return (d.Docker != nil && d.Docker.Enabled) ||
		(d.Podman != nil && d.Podman.Enabled)
}

// hasOrchestratorDiscovery checks if orchestrator discovery is enabled.
//
// Returns:
//   - bool: true if kubernetes or nomad discovery is enabled.
func (d *DiscoveryConfig) hasOrchestratorDiscovery() bool {
	// check orchestrators: kubernetes, nomad
	return (d.Kubernetes != nil && d.Kubernetes.Enabled) ||
		(d.Nomad != nil && d.Nomad.Enabled)
}
