package pool

import (
	"testing"
)

func TestNewBufferPool(t *testing.T) {
	p := NewBufferPool()
	if p == nil {
		t.Fatal("NewBufferPool() returned nil")
	}
}

func TestGet_ReturnsResetBuffer(t *testing.T) {
	p := NewBufferPool()
	buf := p.Get()
	if buf == nil {
		t.Fatal("Get() returned nil")
	}
	// Write some data.
	buf.WriteString("test data")
	if buf.Len() != 9 {
		t.Fatalf("buffer should have 9 bytes, got %d", buf.Len())
	}
	// Return to pool and get again — should be reset.
	p.Put(buf)
	buf2 := p.Get()
	if buf2.Len() != 0 {
		t.Errorf("buffer from pool should be reset, got %d bytes", buf2.Len())
	}
}

func TestPut_Nil(t *testing.T) {
	p := NewBufferPool()
	// Put(nil) should not panic.
	p.Put(nil)
	buf := p.Get()
	if buf == nil {
		t.Fatal("Get() after Put(nil) should still work")
	}
}

func TestGetPut_Capacity(t *testing.T) {
	p := NewBufferPool()
	buf1 := p.Get()
	// Write beyond initial capacity to force growth.
	for i := 0; i < 2048; i++ {
		buf1.WriteByte('x')
	}
	if buf1.Cap() <= 1024 {
		t.Errorf("expected capacity > 1024 after writes, got %d", buf1.Cap())
	}
	p.Put(buf1)
	// The next Get may reuse the grown buffer.
	buf2 := p.Get()
	_ = buf2 // Just ensure it works without panicking.
}

func TestConcurrentGetPut(t *testing.T) {
	p := NewBufferPool()
	done := make(chan struct{})
	for i := 0; i < 50; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			buf := p.Get()
			buf.WriteString("concurrent write")
			_ = buf.Len()
			p.Put(buf)
		}()
	}
	for i := 0; i < 50; i++ {
		<-done
	}
}
