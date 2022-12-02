package torrentfile

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/aryanA101a/villi/p2p"
	"github.com/aryanA101a/villi/peers"
	bencode "github.com/zeebo/bencode"
	"golang.org/x/exp/maps"
)

// Port used for Bit-Torrent
const Port uint16 = 6881

type TorrentFile struct {
	Announce    []string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength uint
	Length      uint64
	Name        string
	Files       []*file
}

type bencodeInfo struct {
	Pieces      string             `bencode:"pieces"`
	PieceLength uint               `bencode:"piece length"`
	Length      uint64             `bencode:"length"`
	Name        string             `bencode:"name"`
	Files       bencode.RawMessage `bencode:"files"`
}

type bencodeTorrent struct {
	Announce     string             `bencode:"announce"`
	AnnounceList [][]string         `bencode:"announce-list"`
	Info         bencode.RawMessage `bencode:"info"`
}
type bencodeInfoFile struct {
	Path   []string `bencode:"path"`
	Length uint64   `bencode:"length"`
}
type file struct {
	Path        string
	Length      uint64
	FilePointer *os.File
}

func (t *TorrentFile) DownloadToFile(path string) error {
	var peerID [20]byte
	_, err := rand.Read(peerID[:])
	if err != nil {
		return nil
	}
	var peerList []peers.Peer
	peerDict := make(map[string]peers.Peer)

	announceList := t.Announce
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(announceList), func(i, j int) {
		announceList[i], announceList[j] = announceList[j], announceList[i]
	})
	for _, announceURL := range announceList {
		if len(peerDict) >= 30 {
			break
		}
		u, err := url.Parse(announceURL)
		if err != nil {
			continue
		}

		log.Println("Contacting tracker[", announceURL, "] for peer list...")

		var result []peers.Peer
		result, err = t.requestPeers(u, peerID, Port)
		if err != nil {
			log.Println("Failed(", err, "). Trying again...")
			continue
		}

		log.Println("------------------")
		log.Println(result)
		for _, peer := range result {
			if _, ok := peerDict[peer.String()]; !ok {
				peerDict[peer.String()] = peer
			}
		}
		log.Println(peerDict)
		log.Println("------------------")

	}
	peerList = append(peerList, maps.Values(peerDict)...)
	log.Printf("Got %d peers", len(peerList))
	log.Print("PeerList:")
	log.Println(peerList)
	if peerList == nil {
		// panic("Unable to receive peers! Problem with the torrent or internet")

		return err

	}
	// := t.requestPeers(peerID, Port)
	// if err != nil {
	// return err
	// }

	torrent := p2p.Torrent{
		Peers:       peerList,
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

	offset := uint64(0)
	for _, f := range t.Files {
		_, err = f.FilePointer.Write(buf[offset:f.Length])
		if err != nil {
			return err
		}
		offset += f.Length
		err = f.FilePointer.Close()
		if err != nil {
			return err
		}
	}

	return nil

}

func Open(inPath string,outPath string) (TorrentFile, error) {
	file, err := os.Open(inPath)
	if err != nil {

		return TorrentFile{}, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return TorrentFile{}, err
	}

	bto := bencodeTorrent{}
	err = bencode.DecodeBytes(data, &bto)
	if err != nil {
		return TorrentFile{}, err
	}

	return bto.toTorrentFile(outPath)
}

// func (i *bencodeInfo) hash() ([20]byte, error) {
// 	var buf bytes.Buffer
// 	err := bencode.Marshal(&buf, *i)
// 	if err != nil {
// 		return [20]byte{}, err
// 	}
// 	// log.Println(*i)
// 	h := sha1.Sum(buf.Bytes())
// 	return h, nil
// }

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

func (bto *bencodeTorrent) toTorrentFile(outPath string) (TorrentFile, error) {

	bencodeInfo := bencodeInfo{}
	err := bencode.DecodeBytes(bto.Info, &bencodeInfo)
	log.Println(bencodeInfo)
	if err != nil {
		return TorrentFile{}, err
	}

	//sha1 hash of info dict in .torrent file
	infoHash := sha1.Sum(bto.Info)

	//a slice containing sha1 hash of each piece
	pieceHashes, err := bencodeInfo.splitPiecesHashes()
	if err != nil {
		return TorrentFile{}, err
	}
	var length uint64
	files := make([]*file, 0)
	if err != nil {
		return TorrentFile{}, err
	}
	
	if bencodeInfo.Length > 0 {
		var filePointer *os.File
		name := path.Join(outPath, bencodeInfo.Name)
		filePointer, err = os.Create(name)
		if err != nil {
		log.Println(outPath)

			return TorrentFile{}, err
		}
		log.Println("erssrur")

		files = append(files, &file{
			Path:        name,
			Length:      bencodeInfo.Length,
			FilePointer: filePointer,
		})
		length = bencodeInfo.Length

	} else {
		err = os.Mkdir(path.Join(outPath, bencodeInfo.Name), os.ModePerm)
			if err != nil && os.IsNotExist(err) {
				return TorrentFile{}, err
			}

		bencodeInfoFiles := make([]*bencodeInfoFile, 0)
		err = bencode.DecodeBytes(bencodeInfo.Files, &bencodeInfoFiles)
		if err != nil {
			return TorrentFile{}, err
		}

		for _, f := range bencodeInfoFiles {
			var filePointer *os.File
			name := path.Join(outPath, bencodeInfo.Name+"/"+f.Path[0])
			filePointer, err = os.Create(name)
			if err != nil {
				return TorrentFile{}, err
			}

			files = append(files, &file{
				Path:        name,
				Length:      f.Length,
				FilePointer: filePointer,
			})
			length += f.Length
		}
	}

	//parse tracker urls
	var announceList []string
	if len(bto.AnnounceList) > 0 {
		for _, tier := range bto.AnnounceList {
			announceList = append(announceList, tier...)
		}
	} else {
		announceList = append(announceList, bto.Announce)
	}

	t := TorrentFile{
		Announce:    announceList,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bencodeInfo.PieceLength,
		Length:      length,
		Name:        bencodeInfo.Name,
		Files:       files,
	}
	return t, nil
}
