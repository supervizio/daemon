// Package collector provides internal tests for sandbox.go.
// It tests internal implementation details using white-box testing.
package collector

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_SandboxCollector_detectSandbox tests the detectSandbox method.
// It verifies that sandbox detection works correctly for various scenarios.
//
// Params:
//   - t: the testing context.
func Test_SandboxCollector_detectSandbox(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// setupEndpoints creates test endpoints and returns them.
		setupEndpoints func(t *testing.T) []string
		// checkName is the sandbox name.
		checkName string
		// wantDetected indicates if sandbox should be detected.
		wantDetected bool
	}{
		{
			name: "detects_existing_single_endpoint",
			setupEndpoints: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				tmpPath := filepath.Join(tmpDir, "sandbox_test")
				err := os.WriteFile(tmpPath, []byte{}, 0600)
				assert.NoError(t, err)
				return []string{tmpPath}
			},
			checkName:    "Test1",
			wantDetected: true,
		},
		{
			name: "not_detected_nonexistent_endpoint",
			setupEndpoints: func(t *testing.T) []string {
				return []string{"/nonexistent/path/to/socket"}
			},
			checkName:    "Test2",
			wantDetected: false,
		},
		{
			name: "detects_first_existing_endpoint",
			setupEndpoints: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				tmpPath1 := filepath.Join(tmpDir, "sandbox_test1")
				err := os.WriteFile(tmpPath1, []byte{}, 0600)
				assert.NoError(t, err)

				tmpPath2 := filepath.Join(tmpDir, "sandbox_test2")
				err = os.WriteFile(tmpPath2, []byte{}, 0600)
				assert.NoError(t, err)

				return []string{tmpPath1, tmpPath2}
			},
			checkName:    "Test3",
			wantDetected: true,
		},
		{
			name: "detects_second_endpoint_when_first_missing",
			setupEndpoints: func(t *testing.T) []string {
				tmpDir := t.TempDir()
				tmpPath := filepath.Join(tmpDir, "sandbox_test")
				err := os.WriteFile(tmpPath, []byte{}, 0600)
				assert.NoError(t, err)

				return []string{"/nonexistent/first/path", tmpPath}
			},
			checkName:    "Test4",
			wantDetected: true,
		},
		{
			name: "not_detected_all_endpoints_missing",
			setupEndpoints: func(t *testing.T) []string {
				return []string{
					"/nonexistent/first/path",
					"/nonexistent/second/path",
					"/nonexistent/third/path",
				}
			},
			checkName:    "Test5",
			wantDetected: false,
		},
		{
			name: "not_detected_empty_endpoints_list",
			setupEndpoints: func(t *testing.T) []string {
				return []string{}
			},
			checkName:    "Test6",
			wantDetected: false,
		},
		{
			name: "not_detected_nil_endpoints",
			setupEndpoints: func(t *testing.T) []string {
				return nil
			},
			checkName:    "Test7",
			wantDetected: false,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup test endpoints.
			endpoints := tt.setupEndpoints(t)

			// Create check.
			check := sandboxCheck{
				name:      tt.checkName,
				endpoints: endpoints,
			}

			// Create collector.
			c := NewSandboxCollector()

			// Call the method.
			result := c.detectSandbox(check)

			// Verify result name matches check name.
			assert.Equal(t, tt.checkName, result.Name)

			// Verify detection result.
			assert.Equal(t, tt.wantDetected, result.Detected)

			// Verify endpoint if detected.
			if tt.wantDetected {
				assert.NotEmpty(t, result.Endpoint)
				// Endpoint should be one of the provided endpoints.
				assert.Contains(t, endpoints, result.Endpoint)
			} else {
				assert.Empty(t, result.Endpoint)
			}
		})
	}
}

// Test_SandboxCollector_detectSandbox_stops_after_first_match tests that detectSandbox
// stops checking after finding the first matching endpoint.
//
// Params:
//   - t: the testing context.
func Test_SandboxCollector_detectSandbox_stops_after_first_match(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// setupEndpoints creates test endpoints and returns them with expected index.
		setupEndpoints func(t *testing.T) ([]string, int)
	}{
		{
			name: "returns_first_match_not_second",
			setupEndpoints: func(t *testing.T) ([]string, int) {
				tmpDir := t.TempDir()
				tmpPath1 := filepath.Join(tmpDir, "sandbox_first")
				err := os.WriteFile(tmpPath1, []byte{}, 0600)
				assert.NoError(t, err)

				tmpPath2 := filepath.Join(tmpDir, "sandbox_second")
				err = os.WriteFile(tmpPath2, []byte{}, 0600)
				assert.NoError(t, err)

				return []string{tmpPath1, tmpPath2}, 0
			},
		},
		{
			name: "skips_nonexistent_returns_first_existing",
			setupEndpoints: func(t *testing.T) ([]string, int) {
				tmpDir := t.TempDir()
				tmpPath1 := filepath.Join(tmpDir, "sandbox_first")
				err := os.WriteFile(tmpPath1, []byte{}, 0600)
				assert.NoError(t, err)

				tmpPath2 := filepath.Join(tmpDir, "sandbox_second")
				err = os.WriteFile(tmpPath2, []byte{}, 0600)
				assert.NoError(t, err)

				return []string{"/nonexistent/path", tmpPath1, tmpPath2}, 1
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Setup endpoints.
			endpoints, expectedIdx := tt.setupEndpoints(t)

			// Create collector.
			c := NewSandboxCollector()

			// Create check.
			check := sandboxCheck{
				name:      "MultiEndpoint",
				endpoints: endpoints,
			}

			// Call the method.
			result := c.detectSandbox(check)

			// Verify detection succeeded.
			assert.True(t, result.Detected)

			// Verify it returned the expected matching endpoint.
			assert.Equal(t, endpoints[expectedIdx], result.Endpoint)
		})
	}
}

// Test_sandboxCheck tests the sandboxCheck struct.
// It verifies that the struct fields work correctly.
//
// Params:
//   - t: the testing context.
func Test_sandboxCheck(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// checkName is the sandbox name.
		checkName string
		// endpoints is the list of endpoints.
		endpoints []string
	}{
		{
			name:      "docker_check",
			checkName: "Docker",
			endpoints: []string{"/var/run/docker.sock"},
		},
		{
			name:      "podman_check",
			checkName: "Podman",
			endpoints: []string{"/var/run/podman/podman.sock", "/run/podman/podman.sock"},
		},
		{
			name:      "kubernetes_check",
			checkName: "Kubernetes",
			endpoints: []string{"/var/run/secrets/kubernetes.io/serviceaccount/token"},
		},
		{
			name:      "empty_endpoints",
			checkName: "Empty",
			endpoints: []string{},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create check struct.
			check := sandboxCheck{
				name:      tt.checkName,
				endpoints: tt.endpoints,
			}

			// Verify fields.
			assert.Equal(t, tt.checkName, check.name)
			assert.Equal(t, tt.endpoints, check.endpoints)
		})
	}
}

// Test_getSandboxChecks tests the getSandboxChecks function.
// It verifies that all expected sandbox checks are returned.
//
// Params:
//   - t: the testing context.
func Test_getSandboxChecks(t *testing.T) {
	t.Parallel()

	// Define test cases for table-driven testing.
	tests := []struct {
		// name is the test case name.
		name string
		// wantMinChecks is the minimum expected number of checks.
		wantMinChecks int
		// wantNames is the list of expected sandbox names.
		wantNames []string
	}{
		{
			name:          "returns_all_sandbox_checks",
			wantMinChecks: 6,
			wantNames:     []string{"Docker", "Podman", "containerd", "Kubernetes", "LXC/LXD", "CRI-O"},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Call the function.
			checks := getSandboxChecks()

			// Verify minimum number of checks.
			assert.GreaterOrEqual(t, len(checks), tt.wantMinChecks)

			// Verify each check has non-empty name and at least one endpoint.
			for _, check := range checks {
				assert.NotEmpty(t, check.name)
				assert.NotEmpty(t, check.endpoints)
			}

			// Verify expected names are present.
			names := make([]string, len(checks))
			for i, check := range checks {
				names[i] = check.name
			}
			for _, wantName := range tt.wantNames {
				assert.Contains(t, names, wantName)
			}
		})
	}
}
