package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/schollz/kiki/src/database"
)

var restApi HttpRestApi

// type ApiRoute struct {
// 	Name        string
// 	Method      string
// 	Pattern     string
// 	HandlerFunc http.HandlerFunc
// }
//
// type ApiRoutes []ApiRoute
//
// var routes = ApiRoutes{
// 	// Health check
// 	apiRoute{"Ping", "GET", "/ping", http_helpers.PingHandler},
// 	apiRoute{"Login", "POST", "/login", users.LoginHandler},
// 	apiRoute{"CreateUser", "POST", "/api/v1/user", users.CreateUserHandler},
// 	apiRoute{"DeleteUser", "DELETE", "/api/v1/user/{user_id}", users.DeleteUserHandler},
// 	apiRoute{"CreateCustomer", "POST", "/api/v1/customer", users.CreateCustomerHandler},
// 	apiRoute{"GetCustomer", "GET", "/api/v1/customer/{customer_id}", users.GetCustomerHandler},
// 	apiRoute{"DeleteCustomer", "DELETE", "/api/v1/customer/{customer_id}", users.DeleteCustomerHandler},
// }
//
// func AttachHttpApiHandlers() {
// 	for _, route := range routes {
// 		var handler http.Handler
// 		// log.Println("Attaching HTTP handler for route:", route.Method, route.Pattern)
// 		Log.Info("Attaching HTTP handler for route: ", route.Method, " ", route.Pattern)
// 		handler = route.HandlerFunc
// 		router.
// 			Methods(route.Method).
// 			Path(route.Pattern).
// 			Name(route.Name).
// 			Handler(handler)
// 	}
// }

// HttpRestApiRoutes map[string]func(*gin.Context)

type HttpRestApi struct {
	PrimaryUserId string
	Db            database.DatabaseAPI
}

func (self HttpRestApi) AttachToRouter(router *gin.Engine) {
	router.GET("/api/v1/posts", self.GetPosts)
	router.GET("/api/v1/post/:post_id", self.GetPost)
	router.GET("/api/v1/post/:post_id/comments", self.GetPostComments)
	router.GET("/api/v1/user", self.GetPrimaryUser)
	router.GET("/api/v1/user/:user_id", self.GetUser)
}

func (self HttpRestApi) GetPosts(c *gin.Context) {
	posts, err := self.Db.GetPostsForApi()
	self.apiPostsHandler(c, posts, err)
}

func (self HttpRestApi) GetPost(c *gin.Context) {
	post_id := c.Param("post_id")
	posts, err := self.Db.GetPostForApi(post_id)
	self.apiPostsHandler(c, posts, err)
}

func (self HttpRestApi) GetPostComments(c *gin.Context) {
	post_id := c.Param("post_id")
	posts, err := self.Db.GetPostCommentsForApi(post_id)
	self.apiPostsHandler(c, posts, err)
}

func (self HttpRestApi) GetPrimaryUser(c *gin.Context) {
	user_id := self.PrimaryUserId
	self.apiFetchUserHandler(c, user_id)
}

func (self HttpRestApi) GetUser(c *gin.Context) {
	user_id := c.Param("user_id")
	self.apiFetchUserHandler(c, user_id)
}

func (self HttpRestApi) apiSuccessHandler(c *gin.Context, h gin.H) {
	logger.Log.Info(fmt.Sprintf("%v %v %v [%v]", c.Request.RemoteAddr, c.Request.Method, c.Request.URL, http.StatusOK))
	c.JSON(http.StatusOK, h)
}

func (self HttpRestApi) apiErrorHandler(c *gin.Context, err error) {
	logger.Log.Error(fmt.Sprintf("%v %v %v [%v]", c.Request.RemoteAddr, c.Request.Method, c.Request.URL, http.StatusInternalServerError))
	logger.Log.Warn(err)
	c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
}

func (self HttpRestApi) apiPostsHandler(c *gin.Context, posts []database.ApiBasicPost, err error) {
	if err != nil {
		self.apiErrorHandler(c, err)
		return
	}

	self.apiSuccessHandler(c, gin.H{
		"status": "ok",
		"data": gin.H{
			"posts": posts,
		},
	})
}

func (self HttpRestApi) apiFetchUserHandler(c *gin.Context, user_id string) {
	user, err := self.Db.GetUserForApi(user_id)
	self.apiUserHandler(c, user, err)
}

func (self HttpRestApi) apiUserHandler(c *gin.Context, user database.ApiUser, err error) {
	if err != nil {
		self.apiErrorHandler(c, err)
		return
	}

	self.apiSuccessHandler(c, gin.H{
		"status": "ok",
		"data": gin.H{
			"user": user,
		},
	})
}
