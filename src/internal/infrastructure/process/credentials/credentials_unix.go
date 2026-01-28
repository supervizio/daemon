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

	"github.com/kodflow/daemon/internal/infrastructure/process"
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

// NewManager returns a new credential Manager.
//
// Returns:
//   - *Manager: initialized credential manager for Unix systems
func NewManager() *Manager { return &Manager{} }

// New returns a new credential Manager.
//
// Returns:
//   - *Manager: initialized credential manager for Unix systems
func New() *Manager { return &Manager{} }

// LookupUser resolves a user by name or UID, with UID fallback.
//
// Params:
//   - nameOrID: username string or numeric UID to resolve
//
// Returns:
//   - *User: resolved user with UID, GID, username, and home directory
//   - error: ErrUserNotFound if neither name nor UID lookup succeeds
func (m *Manager) LookupUser(nameOrID string) (*User, error) {
	lookedUpUser, err := user.Lookup(nameOrID)
	// Fallback to numeric UID for minimal containers without passwd file.
	if err != nil {
		lookedUpUser, err = user.LookupId(nameOrID)
		// Both lookups failed.
		if err != nil {
			return nil, process.WrapError("lookup user", ErrUserNotFound)
		}
	}
	uid, _ := strconv.ParseUint(lookedUpUser.Uid, baseDecimal, bitSize32)
	gid, _ := strconv.ParseUint(lookedUpUser.Gid, baseDecimal, bitSize32)
	return &User{
		UID:      uint32(uid),
		GID:      uint32(gid),
		Username: lookedUpUser.Username,
		HomeDir:  lookedUpUser.HomeDir,
	}, nil
}

// LookupGroup resolves a group by name or GID, with GID fallback.
//
// Params:
//   - nameOrID: group name string or numeric GID to resolve
//
// Returns:
//   - *Group: resolved group with GID and name
//   - error: ErrGroupNotFound if neither name nor GID lookup succeeds
func (m *Manager) LookupGroup(nameOrID string) (*Group, error) {
	lookedUpGroup, err := user.LookupGroup(nameOrID)
	// Fallback to numeric GID for minimal containers without group file.
	if err != nil {
		lookedUpGroup, err = user.LookupGroupId(nameOrID)
		// Both lookups failed.
		if err != nil {
			return nil, process.WrapError("lookup group", ErrGroupNotFound)
		}
	}
	gid, _ := strconv.ParseUint(lookedUpGroup.Gid, baseDecimal, bitSize32)
	return &Group{
		GID:  uint32(gid),
		Name: lookedUpGroup.Name,
	}, nil
}

// ResolveCredentials converts user/group names to numeric IDs.
// Supports numeric fallback for minimal containers without passwd/group files.
//
// Params:
//   - username: user name or numeric UID (empty string skips user resolution)
//   - groupname: group name or numeric GID (empty string inherits from user)
//
// Returns:
//   - uid: resolved user ID (0 if username empty)
//   - gid: resolved group ID (inherits from user if groupname empty)
//   - err: lookup error if resolution fails and numeric fallback is invalid
func (m *Manager) ResolveCredentials(username, groupname string) (uid, gid uint32, err error) {
	// Handle user resolution with numeric fallback.
	if username != "" {
		resolvedUser, lookupErr := m.LookupUser(username)
		// User lookup failed; try numeric fallback for scratch containers.
		if lookupErr != nil {
			// Numeric string can be used directly as UID.
			if id, parseErr := strconv.ParseUint(username, baseDecimal, bitSize32); parseErr == nil {
				uid = uint32(id)
			} else {
				// Non-numeric and lookup failed.
				return 0, 0, fmt.Errorf("looking up user %s: %w", username, lookupErr)
			}
		} else {
			// User found via system lookup.
			uid = resolvedUser.UID
			// Inherit primary group when no explicit group specified.
			if groupname == "" {
				gid = resolvedUser.GID
			}
		}
	}
	// Handle group resolution with numeric fallback.
	if groupname != "" {
		resolvedGroup, lookupErr := m.LookupGroup(groupname)
		// Group lookup failed; try numeric fallback for scratch containers.
		if lookupErr != nil {
			// Numeric string can be used directly as GID.
			if id, parseErr := strconv.ParseUint(groupname, baseDecimal, bitSize32); parseErr == nil {
				gid = uint32(id)
			} else {
				// Non-numeric and lookup failed.
				return 0, 0, fmt.Errorf("looking up group %s: %w", groupname, lookupErr)
			}
		} else {
			// Group found via system lookup.
			gid = resolvedGroup.GID
		}
	}
	return uid, gid, nil
}

// ApplyCredentials sets syscall credentials on a command for privilege drop.
//
// Params:
//   - cmd: exec.Cmd to configure with credentials
//   - uid: user ID to run process as
//   - gid: group ID to run process as
//
// Returns:
//   - error: always nil (interface compliance)
func (m *Manager) ApplyCredentials(cmd *exec.Cmd, uid, gid uint32) error {
	// Skip credential setup when running as root (uid=0, gid=0).
	if uid == 0 && gid == 0 {
		return nil
	}
	// Initialize SysProcAttr if not already set by caller.
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: uid,
		Gid: gid,
	}
	return nil
}
