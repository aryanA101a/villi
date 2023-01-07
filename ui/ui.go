package ui

type FileName string
type Status string
type FileSize string
type ConnectedPeers int
type Peers int

type Progress struct {
	Ratio      float64
	Downloaded uint64
}
type Meta struct {
	FileName       FileName
	Status         Status
	FileSize       FileSize
	ConnectedPeers ConnectedPeers
	Peers          Peers
}

var UpdateUI func(x interface{})
