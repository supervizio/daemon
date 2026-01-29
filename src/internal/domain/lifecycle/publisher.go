// Package lifecycle provides domain types for daemon lifecycle management.
package lifecycle

// Publisher defines the interface for publishing events.
type Publisher interface {
	// Publish publishes an event to all subscribers.
	Publish(event Event)
	// Subscribe returns a channel that receives events.
	Subscribe() <-chan Event
	// Unsubscribe removes a subscription.
	Unsubscribe(ch <-chan Event)
}

// Handler is a function that handles events.
type Handler func(Event)

// Filter is a function that filters events.
// Returns true if the event should be passed through.
type Filter func(Event) bool

// FilterByType returns a filter that only passes events of the given types.
//
// Params:
//   - types: event types to include in the filter
//
// Returns:
//   - Filter: filter function that passes only matching event types
func FilterByType(types ...Type) Filter {
	typeSet := make(map[Type]struct{}, len(types))
	// build set of allowed types
	for _, t := range types {
		typeSet[t] = struct{}{}
	}
	// return filter closure
	return func(e Event) bool {
		// check if event type is in the set
		_, ok := typeSet[e.Type]
		// return membership result
		return ok
	}
}

// FilterByCategory returns a filter that only passes events of the given category.
//
// Params:
//   - category: the event category to filter by (process, mesh, kubernetes, system, daemon)
//
// Returns:
//   - Filter: filter function that passes only events matching the category
func FilterByCategory(category string) Filter {
	// return filter closure
	return func(e Event) bool {
		// return category match result
		return e.Type.Category() == category
	}
}

// FilterByServiceName returns a filter that only passes events for the given service.
//
// Params:
//   - serviceName: the service name to filter by
//
// Returns:
//   - Filter: filter function that passes only events for the specified service
func FilterByServiceName(serviceName string) Filter {
	// return filter closure
	return func(e Event) bool {
		// return service name match result
		return e.ServiceName == serviceName
	}
}
