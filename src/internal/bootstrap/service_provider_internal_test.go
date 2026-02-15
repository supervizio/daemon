// Package bootstrap provides internal tests for service provider.
package bootstrap

import (
	"testing"

	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// TestCountTotalListeners tests the countTotalListeners helper function.
func TestCountTotalListeners(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		snapshots []appsupervisor.ServiceSnapshotForTUI
		expected  int
	}{
		{
			name:      "empty snapshots returns zero",
			snapshots: []appsupervisor.ServiceSnapshotForTUI{},
			expected:  0,
		},
		{
			name: "single service with listeners",
			snapshots: []appsupervisor.ServiceSnapshotForTUI{
				{Listeners: make([]appsupervisor.ListenerSnapshotForTUI, 3)},
			},
			expected: 3,
		},
		{
			name: "multiple services with listeners",
			snapshots: []appsupervisor.ServiceSnapshotForTUI{
				{Listeners: make([]appsupervisor.ListenerSnapshotForTUI, 2)},
				{Listeners: make([]appsupervisor.ListenerSnapshotForTUI, 3)},
			},
			expected: 5,
		},
		{
			name: "service without listeners",
			snapshots: []appsupervisor.ServiceSnapshotForTUI{
				{Listeners: nil},
			},
			expected: 0,
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := countTotalListeners(tc.snapshots)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestAppendConvertedListeners tests the appendConvertedListeners helper function.
func TestAppendConvertedListeners(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		destSize    int
		listeners   []appsupervisor.ListenerSnapshotForTUI
		expectedLen int
	}{
		{
			name:     "append to empty slice",
			destSize: 0,
			listeners: []appsupervisor.ListenerSnapshotForTUI{
				{Name: "http", Port: 8080},
			},
			expectedLen: 1,
		},
		{
			name:     "append to existing slice",
			destSize: 2,
			listeners: []appsupervisor.ListenerSnapshotForTUI{
				{Name: "http", Port: 8080},
				{Name: "grpc", Port: 9090},
			},
			expectedLen: 4,
		},
		{
			name:        "append empty listeners",
			destSize:    3,
			listeners:   []appsupervisor.ListenerSnapshotForTUI{},
			expectedLen: 3,
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// Pre-allocate dest with capacity for append.
			dest := make([]model.ListenerSnapshot, tc.destSize, tc.destSize+len(tc.listeners))
			result := appendConvertedListeners(dest, tc.listeners)
			assert.Equal(t, tc.expectedLen, len(result))
			// Verify that original listeners were preserved.
			if tc.destSize > 0 {
				// Check that appending doesn't modify existing elements.
				assert.Equal(t, dest[:tc.destSize], result[:tc.destSize])
			}
		})
	}
}

// mockServiceSnapshotsForTUIer mocks the ServiceSnapshotsForTUIer interface.
type mockServiceSnapshotsForTUIer struct {
	snapshots []appsupervisor.ServiceSnapshotForTUI
}

// ServiceSnapshotsForTUI returns the mock snapshots.
func (m *mockServiceSnapshotsForTUIer) ServiceSnapshotsForTUI() []appsupervisor.ServiceSnapshotForTUI {
	// Return stored snapshots.
	return m.snapshots
}

// TestListServices tests the ListServices method.
func TestSupervisorServiceLister_ListServices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		snapshots   []appsupervisor.ServiceSnapshotForTUI
		expectedLen int
	}{
		{
			name:        "empty snapshots",
			snapshots:   []appsupervisor.ServiceSnapshotForTUI{},
			expectedLen: 0,
		},
		{
			name: "single service",
			snapshots: []appsupervisor.ServiceSnapshotForTUI{
				{Name: "test", StateInt: 1, PID: 123},
			},
			expectedLen: 1,
		},
		{
			name: "service with listeners",
			snapshots: []appsupervisor.ServiceSnapshotForTUI{
				{
					Name:     "web",
					StateInt: 1,
					Listeners: []appsupervisor.ListenerSnapshotForTUI{
						{Name: "http", Port: 8080, Protocol: "tcp"},
					},
				},
			},
			expectedLen: 1,
		},
		{
			name: "multiple services",
			snapshots: []appsupervisor.ServiceSnapshotForTUI{
				{Name: "svc1", StateInt: 1},
				{Name: "svc2", StateInt: 0},
			},
			expectedLen: 2,
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mock := &mockServiceSnapshotsForTUIer{snapshots: tc.snapshots}
			lister := &supervisorServiceLister{sup: mock}
			result := lister.ListServices()
			assert.Len(t, result, tc.expectedLen)
		})
	}
}
