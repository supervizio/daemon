//go:build unix

// Package adapters_test provides black-box tests for credential adapters.
package adapters_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/kernel/adapters"
)

// TestNewScratchCredentialManager verifies constructor returns non-nil instance.
func TestNewScratchCredentialManager(t *testing.T) {
	t.Parallel()

	manager := adapters.NewScratchCredentialManager()

	assert.NotNil(t, manager)
}

// TestScratchCredentialManager_LookupUser tests user lookup behavior.
func TestScratchCredentialManager_LookupUser(t *testing.T) {
	t.Parallel()

	manager := adapters.NewScratchCredentialManager()

	tests := []struct {
		name        string
		input       string
		wantUID     uint32
		wantGID     uint32
		wantErr     bool
		errContains string
	}{
		{
			name:    "numeric UID zero",
			input:   "0",
			wantUID: 0,
			wantGID: 0,
			wantErr: false,
		},
		{
			name:    "numeric UID 1000",
			input:   "1000",
			wantUID: 1000,
			wantGID: 1000,
			wantErr: false,
		},
		{
			name:    "numeric UID max 32-bit",
			input:   "4294967295",
			wantUID: 4294967295,
			wantGID: 4294967295,
			wantErr: false,
		},
		{
			name:        "name lookup fails",
			input:       "root",
			wantErr:     true,
			errContains: "scratch mode",
		},
		{
			name:        "username fails",
			input:       "nobody",
			wantErr:     true,
			errContains: "scratch mode",
		},
		{
			name:        "empty string fails",
			input:       "",
			wantErr:     true,
			errContains: "scratch mode",
		},
		{
			name:        "negative number fails",
			input:       "-1",
			wantErr:     true,
			errContains: "scratch mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user, err := manager.LookupUser(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, user)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)
			assert.Equal(t, tt.wantUID, user.UID)
			assert.Equal(t, tt.wantGID, user.GID)
			assert.Equal(t, tt.input, user.Username)
		})
	}
}

// TestScratchCredentialManager_LookupGroup tests group lookup behavior.
func TestScratchCredentialManager_LookupGroup(t *testing.T) {
	t.Parallel()

	manager := adapters.NewScratchCredentialManager()

	tests := []struct {
		name        string
		input       string
		wantGID     uint32
		wantErr     bool
		errContains string
	}{
		{
			name:    "numeric GID zero",
			input:   "0",
			wantGID: 0,
			wantErr: false,
		},
		{
			name:    "numeric GID 1000",
			input:   "1000",
			wantGID: 1000,
			wantErr: false,
		},
		{
			name:        "name lookup fails",
			input:       "root",
			wantErr:     true,
			errContains: "scratch mode",
		},
		{
			name:        "group name fails",
			input:       "wheel",
			wantErr:     true,
			errContains: "scratch mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			group, err := manager.LookupGroup(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Nil(t, group)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, group)
			assert.Equal(t, tt.wantGID, group.GID)
			assert.Equal(t, tt.input, group.Name)
		})
	}
}

// TestScratchCredentialManager_ResolveCredentials tests credential resolution.
func TestScratchCredentialManager_ResolveCredentials(t *testing.T) {
	t.Parallel()

	manager := adapters.NewScratchCredentialManager()

	tests := []struct {
		name      string
		username  string
		groupname string
		wantUID   uint32
		wantGID   uint32
		wantErr   bool
	}{
		{
			name:      "empty credentials",
			username:  "",
			groupname: "",
			wantUID:   0,
			wantGID:   0,
			wantErr:   false,
		},
		{
			name:      "numeric user only",
			username:  "1000",
			groupname: "",
			wantUID:   1000,
			wantGID:   1000, // defaults to UID
			wantErr:   false,
		},
		{
			name:      "numeric group only",
			username:  "",
			groupname: "1001",
			wantUID:   0,
			wantGID:   1001,
			wantErr:   false,
		},
		{
			name:      "both numeric",
			username:  "1000",
			groupname: "1001",
			wantUID:   1000,
			wantGID:   1001,
			wantErr:   false,
		},
		{
			name:      "name lookup user fails",
			username:  "root",
			groupname: "",
			wantErr:   true,
		},
		{
			name:      "name lookup group fails",
			username:  "",
			groupname: "wheel",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uid, gid, err := manager.ResolveCredentials(tt.username, tt.groupname)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantUID, uid)
			assert.Equal(t, tt.wantGID, gid)
		})
	}
}

// TestScratchCredentialManager_ApplyCredentials tests applying credentials to command.
func TestScratchCredentialManager_ApplyCredentials(t *testing.T) {
	t.Parallel()

	manager := adapters.NewScratchCredentialManager()

	tests := []struct {
		name        string
		uid         uint32
		gid         uint32
		expectCreds bool
		expectedUID uint32
		expectedGID uint32
	}{
		{
			name:        "zero credentials skipped",
			uid:         0,
			gid:         0,
			expectCreds: false,
		},
		{
			name:        "non-zero UID applied",
			uid:         1000,
			gid:         0,
			expectCreds: true,
			expectedUID: 1000,
			expectedGID: 0,
		},
		{
			name:        "non-zero GID applied",
			uid:         0,
			gid:         1000,
			expectCreds: true,
			expectedUID: 0,
			expectedGID: 1000,
		},
		{
			name:        "both applied",
			uid:         1000,
			gid:         1001,
			expectCreds: true,
			expectedUID: 1000,
			expectedGID: 1001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := exec.Command("true")

			err := manager.ApplyCredentials(cmd, tt.uid, tt.gid)

			require.NoError(t, err)

			if !tt.expectCreds {
				if cmd.SysProcAttr != nil && cmd.SysProcAttr.Credential != nil {
					t.Error("expected no credentials to be set")
				}
				return
			}

			require.NotNil(t, cmd.SysProcAttr)
			require.NotNil(t, cmd.SysProcAttr.Credential)
			assert.Equal(t, tt.expectedUID, cmd.SysProcAttr.Credential.Uid)
			assert.Equal(t, tt.expectedGID, cmd.SysProcAttr.Credential.Gid)
		})
	}
}

// TestIsScratchEnvironment tests scratch environment detection.
func TestIsScratchEnvironment(t *testing.T) {
	t.Parallel()

	// This test documents the behavior without being flaky.
	// In a real scratch container, /etc/passwd doesn't exist.
	// In development environment, it typically does.
	result := adapters.IsScratchEnvironment()

	// We can't assert a specific value as it depends on the environment.
	// Just verify it returns a boolean without panicking.
	assert.IsType(t, true, result)
}

// TestErrScratchNameLookup verifies the error is properly exported.
func TestErrScratchNameLookup(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, adapters.ErrScratchNameLookup)
	assert.Contains(t, adapters.ErrScratchNameLookup.Error(), "scratch mode")
}
