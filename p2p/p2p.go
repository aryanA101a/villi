package p2p

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aryanA101a/villi/client"
	"github.com/aryanA101a/villi/message"
	"github.com/aryanA101a/villi/peers"
	"github.com/aryanA101a/villi/ui"
)

const MaxBlockSize = 16384

const MaxBacklog = 5

type Torrent struct {
	Peers          []peers.Peer
	PeerID         [20]byte
	InfoHash       [20]byte
	PieceHashes    [][20]byte
	PieceLength    uint
	Length         uint64
	Name           string
	ConnectedPeers int
}

type pieceWork struct {
	index  int
	hash   [20]byte
	length int
}

type pieceResult struct {
	index int
	buf   []byte
}

type pieceProgress struct {
	index      int
	client     *client.Client
	buf        []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read() //blocking call
	if err != nil {
		return err
	}
	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false
	case message.MsgChoke:
		state.client.Choked = true
	case message.MsgHave:
		index, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(index)
	case message.MsgPiece:
		n, err := message.ParsePiece(state.index, state.buf, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func attemptDownloadPiece(c *client.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		index:  pw.index,
		client: c,
		buf:    make([]byte, pw.length),
	}

	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.length {
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.length {
				blockSize := MaxBlockSize
				if pw.length-state.requested < blockSize {
					blockSize = pw.length - state.requested
				}

				err := c.SendRequest(pw.index, state.requested, blockSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += blockSize

			}
		}
		err := state.readMessage()
		if err != nil {
			return nil, err
		}

	}
	return state.buf, nil
}

func checkIntegrity(pw *pieceWork, buf []byte) error {
	hash := sha1.Sum(buf)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("Index %d failed integrity check", pw.index)
	}
	return nil
}

func (t *Torrent) startDownloadWorker(connectedPeersLock *sync.Mutex, peer peers.Peer, workQuene chan *pieceWork, results chan *pieceResult) {

	var c *client.Client

	c, err := client.New(peer, t.PeerID, t.InfoHash)
	if err != nil {
		log.Println(err.Error())
		log.Printf("Could not handshake with %s. Disconnecting\n\n\n", peer.IP)

		return

	}

	connectedPeersLock.Lock()
	t.ConnectedPeers++
	connectedPeersLock.Unlock()

	defer c.Conn.Close()
	defer func() {

		connectedPeersLock.Lock()
		if t.ConnectedPeers != 0 {
			t.ConnectedPeers--
		}
		connectedPeersLock.Unlock()
	}()

	log.Printf("Completed handshake with %s\n", peer.IP)

	c.SendUnchoke()
	c.SendInterested()

	for pw := range workQuene {
		if !c.Bitfield.HasPiece(pw.index) {
			workQuene <- pw
			return
		}

		buf, err := attemptDownloadPiece(c, pw)
		if err != nil {
			log.Println("Existing", err)
			workQuene <- pw
			return
		}

		err = checkIntegrity(pw, buf)
		if err != nil {
			log.Printf("Piece #%d failed integrity check\n", pw.index)
			workQuene <- pw
			continue
		}

		c.SendHave(pw.index)
		results <- &pieceResult{pw.index, buf}
	}

}

func (t *Torrent) calculateBoundsForPiece(index uint) (begin uint64, end uint64) {
	begin = uint64(index * t.PieceLength)
	end = begin + uint64(t.PieceLength)
	if end > t.Length {
		end = t.Length
	}
	return begin, end
}

func (t *Torrent) calculatePieceSize(index uint) int {
	begin, end := t.calculateBoundsForPiece(index)
	return int(end - begin)
}

func (t *Torrent) Download() ([]byte, error) {
	log.Println("Starting download for", t.Name)
	// timeout := make(chan bool, 1)
	workQuene := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceResult)
	for index, hash := range t.PieceHashes {
		length := t.calculatePieceSize(uint(index))
		workQuene <- &pieceWork{index, hash, length}
	}
	var connectedPeersLock sync.Mutex
	log.Println(t.Peers)
	for _, peer := range t.Peers {
		go t.startDownloadWorker(&connectedPeersLock, peer, workQuene, results)
	}

	buf := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		log.Println("bres")

		// var res *pieceResult
		// select {
		// case r := <-results:
		// 	res = r
		// case <-timeout:
		// 	if (runtime.NumGoroutine() - 1) == 0 {
		// 		log.Println("timeout")
		// 		err := fmt.Errorf("cannot download")
		// 		return nil, err
		// 	}
		// 	continue
		// }
		res := <-results
		log.Println("ares")

		begin, end := t.calculateBoundsForPiece(uint(res.index))
		copy(buf[begin:end], res.buf)
		donePieces++

		ratio := float64(donePieces) / float64(len(t.PieceHashes))
		downloaded := uint64(donePieces) * uint64(t.PieceLength)
		if donePieces == len(t.PieceHashes) {
			downloaded = (downloaded - uint64(t.PieceLength)) + uint64(t.calculatePieceSize(uint(len(t.PieceHashes)-1)))
		}
		ui.UpdateUI(ui.Progress{
			Ratio:      ratio,
			Downloaded: downloaded,
		})
		ui.UpdateUI(ui.ConnectedPeers(t.ConnectedPeers))

		log.Printf("(%0.2f%%) Downloaded piece %d from %d peers\n", ratio*100, res.index, t.ConnectedPeers)
	}
	close(workQuene)
	return buf, nil
}
