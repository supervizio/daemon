package bootstrap

import (
	"testing"
)

// mockReaperInternal is a test double for the reaper interface.
type mockReaperInternal struct {
	isPID1      bool
	startCalled bool
	stopCalled  bool
	reapCount   int
}

// Start marks that Start was called.
func (m *mockReaperInternal) Start() {
	// Record that Start was called.
	m.startCalled = true
}

// Stop marks that Stop was called.
func (m *mockReaperInternal) Stop() {
	// Record that Stop was called.
	m.stopCalled = true
}

// ReapOnce returns the configured reap count.
//
// Returns:
//   - int: the number of processes reaped.
func (m *mockReaperInternal) ReapOnce() int {
	// Return configured reap count.
	return m.reapCount
}

// IsPID1 returns the configured PID1 status.
//
// Returns:
//   - bool: true if mocked as PID1, false otherwise.
func (m *mockReaperInternal) IsPID1() bool {
	// Return configured PID1 status.
	return m.isPID1
}

// TestProvideReaper_Internal verifies ProvideReaper behavior for PID1 scenarios.
func TestProvideReaper_Internal(t *testing.T) {
	t.Parallel()

	// Define test cases for ProvideReaper.
	tests := []struct {
		name       string
		isPID1     bool
		wantNil    bool
		wantSameAs bool
	}{
		{
			name:       "returns reaper when PID1",
			isPID1:     true,
			wantNil:    false,
			wantSameAs: true,
		},
		{
			name:       "returns nil when not PID1",
			isPID1:     false,
			wantNil:    true,
			wantSameAs: false,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a mock reaper with configured PID1 status.
			mock := &mockReaperInternal{isPID1: tt.isPID1}

			// Call ProvideReaper.
			result := ProvideReaper(mock)

			// Verify nil expectation.
			if tt.wantNil && result != nil {
				t.Error("ProvideReaper should return nil")
			}

			// Verify non-nil expectation.
			if !tt.wantNil && result == nil {
				t.Error("ProvideReaper should return non-nil")
			}

			// Verify same instance expectation.
			if tt.wantSameAs && result != mock {
				t.Error("ProvideReaper returned different instance")
			}
		})
	}
}

// TestMockReaperImplementsInterface_Internal verifies the mock reaper methods work correctly.
func TestMockReaperImplementsInterface_Internal(t *testing.T) {
	t.Parallel()

	// Define test cases for mock reaper interface verification.
	tests := []struct {
		name          string
		isPID1        bool
		reapCount     int
		callStart     bool
		callStop      bool
		wantIsPID1    bool
		wantReapCount int
	}{
		{
			name:          "mock with PID1 true and reap count 5",
			isPID1:        true,
			reapCount:     5,
			callStart:     true,
			callStop:      true,
			wantIsPID1:    true,
			wantReapCount: 5,
		},
		{
			name:          "mock with PID1 false and reap count 0",
			isPID1:        false,
			reapCount:     0,
			callStart:     false,
			callStop:      false,
			wantIsPID1:    false,
			wantReapCount: 0,
		},
		{
			name:          "mock with PID1 true and only start",
			isPID1:        true,
			reapCount:     10,
			callStart:     true,
			callStop:      false,
			wantIsPID1:    true,
			wantReapCount: 10,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a mock reaper with configured values.
			mock := &mockReaperInternal{
				isPID1:    tt.isPID1,
				reapCount: tt.reapCount,
			}

			// Verify IsPID1 works correctly.
			if mock.IsPID1() != tt.wantIsPID1 {
				t.Errorf("IsPID1 returned %v, want %v", mock.IsPID1(), tt.wantIsPID1)
			}

			// Call Start if configured.
			if tt.callStart {
				mock.Start()
				// Verify Start was called.
				if !mock.startCalled {
					t.Error("Start should have been called")
				}
			}

			// Call Stop if configured.
			if tt.callStop {
				mock.Stop()
				// Verify Stop was called.
				if !mock.stopCalled {
					t.Error("Stop should have been called")
				}
			}

			// Call ReapOnce to verify it works.
			count := mock.ReapOnce()

			// Verify correct count was returned.
			if count != tt.wantReapCount {
				t.Errorf("ReapOnce returned %d, expected %d", count, tt.wantReapCount)
			}
		})
	}
}
