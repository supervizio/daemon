package target_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	t.Parallel()

	// testCase defines a test case for target constants and defaults.
	type testCase struct {
		name       string
		setupFunc  func() *target.ExternalTarget
		verifyFunc func(*testing.T, *target.ExternalTarget)
	}

	// tests defines all test cases for target constants.
	tests := []testCase{
		{
			name: "default values are correctly applied",
			setupFunc: func() *target.ExternalTarget {
				return target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, 30*time.Second, tgt.Interval, "default interval should be 30s")
				assert.Equal(t, 5*time.Second, tgt.Timeout, "default timeout should be 5s")
				assert.Equal(t, 1, tgt.SuccessThreshold, "default success threshold should be 1")
				assert.Equal(t, 3, tgt.FailureThreshold, "default failure threshold should be 3")
				assert.NotNil(t, tgt.Labels, "labels map should be initialized")
			},
		},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tgt := tc.setupFunc()
			tc.verifyFunc(t, tgt)
		})
	}
}
