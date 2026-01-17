// Package shared_test provides external tests for the shared package.
package shared_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/shared"
)

// TestNewOSFileSystem tests the NewOSFileSystem constructor.
//
// Params:
//   - t: testing context.
func TestNewOSFileSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "returns_non_nil_instance"},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call NewOSFileSystem.
			fs := shared.NewOSFileSystem()

			// Verify result is non-nil.
			assert.NotNil(t, fs)
		})
	}
}

// TestOSFileSystem_Stat tests the OSFileSystem.Stat method.
//
// Params:
//   - t: testing context.
func TestOSFileSystem_Stat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		wantErr   bool
		checkInfo func(t *testing.T, info os.FileInfo)
	}{
		{
			name: "returns_file_info_for_existing_file",
			setup: func(t *testing.T) string {
				// Create a temporary file.
				dir := t.TempDir()
				path := filepath.Join(dir, "test.txt")
				require.NoError(t, os.WriteFile(path, []byte("content"), 0o644))
				return path
			},
			wantErr: false,
			checkInfo: func(t *testing.T, info os.FileInfo) {
				assert.Equal(t, "test.txt", info.Name())
				assert.False(t, info.IsDir())
			},
		},
		{
			name: "returns_file_info_for_directory",
			setup: func(t *testing.T) string {
				// Return temp directory path.
				return t.TempDir()
			},
			wantErr: false,
			checkInfo: func(t *testing.T, info os.FileInfo) {
				assert.True(t, info.IsDir())
			},
		},
		{
			name: "returns_error_for_nonexistent_path",
			setup: func(t *testing.T) string {
				// Return path that does not exist.
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantErr:   true,
			checkInfo: nil,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test case.
			path := tt.setup(t)
			fs := shared.NewOSFileSystem()

			// Call Stat.
			info, err := fs.Stat(path)

			// Verify expected result.
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				if tt.checkInfo != nil {
					tt.checkInfo(t, info)
				}
			}
		})
	}
}

// TestOSFileSystem_ReadFile tests the OSFileSystem.ReadFile method.
//
// Params:
//   - t: testing context.
func TestOSFileSystem_ReadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		wantErr     bool
		wantContent []byte
	}{
		{
			name: "reads_file_contents",
			setup: func(t *testing.T) string {
				// Create a file with known content.
				dir := t.TempDir()
				path := filepath.Join(dir, "test.txt")
				require.NoError(t, os.WriteFile(path, []byte("hello world"), 0o644))
				return path
			},
			wantErr:     false,
			wantContent: []byte("hello world"),
		},
		{
			name: "reads_empty_file",
			setup: func(t *testing.T) string {
				// Create an empty file.
				dir := t.TempDir()
				path := filepath.Join(dir, "empty.txt")
				require.NoError(t, os.WriteFile(path, []byte{}, 0o644))
				return path
			},
			wantErr:     false,
			wantContent: []byte{},
		},
		{
			name: "returns_error_for_nonexistent_file",
			setup: func(t *testing.T) string {
				// Return path that does not exist.
				return filepath.Join(t.TempDir(), "nonexistent.txt")
			},
			wantErr:     true,
			wantContent: nil,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test case.
			path := tt.setup(t)
			fs := shared.NewOSFileSystem()

			// Call ReadFile.
			content, err := fs.ReadFile(path)

			// Verify expected result.
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantContent, content)
			}
		})
	}
}

// TestDefaultFileSystem tests that DefaultFileSystem is properly initialized.
//
// Params:
//   - t: testing context.
func TestDefaultFileSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "is_non_nil"},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify DefaultFileSystem is initialized.
			assert.NotNil(t, shared.DefaultFileSystem)
		})
	}
}
