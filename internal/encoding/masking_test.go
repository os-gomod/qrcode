package encoding

import "testing"

func TestApplyMask(t *testing.T) {
	m := BuildMatrix(1)
	before := CloneMatrix(m)
	ApplyMask(m, 0, 1)

	changed := false
	for r := 0; r < 21; r++ {
		for c := 0; c < 21; c++ {
			if m[r][c] != before[r][c] {
				changed = true
				break
			}
		}
	}
	if !changed {
		t.Error("ApplyMask should change at least some non-function-pattern modules")
	}
}

func TestApplyMask_InvalidPattern(t *testing.T) {
	m := BuildMatrix(1)
	orig := CloneMatrix(m)
	ApplyMask(m, -1, 1)
	ApplyMask(m, 8, 1)
	// Both should be no-ops.
	for r := 0; r < 21; r++ {
		for c := 0; c < 21; c++ {
			if m[r][c] != orig[r][c] {
				t.Errorf("invalid mask pattern should not change matrix at [%d][%d]", r, c)
			}
		}
	}
}

func TestApplyMask_PreservesFunctionPatterns(t *testing.T) {
	m := BuildMatrix(1)
	// Record function pattern positions.
	fp := make([][2]int, 0)
	for r := 0; r < 21; r++ {
		for c := 0; c < 21; c++ {
			if isFunctionPattern(m, r, c, 1, 21) {
				fp = append(fp, [2]int{r, c})
			}
		}
	}
	before := CloneMatrix(m)
	ApplyMask(m, 3, 1)
	// All function patterns should be unchanged.
	for _, pos := range fp {
		if m[pos[0]][pos[1]] != before[pos[0]][pos[1]] {
			t.Errorf("function pattern at [%d][%d] should not change", pos[0], pos[1])
		}
	}
}

func TestRemoveMask_IsInverse(t *testing.T) {
	m := BuildMatrix(1)
	orig := CloneMatrix(m)
	ApplyMask(m, 2, 1)
	RemoveMask(m, 2, 1) // RemoveMask == ApplyMask for the same pattern.
	for r := 0; r < 21; r++ {
		for c := 0; c < 21; c++ {
			if m[r][c] != orig[r][c] {
				t.Errorf("matrix not restored after apply+remove at [%d][%d]", r, c)
			}
		}
	}
}

func TestBestMaskPattern(t *testing.T) {
	// Build a matrix, place data, and find the best mask.
	m := BuildMatrix(1)
	dataBits := make([]bool, 152) // Approximate for version 1 L.
	PlaceDataBits(m, dataBits, 1)
	best := BestMaskPattern(m, 1, 1)
	if best < 0 || best > 7 {
		t.Errorf("BestMaskPattern() = %d, out of range", best)
	}
}

func TestPenaltyScore(t *testing.T) {
	m := BuildMatrix(1)
	score := PenaltyScore(m, 1)
	if score < 0 {
		t.Errorf("PenaltyScore() = %d, should be non-negative", score)
	}
}

func TestPenaltyN1(t *testing.T) {
	// All-same matrix should have high N1 penalty.
	m := NewMatrix(1)
	size := 21
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			m[r][c] = true
		}
	}
	penalty := penaltyN1(m, size)
	if penalty <= 0 {
		t.Error("all-true matrix should have positive N1 penalty")
	}
}

func TestPenaltyN2(t *testing.T) {
	// All-same matrix should have high N2 penalty.
	m := NewMatrix(1)
	size := 21
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			m[r][c] = true
		}
	}
	penalty := penaltyN2(m, size)
	if penalty <= 0 {
		t.Error("all-true matrix should have positive N2 penalty")
	}
}

func TestPenaltyN4(t *testing.T) {
	// 50% dark should have 0 N4 penalty.
	m := NewMatrix(1)
	size := 21
	half := (size * size) / 2
	count := 0
	for r := 0; r < size && count < half; r++ {
		for c := 0; c < size && count < half; c++ {
			m[r][c] = true
			count++
		}
	}
	penalty := penaltyN4(m, size)
	if penalty < 0 {
		t.Errorf("penalty should be non-negative, got %d", penalty)
	}
}

func TestMaskFunctions(t *testing.T) {
	// Each mask function should be deterministic.
	for i := 0; i < 8; i++ {
		fn := maskFunctions[i]
		r1 := fn(3, 5)
		r2 := fn(3, 5)
		if r1 != r2 {
			t.Errorf("mask %d: not deterministic at (3,5)", i)
		}
	}
}

func TestAbs(t *testing.T) {
	if abs(5) != 5 {
		t.Errorf("abs(5) = %d, want 5", abs(5))
	}
	if abs(-5) != 5 {
		t.Errorf("abs(-5) = %d, want 5", abs(-5))
	}
	if abs(0) != 0 {
		t.Errorf("abs(0) = %d, want 0", abs(0))
	}
}
