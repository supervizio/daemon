// Package events provides event bus implementation for lifecycle.Publisher.
package events

import (
	"sync"

	"github.com/kodflow/daemon/internal/domain/lifecycle"
)

const (
	// defaultBufferSize is the default channel buffer size for subscribers.
	defaultBufferSize int = 64
)

// Bus implements lifecycle.Publisher with a simple pub/sub pattern.
//
// It maintains a set of subscribers and broadcasts events to all of them.
// Events are sent non-blocking; slow subscribers may miss events.
type Bus struct {
	mu          sync.RWMutex
	subscribers map[<-chan lifecycle.Event]chan lifecycle.Event
	bufferSize  int
	closed      bool
}

// BusOption configures Bus behavior.
type BusOption func(*Bus)

// WithBufferSize sets the subscriber channel buffer size.
//
// Params:
//   - size: buffer size for subscriber channels (default: 64)
//
// Returns:
//   - BusOption: configuration option
func WithBufferSize(size int) BusOption {
	// Return closure that applies buffer size configuration.
	return func(b *Bus) {
		// Only apply if size is positive to maintain default behavior.
		if size > 0 {
			b.bufferSize = size
		}
	}
}

// NewBus creates a new event bus.
//
// Params:
//   - opts: optional configuration options
//
// Returns:
//   - *Bus: new event bus instance
func NewBus(opts ...BusOption) *Bus {
	b := &Bus{
		subscribers: make(map[<-chan lifecycle.Event]chan lifecycle.Event, 0),
		bufferSize:  defaultBufferSize,
	}
	// Apply all provided options to configure the bus.
	for _, opt := range opts {
		opt(b)
	}

	// Return the fully configured bus instance.
	return b
}

// Publish broadcasts an event to all subscribers (non-blocking; drops if buffer full).
//
// Params:
//   - event: the event to publish
func (b *Bus) Publish(event lifecycle.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Skip publishing if bus is already closed.
	if b.closed {
		// Silently return when closed to avoid panic.
		return
	}

	// Send event to all active subscribers.
	for _, ch := range b.subscribers {
		select {
		case ch <- event:
			// Event sent successfully to this subscriber.
		default:
			// Subscriber buffer full; drop event to avoid blocking.
		}
	}
}

// Subscribe creates a new subscription channel that receives events until Unsubscribe or Close.
//
// Returns:
//   - <-chan lifecycle.Event: channel for receiving events
func (b *Bus) Subscribe() <-chan lifecycle.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Return a closed channel if the bus is already closed.
	if b.closed {
		ch := make(chan lifecycle.Event)
		close(ch)

		// Return closed channel to signal bus is unavailable.
		return ch
	}

	// Create new subscriber channel with configured buffer size.
	ch := make(chan lifecycle.Event, b.bufferSize)
	b.subscribers[ch] = ch

	// Return the new subscription channel.
	return ch
}

// Unsubscribe removes a subscription (idempotent; safe with unknown channels).
//
// Params:
//   - ch: the subscription channel to remove
func (b *Bus) Unsubscribe(ch <-chan lifecycle.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Close and remove the subscription if it exists.
	if writeCh, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(writeCh)
	}
}

// Close shuts down the event bus and closes all subscriber channels (Publish becomes no-op).
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Prevent multiple close operations.
	if b.closed {
		// Already closed, nothing to do.
		return
	}

	// Mark bus as closed and close all subscriber channels.
	b.closed = true

	// Iterate over all subscribers to close and remove them.
	for readCh, writeCh := range b.subscribers {
		delete(b.subscribers, readCh)
		close(writeCh)
	}
}

// SubscriberCount returns the current number of subscribers.
//
// Returns:
//   - int: number of active subscribers
func (b *Bus) SubscriberCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return the count of active subscribers.
	return len(b.subscribers)
}

// compile-time interface check
var _ lifecycle.Publisher = (*Bus)(nil)
