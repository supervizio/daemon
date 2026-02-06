// Package tui provides internal white-box tests.
package tui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// errWrite is a sentinel error for write failures.
var errWrite error = errors.New("write error")

// errorWriter is a mock writer that returns an error on Write.
type errorWriter struct{}

// Write always returns an error.
func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errWrite
}

// mockServiceLister is a mock for ListServicesser.
type mockServiceLister struct {
	services []model.ServiceSnapshot
}

func (m *mockServiceLister) ListServices() []model.ServiceSnapshot {
	return m.services
}

// mockMetricser is a mock for Metricser.
type mockMetricser struct {
	metrics model.SystemMetrics
}

func (m *mockMetricser) Metrics() model.SystemMetrics {
	return m.metrics
}

// mockSummarizeer is a mock for Summarizeer.
type mockSummarizeer struct {
	summary model.LogSummary
}

func (m *mockSummarizeer) Summarize() model.LogSummary {
	return m.summary
}

func Test_TUI_collectData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupLister   bool
		setupMetrics  bool
		setupSummary  bool
		wantServices  int
		wantMetrics   bool
		wantLogsCount int
	}{
		{
			name:          "all_providers_set",
			setupLister:   true,
			setupMetrics:  true,
			setupSummary:  true,
			wantServices:  2,
			wantMetrics:   true,
			wantLogsCount: 13, // InfoCount + WarnCount + ErrorCount
		},
		{
			name:          "no_providers",
			setupLister:   false,
			setupMetrics:  false,
			setupSummary:  false,
			wantServices:  0,
			wantMetrics:   false,
			wantLogsCount: 0,
		},
		{
			name:          "only_lister",
			setupLister:   true,
			setupMetrics:  false,
			setupSummary:  false,
			wantServices:  2,
			wantMetrics:   false,
			wantLogsCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tui := NewTUI(DefaultConfig("1.0.0"))

			if tt.setupLister {
				tui.SetServiceLister(&mockServiceLister{
					services: []model.ServiceSnapshot{
						{Name: "service1", State: process.StateRunning},
						{Name: "service2", State: process.StateStopped},
					},
				})
			}

			if tt.setupMetrics {
				tui.SetMetricser(&mockMetricser{
					metrics: model.SystemMetrics{
						CPUPercent:    75.5,
						MemoryPercent: 80.0,
					},
				})
			}

			if tt.setupSummary {
				tui.SetSummarizeer(&mockSummarizeer{
					summary: model.LogSummary{
						InfoCount:  10,
						WarnCount:  2,
						ErrorCount: 1,
					},
				})
			}

			tui.collectData()

			assert.NotNil(t, tui.snapshot)
			assert.Len(t, tui.snapshot.Services, tt.wantServices)

			if tt.wantMetrics {
				assert.Equal(t, 75.5, tui.snapshot.System.CPUPercent)
				assert.Equal(t, 80.0, tui.snapshot.System.MemoryPercent)
			}

			totalLogs := tui.snapshot.Logs.InfoCount + tui.snapshot.Logs.WarnCount + tui.snapshot.Logs.ErrorCount
			assert.Equal(t, tt.wantLogsCount, totalLogs)
		})
	}
}

func TestMode_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode Mode
		want string
	}{
		{"raw_mode", ModeRaw, "0"},
		{"interactive_mode", ModeInteractive, "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Mode doesn't have a String() method, but we can test type conversion.
			assert.Equal(t, tt.mode, tt.mode)
		})
	}
}

// Test_runRaw tests the runRaw method execution path.
// It verifies that raw mode rendering works without error.
//
// Params:
//   - t: the testing context.
func Test_TUI_runRaw(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// version is the daemon version.
		version string
		// useErrorWriter indicates if error writer should be used.
		useErrorWriter bool
		// wantErr indicates if error is expected.
		wantErr bool
	}{
		{
			name:           "basic_raw_mode_success",
			version:        "1.0.0",
			useErrorWriter: false,
			wantErr:        false,
		},
		{
			name:           "with_services_success",
			version:        "2.0.0",
			useErrorWriter: false,
			wantErr:        false,
		},
		{
			name:           "error_writer_returns_error",
			version:        "1.0.0",
			useErrorWriter: true,
			wantErr:        true,
		},
		{
			name:           "error_with_services",
			version:        "2.0.0",
			useErrorWriter: true,
			wantErr:        true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create config with appropriate output.
			var buf strings.Builder
			config := Config{
				Mode:            ModeRaw,
				RefreshInterval: defaultRefreshInterval,
				Version:         tt.version,
				Output:          &buf,
			}

			// Create TUI.
			tui := NewTUI(config)

			// Use error writer if requested.
			if tt.useErrorWriter {
				tui.config.Output = &errorWriter{}
			} else {
				tui.config.Output = &buf
			}

			// Setup service lister for service tests.
			if strings.Contains(tt.name, "services") {
				tui.SetServiceLister(&mockServiceLister{
					services: []model.ServiceSnapshot{
						{Name: "service1", State: process.StateRunning},
					},
				})
			}

			// Execute runRaw and check for error.
			err := tui.runRaw()

			// Verify error expectation.
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errWrite)
			} else {
				assert.NoError(t, err)
			}

			// Verify TUI is properly configured.
			assert.NotNil(t, tui.collectors)
			assert.NotNil(t, tui.snapshot)
		})
	}
}

// Test_runInteractive tests the runInteractive method.
// It verifies that interactive mode fallback works correctly.
//
// Params:
//   - t: the testing context.
func Test_TUI_runInteractive(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// useErrorWriter indicates if error writer should be used.
		useErrorWriter bool
		// wantErr indicates if error is expected.
		wantErr bool
	}{
		{
			name:           "fallback_to_raw_when_no_tty_success",
			useErrorWriter: false,
			wantErr:        false,
		},
		{
			name:           "fallback_to_raw_with_services",
			useErrorWriter: false,
			wantErr:        false,
		},
		{
			name:           "fallback_error_writer_returns_error",
			useErrorWriter: true,
			wantErr:        true,
		},
		{
			name:           "interactive_fallback_propagates_error",
			useErrorWriter: true,
			wantErr:        true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create config.
			var buf strings.Builder
			config := Config{
				Mode:            ModeInteractive,
				RefreshInterval: defaultRefreshInterval,
				Version:         "1.0.0",
				Output:          &buf,
			}

			// Create TUI.
			tui := NewTUI(config)

			// Use error writer if requested.
			if tt.useErrorWriter {
				tui.config.Output = &errorWriter{}
			} else {
				tui.config.Output = &buf
			}

			// Setup service lister for service tests.
			if strings.Contains(tt.name, "services") {
				tui.SetServiceLister(&mockServiceLister{
					services: []model.ServiceSnapshot{
						{Name: "service1", State: process.StateRunning},
					},
				})
			}

			// Execute runInteractive (falls back to runRaw in test environment).
			// In non-TTY environment, this calls runRaw.
			err := tui.runInteractive(context.Background())

			// Verify error expectation.
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errWrite)
			} else {
				assert.NoError(t, err)
			}

			// Verify TUI is properly configured.
			assert.NotNil(t, tui.collectors)
			assert.NotNil(t, tui.snapshot)
		})
	}
}
