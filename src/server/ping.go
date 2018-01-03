package server

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

// PingHandler provides an api route for server health check
func PingHandler(c *gin.Context) {
	result := make(map[string]interface{})
	result["result"] = "pong"
	result["registered"] = startTime.UTC()
	result["uptime"] = time.Since(startTime).Seconds()
	result["num_cores"] = runtime.NumCPU()

	c.Writer.Header().Set("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": result})
}
