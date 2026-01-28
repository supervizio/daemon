package logging_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConfig struct {
	logPath string
}

func (m *mockConfig) GetServiceLogPath(serviceName, logFile string) string {
	return m.logPath + "/" + serviceName + "/" + logFile
}

type mockServiceLogging struct {
	stdout config.LogStreamConfig
	stderr config.LogStreamConfig
}

func (m *mockServiceLogging) StdoutConfig() *config.LogStreamConfig {
	return &m.stdout
}

func (m *mockServiceLogging) StderrConfig() *config.LogStreamConfig {
	return &m.stderr
}

func TestNewCapture(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
		stdoutFile  string
		stderrFile  string
		wantErr     bool
	}{
		{
			name:        "passthrough to os stdout/stderr",
			serviceName: "test",
			stdoutFile:  "",
			stderrFile:  "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfig{logPath: t.TempDir()}
			svcCfg := &mockServiceLogging{
				stdout: config.LogStreamConfig{FilePath: tt.stdoutFile},
				stderr: config.LogStreamConfig{FilePath: tt.stderrFile},
			}

			capture, err := logging.NewCapture(tt.serviceName, cfg, svcCfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, capture)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, capture)
				if capture != nil {
					_ = capture.Close()
				}
			}
		})
	}
}

func TestCapture_Stdout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "get stdout writer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfig{logPath: t.TempDir()}
			svcCfg := &mockServiceLogging{
				stdout: config.LogStreamConfig{},
				stderr: config.LogStreamConfig{},
			}

			capture, err := logging.NewCapture("test", cfg, svcCfg)
			require.NoError(t, err)
			defer capture.Close()

			writer := capture.Stdout()
			assert.NotNil(t, writer)
		})
	}
}

func TestCapture_Stderr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "get stderr writer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfig{logPath: t.TempDir()}
			svcCfg := &mockServiceLogging{
				stdout: config.LogStreamConfig{},
				stderr: config.LogStreamConfig{},
			}

			capture, err := logging.NewCapture("test", cfg, svcCfg)
			require.NoError(t, err)
			defer capture.Close()

			writer := capture.Stderr()
			assert.NotNil(t, writer)
		})
	}
}

func TestCapture_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "close capture once"},
		{name: "close capture multiple times"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := &mockConfig{logPath: t.TempDir()}
			svcCfg := &mockServiceLogging{
				stdout: config.LogStreamConfig{},
				stderr: config.LogStreamConfig{},
			}

			capture, err := logging.NewCapture("test", cfg, svcCfg)
			require.NoError(t, err)

			err = capture.Close()
			assert.NoError(t, err)

			// Second close should also succeed.
			err = capture.Close()
			assert.NoError(t, err)
		})
	}
}
