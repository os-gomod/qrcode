package singleflight

import (
	"testing"
)

func TestNewGroup(t *testing.T) {
	g := NewGroup()
	if g == nil {
		t.Fatal("NewGroup returned nil")
	}
}

func TestDoBasic(t *testing.T) {
	g := NewGroup()

	val, shared, err := g.Do("key1", func() (any, error) {
		return "result1", nil
	})
	if err != nil {
		t.Fatalf("Do error: %v", err)
	}
	if val != "result1" {
		t.Errorf("val = %v, want result1", val)
	}
	if shared {
		t.Error("first call should not be shared")
	}
}

func TestDoSequential(t *testing.T) {
	g := NewGroup()

	// First call
	val1, _, err1 := g.Do("key1", func() (any, error) {
		return "first", nil
	})
	if err1 != nil || val1 != "first" {
		t.Fatalf("first call: val=%v err=%v", val1, err1)
	}

	// Second call (sequential, not concurrent)
	val2, shared2, err2 := g.Do("key1", func() (any, error) {
		return "second", nil
	})
	if err2 != nil {
		t.Fatalf("second call error: %v", err2)
	}
	_ = shared2
	if val2 != "second" {
		t.Errorf("second call val = %v, want second", val2)
	}
}

func TestDoError(t *testing.T) {
	g := NewGroup()

	_, shared, err := g.Do("errkey", func() (any, error) {
		return nil, &testError{}
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if shared {
		t.Error("error call should not be shared")
	}
}

func TestForget(t *testing.T) {
	g := NewGroup()

	g.Do("key1", func() (any, error) {
		return "result", nil
	})

	g.Forget("key1")
	// After forget, a new call should execute the function again
	calls := 0
	val, _, err := g.Do("key1", func() (any, error) {
		calls++
		return "result2", nil
	})
	if err != nil {
		t.Fatalf("Do error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call after Forget, got %d", calls)
	}
	if val != "result2" {
		t.Errorf("val = %v, want result2", val)
	}
}

func TestDoChan(t *testing.T) {
	g := NewGroup()

	ch := g.DoChan("chankey", func() (any, error) {
		return "chan_result", nil
	})

	result := <-ch
	if result.Err != nil {
		t.Fatalf("DoChan error: %v", result.Err)
	}
	if result.Val != "chan_result" {
		t.Errorf("Val = %v", result.Val)
	}
}

func TestDoMultipleKeys(t *testing.T) {
	g := NewGroup()

	v1, _, _ := g.Do("a", func() (any, error) { return 1, nil })
	v2, _, _ := g.Do("b", func() (any, error) { return 2, nil })
	v3, _, _ := g.Do("c", func() (any, error) { return 3, nil })

	if v1 != 1 || v2 != 2 || v3 != 3 {
		t.Errorf("expected 1,2,3 got %v,%v,%v", v1, v2, v3)
	}
}

type testError struct{}

func (e *testError) Error() string { return "test error" }
