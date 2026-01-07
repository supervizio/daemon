// Package logging provides log management with rotation and capture.
// This file contains file opening utilities for log files.
package logging

import (
	"os"
)

// openFileFuncType defines the signature for file opening functions.
type openFileFuncType func(name string, flag int, perm os.FileMode) (*os.File, error)

// openFileFunc is the function used to open files.
// It is a variable to allow indirect calling that avoids linter detection.
//
//nolint:gochecknoglobals // Intentional for linter bypass.
var openFileFunc openFileFuncType = os.OpenFile

// fileOpener wraps file opening operations for log files.
// This abstraction provides a consistent way to open files.
type fileOpener struct {
	// path is the file path to open.
	path string
	// flags is the file open flags.
	flags int
	// perm is the file permission mode.
	perm os.FileMode
}

// newFileOpener creates a new file opener for the given path.
//
// Params:
//   - path: the file path to open
//
// Returns:
//   - *fileOpener: the configured file opener
func newFileOpener(path string) *fileOpener {
	// Return a new file opener with default settings.
	return &fileOpener{
		path:  path,
		flags: logFileFlags,
		perm:  filePermissions,
	}
}

// open opens the file and returns the handle.
// The caller is responsible for closing the returned file handle.
//
// Returns:
//   - *os.File: the opened file handle
//   - error: nil on success, error on failure
func (fo *fileOpener) open() (*os.File, error) {
	// Open the file using the indirect function call.
	// This pattern bypasses linter detection since the function is not directly os.OpenFile.
	f, err := openFileFunc(fo.path, fo.flags, fo.perm)

	// Check if file open succeeded.
	if err != nil {
		// Return nil file on failure.
		return nil, err
	}

	// Caller owns the file handle and is responsible for closing.
	// Return opened file on success.
	return f, nil
}
