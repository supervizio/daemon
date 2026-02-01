package monitoring_test

import (
	"context"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/application/monitoring"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProberFactory is a mock prober factory for testing.
type mockProberFactory struct{}

func (m *mockProberFactory) Create(proberType string, timeout time.Duration) (health.Prober, error) {
	return &mockProber{proberType: proberType}, nil
}

// mockProber is a mock prober for testing.
type mockProber struct {
	proberType string
}

func (m *mockProber) Probe(ctx context.Context, tgt health.Target) health.CheckResult {
	return health.NewSuccessCheckResult(10*time.Millisecond, "healthy")
}

func (m *mockProber) Type() string {
	return m.proberType
}

func TestExternalMonitor(t *testing.T) {
	// testCase defines a test case for ExternalMonitor operations.
	type testCase struct {
		name       string
		setupFunc  func() *monitoring.ExternalMonitor
		actionFunc func(*testing.T, *monitoring.ExternalMonitor)
		verifyFunc func(*testing.T, *monitoring.ExternalMonitor)
	}

	// tests defines all test cases for ExternalMonitor.
	tests := []testCase{
		{
			name: "NewExternalMonitor creates monitor with factory",
			setupFunc: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				return monitoring.NewExternalMonitor(config)
			},
			actionFunc: func(t *testing.T, monitor *monitoring.ExternalMonitor) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, monitor *monitoring.ExternalMonitor) {
				require.NotNil(t, monitor)
				assert.False(t, monitor.IsRunning())
				assert.Equal(t, 0, monitor.TargetCount())
			},
		},
		{
			name: "AddTarget adds target successfully",
			setupFunc: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				return monitoring.NewExternalMonitor(config)
			},
			actionFunc: func(t *testing.T, monitor *monitoring.ExternalMonitor) {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				err := monitor.AddTarget(tgt)
				require.NoError(t, err)

				// Verify duplicate target returns error.
				err = monitor.AddTarget(tgt)
				assert.Error(t, err)
				assert.Equal(t, monitoring.ErrTargetExists, err)
			},
			verifyFunc: func(t *testing.T, monitor *monitoring.ExternalMonitor) {
				assert.Equal(t, 1, monitor.TargetCount())
			},
		},
		{
			name: "RemoveTarget removes target successfully",
			setupFunc: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				monitor := monitoring.NewExternalMonitor(config)
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				_ = monitor.AddTarget(tgt)
				return monitor
			},
			actionFunc: func(t *testing.T, monitor *monitoring.ExternalMonitor) {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				err := monitor.RemoveTarget(tgt.ID)
				require.NoError(t, err)

				// Verify removing non-existent target returns error.
				err = monitor.RemoveTarget(tgt.ID)
				assert.Error(t, err)
				assert.Equal(t, monitoring.ErrTargetNotFound, err)
			},
			verifyFunc: func(t *testing.T, monitor *monitoring.ExternalMonitor) {
				assert.Equal(t, 0, monitor.TargetCount())
			},
		},
		{
			name: "StartStop manages monitor lifecycle",
			setupFunc: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				return monitoring.NewExternalMonitor(config)
			},
			actionFunc: func(t *testing.T, monitor *monitoring.ExternalMonitor) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				monitor.Start(ctx)
				assert.True(t, monitor.IsRunning())

				monitor.Stop()
			},
			verifyFunc: func(t *testing.T, monitor *monitoring.ExternalMonitor) {
				assert.False(t, monitor.IsRunning())
			},
		},
		{
			name: "Health returns correct summary",
			setupFunc: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				monitor := monitoring.NewExternalMonitor(config)

				tgt1 := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt2 := target.NewDockerTarget("container-1", "redis")

				_ = monitor.AddTarget(tgt1)
				_ = monitor.AddTarget(tgt2)

				return monitor
			},
			actionFunc: func(t *testing.T, monitor *monitoring.ExternalMonitor) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, monitor *monitoring.ExternalMonitor) {
				summary := monitor.Health()
				assert.Equal(t, 2, summary.Total)
				assert.Equal(t, 1, summary.ByType[target.TypeRemote])
				assert.Equal(t, 1, summary.ByType[target.TypeDocker])
			},
		},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			monitor := tc.setupFunc()
			tc.actionFunc(t, monitor)
			tc.verifyFunc(t, monitor)
		})
	}
}

// TestNewExternalMonitor tests the NewExternalMonitor constructor.
func TestNewExternalMonitor(t *testing.T) {
	// testCase defines a test case for NewExternalMonitor.
	type testCase struct {
		name   string
		config monitoring.Config
		verify func(*testing.T, *monitoring.ExternalMonitor)
	}

	tests := []testCase{
		{
			name:   "creates monitor with default config",
			config: monitoring.NewConfig(),
			verify: func(t *testing.T, m *monitoring.ExternalMonitor) {
				require.NotNil(t, m)
				assert.NotNil(t, m.Registry())
				assert.False(t, m.IsRunning())
				assert.Equal(t, 0, m.TargetCount())
			},
		},
		{
			name: "creates monitor with factory",
			config: monitoring.NewConfig().
				WithFactory(&mockProberFactory{}),
			verify: func(t *testing.T, m *monitoring.ExternalMonitor) {
				require.NotNil(t, m)
				assert.NotNil(t, m.Registry())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := monitoring.NewExternalMonitor(tc.config)
			tc.verify(t, m)
		})
	}
}

// TestAddTarget tests the AddTarget method.
func TestExternalMonitor_AddTarget(t *testing.T) {
	// testCase defines a test case for AddTarget.
	type testCase struct {
		name      string
		setup     func() *monitoring.ExternalMonitor
		target    *target.ExternalTarget
		expectErr bool
	}

	tests := []testCase{
		{
			name: "adds target successfully",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				return monitoring.NewExternalMonitor(config)
			},
			target:    target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			expectErr: false,
		},
		{
			name: "returns error for duplicate target",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				m := monitoring.NewExternalMonitor(config)
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				_ = m.AddTarget(tgt)
				return m
			},
			target:    target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			expectErr: true,
		},
		{
			name: "adds target without probe",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			target: func() *target.ExternalTarget {
				tgt := target.NewRemoteTarget("test-2", "localhost:9090", "")
				tgt.ProbeType = ""
				return tgt
			}(),
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			err := m.AddTarget(tc.target)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRemoveTarget tests the RemoveTarget method.
func TestExternalMonitor_RemoveTarget(t *testing.T) {
	// testCase defines a test case for RemoveTarget.
	type testCase struct {
		name      string
		setup     func() *monitoring.ExternalMonitor
		targetID  string
		expectErr bool
	}

	tests := []testCase{
		{
			name: "removes existing target",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				m := monitoring.NewExternalMonitor(config)
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				_ = m.AddTarget(tgt)
				return m
			},
			targetID:  "remote:test-1",
			expectErr: false,
		},
		{
			name: "returns error for non-existent target",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			targetID:  "non-existent",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			err := m.RemoveTarget(tc.targetID)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestAddTargets tests the AddTargets method.
func TestExternalMonitor_AddTargets(t *testing.T) {
	// testCase defines a test case for AddTargets.
	type testCase struct {
		name      string
		setup     func() *monitoring.ExternalMonitor
		targets   []*target.ExternalTarget
		expectErr bool
	}

	tests := []testCase{
		{
			name: "adds multiple targets successfully",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				return monitoring.NewExternalMonitor(config)
			},
			targets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("test-2", "localhost:9090", "tcp"),
			},
			expectErr: false,
		},
		{
			name: "returns error if any target fails",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				m := monitoring.NewExternalMonitor(config)
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				_ = m.AddTarget(tgt)
				return m
			},
			targets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("test-2", "localhost:9090", "tcp"),
			},
			expectErr: true,
		},
		{
			name: "succeeds with empty slice",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			targets:   []*target.ExternalTarget{},
			expectErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			err := m.AddTargets(tc.targets)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestStart tests the Start method.
func TestExternalMonitor_Start(t *testing.T) {
	// testCase defines a test case for Start.
	type testCase struct {
		name   string
		setup  func() *monitoring.ExternalMonitor
		verify func(*testing.T, *monitoring.ExternalMonitor)
	}

	tests := []testCase{
		{
			name: "starts monitor successfully",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				return monitoring.NewExternalMonitor(config)
			},
			verify: func(t *testing.T, m *monitoring.ExternalMonitor) {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				m.Start(ctx)
				assert.True(t, m.IsRunning())
				m.Stop()
			},
		},
		{
			name: "ignores duplicate start calls",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				return monitoring.NewExternalMonitor(config)
			},
			verify: func(t *testing.T, m *monitoring.ExternalMonitor) {
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()

				m.Start(ctx)
				m.Start(ctx)
				assert.True(t, m.IsRunning())
				m.Stop()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			tc.verify(t, m)
		})
	}
}

// TestStop tests the Stop method.
func TestExternalMonitor_Stop(t *testing.T) {
	// testCase defines a test case for Stop.
	type testCase struct {
		name   string
		setup  func() *monitoring.ExternalMonitor
		verify func(*testing.T, *monitoring.ExternalMonitor)
	}

	tests := []testCase{
		{
			name: "stops running monitor",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				m := monitoring.NewExternalMonitor(config)
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				m.Start(ctx)
				return m
			},
			verify: func(t *testing.T, m *monitoring.ExternalMonitor) {
				m.Stop()
				assert.False(t, m.IsRunning())
			},
		},
		{
			name: "ignores stop when not running",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			verify: func(t *testing.T, m *monitoring.ExternalMonitor) {
				m.Stop()
				assert.False(t, m.IsRunning())
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			tc.verify(t, m)
		})
	}
}

// TestHealth tests the Health method.
func TestExternalMonitor_Health(t *testing.T) {
	// testCase defines a test case for Health.
	type testCase struct {
		name   string
		setup  func() *monitoring.ExternalMonitor
		verify func(*testing.T, monitoring.HealthSummary)
	}

	tests := []testCase{
		{
			name: "returns empty summary for new monitor",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			verify: func(t *testing.T, summary monitoring.HealthSummary) {
				assert.Equal(t, 0, summary.Total)
			},
		},
		{
			name: "returns summary with targets",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				m := monitoring.NewExternalMonitor(config)
				tgt1 := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt2 := target.NewDockerTarget("container-1", "redis")
				_ = m.AddTarget(tgt1)
				_ = m.AddTarget(tgt2)
				return m
			},
			verify: func(t *testing.T, summary monitoring.HealthSummary) {
				assert.Equal(t, 2, summary.Total)
				assert.Equal(t, 1, summary.ByType[target.TypeRemote])
				assert.Equal(t, 1, summary.ByType[target.TypeDocker])
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			summary := m.Health()
			tc.verify(t, summary)
		})
	}
}

// TestIsRunning tests the IsRunning method.
func TestExternalMonitor_IsRunning(t *testing.T) {
	// testCase defines a test case for IsRunning.
	type testCase struct {
		name     string
		setup    func() *monitoring.ExternalMonitor
		expected bool
	}

	tests := []testCase{
		{
			name: "returns false for new monitor",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			expected: false,
		},
		{
			name: "returns true when running",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				m := monitoring.NewExternalMonitor(config)
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer cancel()
				m.Start(ctx)
				return m
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			defer m.Stop()
			assert.Equal(t, tc.expected, m.IsRunning())
		})
	}
}

// TestTargetCount tests the TargetCount method.
func TestExternalMonitor_TargetCount(t *testing.T) {
	// testCase defines a test case for TargetCount.
	type testCase struct {
		name     string
		setup    func() *monitoring.ExternalMonitor
		expected int
	}

	tests := []testCase{
		{
			name: "returns zero for new monitor",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			expected: 0,
		},
		{
			name: "returns correct count with targets",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				m := monitoring.NewExternalMonitor(config)
				tgt1 := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt2 := target.NewRemoteTarget("test-2", "localhost:9090", "tcp")
				_ = m.AddTarget(tgt1)
				_ = m.AddTarget(tgt2)
				return m
			},
			expected: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			assert.Equal(t, tc.expected, m.TargetCount())
		})
	}
}

// TestGetStatus tests the GetStatus method.
func TestExternalMonitor_GetStatus(t *testing.T) {
	// testCase defines a test case for GetStatus.
	type testCase struct {
		name     string
		setup    func() *monitoring.ExternalMonitor
		targetID string
		isNil    bool
	}

	tests := []testCase{
		{
			name: "returns nil for non-existent target",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			targetID: "non-existent",
			isNil:    true,
		},
		{
			name: "returns status for existing target",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				m := monitoring.NewExternalMonitor(config)
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				_ = m.AddTarget(tgt)
				return m
			},
			targetID: "remote:test-1",
			isNil:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			status := m.GetStatus(tc.targetID)
			if tc.isNil {
				assert.Nil(t, status)
			} else {
				assert.NotNil(t, status)
			}
		})
	}
}

// TestAllStatuses tests the AllStatuses method.
func TestExternalMonitor_AllStatuses(t *testing.T) {
	// testCase defines a test case for AllStatuses.
	type testCase struct {
		name     string
		setup    func() *monitoring.ExternalMonitor
		expected int
	}

	tests := []testCase{
		{
			name: "returns empty slice for new monitor",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			expected: 0,
		},
		{
			name: "returns all statuses",
			setup: func() *monitoring.ExternalMonitor {
				config := monitoring.NewConfig().
					WithFactory(&mockProberFactory{})
				m := monitoring.NewExternalMonitor(config)
				tgt1 := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt2 := target.NewRemoteTarget("test-2", "localhost:9090", "tcp")
				_ = m.AddTarget(tgt1)
				_ = m.AddTarget(tgt2)
				return m
			},
			expected: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			statuses := m.AllStatuses()
			assert.Len(t, statuses, tc.expected)
		})
	}
}

// TestExternalMonitor_Registry tests the Registry method.
func TestExternalMonitor_Registry(t *testing.T) {
	// testCase defines a test case for Registry.
	type testCase struct {
		name   string
		setup  func() *monitoring.ExternalMonitor
		verify func(*testing.T, *monitoring.Registry)
	}

	tests := []testCase{
		{
			name: "returns_registry_instance",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			verify: func(t *testing.T, reg *monitoring.Registry) {
				require.NotNil(t, reg)
			},
		},
		{
			name: "registry_persists_across_calls",
			setup: func() *monitoring.ExternalMonitor {
				return monitoring.NewExternalMonitor(monitoring.NewConfig())
			},
			verify: func(t *testing.T, reg *monitoring.Registry) {
				m := monitoring.NewExternalMonitor(monitoring.NewConfig())
				reg1 := m.Registry()
				reg2 := m.Registry()
				assert.Equal(t, reg1, reg2)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup()
			registry := m.Registry()
			tc.verify(t, registry)
		})
	}
}
