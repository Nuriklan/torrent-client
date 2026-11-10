package download

import (
	"log"
	"time"

	"github.com/Nuriklan/torrent-client/internal/pwp"
	"github.com/Nuriklan/torrent-client/internal/tracker"
)

type PieceWork struct {
	Index  int
	Hash   [20]byte
	Length int
}

type PieceResult struct {
	Index int
	Data  []byte
}

type Torrent struct {
	Peers       []tracker.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func (t *Torrent) pieceSize(index int) int {
	begin := index * t.PieceLength
	end := begin + t.PieceLength
	end = min(end, t.Length)
	return end - begin
}

func (t *Torrent) startWorker(peer tracker.Peer, workQueue chan *PieceWork, results chan *PieceResult) {
	client, err := pwp.NewClient(peer, t.InfoHash, t.PeerID, 5*time.Second)
	if err != nil {
		// peer is unavailable
		return
	}
	defer client.Close()

	err = client.SendInterested()
	if err != nil {
		return
	}

	for piece := range workQueue {
		if client.Bitfield != nil && !client.Bitfield.HasPiece(piece.Index) {
			workQueue <- piece
			time.Sleep(100 * time.Millisecond)
			continue
		}

		data, err := client.DownloadPiece(piece.Index, piece.Length, piece.Hash)
		if err != nil {
			log.Printf("Worker %s failed piece %d: %v", peer, piece.Index, err)
			workQueue <- piece
			return
		}

		results <- &PieceResult{Index: piece.Index, Data: data}
	}

}

func (t *Torrent) Download() ([]byte, error) {
	workQueue := make(chan *PieceWork, len(t.PieceHashes))
	results := make(chan *PieceResult)

	for index, hash := range t.PieceHashes {
		length := t.pieceSize(index)
		workQueue <- &PieceWork{
			Index:  index,
			Hash:   hash,
			Length: length,
		}
	}

	for _, peer := range t.Peers {
		go t.startWorker(peer, workQueue, results)
	}

	buf := make([]byte, t.Length)
	donePieces := 0

	for donePieces < len(t.PieceHashes) {
		res := <-results

		begin := res.Index * t.PieceLength
		end := begin + len(res.Data)
		copy(buf[begin:end], res.Data)

		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d", percent, res.Index, len(t.PieceHashes))
	}

	close(workQueue)
	return buf, nil
}
