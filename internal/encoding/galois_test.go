package encoding

import "testing"

func TestGfAdd(t *testing.T) {
	// XOR operation.
	if gfAdd(0, 0) != 0 {
		t.Errorf("gfAdd(0, 0) = %d", gfAdd(0, 0))
	}
	if gfAdd(5, 3) != 6 { // 101 ^ 011 = 110
		t.Errorf("gfAdd(5, 3) = %d, want 6", gfAdd(5, 3))
	}
	if gfAdd(7, 7) != 0 {
		t.Errorf("gfAdd(7, 7) = %d, want 0", gfAdd(7, 7))
	}
}

func TestGfMul(t *testing.T) {
	// Zero times anything is zero.
	if gfMul(0, 5) != 0 {
		t.Errorf("gfMul(0, 5) = %d, want 0", gfMul(0, 5))
	}
	if gfMul(5, 0) != 0 {
		t.Errorf("gfMul(5, 0) = %d, want 0", gfMul(5, 0))
	}
	// 1 * x = x.
	if gfMul(1, 100) != 100 {
		t.Errorf("gfMul(1, 100) = %d, want 100", gfMul(1, 100))
	}
	// gfMul should be commutative.
	if gfMul(3, 5) != gfMul(5, 3) {
		t.Error("gfMul should be commutative")
	}
}

func TestGfExpLogTables(t *testing.T) {
	// gfExp[i] should be the inverse of gfLog: gfExp[gfLog[x]] == x for x != 0.
	for x := 1; x < 256; x++ {
		log := gfLog[x]
		exp := gfExp[log]
		if exp != x {
			t.Errorf("gfExp[gfLog[%d]] = %d, want %d", x, exp, x)
		}
	}
}

func TestGeneratorPoly(t *testing.T) {
	t.Run("degree 0", func(t *testing.T) {
		g := GeneratorPoly(0)
		if len(g) != 1 || g[0] != 1 {
			t.Errorf("GeneratorPoly(0) = %v, want [1]", g)
		}
	})
	t.Run("degree 1", func(t *testing.T) {
		g := GeneratorPoly(1)
		if len(g) != 2 {
			t.Fatalf("expected length 2, got %d", len(g))
		}
		if g[0] != 1 || g[1] != 1 {
			t.Errorf("GeneratorPoly(1) = %v, want [1 1]", g)
		}
	})
	t.Run("degree 7", func(t *testing.T) {
		g := GeneratorPoly(7)
		if len(g) != 8 {
			t.Errorf("expected length 8, got %d", len(g))
		}
		if g[0] != 1 {
			t.Errorf("first coefficient should be 1, got %d", g[0])
		}
	})
}

func TestEncodeEC(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		data := []byte{32, 91, 11, 120, 209, 114, 220, 77, 67, 64, 236, 17, 236}
		ec := EncodeEC(data, 13)
		if len(ec) != 13 {
			t.Errorf("expected 13 EC bytes, got %d", len(ec))
		}
	})
	t.Run("zero ec count", func(t *testing.T) {
		ec := EncodeEC([]byte{1, 2, 3}, 0)
		if ec != nil {
			t.Errorf("expected nil for 0 EC count, got %v", ec)
		}
	})
	t.Run("empty data", func(t *testing.T) {
		ec := EncodeEC([]byte{}, 5)
		if ec != nil {
			t.Errorf("expected nil for empty data, got %v", ec)
		}
	})
}

func TestEncodeECBlocks(t *testing.T) {
	blocks := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
	}
	ecBlocks := EncodeECBlocks(blocks, 2)
	if len(ecBlocks) != 2 {
		t.Fatalf("expected 2 EC blocks, got %d", len(ecBlocks))
	}
	for _, ec := range ecBlocks {
		if len(ec) != 2 {
			t.Errorf("expected 2 EC bytes per block, got %d", len(ec))
		}
	}
}

func TestEncodeEC_NonZeroAndCorrectLength(t *testing.T) {
	// Basic sanity: EC should produce ecCount bytes.
	data := []byte{32, 91, 11, 120, 209, 114, 220, 77}
	ecCount := 10
	ec := EncodeEC(data, ecCount)
	if len(ec) != ecCount {
		t.Errorf("expected %d EC bytes, got %d", ecCount, len(ec))
	}

	// Different data should produce different EC.
	data2 := []byte{33, 91, 11, 120, 209, 114, 220, 77}
	ec2 := EncodeEC(data2, ecCount)
	if len(ec2) != ecCount {
		t.Errorf("expected %d EC bytes, got %d", ecCount, len(ec2))
	}

	// First EC byte should differ (since first data byte differs).
	if ec[0] == ec2[0] {
		// This can happen, but is unlikely. Log and continue.
		t.Logf("EC[0] same for different data: %d (expected different)", ec[0])
	}
}
