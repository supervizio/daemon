//go:build unix

package adapters

import (
	"fmt"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"

	"github.com/kodflow/daemon/internal/kernel/ports"
)

// UnixCredentialManager implements CredentialManager for Unix systems.
type UnixCredentialManager struct{}

// NewCredentialManager creates a new CredentialManager.
func NewCredentialManager() *UnixCredentialManager {
	return &UnixCredentialManager{}
}

// LookupUser looks up a user by name or numeric UID.
func (m *UnixCredentialManager) LookupUser(nameOrID string) (*ports.User, error) {
	u, err := user.Lookup(nameOrID)
	if err != nil {
		// Try looking up by UID
		u, err = user.LookupId(nameOrID)
		if err != nil {
			return nil, ports.WrapError("lookup user", ports.ErrUserNotFound)
		}
	}

	uid, _ := strconv.ParseUint(u.Uid, 10, 32)
	gid, _ := strconv.ParseUint(u.Gid, 10, 32)

	return &ports.User{
		UID:      uint32(uid),
		GID:      uint32(gid),
		Username: u.Username,
		HomeDir:  u.HomeDir,
	}, nil
}

// LookupGroup looks up a group by name or numeric GID.
func (m *UnixCredentialManager) LookupGroup(nameOrID string) (*ports.Group, error) {
	g, err := user.LookupGroup(nameOrID)
	if err != nil {
		// Try looking up by GID
		g, err = user.LookupGroupId(nameOrID)
		if err != nil {
			return nil, ports.WrapError("lookup group", ports.ErrGroupNotFound)
		}
	}

	gid, _ := strconv.ParseUint(g.Gid, 10, 32)

	return &ports.Group{
		GID:  uint32(gid),
		Name: g.Name,
	}, nil
}

// ResolveCredentials resolves user and group names to UIDs and GIDs.
func (m *UnixCredentialManager) ResolveCredentials(username, groupname string) (uid, gid uint32, err error) {
	// Resolve user
	if username != "" {
		u, lookupErr := m.LookupUser(username)
		if lookupErr != nil {
			// Try as numeric UID
			if id, parseErr := strconv.ParseUint(username, 10, 32); parseErr == nil {
				uid = uint32(id)
			} else {
				return 0, 0, fmt.Errorf("looking up user %s: %w", username, lookupErr)
			}
		} else {
			uid = u.UID
			// Use user's primary group if no group specified
			if groupname == "" {
				gid = u.GID
			}
		}
	}

	// Resolve group
	if groupname != "" {
		g, lookupErr := m.LookupGroup(groupname)
		if lookupErr != nil {
			// Try as numeric GID
			if id, parseErr := strconv.ParseUint(groupname, 10, 32); parseErr == nil {
				gid = uint32(id)
			} else {
				return 0, 0, fmt.Errorf("looking up group %s: %w", groupname, lookupErr)
			}
		} else {
			gid = g.GID
		}
	}

	return uid, gid, nil
}

// ApplyCredentials applies uid/gid credentials to a command.
func (m *UnixCredentialManager) ApplyCredentials(cmd *exec.Cmd, uid, gid uint32) error {
	if uid == 0 && gid == 0 {
		return nil
	}

	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: uid,
		Gid: gid,
	}

	return nil
}
