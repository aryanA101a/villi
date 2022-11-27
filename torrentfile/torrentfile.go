package torrentfile

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/aryanA101a/villi/p2p"
	"github.com/aryanA101a/villi/peers"
	"github.com/jackpal/bencode-go"
)

//Port used for Bit-Torrent
const Port uint16 = 6881

type TorrentFile struct {
	Announce    []string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce     string      `bencode:"announce"`
	AnnounceList [][]string  `bencode:"announce-list"`
	Info         bencodeInfo `bencode:"info"`
}

func (t *TorrentFile) DownloadToFile(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return nil
	}

	var peers []peers.Peer
	for _, announceURL := range t.Announce {
		u, err := url.Parse(announceURL)
		if err != nil {
			continue
		}
		log.Println("Contacting tracker[", announceURL, "] for peer list...")

		peers, err = t.requestPeers(u, peerID, Port)
		if err == nil {
			log.Println("break")
			break
		}
		log.Println("Failed(", err, "). Trying again...")

	}

	if peers == nil {
		// panic("Unable to receive peers! Problem with the torrent or internet")
		
		return err

	}
	// := t.requestPeers(peerID, Port)
	// if err != nil {
		// return err
	// }

	torrent := p2p.Torrent{
		Peers:       peers,
		PeerID:      peerID,
		InfoHash:    t.InfoHash,
		PieceHashes: t.PieceHashes,
		PieceLength: t.PieceLength,
		Length:      t.Length,
		Name:        t.Name,
	}
	buf, err := torrent.Download()
	if err != nil {
		return err
	}
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()
	_, err = outFile.Write(buf)
	if err != nil {
		return err
	}
	return nil

}

func Open(path string) (TorrentFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return TorrentFile{}, err
	}
	defer file.Close()

	bto := bencodeTorrent{}
	err = bencode.Unmarshal(file, &bto)
	if err != nil {
		return TorrentFile{}, err
	}
	return bto.toTorrentFile()
}

func (i *bencodeInfo) hash() ([20]byte, error) {
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, *i)
	if err != nil {
		return [20]byte{}, err
	}
	h := sha1.Sum(buf.Bytes())
	return h, nil
}

func (i *bencodeInfo) splitPiecesHashes() ([][20]byte, error) {
	hashLen := 20 //Length of SHA1 hash
	buf := []byte(i.Pieces)
	if len(buf)%hashLen != 0 {
		err := fmt.Errorf("recieved malformed pieces of length %d", len(buf))
		return nil, err
	}
	numHashes := len(buf) / hashLen
	hashes := make([][20]byte, numHashes)

	for i := 0; i < numHashes; i++ {
		copy(hashes[i][:], buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes, nil
}

func (bto *bencodeTorrent) toTorrentFile() (TorrentFile, error) {
	//sha1 hash of info dict in .torrent file
	infoHash, err := bto.Info.hash()
	if err != nil {
		return TorrentFile{}, err
	}
	//a slice containing sha1 hash of each piece
	pieceHashes, err := bto.Info.splitPiecesHashes()
	if err != nil {
		return TorrentFile{}, err
	}

	//parse tracker urls
	var announceList []string
	if len(bto.AnnounceList) > 0 {
		for _, tier := range bto.AnnounceList {
			for _, announce := range tier {
				announceList = append(announceList, announce)
			}
		}
	} else {
		announceList = append(announceList, bto.Announce)
	}

	t := TorrentFile{
		Announce:    announceList,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:      bto.Info.Length,
		Name:        bto.Info.Name,
	}
	return t, nil
}
