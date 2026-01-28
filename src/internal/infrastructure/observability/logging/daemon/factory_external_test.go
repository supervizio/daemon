package daemon_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/stretchr/testify/assert"
)

func TestBuildLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     config.DaemonLogging
		baseDir string
		wantErr bool
	}{
		{
			name:    "default config",
			cfg:     config.DaemonLogging{},
			baseDir: t.TempDir(),
			wantErr: false,
		},
		{
			name: "console writer",
			cfg: config.DaemonLogging{
				Writers: []config.WriterConfig{
					{Type: "console", Level: "info"},
				},
			},
			baseDir: t.TempDir(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, err := daemon.BuildLogger(tt.cfg, tt.baseDir)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				if logger != nil {
					_ = logger.Close()
				}
			}
		})
	}
}

func TestBuildLoggerWithoutConsole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     config.DaemonLogging
		baseDir string
		wantErr bool
	}{
		{
			name: "skip console writers",
			cfg: config.DaemonLogging{
				Writers: []config.WriterConfig{
					{Type: "console", Level: "info"},
				},
			},
			baseDir: t.TempDir(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, err := daemon.BuildLoggerWithoutConsole(tt.cfg, tt.baseDir)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				if logger != nil {
					_ = logger.Close()
				}
			}
		})
	}
}

func TestDefaultLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "create default logger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := daemon.DefaultLogger()
			assert.NotNil(t, logger)
			_ = logger.Close()
		})
	}
}

func TestNewSilentLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "create silent logger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := daemon.NewSilentLogger()
			assert.NotNil(t, logger)
			_ = logger.Close()
		})
	}
}

func TestBuildLoggerWithBufferedConsole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		cfg            config.DaemonLogging
		baseDir        string
		expectBuffered bool
		wantErr        bool
	}{
		{
			name: "with console writer",
			cfg: config.DaemonLogging{
				Writers: []config.WriterConfig{
					{Type: "console", Level: "info"},
				},
			},
			baseDir:        t.TempDir(),
			expectBuffered: true,
			wantErr:        false,
		},
		{
			name: "without console writer",
			cfg: config.DaemonLogging{
				Writers: []config.WriterConfig{},
			},
			baseDir:        t.TempDir(),
			expectBuffered: true,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger, buffered, err := daemon.BuildLoggerWithBufferedConsole(tt.cfg, tt.baseDir)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
				if tt.expectBuffered {
					assert.NotNil(t, buffered)
				} else {
					assert.Nil(t, buffered)
				}
				if logger != nil {
					_ = logger.Close()
				}
			}
		})
	}
}
