//go:build freebsd || openbsd || netbsd

package terminal

import "syscall"

const ioctlReadTermios uintptr = syscall.TIOCGETA
