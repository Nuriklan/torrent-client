package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"

	"github.com/Nuriklan/torrent-client/internal/torrentfile"
	"github.com/Nuriklan/torrent-client/internal/tracker"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: torrent <path-to-torrent-file>")
	}

	file := os.Args[1]
	torrentFile, err := torrentfile.Open(file)
	if err != nil {
		log.Fatalf("failed to open torrent file %v: %v", file, err)
	}

	var peerID [20]byte
	copy(peerID[:], "-NT0001-")
	_, err = rand.Read(peerID[8:])
	if err != nil {
		log.Fatalf("failed to  generate peer id: %v", err)
	}

	req := tracker.AnnounceRequest{
		InfoHash:   torrentFile.InfoHash,
		PeerID:     peerID,
		Port:       6881,
		Uploaded:   0,
		Downloaded: 0,
		Left:       torrentFile.Length,
		Compact:    true,
		Event:      "started",
	}

	resp, err := tracker.Announce(torrentFile.Announce, req)
	if err != nil {
		log.Fatalf("failed to announce to tracker: %v", err)
	}

	fmt.Printf("\n Success! Tracker interval: %d seconds\n", resp.Interval)
	fmt.Printf("Discovered %d peers:\n", len(resp.Peers))
	for i, peer := range resp.Peers {
		fmt.Printf(" [%02d] %s\n", i+1, peer)
	}
}
