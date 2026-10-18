package pwp

import (
	"bytes"
	"testing"
)

func TestHandshake_RoundTrip(t *testing.T) {
	infoHash := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	peerId := [20]byte{20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	hs := NewHandshake(infoHash, peerId)

	data := hs.Serialize()
	if len(data) != 68 {
		t.Fatalf("expected 68 bytes, got %d", len(data))
	}

	got, err := ReadHandshake(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("unexpected error reading handshake: %v", err)
	}

	if got.Pstr != hs.Pstr {
		t.Errorf("got Pstr %q, want %q", got.Pstr, hs.Pstr)
	}
	if got.InfoHash != hs.InfoHash {
		t.Errorf("got InfoHash %v, want %v", got.InfoHash, hs.InfoHash)
	}
	if got.PeerID != hs.PeerID {
		t.Errorf("got PeerID %v, want %v", got.PeerID, hs.PeerID)
	}
}

func TestReadHandshake_Errors(t *testing.T) {
	validHandshake := NewHandshake([20]byte{}, [20]byte{}).Serialize()

	badPstrlen := make([]byte, len(validHandshake))
	copy(badPstrlen, validHandshake)
	badPstrlen[0] = 20

	badProtocol := make([]byte, len(validHandshake))
	copy(badProtocol, validHandshake)
	badProtocol[1] = 'X'

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
	}{
		{
			name:    "input too short (unexpected EOF)",
			input:   []byte{19, 'B', 'i', 't'},
			wantErr: true,
		},
		{
			name:    "empty input",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "invalid pstrlen",
			input:   badPstrlen,
			wantErr: true,
		},
		{
			name:    "invalid protocol string",
			input:   badProtocol,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			_, err := ReadHandshake(reader)

			if (err != nil) != tt.wantErr {
				t.Errorf("ReadHandshake() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
