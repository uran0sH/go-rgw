package router

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/register", register)
	r.POST("/login", login)
	authorized := r.Group("/")
	authorized.Use(JWTAuthMiddleware())
	{
		authorized.POST("/upload", putObject)
		authorized.GET("/download", getObject)
	}
	return r
}
