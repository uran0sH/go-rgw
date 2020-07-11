package router

import (
	"github.com/gin-gonic/gin"
	"go-rgw/session"
	"net/http"
)

func putObject(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusInternalServerError, "error")
	}
	// 保存对象
	if file != nil {
		src, _ := file.Open()
		session.Save(file.Filename, src)
		c.String(http.StatusOK, "success")
	}
}

func getObject(c *gin.Context) {

}