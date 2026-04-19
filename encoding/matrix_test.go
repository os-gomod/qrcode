package encoding

import "testing"

func TestNewMatrix(t *testing.T) {
	tests := []struct {
		version  int
		wantSize int
	}{
		{version: 1, wantSize: 21},
		{version: 2, wantSize: 25},
		{version: 10, wantSize: 57},
		{version: 40, wantSize: 177},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			m := NewMatrix(tt.version)
			if len(m) != tt.wantSize {
				t.Errorf("NewMatrix(%d): got %d rows, want %d", tt.version, len(m), tt.wantSize)
			}
			for i, row := range m {
				if len(row) != tt.wantSize {
					t.Errorf("NewMatrix(%d) row %d: got %d cols, want %d", tt.version, i, len(row), tt.wantSize)
				}
			}
		})
	}
}

func TestBuildMatrix(t *testing.T) {
	tests := []struct {
		version int
	}{
		{version: 1},
		{version: 2},
		{version: 7},
		{version: 10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			m := BuildMatrix(tt.version)
			expected := MatrixSize(tt.version)
			if len(m) != expected {
				t.Errorf("BuildMatrix(%d): got %d rows, want %d", tt.version, len(m), expected)
			}

			// Check that finder patterns are placed
			// Top-left: (0,0) should be set (part of outer ring)
			if !m[0][0] {
				t.Error("top-left finder pattern not set at (0,0)")
			}
			// Top-right: (0, size-1) should be set
			if !m[0][expected-1] {
				t.Error("top-right finder pattern not set")
			}
			// Bottom-left: (size-1, 0) should be set
			if !m[expected-1][0] {
				t.Error("bottom-left finder pattern not set")
			}
		})
	}
}

func TestPlaceFinderPattern(t *testing.T) {
	m := NewMatrix(10)
	PlaceFinderPattern(m, 0, 0)

	// Check outer corners
	if !m[0][0] {
		t.Error("outer corner (0,0) should be set")
	}
	if !m[0][6] {
		t.Error("outer corner (0,6) should be set")
	}
	if !m[6][0] {
		t.Error("outer corner (6,0) should be set")
	}
	if !m[6][6] {
		t.Error("outer corner (6,6) should be set")
	}
	// Center should be set
	if !m[3][3] {
		t.Error("center (3,3) should be set")
	}
}

func TestMatrixSize(t *testing.T) {
	tests := []struct {
		version int
		want    int
	}{
		{1, 21},
		{2, 25},
		{3, 29},
		{4, 33},
		{5, 37},
		{10, 57},
		{20, 97},
		{40, 177},
	}

	for _, tt := range tests {
		got := MatrixSize(tt.version)
		if got != tt.want {
			t.Errorf("MatrixSize(%d) = %d, want %d", tt.version, got, tt.want)
		}
	}
}

func TestCloneMatrix(t *testing.T) {
	m := BuildMatrix(1)
	PlaceDataBits(m, make([]bool, 100), 1)

	c := CloneMatrix(m)
	if !matricesEqual(m, c) {
		t.Error("CloneMatrix: clone should equal original")
	}

	// Modify clone, original should be unchanged
	c[10][10] = !c[10][10]
	if m[10][10] == c[10][10] {
		t.Error("modifying clone should not affect original")
	}
}

func TestPrintMatrix(t *testing.T) {
	m := BuildMatrix(1)
	s := PrintMatrix(m)
	if s == "" {
		t.Error("PrintMatrix should return non-empty string")
	}
}

func TestVerifyMatrixSize(t *testing.T) {
	m := BuildMatrix(1)
	if err := VerifyMatrixSize(m, 1); err != nil {
		t.Errorf("VerifyMatrixSize: unexpected error: %v", err)
	}

	bad := NewMatrix(2)
	if err := VerifyMatrixSize(bad, 1); err == nil {
		t.Error("VerifyMatrixSize should return error for wrong size")
	}
}

func TestVersionRemainderBits(t *testing.T) {
	tests := []struct {
		version int
		want    int
	}{
		{1, 0},
		{2, 7},
		{7, 0},
		{14, 3},
		{21, 4},
		{28, 3},
	}

	for _, tt := range tests {
		got := versionRemainderBits(tt.version)
		if got != tt.want {
			t.Errorf("versionRemainderBits(%d) = %d, want %d", tt.version, got, tt.want)
		}
	}
}
