package server

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/schollz/kiki/src/feed"
	"github.com/schollz/kiki/src/logging"
)

var (
	// Port defines what port the carrier should listen on
	Port = "8003"

	log = logging.Log
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
func Run() {
	// Startup server
	gin.SetMode(gin.ReleaseMode)

	// r := gin.Default()
	r := gin.New()
	// Standardize logs
	r.Use(MiddleWareHandler(), gin.Recovery())

	r.HEAD("/", func(c *gin.Context) { // handler for the uptime robot
		c.String(http.StatusOK, "OK")
	})
	r.GET("/ping", PingHandler)
	r.POST("/letter", handlerLetter)
	r.POST("/open", handlerOpen)
	r.Run(":" + Port) // listen and serve on 0.0.0.0:Port

	log.Info("Server is listening to port: " + Port)
}

func respondWithJSON(c *gin.Context, message string, err error) {
	success := err == nil
	if !success {
		message = err.Error()
	}
	c.JSON(http.StatusOK, gin.H{"success": success, "message": message})
}

func handlerOpen(c *gin.Context) {
	respondWithJSON(c, "opened envelopes", handleOpen(c))
}

func handleOpen(c *gin.Context) (err error) {
	err = feed.OpenEnvelopes()
	if err != nil {
		return
	}
	err = feed.RegenerateFeed()
	return
}

func handlerAssign(c *gin.Context) {
	respondWithJSON(c, "assigned", handleAssign(c))
}

func handleAssign(c *gin.Context) (err error) {
	assignmentType := c.PostForm("assign")
	assignData := c.PostForm("data")
	if len(assignData) == 0 {
		return errors.New("assigned data cannot be empty")
	}

	fmt.Println(assignmentType)

	// TODO:  feed.PostMessage("assign-"+assignmentType, assignData, true)
	return nil
}

func handlerLetter(c *gin.Context) {
	respondWithJSON(c, "letter added", handleLetter(c))
}

func handleLetter(c *gin.Context) (err error) {
	// AddCORS(c)

	if !strings.Contains(c.Request.RemoteAddr, "127.0.0.1") && !strings.Contains(c.Request.RemoteAddr, "[::1]") {
		return errors.New("must be on local host")
	}

	// bind the payload
	var p feed.Message
	err = c.BindJSON(&p)
	if err != nil {
		return
	}
	return p.Post()
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
	// This should be assigned via the route
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	//.end
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
}
