package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"time"

	"github.com/darkowlzz/openurl"
	"github.com/schollz/kiki/src/keypair"
	"github.com/schollz/kiki/src/logging"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

var (
	Version        = "0.1.0"
	PrivatePort    = "8003"
	MessageFile    = ""
	HTMLMessage    = template.HTML("")
	PublicPort     = "8004"
	RegionPublic   = "4NfD9kWESGycUdbhbrFygNDjFun6NPk6utpkviyE1Ai6"
	RegionPrivate  = "btbsjnjTtgi3aL9z2X8bqb1URVnCo3zqg4fC4co2JEu"
	GenerateRegion = false
	// Location defines where to open up the kiki database
	Location = "."
)

func main() {
	flag.StringVar(&MessageFile, "message-file", MessageFile, "a markdown file for leaving a message")
	flag.StringVar(&PublicPort, "port-external", PublicPort, "external port for the data (this) server")
	flag.StringVar(&PrivatePort, "port-internal", PrivatePort, "internal port for the data (this) server")
	flag.StringVar(&RegionPublic, "region-public", RegionPublic, "region public key")
	flag.StringVar(&RegionPrivate, "region-private", RegionPrivate, "region private key")
	debug := flag.Bool("debug", false, "turn on debug mode")
	versionPrint := flag.Bool("version", false, "print version")
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
	if *versionPrint {
		fmt.Println(Version)
		os.Exit(1)
	}
	if *debug {
		logging.SetLoggingLevel("debug")
	} else {
		logging.SetLoggingLevel("info")
	}
	if MessageFile != "" {
		bMsg, err := ioutil.ReadFile(MessageFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		bMsg = blackfriday.Run(bMsg)
		HTMLMessage = template.HTML(string(bMsg))
	}
	if !*noBrowser {
		go func() {
			time.Sleep(1 * time.Second)
			openurl.Open("http://localhost:" + PrivatePort + "/home")
		}()
	}
	logging.Log.Infof("kiki version %s", Version)
	os.Mkdir(Location, 0755)

	err := Run(*debug)
	if err != nil {
		fmt.Println("error: " + err.Error())
	}
}
