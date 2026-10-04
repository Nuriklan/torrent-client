package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Nuriklan/torrent-client/internal/bencode"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: torrent <path-to-torrent-file>")
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("failed to read file %v", err)
	}

	res, err := bencode.Unmarshal(data)
	if err != nil {
		log.Fatalf("decode error: %v", err)
	}

	fmt.Printf("Decoded: %+v\n", res)
}
