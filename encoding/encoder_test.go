package encoding

import (
	"strings"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		ecLevel   int
		version   int
		wantErr   bool
		errSubstr string
	}{
		{name: "hello world text", data: []byte("Hello, World!"), ecLevel: ECLevelM, version: 0},
		{name: "numeric data", data: []byte("0123456789"), ecLevel: ECLevelL, version: 0},
		{name: "alphanumeric data", data: []byte("HELLO WORLD"), ecLevel: ECLevelL, version: 0},
		{name: "specific version 5", data: []byte("Version5 test data here"), ecLevel: ECLevelM, version: 5},
		{name: "specific version 1", data: []byte("Hi"), ecLevel: ECLevelL, version: 1},
		{name: "all EC levels", data: []byte("EC test"), ecLevel: ECLevelH, version: 0},
		{name: "empty data", data: []byte{}, ecLevel: ECLevelM, version: 0, wantErr: true, errSubstr: "empty"},
		{name: "invalid EC level", data: []byte("test"), ecLevel: 5, version: 0, wantErr: true, errSubstr: "invalid EC level"},
		{name: "invalid version low", data: []byte("test"), ecLevel: ECLevelM, version: -1, wantErr: true, errSubstr: "invalid version"},
		{name: "invalid version high", data: []byte("test"), ecLevel: ECLevelM, version: 41, wantErr: true, errSubstr: "invalid version"},
		{name: "EC level -1", data: []byte("test"), ecLevel: -1, version: 0, wantErr: true, errSubstr: "invalid EC level"},
		{name: "short binary data", data: []byte{0x80, 0x90}, ecLevel: ECLevelM, version: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qr, err := Encode(tt.data, tt.ecLevel, tt.version)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errSubstr)
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("expected error containing %q, got %q", tt.errSubstr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if qr == nil {
				t.Fatal("expected non-nil QRCode")
			}
			if qr.Modules == nil {
				t.Fatal("expected non-nil Modules")
			}
			expectedSize := MatrixSize(qr.Version)
			if qr.Size != expectedSize {
				t.Fatalf("expected Size %d, got %d", expectedSize, qr.Size)
			}
			if len(qr.Modules) != qr.Size {
				t.Fatalf("expected %d rows in Modules, got %d", qr.Size, len(qr.Modules))
			}
			for i, row := range qr.Modules {
				if len(row) != qr.Size {
					t.Fatalf("row %d: expected %d cols, got %d", i, qr.Size, len(row))
				}
			}
		})
	}
}

func TestEncodeWithMask(t *testing.T) {
	tests := []struct {
		name        string
		data        []byte
		ecLevel     int
		version     int
		maskPattern int
		wantErr     bool
	}{
		{name: "auto mask", data: []byte("test"), ecLevel: ECLevelM, version: 0, maskPattern: -1},
		{name: "mask 0", data: []byte("test"), ecLevel: ECLevelM, version: 0, maskPattern: 0},
		{name: "mask 7", data: []byte("test"), ecLevel: ECLevelM, version: 0, maskPattern: 7},
		{name: "invalid mask -2", data: []byte("test"), ecLevel: ECLevelM, version: 0, maskPattern: -2, wantErr: true},
		{name: "invalid mask 8", data: []byte("test"), ecLevel: ECLevelM, version: 0, maskPattern: 8, wantErr: true},
		{name: "empty data", data: []byte{}, ecLevel: ECLevelM, version: 0, maskPattern: 0, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qr, err := EncodeWithMask(tt.data, tt.ecLevel, tt.version, tt.maskPattern)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if qr == nil {
				t.Fatal("expected non-nil QRCode")
			}
			if tt.maskPattern >= 0 {
				if qr.MaskPattern != tt.maskPattern {
					t.Fatalf("expected mask %d, got %d", tt.maskPattern, qr.MaskPattern)
				}
			}
		})
	}
}

func TestBestEncodingMode(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want int
	}{
		{name: "numeric", data: []byte("12345"), want: ModeNumeric},
		{name: "alphanumeric upper", data: []byte("HELLO WORLD 123"), want: ModeAlphanumeric},
		{name: "alphanumeric with space", data: []byte("AB CD"), want: ModeAlphanumeric},
		{name: "byte mode lowercase", data: []byte("hello"), want: ModeByte},
		{name: "byte mode special", data: []byte("test@example.com"), want: ModeByte},
		{name: "empty returns byte", data: []byte{}, want: ModeByte},
		{name: "single digit", data: []byte("5"), want: ModeNumeric},
		{name: "alphanumeric with symbols", data: []byte("A+B*C"), want: ModeAlphanumeric},
		{name: "byte mode with newline", data: []byte("hello\nworld"), want: ModeByte},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BestEncodingMode(tt.data)
			if got != tt.want {
				t.Errorf("BestEncodingMode(%q) = %d, want %d", tt.data, got, tt.want)
			}
		})
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
		{'%', true},
		{'*', true},
		{'+', true},
		{'-', true},
		{'.', true},
		{'/', true},
		{':', true},
		{'a', false},
		{'z', false},
		{'\n', false},
		{0x00, false},
	}

	for _, tt := range tests {
		got := IsAlphanumeric(tt.b)
		if got != tt.want {
			t.Errorf("IsAlphanumeric(%q) = %v, want %v", tt.b, got, tt.want)
		}
	}
}

func TestIsNumeric(t *testing.T) {
	tests := []struct {
		b    byte
		want bool
	}{
		{'0', true},
		{'5', true},
		{'9', true},
		{'A', false},
		{'a', false},
		{' ', false},
	}

	for _, tt := range tests {
		got := IsNumeric(tt.b)
		if got != tt.want {
			t.Errorf("IsNumeric(%q) = %v, want %v", tt.b, got, tt.want)
		}
	}
}

func TestAlphanumericValue(t *testing.T) {
	tests := []struct {
		b       byte
		wantVal int
		wantOk  bool
	}{
		{'0', 0, true},
		{'9', 9, true},
		{'A', 10, true},
		{'Z', 35, true},
		{' ', 36, true},
		{'a', 0, false},
	}

	for _, tt := range tests {
		val, ok := AlphanumericValue(tt.b)
		if ok != tt.wantOk {
			t.Errorf("AlphanumericValue(%q) ok = %v, want %v", tt.b, ok, tt.wantOk)
		}
		if ok && val != tt.wantVal {
			t.Errorf("AlphanumericValue(%q) = %d, want %d", tt.b, val, tt.wantVal)
		}
	}
}

func TestCapacity(t *testing.T) {
	tests := []struct {
		name    string
		version int
		ecLevel int
		mode    int
		wantGt  int
	}{
		{name: "v1 L numeric", version: 1, ecLevel: ECLevelL, mode: ModeNumeric, wantGt: 0},
		{name: "v1 L byte", version: 1, ecLevel: ECLevelL, mode: ModeByte, wantGt: 0},
		{name: "v10 M alphanumeric", version: 10, ecLevel: ECLevelM, mode: ModeAlphanumeric, wantGt: 0},
		{name: "v40 H byte", version: 40, ecLevel: ECLevelH, mode: ModeByte, wantGt: 0},
		{name: "invalid version", version: 0, ecLevel: ECLevelL, mode: ModeByte, wantGt: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Capacity(tt.version, tt.ecLevel, tt.mode)
			if got < tt.wantGt {
				t.Errorf("Capacity() = %d, want >= %d", got, tt.wantGt)
			}
		})
	}
}

func TestDataCodewords(t *testing.T) {
	tests := []struct {
		version  int
		ecLevel  int
		wantZero bool
	}{
		{version: 1, ecLevel: ECLevelL, wantZero: false},
		{version: 1, ecLevel: ECLevelH, wantZero: false},
		{version: 40, ecLevel: ECLevelM, wantZero: false},
		{version: 0, ecLevel: ECLevelL, wantZero: true},
	}

	for _, tt := range tests {
		got := DataCodewords(tt.version, tt.ecLevel)
		if tt.wantZero && got != 0 {
			t.Errorf("DataCodewords(%d, %d) = %d, want 0", tt.version, tt.ecLevel, got)
		}
		if !tt.wantZero && got == 0 {
			t.Errorf("DataCodewords(%d, %d) = 0, want > 0", tt.version, tt.ecLevel)
		}
	}
}

func TestECCodewords(t *testing.T) {
	tests := []struct {
		version  int
		ecLevel  int
		wantZero bool
	}{
		{version: 1, ecLevel: ECLevelL, wantZero: false},
		{version: 1, ecLevel: ECLevelH, wantZero: false},
		{version: 0, ecLevel: ECLevelL, wantZero: true},
	}

	for _, tt := range tests {
		got := ECCodewords(tt.version, tt.ecLevel)
		if tt.wantZero && got != 0 {
			t.Errorf("ECCodewords(%d, %d) = %d, want 0", tt.version, tt.ecLevel, got)
		}
		if !tt.wantZero && got == 0 {
			t.Errorf("ECCodewords(%d, %d) = 0, want > 0", tt.version, tt.ecLevel)
		}
	}
}

func TestBlockCount(t *testing.T) {
	g1, g2 := BlockCount(1, ECLevelL)
	if g1 != 1 || g2 != 0 {
		t.Fatalf("BlockCount(1, L) = (%d, %d), want (1, 0)", g1, g2)
	}

	g1, g2 = BlockCount(5, ECLevelQ)
	if g1 != 2 || g2 != 2 {
		t.Fatalf("BlockCount(5, Q) = (%d, %d), want (2, 2)", g1, g2)
	}

	g1, g2 = BlockCount(0, ECLevelL)
	if g1 != 0 || g2 != 0 {
		t.Error("BlockCount(0, L) should return 0, 0")
	}
}

func TestModeName(t *testing.T) {
	tests := []struct {
		mode int
		want string
	}{
		{ModeNumeric, "Numeric"},
		{ModeAlphanumeric, "Alphanumeric"},
		{ModeByte, "Byte"},
		{ModeKanji, "Kanji"},
		{99, "Unknown"},
	}
	for _, tt := range tests {
		got := modeName(tt.mode)
		if got != tt.want {
			t.Errorf("modeName(%d) = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestECLevelName(t *testing.T) {
	tests := []struct {
		level int
		want  string
	}{
		{ECLevelL, "L"},
		{ECLevelM, "M"},
		{ECLevelQ, "Q"},
		{ECLevelH, "H"},
		{99, "Unknown"},
	}
	for _, tt := range tests {
		got := ecLevelName(tt.level)
		if got != tt.want {
			t.Errorf("ecLevelName(%d) = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestBitsToBytesAndBack(t *testing.T) {
	bits := []bool{true, false, true, true, false, false, false, true}
	b := bitsToBytes(bits)
	if len(b) != 1 {
		t.Fatalf("expected 1 byte, got %d", len(b))
	}
	back := bytesToBits(b)
	if len(back) != 8 {
		t.Fatalf("expected 8 bits, got %d", len(back))
	}
	for i := 0; i < len(bits); i++ {
		if bits[i] != back[i] {
			t.Errorf("bit %d: %v != %v", i, bits[i], back[i])
		}
	}
}
