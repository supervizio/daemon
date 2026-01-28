// Package collector provides data collectors for TUI snapshot.
// All collectors use kernel interfaces (procfs, sysfs, syscalls) - no exec.Command.
package collector

import (
	"net"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

const (
	// defaultCollectorsCap is the default capacity for collectors slice (typical setup has 5 collectors).
	defaultCollectorsCap int = 8
)

// Gatherer interface for gathering system information.
type Gatherer interface {
	// Gather populates the snapshot with collected data.
	// Returns error only for critical failures; partial data is acceptable.
	Gather(snap *model.Snapshot) error
}

// Addrser defines the minimal interface for network interface address retrieval (KTN-API-MINIF).
type Addrser interface {
	Addrs() ([]net.Addr, error)
}

// Collectors aggregates all collectors.
// It provides a convenient way to run multiple collectors and gather data
// into a single snapshot.
type Collectors struct {
	collectors []Gatherer
}

// NewCollectors creates a new collector aggregator.
//
// Returns:
//   - *Collectors: configured collector aggregator
func NewCollectors() *Collectors {
	// Initialize with pre-allocated slice for efficiency.
	return &Collectors{
		collectors: make([]Gatherer, 0, defaultCollectorsCap),
	}
}

// Add adds a collector.
//
// Params:
//   - collector: the collector to add
//
// Returns:
//   - *Collectors: self reference for chaining
func (c *Collectors) Add(gatherer Gatherer) *Collectors {
	c.collectors = append(c.collectors, gatherer)
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
	// Execute each gatherer independently.
	for _, gatherer := range c.collectors {
		// Ignore errors from individual gatherers.
		// They should handle graceful degradation internally.
		_ = gatherer.Gather(snap)
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
