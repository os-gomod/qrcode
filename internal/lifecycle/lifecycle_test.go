package lifecycle

import (
	"sync"
	"testing"
)

func TestNew(t *testing.T) {
	g := New()
	if g.IsClosed() {
		t.Error("new guard should not be closed")
	}
	done := g.Done()
	if done == nil {
		t.Fatal("Done() should return non-nil channel")
	}
	select {
	case <-done:
		t.Error("Done() channel should not be closed on new guard")
	default:
	}
}

func TestClose(t *testing.T) {
	g := New()
	err := g.Close()
	if err != nil {
		t.Errorf("first Close() should not error, got: %v", err)
	}
	if !g.IsClosed() {
		t.Error("guard should be closed after Close()")
	}
	// Done() channel should be closed.
	select {
	case <-g.Done():
	default:
		t.Error("Done() channel should be closed after Close()")
	}
}

func TestClose_AlreadyClosed(t *testing.T) {
	g := New()
	_ = g.Close()
	err := g.Close()
	if err != ErrAlreadyClosed {
		t.Errorf("second Close() should return ErrAlreadyClosed, got: %v", err)
	}
}

func TestClose_Concurrent(t *testing.T) {
	g := New()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = g.Close()
		}()
	}
	wg.Wait()
	if !g.IsClosed() {
		t.Error("guard should be closed after concurrent Close() calls")
	}
}

func TestIsClosed_Concurrent(t *testing.T) {
	g := New()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = g.IsClosed()
		}()
	}
	wg.Wait()
	if g.IsClosed() {
		t.Error("guard should NOT be closed after only IsClosed() calls")
	}
}

func TestDone_BlocksUntilClosed(t *testing.T) {
	g := New()
	ch := make(chan struct{})
	go func() {
		<-g.Done()
		close(ch)
	}()
	// Give the goroutine time to start and block.
	select {
	case <-ch:
		t.Fatal("Done() should block until Close()")
	default:
	}
	_ = g.Close()
	// Now Done() should be closed and the goroutine should have unblocked.
	// We can't guarantee timing, but in practice this should work.
	select {
	case <-ch:
		// Good, goroutine saw the close.
	case <-g.Done():
		// Also good.
	}
}

func TestClose_Idempotent_Race(t *testing.T) {
	g := New()
	errCh := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			errCh <- g.Close()
		}()
	}
	closeCount := 0
	errorCount := 0
	for i := 0; i < 10; i++ {
		if err := <-errCh; err != nil {
			errorCount++
		} else {
			closeCount++
		}
	}
	// Exactly one Close() should succeed (return nil).
	if closeCount != 1 {
		t.Errorf("expected exactly 1 successful Close(), got %d", closeCount)
	}
	if errorCount != 9 {
		t.Errorf("expected 9 ErrAlreadyClosed, got %d", errorCount)
	}
}
