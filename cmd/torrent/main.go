package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Nuriklan/torrent-client/internal/torrentfile"
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

	fmt.Printf("Decoded: %+v\n", torrentFile)
}
