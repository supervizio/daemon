package process

import (
	"os/user"
	"strconv"

	"github.com/kodflow/daemon/internal/kernel"
)

// resolveCredentials resolves user and group names to UIDs and GIDs.
// This function delegates to the kernel.Credentials interface.
func resolveCredentials(username, groupname string) (uid, gid uint32, err error) {
	return kernel.Default.Credentials.ResolveCredentials(username, groupname)
}

// LookupUser looks up a user by name or numeric ID.
func LookupUser(name string) (*user.User, error) {
	u, err := kernel.Default.Credentials.LookupUser(name)
	if err != nil {
		return nil, err
	}
	return &user.User{
		Uid:      strconv.FormatUint(uint64(u.UID), 10),
		Gid:      strconv.FormatUint(uint64(u.GID), 10),
		Username: u.Username,
		HomeDir:  u.HomeDir,
	}, nil
}

// LookupGroup looks up a group by name or numeric ID.
func LookupGroup(name string) (*user.Group, error) {
	g, err := kernel.Default.Credentials.LookupGroup(name)
	if err != nil {
		return nil, err
	}
	return &user.Group{
		Gid:  strconv.FormatUint(uint64(g.GID), 10),
		Name: g.Name,
	}, nil
}
