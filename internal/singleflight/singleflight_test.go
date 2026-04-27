package singleflight

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo_Basic(t *testing.T) {
	g := NewGroup()
	val, shared, err := g.Do("key1", func() (any, error) {
		return "result1", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "result1" {
		t.Errorf("expected 'result1', got %v", val)
	}
	if shared {
		t.Error("first call should not be shared")
	}
}

func TestDo_Error(t *testing.T) {
	g := NewGroup()
	expectedErr := errors.New("call failed")
	_, _, err := g.Do("key2", func() (any, error) {
		return nil, expectedErr
	})
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestDo_Dedup(t *testing.T) {
	g := NewGroup()
	var calls int32
	var wg sync.WaitGroup

	fn := func() (any, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond) // Simulate slow work.
		return "deduped", nil
	}

	results := make([]struct {
		val    any
		shared bool
		err    error
	}, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			val, shared, err := g.Do("dedup-key", fn)
			results[idx] = struct {
				val    any
				shared bool
				err    error
			}{val, shared, err}
		}(i)
	}

	wg.Wait()

	// fn should be called exactly once due to deduplication.
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call, got %d", atomic.LoadInt32(&calls))
	}

	// All results should have the same value.
	for i, r := range results {
		if r.val != "deduped" {
			t.Errorf("result[%d] = %v, want 'deduped'", i, r.val)
		}
		if r.err != nil {
			t.Errorf("result[%d] error = %v", i, r.err)
		}
	}

	// At least 4 should be shared (some may not be due to timing).
	sharedCount := 0
	for _, r := range results {
		if r.shared {
			sharedCount++
		}
	}
	if sharedCount < 4 {
		t.Errorf("expected at least 4 shared, got %d", sharedCount)
	}
}

func TestDo_DifferentKeys(t *testing.T) {
	g := NewGroup()
	v1, _, _ := g.Do("keyA", func() (any, error) { return "A", nil })
	v2, _, _ := g.Do("keyB", func() (any, error) { return "B", nil })
	if v1 != "A" || v2 != "B" {
		t.Errorf("different keys should call fn independently: %v, %v", v1, v2)
	}
}

func TestForget(t *testing.T) {
	g := NewGroup()
	g.Forget("nonexistent") // Should not panic.

	_, _, _ = g.Do("forget-key", func() (any, error) { return "val", nil })
	g.Forget("forget-key")

	// After Forget, a new call should execute fn again.
	var calls int32
	_, _, _ = g.Do("forget-key", func() (any, error) {
		atomic.AddInt32(&calls, 1)
		return "new-val", nil
	})
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected fn to be called again after Forget, got %d calls", atomic.LoadInt32(&calls))
	}
}

func TestDoChan(t *testing.T) {
	g := NewGroup()
	ch := g.DoChan("chan-key", func() (any, error) {
		return "chan-result", nil
	})
	select {
	case result := <-ch:
		if result.Val != "chan-result" {
			t.Errorf("expected 'chan-result', got %v", result.Val)
		}
		if result.Err != nil {
			t.Errorf("unexpected error: %v", result.Err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("DoChan() timed out")
	}
}
