package target_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

func TestWatcherEvents(t *testing.T) {
	t.Parallel()

	// testCase defines a test case for watcher event creation.
	type testCase struct {
		name         string
		setupFunc    func() target.Event
		expectedType target.EventType
		verifyFunc   func(*testing.T, target.Event)
	}

	// tests defines all test cases for watcher events.
	tests := []testCase{
		{
			name: "NewAddedEvent creates correct event",
			setupFunc: func() target.Event {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				return target.NewAddedEvent(tgt)
			},
			expectedType: target.EventAdded,
			verifyFunc: func(t *testing.T, event target.Event) {
				assert.Equal(t, "test:1", event.Target.ID)
				assert.Equal(t, "test", event.Target.Name)
				assert.Equal(t, target.TypeDocker, event.Target.Type)
			},
		},
		{
			name: "NewRemovedEvent creates correct event",
			setupFunc: func() target.Event {
				targetID := "test:1"
				return target.NewRemovedEvent(targetID)
			},
			expectedType: target.EventRemoved,
			verifyFunc: func(t *testing.T, event target.Event) {
				assert.Equal(t, "test:1", event.Target.ID)
			},
		},
		{
			name: "NewUpdatedEvent creates correct event",
			setupFunc: func() target.Event {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				return target.NewUpdatedEvent(tgt)
			},
			expectedType: target.EventUpdated,
			verifyFunc: func(t *testing.T, event target.Event) {
				assert.Equal(t, "test:1", event.Target.ID)
				assert.Equal(t, "test", event.Target.Name)
			},
		},
		{
			name: "NewHealthChangedEvent creates correct event",
			setupFunc: func() target.Event {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				return target.NewHealthChangedEvent(tgt, target.StateUnknown, target.StateHealthy)
			},
			expectedType: target.EventHealthChanged,
			verifyFunc: func(t *testing.T, event target.Event) {
				assert.Equal(t, "test:1", event.Target.ID)
				assert.Equal(t, target.StateUnknown, event.PreviousState)
				assert.Equal(t, target.StateHealthy, event.NewState)
			},
		},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			event := tc.setupFunc()
			assert.Equal(t, tc.expectedType, event.Type)
			tc.verifyFunc(t, event)
		})
	}
}
