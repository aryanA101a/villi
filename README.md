# villi
A Bit-Torrent client written in GO.  
This project aims to mimic the capabilities of a good torrent client using the original BitTorrent specification as a reference.

## Features
- `.torrent` file support
- Fetching Peer lists from both **HTTP** and **UDP** Trackers.
- Fetching pieces **concurrently** from Peers.
- Simple command-line interface

## Usage
1. **Downloading**  
`go run main.go file.torrent output_dir -flag`

2. **Flags**

| __Flag Name__ | __Flag__ | __Description__ | __Default__ |
|-------------|------------|------------|------------|
| Verbose | `-v or -V` | Detailed logs | false |
