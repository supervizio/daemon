//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"context"
	"fmt"
	"maps"
	"net"
	"net/http"
	"strings"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// podmanProbeTypeTCP is the TCP probe type constant.
const podmanProbeTypeTCP string = "tcp"

// podmanMetadataLabels is the number of Podman-specific metadata labels added to targets.
const podmanMetadataLabels int = 2

// PodmanDiscoverer discovers Podman containers via the Podman Engine API.
// It connects to the Podman socket and queries running containers.
// Containers are filtered by labels and automatically configured with TCP probes.
// Podman's API is compatible with Docker's REST API, so we reuse dockerContainer types.
type PodmanDiscoverer struct {
	// socketPath is the path to the Podman socket.
	socketPath string

	// labelFilter are required labels for container selection.
	labelFilter map[string]string

	// client is the HTTP client for Podman API requests.
	client *http.Client
}

// NewPodmanDiscoverer creates a new Podman discoverer.
// It configures an HTTP client with Unix socket transport for Podman API communication.
//
// Params:
//   - socketPath: path to the Podman socket.
//   - labels: required labels for container filtering.
//
// Returns:
//   - *PodmanDiscoverer: a new Podman discoverer.
func NewPodmanDiscoverer(socketPath string, labels map[string]string) *PodmanDiscoverer {
	// Create HTTP client with Unix socket transport for Podman API.
	dialer := &net.Dialer{
		Timeout: dockerDialTimeout,
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			// Connect to Unix socket regardless of network/addr parameters.
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   dockerRequestTimeout,
	}

	// Construct discoverer with socket path and label filter.
	return &PodmanDiscoverer{
		socketPath:  socketPath,
		labelFilter: labels,
		client:      client,
	}
}

// Discover finds all running Podman containers matching the label filter.
// It queries the Podman Engine API and converts matching containers to ExternalTargets.
// Podman's /containers/json endpoint is compatible with Docker's API.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []target.ExternalTarget: the discovered containers.
//   - error: any error during discovery.
func (d *PodmanDiscoverer) Discover(ctx context.Context) ([]target.ExternalTarget, error) {
	// Fetch containers using shared helper for Docker-compatible APIs.
	containers, err := fetchContainers(ctx, d.client, "http://podman/containers/json", "podman")
	// Check for fetch error.
	if err != nil {
		// Return error from fetch operation.
		return nil, err
	}

	// Convert matching containers to external targets.
	var targets []target.ExternalTarget
	// Iterate over discovered containers.
	for _, container := range containers {
		// Skip containers that don't match label filter.
		if !d.matchesLabels(container) {
			continue
		}

		t := d.containerToTarget(container)
		targets = append(targets, t)
	}

	// Return all discovered and converted targets.
	return targets, nil
}

// Type returns the target type for Podman.
//
// Returns:
//   - target.Type: TypePodman.
func (d *PodmanDiscoverer) Type() target.Type {
	// Return Podman type constant for this discoverer.
	return target.TypePodman
}

// matchesLabels checks if a container has all required labels.
// Returns true if no filter is set (accept all) or if all filter labels match.
//
// Params:
//   - container: the container to check.
//
// Returns:
//   - bool: true if container has all required labels.
func (d *PodmanDiscoverer) matchesLabels(container dockerContainer) bool {
	// Check if no label filter is configured.
	if len(d.labelFilter) == 0 {
		// Accept all containers when no filter is set.
		return true
	}

	// Verify each required label exists and matches expected value.
	for key, value := range d.labelFilter {
		containerValue, exists := container.Labels[key]
		// Check if label is missing or value doesn't match.
		if !exists || containerValue != value {
			// Reject container with missing or mismatched label.
			return false
		}
	}

	// All labels match - accept container.
	return true
}

// containerToTarget converts a Podman container to an ExternalTarget.
// It extracts metadata, configures probes from exposed ports, and sets default thresholds.
//
// Params:
//   - container: the Podman container.
//
// Returns:
//   - target.ExternalTarget: the external target.
func (d *PodmanDiscoverer) containerToTarget(container dockerContainer) target.ExternalTarget {
	// Extract container name with fallback to truncated ID.
	name := container.ID[:containerIDDisplayLength]
	// Check if container has named aliases.
	if len(container.Names) > 0 {
		name = strings.TrimPrefix(container.Names[0], "/")
	}

	// Initialize target with Podman-specific configuration.
	t := target.ExternalTarget{
		ID:               "podman:" + container.ID[:containerIDDisplayLength],
		Name:             name,
		Type:             target.TypePodman,
		Source:           target.SourceDiscovered,
		Labels:           make(map[string]string, len(container.Labels)+podmanMetadataLabels),
		Interval:         defaultProbeInterval,
		Timeout:          defaultProbeTimeout,
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}

	// Copy all container labels to target labels.
	maps.Copy(t.Labels, container.Labels)

	// Add Podman-specific metadata as labels.
	t.Labels["podman.state"] = container.State
	t.Labels["podman.status"] = container.Status

	// Configure TCP probe based on exposed ports.
	d.configureProbe(&t, container)

	// Return fully configured target with probe.
	return t
}

// configureProbe configures the probe for a container based on its ports.
// It prefers public (host) ports over private (container) ports for TCP probes.
//
// Params:
//   - t: the target to configure.
//   - container: the Podman container.
func (d *PodmanDiscoverer) configureProbe(t *target.ExternalTarget, container dockerContainer) {
	// Find first TCP port with public mapping for external accessibility.
	for _, port := range container.Ports {
		// Check for TCP port with public mapping.
		if port.Type == podmanProbeTypeTCP && port.PublicPort > 0 {
			addr := fmt.Sprintf("127.0.0.1:%d", port.PublicPort)
			t.ProbeType = podmanProbeTypeTCP
			t.ProbeTarget = health.NewTCPTarget(addr)
			// Return after configuring with first public port.
			return
		}
	}

	// Fallback to first private port if no public port exists.
	if len(container.Ports) > 0 {
		port := container.Ports[0]
		addr := fmt.Sprintf("127.0.0.1:%d", port.PrivatePort)
		t.ProbeType = podmanProbeTypeTCP
		t.ProbeTarget = health.NewTCPTarget(addr)
	}
}
