package router

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-rgw/allocator"
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
	bucketAcl := c.GetHeader("C-Acl")
	userId := c.GetString("userId")
	err := session.CreateBucket(bucketName, userId, bucketAcl)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
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

func createMultipartUpload(c *gin.Context) {
	bucketName := c.Param("bucket")
	objectName := c.Param("object")
	var metadataMap = make(map[string][]string)
	for key, value := range c.Request.Header {
		if strings.HasPrefix(key, metaPrefix) {
			metadataMap[key] = value
		}
	}
	metadata, err := json.Marshal(metadataMap)
	if err != nil {
		c.String(http.StatusInternalServerError, "json marshal error")
	}
	err = session.CreateMultipartUpload(objectName, bucketName, string(metadata), "", true)
	if err != nil {
		c.String(http.StatusInternalServerError, "create failed")
	}
	uploadID := allocator.AllocateUUID()
	c.JSON(http.StatusOK, gin.H{
		"uploadID": uploadID,
	})
}

func uploadPart(c *gin.Context) {
	partID := c.Query("PartNumber")
	uploadID := c.Query("UploadId")
	hash := c.GetHeader("Content-MD5")
	bucketName := c.Param("bucket")
	objectName := c.Param("object")
	body := c.Request.Body

	var metadata = make(map[string][]string)
	for key, value := range c.Request.Header {
		if strings.HasPrefix(key, metaPrefix) {
			metadata[key] = value
		}
	}

	err := session.SaveObjectPart(objectName, bucketName, partID, uploadID, hash, body, metadata)
	if err != nil {
		c.String(http.StatusInternalServerError, "save failed")
		return
	}
	c.Header("ETag", hash)
	c.Status(http.StatusOK)
	return
}

type Part struct {
	PartID string `json:"PartID"`
	ETag   string `json:"ETag"`
}

type CompleteMultipart struct {
	Parts []Part
}

func completeMultipartUpload(c *gin.Context) {
	uploadID := c.Query("UploadId")
	bucketName := c.Param("bucket")
	objectNanme := c.Param("object")
	body := c.Request.Body
	var cache = make([]byte, 256)
	var data []byte
	for {
		n, err := body.Read(cache)
		if err != nil && err != io.EOF {
			c.Status(http.StatusInternalServerError)
			return
		}
		data = append(data, cache[:n]...)
		if err == io.EOF {
			break
		}
	}
	var multipart CompleteMultipart
	err := json.Unmarshal(data, &multipart)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	var partID []string
	for _, value := range multipart.Parts {
		partID = append(partID, value.PartID)
	}
	err = session.CompleteMultipartUpload(bucketName, objectNanme, uploadID, partID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
	return
}

func abortMultipartUpload(c *gin.Context) {
	uploadID := c.Query("UploadId")
	bucketName := c.Param("bucket")
	objectName := c.Param("object")
	err := session.AbortMultipartUpload(bucketName, objectName, uploadID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
	return
}
