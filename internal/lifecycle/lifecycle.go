// Package lifecycle provides a concurrency-safe close guard that tracks
// whether a resource has been closed and signals waiters via a channel.
package lifecycle

import (
	"errors"
	"sync"
)

// ErrAlreadyClosed is returned by [Guard.Close] when the guard is already closed.
var ErrAlreadyClosed = errors.New("lifecycle: already closed")

// Guard is a concurrency-safe close guard. It tracks whether Close has been
// called and provides a Done channel for goroutines to wait on.
type Guard struct {
	mu     sync.RWMutex
	closed bool
	done   chan struct{}
}

// New creates a new open Guard.
func New() *Guard {
	return &Guard{
		done: make(chan struct{}),
	}
}

// Close marks the guard as closed and closes the Done channel. Returns
// [ErrAlreadyClosed] if called more than once.
func (g *Guard) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.closed {
		return ErrAlreadyClosed
	}
	g.closed = true
	close(g.done)
	return nil
}

// IsClosed reports whether the guard has been closed.
func (g *Guard) IsClosed() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.closed
}

// Done returns a channel that is closed when [Guard.Close] is called.
// Callers can use this with select to detect shutdown.
func (g *Guard) Done() <-chan struct{} {
	return g.done
}
