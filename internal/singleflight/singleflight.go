package singleflight

import (
	"sync"
	"sync/atomic"
)

type Result struct {
	Val    any
	Err    error
	Shared bool
}
type call struct {
	wg   sync.WaitGroup
	val  any
	err  error
	dupe int32
}
type Group struct {
	mu    sync.Mutex
	calls map[string]*call
}

func NewGroup() *Group {
	return &Group{
		calls: make(map[string]*call),
	}
}

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

func (g *Group) DoChan(key string, fn func() (any, error)) <-chan Result {
	ch := make(chan Result, 1)
	go func() {
		val, shared, err := g.Do(key, fn)
		ch <- Result{Val: val, Err: err, Shared: shared}
	}()
	return ch
}

func (g *Group) Forget(key string) {
	g.mu.Lock()
	delete(g.calls, key)
	g.mu.Unlock()
}
