//go:build linux

// Package terminal provides Linux-specific terminal constants.
package terminal

import "syscall"

// ioctlReadTermios is the ioctl request code for reading terminal attributes on Linux.
const ioctlReadTermios uintptr = syscall.TCGETS
