// Package logging_test provides external tests for capture.go.
// It tests the public API of the Capture type using black-box testing.
package logging_test

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/infrastructure/logging"
)

// mockCaptureConfig implements captureConfig for testing.
type mockCaptureConfig struct {
	// basePath is the base directory for log files.
	basePath string
}

// GetServiceLogPath returns the full path for a service log file.
//
// Params:
//   - serviceName: the name of the service.
//   - logFile: the log file name.
//
// Returns:
//   - string: the full path to the log file.
func (m *mockCaptureConfig) GetServiceLogPath(serviceName, logFile string) string {
	// Return the combined path.
	return filepath.Join(m.basePath, serviceName, logFile)
}

// mockServiceLogging implements serviceLogging interface for testing.
type mockServiceLogging struct {
	stdout service.LogStreamConfig
	stderr service.LogStreamConfig
}

// StdoutConfig returns the stdout configuration.
func (m *mockServiceLogging) StdoutConfig() *service.LogStreamConfig {
	return &m.stdout
}

// StderrConfig returns the stderr configuration.
func (m *mockServiceLogging) StderrConfig() *service.LogStreamConfig {
	return &m.stderr
}

// createServiceLogging creates a mockServiceLogging with the given file paths.
//
// Params:
//   - stdoutFile: the stdout log file path.
//   - stderrFile: the stderr log file path.
//
// Returns:
//   - *mockServiceLogging: the configured service logging.
func createServiceLogging(stdoutFile, stderrFile string) *mockServiceLogging {
	return &mockServiceLogging{
		stdout: service.LogStreamConfig{FilePath: stdoutFile},
		stderr: service.LogStreamConfig{FilePath: stderrFile},
	}
}

// TestNewCapture tests the NewCapture constructor.
//
// Params:
//   - t: the testing context.
func TestNewCapture(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// stdoutFile is the stdout file path.
		stdoutFile string
		// stderrFile is the stderr file path.
		stderrFile string
		// wantErr indicates whether an error is expected.
		wantErr bool
	}{
		{
			name:       "passthrough_mode",
			stdoutFile: "",
			stderrFile: "",
			wantErr:    false,
		},
		{
			name:       "file_mode_stdout_only",
			stdoutFile: "stdout.log",
			stderrFile: "",
			wantErr:    false,
		},
		{
			name:       "file_mode_both",
			stdoutFile: "stdout.log",
			stderrFile: "stderr.log",
			wantErr:    false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &mockCaptureConfig{basePath: tmpDir}
			svcCfg := createServiceLogging(tt.stdoutFile, tt.stderrFile)

			capture, err := logging.NewCapture("test-service", cfg, svcCfg)

			// Check error expectation.
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, capture)
				// Return early.
				return
			}

			require.NoError(t, err)
			require.NotNil(t, capture)
			defer func() { _ = capture.Close() }()

			// Verify stdout and stderr are not nil.
			assert.NotNil(t, capture.Stdout())
			assert.NotNil(t, capture.Stderr())
		})
	}
}

// TestCapture_Stdout tests the Stdout method.
//
// Params:
//   - t: the testing context.
func TestCapture_Stdout(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// stdoutFile is the stdout file path.
		stdoutFile string
	}{
		{
			name:       "passthrough_returns_writer",
			stdoutFile: "",
		},
		{
			name:       "file_returns_writer",
			stdoutFile: "stdout.log",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &mockCaptureConfig{basePath: tmpDir}
			svcCfg := createServiceLogging(tt.stdoutFile, "")

			capture, err := logging.NewCapture("test-service", cfg, svcCfg)
			require.NoError(t, err)
			defer func() { _ = capture.Close() }()

			stdout := capture.Stdout()
			assert.NotNil(t, stdout)
			assert.Implements(t, (*io.Writer)(nil), stdout)
		})
	}
}

// TestCapture_Stderr tests the Stderr method.
//
// Params:
//   - t: the testing context.
func TestCapture_Stderr(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// stderrFile is the stderr file path.
		stderrFile string
	}{
		{
			name:       "passthrough_returns_writer",
			stderrFile: "",
		},
		{
			name:       "file_returns_writer",
			stderrFile: "stderr.log",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &mockCaptureConfig{basePath: tmpDir}
			svcCfg := createServiceLogging("", tt.stderrFile)

			capture, err := logging.NewCapture("test-service", cfg, svcCfg)
			require.NoError(t, err)
			defer func() { _ = capture.Close() }()

			stderr := capture.Stderr()
			assert.NotNil(t, stderr)
			assert.Implements(t, (*io.Writer)(nil), stderr)
		})
	}
}

// TestCapture_Close tests the Close method.
//
// Params:
//   - t: the testing context.
func TestCapture_Close(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// stdoutFile is the stdout file path.
		stdoutFile string
		// stderrFile is the stderr file path.
		stderrFile string
		// closeCount is the number of times to call Close.
		closeCount int
	}{
		{
			name:       "close_passthrough",
			stdoutFile: "",
			stderrFile: "",
			closeCount: 1,
		},
		{
			name:       "close_file_mode",
			stdoutFile: "stdout.log",
			stderrFile: "stderr.log",
			closeCount: 1,
		},
		{
			name:       "double_close_is_safe",
			stdoutFile: "stdout.log",
			stderrFile: "stderr.log",
			closeCount: 2,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := &mockCaptureConfig{basePath: tmpDir}
			svcCfg := createServiceLogging(tt.stdoutFile, tt.stderrFile)

			capture, err := logging.NewCapture("test-service", cfg, svcCfg)
			require.NoError(t, err)

			// Call Close multiple times.
			for i := range tt.closeCount {
				_ = i
				err = capture.Close()
				assert.NoError(t, err)
			}
		})
	}
}

// TestNewCapture_StdoutCreationError tests NewCapture when stdout writer creation fails.
//
// Params:
//   - t: the testing context.
func TestNewCapture_StdoutCreationError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// stdoutFile is the stdout file path that will cause an error.
		stdoutFile string
	}{
		{
			name:       "stdout_invalid_path_fails",
			stdoutFile: "stdout.log",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Use a non-existent, non-writable path.
			cfg := &mockCaptureConfig{basePath: "/nonexistent/readonly/path"}
			svcCfg := createServiceLogging(tt.stdoutFile, "")

			capture, err := logging.NewCapture("test-service", cfg, svcCfg)
			assert.Error(t, err)
			assert.Nil(t, capture)
		})
	}
}

// TestNewCapture_StderrCreationError tests NewCapture when stderr writer creation fails.
//
// Params:
//   - t: the testing context.
func TestNewCapture_StderrCreationError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// stdoutFile is the stdout file path.
		stdoutFile string
		// stderrFile is the stderr file path that will cause an error.
		stderrFile string
	}{
		{
			name:       "stderr_creation_fails_after_stdout_success",
			stdoutFile: "",
			stderrFile: "stderr.log",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Use a non-existent, non-writable path for stderr only.
			cfg := &mockCaptureConfig{basePath: "/nonexistent/readonly/path"}
			svcCfg := createServiceLogging(tt.stdoutFile, tt.stderrFile)

			capture, err := logging.NewCapture("test-service", cfg, svcCfg)
			assert.Error(t, err)
			assert.Nil(t, capture)
		})
	}
}
