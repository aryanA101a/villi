package main

import (
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
	var flag, inPath, outPath string

	switch len(os.Args) {
	case 4:
		flag = os.Args[3]
		inPath = os.Args[1]
		outPath = os.Args[2]
		ui.UpdateUI = func(x interface{}) {}
		if flag == "-v" || flag == "-V" {
		}
		start(inPath, outPath)
	case 3:
		log.SetOutput(ioutil.Discard)
		inPath = os.Args[1]
		outPath = os.Args[2]

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
		// go func() {
		// 	// v:=time.NewTimer(15 * time.Second)
		// 	// <-v.C
		// 	p.Send(progressMsg(0.5555))
		// }()
		// p2p.UpdateProgress = func(progress progressMsg) {
		// 	p.Send(progress)
		// }
		// p2p.UpdateMeta = func(meta metaMsg) {
		// 	p.Send(meta)
		// }
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
	default:
		log.Println("command line args missing")
		os.Exit(1)
	}

}

func start(inPath string, outPath string) {
	tf, err := torrentfile.Open(inPath, outPath)
	if err != nil {
		log.Fatal(err)
	}
	ui.UpdateUI(ui.FileName(tf.Name))
	ui.UpdateUI(ui.Status("contacting peers..."))
	ui.UpdateUI(ui.FileSize(utils.ConvertToHumanReadable(tf.Length)))
	err = tf.DownloadToFile(outPath)
	if err != nil {
		log.Fatal(err)
	}
}
