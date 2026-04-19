package lifecycle

import (
	"testing"
)

func TestNew(t *testing.T) {
	g := New()
	if g == nil {
		t.Fatal("New returned nil")
	}
	if g.IsClosed() {
		t.Error("new guard should not be closed")
	}
}

func TestClose(t *testing.T) {
	g := New()
	if err := g.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}
	if !g.IsClosed() {
		t.Error("guard should be closed after Close()")
	}
}

func TestDoubleClose(t *testing.T) {
	g := New()
	if err := g.Close(); err != nil {
		t.Fatalf("first Close error: %v", err)
	}
	err := g.Close()
	if err == nil {
		t.Error("second Close should return error")
	}
	if err != ErrAlreadyClosed {
		t.Errorf("second Close error = %v, want ErrAlreadyClosed", err)
	}
}

func TestIsClosed(t *testing.T) {
	g := New()
	if g.IsClosed() {
		t.Error("new guard should not be closed")
	}
	g.Close()
	if !g.IsClosed() {
		t.Error("guard should be closed after Close()")
	}
}

func TestDone(t *testing.T) {
	g := New()
	done := g.Done()
	if done == nil {
		t.Fatal("Done() returned nil channel")
	}

	select {
	case <-done:
		t.Error("Done channel should not be closed yet")
	default:
	}

	g.Close()

	select {
	case <-done:
		// Expected
	default:
		t.Error("Done channel should be closed after Close()")
	}
}
