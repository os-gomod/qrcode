package encoding

import "testing"

func TestGetVersionInfo(t *testing.T) {
	vi, err := GetVersionInfo(1, ECLevelL)
	if err != nil {
		t.Fatalf("GetVersionInfo(1, L) error: %v", err)
	}
	if vi.Version != 1 {
		t.Errorf("expected version 1, got %d", vi.Version)
	}
	if vi.TotalCodewords != 26 {
		t.Errorf("expected 26 total codewords, got %d", vi.TotalCodewords)
	}
	if vi.DataCodewords <= 0 {
		t.Errorf("DataCodewords should be > 0, got %d", vi.DataCodewords)
	}
}

func TestGetVersionInfo_Invalid(t *testing.T) {
	_, err := GetVersionInfo(0, ECLevelL)
	if err == nil {
		t.Error("expected error for version 0")
	}
	_, err = GetVersionInfo(41, ECLevelL)
	if err == nil {
		t.Error("expected error for version 41")
	}
	_, err = GetVersionInfo(1, 5)
	if err == nil {
		t.Error("expected error for EC level 5")
	}
	_, err = GetVersionInfo(1, -1)
	if err == nil {
		t.Error("expected error for negative EC level")
	}
}

func TestDataCapacity(t *testing.T) {
	cap1L := DataCapacity(1, ECLevelL)
	if cap1L <= 0 {
		t.Errorf("DataCapacity(1, L) = %d, want > 0", cap1L)
	}
	// Higher version should have more capacity.
	cap5L := DataCapacity(5, ECLevelL)
	if cap5L <= cap1L {
		t.Errorf("DataCapacity(5, L) = %d should be > DataCapacity(1, L) = %d", cap5L, cap1L)
	}
	// Higher EC level should have less capacity.
	cap1H := DataCapacity(1, ECLevelH)
	if cap1H >= cap1L {
		t.Errorf("DataCapacity(1, H) = %d should be < DataCapacity(1, L) = %d", cap1H, cap1L)
	}
}

func TestMinVersionForData(t *testing.T) {
	t.Run("small data", func(t *testing.T) {
		v, err := MinVersionForData(10, ECLevelL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v < 1 || v > 40 {
			t.Errorf("version %d out of range", v)
		}
	})

	t.Run("invalid ec level", func(t *testing.T) {
		_, err := MinVersionForData(10, 5)
		if err == nil {
			t.Error("expected error for invalid EC level")
		}
	})

	t.Run("negative data length", func(t *testing.T) {
		_, err := MinVersionForData(-1, ECLevelL)
		if err == nil {
			t.Error("expected error for negative data length")
		}
	})

	t.Run("too large data", func(t *testing.T) {
		// Version 40, EC H has the smallest capacity. Request something larger.
		maxCap := DataCapacity(40, ECLevelH)
		_, err := MinVersionForData(maxCap+1, ECLevelH)
		if err == nil {
			t.Error("expected error for data exceeding maximum capacity")
		}
	})

	t.Run("version ordering", func(t *testing.T) {
		v1, _ := MinVersionForData(100, ECLevelL)
		v2, _ := MinVersionForData(200, ECLevelL)
		if v2 < v1 {
			t.Errorf("larger data should need >= version: v1=%d, v2=%d", v1, v2)
		}
	})
}

func TestAllVersions_Valid(t *testing.T) {
	for v := 1; v <= 40; v++ {
		for ec := 0; ec <= 3; ec++ {
			vi, err := GetVersionInfo(v, ec)
			if err != nil {
				t.Errorf("GetVersionInfo(%d, %d) error: %v", v, ec, err)
				continue
			}
			if vi.DataCodewords <= 0 {
				t.Errorf("version %d ec %d: DataCodewords = %d", v, ec, vi.DataCodewords)
			}
			if vi.ECBlocks <= 0 {
				t.Errorf("version %d ec %d: ECBlocks = %d", v, ec, vi.ECBlocks)
			}
			if vi.ECCodewordsPerBlock <= 0 {
				t.Errorf("version %d ec %d: ECCodewordsPerBlock = %d", v, ec, vi.ECCodewordsPerBlock)
			}
			if vi.NumGroups < 1 || vi.NumGroups > 2 {
				t.Errorf("version %d ec %d: NumGroups = %d", v, ec, vi.NumGroups)
			}
		}
	}
}
