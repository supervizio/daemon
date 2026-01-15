//go:build unix

// Package credentials provides credential management for Unix systems.
// It handles user and group lookup, resolution, and credential application to processes.
package credentials

import (
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/kodflow/daemon/internal/infrastructure/kernel/ports"
)

const (
	// baseDecimal is the base for parsing decimal numbers.
	baseDecimal int = 10
	// bitSize32 is the bit size for 32-bit unsigned integers.
	bitSize32 int = 32
)

// Manager implements CredentialManager for Unix systems.
// It provides methods for looking up users and groups, resolving credentials,
// and applying them to processes via syscall.Credential.
type Manager struct{}

// New creates a new credential Manager instance.
//
// Returns:
//   - *Manager: a new credential manager instance
func New() *Manager {
	// Return a new empty instance of Manager.
	return &Manager{}
}

// LookupUser looks up a user by name or numeric UID.
//
// Params:
//   - nameOrID: the username or numeric UID to look up
//
// Returns:
//   - *ports.User: the user information if found
//   - error: an error if the user could not be found
func (m *Manager) LookupUser(nameOrID string) (*ports.User, error) {
	lookedUpUser, err := user.Lookup(nameOrID)
	// Check if the user lookup by name failed.
	if err != nil {
		// Try looking up by UID as fallback.
		lookedUpUser, err = user.LookupId(nameOrID)
		// Check if the lookup by UID also failed.
		if err != nil {
			// Return an error indicating the user was not found.
			return nil, ports.WrapError("lookup user", ports.ErrUserNotFound)
		}
	}

	uid, _ := strconv.ParseUint(lookedUpUser.Uid, baseDecimal, bitSize32)
	gid, _ := strconv.ParseUint(lookedUpUser.Gid, baseDecimal, bitSize32)

	// Return the user information with parsed UID and GID.
	return &ports.User{
		UID:      uint32(uid),
		GID:      uint32(gid),
		Username: lookedUpUser.Username,
		HomeDir:  lookedUpUser.HomeDir,
	}, nil
}

// LookupGroup looks up a group by name or numeric GID.
//
// Params:
//   - nameOrID: the group name or numeric GID to look up
//
// Returns:
//   - *ports.Group: the group information if found
//   - error: an error if the group could not be found
func (m *Manager) LookupGroup(nameOrID string) (*ports.Group, error) {
	lookedUpGroup, err := user.LookupGroup(nameOrID)
	// Check if the group lookup by name failed.
	if err != nil {
		// Try looking up by GID as fallback.
		lookedUpGroup, err = user.LookupGroupId(nameOrID)
		// Check if the lookup by GID also failed.
		if err != nil {
			// Return an error indicating the group was not found.
			return nil, ports.WrapError("lookup group", ports.ErrGroupNotFound)
		}
	}

	gid, _ := strconv.ParseUint(lookedUpGroup.Gid, baseDecimal, bitSize32)

	// Return the group information with parsed GID.
	return &ports.Group{
		GID:  uint32(gid),
		Name: lookedUpGroup.Name,
	}, nil
}

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
func (m *Manager) ResolveCredentials(username, groupname string) (uid, gid uint32, err error) {
	// Check if a username was provided to resolve.
	if username != "" {
		resolvedUser, lookupErr := m.LookupUser(username)
		// Check if the user lookup failed.
		if lookupErr != nil {
			// Try parsing the username as a numeric UID.
			if id, parseErr := strconv.ParseUint(username, baseDecimal, bitSize32); parseErr == nil {
				uid = uint32(id)
			} else {
				// Return an error if both lookup and parsing failed.
				return 0, 0, fmt.Errorf("looking up user %s: %w", username, lookupErr)
			}
		} else {
			// Use the UID from the resolved user.
			uid = resolvedUser.UID
			// Check if no group was specified to use the user's primary group.
			if groupname == "" {
				gid = resolvedUser.GID
			}
		}
	}

	// Check if a groupname was provided to resolve.
	if groupname != "" {
		resolvedGroup, lookupErr := m.LookupGroup(groupname)
		// Check if the group lookup failed.
		if lookupErr != nil {
			// Try parsing the groupname as a numeric GID.
			if id, parseErr := strconv.ParseUint(groupname, baseDecimal, bitSize32); parseErr == nil {
				gid = uint32(id)
			} else {
				// Return an error if both lookup and parsing failed.
				return 0, 0, fmt.Errorf("looking up group %s: %w", groupname, lookupErr)
			}
		} else {
			// Use the GID from the resolved group.
			gid = resolvedGroup.GID
		}
	}

	// Return the resolved credentials.
	return uid, gid, nil
}

// ApplyCredentials applies uid/gid credentials to a command.
//
// Params:
//   - cmd: the command to apply credentials to
//   - uid: the user ID to set
//   - gid: the group ID to set
//
// Returns:
//   - error: always nil for this implementation
func (m *Manager) ApplyCredentials(cmd *exec.Cmd, uid, gid uint32) error {
	// Check if both uid and gid are zero (no credentials to apply).
	if uid == 0 && gid == 0 {
		// Return nil as there are no credentials to apply.
		return nil
	}

	// Check if SysProcAttr is nil and initialize it.
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: uid,
		Gid: gid,
	}

	// Return nil indicating credentials were successfully applied.
	return nil
}
