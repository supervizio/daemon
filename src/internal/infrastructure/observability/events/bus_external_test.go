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

// TestBus_Subscribe is table-driven test for Subscribe method.
func TestBus_Subscribe(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "ReturnsChannel",
			test: func(t *testing.T) {
				bus := events.NewBus()
				defer bus.Close()

				ch := bus.Subscribe()
				require.NotNil(t, ch)
			},
		},
		{
			name: "MultipleSubscriptionsIndependent",
			test: func(t *testing.T) {
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestBus_Unsubscribe is table-driven test for Unsubscribe method.
func TestBus_Unsubscribe(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "RemovesSubscriber",
			test: func(t *testing.T) {
				bus := events.NewBus()
				defer bus.Close()

				ch := bus.Subscribe()
				assert.Equal(t, 1, bus.SubscriberCount())

				bus.Unsubscribe(ch)
				assert.Equal(t, 0, bus.SubscriberCount())

				// channel should be closed
				_, ok := <-ch
				assert.False(t, ok, "channel should be closed after unsubscribe")
			},
		},
		{
			name: "IsIdempotent",
			test: func(t *testing.T) {
				bus := events.NewBus()
				defer bus.Close()

				ch := bus.Subscribe()

				// unsubscribe multiple times should not panic
				bus.Unsubscribe(ch)
				bus.Unsubscribe(ch)
				bus.Unsubscribe(ch)

				assert.Equal(t, 0, bus.SubscriberCount())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestBus_Publish is table-driven test for Publish method.
//
// Goroutines:
//   - Spawns one goroutine per test case to verify non-blocking Publish behavior.
//   - Lifecycle: goroutine terminates immediately after Publish call completes.
//   - Synchronization: done channel signals goroutine completion with timeout fallback.
func TestBus_Publish(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "DropsWhenBufferFull",
			test: func(t *testing.T) {
				bus := events.NewBus(events.WithBufferSize(2))
				defer bus.Close()

				ch := bus.Subscribe()

				// fill buffer without consuming
				bus.Publish(lifecycle.NewEvent(lifecycle.TypeProcessStarted, "event 1"))
				bus.Publish(lifecycle.NewEvent(lifecycle.TypeProcessStarted, "event 2"))

				// this should not block even though buffer is full
				done := make(chan struct{})
				// Goroutine verifies that Publish is non-blocking when buffer is full.
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestBus_Close is table-driven test for Close method.
func TestBus_Close(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "StopsPublishing",
			test: func(t *testing.T) {
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// TestBus_ConcurrentAccess is table-driven test for thread-safety.
//
// Goroutines:
//   - Spawns numSubscribers subscriber goroutines that receive events.
//   - Spawns numPublishers publisher goroutines that publish events.
//   - Lifecycle: all goroutines terminate when WaitGroup completes or timeout expires.
//   - Synchronization: WaitGroup coordinates completion of all goroutines.
func TestBus_ConcurrentAccess(t *testing.T) {
	tests := []struct {
		name           string
		numSubscribers int
		numPublishers  int
		eventsPerPub   int
		timeoutMs      int
	}{
		{
			name:           "MultipleSubscribersAndPublishers",
			numSubscribers: 10,
			numPublishers:  10,
			eventsPerPub:   100,
			timeoutMs:      500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := events.NewBus()
			defer bus.Close()

			var wg sync.WaitGroup

			// start subscriber goroutines
			for i := range tt.numSubscribers {
				_ = i
				wg.Go(func() {
					ch := bus.Subscribe()
					defer bus.Unsubscribe(ch)

					count := 0
					timeout := time.After(time.Duration(tt.timeoutMs) * time.Millisecond)
					for count < tt.eventsPerPub {
						select {
						case <-ch:
							count++
						case <-timeout:
							return
						}
					}
				})
			}

			// start publisher goroutines
			for i := range tt.numPublishers {
				_ = i
				wg.Go(func() {
					for j := range tt.eventsPerPub {
						_ = j
						bus.Publish(lifecycle.NewEvent(lifecycle.TypeProcessStarted, "concurrent test"))
					}
				})
			}

			wg.Wait()
		})
	}
}

// TestBus_WithBufferSize is table-driven test for buffer size option.
func TestBus_WithBufferSize(t *testing.T) {
	tests := []struct {
		name       string
		bufferSize int
	}{
		{
			name:       "BufferSize128",
			bufferSize: 128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := events.NewBus(events.WithBufferSize(tt.bufferSize))
			defer bus.Close()

			ch := bus.Subscribe()

			// fill buffer
			for i := range tt.bufferSize {
				_ = i
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

			assert.Equal(t, tt.bufferSize, count)
		})
	}
}

// TestBus_ImplementsPublisher is table-driven test for interface compliance.
func TestBus_ImplementsPublisher(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ImplementsPublisher",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher := events.NewBus()
			// compile-time interface check is in bus.go
			require.NotNil(t, publisher)
		})
	}
}
