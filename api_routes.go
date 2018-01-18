package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/schollz/kiki/src/database"
)

func apiSuccessHandler(c *gin.Context, h gin.H) {
	logger.Log.Info(fmt.Sprintf("%v %v %v [%v]", c.Request.RemoteAddr, c.Request.Method, c.Request.URL, http.StatusOK))
	c.JSON(http.StatusOK, h)
}

func apiErrorHandler(c *gin.Context, err error) {
	logger.Log.Error(fmt.Sprintf("%v %v %v [%v]", c.Request.RemoteAddr, c.Request.Method, c.Request.URL, http.StatusInternalServerError))
	logger.Log.Warn(err)
	c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
}

func apiPostsHandler(c *gin.Context, posts []database.ApiBasicPost, err error) {
	if err != nil {
		apiErrorHandler(c, err)
		return
	}

	apiSuccessHandler(c, gin.H{
		"status": "ok",
		"data": gin.H{
			"posts": posts,
		},
	})
}

func apiFetchUserHandler(c *gin.Context, user_id string) {
	user, err := f.ShowUserForApi(user_id)
	apiUserHandler(c, user, err)
}

func apiUserHandler(c *gin.Context, user database.ApiUser, err error) {
	if err != nil {
		apiErrorHandler(c, err)
		return
	}

	apiSuccessHandler(c, gin.H{
		"status": "ok",
		"data": gin.H{
			"user": user,
		},
	})
}
