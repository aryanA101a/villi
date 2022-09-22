package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

//Port used for Bit-Torrent
const Port uint16=6881

type TorrentFile struct{
	Announce string
	InfoHash [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length int
	Name string
}

type bencodeInfo struct{
	Pieces string `bencode:"pieces"`
	PieceLength int `bencode:"piece length"`
	Length int `bencode:"length"`
	Name string `bencode:"name"`
}

type bencodeTorrent struct{
	Announce string `bencode:"announce"`
	Info bencodeInfo `bencode:"info"`
}



func Open(path string) (TorrentFile,error){
	file,err :=os.Open(path)
	if err!=nil{
		return TorrentFile{}, err
	}
	defer file.Close()

	bto :=bencodeTorrent{}
	err=bencode.Unmarshal(file,&bto)
	if err != nil {
		return TorrentFile{},err
	}
	return bto.toTorrentFile()
}

func(i *bencodeInfo) hash()([20]byte,error){
	var buf bytes.Buffer
	err:= bencode.Marshal(&buf,*i)
	if err!=nil{
		return [20]byte{},err
	}
	h:=sha1.Sum(buf.Bytes())
	return h,nil
}

func (i *bencodeInfo) splitPiecesHashes()([][20]byte,error){
	hashLen:=20 //Length of SHA1 hash
	buf:=[]byte(i.Pieces)
	if len(buf)%hashLen !=0{
		err:=fmt.Errorf("recieved malformed pieces of length %d",len(buf))
		return nil,err
	}
	numHashes:=len(buf)/hashLen
	hashes:=make([][20]byte,numHashes)

	for i:=0;i<numHashes;i++{
		copy(hashes[i][:],buf[i*hashLen:(i+1)*hashLen])
	}
	return hashes,nil
}

func (bto *bencodeTorrent) toTorrentFile()(TorrentFile,error){
	//sha1 hash of info dict in .torrent file
	infoHash,err:=bto.Info.hash()
	if err!=nil{
		return TorrentFile{},err
	}
	//a slice containing sha1 hash of each piece
	pieceHashes,err:=bto.Info.splitPiecesHashes()
	if err!=nil{
		return TorrentFile{}, err
	}
	t:= TorrentFile{
		Announce: bto.Announce,
		InfoHash: infoHash,
		PieceHashes: pieceHashes,
		PieceLength: bto.Info.PieceLength,
		Length:bto.Info.Length,
		Name:bto.Info.Name,
	}
	return t,nil
}