package encoding

import "testing"

func TestApplyMask(t *testing.T) {
	version := 1
	matrix := BuildMatrix(version)
	PlaceDataBits(matrix, make([]bool, 100), version)
	original := CloneMatrix(matrix)

	for mask := 0; mask < 8; mask++ {
		t.Run("", func(t *testing.T) {
			cloned := CloneMatrix(original)
			ApplyMask(cloned, mask, 1)

			// Function pattern areas should remain unchanged
			// Check top-left finder pattern area (0-8, 0-8)
			changed := false
			for r := 0; r < 8; r++ {
				for c := 0; c < 8; c++ {
					if cloned[r][c] != original[r][c] {
						changed = true
					}
				}
			}
			// Finder pattern should not be modified
			if changed {
				t.Errorf("mask %d: function pattern modified", mask)
			}
		})
	}
}

func TestApplyMaskInvalidPattern(t *testing.T) {
	matrix := BuildMatrix(1)
	original := CloneMatrix(matrix)

	ApplyMask(matrix, -1, 1)
	if !matricesEqual(matrix, original) {
		t.Error("ApplyMask with -1 should not modify matrix")
	}

	ApplyMask(matrix, 8, 1)
	if !matricesEqual(matrix, original) {
		t.Error("ApplyMask with 8 should not modify matrix")
	}
}

func TestRemoveMask(t *testing.T) {
	matrix := BuildMatrix(1)
	PlaceDataBits(matrix, make([]bool, 100), 1)
	original := CloneMatrix(matrix)

	for mask := 0; mask < 4; mask++ {
		t.Run("", func(t *testing.T) {
			cloned := CloneMatrix(original)
			ApplyMask(cloned, mask, 1)
			RemoveMask(cloned, mask, 1)

			if !matricesEqual(cloned, original) {
				t.Errorf("mask %d: RemoveMask did not reverse ApplyMask", mask)
			}
		})
	}
}

func TestPenaltyScore(t *testing.T) {
	matrix := BuildMatrix(1)
	PlaceDataBits(matrix, make([]bool, 100), 1)

	PlaceFormatInfo(matrix, ECLevelM, 0)
	score := PenaltyScore(matrix, 1)
	if score < 0 {
		t.Errorf("PenaltyScore returned negative: %d", score)
	}

	// Test with all-zero matrix
	zero := NewMatrix(1)
	scoreZero := PenaltyScore(zero, 1)
	if scoreZero < 0 {
		t.Errorf("PenaltyScore returned negative for zero matrix: %d", scoreZero)
	}
}

func TestBestMaskPattern(t *testing.T) {
	version := 1
	matrix := BuildMatrix(version)
	PlaceDataBits(matrix, make([]bool, 100), version)

	best := BestMaskPattern(matrix, ECLevelM, version)
	if best < 0 || best > 7 {
		t.Errorf("BestMaskPattern = %d, want 0-7", best)
	}
}

func matricesEqual(a, b [][]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for r := range a {
		if len(a[r]) != len(b[r]) {
			return false
		}
		for c := range a[r] {
			if a[r][c] != b[r][c] {
				return false
			}
		}
	}
	return true
}
