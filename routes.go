package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/schollz/kiki/src/feed"
	"github.com/schollz/kiki/src/letter"
)

// GET /img
func handleImage(c *gin.Context) {
	id := c.Param("id")
	logger.Log.Debugf("fetching image: %s", id)
	e, err := f.GetEnvelope(id)
	if err != nil {
		logger.Log.Warn(err)
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	mimeType := "image/jpeg"
	if strings.Contains(e.Letter.Purpose, "png") {
		mimeType = "image/png"
	}

	imageBytes, err := base64.StdEncoding.DecodeString(e.Letter.Content)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	c.Data(http.StatusOK, mimeType, imageBytes)
}

// POST /letter
func handleLetter(c *gin.Context) (err error) {
	// bind the payload
	var p letter.Letter
	err = c.BindJSON(&p)
	if err != nil {
		logger.Log.Error(err)
		c.JSON(500, gin.H{"status": "error", "error": err.Error()})
		return
	}
	e, err := f.ProcessLetter(p)
	if err != nil {
		logger.Log.Error(err)
		c.JSON(500, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok", "message": "added " + e.ID, "envelope": e})

	// when a new letter arrives, update everything and then sync servers
	go f.UpdateEverythingAndSync()
	return
}

// GET /ping
func handlePing(c *gin.Context) {
	fmt.Printf("%+v", c.Request)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "error": "pong"})
}

// POST /handshake
func handleHandshake(c *gin.Context) {
	var p feed.Response
	err := c.BindJSON(&p)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "error": "problem binding data"})
		return
	}

	err = f.ValidateKikiInstance(p)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "error": err.Error()})
		return
	}

	signature, err := f.RegionKey.Signature(f.RegionKey)
	personalSignature, err2 := f.PersonalKey.Signature(f.RegionKey)
	if err != nil || err2 != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "error": "problem signing"})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "pong", "region_key": f.RegionKey.Public, "region_signature": signature, "personal_key": f.PersonalKey.Public, "personal_signature": personalSignature})
	}
}

// POST /envelope
func handleEnvelope(c *gin.Context) (err error) {
	// bind the payload
	var p letter.Envelope
	err = c.BindJSON(&p)
	if err != nil {
		return
	}
	err = f.ProcessEnvelope(p)
	f.SignalUpdate()
	return
}

// GET /list
func handleList(c *gin.Context) {
	ids, err := f.GetIDs()
	if err != nil {
		logger.Log.Error(err)
		c.JSON(500, gin.H{"status": "error", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "found IDs", "ids": ids, "region_key": f.RegionKey.Public})
	}
	return
}

// GET /download/ID
// You can always download anything you want but the envelopes are transfered so that the letter is closed up.
func handleDownload(c *gin.Context) {
	id := c.Param("id")
	fmt.Println(id)
	e, err := f.GetEnvelope(id)
	// Close up envelope
	e.Close()
	if err != nil {
		c.JSON(500, gin.H{"status": "error", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "found envelope", "envelope": e})
	}
}

// POST /sync
func handleSync(c *gin.Context) (err error) {
	// bind the payload
	type Payload struct {
		Address string `json:"address" binding"required"`
	}
	var p Payload
	err = c.BindJSON(&p)
	if err != nil {
		logger.Log.Error(err)
		return
	}

	if p.Address == "" {
		logger.Log.Debug("only syncing servers")
		f.SyncServers()
		return
	}

	logger.Log.Debug("syncing...")
	err = f.Sync(p.Address)
	if err != nil {
		logger.Log.Error(err)
		return
	}
	return
}
