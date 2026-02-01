//go:build unix

package discovery

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test kubeconfig YAML constants.
const (
	// testKubeconfigValid is a valid kubeconfig with all required fields.
	testKubeconfigValid string = `
current-context: default
clusters:
- name: default-cluster
  cluster:
    server: https://localhost:6443
    certificate-authority-data: test-ca
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

	// testKubeconfigMissingContext is a kubeconfig missing the current-context field.
	testKubeconfigMissingContext string = `
clusters:
- name: default-cluster
  cluster:
    server: https://localhost:6443
`

	// testKubeconfigMissingCluster is a kubeconfig referencing a non-existent cluster.
	testKubeconfigMissingCluster string = `
current-context: default
contexts:
- name: default
  context:
    cluster: nonexistent-cluster
    user: default-user
`

	// testKubeconfigMissingUser is a kubeconfig referencing a non-existent user.
	testKubeconfigMissingUser string = `
current-context: default
clusters:
- name: default-cluster
  cluster:
    server: https://localhost:6443
contexts:
- name: default
  context:
    cluster: default-cluster
    user: nonexistent-user
`
)

// TestLoadKubeconfig verifies kubeconfig parsing.
func TestLoadKubeconfig(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig string
		wantErr    bool
		wantServer string
		wantToken  string
	}{
		{
			name:       "parses valid kubeconfig",
			kubeconfig: testKubeconfigValid,
			wantErr:    false,
			wantServer: "https://localhost:6443",
			wantToken:  "test-token",
		},
		{
			name:       "fails on missing file",
			kubeconfig: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.kubeconfig == "" {
				_, err := loadKubeconfig("/nonexistent/kubeconfig")
				assert.Error(t, err)
				return
			}

			// Create temporary kubeconfig file.
			tmpDir := t.TempDir()
			kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
			err := os.WriteFile(kubeconfigPath, []byte(tt.kubeconfig), 0600)
			require.NoError(t, err)

			auth, err := loadKubeconfig(kubeconfigPath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantServer, auth.apiServer)
				assert.Equal(t, tt.wantToken, auth.token)
			}
		})
	}
}

// TestNewHTTPClient verifies HTTP client creation with TLS.
func TestNewHTTPClient(t *testing.T) {
	tests := []struct {
		name    string
		auth    *k8sAuth
		wantErr bool
	}{
		{
			name: "creates client without CA cert",
			auth: &k8sAuth{
				apiServer: "https://localhost:6443",
				token:     "test-token",
				caCert:    nil,
			},
			wantErr: false,
		},
		{
			name: "creates client with empty CA cert",
			auth: &k8sAuth{
				apiServer: "https://localhost:6443",
				token:     "test-token",
				caCert:    []byte{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := newHTTPClient(tt.auth)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, client)

				// Verify TLS config is present.
				transport, ok := client.Transport.(*http.Transport)
				require.True(t, ok)
				assert.NotNil(t, transport.TLSClientConfig)
				assert.Equal(t, uint16(tls.VersionTLS12), transport.TLSClientConfig.MinVersion)
			}
		})
	}
}

// TestPodToTarget verifies pod to target conversion.
func TestPodToTarget(t *testing.T) {
	tests := []struct {
		name          string
		pod           k8sPod
		wantID        string
		wantName      string
		wantProbeType string
		wantLabels    map[string]string
	}{
		{
			name: "converts pod with TCP port",
			pod: k8sPod{
				Metadata: k8sMetadata{
					Name:      "nginx-pod",
					Namespace: "default",
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: k8sPodSpec{
					Containers: []k8sContainer{
						{
							Name: "nginx",
							Ports: []k8sPort{
								{
									ContainerPort: 80,
									Protocol:      "TCP",
								},
							},
						},
					},
				},
				Status: k8sPodStatus{
					Phase: "Running",
					PodIP: "10.0.0.1",
				},
			},
			wantID:        "kubernetes:default/nginx-pod",
			wantName:      "nginx-pod",
			wantProbeType: "tcp",
			wantLabels: map[string]string{
				"app":                  "nginx",
				"kubernetes.namespace": "default",
				"kubernetes.pod":       "nginx-pod",
				"kubernetes.phase":     "Running",
			},
		},
		{
			name: "converts pod without ports",
			pod: k8sPod{
				Metadata: k8sMetadata{
					Name:      "worker-pod",
					Namespace: "prod",
					Labels:    map[string]string{},
				},
				Spec: k8sPodSpec{
					Containers: []k8sContainer{
						{
							Name:  "worker",
							Ports: []k8sPort{},
						},
					},
				},
				Status: k8sPodStatus{
					Phase: "Running",
					PodIP: "10.0.0.2",
				},
			},
			wantID:        "kubernetes:prod/worker-pod",
			wantName:      "worker-pod",
			wantProbeType: "",
			wantLabels: map[string]string{
				"kubernetes.namespace": "prod",
				"kubernetes.pod":       "worker-pod",
				"kubernetes.phase":     "Running",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &KubernetesDiscoverer{}

			got := d.podToTarget(tt.pod)

			assert.Equal(t, tt.wantID, got.ID)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, target.TypeKubernetes, got.Type)
			assert.Equal(t, target.SourceDiscovered, got.Source)
			assert.Equal(t, tt.wantProbeType, got.ProbeType)

			// Verify all expected labels are present.
			for key, value := range tt.wantLabels {
				assert.Equal(t, value, got.Labels[key], "label %s mismatch", key)
			}
		})
	}
}

// TestDiscoverNamespace verifies namespace discovery with mock server.
func TestDiscoverNamespace(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		apiHandler http.HandlerFunc
		wantCount  int
		wantErr    bool
	}{
		{
			name:      "discovers pods in namespace",
			namespace: "default",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				// Verify request path.
				assert.Contains(t, r.URL.Path, "/api/v1/namespaces/default/pods")

				// Return pod list.
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(k8sPodList{
					Items: []k8sPod{
						{
							Metadata: k8sMetadata{
								Name:      "pod1",
								Namespace: "default",
								Labels:    map[string]string{},
							},
							Spec: k8sPodSpec{
								Containers: []k8sContainer{},
							},
							Status: k8sPodStatus{
								Phase: "Running",
								PodIP: "10.0.0.1",
							},
						},
					},
				})
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "handles API error",
			namespace: "default",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock K8s API server.
			server := httptest.NewServer(tt.apiHandler)
			defer server.Close()

			// Create discoverer with mock auth.
			d := &KubernetesDiscoverer{
				auth: &k8sAuth{
					apiServer: server.URL,
					token:     "test-token",
					caCert:    nil,
				},
				client: &http.Client{},
			}

			targets, err := d.discoverNamespace(context.Background(), tt.namespace)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, targets, tt.wantCount)
			}
		})
	}
}

// TestConfigureProbe verifies probe configuration logic.
func TestConfigureProbe(t *testing.T) {
	tests := []struct {
		name          string
		pod           k8sPod
		wantProbeType string
		wantAddress   string
	}{
		{
			name: "configures TCP probe for first port",
			pod: k8sPod{
				Spec: k8sPodSpec{
					Containers: []k8sContainer{
						{
							Ports: []k8sPort{
								{ContainerPort: 80, Protocol: "TCP"},
							},
						},
					},
				},
				Status: k8sPodStatus{
					PodIP: "10.0.0.1",
				},
			},
			wantProbeType: "tcp",
			wantAddress:   "10.0.0.1:80",
		},
		{
			name: "uses default TCP when protocol unspecified",
			pod: k8sPod{
				Spec: k8sPodSpec{
					Containers: []k8sContainer{
						{
							Ports: []k8sPort{
								{ContainerPort: 8080, Protocol: ""},
							},
						},
					},
				},
				Status: k8sPodStatus{
					PodIP: "10.0.0.2",
				},
			},
			wantProbeType: "tcp",
			wantAddress:   "10.0.0.2:8080",
		},
		{
			name: "skips UDP ports",
			pod: k8sPod{
				Spec: k8sPodSpec{
					Containers: []k8sContainer{
						{
							Ports: []k8sPort{
								{ContainerPort: 53, Protocol: "UDP"},
							},
						},
					},
				},
				Status: k8sPodStatus{
					PodIP: "10.0.0.3",
				},
			},
			wantProbeType: "",
			wantAddress:   "",
		},
		{
			name: "no probe when no ports",
			pod: k8sPod{
				Spec: k8sPodSpec{
					Containers: []k8sContainer{
						{
							Ports: []k8sPort{},
						},
					},
				},
				Status: k8sPodStatus{
					PodIP: "10.0.0.4",
				},
			},
			wantProbeType: "",
			wantAddress:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &KubernetesDiscoverer{}
			var tgt target.ExternalTarget

			d.configureProbe(&tgt, tt.pod)

			assert.Equal(t, tt.wantProbeType, tgt.ProbeType)
			if tt.wantAddress != "" {
				assert.Equal(t, tt.wantAddress, tgt.ProbeTarget.Address)
			}
		})
	}
}

// TestKubernetesDiscoverer_discoverNamespace verifies namespace discovery with mock server.
func TestKubernetesDiscoverer_discoverNamespace(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		apiHandler http.HandlerFunc
		wantCount  int
		wantErr    bool
	}{
		{
			name:      "discovers pods in namespace",
			namespace: "default",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, r.URL.Path, "/api/v1/namespaces/default/pods")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(k8sPodList{
					Items: []k8sPod{
						{
							Metadata: k8sMetadata{Name: "pod1", Namespace: "default", Labels: map[string]string{}},
							Spec:     k8sPodSpec{Containers: []k8sContainer{}},
							Status:   k8sPodStatus{Phase: "Running", PodIP: "10.0.0.1"},
						},
					},
				})
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name:      "handles API error",
			namespace: "default",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.apiHandler)
			defer server.Close()

			d := &KubernetesDiscoverer{
				auth:   &k8sAuth{apiServer: server.URL, token: "test-token", caCert: nil},
				client: &http.Client{},
			}

			targets, err := d.discoverNamespace(context.Background(), tt.namespace)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, targets, tt.wantCount)
			}
		})
	}
}

// TestKubernetesDiscoverer_podToTarget verifies pod to target conversion.
func TestKubernetesDiscoverer_podToTarget(t *testing.T) {
	tests := []struct {
		name          string
		pod           k8sPod
		wantID        string
		wantName      string
		wantProbeType string
		wantLabels    map[string]string
	}{
		{
			name: "converts pod with TCP port",
			pod: k8sPod{
				Metadata: k8sMetadata{
					Name:      "nginx-pod",
					Namespace: "default",
					Labels:    map[string]string{"app": "nginx"},
				},
				Spec: k8sPodSpec{
					Containers: []k8sContainer{
						{Name: "nginx", Ports: []k8sPort{{ContainerPort: 80, Protocol: "TCP"}}},
					},
				},
				Status: k8sPodStatus{Phase: "Running", PodIP: "10.0.0.1"},
			},
			wantID:        "kubernetes:default/nginx-pod",
			wantName:      "nginx-pod",
			wantProbeType: "tcp",
			wantLabels: map[string]string{
				"app":                  "nginx",
				"kubernetes.namespace": "default",
				"kubernetes.pod":       "nginx-pod",
				"kubernetes.phase":     "Running",
			},
		},
		{
			name: "converts pod without ports",
			pod: k8sPod{
				Metadata: k8sMetadata{Name: "worker-pod", Namespace: "prod", Labels: map[string]string{}},
				Spec:     k8sPodSpec{Containers: []k8sContainer{{Name: "worker", Ports: []k8sPort{}}}},
				Status:   k8sPodStatus{Phase: "Running", PodIP: "10.0.0.2"},
			},
			wantID:        "kubernetes:prod/worker-pod",
			wantName:      "worker-pod",
			wantProbeType: "",
			wantLabels: map[string]string{
				"kubernetes.namespace": "prod",
				"kubernetes.pod":       "worker-pod",
				"kubernetes.phase":     "Running",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &KubernetesDiscoverer{}

			got := d.podToTarget(tt.pod)

			assert.Equal(t, tt.wantID, got.ID)
			assert.Equal(t, tt.wantName, got.Name)
			assert.Equal(t, target.TypeKubernetes, got.Type)
			assert.Equal(t, target.SourceDiscovered, got.Source)
			assert.Equal(t, tt.wantProbeType, got.ProbeType)

			for key, value := range tt.wantLabels {
				assert.Equal(t, value, got.Labels[key], "label %s mismatch", key)
			}
		})
	}
}

// TestKubernetesDiscoverer_configureProbe verifies probe configuration logic.
func TestKubernetesDiscoverer_configureProbe(t *testing.T) {
	tests := []struct {
		name          string
		pod           k8sPod
		wantProbeType string
		wantAddress   string
	}{
		{
			name: "configures TCP probe for first port",
			pod: k8sPod{
				Spec:   k8sPodSpec{Containers: []k8sContainer{{Ports: []k8sPort{{ContainerPort: 80, Protocol: "TCP"}}}}},
				Status: k8sPodStatus{PodIP: "10.0.0.1"},
			},
			wantProbeType: "tcp",
			wantAddress:   "10.0.0.1:80",
		},
		{
			name: "uses default TCP when protocol unspecified",
			pod: k8sPod{
				Spec:   k8sPodSpec{Containers: []k8sContainer{{Ports: []k8sPort{{ContainerPort: 8080, Protocol: ""}}}}},
				Status: k8sPodStatus{PodIP: "10.0.0.2"},
			},
			wantProbeType: "tcp",
			wantAddress:   "10.0.0.2:8080",
		},
		{
			name: "skips UDP ports",
			pod: k8sPod{
				Spec:   k8sPodSpec{Containers: []k8sContainer{{Ports: []k8sPort{{ContainerPort: 53, Protocol: "UDP"}}}}},
				Status: k8sPodStatus{PodIP: "10.0.0.3"},
			},
			wantProbeType: "",
			wantAddress:   "",
		},
		{
			name: "no probe when no ports",
			pod: k8sPod{
				Spec:   k8sPodSpec{Containers: []k8sContainer{{Ports: []k8sPort{}}}},
				Status: k8sPodStatus{PodIP: "10.0.0.4"},
			},
			wantProbeType: "",
			wantAddress:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &KubernetesDiscoverer{}
			var tgt target.ExternalTarget

			d.configureProbe(&tgt, tt.pod)

			assert.Equal(t, tt.wantProbeType, tgt.ProbeType)
			if tt.wantAddress != "" {
				assert.Equal(t, tt.wantAddress, tgt.ProbeTarget.Address)
			}
		})
	}
}

// TestKubeconfigParsing verifies edge cases in kubeconfig parsing.
func TestKubeconfigParsing(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig string
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "missing context",
			kubeconfig: testKubeconfigMissingContext,
			wantErr:    true,
			errMsg:     "no context found",
		},
		{
			name:       "missing cluster",
			kubeconfig: testKubeconfigMissingCluster,
			wantErr:    true,
			errMsg:     "cluster not found",
		},
		{
			name:       "missing user",
			kubeconfig: testKubeconfigMissingUser,
			wantErr:    true,
			errMsg:     "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary kubeconfig file.
			tmpDir := t.TempDir()
			kubeconfigPath := filepath.Join(tmpDir, "kubeconfig")
			err := os.WriteFile(kubeconfigPath, []byte(tt.kubeconfig), 0600)
			require.NoError(t, err)

			_, err = loadKubeconfig(kubeconfigPath)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestKubernetesDiscoverer_buildPodListURL verifies pod list URL construction.
func TestKubernetesDiscoverer_buildPodListURL(t *testing.T) {
	tests := []struct {
		name          string
		apiServer     string
		namespace     string
		labelSelector string
		wantContains  []string
	}{
		{
			name:          "basic url without selector",
			apiServer:     "https://api.example.com:6443",
			namespace:     "default",
			labelSelector: "",
			wantContains:  []string{"https://api.example.com:6443", "/api/v1/namespaces/default/pods"},
		},
		{
			name:          "url with label selector",
			apiServer:     "https://localhost:6443",
			namespace:     "production",
			labelSelector: "app=nginx",
			wantContains:  []string{"https://localhost:6443", "/api/v1/namespaces/production/pods", "labelSelector=app%3Dnginx"},
		},
		{
			name:          "url with complex selector",
			apiServer:     "https://k8s.example.com",
			namespace:     "kube-system",
			labelSelector: "app in (nginx,apache)",
			wantContains:  []string{"https://k8s.example.com", "/api/v1/namespaces/kube-system/pods", "labelSelector="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &KubernetesDiscoverer{
				auth:          &k8sAuth{apiServer: tt.apiServer},
				labelSelector: tt.labelSelector,
			}

			url := d.buildPodListURL(tt.namespace)

			for _, want := range tt.wantContains {
				assert.Contains(t, url, want)
			}
		})
	}
}

// TestKubernetesDiscoverer_fetchPods verifies pod fetching from K8s API.
func TestKubernetesDiscoverer_fetchPods(t *testing.T) {
	tests := []struct {
		name       string
		apiHandler http.HandlerFunc
		wantPodLen int
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "successful fetch",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(k8sPodList{
					Items: []k8sPod{
						{Metadata: k8sMetadata{Name: "pod1"}},
						{Metadata: k8sMetadata{Name: "pod2"}},
					},
				})
			},
			wantPodLen: 2,
			wantErr:    false,
		},
		{
			name: "unauthorized error",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
			},
			wantErr:    true,
			wantErrMsg: "unexpected status code",
		},
		{
			name: "invalid json response",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{invalid json}"))
			},
			wantErr: true,
		},
		{
			name: "empty pod list",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(k8sPodList{Items: []k8sPod{}})
			},
			wantPodLen: 0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.apiHandler)
			defer server.Close()

			d := &KubernetesDiscoverer{
				auth:   &k8sAuth{apiServer: server.URL, token: "test-token"},
				client: &http.Client{},
			}

			pods, err := d.fetchPods(context.Background(), server.URL+"/api/v1/namespaces/default/pods")

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, pods.Items, tt.wantPodLen)
			}
		})
	}
}

// TestKubernetesDiscoverer_filterAndConvertPods verifies pod filtering and conversion.
func TestKubernetesDiscoverer_filterAndConvertPods(t *testing.T) {
	tests := []struct {
		name       string
		podList    *k8sPodList
		wantCount  int
		checkFirst bool
		wantID     string
	}{
		{
			name: "filters running pods with IPs",
			podList: &k8sPodList{
				Items: []k8sPod{
					{
						Metadata: k8sMetadata{Name: "running-pod", Namespace: "default"},
						Status:   k8sPodStatus{Phase: "Running", PodIP: "10.0.0.1"},
					},
					{
						Metadata: k8sMetadata{Name: "pending-pod", Namespace: "default"},
						Status:   k8sPodStatus{Phase: "Pending", PodIP: ""},
					},
				},
			},
			wantCount:  1,
			checkFirst: true,
			wantID:     "kubernetes:default/running-pod",
		},
		{
			name: "skips pods without IP",
			podList: &k8sPodList{
				Items: []k8sPod{
					{
						Metadata: k8sMetadata{Name: "no-ip-pod", Namespace: "default"},
						Status:   k8sPodStatus{Phase: "Running", PodIP: ""},
					},
				},
			},
			wantCount: 0,
		},
		{
			name: "skips non-running pods",
			podList: &k8sPodList{
				Items: []k8sPod{
					{
						Metadata: k8sMetadata{Name: "failed-pod", Namespace: "default"},
						Status:   k8sPodStatus{Phase: "Failed", PodIP: "10.0.0.2"},
					},
					{
						Metadata: k8sMetadata{Name: "succeeded-pod", Namespace: "default"},
						Status:   k8sPodStatus{Phase: "Succeeded", PodIP: "10.0.0.3"},
					},
				},
			},
			wantCount: 0,
		},
		{
			name:      "handles empty pod list",
			podList:   &k8sPodList{Items: []k8sPod{}},
			wantCount: 0,
		},
		{
			name: "processes multiple running pods",
			podList: &k8sPodList{
				Items: []k8sPod{
					{
						Metadata: k8sMetadata{Name: "pod1", Namespace: "ns1"},
						Status:   k8sPodStatus{Phase: "Running", PodIP: "10.0.0.1"},
					},
					{
						Metadata: k8sMetadata{Name: "pod2", Namespace: "ns1"},
						Status:   k8sPodStatus{Phase: "Running", PodIP: "10.0.0.2"},
					},
					{
						Metadata: k8sMetadata{Name: "pod3", Namespace: "ns1"},
						Status:   k8sPodStatus{Phase: "Running", PodIP: "10.0.0.3"},
					},
				},
			},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &KubernetesDiscoverer{}

			targets := d.filterAndConvertPods(tt.podList)

			assert.Len(t, targets, tt.wantCount)
			if tt.checkFirst && len(targets) > 0 {
				assert.Equal(t, tt.wantID, targets[0].ID)
				assert.Equal(t, target.TypeKubernetes, targets[0].Type)
			}
		})
	}
}
