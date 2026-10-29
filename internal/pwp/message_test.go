package pwp

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestMessage_KeepAlive(t *testing.T) {
	var m *Message = nil
	serialized := m.Serialize()

	expected := []byte{0, 0, 0, 0}
	if !bytes.Equal(expected, serialized) {
		t.Fatalf("Serialize() = %v, want %v", serialized, expected)
	}

	got, err := ReadMessage(bytes.NewReader(serialized))
	if err != nil {
		t.Fatalf("unexpected error reading keep-alive: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil message for keep-alive, got %+v", got)
	}
}

func TestMessage_Payload(t *testing.T) {
	tests := []struct {
		name       string
		input      *Message
		wantLength int
		wantErr    bool
	}{
		{
			name:       "message without payload",
			input:      &Message{ID: MsgUnchoke},
			wantLength: 5,
			wantErr:    false,
		},
		{
			name:       "message with payload",
			input:      &Message{ID: MsgUnchoke, Payload: []byte{1, 2, 3, 4}},
			wantLength: 9,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serialized := tt.input.Serialize()

			if len(serialized) != tt.wantLength {
				t.Fatalf("expected length %d, got %d", len(serialized), tt.wantLength)
			}

			got, err := ReadMessage(bytes.NewReader(serialized))
			if err != nil {
				t.Fatalf("unexpected error when reading message %v", err)
			}

			if got == nil {
				t.Fatalf("expected message, got nil")
			}

			if got.ID != tt.input.ID {
				t.Errorf("got ID %v, want %v", got.ID, tt.input.ID)
			}

			if !bytes.Equal(got.Payload, tt.input.Payload) {
				t.Errorf("got Payload %v, want %v", got.Payload, tt.input.Payload)
			}
		})
	}
}

func TestMessage_Errors(t *testing.T) {
	tooLargeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(tooLargeBuf, MaxMessageSize+1)

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "EOF while reading length",
			input:   []byte{0, 0},
			wantErr: true,
		},
		{
			name:    "EOF while reading payload",
			input:   []byte{0, 0, 0, 10, byte(MsgChoke), 1},
			wantErr: true,
		},
		{
			name:    "MaxMessageSize exceeding",
			input:   tooLargeBuf,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadMessage(bytes.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatRequest(t *testing.T) {
	tests := []struct {
		name   string
		index  int
		begin  int
		length int
	}{
		{
			name:   "first block of first piece",
			index:  0,
			begin:  0,
			length: 16384,
		},
		{
			name:   "middle block",
			index:  4,
			begin:  32768,
			length: 16384,
		},
		{
			name:   "tail block with non-standard length",
			index:  10,
			begin:  49152,
			length: 5120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := FormatRequest(tt.index, tt.begin, tt.length)

			if msg.ID != MsgRequest {
				t.Errorf("expected ID MsgRequest, got %v", msg.ID)
			}
			if len(msg.Payload) != 12 {
				t.Errorf("expected payload length 12, got %d", len(msg.Payload))
			}

			index := binary.BigEndian.Uint32(msg.Payload[0:4])
			if index != uint32(tt.index) {
				t.Errorf("expected index %d, got %d", tt.index, index)
			}

			begin := binary.BigEndian.Uint32(msg.Payload[4:8])
			if begin != uint32(tt.begin) {
				t.Errorf("expected begin offset %d, got %d", tt.begin, begin)
			}

			length := binary.BigEndian.Uint32(msg.Payload[8:12])
			if length != uint32(tt.length) {
				t.Errorf("expected length %d, got %d", tt.length, length)
			}
		})
	}
}

func TestParsePiece(t *testing.T) {
	makePieceMsg := func(index, begin uint32, block []byte) *Message {
		payload := make([]byte, 8+len(block))
		binary.BigEndian.PutUint32(payload[0:4], index)
		binary.BigEndian.PutUint32(payload[4:8], begin)
		copy(payload[8:], block)
		return &Message{ID: MsgPiece, Payload: payload}
	}

	tests := []struct {
		name          string
		expectedIndex int
		bufSize       int
		msg           *Message
		wantN         int
		wantErr       bool
		checkBuf      func(t *testing.T, buf []byte)
	}{
		{
			name:          "happy path: valid block at offset 0",
			expectedIndex: 0,
			bufSize:       10,
			msg:           makePieceMsg(0, 0, []byte{1, 2, 3}),
			wantN:         3,
			wantErr:       false,
			checkBuf: func(t *testing.T, buf []byte) {
				expected := []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0}
				if !bytes.Equal(buf, expected) {
					t.Errorf("buf mismatch: got %v, want %v", buf, expected)
				}
			},
		},
		{
			name:          "happy path: valid block with offset",
			expectedIndex: 1,
			bufSize:       10,
			msg:           makePieceMsg(1, 4, []byte{9, 9}),
			wantN:         2,
			wantErr:       false,
			checkBuf: func(t *testing.T, buf []byte) {
				expected := []byte{0, 0, 0, 0, 9, 9, 0, 0, 0, 0}
				if !bytes.Equal(buf, expected) {
					t.Errorf("buf mismatch: got %v, want %v", buf, expected)
				}
			},
		},
		{
			name:          "error message is nil",
			expectedIndex: 0,
			bufSize:       10,
			msg:           nil,
			wantErr:       true,
		},
		{
			name:          "error: wrong message ID",
			expectedIndex: 0,
			bufSize:       10,
			msg:           &Message{ID: MsgChoke},
			wantErr:       true,
		},
		{
			name:          "error: payload too short (< 8 bytes)",
			expectedIndex: 0,
			bufSize:       10,
			msg:           &Message{ID: MsgPiece, Payload: []byte{0, 0, 0, 1}},
			wantErr:       true,
		},
		{
			name:          "error: piece index mismatch",
			expectedIndex: 0,
			bufSize:       10,
			msg:           makePieceMsg(1, 0, []byte{1, 2}),
			wantErr:       true,
		},
		{
			name:          "error: begin offset is out of bounds",
			expectedIndex: 0,
			bufSize:       10,
			msg:           makePieceMsg(0, 10, []byte{1}),
			wantErr:       true,
		},
		{
			name:          "error: block data exceeds buffer length",
			expectedIndex: 0,
			bufSize:       10,
			msg:           makePieceMsg(0, 8, []byte{1, 2, 3}),
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, tt.bufSize)
			n, err := ParsePiece(tt.expectedIndex, buf, tt.msg)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParsePiece() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if n != tt.wantN {
				t.Errorf("ParsePiece() written bytes = %d, want %d", n, tt.wantN)
			}

			if tt.checkBuf != nil {
				tt.checkBuf(t, buf)
			}
		})
	}
}
