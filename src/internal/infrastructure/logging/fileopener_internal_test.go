// Package logging provides internal tests for fileopener.go.
// It tests internal implementation details using white-box testing.
package logging

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_newFileOpener tests the newFileOpener constructor.
//
// Params:
//   - t: the testing context.
func Test_newFileOpener(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// path is the file path to use.
		path string
	}{
		{
			name: "simple_path",
			path: "/tmp/test.log",
		},
		{
			name: "nested_path",
			path: "/var/log/app/test.log",
		},
		{
			name: "relative_path",
			path: "test.log",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			fo := newFileOpener(tt.path)

			assert.NotNil(t, fo)
			assert.Equal(t, tt.path, fo.path)
			assert.Equal(t, logFileFlags, fo.flags)
			assert.Equal(t, filePermissions, fo.perm)
		})
	}
}

// Test_fileOpener_open tests the open method.
//
// Params:
//   - t: the testing context.
func Test_fileOpener_open(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// createDir indicates whether to create the directory.
		createDir bool
		// wantErr indicates whether an error is expected.
		wantErr bool
	}{
		{
			name:      "opens_new_file",
			createDir: true,
			wantErr:   false,
		},
		{
			name:      "fails_without_directory",
			createDir: false,
			wantErr:   true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			var path string

			// Setup path based on directory creation.
			if tt.createDir {
				path = filepath.Join(tmpDir, "test.log")
			} else {
				path = filepath.Join(tmpDir, "nonexistent", "test.log")
			}

			fo := newFileOpener(path)
			f, err := fo.open()

			// Check error expectation.
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, f)
				// Return early.
				return
			}

			require.NoError(t, err)
			require.NotNil(t, f)
			defer func() { _ = f.Close() }()

			// Verify file was created.
			_, statErr := os.Stat(path)
			assert.NoError(t, statErr)
		})
	}
}

// Test_fileOpener_openExistingFile tests opening an existing file.
//
// Params:
//   - t: the testing context.
func Test_fileOpener_openExistingFile(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// content is the existing content.
		content string
	}{
		{
			name:    "empty_file",
			content: "",
		},
		{
			name:    "file_with_content",
			content: "existing content\n",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			// Create file with content if provided.
			if tt.content != "" {
				err := os.WriteFile(path, []byte(tt.content), filePermissions)
				require.NoError(t, err)
			}

			fo := newFileOpener(path)
			f, err := fo.open()

			require.NoError(t, err)
			require.NotNil(t, f)
			defer func() { _ = f.Close() }()

			// Write additional content.
			testData := []byte("new content\n")
			n, err := f.Write(testData)
			require.NoError(t, err)
			assert.Equal(t, len(testData), n)
		})
	}
}
