package lifecycle

import (
	"errors"
	"sync"
)

var ErrAlreadyClosed = errors.New("lifecycle: already closed")

type Guard struct {
	mu     sync.RWMutex
	closed bool
	done   chan struct{}
}

func New() *Guard {
	return &Guard{
		done: make(chan struct{}),
	}
}

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

func (g *Guard) IsClosed() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.closed
}

func (g *Guard) Done() <-chan struct{} {
	return g.done
}
