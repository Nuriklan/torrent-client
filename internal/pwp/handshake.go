package pwp

import (
	"fmt"
	"io"
)

const ProtocolString = "BitTorrent protocol"

type Handshake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(infoHash, peerID [20]byte) *Handshake {
	return &Handshake{
		Pstr:     ProtocolString,
		InfoHash: infoHash,
		PeerID:   peerID,
	}
}

func (h *Handshake) Serialize() []byte {
	buf := make([]byte, 68)

	buf[0] = byte(len(h.Pstr))
	curr := 1
	curr += copy(buf[curr:], []byte(h.Pstr))
	curr += 8
	curr += copy(buf[curr:], h.InfoHash[:])
	copy(buf[curr:], h.PeerID[:])

	return buf
}

func ReadHandshake(r io.Reader) (*Handshake, error) {
	buf := make([]byte, 68)

	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake: %v", err)
	}

	// check protocol length (equals to 19) and the protocol string
	pstrlen := int(buf[0])
	if pstrlen != 19 {
		return nil, fmt.Errorf("unexpected pstrlen: got %d, want 19", pstrlen)
	}

	pstr := string(buf[1 : 1+pstrlen])
	if pstr != ProtocolString {
		return nil, fmt.Errorf("unexpected protocol string: got %q, want %q", pstr, ProtocolString)
	}

	infoHash := buf[28:48]
	peerID := buf[48:68]

	return &Handshake{
		Pstr:     pstr,
		InfoHash: [20]byte(infoHash),
		PeerID:   [20]byte(peerID),
	}, nil
}
