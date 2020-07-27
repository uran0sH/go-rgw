package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-rgw/session"
	"net/http"
	"strings"
)

const metaPrefix = "C-Meta-"

// metadata is included in the request header in a form of key-value pairs and its prefix is "c-meta-"
// request header should contain the bucket(bucketName) and filename
func putObject(c *gin.Context) {
	body := c.Request.Body
	hash := c.GetHeader("Content-MD5")
	var metadata = make(map[string][]string)
	for key, value := range c.Request.Header {
		if strings.HasPrefix(key, metaPrefix) {
			metadata[key] = value
		}
	}
	bucketName := c.Param("bucket")
	objectName := c.Param("object")
	err := session.SaveObject(objectName, bucketName, body, hash, metadata, "")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	c.Header("ETag", hash)
	c.Status(http.StatusOK)
}

func createBucket(c *gin.Context) {
	bucketName := c.Param("bucket")
	session.CreateBucket(bucketName)
	c.Status(http.StatusOK)
}

func getObject(c *gin.Context) {
	bucketName := c.Param("bucket")
	objectName := c.Param("object")
	content, err := session.GetObject(bucketName, objectName)
	if err == nil {
		c.Writer.WriteHeader(http.StatusOK)
		c.Header("Content-Disposition", "attachment; filename="+objectName)
		c.Header("Content-Type", "application/text/plain")
		c.Header("Accept-Length", fmt.Sprintf("%d", len(content)))
		_, _ = c.Writer.Write(content)
	} else {
		c.Status(http.StatusInternalServerError)
	}
}

func initMultipartUpload(c *gin.Context) {

}
