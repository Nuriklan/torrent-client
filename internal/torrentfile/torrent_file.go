package torrentfile

import (
	"fmt"
	"os"

	"github.com/Nuriklan/torrent-client/internal/bencode"
)

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func Open(path string) (*TorrentFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read the file: %w", err)
	}

	bencodedData, err := bencode.Unmarshal(data)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshalling: %w", err)
	}

	dict, ok := bencodedData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid torrent file: expected dictionary at root")
	}

	announce, ok := dict["announce"].(string)
	if !ok {
		return nil, fmt.Errorf("torrent file missing 'announce' field")
	}

	infoHash, ok := dict["_info_hash"].([20]byte)
	if !ok {
		return nil, fmt.Errorf("torrent file missing or invalid '_info_hash'")
	}

	infoDict, ok := dict["info"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("torrent file missing 'info' dictionary")
	}

	pieceLength, ok := infoDict["piece length"].(int)
	if !ok {
		return nil, fmt.Errorf("info map missing 'piece length' value")
	}

	length, ok := infoDict["length"].(int)
	if !ok {
		return nil, fmt.Errorf("info map missing 'length' value")
	}

	piecesStr, ok := infoDict["pieces"].(string)
	if !ok {
		return nil, fmt.Errorf("info map missing 'pieces' value")
	}

	piecesHashes, err := splitPieceHashes([]byte(piecesStr))
	if err != nil {
		return nil, fmt.Errorf("failed to split piece hashes: %w", err)
	}

	name, ok := infoDict["name"].(string)
	if !ok {
		return nil, fmt.Errorf("info map missing 'name' value")
	}

	return &TorrentFile{
		Announce:    announce,
		InfoHash:    infoHash,
		PieceHashes: piecesHashes,
		PieceLength: pieceLength,
		Length:      length,
		Name:        name,
	}, nil
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
