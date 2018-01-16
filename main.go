package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/darkowlzz/openurl"
	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/logging"
)

var (
	// Port defines what port the carrier should listen on
	PublicPort     = "8003"
	PrivatePort    = "8004"
	RegionPublic   = "rbcDfDMIe8qXq4QPtIUtuEylDvlGynx56QgeHUZUZBk="
	RegionPrivate  = "GQf6ZbBbnVGhiHZ_IqRv0AlfqQh1iofmSyFOcp1ti8Q="
	NoSync         = false
	GenerateRegion = false
	// Location defines where to open up the kiki database
	Location = "."
)

func main() {
	flag.StringVar(&PublicPort, "external-port", PublicPort, "external port for the data (this) server")
	flag.StringVar(&PrivatePort, "internal-port", PrivatePort, "internal port for the data (this) server")
	flag.StringVar(&RegionPublic, "region-public", RegionPublic, "region public key")
	flag.StringVar(&RegionPrivate, "region-private", RegionPrivate, "region private key")
	flag.BoolVar(&NoSync, "no-sync", NoSync, "disable syncing")
	debug := flag.Bool("debug", false, "turn on debug mode")
	noBrowser := flag.Bool("no-browser", false, "do not open browser")
	flag.StringVar(&Location, "path", ".", "path to the kiki database folder")
	flag.BoolVar(&GenerateRegion, "generate-region", GenerateRegion, "generate keys for a new region")
	flag.Parse()
	if GenerateRegion {
		keys := keypair.New()
		fmt.Printf(`
	
	New Region:

	Public Key: %s
	
	Private Key: %s

	Start with:

	kiki -region-public '%s' -region-private '%s'

			`, keys.Public, keys.Private, keys.Public, keys.Private)
		os.Exit(1)
	}
	if *debug {
		logging.SetLoggingLevel("debug")
	} else {
		logging.SetLoggingLevel("info")
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
