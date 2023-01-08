package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/aryanA101a/villi/torrentfile"
	"github.com/aryanA101a/villi/ui"
	"github.com/aryanA101a/villi/utils"

	// "github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

var p *tea.Program

func main() {
	
	verboseFlag:=flag.Bool("v",false,"Detailed logging")
	flag.BoolVar(verboseFlag,"verbose",false,"logging")

	helpFlag:=flag.Bool("h",false,"help")
	flag.BoolVar(helpFlag,"help",false,"help")
	flag.Usage=func() {
		fmt.Print(usageText)
	}
	flag.Parse()
var inPath,outPath string
	if(*helpFlag){
		flag.Usage()
		return
	}
	 if *verboseFlag {
		inPath = os.Args[2]
		outPath = os.Args[3]
		ui.UpdateUI = func(x interface{}) {}
		start(inPath, outPath)

	} else {
		inPath = os.Args[1]
		outPath = os.Args[2]
		log.SetOutput(ioutil.Discard)

		m := model{
			meta: ui.Meta{
				Status:         "getting info...",
				FileSize:       "",
				ConnectedPeers: 0,
				Peers:          0,
			},
			progress: ui.Progress{
				Ratio:      0,
				Downloaded: 0,
			},
			progressBar: progress.New(progress.WithDefaultGradient(), progress.WithoutPercentage()),
			err:         nil,
		}
		// // Start Bubble Tea
		p = tea.NewProgram(m)

		// Start the download
		go start(inPath, outPath)

		ui.UpdateUI = func(x interface{}) {
			switch x.(type) {
			case ui.Status:
				p.Send(x)
				return
			case ui.FileSize:
				p.Send(x)
				return
			case ui.ConnectedPeers:
				p.Send(x)
				return
			case ui.Peers:
				p.Send(x)
				return
			case ui.Progress:
				p.Send(x)
				return
			case ui.FileName:
				p.Send(x)
				return
			default:
				return

			}

		}
		if _, err := p.Run(); err != nil {
			log.Println("error running program:", err)
			os.Exit(1)
		}
	}

}

func start(inPath string, outPath string) {
	tf, err := torrentfile.Open(inPath, outPath)
	if err != nil {
		log.Fatal(utils.BoldRed(err))
	}
	ui.UpdateUI(ui.FileName(tf.Name))
	ui.UpdateUI(ui.Status("contacting peers..."))
	ui.UpdateUI(ui.FileSize(utils.ConvertToHumanReadable(tf.Length)))

	err = tf.DownloadToFile(outPath)
	if err != nil {
		log.Fatal(utils.BoldRed(err))
	}
}

var usageText=`Usage: villi [options] torrent_file output_directory

Options:
  -v, --verbose    Enable verbose logging
  -h, --help       Show help message and exit

Examples:
  villi file.torrent /downloads/         Download file.torrent and save to /downloads/
  villi file.torrent /downloads/ -v      Download file.torrent and save to /downloads/ with verbose logging
`