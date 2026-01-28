// Package tui_test provides external tests.
package tui_test

import (
	"testing"
	"time"

	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// mockTUISnapshotser is a mock for TUISnapshotser.
type mockTUISnapshotser struct {
	snapshots []tui.TUISnapshotData
}

func (m *mockTUISnapshotser) TUISnapshots() []tui.TUISnapshotData {
	return m.snapshots
}

// mockProcessMetricsProvider is a mock for ProcessMetricsProvider.
type mockProcessMetricsProvider struct {
	metrics map[string]domainmetrics.ProcessMetrics
}

func (m *mockProcessMetricsProvider) Get(serviceName string) (domainmetrics.ProcessMetrics, bool) {
	if m == nil || m.metrics == nil {
		return domainmetrics.ProcessMetrics{}, false
	}
	metric, ok := m.metrics[serviceName]
	return metric, ok
}

func (m *mockProcessMetricsProvider) Has(serviceName string) bool {
	if m == nil || m.metrics == nil {
		return false
	}
	_, ok := m.metrics[serviceName]
	return ok
}

func TestNewDynamicServiceProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{"with_both_providers"},
		{"with_nil_metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			provider := &mockTUISnapshotser{
				snapshots: []tui.TUISnapshotData{
					{Name: "test", State: process.StateRunning, PID: 100, Uptime: 3600},
				},
			}
			
			var metricsProvider tui.ProcessMetricsProvider
			if tt.name == "with_both_providers" {
				metricsProvider = &mockProcessMetricsProvider{
					metrics: map[string]domainmetrics.ProcessMetrics{
						"test": {
							CPU:    domainmetrics.ProcessCPU{UsagePercent: 25.5},
							Memory: domainmetrics.ProcessMemory{RSS: 1024 * 1024 * 100},
						},
					},
				}
			}
			
			dsp := tui.NewDynamicServiceProvider(provider, metricsProvider)
			assert.NotNil(t, dsp)
		})
	}
}

func TestDynamicServiceProvider_ListServices_Empty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "empty_snapshots"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			provider := &mockTUISnapshotser{
				snapshots: []tui.TUISnapshotData{},
			}

			dsp := tui.NewDynamicServiceProvider(provider, nil)
			services := dsp.ListServices()
			assert.NotNil(t, services)
			assert.Len(t, services, 0)
		})
	}
}

func TestDynamicServiceProvider_ListServices_NilProvider(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "nil_provider"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dsp := tui.NewDynamicServiceProvider(nil, nil)
			services := dsp.ListServices()
			assert.Nil(t, services)
		})
	}
}

func TestDynamicServiceProvider_ListServices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		snapshots      []tui.TUISnapshotData
		withMetrics    bool
		expectedCount  int
	}{
		{
			name: "single_service_no_metrics",
			snapshots: []tui.TUISnapshotData{
				{Name: "api", State: process.StateRunning, PID: 100, Uptime: 3600},
			},
			withMetrics:   false,
			expectedCount: 1,
		},
		{
			name: "single_service_with_metrics",
			snapshots: []tui.TUISnapshotData{
				{Name: "api", State: process.StateRunning, PID: 100, Uptime: 3600},
			},
			withMetrics:   true,
			expectedCount: 1,
		},
		{
			name: "multiple_services",
			snapshots: []tui.TUISnapshotData{
				{Name: "api", State: process.StateRunning, PID: 100, Uptime: 3600},
				{Name: "worker", State: process.StateRunning, PID: 101, Uptime: 1800},
				{Name: "db", State: process.StateStopped, PID: 0, Uptime: 0},
			},
			withMetrics:   true,
			expectedCount: 3,
		},
		{
			name: "mixed_states",
			snapshots: []tui.TUISnapshotData{
				{Name: "svc1", State: process.StateRunning, PID: 100, Uptime: 3600},
				{Name: "svc2", State: process.StateStopped, PID: 0, Uptime: 0},
				{Name: "svc3", State: process.StateFailed, PID: 0, Uptime: 0},
				{Name: "svc4", State: process.StateStarting, PID: 102, Uptime: 10},
			},
			withMetrics:   false,
			expectedCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			provider := &mockTUISnapshotser{
				snapshots: tt.snapshots,
			}
			
			var metricsProvider tui.ProcessMetricsProvider
			if tt.withMetrics {
				mockMetrics := &mockProcessMetricsProvider{
					metrics: make(map[string]domainmetrics.ProcessMetrics),
				}
				for _, snap := range tt.snapshots {
					mockMetrics.metrics[snap.Name] = domainmetrics.ProcessMetrics{
						CPU:    domainmetrics.ProcessCPU{UsagePercent: 25.5},
						Memory: domainmetrics.ProcessMemory{RSS: 1024 * 1024 * 100},
					}
				}
				metricsProvider = mockMetrics
			}
			
			dsp := tui.NewDynamicServiceProvider(provider, metricsProvider)
			services := dsp.ListServices()
			
			assert.Len(t, services, tt.expectedCount)
			
			for i, svc := range services {
				assert.Equal(t, tt.snapshots[i].Name, svc.Name)
				assert.Equal(t, tt.snapshots[i].State, svc.State)
				assert.Equal(t, tt.snapshots[i].PID, svc.PID)
				assert.Equal(t, time.Duration(tt.snapshots[i].Uptime)*time.Second, svc.Uptime)
				
				if tt.withMetrics {
					assert.Equal(t, 25.5, svc.CPUPercent)
					assert.Equal(t, uint64(1024*1024*100), svc.MemoryRSS)
				} else {
					assert.Equal(t, 0.0, svc.CPUPercent)
					assert.Equal(t, uint64(0), svc.MemoryRSS)
				}
			}
		})
	}
}

func TestDynamicServiceProvider_ListServices_PartialMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "partial_metrics_coverage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test when only some services have metrics.
			provider := &mockTUISnapshotser{
				snapshots: []tui.TUISnapshotData{
					{Name: "api", State: process.StateRunning, PID: 100, Uptime: 3600},
					{Name: "worker", State: process.StateRunning, PID: 101, Uptime: 1800},
					{Name: "db", State: process.StateRunning, PID: 102, Uptime: 7200},
				},
			}

			// Only provide metrics for "api" and "db", not "worker".
			metricsProvider := &mockProcessMetricsProvider{
				metrics: map[string]domainmetrics.ProcessMetrics{
					"api": {
						CPU:    domainmetrics.ProcessCPU{UsagePercent: 30.0},
						Memory: domainmetrics.ProcessMemory{RSS: 200 * 1024 * 1024},
					},
					"db": {
						CPU:    domainmetrics.ProcessCPU{UsagePercent: 50.0},
						Memory: domainmetrics.ProcessMemory{RSS: 500 * 1024 * 1024},
					},
				},
			}

			dsp := tui.NewDynamicServiceProvider(provider, metricsProvider)
			services := dsp.ListServices()

			assert.Len(t, services, 3)

			// Check "api" has metrics.
			apiService := services[0]
			assert.Equal(t, "api", apiService.Name)
			assert.Equal(t, 30.0, apiService.CPUPercent)
			assert.Equal(t, uint64(200*1024*1024), apiService.MemoryRSS)

			// Check "worker" has zero metrics (not found).
			workerService := services[1]
			assert.Equal(t, "worker", workerService.Name)
			assert.Equal(t, 0.0, workerService.CPUPercent)
			assert.Equal(t, uint64(0), workerService.MemoryRSS)

			// Check "db" has metrics.
			dbService := services[2]
			assert.Equal(t, "db", dbService.Name)
			assert.Equal(t, 50.0, dbService.CPUPercent)
			assert.Equal(t, uint64(500*1024*1024), dbService.MemoryRSS)
		})
	}
}

func TestDynamicServiceProvider_ListServices_UptimeConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		uptimeSeconds  int64
		expectedUptime time.Duration
	}{
		{"zero_uptime", 0, 0},
		{"one_minute", 60, 60 * time.Second},
		{"one_hour", 3600, 3600 * time.Second},
		{"one_day", 86400, 86400 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			provider := &mockTUISnapshotser{
				snapshots: []tui.TUISnapshotData{
					{Name: "test", State: process.StateRunning, PID: 100, Uptime: tt.uptimeSeconds},
				},
			}
			
			dsp := tui.NewDynamicServiceProvider(provider, nil)
			services := dsp.ListServices()
			
			assert.Len(t, services, 1)
			assert.Equal(t, tt.expectedUptime, services[0].Uptime)
		})
	}
}

func TestDynamicServiceProvider_ListServices_AllStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "all_process_states"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			states := []process.State{
				process.StateRunning,
				process.StateStopped,
				process.StateFailed,
				process.StateStarting,
				process.StateStopping,
			}

			snapshots := make([]tui.TUISnapshotData, len(states))
			for i, state := range states {
				snapshots[i] = tui.TUISnapshotData{
					Name:   "svc",
					State:  state,
					PID:    100 + i,
					Uptime: int64((i + 1) * 100),
				}
			}

			provider := &mockTUISnapshotser{
				snapshots: snapshots,
			}

			dsp := tui.NewDynamicServiceProvider(provider, nil)
			services := dsp.ListServices()

			assert.Len(t, services, len(states))
			for i, svc := range services {
				assert.Equal(t, states[i], svc.State)
			}
		})
	}
}

func TestDynamicServiceProvider_ListServices_Snapshot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "snapshot_type_validation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test that the returned services are proper model.ServiceSnapshot instances.
			provider := &mockTUISnapshotser{
				snapshots: []tui.TUISnapshotData{
					{Name: "api", State: process.StateRunning, PID: 12345, Uptime: 3600},
				},
			}

			metricsProvider := &mockProcessMetricsProvider{
				metrics: map[string]domainmetrics.ProcessMetrics{
					"api": {
						CPU:    domainmetrics.ProcessCPU{UsagePercent: 42.5},
						Memory: domainmetrics.ProcessMemory{RSS: 256 * 1024 * 1024},
					},
				},
			}

			dsp := tui.NewDynamicServiceProvider(provider, metricsProvider)
			services := dsp.ListServices()

			assert.Len(t, services, 1)

			svc := services[0]
			assert.IsType(t, model.ServiceSnapshot{}, svc)
			assert.Equal(t, "api", svc.Name)
			assert.Equal(t, process.StateRunning, svc.State)
			assert.Equal(t, 12345, svc.PID)
			assert.Equal(t, 3600*time.Second, svc.Uptime)
			assert.Equal(t, 42.5, svc.CPUPercent)
			assert.Equal(t, uint64(256*1024*1024), svc.MemoryRSS)
		})
	}
}
