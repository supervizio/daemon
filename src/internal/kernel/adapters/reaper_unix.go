//go:build unix

package adapters

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// UnixZombieReaper implements ZombieReaper for Unix systems.
type UnixZombieReaper struct {
	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
	doneCh  chan struct{}
}

// NewZombieReaper creates a new ZombieReaper.
func NewZombieReaper() *UnixZombieReaper {
	return &UnixZombieReaper{
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}

// Start begins the background reaping loop.
func (r *UnixZombieReaper) Start() {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return
	}
	r.running = true
	r.stopCh = make(chan struct{})
	r.doneCh = make(chan struct{})
	r.mu.Unlock()

	go r.reapLoop()
}

// Stop stops the reaping loop.
func (r *UnixZombieReaper) Stop() {
	r.mu.Lock()
	if !r.running {
		r.mu.Unlock()
		return
	}
	r.running = false
	r.mu.Unlock()

	close(r.stopCh)
	<-r.doneCh
}

// reapLoop continuously reaps zombie processes.
func (r *UnixZombieReaper) reapLoop() {
	defer close(r.doneCh)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGCHLD)
	defer signal.Stop(sigCh)

	for {
		select {
		case <-r.stopCh:
			// Final reap before exit
			r.reapAll()
			return
		case <-sigCh:
			r.reapAll()
		}
	}
}

// reapAll reaps all zombie processes.
func (r *UnixZombieReaper) reapAll() {
	for {
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		if err != nil || pid <= 0 {
			break
		}
	}
}

// ReapOnce performs a single reap cycle and returns the count of reaped processes.
func (r *UnixZombieReaper) ReapOnce() int {
	count := 0
	for {
		pid, err := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
		if err != nil || pid <= 0 {
			break
		}
		count++
	}
	return count
}

// IsPID1 returns true if the current process is running as PID 1.
func (r *UnixZombieReaper) IsPID1() bool {
	return os.Getpid() == 1
}
