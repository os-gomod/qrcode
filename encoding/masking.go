// Package encoding implements QR code data masking and penalty scoring to select
// the optimal mask pattern for readability.
package encoding

type maskFunc func(row, col int) bool

var maskFunctions = [8]maskFunc{
	func(row, col int) bool { return (row+col)%2 == 0 },
	func(row, _ int) bool { return row%2 == 0 },
	func(_, col int) bool { return col%3 == 0 },
	func(row, col int) bool { return (row+col)%3 == 0 },
	func(row, col int) bool { return (row/2+col/3)%2 == 0 },
	func(row, col int) bool { return (row*col)%2+(row*col)%3 == 0 },
	func(row, col int) bool { return ((row*col)%2+(row*col)%3)%2 == 0 },
	func(row, col int) bool { return ((row+col)%2+(row*col)%3)%2 == 0 },
}

// ApplyMask XORs data modules with the specified mask pattern (0–7).
// Function-pattern modules (finders, alignment, timing, format/version info)
// are left unchanged. Modifies the matrix in place.
func ApplyMask(matrix [][]bool, maskPattern, version int) {
	if maskPattern < 0 || maskPattern > 7 {
		return
	}
	size := len(matrix)
	fn := maskFunctions[maskPattern]
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if isFunctionPattern(matrix, row, col, version, size) {
				continue
			}
			if fn(row, col) {
				matrix[row][col] = !matrix[row][col]
			}
		}
	}
}

// RemoveMask removes (toggles) a previously applied mask pattern by applying
// the same XOR operation again. Equivalent to [ApplyMask] with the same pattern.
func RemoveMask(matrix [][]bool, maskPattern, version int) {
	ApplyMask(matrix, maskPattern, version)
}

// PenaltyScore computes the total mask penalty score for the matrix using
// the four standard QR code penalty rules (N1–N4):
//   - N1: runs of same-color modules in rows/columns
//   - N2: 2×2 blocks of same-color modules
//   - N3: finder-like patterns
//   - N4: deviation from 50% dark-module ratio
func PenaltyScore(matrix [][]bool, _ int) int {
	size := len(matrix)
	score := 0
	score += penaltyN1(matrix, size)
	score += penaltyN2(matrix, size)
	score += penaltyN3(matrix, size)
	score += penaltyN4(matrix, size)
	return score
}

func penaltyN1(matrix [][]bool, size int) int {
	penalty := 0
	for row := 0; row < size; row++ {
		count := 1
		for col := 1; col < size; col++ {
			if matrix[row][col] == matrix[row][col-1] {
				count++
			} else {
				if count >= 5 {
					penalty += 3 + (count - 5)
				}
				count = 1
			}
		}
		if count >= 5 {
			penalty += 3 + (count - 5)
		}
	}
	for col := 0; col < size; col++ {
		count := 1
		for row := 1; row < size; row++ {
			if matrix[row][col] == matrix[row-1][col] {
				count++
			} else {
				if count >= 5 {
					penalty += 3 + (count - 5)
				}
				count = 1
			}
		}
		if count >= 5 {
			penalty += 3 + (count - 5)
		}
	}
	return penalty
}

func penaltyN2(matrix [][]bool, size int) int {
	penalty := 0
	for row := 0; row < size-1; row++ {
		for col := 0; col < size-1; col++ {
			val := matrix[row][col]
			//nolint:gocritic // badCond: standard QR code 2x2 block penalty check
			if val == matrix[row][col+1] &&
				val == matrix[row+1][col] &&
				val == matrix[row+1][col+1] {
				penalty += 3
			}
		}
	}
	return penalty
}

func penaltyN3(matrix [][]bool, size int) int {
	penalty := 0
	pattern1 := []bool{true, false, true, true, true, false, true, false, false, false, false}
	pattern2 := []bool{false, false, false, false, true, false, true, true, true, false, true}
	for row := 0; row < size; row++ {
		penalty += countPatternN3(matrix, row, true, size, pattern1, pattern2)
	}
	for col := 0; col < size; col++ {
		penalty += countPatternN3(matrix, col, false, size, pattern1, pattern2)
	}
	return penalty
}

func countPatternN3(matrix [][]bool, index int, isRow bool, size int, pattern1, pattern2 []bool) int {
	penalty := 0
	patLen := len(pattern1)
	for start := 0; start <= size-patLen; start++ {
		match1 := true
		for j := 0; j < patLen; j++ {
			var val bool
			if isRow {
				val = matrix[index][start+j]
			} else {
				val = matrix[start+j][index]
			}
			if val != pattern1[j] {
				match1 = false
				break
			}
		}
		match2 := true
		for j := 0; j < patLen; j++ {
			var val bool
			if isRow {
				val = matrix[index][start+j]
			} else {
				val = matrix[start+j][index]
			}
			if val != pattern2[j] {
				match2 = false
				break
			}
		}
		if !match1 && !match2 {
			continue
		}
		preOk := true
		for j := 1; j <= 4; j++ {
			pos := start - j
			if pos < 0 {
				preOk = false
				break
			}
			var val bool
			if isRow {
				val = matrix[index][pos]
			} else {
				val = matrix[pos][index]
			}
			if val {
				preOk = false
				break
			}
		}
		postOk := true
		for j := 1; j <= 4; j++ {
			pos := start + patLen - 1 + j
			if pos >= size {
				postOk = false
				break
			}
			var val bool
			if isRow {
				val = matrix[index][pos]
			} else {
				val = matrix[pos][index]
			}
			if val {
				postOk = false
				break
			}
		}
		if preOk && postOk {
			penalty += 40
		}
	}
	return penalty
}

func penaltyN4(matrix [][]bool, size int) int {
	total := size * size
	dark := 0
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if matrix[row][col] {
				dark++
			}
		}
	}
	percent := (dark * 100) / total
	deviation := abs(percent - 50)
	steps := (deviation + 4) / 5
	return steps * 10
}

// BestMaskPattern evaluates all eight mask patterns (0–7) and returns the
// one with the lowest total penalty score. For each candidate, it clones the
// base matrix, applies the mask, writes format/version info, and scores the
// result. The matrix passed in is not modified.
func BestMaskPattern(baseMatrix [][]bool, ecLevel, version int) int {
	bestMask := 0
	bestScore := -1
	for mask := 0; mask < 8; mask++ {
		matrix := CloneMatrix(baseMatrix)
		ApplyMask(matrix, mask, version)
		PlaceFormatInfo(matrix, ecLevel, mask)
		PlaceVersionInfo(matrix, version)
		score := PenaltyScore(matrix, version)
		if bestScore < 0 || score < bestScore {
			bestScore = score
			bestMask = mask
		}
	}
	return bestMask
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
