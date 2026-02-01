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

// SystemdDiscoverer discovers systemd services using systemctl.
// It queries running services and creates health checks using systemctl is-active.
type SystemdDiscoverer struct {
	// patterns are glob patterns for service selection.
	// Example: "nginx.service", "postgresql*.service"
	patterns []string
}

// NewSystemdDiscoverer creates a new systemd discoverer.
//
// Params:
//   - patterns: glob patterns for service filtering.
//
// Returns:
//   - *SystemdDiscoverer: a new systemd discoverer.
func NewSystemdDiscoverer(patterns []string) *SystemdDiscoverer {
	// Construct discoverer with service patterns.
	return &SystemdDiscoverer{
		patterns: patterns,
	}
}

// Discover finds all systemd services matching the patterns.
// It executes systemctl to list running services and filters by glob patterns.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []target.ExternalTarget: the discovered services.
//   - error: any error during discovery.
func (d *SystemdDiscoverer) Discover(ctx context.Context) ([]target.ExternalTarget, error) {
	// Get list of running services from systemd.
	units, err := d.listUnits(ctx)
	// Failed to list systemd units.
	if err != nil {
		// Return error when unit listing fails.
		return nil, err
	}

	// Filter units by glob patterns and convert to targets.
	var targets []target.ExternalTarget
	// Iterate over all discovered units.
	for _, unit := range units {
		// Skip unit that doesn't match any pattern.
		if !d.matchesPatterns(unit) {
			continue
		}

		t := d.unitToTarget(unit)
		targets = append(targets, t)
	}

	// Return all discovered targets.
	return targets, nil
}

// Type returns the target type for systemd.
//
// Returns:
//   - target.Type: TypeSystemd.
func (d *SystemdDiscoverer) Type() target.Type {
	// Return systemd type constant for this discoverer.
	return target.TypeSystemd
}

// listUnits returns all running systemd service units.
// It executes systemctl and parses the output to extract unit names.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []string: list of unit names.
//   - error: any error during listing.
func (d *SystemdDiscoverer) listUnits(ctx context.Context) ([]string, error) {
	// Execute systemctl to list running service units.
	cmd := exec.CommandContext(ctx, "systemctl", "list-units", "--type=service", "--state=running", "--no-pager", "--no-legend", "--plain")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// Failed to execute systemctl command.
	if err := cmd.Run(); err != nil {
		// Return error when systemctl execution fails.
		return nil, err
	}

	// Parse output line by line using efficient iterator.
	var units []string
	// Iterate over each line in systemctl output.
	for line := range splitLines(stdout.String()) {
		line = strings.TrimSpace(line)
		// Skip empty lines in systemctl output.
		if line == "" {
			continue
		}

		// Extract unit name (first whitespace-separated field).
		fields := strings.Fields(line)
		// Append unit name if fields are present.
		if len(fields) > 0 {
			units = append(units, fields[0])
		}
	}

	// Return parsed unit names.
	return units, nil
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

// matchesPatterns checks if a unit matches any of the patterns.
// It uses filepath.Match for glob pattern matching.
//
// Params:
//   - unit: the unit name to check.
//
// Returns:
//   - bool: true if unit matches any pattern.
func (d *SystemdDiscoverer) matchesPatterns(unit string) bool {
	// Accept all units when no patterns are configured.
	if len(d.patterns) == 0 {
		// Return true when no filtering is needed.
		return true
	}

	// Check each glob pattern for a match.
	for _, pattern := range d.patterns {
		matched, err := filepath.Match(pattern, unit)
		// Pattern matched - accept this unit.
		if err == nil && matched {
			// Return true when pattern matches.
			return true
		}
	}

	// No patterns matched - reject this unit.
	return false
}

// unitToTarget converts a systemd unit to an ExternalTarget.
// It configures an exec probe using systemctl is-active.
//
// Params:
//   - unit: the systemd unit name.
//
// Returns:
//   - target.ExternalTarget: the external target.
func (d *SystemdDiscoverer) unitToTarget(unit string) target.ExternalTarget {
	// Initialize target with systemd-specific configuration.
	t := target.ExternalTarget{
		ID:               "systemd:" + unit,
		Name:             unit,
		Type:             target.TypeSystemd,
		Source:           target.SourceDiscovered,
		Labels:           make(map[string]string, 1),
		ProbeType:        "exec",
		ProbeTarget:      health.NewExecTarget("systemctl", "is-active", "--quiet", unit),
		Interval:         defaultProbeInterval,
		Timeout:          defaultProbeTimeout,
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}

	// Add unit name as label for filtering and querying.
	t.Labels["systemd.unit"] = unit

	// Return fully configured target with exec probe.
	return t
}
