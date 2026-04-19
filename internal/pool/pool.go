// Package pool provides a sync.Pool-backed pool of [bytes.Buffer] instances
// to reduce allocations in high-throughput QR code rendering paths.
package pool

import (
	"bytes"
	"sync"
)

// BufferPool is a pool of reusable [bytes.Buffer] instances backed by
// [sync.Pool]. Each buffer is reset before being handed out via [BufferPool.Get].
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool creates a new BufferPool. Each new buffer is initialized
// with a 1024-byte capacity.
func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() any {
				return bytes.NewBuffer(make([]byte, 0, 1024))
			},
		},
	}
}

// Get returns a reset [bytes.Buffer] from the pool, ready for reuse.
func (p *BufferPool) Get() *bytes.Buffer {
	b := p.pool.Get().(*bytes.Buffer) //nolint:errcheck // sync.Pool always returns non-nil
	b.Reset()
	return b
}

// Put returns a buffer to the pool for reuse. Nil buffers are silently ignored.
func (p *BufferPool) Put(b *bytes.Buffer) {
	if b != nil {
		p.pool.Put(b)
	}
}
