package hash

import "testing"

func TestHash(t *testing.T) {
	h1 := Hash("hello")
	if h1 == 0 {
		t.Error("Hash should not return 0 for non-empty input")
	}

	// Consistent
	h2 := Hash("hello")
	if h1 != h2 {
		t.Error("Hash should return consistent values")
	}

	// Different inputs should (likely) produce different hashes
	h3 := Hash("world")
	if h1 == h3 {
		t.Error("Hash should produce different values for different inputs")
	}

	// Empty string
	h4 := Hash("")
	// FNV of empty string is a known value
	if h4 == 0 {
		t.Error("Hash of empty string should be non-zero with fnv")
	}
}

func TestHashBytes(t *testing.T) {
	h1 := HashBytes([]byte("hello"))
	if h1 == 0 {
		t.Error("HashBytes should not return 0")
	}

	// Should match Hash for same string
	h2 := Hash("hello")
	if h1 != h2 {
		t.Error("HashBytes and Hash should return same value for same data")
	}

	// Empty
	h3 := HashBytes([]byte{})
	if h3 == 0 {
		t.Error("HashBytes of empty should be non-zero")
	}

	// Different data
	h4 := HashBytes([]byte{0x01, 0x02})
	h5 := HashBytes([]byte{0x02, 0x01})
	if h4 == h5 {
		t.Error("different byte slices should produce different hashes")
	}
}

func TestCombine(t *testing.T) {
	h1 := Hash("a")
	h2 := Hash("b")
	combined := Combine(h1, h2)

	if combined == 0 {
		t.Error("Combine should not return 0")
	}

	// Combine should be deterministic
	combined2 := Combine(h1, h2)
	if combined != combined2 {
		t.Error("Combine should be deterministic")
	}

	// Order matters
	reversed := Combine(h2, h1)
	if combined == reversed {
		// Not strictly required to be different, but good to check
		t.Log("Combine(a,b) == Combine(b,a) - not an error but noted")
	}

	// Identity-like behavior with zero
	zeroCombined := Combine(0, h1)
	if zeroCombined == 0 {
		t.Error("Combine with zero should not be zero (unless h1 counteracts)")
	}
}
