package pwp

import "testing"

func TestBitfield_HasPiece(t *testing.T) {
	bf := Bitfield{0b10000001, 0b00000010}

	tests := []struct {
		name  string
		index int
		want  bool
	}{
		{name: "piece 0 is set (first bit of byte 0)", index: 0, want: true},
		{name: "piece 1 is not set", index: 1, want: false},
		{name: "piece 7 is set (last bit of byte 0)", index: 7, want: true},
		{name: "piece 8 is not set (first bit of byte 1)", index: 8, want: false},
		{name: "piece 14 is set", index: 14, want: true},
		{name: "piece 15 is not set", index: 15, want: false},
		// boundaries
		{name: "negative index", index: -1, want: false},
		{name: "out of bound index", index: 16, want: false},
		{name: "far out of bounds index", index: 100, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bf.HasPiece(tt.index)
			if got != tt.want {
				t.Errorf("HasPiece(%d) = %v, want %v", tt.index, got, tt.want)
			}
		})
	}
}

func TestBitfield_SetPiece(t *testing.T) {
	bf := Bitfield{0b00000000, 0b00000000}

	bf.SetPiece(-1)
	bf.SetPiece(100)

	if bf.HasPiece(0) {
		t.Errorf("expected false, got true")
	}

	bf.SetPiece(0)
	if !bf.HasPiece(0) {
		t.Errorf("expected true, got false")
	}

	bf.SetPiece(7)
	if !bf.HasPiece(7) {
		t.Errorf("expected piece 7 to be set")
	}
	if !bf.HasPiece(0) {
		t.Errorf("expected piece 0 to remain set")
	}
	if bf[0] != 0b10000001 {
		t.Errorf("bf[0]	= %08b, want 0b10000001", bf[0])
	}
}
