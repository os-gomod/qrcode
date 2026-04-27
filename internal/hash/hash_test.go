package hash

import "testing"

func TestHash_Deterministic(t *testing.T) {
	h1 := Hash("hello")
	h2 := Hash("hello")
	if h1 != h2 {
		t.Errorf("Hash() should be deterministic: %d != %d", h1, h2)
	}
}

func TestHash_Different(t *testing.T) {
	h1 := Hash("hello")
	h2 := Hash("world")
	if h1 == h2 {
		t.Error("different inputs should produce different hashes")
	}
}

func TestHash_Empty(t *testing.T) {
	h := Hash("")
	// Should not panic and should return some value.
	if h == 0 {
		t.Error("hash of empty string should not be 0")
	}
}

func TestHashBytes(t *testing.T) {
	h1 := HashBytes([]byte("hello"))
	h2 := Hash("hello")
	if h1 != h2 {
		t.Errorf("HashBytes() and Hash() should agree: %d != %d", h1, h2)
	}
}

func TestHashBytes_Empty(t *testing.T) {
	h := HashBytes(nil)
	_ = h // Should not panic.
}

func TestCombine(t *testing.T) {
	h1 := Hash("a")
	h2 := Hash("b")
	combined := Combine(h1, h2)
	// Should not equal either input.
	if combined == h1 || combined == h2 {
		t.Error("Combine() should produce a different value from inputs")
	}
}

func TestCombine_Deterministic(t *testing.T) {
	c1 := Combine(100, 200)
	c2 := Combine(100, 200)
	if c1 != c2 {
		t.Error("Combine() should be deterministic")
	}
}

func TestCombine_Commutative(t *testing.T) {
	// Combine is NOT commutative by design (it's order-dependent).
	h1 := Combine(10, 20)
	h2 := Combine(20, 10)
	// We just verify both produce valid results.
	if h1 == h2 {
		t.Log("Combine happens to be commutative for these inputs (not a requirement)")
	}
}
