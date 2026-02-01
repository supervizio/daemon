//go:build linux

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"bytes"
	"context"
	"iter"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// OpenRCDiscoverer discovers OpenRC services using rc-status.
// It queries running services and creates health checks using rc-service status.
type OpenRCDiscoverer struct {
	// patterns are glob patterns for service selection.
	// Example: "nginx", "postgresql*"
	patterns []string
}

// NewOpenRCDiscoverer creates a new OpenRC discoverer.
//
// Params:
//   - patterns: glob patterns for service filtering.
//
// Returns:
//   - *OpenRCDiscoverer: a new OpenRC discoverer.
func NewOpenRCDiscoverer(patterns []string) *OpenRCDiscoverer {
	// Construct discoverer with service patterns.
	return &OpenRCDiscoverer{
		patterns: patterns,
	}
}

// Discover finds all OpenRC services matching the patterns.
// It executes rc-status to list running services and filters by glob patterns.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []target.ExternalTarget: the discovered services.
//   - error: any error during discovery.
func (d *OpenRCDiscoverer) Discover(ctx context.Context) ([]target.ExternalTarget, error) {
	// Get list of running services from OpenRC.
	services, err := d.listServices(ctx)
	// Failed to list OpenRC services.
	if err != nil {
		// Return error when service listing fails.
		return nil, err
	}

	// Filter services by glob patterns and convert to targets.
	var targets []target.ExternalTarget
	// Iterate over all discovered services.
	for _, service := range services {
		// Skip service that doesn't match any pattern.
		if !d.matchesPatterns(service) {
			continue
		}

		t := d.serviceToTarget(service)
		targets = append(targets, t)
	}

	// Return all discovered targets.
	return targets, nil
}

// Type returns the target type for OpenRC.
//
// Returns:
//   - target.Type: TypeOpenRC.
func (d *OpenRCDiscoverer) Type() target.Type {
	// Return OpenRC type constant for this discoverer.
	return target.TypeOpenRC
}

// listServices returns all running OpenRC services.
// It executes rc-status and parses the output to extract service names.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []string: list of service names.
//   - error: any error during listing.
func (d *OpenRCDiscoverer) listServices(ctx context.Context) ([]string, error) {
	// Execute rc-status to list running services.
	// Use -s for simple output format: "service_name [started|stopped|...]"
	cmd := exec.CommandContext(ctx, "rc-status", "-s")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Failed to execute rc-status command.
	if err := cmd.Run(); err != nil {
		// Return error when rc-status execution fails.
		return nil, err
	}

	// Parse output line by line using efficient iterator.
	var services []string
	// Iterate over each line in rc-status output.
	for line := range splitLinesOpenRC(stdout.String()) {
		line = strings.TrimSpace(line)
		// Skip empty lines in rc-status output.
		if line == "" {
			continue
		}

		// Extract service name from format: "service_name [started]"
		// Service name is everything before the first space.
		spaceIdx := strings.Index(line, " ")
		if spaceIdx == -1 {
			// Line doesn't contain status bracket - skip malformed output.
			continue
		}

		// Extract service name (before the space).
		serviceName := line[:spaceIdx]
		// Append service name if non-empty.
		if serviceName != "" {
			services = append(services, serviceName)
		}
	}

	// Return parsed service names.
	return services, nil
}

// splitLinesOpenRC returns an iterator over lines in a string.
// This is more efficient than strings.Split for large outputs as it avoids
// allocating a slice for all lines upfront.
//
// Params:
//   - s: the string to split into lines.
//
// Returns:
//   - iter.Seq[string]: an iterator yielding each line.
func splitLinesOpenRC(s string) iter.Seq[string] {
	// Return iterator function for line-by-line traversal.
	return func(yield func(string) bool) {
		start := 0
		// Iterate over each character to find newlines.
		for i := range len(s) {
			// Yield line and check if iteration should continue.
			if s[i] == '\n' {
				// Yield current line to consumer.
				if !yield(s[start:i]) {
					// Consumer stopped iteration early.
					return
				}
				start = i + 1
			}
		}
		// Yield last line if it's non-empty (no trailing newline).
		if start < len(s) {
			// Ignore return value as iteration completes regardless.
			_ = yield(s[start:])
		}
	}
}

// matchesPatterns checks if a service matches any of the patterns.
// It uses filepath.Match for glob pattern matching.
//
// Params:
//   - service: the service name to check.
//
// Returns:
//   - bool: true if service matches any pattern.
func (d *OpenRCDiscoverer) matchesPatterns(service string) bool {
	// Accept all services when no patterns are configured.
	if len(d.patterns) == 0 {
		// Return true when no filtering is needed.
		return true
	}

	// Check each glob pattern for a match.
	for _, pattern := range d.patterns {
		matched, err := filepath.Match(pattern, service)
		// Pattern matched - accept this service.
		if err == nil && matched {
			// Return true when pattern matches.
			return true
		}
	}

	// No patterns matched - reject this service.
	return false
}

// serviceToTarget converts an OpenRC service to an ExternalTarget.
// It configures an exec probe using rc-service status.
//
// Params:
//   - service: the OpenRC service name.
//
// Returns:
//   - target.ExternalTarget: the external target.
func (d *OpenRCDiscoverer) serviceToTarget(service string) target.ExternalTarget {
	// Initialize target with OpenRC-specific configuration.
	t := target.ExternalTarget{
		ID:               "openrc:" + service,
		Name:             service,
		Type:             target.TypeOpenRC,
		Source:           target.SourceDiscovered,
		Labels:           make(map[string]string, 1),
		ProbeType:        "exec",
		ProbeTarget:      health.NewExecTarget("rc-service", service, "status"),
		Interval:         defaultProbeInterval,
		Timeout:          defaultProbeTimeout,
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}

	// Add service name as label for filtering and querying.
	t.Labels["openrc.service"] = service

	// Return fully configured target with exec probe.
	return t
}
