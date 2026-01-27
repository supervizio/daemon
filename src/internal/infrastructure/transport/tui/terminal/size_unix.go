//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

// Package terminal provides Unix-specific terminal size detection.
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
//
// Returns:
//   - Size: detected terminal dimensions
//   - error: syscall.ENOTTY if terminal size cannot be determined
func getTerminalSize() (Size, error) {
	// Check environment variables first (allows explicit override).
	if envSize := getSizeFromEnv(); envSize != DefaultSize {
		// Environment variables set, use them.
		return envSize, nil
	}

	// Try ioctl on stdout, stdin, stderr.
	var ws winsize
	fds := []uintptr{os.Stdout.Fd(), os.Stdin.Fd(), os.Stderr.Fd()}

	// Attempt ioctl on each file descriptor.
	for _, fd := range fds {
		_, _, errno := syscall.Syscall(
			syscall.SYS_IOCTL,
			fd,
			syscall.TIOCGWINSZ,
			uintptr(unsafe.Pointer(&ws)),
		)
		// Check if ioctl succeeded and size is valid.
		if errno == 0 && ws.Col > 0 && ws.Row > 0 {
			// Return detected size.
			return Size{
				Cols: int(ws.Col),
				Rows: int(ws.Row),
			}, nil
		}
	}

	// No valid size detected, return error.
	return DefaultSize, syscall.ENOTTY
}

// getSizeFromEnv reads terminal size from COLUMNS/LINES environment variables.
//
// Returns:
//   - Size: terminal dimensions from environment, or DefaultSize if not set
func getSizeFromEnv() Size {
	size := DefaultSize

	// Check COLUMNS environment variable.
	if cols := os.Getenv("COLUMNS"); cols != "" {
		// Parse columns value.
		if colsNum, err := strconv.Atoi(cols); err == nil && colsNum > 0 {
			size.Cols = colsNum
		}
	}

	// Check LINES environment variable.
	if lines := os.Getenv("LINES"); lines != "" {
		// Parse lines value.
		if linesNum, err := strconv.Atoi(lines); err == nil && linesNum > 0 {
			size.Rows = linesNum
		}
	}

	// Return size from environment or default.
	return size
}

// isTerminal returns true if fd is a terminal.
//
// Params:
//   - fd: file descriptor to check
//
// Returns:
//   - bool: true if fd is a terminal, false otherwise
func isTerminal(fd uintptr) bool {
	var termios syscall.Termios
	// Attempt to read terminal attributes.
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		ioctlReadTermios,
		uintptr(unsafe.Pointer(&termios)),
	)
	// Return true if ioctl succeeded.
	return errno == 0
}
