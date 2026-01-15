//go:build unix

// Package reaper provides platform-specific implementations of kernel interfaces.
// This file implements zombie process reaping for Unix systems.
package reaper

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Reaper implements ZombieReaper for Unix systems.
// It handles automatic reaping of zombie child processes using SIGCHLD signals.
type Reaper struct {
	// mu protects the running state from concurrent access.
	mu sync.Mutex
	// running indicates whether the reaper loop is currently active.
	running bool
	// stopCh is used to signal the reaper loop to stop.
	stopCh chan struct{}
	// doneCh is closed when the reaper loop has fully stopped.
	doneCh chan struct{}
}

// New creates a new ZombieReaper.
//
// Returns:
//   - *Reaper: a new zombie reaper instance
func New() *Reaper {
	// Return a new instance with initialized channels.
	return &Reaper{
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

// Start begins the background reaping loop.
func (r *Reaper) Start() {
	r.mu.Lock()
	// Check if the reaper is already running to avoid duplicate loops.
	if r.running {
		r.mu.Unlock()
		// Return early since the reaper is already running.
		return
	}
	r.running = true
	r.stopCh = make(chan struct{})
	r.doneCh = make(chan struct{})
	r.mu.Unlock()

	go r.reapLoop()
}

// Stop stops the reaping loop.
func (r *Reaper) Stop() {
	r.mu.Lock()
	// Check if the reaper is not running to avoid closing already closed channels.
	if !r.running {
		r.mu.Unlock()
		// Return early since the reaper is not running.
		return
	}
	r.running = false
	r.mu.Unlock()

	close(r.stopCh)
	<-r.doneCh
}

// reapLoop continuously reaps zombie processes.
func (r *Reaper) reapLoop() {
	// Defer closing doneCh to signal completion to Stop().
	defer close(r.doneCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGCHLD)
	// Defer stopping signal notifications to clean up resources.
	defer signal.Stop(sigCh)

	// Iterate indefinitely until stop signal is received.
	for {
		// Select on stop or SIGCHLD channels.
		select {
		// Case stopCh handles the stop signal to terminate the loop.
		case <-r.stopCh:
			// Final reap before exit.
			r.reapAll()
			// Return to exit the reap loop.
			return
		// Case sigCh handles SIGCHLD signal indicating child process state change.
		case <-sigCh:
			r.reapAll()
		}
	}
}

// reapAll reaps all zombie processes.
func (r *Reaper) reapAll() {
	// Iterate until no more zombie processes are available to reap.
	for {
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		// Check if wait4 failed or returned no process to break the loop.
		if err != nil || pid <= 0 {
			break
		}
	}
}

// ReapOnce performs a single reap cycle and returns the count of reaped processes.
//
// Returns:
//   - int: the number of zombie processes reaped
func (r *Reaper) ReapOnce() int {
	count := 0
	// Iterate until no more zombie processes are available to reap.
	for {
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		// Check if wait4 failed or returned no process to break the loop.
		if err != nil || pid <= 0 {
			break
		}
		count++
	}
	// Return the total count of reaped zombie processes.
	return count
}

// IsPID1 returns true if the current process is running as PID 1.
//
// Returns:
//   - bool: true if the current process ID is 1
func (r *Reaper) IsPID1() bool {
	// Return true if the current process ID equals 1.
	return os.Getpid() == 1
}
