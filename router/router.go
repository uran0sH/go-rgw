package router

import (
	"github.com/gin-gonic/gin"
	"go-rgw/auth"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	authorized := registerAuthMiddleware(r, auth.DefaultAuth{})
	{
		authorized.POST("/upload", putObject)
		authorized.GET("/download", getObject)
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
