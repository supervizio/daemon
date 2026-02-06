package monitoring

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_internal(t *testing.T) {
	// testCase defines a test case for Registry internal state.
	type testCase struct {
		name       string
		setupFunc  func() *Registry
		verifyFunc func(*testing.T, *Registry)
	}

	// tests defines all test cases for Registry internal.
	tests := []testCase{
		{
			name: "internal state is initialized correctly",
			setupFunc: func() *Registry {
				return NewRegistry()
			},
			verifyFunc: func(t *testing.T, registry *Registry) {
				assert.NotNil(t, registry.targets)
				assert.NotNil(t, registry.statuses)
				assert.Equal(t, 0, len(registry.targets))
				assert.Equal(t, 0, len(registry.statuses))
			},
		},
		{
			name: "UpdateStatus updates status via callback",
			setupFunc: func() *Registry {
				registry := NewRegistry()
				tgt := target.NewRemoteTarget("test-1", "localhost:8080", "tcp")
				_ = registry.Add(tgt)
				_ = registry.UpdateStatus(tgt.ID, func(status *target.Status) {
					status.State = target.StateHealthy
				})
				return registry
			},
			verifyFunc: func(t *testing.T, registry *Registry) {
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
			tc.verifyFunc(t, registry)
		})
	}
}
