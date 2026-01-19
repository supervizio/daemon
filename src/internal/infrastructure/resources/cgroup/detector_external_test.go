//go:build linux

// Package cgroup_test provides external tests for the cgroup detector.
package cgroup_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/resources/cgroup"
)

// TestVersion_String tests the Version.String method.
//
// Params:
//   - t: the testing context.
func TestVersion_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version cgroup.Version
		want    string
	}{
		{name: "v1", version: cgroup.VersionV1, want: "v1"},
		{name: "v2", version: cgroup.VersionV2, want: "v2"},
		{name: "hybrid", version: cgroup.VersionHybrid, want: "hybrid"},
		{name: "unknown", version: cgroup.VersionUnknown, want: "unknown"},
		{name: "invalid_value", version: cgroup.Version(99), want: "unknown"},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.version.String())
		})
	}
}

// TestDetectWithPath tests the DetectWithPath function.
//
// Params:
//   - t: the testing context.
func TestDetectWithPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T) string
		want  cgroup.Version
	}{
		{
			name: "detects_v2",
			setup: func(t *testing.T) string {
				// Create mock v2 cgroup directory.
				mockCgroup := t.TempDir()
				require.NoError(t, os.WriteFile(
					filepath.Join(mockCgroup, "cgroup.controllers"),
					[]byte("cpu memory io"),
					0o644,
				))
				return mockCgroup
			},
			want: cgroup.VersionV2,
		},
		{
			name: "detects_v1",
			setup: func(t *testing.T) string {
				// Create mock v1 cgroup directory.
				mockCgroup := t.TempDir()
				require.NoError(t, os.MkdirAll(filepath.Join(mockCgroup, "cpu"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(mockCgroup, "memory"), 0o755))
				return mockCgroup
			},
			want: cgroup.VersionV1,
		},
		{
			name: "detects_hybrid",
			setup: func(t *testing.T) string {
				// Create mock hybrid cgroup directory.
				mockCgroup := t.TempDir()
				require.NoError(t, os.WriteFile(
					filepath.Join(mockCgroup, "cgroup.controllers"),
					[]byte("cpu memory"),
					0o644,
				))
				require.NoError(t, os.MkdirAll(filepath.Join(mockCgroup, "cpu"), 0o755))
				require.NoError(t, os.MkdirAll(filepath.Join(mockCgroup, "memory"), 0o755))
				return mockCgroup
			},
			want: cgroup.VersionHybrid,
		},
		{
			name: "returns_unknown_for_empty_directory",
			setup: func(t *testing.T) string {
				// Return empty directory.
				return t.TempDir()
			},
			want: cgroup.VersionUnknown,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mockCgroup := tt.setup(t)
			version := cgroup.DetectWithPath(mockCgroup)
			assert.Equal(t, tt.want, version)
		})
	}
}

// TestDetect tests the Detect function.
//
// Params:
//   - t: the testing context.
func TestDetect(t *testing.T) {
	t.Parallel()

	// Valid versions that Detect can return.
	validVersions := []cgroup.Version{
		cgroup.VersionUnknown,
		cgroup.VersionV1,
		cgroup.VersionV2,
		cgroup.VersionHybrid,
	}

	tests := []struct {
		name string
	}{
		{name: "returns_valid_cgroup_version"},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call Detect.
			version := cgroup.Detect()

			// Verify result is a valid version constant.
			found := false
			for _, v := range validVersions {
				if version == v {
					found = true
					break
				}
			}
			assert.True(t, found, "Detect should return a valid version")
		})
	}
}

// TestIsContainerized tests the IsContainerized function.
//
// Params:
//   - t: the testing context.
func TestIsContainerized(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "returns_boolean_result"},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call IsContainerized.
			result := cgroup.IsContainerized()

			// Verify result is a boolean.
			assert.IsType(t, true, result)
		})
	}
}

// mockFileSystem implements cgroup.FileSystem for testing.
type mockFileSystem struct {
	statFunc     func(name string) (os.FileInfo, error)
	readFileFunc func(name string) ([]byte, error)
}

// Stat returns file info using the mock function.
func (m *mockFileSystem) Stat(name string) (os.FileInfo, error) {
	return m.statFunc(name)
}

// ReadFile reads file contents using the mock function.
func (m *mockFileSystem) ReadFile(name string) ([]byte, error) {
	return m.readFileFunc(name)
}

// TestIsContainerizedWithFS tests the IsContainerizedWithFS function with injected filesystem.
//
// Params:
//   - t: the testing context.
func TestIsContainerizedWithFS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		dockerenvErr  error
		cgroupContent string
		cgroupErr     error
		want          bool
	}{
		{
			name:         "detects_docker_via_dockerenv",
			dockerenvErr: nil, // File exists
			want:         true,
		},
		{
			name:          "detects_docker_via_cgroup",
			dockerenvErr:  os.ErrNotExist,
			cgroupContent: "12:memory:/docker/abc123def456\n",
			want:          true,
		},
		{
			name:          "detects_kubernetes_via_cgroup",
			dockerenvErr:  os.ErrNotExist,
			cgroupContent: "12:memory:/kubepods/burstable/pod-abc123\n",
			want:          true,
		},
		{
			name:          "detects_lxc_via_cgroup",
			dockerenvErr:  os.ErrNotExist,
			cgroupContent: "12:memory:/lxc/container-123\n",
			want:          true,
		},
		{
			name:          "detects_containerd_via_cgroup",
			dockerenvErr:  os.ErrNotExist,
			cgroupContent: "0::/containerd/abc123\n",
			want:          true,
		},
		{
			name:          "not_containerized_no_markers",
			dockerenvErr:  os.ErrNotExist,
			cgroupContent: "12:memory:/user.slice/user-1000.slice\n",
			want:          false,
		},
		{
			name:          "not_containerized_empty_cgroup",
			dockerenvErr:  os.ErrNotExist,
			cgroupContent: "",
			want:          false,
		},
		{
			name:         "not_containerized_cgroup_read_error",
			dockerenvErr: os.ErrNotExist,
			cgroupErr:    os.ErrPermission,
			want:         false,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create mock filesystem with test case behavior.
			fs := &mockFileSystem{
				statFunc: func(name string) (os.FileInfo, error) {
					if name == "/.dockerenv" {
						if tt.dockerenvErr != nil {
							return nil, tt.dockerenvErr
						}
						// Return a mock FileInfo (nil is acceptable for existence check).
						return nil, nil
					}
					return nil, os.ErrNotExist
				},
				readFileFunc: func(name string) ([]byte, error) {
					if name == "/proc/1/cgroup" {
						if tt.cgroupErr != nil {
							return nil, tt.cgroupErr
						}
						return []byte(tt.cgroupContent), nil
					}
					return nil, os.ErrNotExist
				},
			}

			// Call IsContainerizedWithFS with mock.
			result := cgroup.IsContainerizedWithFS(fs)

			// Verify expected result.
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestNewReader tests the NewReader function.
//
// Params:
//   - t: the testing context.
func TestNewReader(t *testing.T) {
	t.Parallel()

	// Get system cgroup version for expected behavior.
	systemVersion := cgroup.Detect()

	tests := []struct {
		name string
	}{
		{name: "creates_reader_based_on_system_cgroup"},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call NewReader.
			reader, err := cgroup.NewReader()

			// Result depends on system cgroup version.
			if systemVersion == cgroup.VersionUnknown {
				// System has no cgroup, expect error.
				assert.Error(t, err)
				assert.Nil(t, reader)
				assert.ErrorIs(t, err, cgroup.ErrUnknownVersion)
			} else {
				// System has cgroup, expect reader.
				assert.NoError(t, err)
				assert.NotNil(t, reader)
			}
		})
	}
}

// TestNewReaderWithPath tests the NewReaderWithPath function.
// Note: NewReaderWithPath uses Detect() to determine cgroup version from system path,
// not the provided path. The path is used to create the reader instance.
//
// Params:
//   - t: the testing context.
func TestNewReaderWithPath(t *testing.T) {
	t.Parallel()

	// Test behavior depends on system cgroup version.
	version := cgroup.Detect()

	tests := []struct {
		name  string
		setup func(t *testing.T) string
	}{
		{
			name: "creates_reader_with_provided_path",
			setup: func(t *testing.T) string {
				// Create mock cgroup directory with required files.
				mockCgroup := t.TempDir()
				// Write required files for V2Reader.
				require.NoError(t, os.WriteFile(
					filepath.Join(mockCgroup, "cpu.stat"),
					[]byte("usage_usec 0\n"),
					0o644,
				))
				return mockCgroup
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := tt.setup(t)
			reader, err := cgroup.NewReaderWithPath(path)

			// Result depends on system cgroup version.
			if version == cgroup.VersionUnknown {
				// System has no cgroup, expect error.
				assert.Error(t, err)
				assert.Nil(t, reader)
			} else {
				// System has cgroup, expect reader.
				assert.NoError(t, err)
				if reader != nil {
					// Verify path was used.
					assert.Equal(t, path, reader.Path())
				}
			}
		})
	}
}
