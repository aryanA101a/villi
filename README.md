<div align="center">
  <img src="https://user-images.githubusercontent.com/23309033/211167294-bfb97561-89aa-4182-8ce8-3e75f6d0ff2a.svg" width="800" height="200" />
 </div>
 
# Villi
![image](https://user-images.githubusercontent.com/23309033/211168707-5657ebe8-2254-4f98-95ac-3f51be5f76b4.png)

A Bit-Torrent client written in GO.  
This project aims to mimic the capabilities of a good torrent client using the original BitTorrent specification as a reference.

## Features
- `.torrent` file support
- **HTTP** and **UDP** Tracker Support
- Terminal User Interface

## Build
`go build`

## Usage
1. **Examples**  
  `./villi file.torrent /downloads/         Download file.torrent and save to /downloads/`  
  `./villi -flag file.torrent /downloads/      Download file.torrent and save to /downloads/ with verbose logging`

2. **Flags**

| __Flag Name__ | __Flag__ | __Description__ | __Default__ |
|-------------|------------|------------|------------|
| Verbose | `-v or --verbose` | Enable verbose logging | false |
| Help | `-h or --help` | Show this help message and exit | false |
