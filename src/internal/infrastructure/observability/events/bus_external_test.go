package events_test

import (
	"sync"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/lifecycle"
	"github.com/kodflow/daemon/internal/infrastructure/observability/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBus_Subscribe_ReturnsChannel verifies Subscribe returns a valid channel.
func TestBus_Subscribe_ReturnsChannel(t *testing.T) {
	bus := events.NewBus()
	defer bus.Close()

	ch := bus.Subscribe()
	require.NotNil(t, ch)
}

// TestBus_Publish_DeliversToSubscribers verifies events are delivered to all subscribers.
func TestBus_Publish_DeliversToSubscribers(t *testing.T) {
	bus := events.NewBus()
	defer bus.Close()

	sub1 := bus.Subscribe()
	sub2 := bus.Subscribe()

	event := lifecycle.NewEvent(lifecycle.TypeProcessStarted, "test service started")

	bus.Publish(event)

	select {
	case received := <-sub1:
		assert.Equal(t, lifecycle.TypeProcessStarted, received.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sub1 did not receive event")
	}

	select {
	case received := <-sub2:
		assert.Equal(t, lifecycle.TypeProcessStarted, received.Type)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("sub2 did not receive event")
	}
}

// TestBus_Unsubscribe_RemovesSubscriber verifies Unsubscribe removes the subscriber.
func TestBus_Unsubscribe_RemovesSubscriber(t *testing.T) {
	bus := events.NewBus()
	defer bus.Close()

	ch := bus.Subscribe()
	assert.Equal(t, 1, bus.SubscriberCount())

	bus.Unsubscribe(ch)
	assert.Equal(t, 0, bus.SubscriberCount())

	// channel should be closed
	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after unsubscribe")
}

// TestBus_Unsubscribe_MultipleTimes verifies Unsubscribe is idempotent.
func TestBus_Unsubscribe_MultipleTimes(t *testing.T) {
	bus := events.NewBus()
	defer bus.Close()

	ch := bus.Subscribe()

	// unsubscribe multiple times should not panic
	bus.Unsubscribe(ch)
	bus.Unsubscribe(ch)
	bus.Unsubscribe(ch)

	assert.Equal(t, 0, bus.SubscriberCount())
}

// TestBus_Publish_DropsWhenBufferFull verifies slow subscribers don't block.
func TestBus_Publish_DropsWhenBufferFull(t *testing.T) {
	bus := events.NewBus(events.WithBufferSize(2))
	defer bus.Close()

	ch := bus.Subscribe()

	// fill buffer without consuming
	bus.Publish(lifecycle.NewEvent(lifecycle.TypeProcessStarted, "event 1"))
	bus.Publish(lifecycle.NewEvent(lifecycle.TypeProcessStarted, "event 2"))

	// this should not block even though buffer is full
	done := make(chan struct{})
	go func() {
		bus.Publish(lifecycle.NewEvent(lifecycle.TypeProcessStarted, "event 3"))
		close(done)
	}()

	select {
	case <-done:
		// ok - publish didn't block
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Publish blocked when buffer was full")
	}

	// consume the buffered events
	<-ch
	<-ch
}

// TestBus_Close_StopsPublishing verifies Close shuts down the bus.
func TestBus_Close_StopsPublishing(t *testing.T) {
	bus := events.NewBus()
	ch := bus.Subscribe()

	bus.Close()

	// channel should be closed
	_, ok := <-ch
	assert.False(t, ok, "channel should be closed after bus.Close")

	// publish after close should not panic
	bus.Publish(lifecycle.NewEvent(lifecycle.TypeProcessStarted, "test"))

	// subscribe after close returns closed channel
	ch2 := bus.Subscribe()
	_, ok = <-ch2
	assert.False(t, ok, "new subscription after close should return closed channel")
}

// TestBus_ConcurrentAccess verifies thread-safety of the bus.
func TestBus_ConcurrentAccess(t *testing.T) {
	bus := events.NewBus()
	defer bus.Close()

	var wg sync.WaitGroup
	const goroutines = 10
	const events = 100

	// start subscriber goroutines
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := bus.Subscribe()
			defer bus.Unsubscribe(ch)

			count := 0
			timeout := time.After(500 * time.Millisecond)
			for count < events {
				select {
				case <-ch:
					count++
				case <-timeout:
					return
				}
			}
		}()
	}

	// start publisher goroutines
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < events; j++ {
				bus.Publish(lifecycle.NewEvent(lifecycle.TypeProcessStarted, "concurrent test"))
			}
		}()
	}

	wg.Wait()
}

// TestBus_WithBufferSize verifies buffer size option works.
func TestBus_WithBufferSize(t *testing.T) {
	bus := events.NewBus(events.WithBufferSize(128))
	defer bus.Close()

	ch := bus.Subscribe()

	// fill buffer
	for i := 0; i < 128; i++ {
		bus.Publish(lifecycle.NewEvent(lifecycle.TypeProcessStarted, "test"))
	}

	// verify all events were buffered
	count := 0
	timeout := time.After(100 * time.Millisecond)
outer:
	for {
		select {
		case <-ch:
			count++
		case <-timeout:
			break outer
		}
	}

	assert.Equal(t, 128, count)
}

// TestBus_ImplementsPublisher verifies Bus implements lifecycle.Publisher.
func TestBus_ImplementsPublisher(t *testing.T) {
	var publisher lifecycle.Publisher = events.NewBus()
	require.NotNil(t, publisher)
}
