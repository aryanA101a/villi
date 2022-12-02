package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/aryanA101a/villi/torrentfile"
)

func main() {
	var flag, inPath, outPath string
	
	switch len(os.Args) {
	case 4:
		flag = os.Args[3]
		if flag == "-v" || flag == "-V" {
			log.SetOutput(ioutil.Discard)
		}
		fallthrough
	case 3:
		inPath = os.Args[1]
		outPath = os.Args[2]
	default:
		log.Println("command line args missing")
		os.Exit(1)
	}

	

	tf, err := torrentfile.Open(inPath, outPath)
	if err != nil {
		log.Fatal(err)
	}

	err = tf.DownloadToFile(outPath)
	if err != nil {
		log.Fatal(err)
	}
}
