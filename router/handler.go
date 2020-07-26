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

	var metadata = make(map[string][]string)
	for key, value := range c.Request.Header {
		//TODO delete
		fmt.Println(key, value)
		if strings.HasPrefix(key, metaPrefix) {
			metadata[key] = value
		}
	}
	bucketName := c.Param("bucket")
	objectName := c.Param("object")
	err := session.SaveObject(objectName, bucketName, body, metadata, "")
	if err != nil {
		c.String(http.StatusInternalServerError, "save failed")
		return
	}
	c.String(http.StatusOK, "success")
}

func createBucket(c *gin.Context) {
	bucketName := c.Param("bucket")
	session.CreateBucket(bucketName)
}

//func getObject(c *gin.Context) {
//	filename := c.Query("filename")
//	content, err := session.GetObject(filename)
//	if err == nil {
//		c.Writer.WriteHeader(http.StatusOK)
//		c.Header("Content-Disposition", "attachment; filename="+filename)
//		c.Header("Content-Type", "application/text/plain")
//		fmt.Println(len(content))
//		c.Header("Accept-Length", fmt.Sprintf("%d", len(content)))
//		_, _ = c.Writer.Write(content)
//	} else {
//		c.String(http.StatusInternalServerError, "failed")
//	}
//}
