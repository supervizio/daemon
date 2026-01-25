// Package collector provides data collectors for TUI snapshot.
package collector

import (
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// LimitsCollector gathers cgroup and resource limit information.
type LimitsCollector struct{}

// NewLimitsCollector creates a limits collector.
func NewLimitsCollector() *LimitsCollector {
	return &LimitsCollector{}
}

// CollectInto populates resource limits.
func (c *LimitsCollector) CollectInto(snap *model.Snapshot) error {
	limits := &snap.Limits

	// Collect cgroup limits (platform-specific).
	collectCgroupLimits(limits)

	// Determine if any limits are set.
	limits.HasLimits = limits.CPUQuota > 0 ||
		limits.MemoryMax > 0 ||
		limits.PIDsMax > 0 ||
		limits.CPUSet != ""

	return nil
}
