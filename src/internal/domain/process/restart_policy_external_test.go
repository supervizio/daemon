// Package process_test provides external tests for restart_policy.go.
// It tests the public API of RestartTracker using black-box testing.
package process_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// TestNewRestartTracker tests the NewRestartTracker constructor.
//
// Params:
//   - t: the testing context.
func TestNewRestartTracker(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		policy     service.RestartPolicy
		maxRetries int
		delay      shared.Duration
	}{
		{
			name:       "on failure policy",
			policy:     service.RestartOnFailure,
			maxRetries: 5,
			delay:      shared.Seconds(2),
		},
		{
			name:       "always policy",
			policy:     service.RestartAlways,
			maxRetries: 3,
			delay:      shared.Seconds(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     tt.policy,
				MaxRetries: tt.maxRetries,
				Delay:      tt.delay,
			}

			tracker := process.NewRestartTracker(cfg)

			assert.NotNil(t, tracker)
			assert.Equal(t, 0, tracker.Attempts())
		})
	}
}

// TestRestartTracker_ShouldRestart_Always tests restart behavior with RestartAlways policy.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_ShouldRestart_Always(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		exitCode       int
		attempts       int
		maxRetries     int
		expectedResult bool
	}{
		{
			name:           "restart on exit code 0",
			exitCode:       0,
			attempts:       0,
			maxRetries:     3,
			expectedResult: true,
		},
		{
			name:           "restart on non-zero exit code",
			exitCode:       1,
			attempts:       0,
			maxRetries:     3,
			expectedResult: true,
		},
		{
			name:           "no restart after max retries",
			exitCode:       1,
			attempts:       3,
			maxRetries:     3,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     service.RestartAlways,
				MaxRetries: tt.maxRetries,
				Delay:      shared.Seconds(1),
			}

			tracker := process.NewRestartTracker(cfg)

			// Record the specified number of attempts.
			for range tt.attempts {
				tracker.RecordAttempt()
			}

			assert.Equal(t, tt.expectedResult, tracker.ShouldRestart(tt.exitCode))
		})
	}
}

// TestRestartTracker_ShouldRestart_OnFailure tests restart behavior with RestartOnFailure policy.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_ShouldRestart_OnFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		exitCode       int
		attempts       int
		maxRetries     int
		expectedResult bool
	}{
		{
			name:           "no restart on exit code 0",
			exitCode:       0,
			attempts:       0,
			maxRetries:     3,
			expectedResult: false,
		},
		{
			name:           "restart on exit code 1",
			exitCode:       1,
			attempts:       0,
			maxRetries:     3,
			expectedResult: true,
		},
		{
			name:           "restart on exit code 127",
			exitCode:       127,
			attempts:       0,
			maxRetries:     3,
			expectedResult: true,
		},
		{
			name:           "no restart after max retries",
			exitCode:       1,
			attempts:       3,
			maxRetries:     3,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     service.RestartOnFailure,
				MaxRetries: tt.maxRetries,
				Delay:      shared.Seconds(1),
			}

			tracker := process.NewRestartTracker(cfg)

			// Record the specified number of attempts.
			for range tt.attempts {
				tracker.RecordAttempt()
			}

			assert.Equal(t, tt.expectedResult, tracker.ShouldRestart(tt.exitCode))
		})
	}
}

// TestRestartTracker_ShouldRestart_Never tests restart behavior with RestartNever policy.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_ShouldRestart_Never(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		exitCode int
	}{
		{
			name:     "no restart on exit code 0",
			exitCode: 0,
		},
		{
			name:     "no restart on exit code 1",
			exitCode: 1,
		},
		{
			name:     "no restart on exit code 127",
			exitCode: 127,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     service.RestartNever,
				MaxRetries: 3,
				Delay:      shared.Seconds(1),
			}

			tracker := process.NewRestartTracker(cfg)

			assert.False(t, tracker.ShouldRestart(tt.exitCode))
		})
	}
}

// TestRestartTracker_ShouldRestart_Unless tests restart behavior with RestartUnless policy.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_ShouldRestart_Unless(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		exitCode       int
		attempts       int
		expectedResult bool
	}{
		{
			name:           "restart on exit code 0",
			exitCode:       0,
			attempts:       0,
			expectedResult: true,
		},
		{
			name:           "restart on exit code 1",
			exitCode:       1,
			attempts:       0,
			expectedResult: true,
		},
		{
			name:           "restart ignores max retries",
			exitCode:       1,
			attempts:       3,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     service.RestartUnless,
				MaxRetries: 3,
				Delay:      shared.Seconds(1),
			}

			tracker := process.NewRestartTracker(cfg)

			// Record the specified number of attempts.
			for range tt.attempts {
				tracker.RecordAttempt()
			}

			assert.Equal(t, tt.expectedResult, tracker.ShouldRestart(tt.exitCode))
		})
	}
}

// TestRestartTracker_RecordAttempt tests the attempt counting functionality.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_RecordAttempt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		recordCount      int
		expectedAttempts int
	}{
		{
			name:             "initial count is zero",
			recordCount:      0,
			expectedAttempts: 0,
		},
		{
			name:             "one attempt recorded",
			recordCount:      1,
			expectedAttempts: 1,
		},
		{
			name:             "two attempts recorded",
			recordCount:      2,
			expectedAttempts: 2,
		},
		{
			name:             "five attempts recorded",
			recordCount:      5,
			expectedAttempts: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     service.RestartOnFailure,
				MaxRetries: 5,
				Delay:      shared.Seconds(1),
			}

			tracker := process.NewRestartTracker(cfg)

			// Record the specified number of attempts.
			for range tt.recordCount {
				tracker.RecordAttempt()
			}

			assert.Equal(t, tt.expectedAttempts, tracker.Attempts())
		})
	}
}

// TestRestartTracker_Reset tests the reset functionality.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_Reset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		attemptsToMake int
	}{
		{
			name:           "reset after no attempts",
			attemptsToMake: 0,
		},
		{
			name:           "reset after two attempts",
			attemptsToMake: 2,
		},
		{
			name:           "reset after five attempts",
			attemptsToMake: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     service.RestartOnFailure,
				MaxRetries: 5,
				Delay:      shared.Seconds(1),
			}

			tracker := process.NewRestartTracker(cfg)

			// Record attempts.
			for range tt.attemptsToMake {
				tracker.RecordAttempt()
			}
			assert.Equal(t, tt.attemptsToMake, tracker.Attempts())

			// Verify reset clears the attempt count.
			tracker.Reset()
			assert.Equal(t, 0, tracker.Attempts())
		})
	}
}

// TestRestartTracker_MaybeReset tests conditional reset based on uptime window.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_MaybeReset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		window           time.Duration
		uptime           time.Duration
		initialAttempts  int
		expectedAttempts int
	}{
		{
			name:             "no reset when uptime is less than window",
			window:           time.Minute,
			uptime:           30 * time.Second,
			initialAttempts:  2,
			expectedAttempts: 2,
		},
		{
			name:             "reset when uptime equals window",
			window:           time.Minute,
			uptime:           time.Minute,
			initialAttempts:  2,
			expectedAttempts: 0,
		},
		{
			name:             "reset when uptime exceeds window",
			window:           time.Minute,
			uptime:           2 * time.Minute,
			initialAttempts:  3,
			expectedAttempts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     service.RestartOnFailure,
				MaxRetries: 5,
				Delay:      shared.Seconds(1),
			}

			tracker := process.NewRestartTracker(cfg)
			tracker.SetWindow(tt.window)

			// Record initial attempts.
			for range tt.initialAttempts {
				tracker.RecordAttempt()
			}

			// Attempt conditional reset.
			tracker.MaybeReset(tt.uptime)

			assert.Equal(t, tt.expectedAttempts, tracker.Attempts())
		})
	}
}

// TestRestartTracker_NextDelay tests exponential backoff delay calculation.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_NextDelay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		attempts      int
		baseDelay     shared.Duration
		maxDelay      shared.Duration
		expectedDelay time.Duration
	}{
		{
			name:          "first attempt delay",
			attempts:      0,
			baseDelay:     shared.Seconds(1),
			maxDelay:      shared.Seconds(30),
			expectedDelay: time.Second,
		},
		{
			name:          "second attempt delay",
			attempts:      1,
			baseDelay:     shared.Seconds(1),
			maxDelay:      shared.Seconds(30),
			expectedDelay: 2 * time.Second,
		},
		{
			name:          "third attempt delay",
			attempts:      2,
			baseDelay:     shared.Seconds(1),
			maxDelay:      shared.Seconds(30),
			expectedDelay: 4 * time.Second,
		},
		{
			name:          "delay capped at max",
			attempts:      5,
			baseDelay:     shared.Seconds(1),
			maxDelay:      shared.Seconds(30),
			expectedDelay: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     service.RestartOnFailure,
				MaxRetries: 10,
				Delay:      tt.baseDelay,
				DelayMax:   tt.maxDelay,
			}

			tracker := process.NewRestartTracker(cfg)

			// Record the specified number of attempts.
			for range tt.attempts {
				tracker.RecordAttempt()
			}

			assert.Equal(t, tt.expectedDelay, tracker.NextDelay())
		})
	}
}

// TestRestartTracker_NextDelay_NoMax tests delay calculation without explicit max.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_NextDelay_NoMax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		attempts    int
		baseDelay   shared.Duration
		expectedMax time.Duration
	}{
		{
			name:        "delay capped at implicit max after 5 attempts",
			attempts:    5,
			baseDelay:   shared.Seconds(1),
			expectedMax: 10 * time.Second,
		},
		{
			name:        "delay capped at implicit max after 10 attempts",
			attempts:    10,
			baseDelay:   shared.Seconds(1),
			expectedMax: 10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     service.RestartOnFailure,
				MaxRetries: 10,
				Delay:      tt.baseDelay,
				// No DelayMax set - should default to base * 10.
			}

			tracker := process.NewRestartTracker(cfg)

			// Record several attempts to test implicit max.
			for range tt.attempts {
				tracker.RecordAttempt()
			}

			// Verify delay is capped at base * 10.
			assert.LessOrEqual(t, tracker.NextDelay(), tt.expectedMax)
		})
	}
}

// TestRestartTracker_IsExhausted tests the exhaustion detection.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_IsExhausted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		attempts   int
		maxRetries int
		exhausted  bool
	}{
		{
			name:       "not exhausted initially",
			attempts:   0,
			maxRetries: 3,
			exhausted:  false,
		},
		{
			name:       "not exhausted after first attempt",
			attempts:   1,
			maxRetries: 3,
			exhausted:  false,
		},
		{
			name:       "not exhausted after second attempt",
			attempts:   2,
			maxRetries: 3,
			exhausted:  false,
		},
		{
			name:       "exhausted after reaching max retries",
			attempts:   3,
			maxRetries: 3,
			exhausted:  true,
		},
		{
			name:       "exhausted after exceeding max retries",
			attempts:   5,
			maxRetries: 3,
			exhausted:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     service.RestartOnFailure,
				MaxRetries: tt.maxRetries,
				Delay:      shared.Seconds(1),
			}

			tracker := process.NewRestartTracker(cfg)

			// Record the specified number of attempts.
			for range tt.attempts {
				tracker.RecordAttempt()
			}

			assert.Equal(t, tt.exhausted, tracker.IsExhausted())
		})
	}
}

// TestRestartTracker_ShouldRestart_UnknownPolicy tests restart behavior with unknown policy.
// This tests the default case in the switch statement.
//
// Params:
//   - t: the testing context.
func TestRestartTracker_ShouldRestart_UnknownPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		policy   service.RestartPolicy
		exitCode int
	}{
		{
			name:     "unknown policy returns false on exit code 0",
			policy:   service.RestartPolicy("unknown-policy"),
			exitCode: 0,
		},
		{
			name:     "unknown policy returns false on exit code 1",
			policy:   service.RestartPolicy("invalid"),
			exitCode: 1,
		},
		{
			name:     "empty policy returns false",
			policy:   service.RestartPolicy(""),
			exitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &service.RestartConfig{
				Policy:     tt.policy,
				MaxRetries: 3,
				Delay:      shared.Seconds(1),
			}

			tracker := process.NewRestartTracker(cfg)

			// Unknown policies should never allow restart.
			assert.False(t, tracker.ShouldRestart(tt.exitCode))
		})
	}
}
