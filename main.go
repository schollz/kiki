package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/darkowlzz/openurl"
	"github.com/schollz/kiki/src/logging"
)

func main() {
	flag.StringVar(&PublicPort, "public", "8003", "port for the data (this) server")
	flag.StringVar(&PrivatePort, "private", "8004", "port for the data (this) server")
	flag.BoolVar(&NoSync, "no-sync", false, "disable syncing")
	debug := flag.Bool("debug", false, "turn on debug mode")
	noBrowser := flag.Bool("no-browser", false, "do not open browser")
	flag.StringVar(&Location, "path", ".", "path to the kiki database folder")
	flag.Parse()
	if *debug {
		logging.Log.Debug(true)
	} else {
		logging.Log.Debug(false)
	}

	if !*noBrowser {
		go openurl.Open("http://localhost:" + PrivatePort)
	}

	os.Mkdir(Location, 0755)

	err := Run()
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
}
