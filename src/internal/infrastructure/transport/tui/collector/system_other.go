//go:build !linux

// Package collector provides data collection for the TUI.
package collector

import (
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// SystemCollector collects system-wide metrics.
// On non-Linux platforms, most metrics are not available.
type SystemCollector struct{}

// NewSystemCollector creates a new system collector.
func NewSystemCollector() *SystemCollector {
	return &SystemCollector{}
}

// CollectInto gathers system metrics.
// On non-Linux platforms, returns zeros (graceful degradation).
func (c *SystemCollector) CollectInto(snap *model.Snapshot) error {
	// CPU, memory, load average, and disk metrics require platform-specific
	// implementations (sysctl on BSD/Darwin, etc.).
	// For now, return zeros - the TUI will display "-" for missing data.
	snap.System.DiskPath = "/"
	return nil
}
