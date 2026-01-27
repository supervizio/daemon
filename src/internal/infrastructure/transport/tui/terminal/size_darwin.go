//go:build darwin

package terminal

import "syscall"

const ioctlReadTermios uintptr = syscall.TIOCGETA
