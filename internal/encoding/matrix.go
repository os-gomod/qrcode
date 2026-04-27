package encoding

import "fmt"

var alignmentPositions = [40][]int{
	nil,
	{6, 18},
	{6, 22},
	{6, 26},
	{6, 30},
	{6, 34},
	{6, 22, 38},
	{6, 24, 42},
	{6, 26, 46},
	{6, 28, 50},
	{6, 30, 54},
	{6, 32, 58},
	{6, 34, 62},
	{6, 26, 46, 66},
	{6, 26, 48, 70},
	{6, 26, 50, 74},
	{6, 30, 54, 78},
	{6, 30, 56, 82},
	{6, 30, 58, 86},
	{6, 34, 62, 90},
	{6, 28, 50, 72, 94},
	{6, 26, 50, 74, 98},
	{6, 30, 54, 78, 102},
	{6, 28, 54, 80, 106},
	{6, 32, 58, 84, 110},
	{6, 30, 58, 86, 114},
	{6, 34, 62, 90, 118},
	{6, 26, 50, 74, 98, 122},
	{6, 30, 54, 78, 102, 126},
	{6, 26, 52, 78, 104, 130},
	{6, 30, 56, 82, 108, 134},
	{6, 34, 60, 86, 112, 138},
	{6, 30, 58, 86, 114, 142},
	{6, 34, 62, 90, 118, 146},
	{6, 30, 54, 78, 102, 126, 150},
	{6, 24, 50, 76, 102, 128, 154},
	{6, 28, 54, 80, 106, 132, 158},
	{6, 32, 58, 84, 110, 136, 162},
	{6, 26, 54, 82, 110, 138, 166},
	{6, 30, 58, 86, 114, 142, 170},
}

func NewMatrix(version int) [][]bool {
	size := version*4 + 17
	matrix := make([][]bool, size)
	for i := range matrix {
		matrix[i] = make([]bool, size)
	}
	return matrix
}

//nolint:gocyclo,cyclop // finder pattern placement requires boundary checks
func PlaceFinderPattern(matrix [][]bool, row, col int) {
	for r := -1; r <= 7; r++ {
		for c := -1; c <= 7; c++ {
			r2, c2 := row+r, col+c
			if r2 < 0 || r2 >= len(matrix) || c2 < 0 || c2 >= len(matrix) {
				continue
			}
			if r >= 0 && r <= 6 && c >= 0 && c <= 6 {
				if (r == 0 || r == 6 || c == 0 || c == 6) ||
					(r >= 2 && r <= 4 && c >= 2 && c <= 4) {
					matrix[r2][c2] = true
				}
			}
		}
	}
}

func PlaceAlignmentPattern(matrix [][]bool, _, row, col int) {
	for r := -2; r <= 2; r++ {
		for c := -2; c <= 2; c++ {
			r2, c2 := row+r, col+c
			if r2 < 0 || r2 >= len(matrix) || c2 < 0 || c2 >= len(matrix) {
				continue
			}
			if r == -2 || r == 2 || c == -2 || c == 2 || (r == 0 && c == 0) {
				matrix[r2][c2] = true
			}
		}
	}
}

func PlaceTimingPatterns(matrix [][]bool, version int) {
	size := version*4 + 17
	for i := 8; i < size-8; i++ {
		if i%2 == 0 {
			matrix[6][i] = true
		}
		if i%2 == 0 {
			matrix[i][6] = true
		}
	}
}

func PlaceDarkModule(matrix [][]bool, version int) {
	matrix[version*4+9][8] = true
}

func PlaceFormatInfo(matrix [][]bool, ecLevel, maskPattern int) {
	if ecLevel < 0 || ecLevel > 3 {
		ecLevel = 0
	}
	if maskPattern < 0 || maskPattern > 7 {
		maskPattern = 0
	}
	data := ecLevelFormatBits[ecLevel]<<3 | maskPattern
	bits := bchEncodeFormat(data)
	formatPositions1 := [][2]int{
		{8, 0},
		{8, 1},
		{8, 2},
		{8, 3},
		{8, 4},
		{8, 5},
		{8, 7},
		{8, 8},
		{7, 8},
		{5, 8},
		{4, 8},
		{3, 8},
		{2, 8},
		{1, 8},
		{0, 8},
	}
	size := len(matrix)
	for i, pos := range formatPositions1 {
		matrix[pos[0]][pos[1]] = (bits>>uint(i))&1 != 0
	}
	for i := range 7 {
		matrix[8][size-1-i] = (bits>>uint(i))&1 != 0
	}
	for i := 7; i < 15; i++ {
		matrix[size-15+i][8] = (bits>>uint(i))&1 != 0
	}
}

func bchEncodeFormat(data int) int {
	gen := 0x537
	encoded := data << 10
	remainder := encoded
	for i := 14; i >= 10; i-- {
		if remainder&(1<<uint(i)) != 0 {
			remainder ^= gen << uint(i-10)
		}
	}
	result := (data << 10) | remainder
	result ^= 0x5412
	return result
}

func PlaceVersionInfo(matrix [][]bool, version int) {
	if version < 7 || version > 40 {
		return
	}
	bits := bchEncodeVersion(version)
	size := len(matrix)
	for i := range 18 {
		row := size - 11 + (i % 3)
		col := (i / 3)
		matrix[row][col] = bits&(1<<(17-i)) != 0
	}
	for i := range 18 {
		row := i / 3
		col := size - 11 + (i % 3)
		matrix[row][col] = bits&(1<<(17-i)) != 0
	}
}

func bchEncodeVersion(version int) int {
	gen := 0x1F25
	encoded := version << 12
	remainder := encoded
	for i := 17; i >= 12; i-- {
		if remainder&(1<<uint(i)) != 0 {
			remainder ^= gen << uint(i-12)
		}
	}
	return (version << 12) | remainder
}

func PlaceDataBits(matrix [][]bool, dataBits []bool, version int) {
	size := version*4 + 17
	bitIndex := 0
	for col := size - 1; col >= 1; col -= 2 {
		if col == 6 {
			col = 5
		}
		upward := ((size-1-col)/2)%2 == 0
		for row := range size {
			var r int
			if upward {
				r = size - 1 - row
			} else {
				r = row
			}
			for dc := range 2 {
				c := col - dc
				if c < 0 {
					continue
				}
				if isFunctionPattern(matrix, r, c, version, size) {
					continue
				}
				if bitIndex < len(dataBits) {
					matrix[r][c] = dataBits[bitIndex]
					bitIndex++
				}
			}
		}
	}
}

//nolint:gocyclo,cyclop // function pattern detection requires comprehensive checks
func isFunctionPattern(_ [][]bool, row, col, version, size int) bool {
	if row <= 8 && col <= 8 {
		return true
	}
	if row <= 8 && col >= size-8 {
		return true
	}
	if row >= size-8 && col <= 8 {
		return true
	}
	if row == 6 || col == 6 {
		return true
	}
	if row == version*4+9 && col == 8 {
		return true
	}
	if version >= 7 {
		if row >= size-11 && row <= size-9 && col <= 5 {
			return true
		}
		if row <= 5 && col >= size-11 && col <= size-9 {
			return true
		}
	}
	positions := alignmentPositions[version-1]
	for _, pr := range positions {
		for _, pc := range positions {
			if pr <= 8 && pc <= 8 {
				continue
			}
			if pr <= 8 && pc >= size-8 {
				continue
			}
			if pr >= size-8 && pc <= 8 {
				continue
			}
			if row >= pr-2 && row <= pr+2 && col >= pc-2 && col <= pc+2 {
				return true
			}
		}
	}
	return false
}

func PlaceAllFinderPatterns(matrix [][]bool, version int) {
	size := version*4 + 17
	PlaceFinderPattern(matrix, 0, 0)
	PlaceFinderPattern(matrix, 0, size-7)
	PlaceFinderPattern(matrix, size-7, 0)
}

func PlaceAllAlignmentPatterns(matrix [][]bool, version int) {
	if version < 2 {
		return
	}
	positions := alignmentPositions[version-1]
	for _, row := range positions {
		for _, col := range positions {
			PlaceAlignmentPattern(matrix, version, row, col)
		}
	}
}

func BuildMatrix(version int) [][]bool {
	matrix := NewMatrix(version)
	PlaceAllFinderPatterns(matrix, version)
	PlaceAllAlignmentPatterns(matrix, version)
	PlaceTimingPatterns(matrix, version)
	PlaceDarkModule(matrix, version)
	return matrix
}

func MatrixSize(version int) int {
	return version*4 + 17
}

func CloneMatrix(matrix [][]bool) [][]bool {
	size := len(matrix)
	clone := make([][]bool, size)
	for i := range matrix {
		clone[i] = make([]bool, size)
		copy(clone[i], matrix[i])
	}
	return clone
}

func PrintMatrix(matrix [][]bool) string {
	size := len(matrix)
	var result []byte
	result = make([]byte, 0, size*(size+1))
	for r := range size {
		for c := range size {
			if matrix[r][c] {
				result = append(result, '#')
			} else {
				result = append(result, '.')
			}
		}
		result = append(result, '\n')
	}
	return string(result)
}

func VerifyMatrixSize(matrix [][]bool, version int) error {
	expected := version*4 + 17
	if len(matrix) != expected {
		return fmt.Errorf("matrix has %d rows, expected %d for version %d", len(matrix), expected, version)
	}
	for i, row := range matrix {
		if len(row) != expected {
			return fmt.Errorf("matrix row %d has %d columns, expected %d for version %d", i, len(row), expected, version)
		}
	}
	return nil
}
