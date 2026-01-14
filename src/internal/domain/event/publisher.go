// Package event provides domain types for event handling.
package event

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
func FilterByType(types ...Type) Filter {
	typeSet := make(map[Type]struct{}, len(types))
	for _, t := range types {
		typeSet[t] = struct{}{}
	}
	return func(e Event) bool {
		_, ok := typeSet[e.Type]
		return ok
	}
}

// FilterByCategory returns a filter that only passes events of the given category.
func FilterByCategory(category string) Filter {
	return func(e Event) bool {
		return e.Type.Category() == category
	}
}

// FilterByServiceName returns a filter that only passes events for the given service.
func FilterByServiceName(serviceName string) Filter {
	return func(e Event) bool {
		return e.ServiceName == serviceName
	}
}
