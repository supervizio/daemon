// Package collector provides data collectors for TUI snapshot.
package collector

import (
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// LimitsCollector gathers cgroup and resource limit information.
// It collects CPU quota, memory limits, PID limits, and CPU sets
// from cgroup v1 or v2 hierarchies.
type LimitsCollector struct{}

// NewLimitsCollector creates a limits collector.
//
// Returns:
//   - *LimitsCollector: configured limits collector
func NewLimitsCollector() *LimitsCollector {
	// Return empty struct (no state needed).
	return &LimitsCollector{}
}

// Gather populates resource limits.
//
// Params:
//   - snap: target snapshot to populate
//
// Returns:
//   - error: always returns nil
func (c *LimitsCollector) Gather(snap *model.Snapshot) error {
	limits := &snap.Limits

	// Collect cgroup limits (platform-specific).
	collectCgroupLimits(limits)

	// Determine if any limits are set.
	limits.HasLimits = limits.CPUQuota > 0 ||
		limits.MemoryMax > 0 ||
		limits.PIDsMax > 0 ||
		limits.CPUSet != ""

	// Always return nil for graceful operation.
	return nil
}
