//go:build unix

// Package credentials provides credential management for scratch containers.
// This implementation only supports numeric UIDs/GIDs, as /etc/passwd is not available.
package credentials

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/kodflow/daemon/internal/infrastructure/kernel/ports"
)

// ErrScratchNameLookup is returned when a name lookup is attempted in scratch mode.
var ErrScratchNameLookup = errors.New("name lookup not available in scratch mode, use numeric UID/GID")

// ScratchManager implements CredentialManager for scratch containers.
// It only supports numeric UID/GID values since /etc/passwd is not available.
type ScratchManager struct{}

// NewScratch creates a new ScratchManager instance for scratch containers.
//
// Returns:
//   - *ScratchManager: a new credential manager for scratch environments
func NewScratch() *ScratchManager {
	return &ScratchManager{}
}

// LookupUser looks up a user by numeric UID only.
// Name lookups will fail as /etc/passwd is not available in scratch containers.
func (m *ScratchManager) LookupUser(nameOrID string) (*ports.User, error) {
	// Try parsing as numeric UID
	uid, err := strconv.ParseUint(nameOrID, baseDecimal, bitSize32)
	if err != nil {
		// Not a numeric UID, return error with helpful message
		return nil, fmt.Errorf("user %q: %w", nameOrID, ErrScratchNameLookup)
	}

	// Return a minimal user with just the UID
	return &ports.User{
		UID:      uint32(uid),
		GID:      uint32(uid), // Default GID to same as UID
		Username: nameOrID,    // Use the numeric string as username
	}, nil
}

// LookupGroup looks up a group by numeric GID only.
// Name lookups will fail as /etc/group is not available in scratch containers.
func (m *ScratchManager) LookupGroup(nameOrID string) (*ports.Group, error) {
	// Try parsing as numeric GID
	gid, err := strconv.ParseUint(nameOrID, baseDecimal, bitSize32)
	if err != nil {
		// Not a numeric GID, return error with helpful message
		return nil, fmt.Errorf("group %q: %w", nameOrID, ErrScratchNameLookup)
	}

	// Return a minimal group with just the GID
	return &ports.Group{
		GID:  uint32(gid),
		Name: nameOrID, // Use the numeric string as group name
	}, nil
}

// ResolveCredentials resolves user and group to UIDs and GIDs.
// Only numeric values are supported in scratch mode.
func (m *ScratchManager) ResolveCredentials(username, groupname string) (uid, gid uint32, err error) {
	// Resolve username if provided
	if username != "" {
		resolvedUser, lookupErr := m.LookupUser(username)
		if lookupErr != nil {
			return 0, 0, fmt.Errorf("resolve user: %w", lookupErr)
		}
		uid = resolvedUser.UID
		// Default GID to user's GID if no group specified
		if groupname == "" {
			gid = resolvedUser.GID
		}
	}

	// Resolve groupname if provided
	if groupname != "" {
		resolvedGroup, lookupErr := m.LookupGroup(groupname)
		if lookupErr != nil {
			return 0, 0, fmt.Errorf("resolve group: %w", lookupErr)
		}
		gid = resolvedGroup.GID
	}

	return uid, gid, nil
}

// ApplyCredentials applies uid/gid credentials to a command.
func (m *ScratchManager) ApplyCredentials(cmd *exec.Cmd, uid, gid uint32) error {
	// Check if both uid and gid are zero (no credentials to apply)
	if uid == 0 && gid == 0 {
		return nil
	}

	// Initialize SysProcAttr if needed
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	// Set credentials
	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: uid,
		Gid: gid,
	}

	return nil
}

// IsScratchEnvironment detects if we're running in a scratch container.
// This checks for the absence of /etc/passwd which is a strong indicator.
func IsScratchEnvironment() bool {
	// Check for /etc/passwd existence
	_, err := os.Stat("/etc/passwd")
	return err != nil
}
