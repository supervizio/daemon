//go:build unix

package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test kubeconfig constants for kubernetes_auth tests.
const (
	// testKubeconfigWithFirstContext uses first context when current-context is not set.
	testKubeconfigWithFirstContext string = `
clusters:
- name: first-cluster
  cluster:
    server: https://first.cluster:6443
    certificate-authority-data: first-ca
contexts:
- name: first-context
  context:
    cluster: first-cluster
    user: first-user
users:
- name: first-user
  user:
    token: first-token
`

	// testKubeconfigInvalidYAML is malformed YAML.
	testKubeconfigInvalidYAML string = `
clusters:
- name: test
  cluster: [invalid yaml structure
`

	// testKubeconfigExplicitPath is a complete kubeconfig for explicit path tests.
	testKubeconfigExplicitPath string = `
current-context: default
clusters:
- name: default-cluster
  cluster:
    server: https://explicit.example.com:6443
contexts:
- name: default
  context:
    cluster: default-cluster
    user: default-user
users:
- name: default-user
  user:
    token: explicit-token
`
)

// TestParseKubeconfigFile verifies kubeconfig file parsing.
func TestParseKubeconfigFile(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantErr    bool
		wantClusts int
	}{
		{
			name:       "valid kubeconfig",
			content:    testKubeconfigWithFirstContext,
			wantErr:    false,
			wantClusts: 1,
		},
		{
			name:       "invalid YAML",
			content:    testKubeconfigInvalidYAML,
			wantErr:    true,
			wantClusts: 0,
		},
		{
			name:       "empty file",
			content:    "",
			wantErr:    false,
			wantClusts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
			err := os.WriteFile(kubeconfigPath, []byte(tt.content), 0600)
			require.NoError(t, err)

			cfg, err := parseKubeconfigFile(kubeconfigPath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, cfg.Clusters, tt.wantClusts)
			}
		})
	}
}

// TestParseKubeconfigFile_NonExistentFile verifies error on missing file.
func TestParseKubeconfigFile_NonExistentFile(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "non-existent file",
			path:    "/nonexistent/path/to/kubeconfig",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseKubeconfigFile(tt.path)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "read kubeconfig")
			}
		})
	}
}

// TestKubeconfig_FindCurrentContextName verifies context name resolution.
func TestKubeconfig_FindCurrentContextName(t *testing.T) {
	tests := []struct {
		name        string
		kubeconfig  kubeconfig
		wantContext string
		wantErr     bool
	}{
		{
			name: "explicit current context",
			kubeconfig: kubeconfig{
				CurrentContext: "my-context",
				Contexts: []struct {
					Name    string            `dto:"in,priv,pub" yaml:"name"`
					Context kubeconfigContext `dto:"in,priv,pub" yaml:"context"`
				}{
					{Name: "my-context"},
					{Name: "other-context"},
				},
			},
			wantContext: "my-context",
			wantErr:     false,
		},
		{
			name: "fallback to first context",
			kubeconfig: kubeconfig{
				CurrentContext: "",
				Contexts: []struct {
					Name    string            `dto:"in,priv,pub" yaml:"name"`
					Context kubeconfigContext `dto:"in,priv,pub" yaml:"context"`
				}{
					{Name: "first-context"},
					{Name: "second-context"},
				},
			},
			wantContext: "first-context",
			wantErr:     false,
		},
		{
			name: "no contexts available",
			kubeconfig: kubeconfig{
				CurrentContext: "",
				Contexts:       nil,
			},
			wantContext: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.kubeconfig.findCurrentContextName()

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errNoContext)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantContext, got)
			}
		})
	}
}

// TestKubeconfig_FindContext verifies context lookup.
func TestKubeconfig_FindContext(t *testing.T) {
	tests := []struct {
		name        string
		kubeconfig  kubeconfig
		contextName string
		wantCluster string
		wantErr     bool
	}{
		{
			name: "context found",
			kubeconfig: kubeconfig{
				Contexts: []struct {
					Name    string            `dto:"in,priv,pub" yaml:"name"`
					Context kubeconfigContext `dto:"in,priv,pub" yaml:"context"`
				}{
					{Name: "test-ctx", Context: kubeconfigContext{Cluster: "test-cluster", User: "test-user"}},
				},
			},
			contextName: "test-ctx",
			wantCluster: "test-cluster",
			wantErr:     false,
		},
		{
			name: "context not found",
			kubeconfig: kubeconfig{
				Contexts: []struct {
					Name    string            `dto:"in,priv,pub" yaml:"name"`
					Context kubeconfigContext `dto:"in,priv,pub" yaml:"context"`
				}{
					{Name: "other-ctx"},
				},
			},
			contextName: "missing-ctx",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.kubeconfig.findContext(tt.contextName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errContextNotFound)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCluster, got.Cluster)
			}
		})
	}
}

// TestKubeconfig_FindCluster verifies cluster lookup.
func TestKubeconfig_FindCluster(t *testing.T) {
	tests := []struct {
		name        string
		kubeconfig  kubeconfig
		clusterName string
		wantServer  string
		wantErr     bool
	}{
		{
			name: "cluster found",
			kubeconfig: kubeconfig{
				Clusters: []struct {
					Name    string            `dto:"in,priv,pub" yaml:"name"`
					Cluster kubeconfigCluster `dto:"in,priv,priv" yaml:"cluster"`
				}{
					{Name: "my-cluster", Cluster: kubeconfigCluster{Server: "https://api.example.com"}},
				},
			},
			clusterName: "my-cluster",
			wantServer:  "https://api.example.com",
			wantErr:     false,
		},
		{
			name: "cluster not found",
			kubeconfig: kubeconfig{
				Clusters: []struct {
					Name    string            `dto:"in,priv,pub" yaml:"name"`
					Cluster kubeconfigCluster `dto:"in,priv,priv" yaml:"cluster"`
				}{
					{Name: "other-cluster"},
				},
			},
			clusterName: "missing-cluster",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.kubeconfig.findCluster(tt.clusterName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errClusterNotFound)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantServer, got.Server)
			}
		})
	}
}

// TestKubeconfig_FindUser verifies user lookup.
func TestKubeconfig_FindUser(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig kubeconfig
		userName   string
		wantToken  string
		wantErr    bool
	}{
		{
			name: "user found",
			kubeconfig: kubeconfig{
				Users: []struct {
					Name string         `dto:"in,priv,pub" yaml:"name"`
					User kubeconfigUser `dto:"in,priv,secret" yaml:"user"`
				}{
					{Name: "my-user", User: kubeconfigUser{Token: "secret-token"}},
				},
			},
			userName:  "my-user",
			wantToken: "secret-token",
			wantErr:   false,
		},
		{
			name: "user not found",
			kubeconfig: kubeconfig{
				Users: []struct {
					Name string         `dto:"in,priv,pub" yaml:"name"`
					User kubeconfigUser `dto:"in,priv,secret" yaml:"user"`
				}{
					{Name: "other-user"},
				},
			},
			userName: "missing-user",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.kubeconfig.findUser(tt.userName)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errUserNotFound)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken, got.Token)
			}
		})
	}
}

// TestBuildK8sAuth verifies k8sAuth construction from cluster and user.
func TestBuildK8sAuth(t *testing.T) {
	tests := []struct {
		name       string
		cluster    *kubeconfigCluster
		user       *kubeconfigUser
		wantServer string
		wantToken  string
		wantCACert bool
	}{
		{
			name: "builds auth with CA cert",
			cluster: &kubeconfigCluster{
				Server:                   "https://api.example.com:6443",
				CertificateAuthorityData: "test-ca-data",
			},
			user: &kubeconfigUser{
				Token: "test-token",
			},
			wantServer: "https://api.example.com:6443",
			wantToken:  "test-token",
			wantCACert: true,
		},
		{
			name: "builds auth without CA cert",
			cluster: &kubeconfigCluster{
				Server:                   "https://localhost:6443",
				CertificateAuthorityData: "",
			},
			user: &kubeconfigUser{
				Token: "local-token",
			},
			wantServer: "https://localhost:6443",
			wantToken:  "local-token",
			wantCACert: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := buildK8sAuth(tt.cluster, tt.user)

			assert.Equal(t, tt.wantServer, auth.apiServer)
			assert.Equal(t, tt.wantToken, auth.token)
			if tt.wantCACert {
				assert.NotEmpty(t, auth.caCert)
			} else {
				assert.Empty(t, auth.caCert)
			}
		})
	}
}

// TestNewHTTPClient_InvalidCACert verifies error on invalid CA certificate.
func TestNewHTTPClient_InvalidCACert(t *testing.T) {
	tests := []struct {
		name    string
		auth    *k8sAuth
		wantErr bool
	}{
		{
			name: "invalid CA cert returns error",
			auth: &k8sAuth{
				apiServer: "https://localhost:6443",
				token:     "test",
				caCert:    []byte("not-a-valid-pem-certificate"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newHTTPClient(tt.auth)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errAppendCACert)
			}
		})
	}
}

// TestLoadInClusterConfig verifies in-cluster configuration loading.
func TestLoadInClusterConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T) (cleanup func())
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "fails when token file missing",
			setupFunc: func(t *testing.T) func() {
				// Don't create any files.
				return func() {}
			},
			wantErr:    true,
			wantErrMsg: "read service account token",
		},
		{
			name: "fails when env vars missing",
			setupFunc: func(t *testing.T) func() {
				tmpDir := t.TempDir()
				tokenPath := filepath.Join(tmpDir, "token")
				caPath := filepath.Join(tmpDir, "ca.crt")
				err := os.WriteFile(tokenPath, []byte("test-token"), 0600)
				require.NoError(t, err)
				err = os.WriteFile(caPath, []byte("test-ca"), 0600)
				require.NoError(t, err)
				// Note: We cannot override the hardcoded paths in loadInClusterConfig.
				// This test verifies the error path when files don't exist.
				return func() {}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setupFunc(t)
			defer cleanup()

			_, err := loadInClusterConfig()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			}
		})
	}
}

// TestLoadKubeconfigOrInCluster verifies kubeconfig fallback logic.
func TestLoadKubeconfigOrInCluster(t *testing.T) {
	tests := []struct {
		name           string
		kubeconfigPath string
		setupFunc      func(t *testing.T) string
		wantErr        bool
		wantErrMsg     string
	}{
		{
			name: "loads explicit kubeconfig path",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
				err := os.WriteFile(kubeconfigPath, []byte(testKubeconfigExplicitPath), 0600)
				require.NoError(t, err)
				return kubeconfigPath
			},
			wantErr: false,
		},
		{
			name: "fails on non-existent explicit path",
			setupFunc: func(t *testing.T) string {
				return "/nonexistent/kubeconfig/path"
			},
			wantErr:    true,
			wantErrMsg: "load kubeconfig",
		},
		{
			name: "falls back to in-cluster when empty path",
			setupFunc: func(t *testing.T) string {
				// Empty path triggers fallback logic.
				return ""
			},
			wantErr:    true, // Will fail because in-cluster files don't exist.
			wantErrMsg: "load in-cluster config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeconfigPath := tt.setupFunc(t)

			auth, err := loadKubeconfigOrInCluster(kubeconfigPath)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, auth)
				assert.NotEmpty(t, auth.apiServer)
			}
		})
	}
}
