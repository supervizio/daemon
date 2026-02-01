// Package discovery provides infrastructure adapters for target discovery.
package discovery

import "time"

// Network timeouts for Docker API communication.
const (
	// dockerDialTimeout is the timeout for establishing socket connections.
	dockerDialTimeout time.Duration = 5 * time.Second

	// dockerRequestTimeout is the timeout for completing HTTP requests.
	dockerRequestTimeout time.Duration = 10 * time.Second
)

// Container ID display length constants.
const (
	// containerIDDisplayLength is the truncated length for container IDs in names/labels.
	containerIDDisplayLength int = 12
)

// Docker metadata label count.
const (
	// dockerMetadataLabels is the number of Docker-specific metadata labels added to targets.
	dockerMetadataLabels int = 2
)

// Default probe configuration for discovered targets.
const (
	// defaultProbeInterval is the default time between health checks.
	defaultProbeInterval time.Duration = 30 * time.Second

	// defaultProbeTimeout is the default timeout for individual probes.
	defaultProbeTimeout time.Duration = 5 * time.Second

	// defaultProbeSuccessThreshold is the consecutive successes needed for healthy status.
	defaultProbeSuccessThreshold int = 1

	// defaultProbeFailureThreshold is the consecutive failures needed for unhealthy status.
	defaultProbeFailureThreshold int = 3
)

// HTTP probe configuration.
const (
	// defaultHTTPStatusCode is the expected HTTP status code for successful health checks.
	defaultHTTPStatusCode int = 200
)

// Kubernetes API timeouts and metadata.
const (
	// kubernetesRequestTimeout is the timeout for Kubernetes API requests.
	kubernetesRequestTimeout time.Duration = 30 * time.Second

	// kubernetesMetadataLabels is the number of Kubernetes-specific metadata labels.
	kubernetesMetadataLabels int = 3
)
