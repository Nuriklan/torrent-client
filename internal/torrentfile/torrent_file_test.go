package torrentfile

import (
	"testing"
)

func TestOpen(t *testing.T) {
	t.Run("valid torrent file", func(t *testing.T) {
		tf, err := Open("../../testdata/sample.torrent")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if tf.Announce == "" {
			t.Errorf("expected announce URL, got empty string")
		}

		if len(tf.PieceHashes) == 0 {
			t.Errorf("expected at least one piece hash")
		}

		totalCapacity := len(tf.PieceHashes) * tf.PieceLength
		if totalCapacity < tf.Length {
			t.Errorf("pieces capacity (%d) must be >= file length (%d)", totalCapacity, tf.Length)
		}
	})
}

func TestSplitPieceHashes(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		wantLen int
	}{
		{
			name:    "invalid length not multiple of 20",
			input:   []byte("1234567890123456789"),
			wantErr: true,
		},
		{
			name:    "valid single hash",
			input:   make([]byte, 20),
			wantErr: false,
			wantLen: 1,
		},
		{
			name:    "valid two hashes",
			input:   make([]byte, 40),
			wantErr: false,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashes, err := splitPieceHashes(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}
			if len(hashes) != tt.wantLen {
				t.Errorf("expected %d hashes, got %d", tt.wantLen, len(hashes))
			}
		})
	}
}
