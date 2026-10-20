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
