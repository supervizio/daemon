//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net"
	"net/http"
	"strings"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// dockerProbeTypeTCP is the TCP probe type constant.
const dockerProbeTypeTCP string = "tcp"

// DockerDiscoverer discovers Docker containers via the Docker Engine API.
// It connects to the Docker socket and queries running containers.
// Containers are filtered by labels and automatically configured with TCP probes.
type DockerDiscoverer struct {
	// socketPath is the path to the Docker socket.
	socketPath string

	// labelFilter are required labels for container selection.
	labelFilter map[string]string

	// client is the HTTP client for Docker API requests.
	client *http.Client
}

// NewDockerDiscoverer creates a new Docker discoverer.
// It configures an HTTP client with Unix socket transport for Docker API communication.
//
// Params:
//   - socketPath: path to the Docker socket.
//   - labels: required labels for container filtering.
//
// Returns:
//   - *DockerDiscoverer: a new Docker discoverer.
func NewDockerDiscoverer(socketPath string, labels map[string]string) *DockerDiscoverer {
	// Create HTTP client with Unix socket transport for Docker API.
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
	return &DockerDiscoverer{
		socketPath:  socketPath,
		labelFilter: labels,
		client:      client,
	}
}

// Discover finds all running Docker containers matching the label filter.
// It queries the Docker Engine API and converts matching containers to ExternalTargets.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []target.ExternalTarget: the discovered containers.
//   - error: any error during discovery.
func (d *DockerDiscoverer) Discover(ctx context.Context) ([]target.ExternalTarget, error) {
	// Build API request for container list.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://docker/containers/json", nil)
	// Check for request creation error.
	if err != nil {
		// Return error with request context.
		return nil, fmt.Errorf("create docker request: %w", err)
	}

	// Execute request against Docker API.
	resp, err := d.client.Do(req)
	// Check for API communication error.
	if err != nil {
		// Return error with API context.
		return nil, fmt.Errorf("docker api request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Verify successful response from Docker API.
	if resp.StatusCode != http.StatusOK {
		// Return error for non-OK status.
		return nil, fmt.Errorf("docker api returned status %d: %w", resp.StatusCode, err)
	}

	// Parse JSON response into container structs.
	var containers []dockerContainer
	// Check for JSON decoding error.
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		// Return error with decode context.
		return nil, fmt.Errorf("decode docker response: %w", err)
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

// Type returns the target type for Docker.
//
// Returns:
//   - target.Type: TypeDocker.
func (d *DockerDiscoverer) Type() target.Type {
	// Return Docker type constant for this discoverer.
	return target.TypeDocker
}

// matchesLabels checks if a container has all required labels.
// Returns true if no filter is set (accept all) or if all filter labels match.
//
// Params:
//   - container: the container to check.
//
// Returns:
//   - bool: true if container has all required labels.
func (d *DockerDiscoverer) matchesLabels(container dockerContainer) bool {
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

// containerToTarget converts a Docker container to an ExternalTarget.
// It extracts metadata, configures probes from exposed ports, and sets default thresholds.
//
// Params:
//   - container: the Docker container.
//
// Returns:
//   - target.ExternalTarget: the external target.
func (d *DockerDiscoverer) containerToTarget(container dockerContainer) target.ExternalTarget {
	// Extract container name with fallback to truncated ID.
	name := container.ID[:containerIDDisplayLength]
	// Check if container has named aliases.
	if len(container.Names) > 0 {
		name = strings.TrimPrefix(container.Names[0], "/")
	}

	// Initialize target with Docker-specific configuration.
	t := target.ExternalTarget{
		ID:               "docker:" + container.ID[:containerIDDisplayLength],
		Name:             name,
		Type:             target.TypeDocker,
		Source:           target.SourceDiscovered,
		Labels:           make(map[string]string, len(container.Labels)+dockerMetadataLabels),
		Interval:         defaultProbeInterval,
		Timeout:          defaultProbeTimeout,
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}

	// Copy all container labels to target labels.
	maps.Copy(t.Labels, container.Labels)

	// Add Docker-specific metadata as labels.
	t.Labels["docker.state"] = container.State
	t.Labels["docker.status"] = container.Status

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
//   - container: the Docker container.
func (d *DockerDiscoverer) configureProbe(t *target.ExternalTarget, container dockerContainer) {
	// Find first TCP port with public mapping for external accessibility.
	for _, port := range container.Ports {
		// Check for TCP port with public mapping.
		if port.Type == dockerProbeTypeTCP && port.PublicPort > 0 {
			addr := fmt.Sprintf("127.0.0.1:%d", port.PublicPort)
			t.ProbeType = dockerProbeTypeTCP
			t.ProbeTarget = health.NewTCPTarget(addr)
			// Return after configuring with first public port.
			return
		}
	}

	// Fallback to first private port if no public port exists.
	if len(container.Ports) > 0 {
		port := container.Ports[0]
		addr := fmt.Sprintf("127.0.0.1:%d", port.PrivatePort)
		t.ProbeType = dockerProbeTypeTCP
		t.ProbeTarget = health.NewTCPTarget(addr)
	}
}
