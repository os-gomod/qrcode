package encoding

import "testing"

func TestEncodeEC(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		ecCount int
		wantNil bool
	}{
		{name: "simple", data: []byte{0x10, 0x20, 0x30}, ecCount: 4},
		{name: "single byte", data: []byte{0x12}, ecCount: 2},
		{name: "larger ec", data: []byte{0x01, 0x02, 0x03, 0x04}, ecCount: 10},
		{name: "zero ec count", data: []byte{0x01}, ecCount: 0, wantNil: true},
		{name: "negative ec count", data: []byte{0x01}, ecCount: -1, wantNil: true},
		{name: "empty data", data: []byte{}, ecCount: 4, wantNil: true},
		{name: "ec count 1", data: []byte{0x40}, ecCount: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeEC(tt.data, tt.ecCount)
			if tt.wantNil {
				if got != nil {
					t.Error("expected nil result")
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil result")
			}
			if len(got) != tt.ecCount {
				t.Errorf("expected %d EC bytes, got %d", tt.ecCount, len(got))
			}
		})
	}
}

func TestEncodeECBlocks(t *testing.T) {
	tests := []struct {
		name       string
		dataBlocks [][]byte
		ecPerBlock int
	}{
		{
			name:       "single block",
			dataBlocks: [][]byte{{0x10, 0x20, 0x30}},
			ecPerBlock: 4,
		},
		{
			name:       "two blocks",
			dataBlocks: [][]byte{{0x01, 0x02}, {0x03, 0x04}},
			ecPerBlock: 2,
		},
		{
			name:       "empty blocks list",
			dataBlocks: [][]byte{},
			ecPerBlock: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EncodeECBlocks(tt.dataBlocks, tt.ecPerBlock)
			if len(got) != len(tt.dataBlocks) {
				t.Fatalf("expected %d blocks, got %d", len(tt.dataBlocks), len(got))
			}
			for i, block := range got {
				if block != nil && len(block) != tt.ecPerBlock {
					t.Errorf("block %d: expected %d bytes, got %d", i, tt.ecPerBlock, len(block))
				}
			}
		})
	}
}
