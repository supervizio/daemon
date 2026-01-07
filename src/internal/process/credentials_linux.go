//go:build linux

package process

import (
	"os/exec"
	"syscall"
)

// setCredentials sets the UID and GID for a command on Linux.
func setCredentials(cmd *exec.Cmd, uid, gid uint32) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: uid,
		Gid: gid,
	}

	return nil
}
