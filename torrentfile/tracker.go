package torrentfile

import (
	"net/url"
	"strconv"
)

func (t *TorrentFile) buildTrackerURL(peerID [20]byte,port uint16)(string,error){
	base,err:=url.Parse(t.Announce)
	if err !=nil{
		return "",err
	}
	params:=url.Values{
		"info_hash": []string{string(t.InfoHash[:])},
		"peer_id":[]string{string(peerID[:])},
		"port":[]string{strconv.Itoa(int(Port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery=params.Encode()
	return base.String(), nil
}