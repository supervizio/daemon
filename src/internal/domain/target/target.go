// Package target provides domain entities for external monitoring targets.
// External targets are services, containers, or hosts that supervizio
// monitors but does not manage (no lifecycle control).
package target

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
)

// ExternalTarget represents an external service, container, or host
// that supervizio monitors but does not manage.
// Unlike managed services, external targets have no lifecycle control.
type ExternalTarget struct {
	// ID is the unique identifier for this target.
	// Format: "<type>:<name>" (e.g., "docker:redis", "systemd:nginx.service").
	ID string

	// Name is the human-readable name for this target.
	Name string

	// Type indicates the kind of target (systemd, docker, kubernetes, etc.).
	Type Type

	// Source indicates how the target was added (static or discovered).
	Source Source

	// Labels are metadata key-value pairs for filtering and grouping.
	Labels map[string]string

	// ProbeType specifies the probe protocol to use.
	// Supported values: "tcp", "udp", "http", "grpc", "exec", "icmp".
	ProbeType string

	// ProbeTarget contains the probe target configuration.
	ProbeTarget health.Target

	// Interval is the time between consecutive probes.
	Interval time.Duration

	// Timeout is the maximum time to wait for a probe response.
	Timeout time.Duration

	// SuccessThreshold is consecutive successes to mark healthy.
	SuccessThreshold int

	// FailureThreshold is consecutive failures to mark unhealthy.
	FailureThreshold int
}

// NewExternalTarget creates a new external target with the specified parameters.
//
// Params:
//   - id: unique identifier for the target.
//   - name: human-readable name.
//   - targetType: the kind of target.
//   - source: how the target was added.
//
// Returns:
//   - *ExternalTarget: a new external target with defaults applied.
func NewExternalTarget(id, name string, targetType Type, source Source) *ExternalTarget {
	// Create target with default timing and threshold values.
	return &ExternalTarget{
		ID:               id,
		Name:             name,
		Type:             targetType,
		Source:           source,
		Labels:           make(map[string]string, defaultMapCapacity),
		Interval:         defaultInterval,
		Timeout:          defaultTimeout,
		SuccessThreshold: defaultSuccessThreshold,
		FailureThreshold: defaultFailureThreshold,
	}
}

// NewRemoteTarget creates a target for monitoring a remote host or endpoint.
//
// Params:
//   - name: the target name.
//   - address: the target address (host:port for TCP/UDP, host for ICMP).
//   - probeType: the probe type ("tcp", "http", "icmp", etc.).
//
// Returns:
//   - *ExternalTarget: a configured remote target.
func NewRemoteTarget(name, address, probeType string) *ExternalTarget {
	id := "remote:" + name
	t := NewExternalTarget(id, name, TypeRemote, SourceStatic)
	t.ProbeType = probeType
	t.ProbeTarget = health.NewTarget(probeType, address)
	// Return configured remote target with probe settings.
	return t
}

// NewDockerTarget creates a target for monitoring a Docker container.
//
// Params:
//   - containerID: the container ID or name.
//   - containerName: the human-readable container name.
//
// Returns:
//   - *ExternalTarget: a configured Docker target.
func NewDockerTarget(containerID, containerName string) *ExternalTarget {
	id := "docker:" + containerID
	t := NewExternalTarget(id, containerName, TypeDocker, SourceDiscovered)
	// Return configured Docker target for container monitoring.
	return t
}

// NewSystemdTarget creates a target for monitoring a systemd service.
//
// Params:
//   - unitName: the systemd unit name (e.g., "nginx.service").
//
// Returns:
//   - *ExternalTarget: a configured systemd target.
func NewSystemdTarget(unitName string) *ExternalTarget {
	id := "systemd:" + unitName
	t := NewExternalTarget(id, unitName, TypeSystemd, SourceDiscovered)
	// Default to exec probe: systemctl is-active <unit>
	t.ProbeType = "exec"
	t.ProbeTarget = health.NewExecTarget("systemctl", "is-active", "--quiet", unitName)
	// Return configured systemd target with default exec probe.
	return t
}

// NewKubernetesTarget creates a target for monitoring a Kubernetes resource.
//
// Params:
//   - namespace: the Kubernetes namespace.
//   - resourceType: the resource type ("pod", "service", "deployment").
//   - resourceName: the resource name.
//
// Returns:
//   - *ExternalTarget: a configured Kubernetes target.
func NewKubernetesTarget(namespace, resourceType, resourceName string) *ExternalTarget {
	id := "kubernetes:" + namespace + "/" + resourceType + "/" + resourceName
	name := resourceName
	t := NewExternalTarget(id, name, TypeKubernetes, SourceDiscovered)
	t.Labels["namespace"] = namespace
	t.Labels["resource_type"] = resourceType
	// Return configured Kubernetes target with resource labels.
	return t
}

// NewNomadTarget creates a target for monitoring a Nomad allocation.
//
// Params:
//   - allocID: the Nomad allocation ID.
//   - taskName: the task name within the allocation.
//   - jobName: the job name for display.
//
// Returns:
//   - *ExternalTarget: a configured Nomad target.
func NewNomadTarget(allocID, taskName, jobName string) *ExternalTarget {
	id := "nomad:" + allocID + "/" + taskName
	name := jobName + "/" + taskName
	t := NewExternalTarget(id, name, TypeNomad, SourceDiscovered)
	t.Labels["alloc_id"] = allocID
	t.Labels["task"] = taskName
	t.Labels["job"] = jobName
	// Return configured Nomad target with allocation labels.
	return t
}

// WithProbe configures the probe for this target.
//
// Params:
//   - probeType: the probe type ("tcp", "http", "icmp", etc.).
//   - probeTarget: the probe target configuration.
//
// Returns:
//   - *ExternalTarget: the target for method chaining.
func (t *ExternalTarget) WithProbe(probeType string, probeTarget health.Target) *ExternalTarget {
	t.ProbeType = probeType
	t.ProbeTarget = probeTarget
	// Return self to enable fluent API pattern.
	return t
}

// WithTiming configures the timing for this target.
//
// Params:
//   - interval: time between probes.
//   - timeout: probe timeout.
//
// Returns:
//   - *ExternalTarget: the target for method chaining.
func (t *ExternalTarget) WithTiming(interval, timeout time.Duration) *ExternalTarget {
	t.Interval = interval
	t.Timeout = timeout
	// Return self to enable fluent API pattern.
	return t
}

// WithThresholds configures the success/failure thresholds.
//
// Params:
//   - success: consecutive successes to mark healthy.
//   - failure: consecutive failures to mark unhealthy.
//
// Returns:
//   - *ExternalTarget: the target for method chaining.
func (t *ExternalTarget) WithThresholds(success, failure int) *ExternalTarget {
	t.SuccessThreshold = success
	t.FailureThreshold = failure
	// Return self to enable fluent API pattern.
	return t
}

// WithLabel adds a label to the target.
//
// Params:
//   - key: the label key.
//   - value: the label value.
//
// Returns:
//   - *ExternalTarget: the target for method chaining.
func (t *ExternalTarget) WithLabel(key, value string) *ExternalTarget {
	// Initialize labels map if nil to handle edge cases.
	if t.Labels == nil {
		t.Labels = make(map[string]string, defaultMapCapacity)
	}
	t.Labels[key] = value
	// Return self to enable fluent API pattern.
	return t
}

// HasProbe checks if a probe is configured for this target.
//
// Returns:
//   - bool: true if a probe type is set.
func (t *ExternalTarget) HasProbe() bool {
	// Check if probe type is non-empty to determine if probe is configured.
	return t.ProbeType != ""
}

// IsHealthCheckable checks if the target can be health-checked.
// A target is checkable if it has a valid probe configuration.
//
// Returns:
//   - bool: true if the target can be probed.
func (t *ExternalTarget) IsHealthCheckable() bool {
	// Validate that probe is configured and timing values are valid.
	return t.HasProbe() && t.Interval > 0 && t.Timeout > 0
}
