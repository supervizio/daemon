// Package health provides domain entities and value objects for health checking.
package health_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/health"
)

// TestStatus_String tests the String method of the Status type.
//
// Params:
//   - t: the testing context.
func TestStatus_String(t *testing.T) {
	tests := []struct {
		name   string
		status health.Status
		want   string
	}{
		{"unknown", health.StatusUnknown, "unknown"},
		{"healthy", health.StatusHealthy, "healthy"},
		{"unhealthy", health.StatusUnhealthy, "unhealthy"},
		{"degraded", health.StatusDegraded, "degraded"},
		{"invalid", health.Status(99), "unknown"},
	}

	// Iterate through all status string test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.String())
		})
	}
}

// TestStatus_IsHealthy tests the IsHealthy method of the Status type.
//
// Params:
//   - t: the testing context.
func TestStatus_IsHealthy(t *testing.T) {
	tests := []struct {
		name      string
		status    health.Status
		isHealthy bool
	}{
		{"healthy is healthy", health.StatusHealthy, true},
		{"unknown is not healthy", health.StatusUnknown, false},
		{"unhealthy is not healthy", health.StatusUnhealthy, false},
		{"degraded is not healthy", health.StatusDegraded, false},
	}

	// Iterate through all IsHealthy test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isHealthy, tt.status.IsHealthy())
		})
	}
}

// TestStatus_IsUnhealthy tests the IsUnhealthy method of the Status type.
//
// Params:
//   - t: the testing context.
func TestStatus_IsUnhealthy(t *testing.T) {
	tests := []struct {
		name        string
		status      health.Status
		isUnhealthy bool
	}{
		{"unhealthy is unhealthy", health.StatusUnhealthy, true},
		{"healthy is not unhealthy", health.StatusHealthy, false},
		{"unknown is not unhealthy", health.StatusUnknown, false},
		{"degraded is not unhealthy", health.StatusDegraded, false},
	}

	// Iterate through all IsUnhealthy test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isUnhealthy, tt.status.IsUnhealthy())
		})
	}
}
