package tracker

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
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
