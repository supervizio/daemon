// Package health_test provides black-box tests for the health package.
package health_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/kodflow/daemon/internal/domain/process"
)

// listenerName generates a listener name from an index.
func listenerName(i int) string {
	return "listener-" + strconv.Itoa(i)
}

// TestNewAggregatedHealth tests aggregated health creation.
func TestNewAggregatedHealth(t *testing.T) {
	tests := []struct {
		name         string
		processState process.State
	}{
		{
			name:         "running_state",
			processState: process.StateRunning,
		},
		{
			name:         "stopped_state",
			processState: process.StateStopped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(tt.processState)

			// Verify fields.
			require.NotNil(t, h)
			assert.Equal(t, tt.processState, h.ProcessState)
			assert.Empty(t, h.Listeners)
			assert.Empty(t, h.CustomStatus)
		})
	}
}

// TestAggregatedHealth_AddListener tests listener addition.
func TestAggregatedHealth_AddListener(t *testing.T) {
	tests := []struct {
		name          string
		listenerName  string
		listenerState listener.State
	}{
		{
			name:          "closed_listener",
			listenerName:  "http",
			listenerState: listener.Closed,
		},
		{
			name:          "ready_listener",
			listenerName:  "grpc",
			listenerState: listener.Ready,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(process.StateRunning)

			// Add listener.
			h.AddListener(tt.listenerName, tt.listenerState)

			// Verify listener was added.
			require.Len(t, h.Listeners, 1)
			assert.Equal(t, tt.listenerName, h.Listeners[0].Name)
			assert.Equal(t, tt.listenerState, h.Listeners[0].State)
		})
	}
}

// TestAggregatedHealth_SetCustomStatus tests custom status setting.
func TestAggregatedHealth_SetCustomStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
	}{
		{
			name:   "empty",
			status: "",
		},
		{
			name:   "healthy",
			status: "HEALTHY",
		},
		{
			name:   "draining",
			status: "DRAINING",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(process.StateRunning)

			// Set custom status.
			h.SetCustomStatus(tt.status)

			// Verify status.
			assert.Equal(t, tt.status, h.CustomStatus)
		})
	}
}

// TestAggregatedHealth_SetLatency tests latency setting.
func TestAggregatedHealth_SetLatency(t *testing.T) {
	tests := []struct {
		name    string
		latency time.Duration
	}{
		{
			name:    "short_latency",
			latency: 10 * time.Millisecond,
		},
		{
			name:    "long_latency",
			latency: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(process.StateRunning)

			// Set latency.
			h.SetLatency(tt.latency)

			// Verify latency.
			assert.Equal(t, tt.latency, h.Latency)
		})
	}
}

// TestAggregatedHealth_Status tests status computation.
func TestAggregatedHealth_Status(t *testing.T) {
	tests := []struct {
		name           string
		processState   process.State
		listeners      []listener.State
		customStatus   string
		expectedStatus health.Status
	}{
		{
			name:           "healthy_running_no_listeners",
			processState:   process.StateRunning,
			listeners:      nil,
			customStatus:   "",
			expectedStatus: health.StatusHealthy,
		},
		{
			name:           "healthy_running_ready_listeners",
			processState:   process.StateRunning,
			listeners:      []listener.State{listener.Ready, listener.Ready},
			customStatus:   "",
			expectedStatus: health.StatusHealthy,
		},
		{
			name:           "unhealthy_stopped",
			processState:   process.StateStopped,
			listeners:      nil,
			customStatus:   "",
			expectedStatus: health.StatusUnhealthy,
		},
		{
			name:           "degraded_listening",
			processState:   process.StateRunning,
			listeners:      []listener.State{listener.Listening},
			customStatus:   "",
			expectedStatus: health.StatusDegraded,
		},
		{
			name:           "degraded_custom_status",
			processState:   process.StateRunning,
			listeners:      []listener.State{listener.Ready},
			customStatus:   "DRAINING",
			expectedStatus: health.StatusDegraded,
		},
		{
			name:           "healthy_with_healthy_custom",
			processState:   process.StateRunning,
			listeners:      []listener.State{listener.Ready},
			customStatus:   "HEALTHY",
			expectedStatus: health.StatusHealthy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(tt.processState)

			// Add listeners.
			for i, state := range tt.listeners {
				h.AddListener(listenerName(i), state)
			}

			// Set custom status.
			h.SetCustomStatus(tt.customStatus)

			// Verify status.
			assert.Equal(t, tt.expectedStatus, h.Status())
		})
	}
}

// TestAggregatedHealth_IsHealthy tests IsHealthy method.
func TestAggregatedHealth_IsHealthy(t *testing.T) {
	tests := []struct {
		name         string
		processState process.State
		expected     bool
	}{
		{
			name:         "running",
			processState: process.StateRunning,
			expected:     true,
		},
		{
			name:         "stopped",
			processState: process.StateStopped,
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(tt.processState)

			// Verify IsHealthy.
			assert.Equal(t, tt.expected, h.IsHealthy())
		})
	}
}

// TestAggregatedHealth_IsDegraded tests IsDegraded method.
func TestAggregatedHealth_IsDegraded(t *testing.T) {
	tests := []struct {
		name         string
		processState process.State
		listeners    []listener.State
		customStatus string
		expected     bool
	}{
		{
			name:         "healthy_not_degraded",
			processState: process.StateRunning,
			listeners:    nil,
			customStatus: "",
			expected:     false,
		},
		{
			name:         "degraded_custom_status",
			processState: process.StateRunning,
			listeners:    nil,
			customStatus: "DRAINING",
			expected:     true,
		},
		{
			name:         "degraded_listener_listening",
			processState: process.StateRunning,
			listeners:    []listener.State{listener.Listening},
			customStatus: "",
			expected:     true,
		},
		{
			name:         "unhealthy_not_degraded",
			processState: process.StateStopped,
			listeners:    nil,
			customStatus: "",
			expected:     false,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(tt.processState)

			// Add listeners.
			for i, state := range tt.listeners {
				h.AddListener(listenerName(i), state)
			}

			// Set custom status.
			h.SetCustomStatus(tt.customStatus)

			// Verify IsDegraded.
			assert.Equal(t, tt.expected, h.IsDegraded())
		})
	}
}

// TestAggregatedHealth_IsUnhealthy tests IsUnhealthy method.
func TestAggregatedHealth_IsUnhealthy(t *testing.T) {
	tests := []struct {
		name         string
		processState process.State
		listeners    []listener.State
		expected     bool
	}{
		{
			name:         "healthy_not_unhealthy",
			processState: process.StateRunning,
			listeners:    nil,
			expected:     false,
		},
		{
			name:         "unhealthy_stopped",
			processState: process.StateStopped,
			listeners:    nil,
			expected:     true,
		},
		{
			name:         "unhealthy_all_closed",
			processState: process.StateRunning,
			listeners:    []listener.State{listener.Closed},
			expected:     true,
		},
		{
			name:         "degraded_not_unhealthy",
			processState: process.StateRunning,
			listeners:    []listener.State{listener.Listening},
			expected:     false,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(tt.processState)

			// Add listeners.
			for i, state := range tt.listeners {
				h.AddListener(listenerName(i), state)
			}

			// Verify IsUnhealthy.
			assert.Equal(t, tt.expected, h.IsUnhealthy())
		})
	}
}

// TestAggregatedHealth_AllListenersReady tests AllListenersReady method.
func TestAggregatedHealth_AllListenersReady(t *testing.T) {
	tests := []struct {
		name      string
		listeners []listener.State
		expected  bool
	}{
		{
			name:      "empty",
			listeners: nil,
			expected:  true,
		},
		{
			name:      "all_ready",
			listeners: []listener.State{listener.Ready, listener.Ready},
			expected:  true,
		},
		{
			name:      "some_listening",
			listeners: []listener.State{listener.Ready, listener.Listening},
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(process.StateRunning)

			// Add listeners.
			for i, state := range tt.listeners {
				h.AddListener(listenerName(i), state)
			}

			// Verify AllListenersReady.
			assert.Equal(t, tt.expected, h.AllListenersReady())
		})
	}
}

// TestAggregatedHealth_ReadyListenerCount tests ReadyListenerCount method.
func TestAggregatedHealth_ReadyListenerCount(t *testing.T) {
	tests := []struct {
		name      string
		listeners []listener.State
		expected  int
	}{
		{
			name:      "empty",
			listeners: nil,
			expected:  0,
		},
		{
			name:      "two_ready",
			listeners: []listener.State{listener.Ready, listener.Ready},
			expected:  2,
		},
		{
			name:      "one_ready_one_listening",
			listeners: []listener.State{listener.Ready, listener.Listening},
			expected:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(process.StateRunning)

			// Add listeners.
			for i, state := range tt.listeners {
				h.AddListener(listenerName(i), state)
			}

			// Verify ReadyListenerCount.
			assert.Equal(t, tt.expected, h.ReadyListenerCount())
		})
	}
}

// TestAggregatedHealth_TotalListenerCount tests TotalListenerCount method.
func TestAggregatedHealth_TotalListenerCount(t *testing.T) {
	tests := []struct {
		name      string
		listeners []listener.State
		expected  int
	}{
		{
			name:      "empty",
			listeners: nil,
			expected:  0,
		},
		{
			name:      "three_listeners",
			listeners: []listener.State{listener.Ready, listener.Listening, listener.Closed},
			expected:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := health.NewAggregatedHealth(process.StateRunning)

			// Add listeners.
			for i, state := range tt.listeners {
				h.AddListener(listenerName(i), state)
			}

			// Verify TotalListenerCount.
			assert.Equal(t, tt.expected, h.TotalListenerCount())
		})
	}
}
