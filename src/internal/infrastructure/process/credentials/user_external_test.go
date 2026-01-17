// Package credentials_test provides black-box tests for credentials package.
package credentials_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/infrastructure/process/credentials"
)

// TestUser_Fields tests the User struct fields.
//
// Params:
//   - t: the testing context
func TestUser_Fields(t *testing.T) {
	tests := []struct {
		name     string
		uid      uint32
		gid      uint32
		username string
		homeDir  string
	}{
		{
			name:     "basic user",
			uid:      1000,
			gid:      1000,
			username: "testuser",
			homeDir:  "/home/testuser",
		},
		{
			name:     "root user",
			uid:      0,
			gid:      0,
			username: "root",
			homeDir:  "/root",
		},
		{
			name:     "empty fields",
			uid:      0,
			gid:      0,
			username: "",
			homeDir:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := credentials.User{
				UID:      tt.uid,
				GID:      tt.gid,
				Username: tt.username,
				HomeDir:  tt.homeDir,
			}

			// Verify all fields are accessible
			assert.Equal(t, tt.uid, user.UID)
			assert.Equal(t, tt.gid, user.GID)
			assert.Equal(t, tt.username, user.Username)
			assert.Equal(t, tt.homeDir, user.HomeDir)
		})
	}
}
