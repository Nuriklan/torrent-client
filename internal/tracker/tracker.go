package tracker

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Nuriklan/torrent-client/internal/bencode"
)

type Peer struct {
	IP   net.IP
	Port uint16
}

func (p Peer) String() string {
	return net.JoinHostPort(p.IP.String(), strconv.Itoa(int(p.Port)))
}

type AnnounceRequest struct {
	InfoHash   [20]byte
	PeerID     [20]byte
	Port       uint16
	Uploaded   int
	Downloaded int
	Left       int
	Compact    bool
	Event      string
}

type AnnounceResponse struct {
	Interval int
	Peers    []Peer
}

func Announce(baseURL string, req AnnounceRequest) (*AnnounceResponse, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	u, err := buildURL(baseURL, req)
	if err != nil {
		return nil, fmt.Errorf("could not build URL: %v", err)
	}

	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("tracker request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tracker returned bad status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	bencodedData, err := bencode.Unmarshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %w", err)
	}

	dict, ok := bencodedData.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected dictionary in tracker response, got %T", bencodedData)
	}

	if reason, ok := dict["failure reason"].(string); ok {
		return nil, fmt.Errorf("tracker rejected request: %s", reason)
	}

	interval, ok := dict["interval"].(int)
	if !ok {
		return nil, fmt.Errorf("tracker response missing 'interval' field")
	}

	peerStr, ok := dict["peers"].(string)
	if !ok {
		return nil, fmt.Errorf("tracker response missing 'peers' field")
	}

	peers, err := unmarshalPeers([]byte(peerStr))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal peers: %w", err)
	}

	return &AnnounceResponse{
		Interval: interval,
		Peers:    peers,
	}, nil
}

func escapeBinaryPercent(b []byte) string {
	res := make([]byte, 0, len(b)*3)
	for _, c := range b {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '_' || c == '.' || c == '~' {
			res = append(res, c)
		} else {
			res = fmt.Appendf(res, "%%%02X", c)
		}
	}
	return string(res)
}

func buildURL(baseURL string, req AnnounceRequest) (*url.URL, error) {
	if req.Port == 0 {
		return nil, fmt.Errorf("invalid port: %d", req.Port)
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse base url: %w", err)
	}

	q := make(url.Values)
	q.Set("port", strconv.FormatUint(uint64(req.Port), 10))
	q.Set("uploaded", strconv.FormatUint(uint64(req.Uploaded), 10))
	q.Set("downloaded", strconv.FormatUint(uint64(req.Downloaded), 10))
	q.Set("left", strconv.FormatUint(uint64(req.Left), 10))

	compactStr := "0"
	if req.Compact {
		compactStr = "1"
	}
	q.Set("compact", compactStr)

	if req.Event != "" {
		q.Set("event", req.Event)
	}

	binaryParams := fmt.Sprintf("info_hash=%s&peer_id=%s",
		escapeBinaryPercent(req.InfoHash[:]),
		escapeBinaryPercent(req.PeerID[:]),
	)

	newQuery := binaryParams + "&" + q.Encode()
	if u.RawQuery != "" {
		u.RawQuery = u.RawQuery + "&" + newQuery
	} else {
		u.RawQuery = newQuery
	}

	return u, nil
}

func unmarshalPeers(peersBin []byte) ([]Peer, error) {
	if len(peersBin)%6 != 0 {
		return nil, fmt.Errorf("malformed peers: length %d is not a multiple of 6", len(peersBin))
	}

	numPeers := len(peersBin) / 6
	peers := make([]Peer, 0, numPeers)

	for i := 0; i < len(peersBin); i += 6 {
		ip := net.IP(peersBin[i : i+4])
		port := binary.BigEndian.Uint16(peersBin[i+4 : i+6])
		peers = append(peers, Peer{IP: ip, Port: port})
	}

	return peers, nil
}
