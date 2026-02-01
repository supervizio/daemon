//go:build freebsd || openbsd || netbsd

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"bytes"
	"context"
	"iter"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// BSDRCDiscoverer discovers BSD rc.d services using platform-specific commands.
// It supports FreeBSD (service), OpenBSD (rcctl), and NetBSD (rc.d directory).
type BSDRCDiscoverer struct {
	// patterns are glob patterns for service selection.
	// Example: "nginx", "postgresql*"
	patterns []string
}

// NewBSDRCDiscoverer creates a new BSD rc.d discoverer.
//
// Params:
//   - patterns: glob patterns for service filtering.
//
// Returns:
//   - *BSDRCDiscoverer: a new BSD rc.d discoverer.
func NewBSDRCDiscoverer(patterns []string) *BSDRCDiscoverer {
	// Construct discoverer with service patterns.
	return &BSDRCDiscoverer{
		patterns: patterns,
	}
}

// Discover finds all BSD rc.d services matching the patterns.
// It uses platform-specific commands to list running services.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []target.ExternalTarget: the discovered services.
//   - error: any error during discovery.
func (d *BSDRCDiscoverer) Discover(ctx context.Context) ([]target.ExternalTarget, error) {
	// Get list of running services from BSD rc.d.
	services, err := d.listServices(ctx)
	// Failed to list BSD rc.d services.
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

// Type returns the target type for BSD rc.d.
//
// Returns:
//   - target.Type: TypeBSDRC.
func (d *BSDRCDiscoverer) Type() target.Type {
	// Return BSD rc.d type constant for this discoverer.
	return target.TypeBSDRC
}

// listServices returns all running BSD rc.d services.
// It uses platform-specific commands based on runtime.GOOS.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []string: list of service names.
//   - error: any error during listing.
func (d *BSDRCDiscoverer) listServices(ctx context.Context) ([]string, error) {
	// Dispatch to platform-specific implementation.
	switch runtime.GOOS {
	case "freebsd":
		// Use FreeBSD's service command.
		return d.listFreeBSDServices(ctx)
	case "openbsd":
		// Use OpenBSD's rcctl command.
		return d.listOpenBSDServices(ctx)
	case "netbsd":
		// Parse NetBSD's rc.d directory.
		return d.listNetBSDServices(ctx)
	default:
		// Return empty list for unsupported platform.
		return nil, nil
	}
}

// listFreeBSDServices lists services on FreeBSD using 'service -l'.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []string: list of service names.
//   - error: any error during listing.
func (d *BSDRCDiscoverer) listFreeBSDServices(ctx context.Context) ([]string, error) {
	// Execute service -l to list all services.
	cmd := exec.CommandContext(ctx, "service", "-l")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Failed to execute service command.
	if err := cmd.Run(); err != nil {
		// Return error when service execution fails.
		return nil, err
	}

	// Parse output line by line using efficient iterator.
	var services []string
	// Iterate over each line in service output.
	for line := range splitLines(stdout.String()) {
		line = strings.TrimSpace(line)
		// Skip empty lines in service output.
		if line == "" {
			continue
		}

		// Add service name (entire line is the service name).
		services = append(services, line)
	}

	// Return parsed service names.
	return services, nil
}

// listOpenBSDServices lists services on OpenBSD using 'rcctl ls started'.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []string: list of service names.
//   - error: any error during listing.
func (d *BSDRCDiscoverer) listOpenBSDServices(ctx context.Context) ([]string, error) {
	// Execute rcctl ls started to list running services.
	cmd := exec.CommandContext(ctx, "rcctl", "ls", "started")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Failed to execute rcctl command.
	if err := cmd.Run(); err != nil {
		// Return error when rcctl execution fails.
		return nil, err
	}

	// Parse output line by line using efficient iterator.
	var services []string
	// Iterate over each line in rcctl output.
	for line := range splitLines(stdout.String()) {
		line = strings.TrimSpace(line)
		// Skip empty lines in rcctl output.
		if line == "" {
			continue
		}

		// Add service name (entire line is the service name).
		services = append(services, line)
	}

	// Return parsed service names.
	return services, nil
}

// listNetBSDServices lists services on NetBSD by parsing /etc/rc.d/.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []string: list of service names.
//   - error: any error during listing.
//
//nolint:unparam // ctx required for interface consistency
func (d *BSDRCDiscoverer) listNetBSDServices(ctx context.Context) ([]string, error) {
	// List all files in /etc/rc.d/ directory.
	matches, err := filepath.Glob("/etc/rc.d/*")
	// Failed to glob rc.d directory.
	if err != nil {
		// Return error when globbing fails.
		return nil, err
	}

	// Extract base names from full paths.
	services := make([]string, 0, len(matches))
	// Iterate over matched files to extract service names.
	for _, match := range matches {
		// Extract service name from file path.
		serviceName := filepath.Base(match)
		services = append(services, serviceName)
	}

	// Return extracted service names.
	return services, nil
}

// splitLines returns an iterator over lines in a string.
// This is more efficient than strings.Split for large outputs as it avoids
// allocating a slice for all lines upfront.
//
// Params:
//   - s: the string to split into lines.
//
// Returns:
//   - iter.Seq[string]: an iterator yielding each line.
func splitLines(s string) iter.Seq[string] {
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
func (d *BSDRCDiscoverer) matchesPatterns(service string) bool {
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

// serviceToTarget converts a BSD rc.d service to an ExternalTarget.
// It configures an exec probe using platform-specific status commands.
//
// Params:
//   - service: the BSD rc.d service name.
//
// Returns:
//   - target.ExternalTarget: the external target.
func (d *BSDRCDiscoverer) serviceToTarget(service string) target.ExternalTarget {
	// Determine platform-specific status command.
	var probeTarget health.Target
	// Select status check command based on platform.
	switch runtime.GOOS {
	case "freebsd":
		// FreeBSD uses 'service <name> status'.
		probeTarget = health.NewExecTarget("service", service, "status")
	case "openbsd":
		// OpenBSD uses 'rcctl check <name>'.
		probeTarget = health.NewExecTarget("rcctl", "check", service)
	case "netbsd":
		// NetBSD uses '/etc/rc.d/<name> status'.
		probeTarget = health.NewExecTarget("/etc/rc.d/"+service, "status")
	default:
		// Fallback to generic service status command.
		probeTarget = health.NewExecTarget("service", service, "status")
	}

	// Initialize target with BSD rc.d-specific configuration.
	t := target.ExternalTarget{
		ID:               "bsd-rc:" + service,
		Name:             service,
		Type:             target.TypeBSDRC,
		Source:           target.SourceDiscovered,
		Labels:           make(map[string]string, 2),
		ProbeType:        "exec",
		ProbeTarget:      probeTarget,
		Interval:         defaultProbeInterval,
		Timeout:          defaultProbeTimeout,
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}

	// Add service name and OS as labels for filtering and querying.
	t.Labels["bsdrc.service"] = service
	t.Labels["bsdrc.os"] = runtime.GOOS

	// Return fully configured target with exec probe.
	return t
}
