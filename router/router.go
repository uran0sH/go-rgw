package router

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/upload", putObject)
	r.GET("/download", getObject)
	return r
}
