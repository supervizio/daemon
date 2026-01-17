//go:build unix

// Package credentials_test provides black-box tests for the adapters package.
// It tests credential management functionality for Unix systems.
package credentials_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/process/credentials"
)

// TestNew tests the New constructor.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestNew(t *testing.T) {
	// Define test cases for New constructor.
	tests := []struct {
		name string
	}{
		{
			name: "returns non-nil manager instance",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			manager := credentials.New()
			// Check if the manager is not nil.
			assert.NotNil(t, manager, "New should return a non-nil instance")
		})
	}
}

// TestNewManager tests the NewManager constructor.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestNewManager(t *testing.T) {
	// Define test cases for NewManager constructor.
	tests := []struct {
		name string
	}{
		{
			name: "returns non-nil manager instance",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			manager := credentials.NewManager()
			// Check if the manager is not nil.
			assert.NotNil(t, manager, "NewManager should return a non-nil instance")
		})
	}
}

// TestManager_LookupUser tests the LookupUser method with various inputs.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestManager_LookupUser(t *testing.T) {
	// Define test cases for LookupUser.
	tests := []struct {
		name        string
		input       string
		expectError bool
		expectedUID uint32
		checkUID    bool
	}{
		{
			name:        "lookup root user by name",
			input:       "root",
			expectError: false,
			expectedUID: 0,
			checkUID:    true,
		},
		{
			name:        "lookup user by UID 0",
			input:       "0",
			expectError: false,
			expectedUID: 0,
			checkUID:    true,
		},
		{
			name:        "lookup nonexistent user",
			input:       "nonexistent_user_12345",
			expectError: true,
			checkUID:    false,
		},
	}

	manager := credentials.New()

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			user, err := manager.LookupUser(tt.input)
			// Check error expectation.
			if tt.expectError {
				// Check if an error was returned for nonexistent user.
				assert.Error(t, err, "should return error for nonexistent user")
			} else if err == nil && tt.checkUID {
				// Check if the user was found.
				assert.NotNil(t, user, "user should not be nil")
				assert.Equal(t, tt.expectedUID, user.UID, "UID should match expected value")
			}
		})
	}
}

// TestManager_LookupGroup tests the LookupGroup method with various inputs.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestManager_LookupGroup(t *testing.T) {
	// Define test cases for LookupGroup.
	tests := []struct {
		name        string
		input       string
		expectError bool
		expectedGID uint32
		checkGID    bool
	}{
		{
			name:        "lookup root group by name",
			input:       "root",
			expectError: false,
			expectedGID: 0,
			checkGID:    true,
		},
		{
			name:        "lookup group by GID 0",
			input:       "0",
			expectError: false,
			expectedGID: 0,
			checkGID:    true,
		},
		{
			name:        "lookup nonexistent group",
			input:       "nonexistent_group_12345",
			expectError: true,
			checkGID:    false,
		},
	}

	manager := credentials.New()

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			group, err := manager.LookupGroup(tt.input)
			// Check error expectation.
			if tt.expectError {
				// Check if an error was returned for nonexistent group.
				assert.Error(t, err, "should return error for nonexistent group")
			} else if err == nil && tt.checkGID {
				// Check if the group was found.
				assert.NotNil(t, group, "group should not be nil")
				assert.Equal(t, tt.expectedGID, group.GID, "GID should match expected value")
			}
		})
	}
}

// TestManager_ResolveCredentials tests the ResolveCredentials method with various inputs.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestManager_ResolveCredentials(t *testing.T) {
	// Define test cases for ResolveCredentials.
	tests := []struct {
		name        string
		user        string
		group       string
		expectError bool
		expectedUID uint32
		expectedGID uint32
		checkUID    bool
		checkGID    bool
	}{
		{
			name:        "resolve empty credentials",
			user:        "",
			group:       "",
			expectError: false,
			expectedUID: 0,
			expectedGID: 0,
			checkUID:    true,
			checkGID:    true,
		},
		{
			name:        "resolve numeric UID",
			user:        "1000",
			group:       "",
			expectError: false,
			expectedUID: 1000,
			checkUID:    true,
			checkGID:    false,
		},
		{
			name:        "resolve numeric GID",
			user:        "",
			group:       "1000",
			expectError: false,
			expectedUID: 0,
			expectedGID: 1000,
			checkUID:    true,
			checkGID:    true,
		},
		{
			name:        "resolve nonexistent user",
			user:        "nonexistent_user_12345",
			group:       "",
			expectError: true,
			checkUID:    false,
			checkGID:    false,
		},
		{
			name:        "resolve nonexistent group",
			user:        "",
			group:       "nonexistent_group_12345",
			expectError: true,
			checkUID:    false,
			checkGID:    false,
		},
		{
			name:        "resolve root user without group uses primary GID",
			user:        "root",
			group:       "",
			expectError: false,
			expectedUID: 0,
			expectedGID: 0,
			checkUID:    true,
			checkGID:    true,
		},
		{
			name:        "resolve root user with explicit group",
			user:        "root",
			group:       "root",
			expectError: false,
			expectedUID: 0,
			expectedGID: 0,
			checkUID:    true,
			checkGID:    true,
		},
		{
			name:        "resolve nonexistent numeric UID as fallback",
			user:        "4294967294",
			group:       "",
			expectError: false,
			expectedUID: 4294967294,
			expectedGID: 0,
			checkUID:    true,
			checkGID:    true,
		},
		{
			name:        "resolve nonexistent numeric GID as fallback",
			user:        "",
			group:       "4294967294",
			expectError: false,
			expectedUID: 0,
			expectedGID: 4294967294,
			checkUID:    true,
			checkGID:    true,
		},
	}

	manager := credentials.New()

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			uid, gid, err := manager.ResolveCredentials(tt.user, tt.group)
			// Check error expectation.
			if tt.expectError {
				// Check if an error was returned.
				assert.Error(t, err, "should return error")
			} else {
				// Check if no error occurred.
				require.NoError(t, err, "should not return error")
				// Check UID if needed.
				if tt.checkUID {
					assert.Equal(t, tt.expectedUID, uid, "uid should match expected value")
				}
				// Check GID if needed.
				if tt.checkGID {
					assert.Equal(t, tt.expectedGID, gid, "gid should match expected value")
				}
			}
		})
	}
}

// TestManager_ApplyCredentials tests the ApplyCredentials method with various inputs.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestManager_ApplyCredentials(t *testing.T) {
	// Define test cases for ApplyCredentials.
	tests := []struct {
		name             string
		uid              uint32
		gid              uint32
		expectNilSysAttr bool
		expectedUID      uint32
		expectedGID      uint32
	}{
		{
			name:             "apply zero credentials",
			uid:              0,
			gid:              0,
			expectNilSysAttr: true,
		},
		{
			name:             "apply non-zero credentials",
			uid:              1000,
			gid:              1000,
			expectNilSysAttr: false,
			expectedUID:      1000,
			expectedGID:      1000,
		},
		{
			name:             "apply credentials with only UID",
			uid:              1000,
			gid:              0,
			expectNilSysAttr: false,
			expectedUID:      1000,
			expectedGID:      0,
		},
	}

	manager := credentials.New()

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("echo", "test")
			err := manager.ApplyCredentials(cmd, tt.uid, tt.gid)
			// Check if no error occurred.
			require.NoError(t, err, "should not return error")
			// Check SysProcAttr based on expectations.
			if tt.expectNilSysAttr {
				// Check if SysProcAttr was not set.
				assert.Nil(t, cmd.SysProcAttr, "SysProcAttr should be nil for zero credentials")
			} else {
				// Check if SysProcAttr was set.
				require.NotNil(t, cmd.SysProcAttr, "SysProcAttr should not be nil")
				// Check if Credential was set correctly.
				require.NotNil(t, cmd.SysProcAttr.Credential, "Credential should not be nil")
				// Check if UID and GID were set correctly.
				assert.Equal(t, tt.expectedUID, cmd.SysProcAttr.Credential.Uid, "UID should match expected value")
				assert.Equal(t, tt.expectedGID, cmd.SysProcAttr.Credential.Gid, "GID should match expected value")
			}
		})
	}
}
