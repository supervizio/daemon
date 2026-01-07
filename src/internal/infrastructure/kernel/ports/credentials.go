// Package ports defines the interfaces for OS abstraction.
package ports

import "os/exec"

// CredentialManager handles user and group credential operations.
type CredentialManager interface {
	// LookupUser looks up a user by name or numeric UID.
	//
	// Params:
	//   - nameOrID: the username or numeric UID to look up
	//
	// Returns:
	//   - *User: the user information if found
	//   - error: an error if the user could not be found
	LookupUser(nameOrID string) (*User, error)

	// LookupGroup looks up a group by name or numeric GID.
	//
	// Params:
	//   - nameOrID: the group name or numeric GID to look up
	//
	// Returns:
	//   - *Group: the group information if found
	//   - error: an error if the group could not be found
	LookupGroup(nameOrID string) (*Group, error)

	// ResolveCredentials resolves user and group names to UIDs and GIDs.
	//
	// Params:
	//   - username: the username to resolve (can be empty)
	//   - groupname: the group name to resolve (can be empty)
	//
	// Returns:
	//   - uid: the resolved user ID
	//   - gid: the resolved group ID
	//   - err: an error if resolution failed
	ResolveCredentials(username, groupname string) (uid, gid uint32, err error)

	// ApplyCredentials applies uid/gid credentials to a command.
	//
	// Params:
	//   - cmd: the command to apply credentials to
	//   - uid: the user ID to set
	//   - gid: the group ID to set
	//
	// Returns:
	//   - error: an error if credentials could not be applied
	ApplyCredentials(cmd *exec.Cmd, uid, gid uint32) error
}
