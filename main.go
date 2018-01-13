package main

import (
	"flag"
	"fmt"

	"github.com/darkowlzz/openurl"
	"github.com/schollz/kiki/src/logging"
)

const (
	DEFAULT_PUBLIC_PORT  = "8003"
	DEFAULT_PRIVATE_PORT = "8004"
)

var (
	NoSync bool
)

func main() {
	flag.StringVar(&PublicPort, "public", DEFAULT_PUBLIC_PORT, "port for the data (this) server")
	flag.StringVar(&PrivatePort, "private", DEFAULT_PRIVATE_PORT, "port for the data (this) server")
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
	err := Run()
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
}
