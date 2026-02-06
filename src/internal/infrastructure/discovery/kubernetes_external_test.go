//go:build unix

// Package discovery_test provides external tests for Kubernetes discoverer.
package discovery_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testKubeconfigDiscovererValid is a valid kubeconfig for discoverer creation tests.
const testKubeconfigDiscovererValid string = `
current-context: default
clusters:
- name: default-cluster
  cluster:
    server: https://localhost:6443
contexts:
- name: default
  context:
    cluster: default-cluster
    user: default-user
users:
- name: default-user
  user:
    token: test-token
`

// TestNewKubernetesDiscoverer verifies discoverer creation.
func TestNewKubernetesDiscoverer(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T) string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "fails without kubeconfig",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/kubeconfig"
			},
			wantErr:    true,
			wantErrMsg: "load kubernetes auth",
		},
		{
			name: "succeeds with valid kubeconfig",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
				err := os.WriteFile(kubeconfigPath, []byte(testKubeconfigDiscovererValid), 0600)
				require.NoError(t, err)
				return kubeconfigPath
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeconfigPath := tt.setupFunc(t)

			cfg := &config.KubernetesDiscoveryConfig{
				Enabled:        true,
				KubeconfigPath: kubeconfigPath,
				Namespaces:     []string{"default"},
			}

			discoverer, err := discovery.NewKubernetesDiscoverer(cfg)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, discoverer)
			}
		})
	}
}

// TestKubernetesDiscoverer_Type verifies the Type method returns the correct target type.
//
// Params:
//   - t: testing context for assertions.
func TestKubernetesDiscoverer_Type(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		expectedType target.Type
	}{
		{
			name:         "returns kubernetes type",
			expectedType: target.TypeKubernetes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a temporary kubeconfig for the discoverer.
			tmpDir := t.TempDir()
			kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
			err := os.WriteFile(kubeconfigPath, []byte(testKubeconfigDiscovererValid), 0600)
			require.NoError(t, err)

			cfg := &config.KubernetesDiscoveryConfig{
				Enabled:        true,
				KubeconfigPath: kubeconfigPath,
				Namespaces:     []string{"default"},
			}

			discoverer, err := discovery.NewKubernetesDiscoverer(cfg)
			require.NoError(t, err)

			// Verify Type returns correct value.
			assert.Equal(t, tt.expectedType, discoverer.Type())
		})
	}
}

// TestKubernetesDiscoverer_Discover verifies the Discover method behavior.
//
// Params:
//   - t: testing context for assertions.
func TestKubernetesDiscoverer_Discover(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "returns error when kubernetes api unreachable",
			wantErr:    true,
			wantErrMsg: "discover namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a temporary kubeconfig for the discoverer.
			tmpDir := t.TempDir()
			kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
			err := os.WriteFile(kubeconfigPath, []byte(testKubeconfigDiscovererValid), 0600)
			require.NoError(t, err)

			cfg := &config.KubernetesDiscoveryConfig{
				Enabled:        true,
				KubeconfigPath: kubeconfigPath,
				Namespaces:     []string{"default"},
			}

			discoverer, err := discovery.NewKubernetesDiscoverer(cfg)
			require.NoError(t, err)

			// Call Discover with timeout context.
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			targets, err := discoverer.Discover(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, targets)
			}
		})
	}
}
