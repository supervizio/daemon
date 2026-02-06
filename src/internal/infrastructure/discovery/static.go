// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"context"
	"maps"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// Probe type constants for static target configuration.
const (
	// probeTypeTCP is the TCP probe type constant.
	probeTypeTCP string = "tcp"

	// probeTypeUDP is the UDP probe type constant.
	probeTypeUDP string = "udp"

	// probeTypeHTTP is the HTTP probe type constant.
	probeTypeHTTP string = "http"

	// probeTypeHTTPS is the HTTPS probe type constant.
	probeTypeHTTPS string = "https"

	// probeTypeICMP is the ICMP probe type constant.
	probeTypeICMP string = "icmp"

	// probeTypePing is the ping probe type constant.
	probeTypePing string = "ping"

	// probeTypeExec is the exec probe type constant.
	probeTypeExec string = "exec"
)

// StaticDiscoverer creates targets from static configuration.
// It converts configured targets into ExternalTarget entities with appropriate probes.
type StaticDiscoverer struct {
	// targets are the configured static targets.
	targets []config.TargetConfig
}

// NewStaticDiscoverer creates a discoverer from static target configurations.
//
// Params:
//   - targets: the target configurations.
//
// Returns:
//   - *StaticDiscoverer: a new static discoverer.
func NewStaticDiscoverer(targets []config.TargetConfig) *StaticDiscoverer {
	// Construct discoverer with static target list.
	return &StaticDiscoverer{
		targets: targets,
	}
}

// Discover returns targets from static configuration.
// This method always succeeds as it reads from memory (no I/O).
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []target.ExternalTarget: the configured targets.
//   - error: always nil for static discovery.
func (d *StaticDiscoverer) Discover(ctx context.Context) ([]target.ExternalTarget, error) {
	// Check for context cancellation before processing.
	select {
	case <-ctx.Done():
		// Return early if context is cancelled.
		return nil, ctx.Err()
	default:
	}

	// Convert configurations to targets with pre-allocated capacity.
	result := make([]target.ExternalTarget, 0, len(d.targets))
	// Iterate over all configured targets.
	for i := range d.targets {
		t := d.configToTarget(&d.targets[i])
		result = append(result, t)
	}

	// Return all static targets.
	return result, nil
}

// Type returns the target type for static targets.
//
// Returns:
//   - target.Type: TypeRemote (static targets are usually remote).
func (d *StaticDiscoverer) Type() target.Type {
	// Return remote type constant for static targets.
	return target.TypeRemote
}

// configToTarget converts a TargetConfig to an ExternalTarget.
// It parses the target type, applies labels, and configures the appropriate probe.
//
// Params:
//   - cfg: the target configuration.
//
// Returns:
//   - target.ExternalTarget: the external target.
func (d *StaticDiscoverer) configToTarget(cfg *config.TargetConfig) target.ExternalTarget {
	// Determine target type from configuration string.
	targetType := d.parseTargetType(cfg.Type)

	// Initialize target with static configuration values.
	t := target.ExternalTarget{
		ID:               targetType.String() + ":" + cfg.Name,
		Name:             cfg.Name,
		Type:             targetType,
		Source:           target.SourceStatic,
		Labels:           make(map[string]string, len(cfg.Labels)),
		Interval:         cfg.Interval.Duration(),
		Timeout:          cfg.Timeout.Duration(),
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}

	// Copy all configured labels to target labels.
	maps.Copy(t.Labels, cfg.Labels)

	// Configure probe based on type and address.
	d.configureProbe(&t, cfg) // cfg is already a pointer

	// Return fully configured target.
	return t
}

// parseTargetType converts a type string to target.Type.
// It handles common type aliases and defaults to TypeRemote for unknown types.
//
// Params:
//   - typeStr: the type string.
//
// Returns:
//   - target.Type: the parsed type.
func (d *StaticDiscoverer) parseTargetType(typeStr string) target.Type {
	// Map configuration string to domain type constant.
	switch typeStr {
	// Handle systemd init system type.
	case "systemd":
		// Return systemd type constant.
		return target.TypeSystemd
	// Handle docker container type.
	case "docker":
		// Return docker type constant.
		return target.TypeDocker
	// Handle kubernetes orchestrator type.
	case "kubernetes", "k8s":
		// Return kubernetes type constant.
		return target.TypeKubernetes
	// Handle nomad orchestrator type.
	case "nomad":
		// Return nomad type constant.
		return target.TypeNomad
	// Handle remote or empty type.
	case "remote", "":
		// Return remote type as default.
		return target.TypeRemote
	// Handle unknown or custom types.
	default:
		// Return custom type for unrecognized types.
		return target.TypeCustom
	}
}

// configureProbe configures the probe from TargetConfig.
// It creates the appropriate health check target based on probe type.
//
// Params:
//   - t: the target to configure.
//   - cfg: the target configuration.
func (d *StaticDiscoverer) configureProbe(t *target.ExternalTarget, cfg *config.TargetConfig) {
	probe := cfg.Probe
	// Skip probe configuration when type is empty.
	if probe.Type == "" {
		// Return early without configuring probe.
		return
	}

	t.ProbeType = probe.Type

	// Create probe target based on type and configuration.
	switch probe.Type {
	// Handle TCP connectivity probe.
	case probeTypeTCP:
		// Create TCP probe target.
		t.ProbeTarget = health.NewTCPTarget(cfg.Address)

	// Handle UDP connectivity probe.
	case probeTypeUDP:
		// Create UDP probe target.
		t.ProbeTarget = health.NewUDPTarget(cfg.Address)

	// Handle HTTP/HTTPS endpoint probe.
	case probeTypeHTTP, probeTypeHTTPS:
		// Create HTTP probe with GET method and default status.
		method := "GET"
		t.ProbeTarget = health.NewHTTPTarget(cfg.Address, method, defaultHTTPStatusCode)

	// Handle ICMP/ping reachability probe.
	case probeTypeICMP, probeTypePing:
		// Create ICMP probe target.
		t.ProbeTarget = health.NewICMPTarget(cfg.Address)

	// Handle command execution probe.
	case probeTypeExec:
		// Check if command is specified.
		if probe.Command != "" {
			// Create exec probe with command and args.
			t.ProbeTarget = health.NewExecTarget(probe.Command, probe.Args...)
		}
	}
}
