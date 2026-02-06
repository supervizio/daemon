// Package target provides domain entities for external monitoring targets.
// External targets are services, containers, or hosts that supervizio
// monitors but does not manage (no lifecycle control).
package target

import "context"

// EventType represents the type of target event.
type EventType string

// Event type constants define the possible target events.
const (
	// EventAdded indicates a new target was discovered.
	EventAdded EventType = "added"

	// EventRemoved indicates a target was removed.
	EventRemoved EventType = "removed"

	// EventUpdated indicates a target's configuration changed.
	EventUpdated EventType = "updated"

	// EventHealthChanged indicates a target's health state changed.
	EventHealthChanged EventType = "health_changed"
)

// Event represents a target lifecycle event.
// It carries information about target changes like additions,
// removals, updates, and health state transitions.
type Event struct {
	// Type is the kind of event.
	Type EventType

	// Target is the affected target.
	// For EventRemoved, only ID will be populated.
	Target ExternalTarget

	// PreviousState is the previous health state (for EventHealthChanged).
	PreviousState State

	// NewState is the new health state (for EventHealthChanged).
	NewState State
}

// Watcher is a port interface for watching target changes in real-time.
// Infrastructure adapters implement this interface to provide
// live updates from platforms like Docker events or Kubernetes watches.
type Watcher interface {
	// Watch starts watching for target changes.
	// Returns a channel that receives events until context is cancelled.
	//
	// Params:
	//   - ctx: context for cancellation.
	//
	// Returns:
	//   - <-chan Event: channel of target events.
	//   - error: any error starting the watch.
	Watch(ctx context.Context) (<-chan Event, error)

	// Type returns the target type this watcher handles.
	//
	// Returns:
	//   - Type: the target type.
	Type() Type
}

// NewAddedEvent creates an event for a newly discovered target.
//
// Params:
//   - target: the new target.
//
// Returns:
//   - Event: an added event.
func NewAddedEvent(target *ExternalTarget) Event {
	// Create event indicating a new target was discovered.
	return Event{
		Type:   EventAdded,
		Target: *target,
	}
}

// NewRemovedEvent creates an event for a removed target.
//
// Params:
//   - targetID: the ID of the removed target.
//
// Returns:
//   - Event: a removed event.
func NewRemovedEvent(targetID string) Event {
	// Create event with minimal target data (only ID needed for removal).
	return Event{
		Type: EventRemoved,
		Target: ExternalTarget{
			ID: targetID,
		},
	}
}

// NewUpdatedEvent creates an event for an updated target.
//
// Params:
//   - target: the updated target.
//
// Returns:
//   - Event: an updated event.
func NewUpdatedEvent(target *ExternalTarget) Event {
	// Create event indicating target configuration changed.
	return Event{
		Type:   EventUpdated,
		Target: *target,
	}
}

// NewHealthChangedEvent creates an event for a health state change.
//
// Params:
//   - target: the target whose health changed.
//   - previousState: the previous health state.
//   - newState: the new health state.
//
// Returns:
//   - Event: a health changed event.
func NewHealthChangedEvent(target *ExternalTarget, previousState, newState State) Event {
	// Create event with state transition information for health monitoring.
	return Event{
		Type:          EventHealthChanged,
		Target:        *target,
		PreviousState: previousState,
		NewState:      newState,
	}
}
