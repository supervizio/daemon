package process

import (
	"fmt"
	"os/exec"
	"os/user"
	"strconv"

	"github.com/kodflow/daemon/internal/config"
)

// applyCredentials sets up user/group credentials for a command.
func applyCredentials(cmd *exec.Cmd, cfg *config.ServiceConfig) error {
	if cfg.User == "" && cfg.Group == "" {
		return nil
	}

	uid, gid, err := resolveCredentials(cfg.User, cfg.Group)
	if err != nil {
		return err
	}

	return setCredentials(cmd, uid, gid)
}

// resolveCredentials resolves user and group names to UIDs and GIDs.
func resolveCredentials(username, groupname string) (uid, gid uint32, err error) {
	// Resolve user
	if username != "" {
		u, err := user.Lookup(username)
		if err != nil {
			// Try as numeric UID
			if id, parseErr := strconv.ParseUint(username, 10, 32); parseErr == nil {
				uid = uint32(id)
			} else {
				return 0, 0, fmt.Errorf("looking up user %s: %w", username, err)
			}
		} else {
			id, _ := strconv.ParseUint(u.Uid, 10, 32)
			uid = uint32(id)
			// Use user's primary group if no group specified
			if groupname == "" {
				id, _ := strconv.ParseUint(u.Gid, 10, 32)
				gid = uint32(id)
			}
		}
	}

	// Resolve group
	if groupname != "" {
		g, err := user.LookupGroup(groupname)
		if err != nil {
			// Try as numeric GID
			if id, parseErr := strconv.ParseUint(groupname, 10, 32); parseErr == nil {
				gid = uint32(id)
			} else {
				return 0, 0, fmt.Errorf("looking up group %s: %w", groupname, err)
			}
		} else {
			id, _ := strconv.ParseUint(g.Gid, 10, 32)
			gid = uint32(id)
		}
	}

	return uid, gid, nil
}

// LookupUser looks up a user by name or numeric ID.
func LookupUser(name string) (*user.User, error) {
	u, err := user.Lookup(name)
	if err != nil {
		// Try looking up by UID
		return user.LookupId(name)
	}
	return u, nil
}

// LookupGroup looks up a group by name or numeric ID.
func LookupGroup(name string) (*user.Group, error) {
	g, err := user.LookupGroup(name)
	if err != nil {
		// Try looking up by GID
		return user.LookupGroupId(name)
	}
	return g, nil
}
