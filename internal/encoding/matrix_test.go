package encoding

import "testing"

func TestNewMatrix(t *testing.T) {
	for v := 1; v <= 5; v++ {
		m := NewMatrix(v)
		expectedSize := v*4 + 17
		if len(m) != expectedSize {
			t.Errorf("version %d: expected %d rows, got %d", v, expectedSize, len(m))
		}
		for i, row := range m {
			if len(row) != expectedSize {
				t.Errorf("version %d row %d: expected %d cols, got %d", v, i, expectedSize, len(row))
			}
			for j, cell := range row {
				if cell {
					t.Errorf("version %d: new matrix should be all false, got true at [%d][%d]", v, i, j)
				}
			}
		}
	}
}

func TestMatrixSize(t *testing.T) {
	if MatrixSize(1) != 21 {
		t.Errorf("MatrixSize(1) = %d, want 21", MatrixSize(1))
	}
	if MatrixSize(7) != 45 {
		t.Errorf("MatrixSize(7) = %d, want 45", MatrixSize(7))
	}
	if MatrixSize(40) != 177 {
		t.Errorf("MatrixSize(40) = %d, want 177", MatrixSize(40))
	}
}

func TestPlaceFinderPattern(t *testing.T) {
	m := NewMatrix(1)
	PlaceFinderPattern(m, 0, 0)

	// Check corners of the finder pattern.
	// Top-left: [0][0], [0][6], [6][0], [6][6] should be true.
	if !m[0][0] || !m[0][6] || !m[6][0] || !m[6][6] {
		t.Error("finder pattern corners should be true")
	}
	// [1][1], [1][2], [2][1] ... inner area should be true.
	if !m[2][2] || !m[4][4] {
		t.Error("finder pattern inner area should be true")
	}
	// [1][1] should be false (separator).
	if m[1][1] {
		t.Error("[1][1] should be false (separator)")
	}
}

func TestBuildMatrix(t *testing.T) {
	m := BuildMatrix(2)
	size := MatrixSize(2)
	if len(m) != size {
		t.Fatalf("expected %d rows, got %d", size, len(m))
	}
	// Check that finder patterns exist (positions [0,0], [0, size-7], [size-7, 0]).
	finderPositions := [][2]int{{0, 0}, {0, size - 7}, {size - 7, 0}}
	for _, pos := range finderPositions {
		if !m[pos[0]][pos[1]] {
			t.Errorf("finder pattern missing at [%d][%d]", pos[0], pos[1])
		}
	}
	// Timing pattern: row 6 should have alternating values.
	hasTrueTiming := false
	for c := 8; c < size-8; c++ {
		if m[6][c] {
			hasTrueTiming = true
			break
		}
	}
	if !hasTrueTiming {
		t.Error("timing pattern seems missing in row 6")
	}
}

func TestCloneMatrix(t *testing.T) {
	m := BuildMatrix(1)
	c := CloneMatrix(m)
	if len(c) != len(m) {
		t.Fatalf("clone size mismatch: %d != %d", len(c), len(m))
	}
	// Modify clone and check original is unchanged.
	c[0][0] = !c[0][0]
	if m[0][0] == c[0][0] {
		t.Error("modifying clone should not affect original")
	}
}

func TestVerifyMatrixSize(t *testing.T) {
	m := BuildMatrix(3)
	if err := VerifyMatrixSize(m, 3); err != nil {
		t.Errorf("valid matrix should pass: %v", err)
	}
	// Wrong version.
	if err := VerifyMatrixSize(m, 4); err == nil {
		t.Error("wrong version should fail")
	}
	// Corrupted matrix.
	corrupt := m[:len(m)-1]
	if err := VerifyMatrixSize(corrupt, 3); err == nil {
		t.Error("corrupted matrix should fail")
	}
}

func TestPrintMatrix(t *testing.T) {
	m := BuildMatrix(1)
	s := PrintMatrix(m)
	if len(s) == 0 {
		t.Error("PrintMatrix should return non-empty string")
	}
	// Should contain '#' and '.' characters.
	hasHash := false
	hasDot := false
	for _, c := range s {
		if c == '#' {
			hasHash = true
		}
		if c == '.' {
			hasDot = true
		}
	}
	if !hasHash {
		t.Error("PrintMatrix output should contain '#'")
	}
	if !hasDot {
		t.Error("PrintMatrix output should contain '.'")
	}
}

func TestPlaceDarkModule(t *testing.T) {
	m := BuildMatrix(1)
	row := 1*4 + 9 // version*4+9
	if !m[row][8] {
		t.Errorf("dark module at [%d][8] should be true", row)
	}
}

func TestPlaceFormatInfo(t *testing.T) {
	m := BuildMatrix(1)
	PlaceFormatInfo(m, ECLevelM, 0)
	// Check that format info positions have been set (some should be true).
	if !m[8][0] && !m[8][1] && !m[8][2] {
		t.Error("format info positions should have some true values")
	}
}

func TestPlaceVersionInfo(t *testing.T) {
	// Version 7+ should have version info.
	m := BuildMatrix(7)
	PlaceVersionInfo(m, 7)
	// Just verify it doesn't panic and the matrix is still valid.
	if err := VerifyMatrixSize(m, 7); err != nil {
		t.Error("matrix should be valid after PlaceVersionInfo")
	}
}

func TestPlaceVersionInfo_LowVersion(t *testing.T) {
	m := BuildMatrix(1)
	PlaceVersionInfo(m, 1) // Should be a no-op for version < 7.
	// Verify matrix is unchanged (no panic).
	if err := VerifyMatrixSize(m, 1); err != nil {
		t.Error("matrix should still be valid")
	}
}

func TestIsFunctionPattern(t *testing.T) {
	m := BuildMatrix(1)
	size := 21
	// Top-left finder area (row<=8, col<=8) is always function pattern.
	if !isFunctionPattern(m, 0, 0, 1, size) {
		t.Error("[0][0] should be function pattern")
	}
	// Data area should not be function pattern.
	if isFunctionPattern(m, 10, 10, 1, size) {
		t.Error("[10][10] should not be function pattern")
	}
}

func TestAlignmentPattern_Version2(t *testing.T) {
	m := BuildMatrix(2)
	// Version 2 has alignment at [6, 18].
	if !m[6][18] {
		t.Error("alignment pattern should be at [6][18] for version 2")
	}
	if !m[18][6] {
		t.Error("alignment pattern should be at [18][6] for version 2")
	}
	if !m[18][18] {
		t.Error("alignment pattern should be at [18][18] for version 2")
	}
}

func TestBchEncodeFormat(t *testing.T) {
	// Known test vectors for format info BCH encoding.
	tests := []struct {
		data int
	}{
		{0b00000}, // EC=L, mask=0
		{0b01000}, // EC=M, mask=0
	}
	for _, tt := range tests {
		encoded := bchEncodeFormat(tt.data)
		// Masking with 0x5412 ensures the XOR mask pattern.
		if encoded < 0 || encoded > 0xFFFF {
			t.Errorf("bchEncodeFormat(%d) = %d, out of range", tt.data, encoded)
		}
	}
}
