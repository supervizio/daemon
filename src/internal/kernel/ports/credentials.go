// Package ports defines the interfaces for OS abstraction.
package ports

import "os/exec"

// User represents a system user.
type User struct {
	UID      uint32
	GID      uint32
	Username string
	HomeDir  string
}

// Group represents a system group.
type Group struct {
	GID  uint32
	Name string
}

// CredentialManager handles user and group credential operations.
type CredentialManager interface {
	// LookupUser looks up a user by name or numeric UID.
	LookupUser(nameOrID string) (*User, error)

	// LookupGroup looks up a group by name or numeric GID.
	LookupGroup(nameOrID string) (*Group, error)

	// ResolveCredentials resolves user and group names to UIDs and GIDs.
	ResolveCredentials(username, groupname string) (uid, gid uint32, err error)

	// ApplyCredentials applies uid/gid credentials to a command.
	ApplyCredentials(cmd *exec.Cmd, uid, gid uint32) error
}
