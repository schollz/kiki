package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/schollz/kiki/src/letter"
)

// GET /img
func handleImage(c *gin.Context) {
	AddCORS(c)
	id := c.Param("id")
	log.Debugf("fetching image: %s", id)
	e, err := f.GetEnvelope(id)
	if err != nil {
		log.Warn(err)
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	mimeType := "image/jpeg"
	if strings.Contains(e.Letter.Purpose, "png") {
		mimeType = "image/png"
	}

	imageBytes, err := base64.URLEncoding.DecodeString(e.Letter.Content)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/plain", []byte(err.Error()))
		return
	}

	c.Data(http.StatusOK, mimeType, imageBytes)
}

// POST /letter
func handleLetter(c *gin.Context) (err error) {
	AddCORS(c)

	// bind the payload
	var p letter.Letter
	err = c.BindJSON(&p)
	if err != nil {
		log.Debug(err)
		return
	}
	log.Debug(p)
	err = f.ProcessLetter(p)
	return
}

// GET /ping
func handlePing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": f.RegionKey.Public})
}

// POST /envelope
func handleEnvelope(c *gin.Context) (err error) {
	AddCORS(c)

	// bind the payload
	var p letter.Envelope
	err = c.BindJSON(&p)
	if err != nil {
		return
	}
	err = f.ProcessEnvelope(p)
	return
}

// GET /list
func handleList(c *gin.Context) {
	AddCORS(c)
	ids, err := f.GetIDs()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "found IDs", "ids": ids, "region_key": f.RegionKey.Public})
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
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "found envelope", "envelope": e})
	}
}

// POST /sync
func handleSync(c *gin.Context) (err error) {
	AddCORS(c)

	// bind the payload
	type Payload struct {
		Address string `json:"address" binding"required"`
	}
	var p Payload
	err = c.BindJSON(&p)
	if err != nil {
		log.Critical(err)
		return
	}

	err = f.Sync(p.Address)
	if err != nil {
		log.Critical(err)
		return
	}
	return
}
