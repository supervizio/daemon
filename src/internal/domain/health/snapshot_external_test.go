// Package health provides domain entities and value objects for health checking.
package health_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/health"
)

func TestSubjectSnapshot_IsReady(t *testing.T) {
	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{"ready", health.SubjectReady, true},
		{"running", health.SubjectRunning, true},
		{"listening", health.SubjectListening, false},
		{"closed", health.SubjectClosed, false},
		{"unknown", health.SubjectUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := health.NewSubjectSnapshot("test", "listener", tt.state)
			assert.Equal(t, tt.expected, s.IsReady())
		})
	}
}

func TestSubjectSnapshot_IsListening(t *testing.T) {
	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{"listening", health.SubjectListening, true},
		{"ready_is_also_listening", health.SubjectReady, true},
		{"closed_not_listening", health.SubjectClosed, false},
		{"stopped_not_listening", health.SubjectStopped, false},
		{"unknown_not_listening", health.SubjectUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := health.NewSubjectSnapshot("test", "listener", tt.state)
			assert.Equal(t, tt.expected, s.IsListening())
		})
	}
}

func TestSubjectSnapshot_IsClosed(t *testing.T) {
	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{"closed", health.SubjectClosed, true},
		{"stopped", health.SubjectStopped, true},
		{"failed", health.SubjectFailed, true},
		{"ready", health.SubjectReady, false},
		{"running", health.SubjectRunning, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := health.NewSubjectSnapshot("test", "process", tt.state)
			assert.Equal(t, tt.expected, s.IsClosed())
		})
	}
}
