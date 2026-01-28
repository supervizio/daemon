package bootstrap

import (
	"testing"

	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
)

// mockSnapshotsProvider is a test double for ServiceSnapshotsForTUIer.
type mockSnapshotsProvider struct {
	snapshots []appsupervisor.ServiceSnapshotForTUI
}

// ServiceSnapshotsForTUI returns the configured snapshots.
//
// Returns:
//   - []appsupervisor.ServiceSnapshotForTUI: the configured snapshots.
func (m *mockSnapshotsProvider) ServiceSnapshotsForTUI() []appsupervisor.ServiceSnapshotForTUI {
	// Return configured snapshots.
	return m.snapshots
}

// Test_supervisorServiceLister_ListServices verifies the ListServices method.
//
// Params:
//   - t: testing context for assertions.
func Test_supervisorServiceLister_ListServices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		snapshots []appsupervisor.ServiceSnapshotForTUI
		wantCount int
	}{
		{
			name:      "empty_snapshots",
			snapshots: nil,
			wantCount: 0,
		},
		{
			name: "single_snapshot",
			snapshots: []appsupervisor.ServiceSnapshotForTUI{
				{
					Name:     "test-service",
					StateInt: 1,
					PID:      1234,
				},
			},
			wantCount: 1,
		},
		{
			name: "multiple_snapshots",
			snapshots: []appsupervisor.ServiceSnapshotForTUI{
				{Name: "service-a", StateInt: 1},
				{Name: "service-b", StateInt: 2},
				{Name: "service-c", StateInt: 1},
			},
			wantCount: 3,
		},
		{
			name: "snapshot_with_listeners",
			snapshots: []appsupervisor.ServiceSnapshotForTUI{
				{
					Name:     "service-with-listeners",
					StateInt: 1,
					Listeners: []appsupervisor.ListenerSnapshotForTUI{
						{Name: "http", Port: 8080, Protocol: "tcp"},
						{Name: "grpc", Port: 9090, Protocol: "tcp"},
					},
				},
			},
			wantCount: 1,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a mock lister with configured snapshots.
			mock := &mockSnapshotsProvider{snapshots: tt.snapshots}
			lister := &supervisorServiceLister{sup: mock}

			// Call ListServices.
			result := lister.ListServices()

			// Verify count matches expectation.
			if len(result) != tt.wantCount {
				t.Errorf("ListServices() returned %d items, want %d", len(result), tt.wantCount)
			}

			// Verify service names match (if any).
			for i, snap := range result {
				if i < len(tt.snapshots) && snap.Name != tt.snapshots[i].Name {
					t.Errorf("ListServices()[%d].Name = %s, want %s", i, snap.Name, tt.snapshots[i].Name)
				}
			}

			// Verify listener count for snapshot with listeners.
			if tt.name == "snapshot_with_listeners" && len(result) > 0 {
				if len(result[0].Listeners) != 2 {
					t.Errorf("ListServices()[0].Listeners count = %d, want 2", len(result[0].Listeners))
				}
			}
		})
	}
}
