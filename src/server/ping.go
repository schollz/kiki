package server

import (
	// "encoding/json"
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

	// var data map[string]interface{}
	// data = make(map[string]interface{})
	// data["status"] = "success"
	result := make(map[string]interface{})
	result["result"] = "pong"
	result["registered"] = startTime.UTC()
	result["uptime"] = time.Since(startTime).Seconds()
	result["num_cores"] = runtime.NumCPU()
	// data["data"] = result

	c.Writer.Header().Set("Content-Type", "application/json")

	// js, err := json.Marshal(data)
	// if err != nil {
	// 	http.Error(w, `{"status": "error", "message": "`+err.Error()+`"}`, http.StatusInternalServerError)
	// 	return
	// }

	// w.WriteHeader(http.StatusOK)
	// w.Write(js)

	AddCORS(c)

	c.JSON(http.StatusOK, gin.H{"status": "ok", "data": result})
}
