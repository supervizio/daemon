// Package monitoring provides the application service for external target monitoring.
package monitoring

import (
	"maps"
	"slices"
	"sync"

	"github.com/kodflow/daemon/internal/domain/target"
)

// defaultMapCapacity is the initial capacity for target and status maps.
const defaultMapCapacity int = 16

// defaultTypeMapCapacity is the initial capacity for type maps (8 target types).
const defaultTypeMapCapacity int = 8

// defaultStateMapCapacity is the initial capacity for state maps (4 health states).
const defaultStateMapCapacity int = 4

// Registry provides thread-safe storage for external targets and their status.
// It maintains both targets and their corresponding health statuses.
type Registry struct {
	// mu protects concurrent access to the registry.
	mu sync.RWMutex

	// targets maps target ID to target.
	targets map[string]*target.ExternalTarget

	// statuses maps target ID to status.
	statuses map[string]*target.Status
}

// NewRegistry creates a new empty registry.
//
// Returns:
//   - *Registry: a new registry instance.
func NewRegistry() *Registry {
	// construct registry with empty maps
	return &Registry{
		targets:  make(map[string]*target.ExternalTarget, defaultMapCapacity),
		statuses: make(map[string]*target.Status, defaultMapCapacity),
	}
}

// Add adds a target to the registry.
// If the target already exists, returns ErrTargetExists.
//
// Params:
//   - t: the target to add.
//
// Returns:
//   - error: ErrTargetExists if target already exists.
func (r *Registry) Add(t *target.ExternalTarget) error {
	// Lock for thread-safe update.
	r.mu.Lock()
	defer r.mu.Unlock()

	// check if target already exists
	if _, exists := r.targets[t.ID]; exists {
		// return error for duplicate target
		return ErrTargetExists
	}

	// Store target and create initial status.
	r.targets[t.ID] = t
	r.statuses[t.ID] = target.NewStatus(t)

	// Return nil on successful target registration.
	return nil
}

// AddOrUpdate adds or updates a target in the registry.
//
// Params:
//   - t: the target to add or update.
func (r *Registry) AddOrUpdate(t *target.ExternalTarget) {
	// Lock for thread-safe update.
	r.mu.Lock()
	defer r.mu.Unlock()

	// Store or update target.
	r.targets[t.ID] = t

	// check if status exists
	if _, exists := r.statuses[t.ID]; !exists {
		// Create initial status for new target.
		r.statuses[t.ID] = target.NewStatus(t)
	}
}

// Remove removes a target from the registry.
//
// Params:
//   - id: the target ID to remove.
//
// Returns:
//   - error: ErrTargetNotFound if target not found.
func (r *Registry) Remove(id string) error {
	// Lock for thread-safe update.
	r.mu.Lock()
	defer r.mu.Unlock()

	// check if target exists
	if _, exists := r.targets[id]; !exists {
		// return error for missing target
		return ErrTargetNotFound
	}

	// Remove target and status.
	delete(r.targets, id)
	delete(r.statuses, id)

	// Return nil on successful target removal.
	return nil
}

// Get returns a target by ID.
//
// Params:
//   - id: the target ID.
//
// Returns:
//   - *target.ExternalTarget: the target or nil if not found.
func (r *Registry) Get(id string) *target.ExternalTarget {
	// Lock for thread-safe read.
	r.mu.RLock()
	defer r.mu.RUnlock()

	// return target or nil
	return r.targets[id]
}

// GetStatus returns the status for a target.
//
// Params:
//   - id: the target ID.
//
// Returns:
//   - *target.Status: the status or nil if not found.
func (r *Registry) GetStatus(id string) *target.Status {
	// Lock for thread-safe read.
	r.mu.RLock()
	defer r.mu.RUnlock()

	// return status or nil
	return r.statuses[id]
}

// All returns all targets in the registry.
//
// Returns:
//   - []*target.ExternalTarget: slice of all targets.
func (r *Registry) All() []*target.ExternalTarget {
	// Lock for thread-safe read.
	r.mu.RLock()
	defer r.mu.RUnlock()

	// return all targets using maps.Values
	return slices.Collect(maps.Values(r.targets))
}

// AllStatuses returns all target statuses.
//
// Returns:
//   - []*target.Status: slice of all statuses.
func (r *Registry) AllStatuses() []*target.Status {
	// Lock for thread-safe read.
	r.mu.RLock()
	defer r.mu.RUnlock()

	// return all statuses using maps.Values
	return slices.Collect(maps.Values(r.statuses))
}

// Count returns the number of targets in the registry.
//
// Returns:
//   - int: the number of targets.
func (r *Registry) Count() int {
	// Lock for thread-safe read.
	r.mu.RLock()
	defer r.mu.RUnlock()

	// return target count
	return len(r.targets)
}

// ByType returns all targets of a specific type.
//
// Params:
//   - targetType: the type to filter by.
//
// Returns:
//   - []*target.ExternalTarget: slice of matching targets.
func (r *Registry) ByType(targetType target.Type) []*target.ExternalTarget {
	// Lock for thread-safe read.
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create result slice with estimated capacity.
	var result []*target.ExternalTarget
	// iterate over all targets
	for _, t := range r.targets {
		// check if type matches
		if t.Type == targetType {
			result = append(result, t)
		}
	}

	// return matching targets
	return result
}

// ByState returns all targets in a specific state.
//
// Params:
//   - state: the health state to filter by.
//
// Returns:
//   - []*target.ExternalTarget: slice of matching targets.
func (r *Registry) ByState(state target.State) []*target.ExternalTarget {
	// Lock for thread-safe read.
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create result slice.
	var result []*target.ExternalTarget
	// iterate over all statuses
	for id, s := range r.statuses {
		// check if state matches
		if s.State == state {
			// check if target exists
			if t, exists := r.targets[id]; exists {
				result = append(result, t)
			}
		}
	}

	// return matching targets
	return result
}

// UpdateStatus updates the status of a target with a callback.
// This ensures atomic status updates.
//
// Params:
//   - id: the target ID.
//   - fn: the update function.
//
// Returns:
//   - error: ErrTargetNotFound if target not found.
func (r *Registry) UpdateStatus(id string, fn func(*target.Status)) error {
	// Lock for thread-safe update.
	r.mu.Lock()
	defer r.mu.Unlock()

	status, exists := r.statuses[id]
	// check if status exists
	if !exists {
		// return error for missing target
		return ErrTargetNotFound
	}

	// Apply update function.
	fn(status)

	// Return nil after status update applied.
	return nil
}

// HealthSummary returns a summary of target health states.
//
// Returns:
//   - HealthSummary: counts of targets by state.
func (r *Registry) HealthSummary() HealthSummary {
	// Lock for thread-safe read.
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Construct summary using constructor.
	summary := NewHealthSummary()
	summary.Total = len(r.targets)

	// Count by type.
	for _, t := range r.targets {
		summary.ByType[t.Type]++
	}

	// Count by state.
	for _, s := range r.statuses {
		summary.ByState[s.State]++
	}

	// return computed summary
	return summary
}
