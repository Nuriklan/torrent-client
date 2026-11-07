package pwp

import (
	"crypto/sha1"
	"fmt"
	"net"
	"time"

	"github.com/Nuriklan/torrent-client/internal/tracker"
)

const (
	maxBlockSize = 16384
	maxBacklog   = 5
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield Bitfield
	Peer     tracker.Peer
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewClient(peer tracker.Peer, infoHash, peerId [20]byte, timeout time.Duration) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.String(), timeout)
	if err != nil {
		return nil, fmt.Errorf("could not connect to peer: %v", err)
	}

	hs := NewHandshake(infoHash, peerId)
	conn.Write(hs.Serialize())

	resHandshake, err := ReadHandshake(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read handshake: %w", err)
	}

	if resHandshake.InfoHash != infoHash {
		conn.Close()
		return nil, fmt.Errorf("infoHash is not correct")
	}

	msg, err := ReadMessage(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("could not read message: %v", err)
	}

	var bf Bitfield
	if msg != nil && msg.ID == MsgBitfield {
		bf = Bitfield(msg.Payload)
	}

	return &Client{
		Conn:     conn,
		Choked:   true,
		Bitfield: bf,
		Peer:     peer,
		InfoHash: infoHash,
		PeerID:   resHandshake.PeerID,
	}, nil
}

func (c *Client) Close() error {
	return c.Conn.Close()
}

func (c *Client) Read() (*Message, error) {
	return ReadMessage(c.Conn)
}

func (c *Client) SendInterested() error {
	msg := &Message{ID: MsgInterested}
	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// SendRequest formats and sends a Request message (ID 6) to the peer,
// asking for a block of length bytes at offset begin within piece index.
func (c *Client) SendRequest(index, begin, length int) error {
	msg := FormatRequest(index, begin, length)

	_, err := c.Conn.Write(msg.Serialize())
	return err
}

// Download piece
func (c *Client) DownloadPiece(index, pieceLength int, expectedHash [20]byte) ([]byte, error) {
	buf := make([]byte, pieceLength)
	requested := 0
	downloaded := 0
	backlog := 0

	for downloaded < pieceLength {
		for !c.Choked && backlog < maxBacklog && requested < pieceLength {
			// filling requests queue
			blockSize := maxBlockSize
			if pieceLength-requested < blockSize {
				blockSize = pieceLength - requested
			}

			err := c.SendRequest(index, requested, blockSize)
			if err != nil {
				return nil, fmt.Errorf("failed to send request: %w", err)
			}

			backlog++
			requested += blockSize
		}

		// reading response from peer
		msg, err := c.Read()
		if err != nil {
			return nil, fmt.Errorf("failed to read message: %w", err)
		}
		if msg == nil {
			continue
		}

		// process message type
		switch msg.ID {
		case MsgChoke:
			c.Choked = true
		case MsgUnchoke:
			c.Choked = false
		case MsgPiece:
			n, err := ParsePiece(index, buf, msg)
			if err != nil {
				return nil, fmt.Errorf("failed to parse piece: %w", err)
			}
			downloaded += n
			backlog--
		}
	}

	hash := sha1.Sum(buf)
	if hash != expectedHash {
		return nil, fmt.Errorf("piece %d hash mismatch: expected %x, got %x", index, expectedHash, hash)
	}

	return buf, nil
}
