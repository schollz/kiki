package main

import (
	"flag"
	"fmt"

	"github.com/schollz/kiki/src/logging"
)

const (
	DEFAULT_PUBLIC_PORT  = "8003"
	DEFAULT_PRIVATE_PORT = "8004"
)

func main() {
	flag.StringVar(&PublicPort, "public", DEFAULT_PUBLIC_PORT, "port for the data (this) server")
	flag.StringVar(&PrivatePort, "private", DEFAULT_PRIVATE_PORT, "port for the data (this) server")
	debug := flag.Bool("debug", false, "turn on debug mode")
	flag.StringVar(&Location, "path", ".", "path to the kiki database folder")
	flag.Parse()
	if *debug {
		logging.Log.Debug(true)
	} else {
		logging.Log.Debug(false)
	}

	err := Run()
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
}
