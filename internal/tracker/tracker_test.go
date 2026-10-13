package tracker

import (
	"net/url"
	"strings"
	"testing"	
)

func TestBuildURL(t *testing.T) {
	infoHash := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	peerID := [20]byte{'N', 'T', '-', '0', '0', '0', '1', '-'}

	tests := []struct {
		name        string
		baseURL     string
		req         AnnounceRequest
		wantErr     bool
		errContains string
		checkFunc   func(t *testing.T, u *url.URL)
	}{
		{
			name:    "port zero return error",
			baseURL: "http://example.com/announce",
			req: AnnounceRequest{
				Port: 0,
			},
			wantErr:     true,
			errContains: "invalid port",
		},
		{
			name:    "invalid base url return error",
			baseURL: "://invalid-url",
			req: AnnounceRequest{
				Port: 6881,
			},
			wantErr:     true,
			errContains: "could not parse base url",
		},
		{
			name:    "successful build with all parameters",
			baseURL: "http://example.com/announce",
			req: AnnounceRequest{
				InfoHash:   infoHash,
				PeerID:     peerID,
				Port:       6881,
				Uploaded:   0,
				Downloaded: 0,
				Left:       1024,
				Compact:    true,
				Event:      "started",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, u *url.URL) {
				// check binary hash in RaqQuery
				expectedHash := "%01%02%03%04%05%06%07%08%09%0A%0B%0C%0D%0E%0F%10%11%12%13%14"
				if !strings.Contains(u.RawQuery, "info_hash="+expectedHash) {
					t.Errorf("RaqQuery missing expected info_hash; got %s", u.RawQuery)
				}

				// check query parameters
				q := u.Query()
				if got := q.Get("port"); got != "6881" {
					t.Errorf("exptected port 6881, got %s", got)
				}
				if got := q.Get("compact"); got != "1" {
					t.Errorf("expected compact 1, got %s", got)
				}
				if got := q.Get("event"); got != "started" {
					t.Errorf("expected event started, got %s", got)
				}
			},
		},
		{
			name:    "preserves existing query parameters in baseURL",
			baseURL: "http://example.com/announce?token=secret123",
			req: AnnounceRequest{
				Port:     6881,
				Compact:  true,
				InfoHash: infoHash,
				PeerID:   peerID,
			},
			wantErr: false,
			checkFunc: func(t *testing.T, u *url.URL) {
				q := u.Query()
				if got := q.Get("token"); got != "secret123" {
					t.Errorf("expected token secret123, got %s", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildURL(tt.baseURL, tt.req)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, got)
			}
		})
	} 
}
