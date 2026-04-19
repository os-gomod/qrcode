package pool

import (
	"testing"
)

func TestNewBufferPool(t *testing.T) {
	p := NewBufferPool()
	if p == nil {
		t.Fatal("NewBufferPool returned nil")
	}
}

func TestGetPut(t *testing.T) {
	p := NewBufferPool()

	b := p.Get()
	if b == nil {
		t.Fatal("Get returned nil")
	}
	if b.Len() != 0 {
		t.Error("Get should return empty buffer")
	}

	b.WriteString("hello world")

	// Put it back
	p.Put(b)

	// Get again - should reuse
	b2 := p.Get()
	if b2 == nil {
		t.Fatal("second Get returned nil")
	}
	if b2.Len() != 0 {
		t.Error("Get after Put should return empty buffer")
	}

	// Put nil should not panic
	p.Put(nil)
}

func TestMultipleGetPut(t *testing.T) {
	p := NewBufferPool()

	// Get multiple buffers
	b1 := p.Get()
	b2 := p.Get()

	b1.WriteString("buffer1")
	b2.WriteString("buffer2")

	if b1.String() != "buffer1" || b2.String() != "buffer2" {
		t.Error("buffers should be independent")
	}

	p.Put(b1)
	p.Put(b2)

	// Verify they're properly reset
	b3 := p.Get()
	if b3.Len() != 0 {
		t.Error("buffer should be empty after recycling")
	}
	p.Put(b3)
}
