package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/schollz/kiki/src/feed"
	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/web"
)

var (
	f      *feed.Feed
	logger = logging.New()
)

func MiddleWareHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request
		logger.Log.Debug(fmt.Sprintf("%v %v %v", c.Request.RemoteAddr, c.Request.Method, c.Request.URL))
		// Add base headers
		AddCORS(c)
		// Run next function
		c.Next()
	}
}

func minus(a, b int) string {
	return strconv.FormatInt(int64(a)-int64(b), 10)
}

// Run will start the server listening
func Run(verbose bool) (err error) {
	if !verbose {
		logger.SetLevel("info")
	}
	// Startup feed
	logger.Log.Debug("opening feed")
	f, err = feed.New(Location, RegionPublic, RegionPrivate, verbose)
	if err != nil {
		return
	}
	f.Debug(verbose)
	logger.Log.Debug("opened feed")
	err = f.SetRegionKey(RegionPublic, RegionPrivate)
	if err != nil {
		return
	}
	logger.Log.Infof("Region public: %s", f.RegionKey.Public)
	logger.Log.Infof("Region private: %s", f.RegionKey.Private)
	err = f.Save()
	if err != nil {
		return
	}

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func(f *feed.Feed) {
		<-c
		f.Cleanup()
		os.Exit(1)
	}(f)
	defer f.Cleanup()

	// Startup server
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(MiddleWareHandler(), gin.Recovery()) // Standardize logs
	r.HTMLRender = loadTemplates("index.tmpl", "client.html")
	r.HEAD("/", func(c *gin.Context) { // handler for the uptime robot
		c.String(http.StatusOK, "OK")
	})
	r.GET("/home", handleHome)
	r.GET("/", handleSlash)

	// REST Api
	r.GET("/client", func(c *gin.Context) {
		c.HTML(http.StatusOK, "client.html", nil)
	})
	restApi = HttpRestApi{Db: f.GetDatabase(), PrimaryUserId: f.PersonalKey.Public, RegionPublicId: f.RegionKey.Public}
	restApi.AttachToRouter(r)
	//.end

	// /api/v1/friendsrout is depricated. Please use /api/v1/user or /api/v1/user/:user_id.
	r.GET("/api/v1/friendsrout", func(c *gin.Context) {
		logger.Log.Warn("/api/v1/friendsrout is depricated. Please use /api/v1/user or /api/v1/user/:user_id.")
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"data": gin.H{
				"friends": f.GetUserFriends(),
			},
		})
	})

	r.GET("/static/:file", func(c *gin.Context) {
		file := c.Param("file")
		filename := "static/" + file
		data, err := Asset(filename)
		if err != nil {
			c.String(http.StatusInternalServerError, "uh-oh")
		} else {
			c.Data(http.StatusOK, web.ContentType(filename), data)
		}
	})
	r.GET("/ping", handlePing)
	r.POST("/handshake", handleHandshake)
	r.GET("/img/:id", handleImage)
	r.POST("/letter", handlerLetter)       // post to put in letter (local only)
	r.OPTIONS("/letter", handlePing)       // post to put in letter (local only)
	r.POST("/sync", handlerSync)           // tell server to sync with another server (local only)
	r.OPTIONS("/sync", handlePing)         // post to put in letter (local only)
	r.GET("/list", handleList)             // GET list of all envelope IDs
	r.POST("/envelope", handlerEnvelope)   // post to put into database (public)
	r.GET("/download/:id", handleDownload) // download a specific envelope
	r.GET("/test", func(c *gin.Context) {
		message := ""
		f.TestStuff()
		c.JSON(http.StatusOK, gin.H{"success": err == nil, "message": message})
	})
	r.GET("/exit", func(c *gin.Context) {
		f.Cleanup()
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "exited"})
		go func() {
			time.Sleep(200 * time.Millisecond)
			os.Exit(0)
		}()
	})

	// PUBLIC FACING ROUTES
	publicRouter := gin.New()
	publicRouter.Use(MiddleWareHandler(), gin.Recovery())
	publicRouter.GET("/ping", handlePing) // PING a kiki server to see if it is available
	publicRouter.POST("/ping", handleHandshake)
	publicRouter.GET("/list", handleList)             // GET list of all envelope IDs
	publicRouter.POST("/envelope", handlerEnvelope)   // post to put into database (public)
	publicRouter.GET("/download/:id", handleDownload) // download a specific envelope

	// find a pair of ports
	for {
		ln1, err1 := net.Listen("tcp", ":"+PublicPort)
		ln2, err2 := net.Listen("tcp", ":"+PrivatePort)
		if err1 == nil {
			ln1.Close()
		}
		if err2 == nil {
			ln2.Close()
		}
		if err1 != nil || err2 != nil {
			portNum, _ := strconv.Atoi(PublicPort)
			PublicPort = strconv.Itoa(portNum + 2)
			portNum, _ = strconv.Atoi(PrivatePort)
			PrivatePort = strconv.Itoa(portNum + 2)
		} else {
			break
		}
	}

	go (func() {
		logger.Log.Infof("Running public router on 0.0.0.0:%s", PublicPort)
		err = publicRouter.Run(":" + PublicPort)
		if err != nil {
			logger.Log.Error(err)
			panic(err)
		}
	})()

	// private routes bind to localhost
	logger.Log.Infof("Running private router on localhost:%s", PrivatePort)
	err = r.Run("localhost:" + PrivatePort)
	if nil != err {
		logger.Log.Error(err)
		return
	}
	return
}

func respondWithJSON(c *gin.Context, message string, err error) {
	if nil != err {
		logger.Log.Error(fmt.Sprintf("%v %v %v [%v]", c.Request.RemoteAddr, c.Request.Method, c.Request.URL, 500))
		logger.Log.Warn(err)
		c.JSON(500, gin.H{"status": "error", "error": err.Error()})
		return
	}
	logger.Log.Info(fmt.Sprintf("%v %v %v [%v]", c.Request.RemoteAddr, c.Request.Method, c.Request.URL, 200))
	logger.Log.Debug(message)
	c.JSON(http.StatusOK, gin.H{"status": "ok", "message": message})
}

func handlerLetter(c *gin.Context) {
	handleLetter(c)
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

// func ValidateLocalAddress(c *gin.Context) (valid bool) {
// 	clientIP, errIP := web.GetClientIPHelper(c.Request)
// 	if errIP != nil {
// 		return
// 	}
// 	logger.Log.Debugf("Got IP adddress: '%s'", clientIP)
// 	return clientIP == "127.0.0.1" || clientIP == "::1"
// }

func loadTemplates(list ...string) multitemplate.Render {
	r := multitemplate.New()
	funcMap := template.FuncMap{
		"minus": minus,
	}
	for _, x := range list {
		templateString, err := Asset("templates/" + x)
		if err != nil {
			panic(err)
		}
		tmplMessage, err := template.New(x).Funcs(funcMap).Parse(string(templateString))
		if err != nil {
			panic(err)
		}
		r.Add(x, tmplMessage)
	}
	return r
}
