package main

import (
	"flag"

	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/server"
)

func main() {
	port := flag.String("port", "8003", "port for the data (this) server")
	debug := flag.Bool("debug", false, "turn on debug mode")
	flag.Parse()
	if *debug {
		logging.Log.Debug(true)
	}
	server.Port = *port
	server.Run()
}
