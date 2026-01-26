//go:build freebsd || openbsd || netbsd || dragonfly

package terminal

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA
