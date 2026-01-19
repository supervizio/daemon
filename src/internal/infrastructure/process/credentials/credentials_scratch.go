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
)

// ErrScratchNameLookup is returned when a name lookup is attempted in scratch mode.
var ErrScratchNameLookup error = errors.New("name lookup not available in scratch mode, use numeric UID/GID")

// ScratchManager implements CredentialManager for scratch containers.
// It only supports numeric UID/GID values since /etc/passwd is not available.
type ScratchManager struct{}

// NewScratchManager creates a new ScratchManager instance for scratch containers.
//
// Returns:
//   - *ScratchManager: a new credential manager for scratch environments
func NewScratchManager() *ScratchManager {
	// Create and return new ScratchManager instance.
	return &ScratchManager{}
}

// NewScratch creates a new ScratchManager instance for scratch containers.
//
// Returns:
//   - *ScratchManager: a new credential manager for scratch environments
func NewScratch() *ScratchManager {
	// Create and return new ScratchManager instance.
	return &ScratchManager{}
}

// LookupUser looks up a user by numeric UID only.
// Name lookups will fail as /etc/passwd is not available in scratch containers.
//
// Params:
//   - nameOrID: the username or numeric UID to look up
//
// Returns:
//   - *User: the user information if found
//   - error: an error if the user could not be found
func (m *ScratchManager) LookupUser(nameOrID string) (*User, error) {
	// Try parsing as numeric UID.
	uid, err := strconv.ParseUint(nameOrID, baseDecimal, bitSize32)
	// Check if parsing failed indicating non-numeric value.
	if err != nil {
		// Return error with helpful message for name lookup.
		return nil, fmt.Errorf("user %q: %w", nameOrID, ErrScratchNameLookup)
	}

	// Return a minimal user with just the UID.
	return &User{
		UID:      uint32(uid),
		GID:      uint32(uid), // Default GID to same as UID
		Username: nameOrID,    // Use the numeric string as username
	}, nil
}

// LookupGroup looks up a group by numeric GID only.
// Name lookups will fail as /etc/group is not available in scratch containers.
//
// Params:
//   - nameOrID: the group name or numeric GID to look up
//
// Returns:
//   - *Group: the group information if found
//   - error: an error if the group could not be found
func (m *ScratchManager) LookupGroup(nameOrID string) (*Group, error) {
	// Try parsing as numeric GID.
	gid, err := strconv.ParseUint(nameOrID, baseDecimal, bitSize32)
	// Check if parsing failed indicating non-numeric value.
	if err != nil {
		// Return error with helpful message for name lookup.
		return nil, fmt.Errorf("group %q: %w", nameOrID, ErrScratchNameLookup)
	}

	// Return a minimal group with just the GID.
	return &Group{
		GID:  uint32(gid),
		Name: nameOrID, // Use the numeric string as group name
	}, nil
}

// ResolveCredentials resolves user and group to UIDs and GIDs.
// Only numeric values are supported in scratch mode.
//
// Params:
//   - username: the username to resolve (can be empty)
//   - groupname: the group name to resolve (can be empty)
//
// Returns:
//   - uid: the resolved user ID
//   - gid: the resolved group ID
//   - err: an error if resolution failed
func (m *ScratchManager) ResolveCredentials(username, groupname string) (uid, gid uint32, err error) {
	// Check if username was provided for resolution.
	if username != "" {
		resolvedUser, lookupErr := m.LookupUser(username)
		// Check if user lookup failed.
		if lookupErr != nil {
			// Return error with context about user resolution failure.
			return 0, 0, fmt.Errorf("resolve user: %w", lookupErr)
		}
		uid = resolvedUser.UID
		// Check if no group specified to use user's default GID.
		if groupname == "" {
			gid = resolvedUser.GID
		}
	}

	// Check if groupname was provided for resolution.
	if groupname != "" {
		resolvedGroup, lookupErr := m.LookupGroup(groupname)
		// Check if group lookup failed.
		if lookupErr != nil {
			// Return error with context about group resolution failure.
			return 0, 0, fmt.Errorf("resolve group: %w", lookupErr)
		}
		gid = resolvedGroup.GID
	}

	// Return successfully resolved credentials.
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
//   - error: an error if credentials could not be applied
func (m *ScratchManager) ApplyCredentials(cmd *exec.Cmd, uid, gid uint32) error {
	// Check if both uid and gid are zero (no credentials to apply).
	if uid == 0 && gid == 0 {
		// Return nil as there are no credentials to apply.
		return nil
	}

	// Check if SysProcAttr is nil and initialize it.
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	// Set credentials on command.
	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: uid,
		Gid: gid,
	}

	// Return nil indicating credentials were successfully applied.
	return nil
}

// IsScratchEnvironment detects if we're running in a scratch container.
// This checks for the absence of /etc/passwd which is a strong indicator.
//
// Returns:
//   - bool: true if the current environment is scratch
func IsScratchEnvironment() bool {
	// Check for /etc/passwd existence.
	_, err := os.Stat("/etc/passwd")
	// Return true if stat failed indicating scratch environment.
	return err != nil
}
