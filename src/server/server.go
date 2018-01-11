package server

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/schollz/kiki/src/feed"
	"github.com/schollz/kiki/src/logging"
)

var (
	// Port defines what port the carrier should listen on
	Port = "8003"
	// Location defines where to open up the kiki database
	Location = "."
	f        feed.Feed
	log      = logging.Log
)

func MiddleWareHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request
		log.Debug(fmt.Sprintf("%v %v %v", c.Request.RemoteAddr, c.Request.Method, c.Request.URL))
		// Add base headers
		AddCORS(c)
		// Run next function
		c.Next()
	}
}

// Run will start the server listening
func Run() (err error) {
	// Startup feed
	log.Debug("opening feed")
	f, err = feed.Open(Location)
	if err != nil {
		log.Debug("feed not found, creating new one")
		var err2 error
		f, err2 = feed.New(Location)
		if err2 != nil {
			return err2
		}
	}

	// Startup server
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	// Standardize logs
	r.Use(MiddleWareHandler(), gin.Recovery())

	r.HEAD("/", func(c *gin.Context) { // handler for the uptime robot
		c.String(http.StatusOK, "OK")
	})
	r.GET("/ping", handlePing)
	r.POST("/letter", handlerLetter)       // post to put in letter (local only)
	r.OPTIONS("/letter", handlePing)       // post to put in letter (local only)
	r.GET("/download/:id", handleDownload) // download a specific envelope
	r.GET("/list", handleList)             // list all the current envelopes
	r.POST("/envelope", handlerEnvelope)   // post to put into database (public)
	r.POST("/sync", handlerSync)           // tell server to sync with another server (local only)
	r.GET("/test", func(c *gin.Context) {
		message := ""
		fmt.Println(f.IsKikiInstance("http://localhost:8004"))
		err := f.ShowFeed()
		if err != nil {
			message = err.Error()
		}
		c.JSON(http.StatusOK, gin.H{"success": err == nil, "message": message})
	})
	err = r.Run(":" + Port) // listen and serve on 0.0.0.0:Port
	return
}

func respondWithJSON(c *gin.Context, message string, err error) {
	success := err == nil
	if !success {
		message = err.Error()
	}
	c.JSON(http.StatusOK, gin.H{"success": success, "message": message})
}

func handlerLetter(c *gin.Context) {
	respondWithJSON(c, "letter added", handleLetter(c))
}

func handlerEnvelope(c *gin.Context) {
	respondWithJSON(c, "envelope added", handleEnvelope(c))
}

func handlerSync(c *gin.Context) {
	respondWithJSON(c, "synced", handleSync(c))
}

func readFormFile(file *multipart.FileHeader) (data []byte, err error) {
	src, err := file.Open()
	if err != nil {
		return
	}
	defer src.Close()
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, src)
	data = buf.Bytes()
	return
}

func AddCORS(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
}
