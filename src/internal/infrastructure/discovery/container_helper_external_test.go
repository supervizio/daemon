//go:build unix

// Package discovery_test provides external tests for the container helper.
package discovery_test

import (
	"net/http"
	"testing"
)

// mockDoer is a mock HTTP client implementing the Doer interface pattern.
type mockDoer struct {
	response *http.Response
	err      error
}

// Do implements the Doer interface.
func (m *mockDoer) Do(_ *http.Request) (*http.Response, error) {
	return m.response, m.err
}

// TestDoerInterface verifies that http.Client satisfies the Doer interface pattern.
func TestDoerInterface(t *testing.T) {
	tests := []struct {
		name string
		doer interface {
			Do(*http.Request) (*http.Response, error)
		}
		wantImpl bool
	}{
		{
			name:     "http.Client implements Doer",
			doer:     &http.Client{},
			wantImpl: true,
		},
		{
			name:     "mock implements Doer",
			doer:     &mockDoer{},
			wantImpl: true,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the type implements the Doer interface pattern.
			if tt.doer == nil {
				t.Error("expected non-nil Doer implementation")
			}
		})
	}
}
