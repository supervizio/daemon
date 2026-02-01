package target_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

func TestState_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state target.State
		want  string
	}{
		{"unknown", target.StateUnknown, "unknown"},
		{"healthy", target.StateHealthy, "healthy"},
		{"unhealthy", target.StateUnhealthy, "unhealthy"},
		{"degraded", target.StateDegraded, "degraded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.state.String())
		})
	}
}

func TestState_IsHealthy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state target.State
		want  bool
	}{
		{"healthy is healthy", target.StateHealthy, true},
		{"unknown not healthy", target.StateUnknown, false},
		{"unhealthy not healthy", target.StateUnhealthy, false},
		{"degraded not healthy", target.StateDegraded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.state.IsHealthy())
		})
	}
}

func TestStatus_UpdateFromProbe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		initialState     target.State
		probeResult      health.CheckResult
		successThreshold int
		failureThreshold int
		wantState        target.State
		wantSuccesses    int
		wantFailures     int
	}{
		{
			name:         "success increments success count",
			initialState: target.StateUnknown,
			probeResult: health.CheckResult{
				Success: true,
				Output:  "OK",
			},
			successThreshold: 1,
			failureThreshold: 3,
			wantState:        target.StateHealthy,
			wantSuccesses:    1,
			wantFailures:     0,
		},
		{
			name:         "failure increments failure count",
			initialState: target.StateHealthy,
			probeResult: health.CheckResult{
				Success: false,
				Output:  "Failed",
			},
			successThreshold: 1,
			failureThreshold: 3,
			wantState:        target.StateHealthy,
			wantSuccesses:    0,
			wantFailures:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
			status := target.NewStatus(tgt)
			status.State = tt.initialState

			status.UpdateFromProbe(tt.probeResult, tt.successThreshold, tt.failureThreshold)

			assert.Equal(t, tt.wantState, status.State)
			assert.Equal(t, tt.wantSuccesses, status.ConsecutiveSuccesses)
			assert.Equal(t, tt.wantFailures, status.ConsecutiveFailures)
		})
	}
}

func TestStatus(t *testing.T) {
	t.Parallel()

	// testCase defines a test case for Status operations.
	type testCase struct {
		name       string
		setupFunc  func() *target.Status
		verifyFunc func(*testing.T, *target.Status)
	}

	// tests defines all test cases for Status.
	tests := []testCase{
		{
			name: "NewStatus initializes correctly",
			setupFunc: func() *target.Status {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				return target.NewStatus(tgt)
			},
			verifyFunc: func(t *testing.T, status *target.Status) {
				assert.NotNil(t, status)
				assert.Equal(t, "test:1", status.TargetID)
				assert.Equal(t, "test", status.TargetName)
				assert.Equal(t, target.TypeDocker, status.TargetType)
				assert.Equal(t, target.StateUnknown, status.State)
			},
		},
		{
			name: "MarkHealthy updates state correctly",
			setupFunc: func() *target.Status {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				status := target.NewStatus(tgt)
				status.State = target.StateUnhealthy
				status.ConsecutiveFailures = 5
				status.MarkHealthy("recovered")
				return status
			},
			verifyFunc: func(t *testing.T, status *target.Status) {
				assert.Equal(t, target.StateHealthy, status.State)
				assert.Equal(t, "recovered", status.Message)
				assert.Equal(t, 0, status.ConsecutiveFailures)
				assert.False(t, status.LastStateChange.IsZero())
			},
		},
		{
			name: "MarkUnhealthy updates state correctly",
			setupFunc: func() *target.Status {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				status := target.NewStatus(tgt)
				status.State = target.StateHealthy
				status.ConsecutiveSuccesses = 5
				status.MarkUnhealthy("probe failed")
				return status
			},
			verifyFunc: func(t *testing.T, status *target.Status) {
				assert.Equal(t, target.StateUnhealthy, status.State)
				assert.Equal(t, "probe failed", status.Message)
				assert.Equal(t, 0, status.ConsecutiveSuccesses)
				assert.False(t, status.LastStateChange.IsZero())
			},
		},
		{
			name: "Latency returns 0 when no probe",
			setupFunc: func() *target.Status {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				return target.NewStatus(tgt)
			},
			verifyFunc: func(t *testing.T, status *target.Status) {
				assert.Equal(t, time.Duration(0), status.Latency(), "no probe should return 0")
			},
		},
		{
			name: "Latency returns correct value after probe",
			setupFunc: func() *target.Status {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				status := target.NewStatus(tgt)
				result := health.CheckResult{
					Success: true,
					Latency: 100 * time.Millisecond,
				}
				status.UpdateFromProbe(result, 1, 3)
				return status
			},
			verifyFunc: func(t *testing.T, status *target.Status) {
				assert.Equal(t, 100*time.Millisecond, status.Latency())
			},
		},
		{
			name: "SinceLastProbe returns 0 when no probe",
			setupFunc: func() *target.Status {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				return target.NewStatus(tgt)
			},
			verifyFunc: func(t *testing.T, status *target.Status) {
				assert.Equal(t, time.Duration(0), status.SinceLastProbe(), "no probe should return 0")
			},
		},
		{
			name: "SinceLastProbe returns positive duration after probe",
			setupFunc: func() *target.Status {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				status := target.NewStatus(tgt)
				result := health.CheckResult{Success: true}
				status.UpdateFromProbe(result, 1, 3)
				return status
			},
			verifyFunc: func(t *testing.T, status *target.Status) {
				assert.Greater(t, status.SinceLastProbe(), time.Duration(0))
			},
		},
		{
			name: "SinceLastStateChange returns 0 when no state change",
			setupFunc: func() *target.Status {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				return target.NewStatus(tgt)
			},
			verifyFunc: func(t *testing.T, status *target.Status) {
				assert.Equal(t, time.Duration(0), status.SinceLastStateChange(), "no state change should return 0")
			},
		},
		{
			name: "SinceLastStateChange returns positive duration after state change",
			setupFunc: func() *target.Status {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				status := target.NewStatus(tgt)
				status.MarkHealthy("ok")
				return status
			},
			verifyFunc: func(t *testing.T, status *target.Status) {
				assert.Greater(t, status.SinceLastStateChange(), time.Duration(0))
			},
		},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			status := tc.setupFunc()
			tc.verifyFunc(t, status)
		})
	}
}
