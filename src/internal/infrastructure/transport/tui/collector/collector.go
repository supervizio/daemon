// Package collector provides data collectors for TUI snapshot.
// All collectors use kernel interfaces (procfs, sysfs, syscalls) - no exec.Command.
package collector

import (
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

const (
	// defaultCollectorsCap is the default capacity for collectors slice (typical setup has 5 collectors).
	defaultCollectorsCap int = 8
)

// Collector interface for gathering system information.
type Collector interface {
	// CollectInto populates the snapshot with collected data.
	// Returns error only for critical failures; partial data is acceptable.
	CollectInto(snap *model.Snapshot) error
}

// Collectors aggregates all collectors.
// It provides a convenient way to run multiple collectors and gather data
// into a single snapshot.
type Collectors struct {
	collectors []Collector
}

// NewCollectors creates a new collector aggregator.
//
// Returns:
//   - *Collectors: configured collector aggregator
func NewCollectors() *Collectors {
	// Initialize with pre-allocated slice for efficiency.
	return &Collectors{
		collectors: make([]Collector, 0, defaultCollectorsCap),
	}
}

// Add adds a collector.
//
// Params:
//   - collector: the collector to add
//
// Returns:
//   - *Collectors: self reference for chaining
func (c *Collectors) Add(collector Collector) *Collectors {
	c.collectors = append(c.collectors, collector)
	// Return self for fluent interface.
	return c
}

// CollectAll runs all collectors and populates the snapshot.
// Errors are logged but don't stop collection.
//
// Params:
//   - snap: target snapshot to populate
//
// Returns:
//   - error: always returns nil (errors handled gracefully)
func (c *Collectors) CollectAll(snap *model.Snapshot) error {
	// Execute each collector independently.
	for _, collector := range c.collectors {
		// Ignore errors from individual collectors.
		// They should handle graceful degradation internally.
		_ = collector.CollectInto(snap)
	}
	// Always return nil for graceful degradation.
	return nil
}

// DefaultCollectors returns the standard set of collectors.
//
// Params:
//   - version: daemon version string
//
// Returns:
//   - *Collectors: configured collector set
func DefaultCollectors(version string) *Collectors {
	// Build standard collector chain.
	return NewCollectors().
		Add(NewContextCollector(version)).
		Add(NewLimitsCollector()).
		Add(NewSystemCollector()).
		Add(NewNetworkCollector()).
		Add(NewSandboxCollector())
}

// SetConfigPath sets the config path on the context collector.
//
// Params:
//   - path: configuration file path
func (c *Collectors) SetConfigPath(path string) {
	// Search for context collector in the chain.
	for _, collector := range c.collectors {
		// Check if this is a context collector.
		if ctx, ok := collector.(*ContextCollector); ok {
			ctx.SetConfigPath(path)
			// Stop after finding the first context collector.
			return
		}
	}
}
