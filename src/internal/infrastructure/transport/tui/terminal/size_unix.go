//go:build linux || darwin || freebsd || openbsd || netbsd

package terminal

import (
	"os"
	"strconv"
	"syscall"
	"unsafe"
)

// winsize matches the C struct winsize.
type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

// getTerminalSize returns the terminal size using TIOCGWINSZ ioctl.
// COLUMNS/LINES environment variables take precedence when set (standard Unix behavior).
func getTerminalSize() (Size, error) {
	// Check environment variables first (allows explicit override).
	if envSize := getSizeFromEnv(); envSize != DefaultSize {
		return envSize, nil
	}

	// Try ioctl on stdout, stdin, stderr.
	var ws winsize
	fds := []uintptr{os.Stdout.Fd(), os.Stdin.Fd(), os.Stderr.Fd()}

	for _, fd := range fds {
		_, _, errno := syscall.Syscall(
			syscall.SYS_IOCTL,
			fd,
			syscall.TIOCGWINSZ,
			uintptr(unsafe.Pointer(&ws)),
		)
		if errno == 0 && ws.Col > 0 && ws.Row > 0 {
			return Size{
				Cols: int(ws.Col),
				Rows: int(ws.Row),
			}, nil
		}
	}

	return DefaultSize, syscall.ENOTTY
}

// getSizeFromEnv reads terminal size from COLUMNS/LINES environment variables.
func getSizeFromEnv() Size {
	size := DefaultSize

	if cols := os.Getenv("COLUMNS"); cols != "" {
		if c, err := strconv.Atoi(cols); err == nil && c > 0 {
			size.Cols = c
		}
	}

	if lines := os.Getenv("LINES"); lines != "" {
		if l, err := strconv.Atoi(lines); err == nil && l > 0 {
			size.Rows = l
		}
	}

	return size
}

// isTerminal returns true if fd is a terminal.
func isTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		ioctlReadTermios,
		uintptr(unsafe.Pointer(&termios)),
	)
	return errno == 0
}
