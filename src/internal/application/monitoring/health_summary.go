// Package monitoring provides the application service for external target monitoring.
package monitoring

import "github.com/kodflow/daemon/internal/domain/target"

// HealthSummary contains counts of targets by state and type.
// It provides an aggregate view of the registry's health status.
type HealthSummary struct {
	// Total is the total number of targets.
	Total int

	// ByType counts targets by type.
	ByType map[target.Type]int

	// ByState counts targets by health state.
	ByState map[target.State]int
}

// NewHealthSummary creates a new HealthSummary with initialized maps.
//
// Returns:
//   - HealthSummary: a new health summary instance.
func NewHealthSummary() HealthSummary {
	// construct summary with empty maps
	return HealthSummary{
		Total:   0,
		ByType:  make(map[target.Type]int, defaultTypeMapCapacity),
		ByState: make(map[target.State]int, defaultStateMapCapacity),
	}
}

// HealthyCount returns the number of healthy targets.
//
// Returns:
//   - int: count of healthy targets.
func (h HealthSummary) HealthyCount() int {
	// return healthy count from state map
	return h.ByState[target.StateHealthy]
}

// UnhealthyCount returns the number of unhealthy targets.
//
// Returns:
//   - int: count of unhealthy targets.
func (h HealthSummary) UnhealthyCount() int {
	// return unhealthy count from state map
	return h.ByState[target.StateUnhealthy]
}

// UnknownCount returns the number of targets in unknown state.
//
// Returns:
//   - int: count of unknown targets.
func (h HealthSummary) UnknownCount() int {
	// return unknown count from state map
	return h.ByState[target.StateUnknown]
}
