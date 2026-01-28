package daemon

import (
	"path/filepath"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/stretchr/testify/assert"
)

func TestBuildWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     config.WriterConfig
		baseDir string
		wantErr bool
	}{
		{
			name:    "console writer",
			cfg:     config.WriterConfig{Type: "console"},
			baseDir: "",
			wantErr: false,
		},
		{
			name:    "file writer missing path",
			cfg:     config.WriterConfig{Type: "file"},
			baseDir: "",
			wantErr: true,
		},
		{
			name:    "json writer missing path",
			cfg:     config.WriterConfig{Type: "json"},
			baseDir: "",
			wantErr: true,
		},
		{
			name:    "unknown writer type",
			cfg:     config.WriterConfig{Type: "unknown"},
			baseDir: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer, err := buildWriter(tt.cfg, tt.baseDir)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, writer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, writer)
				if writer != nil {
					_ = writer.Close()
				}
			}
		})
	}
}

func TestResolvePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    string
		baseDir string
		wantErr bool
	}{
		{
			name:    "absolute path",
			path:    "/var/log/daemon.log",
			baseDir: "/tmp",
			wantErr: false,
		},
		{
			name:    "relative path with baseDir",
			path:    "daemon.log",
			baseDir: "/var/log",
			wantErr: false,
		},
		{
			name:    "relative path without baseDir",
			path:    "daemon.log",
			baseDir: "",
			wantErr: false,
		},
		{
			name:    "path escaping baseDir",
			path:    "../../../etc/passwd",
			baseDir: "/var/log/daemon",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resolved, err := resolvePath(tt.path, tt.baseDir)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, resolved)
			} else {
				assert.NoError(t, err)
				if filepath.IsAbs(tt.path) {
					assert.Equal(t, tt.path, resolved)
				}
			}
		})
	}
}
