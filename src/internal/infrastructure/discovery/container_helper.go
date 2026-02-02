//go:build unix

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// Doer abstracts HTTP client operations for testing and flexibility.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// errUnexpectedContainerStatus is returned when container runtime API returns non-OK status.
var errUnexpectedContainerStatus error = errors.New("unexpected status code")

// fetchContainers fetches containers from a container runtime API.
// It handles HTTP request creation, execution, and JSON decoding for Docker-compatible APIs.
//
// Params:
//   - ctx: context for cancellation.
//   - client: HTTP client configured for the runtime socket.
//   - apiURL: the API URL (e.g., "http://docker/containers/json").
//   - runtimeName: name of the runtime for error messages (e.g., "docker", "podman").
//
// Returns:
//   - []dockerContainer: the fetched containers.
//   - error: any error during fetch.
func fetchContainers(
	ctx context.Context,
	client Doer,
	apiURL, runtimeName string,
) ([]dockerContainer, error) {
	// Build API request for container list.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, http.NoBody)
	// Check for request creation error.
	if err != nil {
		// Return error with request context.
		return nil, fmt.Errorf("create %s request: %w", runtimeName, err)
	}

	// Execute request against runtime API.
	resp, err := client.Do(req)
	// Check for API communication error.
	if err != nil {
		// Return error with API context.
		return nil, fmt.Errorf("%s api request: %w", runtimeName, err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Verify successful response from runtime API.
	if resp.StatusCode != http.StatusOK {
		// Return error for non-OK status.
		return nil, fmt.Errorf("%s api: %w (status %d)", runtimeName, errUnexpectedContainerStatus, resp.StatusCode)
	}

	// Parse JSON response into container structs.
	var containers []dockerContainer
	// Check for JSON decoding error.
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		// Return error with decode context.
		return nil, fmt.Errorf("decode %s response: %w", runtimeName, err)
	}

	// Return fetched containers.
	return containers, nil
}

// containerToTargetParams holds parameters for container to target conversion.
type containerToTargetParams struct {
	Container      dockerContainer
	RuntimePrefix  string      // e.g., "docker" or "podman"
	TargetType     target.Type // e.g., target.TypeDocker
	MetadataLabels int         // Number of extra labels to allocate
	ProbeType      string      // e.g., "tcp"
}

// containerToExternalTarget converts a Docker-compatible container to an ExternalTarget.
// This shared function is used by both Docker and Podman discoverers.
//
// Params:
//   - params: conversion parameters
//
// Returns:
//   - target.ExternalTarget: the external target
func containerToExternalTarget(params containerToTargetParams) target.ExternalTarget {
	container := params.Container

	// Extract container name with fallback to truncated ID.
	name := container.ID[:containerIDDisplayLength]
	// Check if container has named aliases.
	if len(container.Names) > 0 {
		name = strings.TrimPrefix(container.Names[0], "/")
	}

	// Initialize target with runtime-specific configuration.
	t := target.ExternalTarget{
		ID:               params.RuntimePrefix + ":" + container.ID[:containerIDDisplayLength],
		Name:             name,
		Type:             params.TargetType,
		Source:           target.SourceDiscovered,
		Labels:           make(map[string]string, len(container.Labels)+params.MetadataLabels),
		Interval:         defaultProbeInterval,
		Timeout:          defaultProbeTimeout,
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}

	// Copy all container labels to target labels.
	maps.Copy(t.Labels, container.Labels)

	// Add runtime-specific metadata as labels.
	t.Labels[params.RuntimePrefix+".state"] = container.State
	t.Labels[params.RuntimePrefix+".status"] = container.Status

	// Configure TCP probe based on exposed ports.
	configureContainerProbe(&t, container, params.ProbeType)

	// Return fully configured target with probe.
	return t
}

// configureContainerProbe configures the probe for a container based on its ports.
// It prefers public (host) ports over private (container) ports for TCP probes.
//
// Params:
//   - t: the target to configure.
//   - container: the Docker-compatible container.
//   - probeType: the probe type string (e.g., "tcp").
func configureContainerProbe(t *target.ExternalTarget, container dockerContainer, probeType string) {
	// Find first TCP port with public mapping for external accessibility.
	for _, port := range container.Ports {
		// Check for TCP port with public mapping.
		if port.Type == probeType && port.PublicPort > 0 {
			addr := fmt.Sprintf("127.0.0.1:%d", port.PublicPort)
			t.ProbeType = probeType
			t.ProbeTarget = health.NewTCPTarget(addr)
			// Return after configuring with first public port.
			return
		}
	}

	// Fallback to first private port if no public port exists.
	if len(container.Ports) > 0 {
		port := container.Ports[0]
		addr := fmt.Sprintf("127.0.0.1:%d", port.PrivatePort)
		t.ProbeType = probeType
		t.ProbeTarget = health.NewTCPTarget(addr)
	}
}
