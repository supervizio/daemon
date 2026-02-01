package target_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

func TestNewExternalTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		targetName string
		targetType target.Type
		source     target.Source
	}{
		{
			name:       "docker target",
			id:         "docker:redis",
			targetName: "redis",
			targetType: target.TypeDocker,
			source:     target.SourceDiscovered,
		},
		{
			name:       "systemd target",
			id:         "systemd:nginx.service",
			targetName: "nginx.service",
			targetType: target.TypeSystemd,
			source:     target.SourceStatic,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tgt := target.NewExternalTarget(tt.id, tt.targetName, tt.targetType, tt.source)

			assert.NotNil(t, tgt)
			assert.Equal(t, tt.id, tgt.ID)
			assert.Equal(t, tt.targetName, tgt.Name)
			assert.Equal(t, tt.targetType, tgt.Type)
			assert.Equal(t, tt.source, tgt.Source)
			assert.NotNil(t, tgt.Labels)
			assert.Greater(t, tgt.Interval, time.Duration(0))
			assert.Greater(t, tgt.Timeout, time.Duration(0))
			assert.Greater(t, tgt.SuccessThreshold, 0)
			assert.Greater(t, tgt.FailureThreshold, 0)
		})
	}
}

func TestTargetFactories(t *testing.T) {
	t.Parallel()

	// testCase defines a test case for target factory functions.
	type testCase struct {
		name       string
		createFunc func() *target.ExternalTarget
		verifyFunc func(*testing.T, *target.ExternalTarget)
	}

	// tests defines all test cases for target factories.
	tests := []testCase{
		{
			name: "NewRemoteTarget creates remote target",
			createFunc: func() *target.ExternalTarget {
				return target.NewRemoteTarget("web-server", "example.com:80", "tcp")
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, "remote:web-server", tgt.ID)
				assert.Equal(t, "web-server", tgt.Name)
				assert.Equal(t, target.TypeRemote, tgt.Type)
				assert.Equal(t, target.SourceStatic, tgt.Source)
				assert.Equal(t, "tcp", tgt.ProbeType)
				assert.NotNil(t, tgt.ProbeTarget)
			},
		},
		{
			name: "NewDockerTarget creates docker target",
			createFunc: func() *target.ExternalTarget {
				return target.NewDockerTarget("abc123", "redis-container")
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, "docker:abc123", tgt.ID)
				assert.Equal(t, "redis-container", tgt.Name)
				assert.Equal(t, target.TypeDocker, tgt.Type)
				assert.Equal(t, target.SourceDiscovered, tgt.Source)
			},
		},
		{
			name: "NewSystemdTarget creates systemd target",
			createFunc: func() *target.ExternalTarget {
				return target.NewSystemdTarget("nginx.service")
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, "systemd:nginx.service", tgt.ID)
				assert.Equal(t, "nginx.service", tgt.Name)
				assert.Equal(t, target.TypeSystemd, tgt.Type)
				assert.Equal(t, target.SourceDiscovered, tgt.Source)
				assert.Equal(t, "exec", tgt.ProbeType)
				assert.NotNil(t, tgt.ProbeTarget)
			},
		},
		{
			name: "NewKubernetesTarget creates kubernetes target",
			createFunc: func() *target.ExternalTarget {
				return target.NewKubernetesTarget("default", "pod", "nginx-xyz")
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, "kubernetes:default/pod/nginx-xyz", tgt.ID)
				assert.Equal(t, "nginx-xyz", tgt.Name)
				assert.Equal(t, target.TypeKubernetes, tgt.Type)
				assert.Equal(t, target.SourceDiscovered, tgt.Source)
				assert.Equal(t, "default", tgt.Labels["namespace"])
				assert.Equal(t, "pod", tgt.Labels["resource_type"])
			},
		},
		{
			name: "NewNomadTarget creates nomad target",
			createFunc: func() *target.ExternalTarget {
				return target.NewNomadTarget("alloc123", "web", "my-job")
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, "nomad:alloc123/web", tgt.ID)
				assert.Equal(t, "my-job/web", tgt.Name)
				assert.Equal(t, target.TypeNomad, tgt.Type)
				assert.Equal(t, target.SourceDiscovered, tgt.Source)
				assert.Equal(t, "alloc123", tgt.Labels["alloc_id"])
				assert.Equal(t, "web", tgt.Labels["task"])
				assert.Equal(t, "my-job", tgt.Labels["job"])
			},
		},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tgt := tc.createFunc()
			tc.verifyFunc(t, tgt)
		})
	}
}

func TestExternalTarget_Methods(t *testing.T) {
	t.Parallel()

	// testCase defines a test case for ExternalTarget methods.
	type testCase struct {
		name       string
		setupFunc  func() *target.ExternalTarget
		verifyFunc func(*testing.T, *target.ExternalTarget)
	}

	// tests defines all test cases for ExternalTarget methods.
	tests := []testCase{
		{
			name: "WithProbe sets probe and returns self",
			setupFunc: func() *target.ExternalTarget {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				probeTarget := health.NewTarget("http", "localhost:8080")
				return tgt.WithProbe("http", probeTarget)
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, "http", tgt.ProbeType)
				assert.NotNil(t, tgt.ProbeTarget)
			},
		},
		{
			name: "WithTiming sets timing and returns self",
			setupFunc: func() *target.ExternalTarget {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				return tgt.WithTiming(60*time.Second, 10*time.Second)
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, 60*time.Second, tgt.Interval)
				assert.Equal(t, 10*time.Second, tgt.Timeout)
			},
		},
		{
			name: "WithThresholds sets thresholds and returns self",
			setupFunc: func() *target.ExternalTarget {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				return tgt.WithThresholds(2, 5)
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, 2, tgt.SuccessThreshold)
				assert.Equal(t, 5, tgt.FailureThreshold)
			},
		},
		{
			name: "WithLabel sets label and returns self",
			setupFunc: func() *target.ExternalTarget {
				tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
				return tgt.WithLabel("env", "production")
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, "production", tgt.Labels["env"])
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

func TestExternalTarget_HasProbe(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		probeType string
		want      bool
	}{
		{"with probe", "http", true},
		{"without probe", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
			tgt.ProbeType = tt.probeType

			assert.Equal(t, tt.want, tgt.HasProbe())
		})
	}
}

func TestExternalTarget_IsHealthCheckable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setup      func(*target.ExternalTarget)
		want       bool
		wantReason string
	}{
		{
			name: "fully configured",
			setup: func(tgt *target.ExternalTarget) {
				tgt.ProbeType = "http"
				tgt.Interval = 30 * time.Second
				tgt.Timeout = 5 * time.Second
			},
			want:       true,
			wantReason: "should be checkable",
		},
		{
			name: "no probe type",
			setup: func(tgt *target.ExternalTarget) {
				tgt.ProbeType = ""
			},
			want:       false,
			wantReason: "should not be checkable without probe",
		},
		{
			name: "zero interval",
			setup: func(tgt *target.ExternalTarget) {
				tgt.ProbeType = "http"
				tgt.Interval = 0
			},
			want:       false,
			wantReason: "should not be checkable with zero interval",
		},
		{
			name: "zero timeout",
			setup: func(tgt *target.ExternalTarget) {
				tgt.ProbeType = "http"
				tgt.Timeout = 0
			},
			want:       false,
			wantReason: "should not be checkable with zero timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tgt := target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic)
			tt.setup(tgt)

			assert.Equal(t, tt.want, tgt.IsHealthCheckable(), tt.wantReason)
		})
	}
}

func TestExternalTarget(t *testing.T) {
	t.Parallel()

	// testCase defines a test case for ExternalTarget fluent API.
	type testCase struct {
		name       string
		setupFunc  func() *target.ExternalTarget
		verifyFunc func(*testing.T, *target.ExternalTarget)
	}

	// tests defines all test cases for ExternalTarget.
	tests := []testCase{
		{
			name: "fluent API chains all setters",
			setupFunc: func() *target.ExternalTarget {
				probeTarget := health.NewTarget("http", "localhost:8080")
				return target.NewExternalTarget("test:1", "test", target.TypeDocker, target.SourceStatic).
					WithProbe("http", probeTarget).
					WithTiming(60*time.Second, 10*time.Second).
					WithThresholds(2, 5).
					WithLabel("env", "production").
					WithLabel("team", "platform")
			},
			verifyFunc: func(t *testing.T, tgt *target.ExternalTarget) {
				assert.Equal(t, "http", tgt.ProbeType)
				assert.NotNil(t, tgt.ProbeTarget)
				assert.Equal(t, 60*time.Second, tgt.Interval)
				assert.Equal(t, 10*time.Second, tgt.Timeout)
				assert.Equal(t, 2, tgt.SuccessThreshold)
				assert.Equal(t, 5, tgt.FailureThreshold)
				assert.Equal(t, "production", tgt.Labels["env"])
				assert.Equal(t, "platform", tgt.Labels["team"])
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
