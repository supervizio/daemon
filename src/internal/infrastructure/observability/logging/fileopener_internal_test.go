package logging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileOpener(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{
			name: "create file opener",
			path: "/tmp/test.log",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			opener := newFileOpener(tt.path)
			assert.NotNil(t, opener)
			assert.Equal(t, tt.path, opener.path)
			assert.Equal(t, logFileFlags, opener.flags)
			assert.Equal(t, filePermissions, opener.perm)
		})
	}
}

func TestFileOpener_Open(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "open new file",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			opener := newFileOpener(path)
			file, err := opener.open()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, file)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, file)
				if file != nil {
					_ = file.Close()
				}
			}
		})
	}
}

func TestOpenFileFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "verify openFileFunc is set"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.NotNil(t, openFileFunc)

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			file, err := openFileFunc(path, os.O_CREATE|os.O_WRONLY, 0o600)
			require.NoError(t, err)
			assert.NotNil(t, file)
			_ = file.Close()
		})
	}
}
