package target_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
)

func TestNewDiscoveryConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want func(*testing.T, target.DiscoveryConfig)
	}{
		{
			name: "creates config with defaults",
			want: func(t *testing.T, cfg target.DiscoveryConfig) {
				assert.False(t, cfg.Enabled, "should be disabled by default")
				assert.NotNil(t, cfg.Labels, "labels map should be initialized")
				assert.Nil(t, cfg.Patterns, "patterns should be nil")
				assert.Nil(t, cfg.Namespaces, "namespaces should be nil")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := target.NewDiscoveryConfig()
			tt.want(t, cfg)
		})
	}
}

func TestDiscoveryConfig_WithEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		enabled bool
	}{
		{"enable discovery", true},
		{"disable discovery", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := target.NewDiscoveryConfig().WithEnabled(tt.enabled)
			assert.Equal(t, tt.enabled, cfg.Enabled)
		})
	}
}

func TestDiscoveryConfig_WithPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		patterns []string
	}{
		{"single pattern", []string{"*.service"}},
		{"multiple patterns", []string{"nginx.*", "*.target"}},
		{"empty patterns", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := target.NewDiscoveryConfig().WithPatterns(tt.patterns...)
			assert.Equal(t, tt.patterns, cfg.Patterns)
		})
	}
}

func TestDiscoveryConfig_WithNamespaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		namespaces []string
	}{
		{"single namespace", []string{"default"}},
		{"multiple namespaces", []string{"default", "kube-system"}},
		{"empty namespaces", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := target.NewDiscoveryConfig().WithNamespaces(tt.namespaces...)
			assert.Equal(t, tt.namespaces, cfg.Namespaces)
		})
	}
}

func TestDiscoveryConfig_WithSocketPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{"docker socket", "/var/run/docker.sock"},
		{"podman socket", "/run/podman/podman.sock"},
		{"custom path", "/custom/socket.sock"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := target.NewDiscoveryConfig().WithSocketPath(tt.path)
			assert.Equal(t, tt.path, cfg.SocketPath)
		})
	}
}

func TestDiscoveryConfig(t *testing.T) {
	t.Parallel()

	// testCase defines a test case for DiscoveryConfig fluent API.
	type testCase struct {
		name       string
		setupFunc  func() target.DiscoveryConfig
		verifyFunc func(*testing.T, target.DiscoveryConfig)
	}

	// tests defines all test cases for DiscoveryConfig.
	tests := []testCase{
		{
			name: "fluent API chains all setters",
			setupFunc: func() target.DiscoveryConfig {
				return target.NewDiscoveryConfig().
					WithEnabled(true).
					WithPatterns("*.service").
					WithNamespaces("default").
					WithSocketPath("/var/run/docker.sock")
			},
			verifyFunc: func(t *testing.T, cfg target.DiscoveryConfig) {
				assert.True(t, cfg.Enabled)
				assert.Equal(t, []string{"*.service"}, cfg.Patterns)
				assert.Equal(t, []string{"default"}, cfg.Namespaces)
				assert.Equal(t, "/var/run/docker.sock", cfg.SocketPath)
			},
		},
	}

	// run all test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := tc.setupFunc()
			tc.verifyFunc(t, cfg)
		})
	}
}
