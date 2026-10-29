package pwp

import (
	"encoding/binary"
	"fmt"
	"io"
)

type MessageID byte

const (
	MsgChoke         MessageID = 0
	MsgUnchoke       MessageID = 1
	MsgInterested    MessageID = 2
	MsgNotInterested MessageID = 3
	MsgHave          MessageID = 4
	MsgBitfield      MessageID = 5
	MsgRequest       MessageID = 6
	MsgPiece         MessageID = 7
	MsgCancel        MessageID = 8
)

type Message struct {
	ID      MessageID
	Payload []byte
}

const MaxMessageSize = 256 * 1024

func (m *Message) Serialize() []byte {
	// Keep-Alive message
	if m == nil {
		return make([]byte, 4)
	}

	length := uint32(len(m.Payload) + 1)
	buf := make([]byte, 4+length)
	binary.BigEndian.PutUint32(buf[0:4], length)
	buf[4] = byte(m.ID)
	copy(buf[5:], m.Payload)

	return buf
}

func ReadMessage(r io.Reader) (*Message, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf)
	if length > MaxMessageSize {
		return nil, fmt.Errorf("message too large: %d bytes", length)
	}

	// Keep-Alive message
	if length == 0 {
		return nil, nil
	}

	messageBuf := make([]byte, length)
	_, err = io.ReadFull(r, messageBuf)
	if err != nil {
		return nil, err
	}

	return &Message{
		ID:      MessageID(messageBuf[0]),
		Payload: messageBuf[1:],
	}, nil
}

// FormatRequest creates a BitTorrent request message for a specific piece block.
func FormatRequest(index, begin, length int) *Message {
	payload := make([]byte, 12)

	binary.BigEndian.PutUint32(payload[0:4], uint32(index))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))

	return &Message{ID: MsgRequest, Payload: payload}
}

// ParsePiece validates the payload of a Piece message and copies
// the recieved block data into the destination buffer buf.
// It returns the number of written bytes or an error if the block is malformed.
func ParsePiece(index int, buf []byte, msg *Message) (int, error) {
	if msg == nil {
		return 0, fmt.Errorf("message is nil")
	}
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("expected MsgPiece (%d), got %d", MsgPiece, msg.ID)
	}

	payload := msg.Payload
	payloadLength := len(payload)
	if payloadLength < 8 {
		return 0, fmt.Errorf("invalid payload length: %d", payloadLength)
	}

	parsedIndex := binary.BigEndian.Uint32(payload[0:4])
	if int(parsedIndex) != index {
		return 0, fmt.Errorf("expected index %d, got %d", index, parsedIndex)
	}

	begin := int(binary.BigEndian.Uint32(payload[4:8]))
	if begin >= len(buf) {
		return 0, fmt.Errorf("'begin' offset (%d) is out of buffer bounds (%d)", begin, len(buf))
	}

	block := payload[8:]
	if begin+len(block) > len(buf) {
		return 0, fmt.Errorf("block data [%d:%d] exceeds buffer length (%d)", begin, begin+len(block), len(buf))
	}

	copy(buf[begin:], block)
	return len(block), nil
}
