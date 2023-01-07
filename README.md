# villi
A Bit-Torrent client written in GO.  
This project aims to mimic the capabilities of a good torrent client using the original BitTorrent specification as a reference.

## Features
- `.torrent` file support
- **HTTP** and **UDP** Tracker Support
- Terminal User Interface

## Usage
1. **Examples**
  `villi file.torrent /downloads/         Download file.torrent and save to /downloads/`
  `villi file.torrent /downloads/ -v      Download file.torrent and save to /downloads/ with verbose logging`

2. **Flags**

| __Flag Name__ | __Flag__ | __Description__ | __Default__ |
|-------------|------------|------------|------------|
| Verbose | `-v or --verbose` | Enable verbose logging | false |
| Help | `-h or --help` | Show this help message and exit | false |
