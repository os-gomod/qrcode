// Package singleflight provides a mechanism to suppress duplicate function
// calls. When multiple goroutines call [Group.Do] with the same key, only the
// first call executes the function; subsequent callers block and share the result.
package singleflight

import (
	"sync"
	"sync/atomic"
)

// Result holds the outcome of a deduplicated function call.
type Result struct {
	// Val is the value returned by the function.
	Val any
	// Err is the error returned by the function.
	Err error
	// Shared reports whether the result was shared with other callers.
	Shared bool
}
type call struct {
	wg   sync.WaitGroup
	val  any
	err  error
	dupe int32
}

// Group manages deduplicated function calls keyed by string.
type Group struct {
	mu    sync.Mutex
	calls map[string]*call
}

// NewGroup creates a new Group ready to track deduplicated calls.
func NewGroup() *Group {
	return &Group{
		calls: make(map[string]*call),
	}
}

// Do executes fn for the given key, ensuring that concurrent callers
// with the same key share the same result. Returns the value, whether this
// call shared the result with another in-flight call, and an error.
func (g *Group) Do(key string, fn func() (any, error)) (any, bool, error) {
	g.mu.Lock()
	if c, ok := g.calls[key]; ok {
		atomic.AddInt32(&c.dupe, 1)
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, atomic.LoadInt32(&c.dupe) > 1, c.err
	}
	c := &call{}
	c.wg.Add(1)
	g.calls[key] = c
	g.mu.Unlock()
	c.val, c.err = fn()
	c.wg.Done()
	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()
	shared := atomic.LoadInt32(&c.dupe) > 0
	return c.val, shared, c.err
}

// DoChan is like [Group.Do] but returns a channel that receives the
// [Result] when the function completes.
func (g *Group) DoChan(key string, fn func() (any, error)) <-chan Result {
	ch := make(chan Result, 1)
	go func() {
		val, shared, err := g.Do(key, fn)
		ch <- Result{Val: val, Err: err, Shared: shared}
	}()
	return ch
}

// Forget discards the in-flight call for the given key, if any. Subsequent
// calls to [Group.Do] will execute the function again.
func (g *Group) Forget(key string) {
	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()
}
