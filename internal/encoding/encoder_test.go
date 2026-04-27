package encoding

import (
	"testing"
)

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		b    byte
		want bool
	}{
		{'0', true},
		{'9', true},
		{'A', false},
		{'a', false},
		{' ', false},
		{'/', false},
		{0, false},
		{255, false},
	}
	for _, tt := range tests {
		got := IsNumeric(tt.b)
		if got != tt.want {
			t.Errorf("IsNumeric(%q) = %v, want %v", tt.b, got, tt.want)
		}
	}
}

func TestIsAlphanumeric(t *testing.T) {
	tests := []struct {
		b    byte
		want bool
	}{
		{'0', true},
		{'9', true},
		{'A', true},
		{'Z', true},
		{' ', true},
		{'$', true},
		{'*', true},
		{'+', true},
		{'-', true},
		{'.', true},
		{'/', true},
		{':', true},
		{'a', false},
		{'z', false},
		{'@', false},
		{'#', false},
		{0, false},
	}
	for _, tt := range tests {
		got := IsAlphanumeric(tt.b)
		if got != tt.want {
			t.Errorf("IsAlphanumeric(%q) = %v, want %v", tt.b, got, tt.want)
		}
	}
}

func TestAlphanumericValue(t *testing.T) {
	val, ok := AlphanumericValue('A')
	if !ok || val != 10 {
		t.Errorf("AlphanumericValue('A') = %d, %v; want 10, true", val, ok)
	}
	val, ok = AlphanumericValue('Z')
	if !ok || val != 35 {
		t.Errorf("AlphanumericValue('Z') = %d, %v; want 35, true", val, ok)
	}
	_, ok = AlphanumericValue('a')
	if ok {
		t.Error("AlphanumericValue('a') should return false")
	}
}

func TestBestEncodingMode(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{"numeric", []byte("12345"), ModeNumeric},
		{"alphanumeric", []byte("HELLO WORLD"), ModeAlphanumeric},
		{"byte", []byte("hello world"), ModeByte},
		{"empty", []byte{}, ModeByte},
		{"kanji", []byte{0x82, 0xA0}, ModeKanji},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BestEncodingMode(tt.data)
			if got != tt.want {
				t.Errorf("BestEncodingMode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestEncode_Basic(t *testing.T) {
	qr, err := Encode([]byte("Hello"), 1, 0) // EC L, auto version.
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if qr == nil {
		t.Fatal("Encode() returned nil")
	}
	if qr.Version < 1 || qr.Version > 40 {
		t.Errorf("invalid version %d", qr.Version)
	}
	if qr.Size < 21 {
		t.Errorf("invalid size %d", qr.Size)
	}
	if qr.ECLevel != 1 {
		t.Errorf("expected EC level 1 (M), got %d", qr.ECLevel)
	}
	if qr.MaskPattern < 0 || qr.MaskPattern > 7 {
		t.Errorf("invalid mask pattern %d", qr.MaskPattern)
	}
	if qr.Modules == nil {
		t.Fatal("Modules should not be nil")
	}
	if len(qr.Modules) != qr.Size {
		t.Errorf("Modules rows = %d, want %d", len(qr.Modules), qr.Size)
	}
	if len(qr.Data) != 5 {
		t.Errorf("Data = %q, want 'Hello'", qr.Data)
	}
}

func TestEncode_WithVersion(t *testing.T) {
	qr, err := Encode([]byte("Hello"), 1, 2)
	if err != nil {
		t.Fatalf("Encode() error: %v", err)
	}
	if qr.Version != 2 {
		t.Errorf("expected version 2, got %d", qr.Version)
	}
}

func TestEncode_InvalidECLevel(t *testing.T) {
	_, err := Encode([]byte("Hello"), 5, 0)
	if err == nil {
		t.Error("expected error for invalid EC level")
	}
}

func TestEncode_EmptyData(t *testing.T) {
	_, err := Encode([]byte{}, 1, 0)
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestEncode_InvalidVersion(t *testing.T) {
	_, err := Encode([]byte("Hello"), 1, 50)
	if err == nil {
		t.Error("expected error for version > 40")
	}
	_, err = Encode([]byte("Hello"), 1, -1)
	if err == nil {
		t.Error("expected error for version < 1")
	}
}

func TestEncodeWithMask_Basic(t *testing.T) {
	qr, err := EncodeWithMask([]byte("Test"), 1, 0, 0)
	if err != nil {
		t.Fatalf("EncodeWithMask() error: %v", err)
	}
	if qr.MaskPattern != 0 {
		t.Errorf("expected mask 0, got %d", qr.MaskPattern)
	}
}

func TestEncodeWithMask_Auto(t *testing.T) {
	qr, err := EncodeWithMask([]byte("Test"), 1, 0, -1)
	if err != nil {
		t.Fatalf("EncodeWithMask() auto error: %v", err)
	}
	if qr.MaskPattern < 0 || qr.MaskPattern > 7 {
		t.Errorf("invalid auto mask pattern %d", qr.MaskPattern)
	}
}

func TestEncodeWithMask_InvalidMask(t *testing.T) {
	_, err := EncodeWithMask([]byte("Test"), 1, 0, 8)
	if err == nil {
		t.Error("expected error for mask > 7")
	}
}

func TestEncodeWithMask_InvalidECLevel(t *testing.T) {
	_, err := EncodeWithMask([]byte("Test"), -1, 0, 0)
	if err == nil {
		t.Error("expected error for invalid EC level")
	}
}

func TestEncode_Deterministic(t *testing.T) {
	qr1, _ := Encode([]byte("SameData"), 2, 5)
	qr2, _ := Encode([]byte("SameData"), 2, 5)
	if qr1.Version != qr2.Version || qr1.Size != qr2.Size || qr1.MaskPattern != qr2.MaskPattern {
		t.Error("same input should produce same output")
	}
	for r := 0; r < qr1.Size; r++ {
		for c := 0; c < qr1.Size; c++ {
			if qr1.Modules[r][c] != qr2.Modules[r][c] {
				t.Errorf("modules differ at [%d][%d]", r, c)
			}
		}
	}
}

func TestEncode_Metadata(t *testing.T) {
	qr, _ := Encode([]byte("meta-test"), 1, 0)
	if qr.Metadata == nil {
		t.Fatal("metadata should not be nil")
	}
	if qr.Metadata["mode"] != "Byte" {
		t.Errorf("expected mode=Byte, got %s", qr.Metadata["mode"])
	}
	if qr.Metadata["ecLevel"] != "M" {
		t.Errorf("expected ecLevel=M, got %s", qr.Metadata["ecLevel"])
	}
}

func TestEncode_NumericMode(t *testing.T) {
	qr, err := Encode([]byte("1234567890"), 0, 0)
	if err != nil {
		t.Fatalf("numeric encode error: %v", err)
	}
	if qr.Metadata["mode"] != "Numeric" {
		t.Errorf("expected Numeric mode, got %s", qr.Metadata["mode"])
	}
}

func TestEncode_AlphanumericMode(t *testing.T) {
	qr, err := Encode([]byte("HELLO WORLD"), 1, 0)
	if err != nil {
		t.Fatalf("alphanumeric encode error: %v", err)
	}
	if qr.Metadata["mode"] != "Alphanumeric" {
		t.Errorf("expected Alphanumeric mode, got %s", qr.Metadata["mode"])
	}
}

func TestCharCountBits(t *testing.T) {
	tests := []struct {
		mode    int
		version int
		want    int
	}{
		{ModeNumeric, 1, 10},
		{ModeNumeric, 10, 12},
		{ModeAlphanumeric, 1, 9},
		{ModeAlphanumeric, 10, 11},
		{ModeByte, 1, 8},
		{ModeByte, 10, 16},
		{ModeKanji, 1, 8},
		{ModeKanji, 10, 10},
		{ModeByte, 0, 8}, // default case
	}
	for _, tt := range tests {
		got := charCountBits(tt.mode, tt.version)
		if got != tt.want {
			t.Errorf("charCountBits(%d, %d) = %d, want %d", tt.mode, tt.version, got, tt.want)
		}
	}
}

func TestBitsToBytes_BytesToBits_Roundtrip(t *testing.T) {
	original := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	bits := bytesToBits(original)
	recovered := bitsToBytes(bits)
	if len(recovered) != len(original) {
		t.Fatalf("length mismatch: %d != %d", len(recovered), len(original))
	}
	for i := range original {
		if recovered[i] != original[i] {
			t.Errorf("byte %d: %02X != %02X", i, recovered[i], original[i])
		}
	}
}

func TestAppendBits(t *testing.T) {
	var bits []bool
	appendBits(&bits, 0xA, 4) // 1010
	if len(bits) != 4 {
		t.Fatalf("expected 4 bits, got %d", len(bits))
	}
	if bits[0] != true || bits[1] != false || bits[2] != true || bits[3] != false {
		t.Errorf("bits = %v, want [true false true false]", bits)
	}
}

func TestSplitIntoBlocks(t *testing.T) {
	// Version 1, EC L: 1 block of 19 data codewords.
	data := make([]byte, 19)
	for i := range data {
		data[i] = byte(i)
	}
	vi := &VersionInfo{
		Group1Blocks:        1,
		Group1DataCodewords: 19,
	}
	blocks := splitIntoBlocks(data, vi)
	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	if len(blocks[0]) != 19 {
		t.Errorf("expected 19 bytes, got %d", len(blocks[0]))
	}
}

func TestSplitIntoBlocks_TwoGroups(t *testing.T) {
	data := make([]byte, 100)
	vi := &VersionInfo{
		Group1Blocks:        2,
		Group1DataCodewords: 30,
		Group2Blocks:        2,
		Group2DataCodewords: 20,
	}
	blocks := splitIntoBlocks(data, vi)
	if len(blocks) != 4 {
		t.Fatalf("expected 4 blocks, got %d", len(blocks))
	}
	if len(blocks[0]) != 30 || len(blocks[2]) != 20 {
		t.Errorf("block sizes wrong: %d, %d", len(blocks[0]), len(blocks[2]))
	}
}

func TestModeName(t *testing.T) {
	if modeName(ModeNumeric) != "Numeric" {
		t.Errorf("got %q", modeName(ModeNumeric))
	}
	if modeName(ModeAlphanumeric) != "Alphanumeric" {
		t.Errorf("got %q", modeName(ModeAlphanumeric))
	}
	if modeName(ModeByte) != "Byte" {
		t.Errorf("got %q", modeName(ModeByte))
	}
	if modeName(ModeKanji) != "Kanji" {
		t.Errorf("got %q", modeName(ModeKanji))
	}
	if modeName(99) != "Unknown" {
		t.Errorf("got %q", modeName(99))
	}
}

func TestECLevelName(t *testing.T) {
	if ecLevelName(ECLevelL) != "L" {
		t.Errorf("got %q", ecLevelName(ECLevelL))
	}
	if ecLevelName(ECLevelM) != "M" {
		t.Errorf("got %q", ecLevelName(ECLevelM))
	}
	if ecLevelName(ECLevelQ) != "Q" {
		t.Errorf("got %q", ecLevelName(ECLevelQ))
	}
	if ecLevelName(ECLevelH) != "H" {
		t.Errorf("got %q", ecLevelName(ECLevelH))
	}
	if ecLevelName(99) != "Unknown" {
		t.Errorf("got %q", ecLevelName(99))
	}
}

func TestCapacity(t *testing.T) {
	// Version 1, EC L, Byte mode should have a known capacity.
	capByte := Capacity(1, ECLevelL, ModeByte)
	if capByte <= 0 {
		t.Errorf("Capacity(1, L, Byte) = %d, want > 0", capByte)
	}
	// Numeric should hold more characters than byte.
	capNum := Capacity(1, ECLevelL, ModeNumeric)
	if capNum <= capByte {
		t.Errorf("Numeric capacity (%d) should be > Byte capacity (%d)", capNum, capByte)
	}
}

func TestDataCodewords(t *testing.T) {
	d := DataCodewords(1, ECLevelL)
	if d <= 0 {
		t.Errorf("DataCodewords(1, L) = %d, want > 0", d)
	}
}

func TestECCodewords(t *testing.T) {
	ec := ECCodewords(1, ECLevelH)
	if ec <= 0 {
		t.Errorf("ECCodewords(1, H) = %d, want > 0", ec)
	}
	// Higher EC level should have more EC codewords.
	ecL := ECCodewords(1, ECLevelL)
	if ec <= ecL {
		t.Errorf("EC H (%d) should have more EC codewords than EC L (%d)", ec, ecL)
	}
}

func TestBlockCount(t *testing.T) {
	g1, g2 := BlockCount(1, ECLevelL)
	if g1 <= 0 {
		t.Errorf("Group1Blocks = %d, want > 0", g1)
	}
	_ = g2
}

func TestAddTerminatorAndPadding(t *testing.T) {
	tests := []struct {
		name      string
		totalBits int
		inputLen  int
		expectLen int
	}{
		{"exact", 24, 20, 24},
		{"short", 40, 8, 40},
		{"very_short", 8, 4, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bits := make([]bool, tt.inputLen)
			for i := range bits {
				bits[i] = true
			}
			addTerminatorAndPadding(&bits, tt.totalBits)
			if len(bits) != tt.expectLen {
				t.Errorf("expected %d bits, got %d", tt.expectLen, len(bits))
			}
		})
	}
}

func TestEncode_LargeData(t *testing.T) {
	// A large but valid payload.
	data := make([]byte, 500)
	for i := range data {
		data[i] = byte('A' + (i % 26))
	}
	qr, err := Encode(data, 1, 0)
	if err != nil {
		t.Fatalf("large data encode error: %v", err)
	}
	if qr.Version < 5 {
		t.Errorf("expected higher version for large data, got %d", qr.Version)
	}
}

func TestInterleaveBlocks_Simple(t *testing.T) {
	dataBlocks := [][]byte{
		{1, 2, 3},
		{4, 5, 6},
	}
	ecBlocks := [][]byte{
		{7, 8},
		{9, 10},
	}
	vi := &VersionInfo{
		Group1DataCodewords: 3,
		ECCodewordsPerBlock: 2,
	}
	result := interleaveBlocks(dataBlocks, ecBlocks, vi)
	// Expected: [1,4,2,5,3,6,7,9,8,10]
	expected := []byte{1, 4, 2, 5, 3, 6, 7, 9, 8, 10}
	if len(result) != len(expected) {
		t.Fatalf("length: got %d, want %d", len(result), len(expected))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("result[%d] = %d, want %d", i, result[i], expected[i])
		}
	}
}
