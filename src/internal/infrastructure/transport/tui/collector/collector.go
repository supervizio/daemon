// Package collector provides data collectors for TUI snapshot.
// All collectors use kernel interfaces (procfs, sysfs, syscalls) - no exec.Command.
package collector

import (
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// Collector interface for gathering system information.
type Collector interface {
	// CollectInto populates the snapshot with collected data.
	// Returns error only for critical failures; partial data is acceptable.
	CollectInto(snap *model.Snapshot) error
}

// Collectors aggregates all collectors.
type Collectors struct {
	collectors []Collector
}

// Default capacity for collectors slice (typical setup has 5 collectors).
const defaultCollectorsCap = 8

// NewCollectors creates a new collector aggregator.
func NewCollectors() *Collectors {
	return &Collectors{
		collectors: make([]Collector, 0, defaultCollectorsCap),
	}
}

// Add adds a collector.
func (c *Collectors) Add(collector Collector) *Collectors {
	c.collectors = append(c.collectors, collector)
	return c
}

// CollectAll runs all collectors and populates the snapshot.
// Errors are logged but don't stop collection.
func (c *Collectors) CollectAll(snap *model.Snapshot) error {
	for _, collector := range c.collectors {
		// Ignore errors from individual collectors.
		// They should handle graceful degradation internally.
		_ = collector.CollectInto(snap)
	}
	return nil
}

// DefaultCollectors returns the standard set of collectors.
func DefaultCollectors(version string) *Collectors {
	return NewCollectors().
		Add(NewContextCollector(version)).
		Add(NewLimitsCollector()).
		Add(NewSystemCollector()).
		Add(NewNetworkCollector()).
		Add(NewSandboxCollector())
}

// SetConfigPath sets the config path on the context collector.
func (c *Collectors) SetConfigPath(path string) {
	for _, collector := range c.collectors {
		if ctx, ok := collector.(*ContextCollector); ok {
			ctx.SetConfigPath(path)
			return
		}
	}
}
