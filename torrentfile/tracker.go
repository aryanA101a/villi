package torrentfile

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/aryanA101a/villi/peers"
	"github.com/jackpal/bencode-go"
)

type bencodeTrackerResp struct {
	Interval int    `bencode:"interval"`
	Peers    string `bencode:"peers"`
}

func (t *TorrentFile) buildTrackerURL(announceURL string, peerID [20]byte, port uint16) (string, error) {
	base, err := url.Parse(announceURL)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{string(t.InfoHash[:])},
		"peer_id":    []string{string(peerID[:])},
		"port":       []string{strconv.Itoa(int(Port))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.FormatUint(t.Length,10)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

func (t *TorrentFile) requestPeers(announceURL *url.URL, peerID [20]byte, port uint16) ([]peers.Peer, error) {
	var peers []peers.Peer
	var err error

	switch announceURL.Scheme {
	case "http":
		peers, err = t.requestPeersHTTP(announceURL, peerID, port)
	case "udp":
		peers, err = t.requestPeersUDP(announceURL, peerID, port)
	default:
		err = fmt.Errorf("announce url not recognized")
	}
	return peers, err
}

func (t *TorrentFile) requestPeersHTTP(announceURL *url.URL, peerID [20]byte, port uint16) ([]peers.Peer, error) {
	url, err := t.buildTrackerURL(announceURL.String(), peerID, port)
	if err != nil {
		return nil, err
	}

	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	trackerResp := bencodeTrackerResp{}
	err = bencode.Unmarshal(resp.Body, &trackerResp)
	if err != nil {
		return nil, err
	}

	return peers.Unmarshal([]byte(trackerResp.Peers))
}

func (t *TorrentFile) requestPeersUDP(announceURL *url.URL, peerID [20]byte, port uint16) ([]peers.Peer, error) {
	
	conn, err := net.Dial("udp", announceURL.Host)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var connID uint64
	err = conn.SetDeadline(time.Now().Add(4 * time.Second))
	if err != nil {
		return nil, err
	}
	connID, err = connectReqUDP(conn)
	if err != nil {
		return nil, err
	}

	

	var peers []peers.Peer
	peers, err = announceReqUDP(conn, connID, peerID, port, *t)
	
	if err != nil {
		return nil, err
	}


	return peers, nil

}

func connectReqUDP(conn net.Conn) (uint64, error) {

	/*
		connect request:
			Offset  Size            Name            Value
			0       64-bit integer  protocol_id     0x41727101980 // magic constant
			8       32-bit integer  action          0 // connect
			12      32-bit integer  transaction_id
			16

		connect response:
			Offset  Size            Name            Value
			0       32-bit integer  action          0 // connect
			4       32-bit integer  transaction_id
			8       64-bit integer  connection_id
			16

	*/

	connectPacket, err := buildConnectPacket()
	if err != nil {
		return 0, err
	}

	transactionID := binary.BigEndian.Uint32(connectPacket[12:])

	_, err = conn.Write(connectPacket)
	if err != nil {
		return 0, err
	}

	respBytes := make([]byte, 16)
	var respLen int
	respLen, err = conn.Read(respBytes)
	if err != nil {
		return 0, err
	}

	if respLen != 16 {
		err = fmt.Errorf("unexpected response size %d", respLen)
		return 0, err
	}

	if binary.BigEndian.Uint32(respBytes[0:4]) != 0 {
		err = fmt.Errorf("unexpected connect response action")
		return 0, err
	}

	if transactionID != binary.BigEndian.Uint32(respBytes[4:8]) {
		err = fmt.Errorf("TransactionID does not match")
		return 0, err
	}

	connectID := binary.BigEndian.Uint64(respBytes[8:])
	return connectID, nil
}

func buildConnectPacket() ([]byte, error) {
	packet := make([]byte, 16)
	transactionID := make([]byte, 4)
	_, err := rand.Read(transactionID)
	if err != nil {
		return nil, err
	}

	binary.BigEndian.PutUint64(packet[:8], uint64(0x41727101980))
	binary.BigEndian.PutUint32(packet[8:12], uint32(0))
	binary.BigEndian.PutUint32(packet[12:], binary.BigEndian.Uint32(transactionID[:]))

	return packet, err
}

func announceReqUDP(conn net.Conn, connectID uint64, peerID [20]byte, port uint16, t TorrentFile) ([]peers.Peer, error) {
	/*
		IPv4 announce request:
			Offset  Size    Name    Value
			0       64-bit integer  connection_id
			8       32-bit integer  action          1 // announce
			12      32-bit integer  transaction_id
			16      20-byte string  info_hash
			36      20-byte string  peer_id
			56      64-bit integer  downloaded
			64      64-bit integer  left
			72      64-bit integer  uploaded
			80      32-bit integer  event           0 // 0: none; 1: completed; 2: started; 3: stopped
			84      32-bit integer  IP address      0 // default
			88      32-bit integer  key
			92      32-bit integer  num_want        -1 // default
			96      16-bit integer  port
			98

		IPv4 announce response:
			Offset      Size            Name            Value
			0           32-bit integer  action          1 // announce
			4           32-bit integer  transaction_id
			8           32-bit integer  interval
			12          32-bit integer  leechers
			16          32-bit integer  seeders
			20 + 6 * n  32-bit integer  IP address
			24 + 6 * n  16-bit integer  TCP port
			20 + 6 * N
	*/

	announcePacket, err := buildAnnouncePacket(connectID, peerID, port, t)
	if err != nil {
		return nil, err
	}
	transactionID := announcePacket[12:16]
	_, err = conn.Write(announcePacket)
	if err != nil {
		return nil, err
	}

	respBuffer := new(bytes.Buffer)
	var respLen int
	respBytes := make([]byte, 4096)
	respLen, err = conn.Read(respBytes)
	if err != nil {
		return nil, err
	}

	err = binary.Write(respBuffer, binary.BigEndian, respBytes[:respLen])
	if err != nil {
		return nil, err
	}

	if respLen <= 20 {
		err = fmt.Errorf("unexpected response size")
		return nil, err
	}

	if binary.BigEndian.Uint32(transactionID[:]) != binary.BigEndian.Uint32(respBuffer.Bytes()[4:8]) {
		err = fmt.Errorf("transaction id not matching")
		return nil, err
	}

	action := binary.BigEndian.Uint32(respBuffer.Bytes()[0:4])
	if action != 1 {
		if action == 3 {
			err = fmt.Errorf("%d:unexpected announce response action | message:%s", action, string(respBuffer.Bytes()[8:]))
		} else {
			err = fmt.Errorf("%d:unexpected announce response action", action)

		}
		return nil, err
	}

	peerList, err := peers.Unmarshal(respBuffer.Bytes()[20:])
	if err != nil {
		return nil, err
	}

	return peerList, nil
}

func buildAnnouncePacket(connID uint64, peerID [20]byte, port uint16, t TorrentFile) ([]byte, error) {
	announcePacket := new(bytes.Buffer)

	transactionID := make([]byte, 4)
	_, err := rand.Read(transactionID[:])
	if err != nil {
		return nil, err
	}

	//connection id
	err = binary.Write(announcePacket, binary.BigEndian, connID)
	if err != nil {
		return nil, err
	}

	//action
	err = binary.Write(announcePacket, binary.BigEndian, uint32(1))
	if err != nil {
		return nil, err
	}

	//transaction id
	
	err = binary.Write(announcePacket, binary.BigEndian, transactionID)
	if err != nil {
		return nil, err
	}

	//infohash
	err = binary.Write(announcePacket, binary.BigEndian, t.InfoHash)
	if err != nil {
		return nil, err
	}

	//peer id
	err = binary.Write(announcePacket, binary.BigEndian, get_peer_id())
	if err != nil {
		return nil, err
	}

	//downloaded
	err = binary.Write(announcePacket, binary.BigEndian, uint64(0))
	if err != nil {
		return nil, err
	}

	//left
	err = binary.Write(announcePacket, binary.BigEndian, uint64(0))
	if err != nil {
		return nil, err
	}

	//uploaded
	err = binary.Write(announcePacket, binary.BigEndian, uint64(0))
	if err != nil {
		return nil, err
	}

	//event
	err = binary.Write(announcePacket, binary.BigEndian, uint32(0))
	if err != nil {
		return nil, err
	}

	//ip address
	err = binary.Write(announcePacket, binary.BigEndian, uint32(0))
	if err != nil {
		return nil, err
	}

	//key
	err = binary.Write(announcePacket, binary.BigEndian, uint32(0))
	if err != nil {
		return nil, err
	}

	//num want
	err = binary.Write(announcePacket, binary.BigEndian, int32(-1))
	if err != nil {
		return nil, err
	}

	//port
	err = binary.Write(announcePacket, binary.BigEndian, uint16(8000))
	if err != nil {
		return nil, err
	}

	return announcePacket.Bytes(), nil
}

func get_peer_id() [20]byte {
	buf:=new(bytes.Buffer)
	binary.Write(buf,binary.LittleEndian,time.Now().Unix())
	return sha1.Sum(buf.Bytes())

}
