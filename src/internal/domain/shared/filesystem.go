// Package shared provides common value objects and interfaces for the domain layer.
package shared

import "os"

// FileSystem abstracts file system operations for testing.
// It allows filesystem-dependent code to be tested with mock implementations.
// Implementations can return fixed values for testing or use real OS calls.
type FileSystem interface {
	// Stat returns file info for the given path.
	Stat(name string) (os.FileInfo, error)
	// ReadFile reads the entire contents of the file.
	ReadFile(name string) ([]byte, error)
}

// OSFileSystem implements FileSystem using the real os package.
// It is a stateless implementation that delegates to os.Stat and os.ReadFile.
type OSFileSystem struct{}

// NewOSFileSystem creates a new OSFileSystem instance.
//
// Returns:
//   - *OSFileSystem: a new filesystem that uses real OS calls.
func NewOSFileSystem() *OSFileSystem {
	return &OSFileSystem{}
}

// Stat returns file info using os.Stat.
//
// Params:
//   - name: path to the file
//
// Returns:
//   - os.FileInfo: file information
//   - error: error if file cannot be stat'd
func (OSFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

// ReadFile reads file contents using os.ReadFile.
//
// Params:
//   - name: path to the file
//
// Returns:
//   - []byte: file contents
//   - error: error if file cannot be read
func (OSFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// DefaultFileSystem is the default filesystem instance using real OS calls.
var DefaultFileSystem FileSystem = &OSFileSystem{}
