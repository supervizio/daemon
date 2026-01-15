// Package credentials provides credential management interfaces and types.
package credentials

import (
	"errors"
	"os/exec"
)

// Sentinel errors for credential operations.
var (
	// ErrUserNotFound indicates that the specified user could not be found.
	ErrUserNotFound = errors.New("user not found")
	// ErrGroupNotFound indicates that the specified group could not be found.
	ErrGroupNotFound = errors.New("group not found")
)

// User represents a system user.
// It contains identification and profile information from the OS user database.
type User struct {
	// UID is the numeric user identifier.
	UID uint32
	// GID is the numeric primary group identifier for this user.
	GID uint32
	// Username is the login name of the user.
	Username string
	// HomeDir is the path to the user's home directory.
	HomeDir string
}

// Group represents a system group.
// It contains identification information from the OS group database.
type Group struct {
	// GID is the numeric group identifier.
	GID uint32
	// Name is the name of the group.
	Name string
}

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
