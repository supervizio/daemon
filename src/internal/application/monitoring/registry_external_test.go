package monitoring_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/application/monitoring"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	// testCase defines a test case for Registry operations.
	type testCase struct {
		name       string
		setupFunc  func() *monitoring.Registry
		actionFunc func(*testing.T, *monitoring.Registry)
		verifyFunc func(*testing.T, *monitoring.Registry)
	}

	// tests defines all test cases for Registry.
	tests := []testCase{
		{
			name: "NewRegistry creates empty registry",
			setupFunc: func() *monitoring.Registry {
				return monitoring.NewRegistry()
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				require.NotNil(t, registry)
				assert.Equal(t, 0, registry.Count())
			},
		},
		{
			name: "Add adds target successfully",
			setupFunc: func() *monitoring.Registry {
				return monitoring.NewRegistry()
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				err := registry.Add(tgt)
				require.NoError(t, err)

				// Verify duplicate target returns error.
				err = registry.Add(tgt)
				assert.Error(t, err)
				assert.Equal(t, monitoring.ErrTargetExists, err)
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				assert.Equal(t, 1, registry.Count())
			},
		},
		{
			name: "AddOrUpdate adds and updates targets",
			setupFunc: func() *monitoring.Registry {
				return monitoring.NewRegistry()
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt.Name = "original"
				registry.AddOrUpdate(tgt)
				assert.Equal(t, 1, registry.Count())

				tgt.Name = "updated"
				registry.AddOrUpdate(tgt)
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				assert.Equal(t, 1, registry.Count())
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				retrieved := registry.Get(tgt.ID)
				require.NotNil(t, retrieved)
				assert.Equal(t, "updated", retrieved.Name)
			},
		},
		{
			name: "Remove removes target successfully",
			setupFunc: func() *monitoring.Registry {
				registry := monitoring.NewRegistry()
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				_ = registry.Add(tgt)
				return registry
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				err := registry.Remove(tgt.ID)
				require.NoError(t, err)

				// Verify removing non-existent target returns error.
				err = registry.Remove(tgt.ID)
				assert.Error(t, err)
				assert.Equal(t, monitoring.ErrTargetNotFound, err)
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				assert.Equal(t, 0, registry.Count())
			},
		},
		{
			name: "Get retrieves target correctly",
			setupFunc: func() *monitoring.Registry {
				registry := monitoring.NewRegistry()
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt.Name = "test-service"
				_ = registry.Add(tgt)
				return registry
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				retrieved := registry.Get(tgt.ID)
				require.NotNil(t, retrieved)
				assert.Equal(t, tgt.ID, retrieved.ID)
				assert.Equal(t, "test-service", retrieved.Name)

				// Verify non-existent target returns nil.
				retrieved = registry.Get("non-existent")
				assert.Nil(t, retrieved)
			},
		},
		{
			name: "GetStatus returns status for target",
			setupFunc: func() *monitoring.Registry {
				registry := monitoring.NewRegistry()
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				_ = registry.Add(tgt)
				return registry
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				status := registry.GetStatus(tgt.ID)
				require.NotNil(t, status)
				assert.Equal(t, target.StateUnknown, status.State)
			},
		},
		{
			name: "ByType filters targets by type",
			setupFunc: func() *monitoring.Registry {
				registry := monitoring.NewRegistry()
				remote1 := target.NewRemoteTarget("remote-1", "localhost:8080", "tcp")
				remote2 := target.NewRemoteTarget("remote-2", "localhost:8081", "tcp")
				docker1 := target.NewDockerTarget("container-1", "redis")
				_ = registry.Add(remote1)
				_ = registry.Add(remote2)
				_ = registry.Add(docker1)
				return registry
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				remotes := registry.ByType(target.TypeRemote)
				assert.Equal(t, 2, len(remotes))

				dockers := registry.ByType(target.TypeDocker)
				assert.Equal(t, 1, len(dockers))
			},
		},
		{
			name: "HealthSummary returns correct summary",
			setupFunc: func() *monitoring.Registry {
				registry := monitoring.NewRegistry()
				remote := target.NewRemoteTarget("remote-1", "localhost:8080", "tcp")
				docker := target.NewDockerTarget("container-1", "redis")
				_ = registry.Add(remote)
				_ = registry.Add(docker)
				return registry
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				summary := registry.HealthSummary()
				assert.Equal(t, 2, summary.Total)
				assert.Equal(t, 1, summary.ByType[target.TypeRemote])
				assert.Equal(t, 1, summary.ByType[target.TypeDocker])
				assert.Equal(t, 2, summary.UnknownCount())
			},
		},
		{
			name: "All returns all targets",
			setupFunc: func() *monitoring.Registry {
				registry := monitoring.NewRegistry()
				tgt1 := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt2 := target.NewRemoteTarget("test-2", "localhost:8081", "tcp")
				_ = registry.Add(tgt1)
				_ = registry.Add(tgt2)
				return registry
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				all := registry.All()
				assert.Equal(t, 2, len(all))
			},
		},
		{
			name: "AllStatuses returns all statuses",
			setupFunc: func() *monitoring.Registry {
				registry := monitoring.NewRegistry()
				tgt1 := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt2 := target.NewRemoteTarget("test-2", "localhost:8081", "tcp")
				_ = registry.Add(tgt1)
				_ = registry.Add(tgt2)
				return registry
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				statuses := registry.AllStatuses()
				assert.Equal(t, 2, len(statuses))
			},
		},
		{
			name: "Count returns correct count",
			setupFunc: func() *monitoring.Registry {
				registry := monitoring.NewRegistry()
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				_ = registry.Add(tgt)
				return registry
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				assert.Equal(t, 1, registry.Count())
			},
		},
		{
			name: "ByState filters targets by state",
			setupFunc: func() *monitoring.Registry {
				registry := monitoring.NewRegistry()
				tgt1 := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt2 := target.NewRemoteTarget("test-2", "localhost:8081", "tcp")
				_ = registry.Add(tgt1)
				_ = registry.Add(tgt2)
				_ = registry.UpdateStatus(tgt1.ID, func(status *target.Status) {
					status.State = target.StateHealthy
				})
				return registry
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				// No action needed
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				healthy := registry.ByState(target.StateHealthy)
				unknown := registry.ByState(target.StateUnknown)
				assert.Equal(t, 1, len(healthy))
				assert.Equal(t, 1, len(unknown))
			},
		},
		{
			name: "UpdateStatus updates status via callback",
			setupFunc: func() *monitoring.Registry {
				registry := monitoring.NewRegistry()
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				_ = registry.Add(tgt)
				return registry
			},
			actionFunc: func(t *testing.T, registry *monitoring.Registry) {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				err := registry.UpdateStatus(tgt.ID, func(status *target.Status) {
					status.State = target.StateHealthy
				})
				assert.NoError(t, err)

				// Verify update for non-existent target returns error.
				err = registry.UpdateStatus("non-existent", func(status *target.Status) {})
				assert.Error(t, err)
				assert.Equal(t, monitoring.ErrTargetNotFound, err)
			},
			verifyFunc: func(t *testing.T, registry *monitoring.Registry) {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				status := registry.GetStatus(tgt.ID)
				assert.NotNil(t, status)
				assert.Equal(t, target.StateHealthy, status.State)
			},
		},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			registry := tc.setupFunc()
			tc.actionFunc(t, registry)
			tc.verifyFunc(t, registry)
		})
	}
}

func TestHealthSummary_Counts(t *testing.T) {
	// testCase defines a test case for HealthSummary count methods.
	type testCase struct {
		name       string
		setupFunc  func() monitoring.HealthSummary
		verifyFunc func(*testing.T, monitoring.HealthSummary)
	}

	// tests defines all test cases for HealthSummary counts.
	tests := []testCase{
		{
			name: "count methods return correct values",
			setupFunc: func() monitoring.HealthSummary {
				summary := monitoring.NewHealthSummary()
				summary.Total = 10
				summary.ByState[target.StateHealthy] = 5
				summary.ByState[target.StateUnhealthy] = 3
				summary.ByState[target.StateUnknown] = 2
				return summary
			},
			verifyFunc: func(t *testing.T, summary monitoring.HealthSummary) {
				assert.Equal(t, 10, summary.Total)
				assert.Equal(t, 5, summary.HealthyCount())
				assert.Equal(t, 3, summary.UnhealthyCount())
				assert.Equal(t, 2, summary.UnknownCount())
			},
		},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			summary := tc.setupFunc()
			tc.verifyFunc(t, summary)
		})
	}
}

func TestNewRegistry(t *testing.T) {
	tests := []struct {
		name          string
		wantNotNil    bool
		wantInitCount int
	}{
		{
			name:          "creates non-nil registry",
			wantNotNil:    true,
			wantInitCount: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			if tt.wantNotNil {
				require.NotNil(t, registry)
			}
			assert.Equal(t, tt.wantInitCount, registry.Count())
		})
	}
}

func TestRegistry_Add(t *testing.T) {
	tests := []struct {
		name      string
		targets   []*target.ExternalTarget
		wantErr   []error
		wantCount int
	}{
		{
			name: "add single target successfully",
			targets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			},
			wantErr:   []error{nil},
			wantCount: 1,
		},
		{
			name: "add duplicate target returns error",
			targets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			},
			wantErr:   []error{nil, monitoring.ErrTargetExists},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			for i, tgt := range tt.targets {
				err := registry.Add(tgt)
				if tt.wantErr[i] != nil {
					assert.ErrorIs(t, err, tt.wantErr[i])
				} else {
					require.NoError(t, err)
				}
			}
			assert.Equal(t, tt.wantCount, registry.Count())
		})
	}
}

func TestRegistry_AddOrUpdate(t *testing.T) {
	tests := []struct {
		name         string
		setupTargets []*target.ExternalTarget
		updateTarget *target.ExternalTarget
		wantCount    int
		wantName     string
	}{
		{
			name:         "add new target",
			setupTargets: []*target.ExternalTarget{},
			updateTarget: func() *target.ExternalTarget {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt.Name = "original"
				return tgt
			}(),
			wantCount: 1,
			wantName:  "original",
		},
		{
			name: "update existing target",
			setupTargets: []*target.ExternalTarget{
				func() *target.ExternalTarget {
					tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
					tgt.Name = "original"
					return tgt
				}(),
			},
			updateTarget: func() *target.ExternalTarget {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt.Name = "updated"
				return tgt
			}(),
			wantCount: 1,
			wantName:  "updated",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			for _, tgt := range tt.setupTargets {
				registry.AddOrUpdate(tgt)
			}
			registry.AddOrUpdate(tt.updateTarget)
			assert.Equal(t, tt.wantCount, registry.Count())
			retrieved := registry.Get(tt.updateTarget.ID)
			require.NotNil(t, retrieved)
			assert.Equal(t, tt.wantName, retrieved.Name)
		})
	}
}

func TestRegistry_Remove(t *testing.T) {
	tests := []struct {
		name        string
		setupTarget *target.ExternalTarget
		removeID    string
		wantErr     error
		wantCount   int
	}{
		{
			name:        "remove existing target",
			setupTarget: target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			removeID:    target.NewRemoteTarget("test-1", "localhost:8080", "tcp").ID,
			wantErr:     nil,
			wantCount:   0,
		},
		{
			name:        "remove non-existent target",
			setupTarget: target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			removeID:    "non-existent",
			wantErr:     monitoring.ErrTargetNotFound,
			wantCount:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			if tt.setupTarget != nil {
				_ = registry.Add(tt.setupTarget)
			}
			err := registry.Remove(tt.removeID)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.wantCount, registry.Count())
		})
	}
}

func TestRegistry_Get(t *testing.T) {
	tests := []struct {
		name        string
		setupTarget *target.ExternalTarget
		getID       string
		wantNil     bool
		wantName    string
	}{
		{
			name: "get existing target",
			setupTarget: func() *target.ExternalTarget {
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				tgt.Name = "test-service"
				return tgt
			}(),
			getID:    target.NewRemoteTarget("test-1", "localhost:8080", "tcp").ID,
			wantNil:  false,
			wantName: "test-service",
		},
		{
			name:        "get non-existent target",
			setupTarget: target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			getID:       "non-existent",
			wantNil:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			if tt.setupTarget != nil {
				_ = registry.Add(tt.setupTarget)
			}
			retrieved := registry.Get(tt.getID)
			if tt.wantNil {
				assert.Nil(t, retrieved)
			} else {
				require.NotNil(t, retrieved)
				assert.Equal(t, tt.wantName, retrieved.Name)
			}
		})
	}
}

func TestRegistry_GetStatus(t *testing.T) {
	tests := []struct {
		name        string
		setupTarget *target.ExternalTarget
		getID       string
		wantNil     bool
		wantState   target.State
	}{
		{
			name:        "get status for existing target",
			setupTarget: target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			getID:       target.NewRemoteTarget("test-1", "localhost:8080", "tcp").ID,
			wantNil:     false,
			wantState:   target.StateUnknown,
		},
		{
			name:        "get status for non-existent target",
			setupTarget: target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			getID:       "non-existent",
			wantNil:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			if tt.setupTarget != nil {
				_ = registry.Add(tt.setupTarget)
			}
			status := registry.GetStatus(tt.getID)
			if tt.wantNil {
				assert.Nil(t, status)
			} else {
				require.NotNil(t, status)
				assert.Equal(t, tt.wantState, status.State)
			}
		})
	}
}

func TestRegistry_All(t *testing.T) {
	tests := []struct {
		name         string
		setupTargets []*target.ExternalTarget
		wantCount    int
	}{
		{
			name:         "empty registry",
			setupTargets: []*target.ExternalTarget{},
			wantCount:    0,
		},
		{
			name: "multiple targets",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("test-2", "localhost:8081", "tcp"),
			},
			wantCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			for _, tgt := range tt.setupTargets {
				_ = registry.Add(tgt)
			}
			all := registry.All()
			assert.Equal(t, tt.wantCount, len(all))
		})
	}
}

func TestRegistry_AllStatuses(t *testing.T) {
	tests := []struct {
		name         string
		setupTargets []*target.ExternalTarget
		wantCount    int
	}{
		{
			name:         "empty registry",
			setupTargets: []*target.ExternalTarget{},
			wantCount:    0,
		},
		{
			name: "multiple targets",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("test-2", "localhost:8081", "tcp"),
			},
			wantCount: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			for _, tgt := range tt.setupTargets {
				_ = registry.Add(tgt)
			}
			statuses := registry.AllStatuses()
			assert.Equal(t, tt.wantCount, len(statuses))
		})
	}
}

func TestRegistry_Count(t *testing.T) {
	tests := []struct {
		name         string
		setupTargets []*target.ExternalTarget
		wantCount    int
	}{
		{
			name:         "empty registry",
			setupTargets: []*target.ExternalTarget{},
			wantCount:    0,
		},
		{
			name: "single target",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			},
			wantCount: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			for _, tgt := range tt.setupTargets {
				_ = registry.Add(tgt)
			}
			assert.Equal(t, tt.wantCount, registry.Count())
		})
	}
}

func TestRegistry_ByType(t *testing.T) {
	tests := []struct {
		name         string
		setupTargets []*target.ExternalTarget
		filterType   target.Type
		wantCount    int
	}{
		{
			name: "filter remote targets",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("remote-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("remote-2", "localhost:8081", "tcp"),
				target.NewDockerTarget("container-1", "redis"),
			},
			filterType: target.TypeRemote,
			wantCount:  2,
		},
		{
			name: "filter docker targets",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("remote-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("remote-2", "localhost:8081", "tcp"),
				target.NewDockerTarget("container-1", "redis"),
			},
			filterType: target.TypeDocker,
			wantCount:  1,
		},
		{
			name: "filter non-existent type",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("remote-1", "localhost:8080", "tcp"),
				target.NewDockerTarget("container-1", "redis"),
			},
			filterType: target.TypeSystemd,
			wantCount:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			for _, tgt := range tt.setupTargets {
				_ = registry.Add(tgt)
			}
			filtered := registry.ByType(tt.filterType)
			assert.Equal(t, tt.wantCount, len(filtered))
		})
	}
}

func TestRegistry_ByState(t *testing.T) {
	tests := []struct {
		name         string
		setupTargets []*target.ExternalTarget
		updateIDs    []string
		updateState  target.State
		filterState  target.State
		wantCount    int
	}{
		{
			name: "filter healthy targets",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("test-2", "localhost:8081", "tcp"),
			},
			updateIDs:   []string{target.NewRemoteTarget("test-1", "localhost:8080", "tcp").ID},
			updateState: target.StateHealthy,
			filterState: target.StateHealthy,
			wantCount:   1,
		},
		{
			name: "filter unknown targets",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("test-2", "localhost:8081", "tcp"),
			},
			updateIDs:   []string{target.NewRemoteTarget("test-1", "localhost:8080", "tcp").ID},
			updateState: target.StateHealthy,
			filterState: target.StateUnknown,
			wantCount:   1,
		},
		{
			name: "filter unhealthy targets",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
				target.NewRemoteTarget("test-2", "localhost:8081", "tcp"),
			},
			updateIDs:   []string{},
			updateState: target.StateHealthy,
			filterState: target.StateUnhealthy,
			wantCount:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			for _, tgt := range tt.setupTargets {
				_ = registry.Add(tgt)
			}
			for _, id := range tt.updateIDs {
				_ = registry.UpdateStatus(id, func(status *target.Status) {
					status.State = tt.updateState
				})
			}
			filtered := registry.ByState(tt.filterState)
			assert.Equal(t, tt.wantCount, len(filtered))
		})
	}
}

func TestRegistry_UpdateStatus(t *testing.T) {
	tests := []struct {
		name        string
		setupTarget *target.ExternalTarget
		updateID    string
		updateState target.State
		wantErr     error
		wantState   target.State
	}{
		{
			name:        "update existing target status",
			setupTarget: target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			updateID:    target.NewRemoteTarget("test-1", "localhost:8080", "tcp").ID,
			updateState: target.StateHealthy,
			wantErr:     nil,
			wantState:   target.StateHealthy,
		},
		{
			name:        "update non-existent target",
			setupTarget: target.NewRemoteTarget("test-1", "localhost:8080", "tcp"),
			updateID:    "non-existent",
			updateState: target.StateHealthy,
			wantErr:     monitoring.ErrTargetNotFound,
			wantState:   target.StateUnknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			if tt.setupTarget != nil {
				_ = registry.Add(tt.setupTarget)
			}
			err := registry.UpdateStatus(tt.updateID, func(status *target.Status) {
				status.State = tt.updateState
			})
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				status := registry.GetStatus(tt.updateID)
				assert.NotNil(t, status)
				assert.Equal(t, tt.wantState, status.State)
			}
		})
	}
}

func TestRegistry_HealthSummary(t *testing.T) {
	tests := []struct {
		name         string
		setupTargets []*target.ExternalTarget
		wantTotal    int
		wantRemote   int
		wantDocker   int
		wantUnknown  int
	}{
		{
			name:         "empty registry",
			setupTargets: []*target.ExternalTarget{},
			wantTotal:    0,
			wantRemote:   0,
			wantDocker:   0,
			wantUnknown:  0,
		},
		{
			name: "mixed target types",
			setupTargets: []*target.ExternalTarget{
				target.NewRemoteTarget("remote-1", "localhost:8080", "tcp"),
				target.NewDockerTarget("container-1", "redis"),
			},
			wantTotal:   2,
			wantRemote:  1,
			wantDocker:  1,
			wantUnknown: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := monitoring.NewRegistry()
			for _, tgt := range tt.setupTargets {
				_ = registry.Add(tgt)
			}
			summary := registry.HealthSummary()
			assert.Equal(t, tt.wantTotal, summary.Total)
			assert.Equal(t, tt.wantRemote, summary.ByType[target.TypeRemote])
			assert.Equal(t, tt.wantDocker, summary.ByType[target.TypeDocker])
			assert.Equal(t, tt.wantUnknown, summary.UnknownCount())
		})
	}
}

func TestHealthSummary_HealthyCount(t *testing.T) {
	tests := []struct {
		name         string
		healthyVal   int
		unhealthyVal int
		unknownVal   int
		wantHealthy  int
	}{
		{
			name:         "with healthy targets",
			healthyVal:   5,
			unhealthyVal: 3,
			unknownVal:   2,
			wantHealthy:  5,
		},
		{
			name:         "no healthy targets",
			healthyVal:   0,
			unhealthyVal: 5,
			unknownVal:   3,
			wantHealthy:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := monitoring.NewHealthSummary()
			summary.Total = tt.healthyVal + tt.unhealthyVal + tt.unknownVal
			summary.ByState[target.StateHealthy] = tt.healthyVal
			summary.ByState[target.StateUnhealthy] = tt.unhealthyVal
			summary.ByState[target.StateUnknown] = tt.unknownVal
			assert.Equal(t, tt.wantHealthy, summary.HealthyCount())
		})
	}
}

func TestHealthSummary_UnhealthyCount(t *testing.T) {
	tests := []struct {
		name          string
		healthyVal    int
		unhealthyVal  int
		unknownVal    int
		wantUnhealthy int
	}{
		{
			name:          "with unhealthy targets",
			healthyVal:    5,
			unhealthyVal:  3,
			unknownVal:    2,
			wantUnhealthy: 3,
		},
		{
			name:          "no unhealthy targets",
			healthyVal:    5,
			unhealthyVal:  0,
			unknownVal:    3,
			wantUnhealthy: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := monitoring.NewHealthSummary()
			summary.Total = tt.healthyVal + tt.unhealthyVal + tt.unknownVal
			summary.ByState[target.StateHealthy] = tt.healthyVal
			summary.ByState[target.StateUnhealthy] = tt.unhealthyVal
			summary.ByState[target.StateUnknown] = tt.unknownVal
			assert.Equal(t, tt.wantUnhealthy, summary.UnhealthyCount())
		})
	}
}

func TestHealthSummary_UnknownCount(t *testing.T) {
	tests := []struct {
		name         string
		healthyVal   int
		unhealthyVal int
		unknownVal   int
		wantUnknown  int
	}{
		{
			name:         "with unknown targets",
			healthyVal:   5,
			unhealthyVal: 3,
			unknownVal:   2,
			wantUnknown:  2,
		},
		{
			name:         "no unknown targets",
			healthyVal:   5,
			unhealthyVal: 3,
			unknownVal:   0,
			wantUnknown:  0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := monitoring.NewHealthSummary()
			summary.Total = tt.healthyVal + tt.unhealthyVal + tt.unknownVal
			summary.ByState[target.StateHealthy] = tt.healthyVal
			summary.ByState[target.StateUnhealthy] = tt.unhealthyVal
			summary.ByState[target.StateUnknown] = tt.unknownVal
			assert.Equal(t, tt.wantUnknown, summary.UnknownCount())
		})
	}
}
