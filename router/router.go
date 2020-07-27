package router

import (
	"github.com/gin-gonic/gin"
	"go-rgw/auth"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	authorized := registerAuthMiddleware(r, auth.DefaultAuth{})
	{
		authorized.GET("/createbucket/:bucket", createBucket)
		authorized.POST("/upload/:bucket/:object", putObject)
		authorized.GET("/download/:bucket/:object", getObject)
		authorized.POST("/multipartupload/init", initMultipartUpload)
	}
	return r
}

func registerAuthMiddleware(e *gin.Engine, auth auth.Auth) *gin.RouterGroup {
	e.POST("/register", auth.CreateUser)
	e.POST("/login", auth.Login)
	g := e.Group("/")
	g.Use(auth.Auth())
	return g
}
