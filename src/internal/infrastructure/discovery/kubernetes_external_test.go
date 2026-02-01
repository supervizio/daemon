//go:build ignore

// TODO: Enable when auth injection mechanism is implemented for KubernetesDiscoverer.

package discovery_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/kodflow/daemon/internal/infrastructure/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKubernetesDiscoverer_Type verifies the discoverer returns correct type.
func TestKubernetesDiscoverer_Type(t *testing.T) {
	tests := []struct {
		name string
		want target.Type
	}{
		{
			name: "returns TypeKubernetes",
			want: target.TypeKubernetes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock K8s API server.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Return empty pod list.
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"items": []any{},
				})
			}))
			defer server.Close()

			cfg := &config.KubernetesDiscoveryConfig{
				Enabled: true,
			}

			// Create discoverer with mock server (use test helper when available).
			d := newTestKubernetesDiscoverer(t, server.URL, cfg)

			got := d.Type()

			assert.Equal(t, tt.want, got)
		})
	}
}

// TestKubernetesDiscoverer_Discover verifies pod discovery from K8s API.
func TestKubernetesDiscoverer_Discover(t *testing.T) {
	tests := []struct {
		name       string
		apiHandler http.HandlerFunc
		namespaces []string
		wantCount  int
		wantErr    bool
	}{
		{
			name: "discovers running pods",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				// Return pod list with one running pod.
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"items": []map[string]any{
						{
							"metadata": map[string]any{
								"name":      "nginx-pod",
								"namespace": "default",
								"labels": map[string]string{
									"app": "nginx",
								},
							},
							"spec": map[string]any{
								"containers": []map[string]any{
									{
										"name": "nginx",
										"ports": []map[string]any{
											{
												"containerPort": 80,
												"protocol":      "TCP",
											},
										},
									},
								},
							},
							"status": map[string]any{
								"phase": "Running",
								"podIP": "10.0.0.1",
							},
						},
					},
				})
			},
			namespaces: []string{"default"},
			wantCount:  1,
			wantErr:    false,
		},
		{
			name: "skips non-running pods",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				// Return pod list with pending pod.
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"items": []map[string]any{
						{
							"metadata": map[string]any{
								"name":      "pending-pod",
								"namespace": "default",
								"labels":    map[string]string{},
							},
							"spec": map[string]any{
								"containers": []map[string]any{},
							},
							"status": map[string]any{
								"phase": "Pending",
								"podIP": "",
							},
						},
					},
				})
			},
			namespaces: []string{"default"},
			wantCount:  0,
			wantErr:    false,
		},
		{
			name: "handles empty response",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				// Return empty pod list.
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"items": []any{},
				})
			},
			namespaces: []string{"default"},
			wantCount:  0,
			wantErr:    false,
		},
		{
			name: "handles API error",
			apiHandler: func(w http.ResponseWriter, r *http.Request) {
				// Return 500 error.
				w.WriteHeader(http.StatusInternalServerError)
			},
			namespaces: []string{"default"},
			wantCount:  0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock K8s API server.
			server := httptest.NewServer(tt.apiHandler)
			defer server.Close()

			cfg := &config.KubernetesDiscoveryConfig{
				Enabled:    true,
				Namespaces: tt.namespaces,
			}

			d := newTestKubernetesDiscoverer(t, server.URL, cfg)

			targets, err := d.Discover(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, targets, tt.wantCount)

				// Verify target fields for non-empty results.
				if tt.wantCount > 0 {
					assert.Equal(t, target.TypeKubernetes, targets[0].Type)
					assert.Equal(t, target.SourceDiscovered, targets[0].Source)
					assert.NotEmpty(t, targets[0].ID)
					assert.NotEmpty(t, targets[0].Name)
				}
			}
		})
	}
}

// TestKubernetesDiscoverer_Discover_LabelSelector verifies label selector filtering.
func TestKubernetesDiscoverer_Discover_LabelSelector(t *testing.T) {
	tests := []struct {
		name          string
		labelSelector string
		wantQuery     bool
	}{
		{
			name:          "applies label selector",
			labelSelector: "app=nginx,version=v1",
			wantQuery:     true,
		},
		{
			name:          "no label selector",
			labelSelector: "",
			wantQuery:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedQuery string

			// Create mock K8s API server that captures query params.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				receivedQuery = r.URL.RawQuery
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"items": []any{},
				})
			}))
			defer server.Close()

			cfg := &config.KubernetesDiscoveryConfig{
				Enabled:       true,
				Namespaces:    []string{"default"},
				LabelSelector: tt.labelSelector,
			}

			d := newTestKubernetesDiscoverer(t, server.URL, cfg)

			_, err := d.Discover(context.Background())
			require.NoError(t, err)

			if tt.wantQuery {
				assert.Contains(t, receivedQuery, "labelSelector=")
			} else {
				assert.NotContains(t, receivedQuery, "labelSelector=")
			}
		})
	}
}

// newTestKubernetesDiscoverer creates a test discoverer with mock API server.
// This is a helper for tests that bypasses kubeconfig loading.
func newTestKubernetesDiscoverer(t *testing.T, apiServer string, cfg *config.KubernetesDiscoveryConfig) *discovery.KubernetesDiscoverer {
	t.Helper()

	// For tests, we'd need to expose a way to inject the auth.
	// For now, this will fail in real tests. This is a placeholder.
	// In production code, we'd add a WithAuth option or similar.

	// Skip test if we can't create discoverer with mock server.
	// TODO: implement auth injection mechanism.
	t.Skip("newTestKubernetesDiscoverer requires auth injection mechanism")
	return nil
}
