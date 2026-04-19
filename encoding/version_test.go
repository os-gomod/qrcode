package encoding

import "testing"

func TestGetVersionInfo(t *testing.T) {
	tests := []struct {
		name    string
		version int
		ecLevel int
		wantErr bool
		errSub  string
	}{
		{name: "v1 L", version: 1, ecLevel: ECLevelL},
		{name: "v1 H", version: 1, ecLevel: ECLevelH},
		{name: "v10 M", version: 10, ecLevel: ECLevelM},
		{name: "v40 L", version: 40, ecLevel: ECLevelL},
		{name: "invalid version 0", version: 0, ecLevel: ECLevelL, wantErr: true, errSub: "invalid version"},
		{name: "invalid version 41", version: 41, ecLevel: ECLevelL, wantErr: true, errSub: "invalid version"},
		{name: "invalid EC -1", version: 1, ecLevel: -1, wantErr: true, errSub: "invalid EC level"},
		{name: "invalid EC 4", version: 1, ecLevel: 4, wantErr: true, errSub: "invalid EC level"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vi, err := GetVersionInfo(tt.version, tt.ecLevel)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if vi.Version != tt.version {
				t.Errorf("Version = %d, want %d", vi.Version, tt.version)
			}
			if vi.TotalCodewords <= 0 {
				t.Error("TotalCodewords should be > 0")
			}
			if vi.DataCodewords <= 0 {
				t.Error("DataCodewords should be > 0")
			}
			if vi.ECCodewordsPerBlock <= 0 {
				t.Error("ECCodewordsPerBlock should be > 0")
			}
			if vi.ECBlocks <= 0 {
				t.Error("ECBlocks should be > 0")
			}
			if vi.NumGroups != 1 && vi.NumGroups != 2 {
				t.Errorf("NumGroups should be 1 or 2, got %d", vi.NumGroups)
			}
		})
	}
}

func TestDataCapacity(t *testing.T) {
	tests := []struct {
		version  int
		ecLevel  int
		wantZero bool
	}{
		{1, ECLevelL, false},
		{1, ECLevelH, false},
		{10, ECLevelM, false},
		{40, ECLevelL, false},
		{0, ECLevelL, true},
		{41, ECLevelL, true},
	}

	for _, tt := range tests {
		got := DataCapacity(tt.version, tt.ecLevel)
		if tt.wantZero {
			if got != 0 {
				t.Errorf("DataCapacity(%d, %d) = %d, want 0", tt.version, tt.ecLevel, got)
			}
		} else {
			if got <= 0 {
				t.Errorf("DataCapacity(%d, %d) = %d, want > 0", tt.version, tt.ecLevel, got)
			}
		}
	}
}

func TestMinVersionForData(t *testing.T) {
	tests := []struct {
		name       string
		dataLen    int
		ecLevel    int
		wantErr    bool
		wantGtZero bool
	}{
		{name: "small data", dataLen: 10, ecLevel: ECLevelL},
		{name: "medium data", dataLen: 100, ecLevel: ECLevelM},
		{name: "exact capacity", dataLen: 19, ecLevel: ECLevelL},
		{name: "negative length", dataLen: -1, ecLevel: ECLevelL, wantErr: true},
		{name: "invalid EC", dataLen: 10, ecLevel: 5, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := MinVersionForData(tt.dataLen, tt.ecLevel)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v < 1 || v > 40 {
				t.Errorf("version = %d, want 1-40", v)
			}
		})
	}
}

func TestVersionInfoConsistency(t *testing.T) {
	// Verify data codewords = group1*dcw1 + group2*dcw2
	for v := 1; v <= 40; v++ {
		for lvl := 0; lvl < 4; lvl++ {
			vi, err := GetVersionInfo(v, lvl)
			if err != nil {
				continue
			}
			expected := vi.Group1Blocks*vi.Group1DataCodewords + vi.Group2Blocks*vi.Group2DataCodewords
			if vi.DataCodewords != expected {
				t.Errorf("v=%d lvl=%d: DataCodewords=%d, expected %d (g1=%d*%d + g2=%d*%d)",
					v, lvl, vi.DataCodewords, expected,
					vi.Group1Blocks, vi.Group1DataCodewords,
					vi.Group2Blocks, vi.Group2DataCodewords)
			}
			total := vi.DataCodewords + vi.ECBlocks*vi.ECCodewordsPerBlock
			if total != vi.TotalCodewords {
				t.Errorf("v=%d lvl=%d: total=%d != TotalCodewords=%d", v, lvl, total, vi.TotalCodewords)
			}
		}
	}
}
