// Package health provides domain entities and value objects for health checking.
package health

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/process"
)

// Test_AggregatedHealth_computeListenerStatus tests the computeListenerStatus helper method.
func Test_AggregatedHealth_computeListenerStatus(t *testing.T) {
	tests := []struct {
		name           string
		subjects       []SubjectStatus
		expectedStatus Status
	}{
		{
			name:           "empty_subjects_healthy",
			subjects:       nil,
			expectedStatus: StatusHealthy,
		},
		{
			name: "all_ready_healthy",
			subjects: []SubjectStatus{
				{Name: "http", State: SubjectReady},
				{Name: "grpc", State: SubjectReady},
			},
			expectedStatus: StatusHealthy,
		},
		{
			name: "some_listening_degraded",
			subjects: []SubjectStatus{
				{Name: "http", State: SubjectReady},
				{Name: "grpc", State: SubjectListening},
			},
			expectedStatus: StatusDegraded,
		},
		{
			name: "all_closed_unhealthy",
			subjects: []SubjectStatus{
				{Name: "http", State: SubjectClosed},
				{Name: "grpc", State: SubjectClosed},
			},
			expectedStatus: StatusUnhealthy,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health with running process.
			h := &AggregatedHealth{
				ProcessState: process.StateRunning,
				Subjects:     tt.subjects,
			}

			// Compute listener status.
			status := h.computeListenerStatus()

			// Verify expected status.
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

// Test_AggregatedHealth_hasAnyListenerListening tests the hasAnyListenerListening helper method.
func Test_AggregatedHealth_hasAnyListenerListening(t *testing.T) {
	tests := []struct {
		name     string
		subjects []SubjectStatus
		expected bool
	}{
		{
			name:     "empty_subjects",
			subjects: nil,
			expected: false,
		},
		{
			name: "none_listening",
			subjects: []SubjectStatus{
				{Name: "http", State: SubjectClosed},
				{Name: "grpc", State: SubjectClosed},
			},
			expected: false,
		},
		{
			name: "some_listening",
			subjects: []SubjectStatus{
				{Name: "http", State: SubjectListening},
				{Name: "grpc", State: SubjectReady},
			},
			expected: true,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health.
			h := &AggregatedHealth{
				ProcessState: process.StateRunning,
				Subjects:     tt.subjects,
			}

			// Check if any listener is listening.
			result := h.hasAnyListenerListening()

			// Verify expected result.
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_AggregatedHealth_hasNonHealthyCustomStatus tests the hasNonHealthyCustomStatus helper method.
func Test_AggregatedHealth_hasNonHealthyCustomStatus(t *testing.T) {
	tests := []struct {
		name         string
		customStatus string
		expected     bool
	}{
		{
			name:         "empty_status",
			customStatus: "",
			expected:     false,
		},
		{
			name:         "healthy_status",
			customStatus: "HEALTHY",
			expected:     false,
		},
		{
			name:         "draining_status",
			customStatus: "DRAINING",
			expected:     true,
		},
		{
			name:         "maintenance_status",
			customStatus: "MAINTENANCE",
			expected:     true,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create aggregated health with custom status.
			h := &AggregatedHealth{
				ProcessState: process.StateRunning,
				CustomStatus: tt.customStatus,
			}

			// Check if custom status is non-healthy.
			result := h.hasNonHealthyCustomStatus()

			// Verify expected result.
			assert.Equal(t, tt.expected, result)
		})
	}
}
