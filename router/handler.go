package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-rgw/session"
	"io"
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
		data := make([]byte, 1024)
		var object []byte
		for {
			n, err := src.Read(data)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(n)
			object = append(object, data[:n]...)
			if n == 0 || err == nil || err == io.EOF{
				break
			}
		}
		err = session.SaveObject(file.Filename, object)
		if err == nil {
			c.String(http.StatusOK, "success")
		} else {
			c.String(http.StatusInternalServerError, "failed")
		}

	}
}

func getObject(c *gin.Context) {
	filename := c.Query("filename")
	content, err := session.GetObject(filename)
	if err == nil {
		c.Writer.WriteHeader(http.StatusOK)
		c.Header("Content-Disposition", "attachment; filename=" + filename)
		c.Header("Content-Type", "application/text/plain")
		fmt.Println(len(content))
		c.Header("Accept-Length", fmt.Sprintf("%d", len(content)))
		_, _ = c.Writer.Write(content)
	} else {
		c.String(http.StatusInternalServerError, "failed")
	}
}