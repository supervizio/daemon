// Package credentials_test provides black-box tests for credentials package.
package credentials_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/infrastructure/process/credentials"
)

// TestGroup_Fields tests the Group struct fields.
//
// Params:
//   - t: the testing context
func TestGroup_Fields(t *testing.T) {
	tests := []struct {
		name      string
		gid       uint32
		groupName string
	}{
		{
			name:      "basic group",
			gid:       1000,
			groupName: "testgroup",
		},
		{
			name:      "root group",
			gid:       0,
			groupName: "root",
		},
		{
			name:      "wheel group",
			gid:       10,
			groupName: "wheel",
		},
		{
			name:      "empty name",
			gid:       0,
			groupName: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := credentials.Group{
				GID:  tt.gid,
				Name: tt.groupName,
			}

			// Verify all fields are accessible
			assert.Equal(t, tt.gid, group.GID)
			assert.Equal(t, tt.groupName, group.Name)
		})
	}
}
