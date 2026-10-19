package pwp

type Bitfield []byte

func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	if byteIndex >= len(bf) || index < 0 {
		return false
	}
	mask := byte(1 << (7 - index%8))
	return bf[byteIndex]&mask != 0
}

func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	if byteIndex >= len(bf) || index < 0 {
		return
	}
	mask := byte(1 << (7 - index%8))
	bf[byteIndex] |= mask
}
