//go:build unix

package discovery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/target"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test JSON constants for Nomad API responses.
const (
	// testNomadSingleAllocation is a JSON response with one running allocation.
	testNomadSingleAllocation string = `[
				{
					"ID": "abcd1234-5678-90ef-ghij-klmnopqrstuv",
					"Name": "web-server.app[0]",
					"JobID": "web-server",
					"TaskGroup": "app",
					"Namespace": "default",
					"ClientStatus": "running",
					"TaskStates": {
						"nginx": {"State": "running", "Failed": false}
					}
				}
			]`

	// testNomadTwoAllocations is a JSON response with two running allocations.
	testNomadTwoAllocations string = `[
				{
					"ID": "abcd1234-5678-90ef-ghij-klmnopqrstuv",
					"Name": "web-server.app[0]",
					"JobID": "web-server",
					"TaskGroup": "app",
					"Namespace": "default",
					"ClientStatus": "running",
					"TaskStates": {
						"nginx": {"State": "running", "Failed": false}
					}
				},
				{
					"ID": "efgh5678-90ab-cdef-1234-567890abcdef",
					"Name": "db-postgres.db[0]",
					"JobID": "db-postgres",
					"TaskGroup": "db",
					"Namespace": "default",
					"ClientStatus": "running",
					"TaskStates": {
						"postgres": {"State": "running", "Failed": false}
					}
				}
			]`

	// testNomadPendingAllocation is a JSON response with a pending allocation.
	testNomadPendingAllocation string = `[
				{
					"ID": "abcd1234-5678-90ef-ghij-klmnopqrstuv",
					"Name": "web-server.app[0]",
					"JobID": "web-server",
					"TaskGroup": "app",
					"Namespace": "default",
					"ClientStatus": "pending",
					"TaskStates": {
						"nginx": {"State": "pending", "Failed": false}
					}
				}
			]`

	// testNomadDetailWithNetwork is a JSON response for allocation detail with networks.
	testNomadDetailWithNetwork string = `{
				"Resources": {
					"Networks": [
						{
							"IP": "192.168.1.10",
							"ReservedPorts": [
								{"Label": "http", "Value": 8080}
							]
						}
					]
				}
			}`
)

// TestNomadDiscoverer_matchesFilters verifies allocation filtering logic.
func TestNomadDiscoverer_matchesFilters(t *testing.T) {
	tests := []struct {
		name        string
		jobFilter   string
		alloc       nomadAllocation
		shouldMatch bool
	}{
		{
			name:      "running allocation with no filter",
			jobFilter: "",
			alloc: nomadAllocation{
				JobID:        "web-server",
				ClientStatus: "running",
			},
			shouldMatch: true,
		},
		{
			name:      "running allocation matches job filter",
			jobFilter: "web-",
			alloc: nomadAllocation{
				JobID:        "web-server",
				ClientStatus: "running",
			},
			shouldMatch: true,
		},
		{
			name:      "running allocation does not match job filter",
			jobFilter: "db-",
			alloc: nomadAllocation{
				JobID:        "web-server",
				ClientStatus: "running",
			},
			shouldMatch: false,
		},
		{
			name:      "pending allocation rejected",
			jobFilter: "",
			alloc: nomadAllocation{
				JobID:        "web-server",
				ClientStatus: "pending",
			},
			shouldMatch: false,
		},
		{
			name:      "dead allocation rejected",
			jobFilter: "",
			alloc: nomadAllocation{
				JobID:        "web-server",
				ClientStatus: "dead",
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &NomadDiscoverer{
				jobFilter: tt.jobFilter,
			}

			result := d.matchesFilters(tt.alloc)

			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

// TestNomadDiscoverer_allocationToTargets verifies allocation conversion.
func TestNomadDiscoverer_allocationToTargets(t *testing.T) {
	tests := []struct {
		name       string
		alloc      nomadAllocation
		detail     *nomadAllocationDetail
		wantCount  int
		checkFirst bool
		wantID     string
		wantName   string
		wantType   target.Type
	}{
		{
			name: "allocation with running task",
			alloc: nomadAllocation{
				ID:        "abcd1234-5678-90ef-ghij-klmnopqrstuv",
				JobID:     "web-server",
				TaskGroup: "app",
				Namespace: "default",
				TaskStates: map[string]nomadTaskState{
					"nginx": {State: "running", Failed: false},
				},
			},
			detail: &nomadAllocationDetail{
				Resources: nomadResources{
					Networks: []nomadNetwork{
						{
							IP: "192.168.1.10",
							ReservedPorts: []nomadPort{
								{Label: "http", Value: 8080},
							},
						},
					},
				},
			},
			wantCount:  1,
			checkFirst: true,
			wantID:     "nomad:abcd1234/nginx",
			wantName:   "web-server/nginx",
			wantType:   target.TypeNomad,
		},
		{
			name: "allocation with multiple running tasks",
			alloc: nomadAllocation{
				ID:        "abcd1234-5678-90ef-ghij-klmnopqrstuv",
				JobID:     "web-server",
				TaskGroup: "app",
				Namespace: "default",
				TaskStates: map[string]nomadTaskState{
					"nginx":   {State: "running", Failed: false},
					"redis":   {State: "running", Failed: false},
					"pending": {State: "pending", Failed: false},
				},
			},
			detail: &nomadAllocationDetail{
				Resources: nomadResources{
					Networks: []nomadNetwork{},
				},
			},
			wantCount:  2,
			checkFirst: false,
		},
		{
			name: "allocation with no running tasks",
			alloc: nomadAllocation{
				ID:        "abcd1234-5678-90ef-ghij-klmnopqrstuv",
				JobID:     "web-server",
				TaskGroup: "app",
				Namespace: "default",
				TaskStates: map[string]nomadTaskState{
					"nginx": {State: "pending", Failed: false},
				},
			},
			detail: &nomadAllocationDetail{
				Resources: nomadResources{
					Networks: []nomadNetwork{},
				},
			},
			wantCount:  0,
			checkFirst: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &NomadDiscoverer{}

			targets := d.allocationToTargets(tt.alloc, tt.detail)

			assert.Len(t, targets, tt.wantCount)
			if tt.checkFirst && tt.wantCount > 0 {
				assert.Equal(t, tt.wantID, targets[0].ID)
				assert.Equal(t, tt.wantName, targets[0].Name)
				assert.Equal(t, tt.wantType, targets[0].Type)
			}
		})
	}
}

// TestNomadDiscoverer_configureProbe verifies probe configuration.
func TestNomadDiscoverer_configureProbe(t *testing.T) {
	tests := []struct {
		name      string
		detail    *nomadAllocationDetail
		wantProbe bool
	}{
		{
			name: "configures probe with reserved port",
			detail: &nomadAllocationDetail{
				Resources: nomadResources{
					Networks: []nomadNetwork{
						{
							IP: "192.168.1.10",
							ReservedPorts: []nomadPort{
								{Label: "http", Value: 8080},
							},
						},
					},
				},
			},
			wantProbe: true,
		},
		{
			name: "configures probe with dynamic port",
			detail: &nomadAllocationDetail{
				Resources: nomadResources{
					Networks: []nomadNetwork{
						{
							IP: "192.168.1.10",
							DynamicPorts: []nomadPort{
								{Label: "db", Value: 5432},
							},
						},
					},
				},
			},
			wantProbe: true,
		},
		{
			name: "prefers reserved over dynamic ports",
			detail: &nomadAllocationDetail{
				Resources: nomadResources{
					Networks: []nomadNetwork{
						{
							IP: "192.168.1.10",
							ReservedPorts: []nomadPort{
								{Label: "http", Value: 8080},
							},
							DynamicPorts: []nomadPort{
								{Label: "metrics", Value: 9090},
							},
						},
					},
				},
			},
			wantProbe: true,
		},
		{
			name: "no networks",
			detail: &nomadAllocationDetail{
				Resources: nomadResources{
					Networks: []nomadNetwork{},
				},
			},
			wantProbe: false,
		},
		{
			name: "no ports",
			detail: &nomadAllocationDetail{
				Resources: nomadResources{
					Networks: []nomadNetwork{
						{IP: "192.168.1.10"},
					},
				},
			},
			wantProbe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &NomadDiscoverer{}
			tgt := target.ExternalTarget{}

			d.configureProbe(&tgt, tt.detail)

			if tt.wantProbe {
				assert.NotEmpty(t, tgt.ProbeType)
				assert.Equal(t, nomadProbeTypeTCP, tgt.ProbeType)
			} else {
				assert.Empty(t, tgt.ProbeType)
			}
		})
	}
}

// TestNomadDiscoverer_Discover_Integration tests the full discovery flow with mock server.
func TestNomadDiscoverer_Discover_Integration(t *testing.T) {
	tests := []struct {
		name             string
		allocationsJSON  string
		detailJSON       string
		namespace        string
		jobFilter        string
		wantTargetCount  int
		wantErr          bool
		checkFirstTarget bool
		wantFirstID      string
	}{
		{
			name:             "discovers running allocations",
			allocationsJSON:  testNomadSingleAllocation,
			detailJSON:       testNomadDetailWithNetwork,
			namespace:        "",
			jobFilter:        "",
			wantTargetCount:  1,
			wantErr:          false,
			checkFirstTarget: true,
			wantFirstID:      "nomad:abcd1234/nginx",
		},
		{
			name:             "filters by job prefix",
			allocationsJSON:  testNomadTwoAllocations,
			detailJSON:       testNomadDetailWithNetwork,
			namespace:        "",
			jobFilter:        "web-",
			wantTargetCount:  1,
			wantErr:          false,
			checkFirstTarget: true,
			wantFirstID:      "nomad:abcd1234/nginx",
		},
		{
			name:            "skips non-running allocations",
			allocationsJSON: testNomadPendingAllocation,
			detailJSON:      `{}`,
			namespace:       "",
			jobFilter:       "",
			wantTargetCount: 0,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Nomad API server.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/v1/allocations" {
					w.Header().Set("Content-Type", "application/json")
					_, err := w.Write([]byte(tt.allocationsJSON))
					require.NoError(t, err)
					return
				}
				if r.URL.Path == "/v1/allocation/abcd1234-5678-90ef-ghij-klmnopqrstuv" ||
					r.URL.Path == "/v1/allocation/efgh5678-90ab-cdef-1234-567890abcdef" {
					w.Header().Set("Content-Type", "application/json")
					_, err := w.Write([]byte(tt.detailJSON))
					require.NoError(t, err)
					return
				}
				w.WriteHeader(http.StatusNotFound)
			}))
			defer server.Close()

			// Create discoverer with mock server.
			cfg := &config.NomadDiscoveryConfig{
				Address:   server.URL,
				Namespace: tt.namespace,
				JobFilter: tt.jobFilter,
			}
			d := NewNomadDiscoverer(cfg)

			// Execute discovery.
			targets, err := d.Discover(context.Background())

			// Verify results.
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, targets, tt.wantTargetCount)
				if tt.checkFirstTarget && len(targets) > 0 {
					assert.Equal(t, tt.wantFirstID, targets[0].ID)
					assert.Equal(t, target.TypeNomad, targets[0].Type)
					assert.Equal(t, target.SourceDiscovered, targets[0].Source)
				}
			}
		})
	}
}

// TestNomadDiscoverer_fetchAllocationDetail verifies detail fetching.
func TestNomadDiscoverer_fetchAllocationDetail(t *testing.T) {
	tests := []struct {
		name         string
		allocID      string
		responseJSON string
		statusCode   int
		wantErr      bool
		wantNetworks int
	}{
		{
			name:         "successful detail fetch",
			allocID:      "abcd1234-5678-90ef-ghij-klmnopqrstuv",
			responseJSON: testNomadDetailWithNetwork,
			statusCode:   http.StatusOK,
			wantErr:      false,
			wantNetworks: 1,
		},
		{
			name:         "allocation not found",
			allocID:      "invalid-id",
			responseJSON: `{}`,
			statusCode:   http.StatusNotFound,
			wantErr:      true,
		},
		{
			name:         "invalid json response",
			allocID:      "abcd1234",
			responseJSON: `{invalid json}`,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				if tt.statusCode == http.StatusOK {
					w.Header().Set("Content-Type", "application/json")
					_, err := w.Write([]byte(tt.responseJSON))
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			// Create discoverer.
			cfg := &config.NomadDiscoveryConfig{Address: server.URL}
			d := NewNomadDiscoverer(cfg)

			// Fetch detail.
			detail, err := d.fetchAllocationDetail(context.Background(), tt.allocID)

			// Verify results.
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, detail)
				assert.Len(t, detail.Resources.Networks, tt.wantNetworks)
			}
		})
	}
}

// TestNomadDiscoverer_fetchAllocations verifies allocation list fetching.
func TestNomadDiscoverer_fetchAllocations(t *testing.T) {
	tests := []struct {
		name           string
		responseJSON   string
		statusCode     int
		namespace      string
		wantAllocCount int
		wantErr        bool
		wantErrMsg     string
	}{
		{
			name:           "successful fetch",
			responseJSON:   testNomadSingleAllocation,
			statusCode:     http.StatusOK,
			namespace:      "",
			wantAllocCount: 1,
			wantErr:        false,
		},
		{
			name:           "fetch multiple allocations",
			responseJSON:   testNomadTwoAllocations,
			statusCode:     http.StatusOK,
			namespace:      "",
			wantAllocCount: 2,
			wantErr:        false,
		},
		{
			name:           "fetch with namespace filter",
			responseJSON:   testNomadSingleAllocation,
			statusCode:     http.StatusOK,
			namespace:      "production",
			wantAllocCount: 1,
			wantErr:        false,
		},
		{
			name:         "api error",
			responseJSON: `{}`,
			statusCode:   http.StatusInternalServerError,
			namespace:    "",
			wantErr:      true,
			wantErrMsg:   "unexpected status code",
		},
		{
			name:         "invalid json",
			responseJSON: `{invalid}`,
			statusCode:   http.StatusOK,
			namespace:    "",
			wantErr:      true,
		},
		{
			name:           "empty allocation list",
			responseJSON:   `[]`,
			statusCode:     http.StatusOK,
			namespace:      "",
			wantAllocCount: 0,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify namespace query param if set.
				if tt.namespace != "" {
					assert.Equal(t, tt.namespace, r.URL.Query().Get("namespace"))
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseJSON))
			}))
			defer server.Close()

			cfg := &config.NomadDiscoveryConfig{
				Address:   server.URL,
				Namespace: tt.namespace,
			}
			d := NewNomadDiscoverer(cfg)

			allocations, err := d.fetchAllocations(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, allocations, tt.wantAllocCount)
			}
		})
	}
}

// TestNomadDiscoverer_processAllocations verifies allocation processing.
func TestNomadDiscoverer_processAllocations(t *testing.T) {
	tests := []struct {
		name        string
		allocations nomadAllocationList
		jobFilter   string
		detailJSON  string
		wantCount   int
	}{
		{
			name: "processes running allocations",
			allocations: nomadAllocationList{
				{
					ID:           "alloc-1",
					JobID:        "web-app",
					TaskGroup:    "app",
					Namespace:    "default",
					ClientStatus: "running",
					TaskStates: map[string]nomadTaskState{
						"nginx": {State: "running", Failed: false},
					},
				},
			},
			jobFilter:  "",
			detailJSON: testNomadDetailWithNetwork,
			wantCount:  1,
		},
		{
			name: "filters by job prefix",
			allocations: nomadAllocationList{
				{
					ID:           "alloc-1",
					JobID:        "web-app",
					ClientStatus: "running",
					TaskStates:   map[string]nomadTaskState{"nginx": {State: "running"}},
				},
				{
					ID:           "alloc-2",
					JobID:        "db-postgres",
					ClientStatus: "running",
					TaskStates:   map[string]nomadTaskState{"postgres": {State: "running"}},
				},
			},
			jobFilter:  "web-",
			detailJSON: testNomadDetailWithNetwork,
			wantCount:  1,
		},
		{
			name: "skips non-running allocations",
			allocations: nomadAllocationList{
				{
					ID:           "alloc-1",
					JobID:        "web-app",
					ClientStatus: "pending",
					TaskStates:   map[string]nomadTaskState{"nginx": {State: "pending"}},
				},
			},
			jobFilter:  "",
			detailJSON: testNomadDetailWithNetwork,
			wantCount:  0,
		},
		{
			name:        "handles empty allocation list",
			allocations: nomadAllocationList{},
			jobFilter:   "",
			detailJSON:  testNomadDetailWithNetwork,
			wantCount:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server for allocation details.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.detailJSON))
			}))
			defer server.Close()

			cfg := &config.NomadDiscoveryConfig{
				Address:   server.URL,
				JobFilter: tt.jobFilter,
			}
			d := NewNomadDiscoverer(cfg)

			targets := d.processAllocations(context.Background(), tt.allocations)

			assert.Len(t, targets, tt.wantCount)
		})
	}
}

// TestNomadDiscoverer_taskToTarget verifies task to target conversion.
func TestNomadDiscoverer_taskToTarget(t *testing.T) {
	tests := []struct {
		name       string
		alloc      nomadAllocation
		taskName   string
		detail     *nomadAllocationDetail
		wantID     string
		wantName   string
		wantType   target.Type
		wantLabels map[string]string
	}{
		{
			name: "converts task with network",
			alloc: nomadAllocation{
				ID:        "abcd1234-5678-90ef-ghij-klmnopqrstuv",
				JobID:     "web-server",
				TaskGroup: "app",
				Namespace: "production",
			},
			taskName: "nginx",
			detail: &nomadAllocationDetail{
				Resources: nomadResources{
					Networks: []nomadNetwork{
						{
							IP: "192.168.1.10",
							ReservedPorts: []nomadPort{
								{Label: "http", Value: 8080},
							},
						},
					},
				},
			},
			wantID:   "nomad:abcd1234/nginx",
			wantName: "web-server/nginx",
			wantType: target.TypeNomad,
			wantLabels: map[string]string{
				"nomad.job":        "web-server",
				"nomad.task":       "nginx",
				"nomad.task_group": "app",
				"nomad.namespace":  "production",
			},
		},
		{
			name: "converts task without network",
			alloc: nomadAllocation{
				ID:        "short-id",
				JobID:     "batch-job",
				TaskGroup: "batch",
				Namespace: "default",
			},
			taskName: "worker",
			detail: &nomadAllocationDetail{
				Resources: nomadResources{
					Networks: []nomadNetwork{},
				},
			},
			wantID:   "nomad:short-id/worker",
			wantName: "batch-job/worker",
			wantType: target.TypeNomad,
			wantLabels: map[string]string{
				"nomad.job":        "batch-job",
				"nomad.task":       "worker",
				"nomad.task_group": "batch",
				"nomad.namespace":  "default",
			},
		},
		{
			name: "truncates long allocation ID",
			alloc: nomadAllocation{
				ID:        "a1b2c3d4e5f6g7h8i9j0",
				JobID:     "long-id-job",
				TaskGroup: "group",
				Namespace: "ns",
			},
			taskName: "task",
			detail: &nomadAllocationDetail{
				Resources: nomadResources{Networks: []nomadNetwork{}},
			},
			wantID:   "nomad:a1b2c3d4/task",
			wantName: "long-id-job/task",
			wantType: target.TypeNomad,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &NomadDiscoverer{}

			tgt := d.taskToTarget(tt.alloc, tt.taskName, tt.detail)

			assert.Equal(t, tt.wantID, tgt.ID)
			assert.Equal(t, tt.wantName, tgt.Name)
			assert.Equal(t, tt.wantType, tgt.Type)
			assert.Equal(t, target.SourceDiscovered, tgt.Source)

			for key, value := range tt.wantLabels {
				assert.Equal(t, value, tgt.Labels[key], "label %s mismatch", key)
			}
		})
	}
}
