package server

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/schollz/kiki/src/feed"
	"github.com/schollz/kiki/src/letter"
)

// POST /letter
func handleLetter(c *gin.Context) (err error) {
	AddCORS(c)

	if !strings.Contains(c.Request.RemoteAddr, "127.0.0.1") && !strings.Contains(c.Request.RemoteAddr, "[::1]") {
		return errors.New("must be on local host")
	}

	// bind the payload
	var p letter.Letter
	err = c.BindJSON(&p)
	if err != nil {
		return
	}
	err = feed.ProcessLetter(p)
	return
}
