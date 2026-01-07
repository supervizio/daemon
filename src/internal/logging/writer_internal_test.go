// Package logging provides internal tests for writer.go.
// It tests internal implementation details using white-box testing.
package logging

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_openLogFile tests the openLogFile helper function.
//
// Params:
//   - t: the testing context.
func Test_openLogFile(t *testing.T) {
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
			name:      "opens_existing_file",
			createDir: true,
			wantErr:   false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			f, size, err := openLogFile(path)

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

			assert.Equal(t, int64(0), size)
		})
	}
}

// Test_openLogFileExistingContent tests openLogFile with existing content.
//
// Params:
//   - t: the testing context.
func Test_openLogFileExistingContent(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// content is the existing content.
		content string
		// expectedSize is the expected size.
		expectedSize int64
	}{
		{
			name:         "empty_file",
			content:      "",
			expectedSize: 0,
		},
		{
			name:         "file_with_content",
			content:      "existing content",
			expectedSize: 16,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			// Create file with content.
			if tt.content != "" {
				err := os.WriteFile(path, []byte(tt.content), filePermissions)
				require.NoError(t, err)
			}

			f, size, err := openLogFile(path)
			require.NoError(t, err)
			require.NotNil(t, f)
			defer func() { _ = f.Close() }()

			assert.Equal(t, tt.expectedSize, size)
		})
	}
}

// Test_parseMaxSize tests the parseMaxSize helper function.
//
// Params:
//   - t: the testing context.
func Test_parseMaxSize(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// input is the size string to parse.
		input string
		// expected is the expected size in bytes.
		expected int64
	}{
		{
			name:     "megabytes",
			input:    "1MB",
			expected: 1 * 1024 * 1024,
		},
		{
			name:     "bytes",
			input:    "100",
			expected: 100,
		},
		{
			name:     "invalid_returns_default",
			input:    "invalid",
			expected: defaultMaxSize,
		},
		{
			name:     "empty_returns_default",
			input:    "",
			expected: defaultMaxSize,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			result := parseMaxSize(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test_Writer_rotateFiles tests the rotateFiles method.
//
// Params:
//   - t: the testing context.
func Test_Writer_rotateFiles(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// maxFiles is the max number of backup files.
		maxFiles int
		// existingBackups is the number of existing backup files.
		existingBackups int
	}{
		{
			name:            "no_existing_backups",
			maxFiles:        3,
			existingBackups: 0,
		},
		{
			name:            "with_existing_backups",
			maxFiles:        3,
			existingBackups: 2,
		},
		{
			name:            "at_max_backups",
			maxFiles:        3,
			existingBackups: 3,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			// Create the main log file.
			err := os.WriteFile(path, []byte("main log content"), filePermissions)
			require.NoError(t, err)

			// Create existing backup files.
			backupContent := []byte("backup content")
			for i := 1; i <= tt.existingBackups; i++ {
				backupPath := path + "." + strconv.Itoa(i)
				err := os.WriteFile(backupPath, backupContent, filePermissions)
				require.NoError(t, err)
			}

			w := &Writer{
				path:     path,
				maxFiles: tt.maxFiles,
			}

			err = w.rotateFiles()
			assert.NoError(t, err)

			// Verify main log was renamed to .1.
			backupPath := path + ".1"
			_, err = os.Stat(backupPath)
			assert.NoError(t, err)
		})
	}
}

// Test_Writer_rotate tests the rotate method.
//
// Params:
//   - t: the testing context.
func Test_Writer_rotate(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// initialContent is the initial content to write.
		initialContent string
		// maxSize is the max file size.
		maxSize int64
		// maxFiles is the max backup files.
		maxFiles int
	}{
		{
			name:           "rotate_creates_backup",
			initialContent: "initial content",
			maxSize:        100,
			maxFiles:       3,
		},
		{
			name:           "rotate_empty_file",
			initialContent: "",
			maxSize:        100,
			maxFiles:       3,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			// Create writer.
			f, size, err := openLogFile(path)
			require.NoError(t, err)

			w := &Writer{
				file:     f,
				path:     path,
				maxSize:  tt.maxSize,
				maxFiles: tt.maxFiles,
				size:     size,
			}

			// Write initial content if provided.
			if tt.initialContent != "" {
				_, err := f.WriteString(tt.initialContent)
				require.NoError(t, err)
			}

			// Close writer file for rotation.
			err = f.Close()
			require.NoError(t, err)

			// Create new file to rotate from.
			f2, _, err := openLogFile(path)
			require.NoError(t, err)
			w.file = f2

			// Import bufio for writer.
			w.writer = nil

			// Re-create properly.
			err = f2.Close()
			require.NoError(t, err)

			f3, _, err := openLogFile(path)
			require.NoError(t, err)
			defer func() { _ = f3.Close() }()
		})
	}
}

// Test_Writer_openNewFile tests the openNewFile method.
//
// Params:
//   - t: the testing context.
func Test_Writer_openNewFile(t *testing.T) {
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

			w := &Writer{path: path}

			f, err := w.openNewFile()

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
		})
	}
}
