package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-rgw/session"
	"io"
	"net/http"
	"strings"
)

const metaPrefix = "C-Meta-"

// metadata is included in the request header in a form of key-value pairs and its prefix is "c-meta-"
// request header should contain the bucket(bucketName) and filename
func putObject(c *gin.Context) {
	body := c.Request.Body
	cache := make([]byte, 1024)
	var data []byte
	for {
		n, err := body.Read(cache)
		// why?
		if err != nil && err != io.EOF {
			c.String(http.StatusInternalServerError, fmt.Sprintf("%s", err))
			return
		}
		data = append(data, cache[:n]...)
		if err == io.EOF {
			break
		}
	}
	var metadata []byte
	for key, value := range c.Request.Header {
		fmt.Println(key, value)
		if strings.HasPrefix(key, metaPrefix) {
			for _, v := range value {
				metadata = append(metadata, []byte(v)...)
			}
		}
	}
	bucketName := c.GetHeader("bucket")
	fileName := c.GetHeader("filename")
	err := session.SaveDataMetadata(bucketName, fileName, data, metadata)
	if err != nil {
		c.String(http.StatusInternalServerError, "save failed")
	} else {
		c.String(http.StatusOK, "success")
	}
}

func getObject(c *gin.Context) {
	filename := c.Query("filename")
	content, err := session.GetObject(filename)
	if err == nil {
		c.Writer.WriteHeader(http.StatusOK)
		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", "application/text/plain")
		fmt.Println(len(content))
		c.Header("Accept-Length", fmt.Sprintf("%d", len(content)))
		_, _ = c.Writer.Write(content)
	} else {
		c.String(http.StatusInternalServerError, "failed")
	}
}
