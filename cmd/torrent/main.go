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
		}

		fmt.Println("Reading next messages from peer:")
		for i := 0; i < 5; i++ {
			msg, err := client.Read()
			if err != nil {
				fmt.Printf("Read error: %v\n", err)
				break
			}
			if msg == nil {
				fmt.Println("  -> Keep-Alive")
				continue
			}
			fmt.Printf("  -> Messsage ID: %d (Payload bytes: %d)\n", msg.ID, len(msg.Payload))
			if msg.ID == pwp.MsgBitfield {
				bf := pwp.Bitfield(msg.Payload)
				fmt.Printf("    Found Bitfield! Length: %d bytes\n", len(bf))
				fmt.Printf("    Peer has piece 0: %v\n", bf.HasPiece(0))
			}
			if msg.ID == pwp.MsgUnchoke {
				fmt.Println("    Peer UNCHOKED us! We can request blocks now!")
			}
		}

		break
	}
}
