package bootstrap_test

import (
	"errors"
	"testing"

	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
	"github.com/kodflow/daemon/internal/bootstrap"
	domainconfig "github.com/kodflow/daemon/internal/domain/config"
)

// mockReaper is a test double for bootstrap.ReaperMinimal interface.
type mockReaper struct {
	isPID1      bool
	startCalled bool
	stopCalled  bool
}

// IsPID1 returns the configured PID 1 status.
//
// Returns:
//   - bool: true if configured as PID 1, false otherwise.
func (m *mockReaper) IsPID1() bool {
	// Return configured PID 1 status.
	return m.isPID1
}

// Start begins the reaper process.
func (m *mockReaper) Start() {
	m.startCalled = true
}

// Stop ends the reaper process.
func (m *mockReaper) Stop() {
	m.stopCalled = true
}

// ReapOnce performs a single reap cycle.
//
// Returns:
//   - int: the number of processes reaped (always 0 in mock).
func (m *mockReaper) ReapOnce() int {
	// Return 0 for mock implementation.
	return 0
}

// TestProvideReaper verifies ProvideReaper behavior with various inputs.
func TestProvideReaper(t *testing.T) {
	t.Parallel()

	// Define test cases for ProvideReaper.
	tests := []struct {
		name     string
		isPID1   bool
		wantNil  bool
		wantSame bool
	}{
		{
			name:     "returns nil when not PID 1",
			isPID1:   false,
			wantNil:  true,
			wantSame: false,
		},
		{
			name:     "returns reaper when PID 1",
			isPID1:   true,
			wantNil:  false,
			wantSame: true,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a mock reaper with configured PID 1 status.
			mock := &mockReaper{isPID1: tt.isPID1}

			// Call ProvideReaper.
			got := bootstrap.ProvideReaper(mock)

			// Verify nil result when not PID 1.
			if (got == nil) != tt.wantNil {
				t.Errorf("ProvideReaper() nil = %v, want nil = %v", got == nil, tt.wantNil)
			}

			// Verify same instance is returned when PID 1.
			if tt.wantSame {
				// Type assert to compare underlying values.
				gotMock, ok := got.(*mockReaper)
				if !ok {
					t.Errorf("ProvideReaper() returned wrong type")
				}
				if gotMock != mock {
					t.Errorf("ProvideReaper() returned different instance")
				}
			}
		})
	}
}

// mockConfigLoader is a test double for appconfig.Loader interface.
type mockConfigLoader struct {
	cfg *domainconfig.Config
	err error
}

// Load returns the configured config or error.
//
// Params:
//   - path: the configuration file path (ignored in mock).
//
// Returns:
//   - *domainconfig.Config: the mocked configuration.
//   - error: the mocked error.
func (m *mockConfigLoader) Load(path string) (*domainconfig.Config, error) {
	// Return configured mock response.
	return m.cfg, m.err
}

// TestLoadConfig verifies LoadConfig behavior with various inputs.
func TestLoadConfig(t *testing.T) {
	t.Parallel()

	// Define test cases for LoadConfig.
	tests := []struct {
		name        string
		cfg         *domainconfig.Config
		err         error
		path        string
		wantErr     bool
		wantNilCfg  bool
		wantSameCfg bool
	}{
		{
			name:        "success returns config without error",
			cfg:         &domainconfig.Config{},
			err:         nil,
			path:        "/test/path",
			wantErr:     false,
			wantNilCfg:  false,
			wantSameCfg: true,
		},
		{
			name:        "error returns nil config with error",
			cfg:         nil,
			err:         errors.New("load failed"),
			path:        "/test/path",
			wantErr:     true,
			wantNilCfg:  true,
			wantSameCfg: false,
		},
		{
			name:        "success with empty path",
			cfg:         &domainconfig.Config{},
			err:         nil,
			path:        "",
			wantErr:     false,
			wantNilCfg:  false,
			wantSameCfg: true,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a mock loader with configured response.
			loader := &mockConfigLoader{
				cfg: tt.cfg,
				err: tt.err,
			}

			// Call LoadConfig.
			cfg, err := bootstrap.LoadConfig(loader, tt.path)

			// Verify error expectation.
			if tt.wantErr && err == nil {
				t.Error("LoadConfig should return error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("LoadConfig returned unexpected error: %v", err)
			}

			// Verify nil config expectation.
			if tt.wantNilCfg && cfg != nil {
				t.Error("LoadConfig should return nil config")
			}
			if !tt.wantNilCfg && cfg == nil {
				t.Error("LoadConfig should return non-nil config")
			}

			// Verify same config expectation.
			if tt.wantSameCfg && cfg != tt.cfg {
				t.Error("LoadConfig returned wrong config")
			}

			// Verify error matches when expected.
			if tt.wantErr && !errors.Is(err, tt.err) {
				t.Errorf("LoadConfig returned wrong error: got %v, want %v", err, tt.err)
			}
		})
	}
}

// TestNewApp verifies App creation with supervisor.
func TestNewApp(t *testing.T) {
	t.Parallel()

	// Define test cases for NewApp.
	tests := []struct {
		name          string
		supervisorNil bool
		wantNilApp    bool
		wantSameSup   bool
	}{
		{
			name:          "creates app with supervisor",
			supervisorNil: false,
			wantNilApp:    false,
			wantSameSup:   true,
		},
		{
			name:          "creates app with nil supervisor",
			supervisorNil: true,
			wantNilApp:    false,
			wantSameSup:   true,
		},
	}

	// Run all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a supervisor instance based on test case.
			var sup *appsupervisor.Supervisor
			if !tt.supervisorNil {
				sup = &appsupervisor.Supervisor{}
			}

			// Call NewApp.
			app := bootstrap.NewApp(sup)

			// Verify app was created.
			if tt.wantNilApp && app != nil {
				t.Error("NewApp should return nil App")
			}
			if !tt.wantNilApp && app == nil {
				t.Fatal("NewApp should return non-nil App")
			}

			// Verify supervisor was set correctly.
			if tt.wantSameSup && app.Supervisor != sup {
				t.Error("NewApp set wrong supervisor")
			}
		})
	}
}
