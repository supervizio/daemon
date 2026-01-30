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
	return func(b *Bus) {
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
		subscribers: make(map[<-chan lifecycle.Event]chan lifecycle.Event),
		bufferSize:  defaultBufferSize,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// Publish broadcasts an event to all subscribers.
//
// Events are sent non-blocking. If a subscriber's buffer is full,
// the event is dropped for that subscriber.
//
// Params:
//   - event: the event to publish
func (b *Bus) Publish(event lifecycle.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.closed {
		return
	}

	for _, ch := range b.subscribers {
		select {
		case ch <- event:
			// sent successfully
		default:
			// subscriber buffer full, drop event
		}
	}
}

// Subscribe creates a new subscription channel.
//
// The returned channel receives published events until Unsubscribe is called
// or the bus is closed.
//
// Returns:
//   - <-chan lifecycle.Event: channel for receiving events
func (b *Bus) Subscribe() <-chan lifecycle.Event {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		// return closed channel if bus is closed
		ch := make(chan lifecycle.Event)
		close(ch)
		return ch
	}

	ch := make(chan lifecycle.Event, b.bufferSize)
	b.subscribers[ch] = ch
	return ch
}

// Unsubscribe removes a subscription.
//
// The channel will be closed and removed from the subscriber list.
// Safe to call multiple times or with unknown channels.
//
// Params:
//   - ch: the subscription channel to remove
func (b *Bus) Unsubscribe(ch <-chan lifecycle.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if writeCh, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(writeCh)
	}
}

// Close shuts down the event bus and closes all subscriber channels.
//
// After Close, Publish becomes a no-op and Subscribe returns closed channels.
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return
	}

	b.closed = true
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
	return len(b.subscribers)
}

// compile-time interface check
var _ lifecycle.Publisher = (*Bus)(nil)
