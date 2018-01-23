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
	Version        = "0.1.0"
	PrivatePort    = "8003"
	PublicPort     = "8004"
	RegionPublic   = "GoAabW4QeCcyeeDWZxu9wFaPAoWhbrwvrFM83JToWk33"
	RegionPrivate  = "6ptaZoSaepphHTqQyCBRBBRF3WyKGoahXUUTVTL5BAQ3"
	GenerateRegion = false
	// Location defines where to open up the kiki database
	Location = "."
)

func main() {
	flag.StringVar(&PublicPort, "external-port", PublicPort, "external port for the data (this) server")
	flag.StringVar(&PrivatePort, "internal-port", PrivatePort, "internal port for the data (this) server")
	flag.StringVar(&RegionPublic, "region-public", RegionPublic, "region public key")
	flag.StringVar(&RegionPrivate, "region-private", RegionPrivate, "region private key")
	debug := flag.Bool("debug", false, "turn on debug mode")
	noBrowser := flag.Bool("no-browser", false, "do not open browser")
	flag.StringVar(&Location, "path", ".kiki", "path to the kiki data")
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
	logging.Log.Infof("kiki version %s", Version)
	os.Mkdir(Location, 0755)

	err := Run(*debug)
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
}
