// Package shared_test provides external tests for the shared domain package.
package shared_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/shared"
)

// TestNewRealClock tests the NewRealClock constructor.
//
// Params:
//   - t: the testing context.
func TestNewRealClock(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "creates_non_nil_clock",
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new RealClock instance.
			clock := shared.NewRealClock()

			// Assert clock is not nil.
			assert.NotNil(t, clock)
		})
	}
}

// TestRealClock_Now tests the RealClock.Now method.
//
// Params:
//   - t: the testing context.
func TestRealClock_Now(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "returns_current_time",
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create clock.
			clock := shared.NewRealClock()

			// Capture time bounds.
			before := time.Now()
			result := clock.Now()
			after := time.Now()

			// Assert time is within bounds.
			assert.True(t, result.After(before) || result.Equal(before))
			assert.True(t, result.Before(after) || result.Equal(after))
		})
	}
}

// TestDefaultClock tests the DefaultClock package variable.
//
// Params:
//   - t: the testing context.
func TestDefaultClock(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "default_clock_is_set",
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assert DefaultClock is not nil.
			assert.NotNil(t, shared.DefaultClock)

			// Assert DefaultClock returns valid time.
			before := time.Now()
			result := shared.DefaultClock.Now()
			after := time.Now()

			// Assert time is within bounds.
			assert.True(t, result.After(before) || result.Equal(before))
			assert.True(t, result.Before(after) || result.Equal(after))
		})
	}
}

// TestNower_Interface tests that RealClock implements Nower interface.
//
// Params:
//   - t: the testing context.
func TestNower_Interface(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "realclock_implements_nower",
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create RealClock and assign to Nower interface.
			var nower shared.Nower = shared.NewRealClock()

			// Assert interface is satisfied.
			assert.NotNil(t, nower)

			// Assert Now method works through interface.
			result := nower.Now()
			assert.False(t, result.IsZero())
		})
	}
}
