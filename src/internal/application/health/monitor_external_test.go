// Package health_test provides black-box tests for the health package.
package health_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apphealth "github.com/kodflow/daemon/internal/application/health"
	domain "github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/kodflow/daemon/internal/domain/process"
)

// mockCreator is a mock implementation of Creator for testing.
type mockCreator struct {
	probers map[string]*mockProber
	err     error
}

// Create returns a mock prober for the given type.
func (f *mockCreator) Create(proberType string, _ time.Duration) (domain.Prober, error) {
	// Return error if configured.
	if f.err != nil {
		return nil, f.err
	}

	// Return prober if exists.
	if p, ok := f.probers[proberType]; ok {
		return p, nil
	}

	// Return default successful prober.
	return &mockProber{
		probeType: proberType,
		result:    domain.CheckResult{Success: true},
	}, nil
}

// TestNewProbeMonitor tests ProbeMonitor creation.
func TestNewProbeMonitor(t *testing.T) {
	tests := []struct {
		name            string
		config          apphealth.ProbeMonitorConfig
		expectedTimeout time.Duration
	}{
		{
			name:            "default_values",
			config:          apphealth.ProbeMonitorConfig{},
			expectedTimeout: domain.DefaultTimeout,
		},
		{
			name: "custom_timeout",
			config: apphealth.ProbeMonitorConfig{
				DefaultTimeout: 10 * time.Second,
			},
			expectedTimeout: 10 * time.Second,
		},
		{
			name: "with_factory",
			config: apphealth.ProbeMonitorConfig{
				Factory:        &mockCreator{},
				DefaultTimeout: 5 * time.Second,
			},
			expectedTimeout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			monitor := apphealth.NewProbeMonitor(tt.config)

			// Verify monitor is not nil.
			require.NotNil(t, monitor)

			// Verify initial status (unhealthy because process is stopped).
			assert.Equal(t, domain.StatusUnhealthy, monitor.Status())
		})
	}
}

// TestProbeMonitor_SetProcessState tests process state updates.
func TestProbeMonitor_SetProcessState(t *testing.T) {
	tests := []struct {
		name     string
		state    process.State
		expected process.State
	}{
		{
			name:     "running_state",
			state:    process.StateRunning,
			expected: process.StateRunning,
		},
		{
			name:     "stopped_state",
			state:    process.StateStopped,
			expected: process.StateStopped,
		},
		{
			name:     "failed_state",
			state:    process.StateFailed,
			expected: process.StateFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{})

			// Set process state.
			monitor.SetProcessState(tt.state)

			// Verify health reflects state.
			h := monitor.Health()
			assert.Equal(t, tt.expected, h.ProcessState)
		})
	}
}

// TestProbeMonitor_SetCustomStatus tests custom status updates.
func TestProbeMonitor_SetCustomStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected string
	}{
		{
			name:     "empty_status",
			status:   "",
			expected: "",
		},
		{
			name:     "healthy_status",
			status:   "HEALTHY",
			expected: "HEALTHY",
		},
		{
			name:     "custom_status",
			status:   "MAINTENANCE",
			expected: "MAINTENANCE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{})

			// Set custom status.
			monitor.SetCustomStatus(tt.status)

			// Verify health reflects status.
			h := monitor.Health()
			assert.Equal(t, tt.expected, h.CustomStatus)
		})
	}
}

// TestProbeMonitor_AddListener tests listener addition.
func TestProbeMonitor_AddListener(t *testing.T) {
	tests := []struct {
		name        string
		listener    *listener.Listener
		factory     *mockCreator
		expectError bool
	}{
		{
			name:        "listener_without_probe",
			listener:    listener.NewListener("test", "tcp", "localhost", 8080),
			factory:     &mockCreator{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor with factory.
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{
				Factory: tt.factory,
			})

			// Add listener.
			err := monitor.AddListener(tt.listener)

			// Verify result.
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProbeMonitor_AddListenerWithBinding tests listener with binding addition.
func TestProbeMonitor_AddListenerWithBinding(t *testing.T) {
	tests := []struct {
		name        string
		listener    *listener.Listener
		binding     *apphealth.ProbeBinding
		factory     *mockCreator
		expectError bool
	}{
		{
			name:        "listener_without_binding",
			listener:    listener.NewListener("test", "tcp", "localhost", 8080),
			binding:     nil,
			factory:     &mockCreator{},
			expectError: false,
		},
		{
			name:     "listener_with_tcp_binding",
			listener: listener.NewListener("test", "tcp", "localhost", 8080),
			binding: &apphealth.ProbeBinding{
				ListenerName: "test",
				Type:         apphealth.ProbeTCP,
				Config: apphealth.ProbeConfig{
					Timeout: time.Second,
				},
			},
			factory:     &mockCreator{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor with factory.
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{
				Factory: tt.factory,
			})

			// Add listener with binding.
			err := monitor.AddListenerWithBinding(tt.listener, tt.binding)

			// Verify result.
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestProbeMonitor_Start tests monitor start.
func TestProbeMonitor_Start(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// callTwice indicates whether to call Start twice.
		callTwice bool
		// setRunning indicates whether to set process state to running.
		setRunning bool
		// expectedHealthy is the expected healthy state after start.
		expectedHealthy bool
	}{
		{
			name:            "starts_monitor_unhealthy_when_stopped",
			callTwice:       false,
			setRunning:      false,
			expectedHealthy: false,
		},
		{
			name:            "starts_monitor_healthy_when_running",
			callTwice:       false,
			setRunning:      true,
			expectedHealthy: true,
		},
		{
			name:            "start_is_idempotent",
			callTwice:       true,
			setRunning:      true,
			expectedHealthy: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{
				Factory: &mockCreator{},
			})

			// Set process state if requested.
			if tt.setRunning {
				monitor.SetProcessState(process.StateRunning)
			}

			// Create context from test context.
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			// Start monitor.
			monitor.Start(ctx)

			// Verify health state after start.
			assert.Equal(t, tt.expectedHealthy, monitor.IsHealthy())

			// Call start again if requested.
			if tt.callTwice {
				monitor.Start(ctx)
				// Verify health state remains consistent after second start.
				assert.Equal(t, tt.expectedHealthy, monitor.IsHealthy())
			}

			// Stop monitor.
			monitor.Stop()
		})
	}
}

// TestProbeMonitor_Stop tests monitor stop.
func TestProbeMonitor_Stop(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// startFirst indicates whether to start before stopping.
		startFirst bool
		// callTwice indicates whether to call Stop twice.
		callTwice bool
	}{
		{
			name:       "stop_without_start",
			startFirst: false,
			callTwice:  false,
		},
		{
			name:       "stop_after_start",
			startFirst: true,
			callTwice:  false,
		},
		{
			name:       "stop_is_idempotent",
			startFirst: true,
			callTwice:  true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{
				Factory: &mockCreator{},
			})
			require.NotNil(t, monitor)

			// Set running state to test stop behavior.
			monitor.SetProcessState(process.StateRunning)

			// Start monitor if requested.
			if tt.startFirst {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				monitor.Start(ctx)
				// Verify monitor is healthy before stop.
				assert.True(t, monitor.IsHealthy())
			}

			// Stop monitor.
			monitor.Stop()

			// Verify monitor can still report status after stop.
			status := monitor.Status()
			assert.NotEmpty(t, status)

			// Call stop again if requested.
			if tt.callTwice {
				monitor.Stop()
				// Verify status remains accessible after double stop.
				statusAfter := monitor.Status()
				assert.Equal(t, status, statusAfter)
			}
		})
	}
}

// TestProbeMonitor_Status tests status retrieval.
func TestProbeMonitor_Status(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// processState is the process state to set.
		processState process.State
		// expectedStatus is the expected status.
		expectedStatus domain.Status
	}{
		{
			name:           "unhealthy_when_stopped",
			processState:   process.StateStopped,
			expectedStatus: domain.StatusUnhealthy,
		},
		{
			name:           "healthy_when_running",
			processState:   process.StateRunning,
			expectedStatus: domain.StatusHealthy,
		},
		{
			name:           "unhealthy_when_failed",
			processState:   process.StateFailed,
			expectedStatus: domain.StatusUnhealthy,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{})

			// Set process state.
			monitor.SetProcessState(tt.processState)

			// Verify status.
			assert.Equal(t, tt.expectedStatus, monitor.Status())
		})
	}
}

// TestProbeMonitor_Health tests health retrieval.
func TestProbeMonitor_Health(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// processState is the process state to set.
		processState process.State
		// customStatus is the custom status to set.
		customStatus string
	}{
		{
			name:         "returns_health_with_default_state",
			processState: process.StateStopped,
			customStatus: "",
		},
		{
			name:         "returns_health_with_running_state",
			processState: process.StateRunning,
			customStatus: "READY",
		},
		{
			name:         "returns_health_with_custom_status",
			processState: process.StateRunning,
			customStatus: "MAINTENANCE",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{})

			// Set process state.
			monitor.SetProcessState(tt.processState)
			// Set custom status if provided.
			if tt.customStatus != "" {
				monitor.SetCustomStatus(tt.customStatus)
			}

			// Get health.
			h := monitor.Health()

			// Verify health is not nil.
			require.NotNil(t, h)
			// Verify process state.
			assert.Equal(t, tt.processState, h.ProcessState)
			// Verify custom status.
			assert.Equal(t, tt.customStatus, h.CustomStatus)
		})
	}
}

// TestProbeMonitor_Latency tests latency retrieval.
func TestProbeMonitor_Latency(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{
			name: "returns_zero_by_default",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{})

			// Verify latency is zero by default.
			assert.Equal(t, time.Duration(0), monitor.Latency())
		})
	}
}

// TestProbeMonitor_IsHealthy tests health status check.
func TestProbeMonitor_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*apphealth.ProbeMonitor)
		expected bool
	}{
		{
			name: "unhealthy_by_default",
			setup: func(_ *apphealth.ProbeMonitor) {
				// No setup - default state.
			},
			expected: false,
		},
		{
			name: "healthy_when_running",
			setup: func(m *apphealth.ProbeMonitor) {
				m.SetProcessState(process.StateRunning)
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create monitor.
			monitor := apphealth.NewProbeMonitor(apphealth.ProbeMonitorConfig{})

			// Apply setup.
			tt.setup(monitor)

			// Verify health.
			assert.Equal(t, tt.expected, monitor.IsHealthy())
		})
	}
}
