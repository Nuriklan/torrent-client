package torrentfile

import "fmt"

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
}

func splitPieceHashes(pieces []byte) ([][20]byte, error) {
	if len(pieces)%20 != 0 {
		return nil, fmt.Errorf("malformed pieces: length %d is not a multiple of 20", len(pieces))
	}

	numHashes := len(pieces) / 20
	hashes := make([][20]byte, 0, numHashes)

	for i := 0; i < len(pieces); i += 20 {
		chunk := pieces[i : i+20]
		hashes = append(hashes, [20]byte(chunk))
	}

	return hashes, nil
}
