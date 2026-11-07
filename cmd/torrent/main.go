package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Nuriklan/torrent-client/internal/pwp"
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

	fmt.Println("\nAttempting to connect to peers...")
	for _, p := range resp.Peers {
		fmt.Printf("Connecting to %s...", p)
		client, err := pwp.NewClient(p, torrentFile.InfoHash, peerID, 3*time.Second)
		if err != nil {
			fmt.Printf("failed: %v\n", err)
			continue
		}
		defer client.Close()

		fmt.Printf("CONNECTED! \n")
		fmt.Printf("  Remote Peer ID: %q\n", string(client.PeerID[:]))
		fmt.Printf("  Has Bitfield: %v (len: %d bytes)\n", client.Bitfield != nil, len(client.Bitfield))

		err = client.SendInterested()
		if err != nil {
			log.Printf("failed to send interested: %v", err)
			continue
		}

		pieceIndex := 0
		pieceLen := torrentFile.PieceLength
		expectedHash := torrentFile.PieceHashes[pieceIndex]

		fmt.Printf("\n🚀 Starting download of Piece %d (%d bytes)...\n", pieceIndex, pieceLen)
		data, err := client.DownloadPiece(pieceIndex, pieceLen, expectedHash)
		if err != nil {
			fmt.Printf("❌ Failed to download piece: %v\n", err)
			continue
		}

		fmt.Printf("🎉 SUCCESS! Piece %d downloaded and verified! Size: %d bytes\n", pieceIndex, len(data))
		break
	}
}
