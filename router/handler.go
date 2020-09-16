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
const acl = "C-Acl"
const grantReadAcl = "C-Grant-Read"
const grantWriteAcl = "C-Grant-Write"
const grantFullControlAcl = "C-Grant-Full-Control"

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
	var hashs []string
	hashs = append(hashs, hash)
	metadata["Content-MD5"] = hashs
	bucketName := c.Param("bucket")
	objectName := c.Param("object")

	userId := c.GetString("userId")
	if userId == "" {
		userId = "root"
	}
	ok, err := session.CouldPut(userId, bucketName)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}

	objectDefaultAcl := c.GetHeader(acl)
	if objectDefaultAcl == "" {
		objectDefaultAcl = session.Private
	}
	var grantee = make(map[string][]string)
	for key, value := range c.Request.Header {
		if key == grantReadAcl {
			grantee[grantReadAcl] = value
		}
		if key == grantWriteAcl {
			grantee[grantWriteAcl] = value
		}
		if key == grantFullControlAcl {
			grantee[grantFullControlAcl] = value
		}
	}
	accessControlList := session.NewAccessControlList(grantee)
	acl := session.NewAcl(userId, objectDefaultAcl, accessControlList)
	aclByte, err := json.Marshal(acl)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	err = session.SaveObject(objectName, bucketName, body, hash, metadata, string(aclByte))
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	c.Header("ETag", hash)
	c.Status(http.StatusOK)
}

func createBucket(c *gin.Context) {
	bucketName := c.Param("bucket")

	userId := c.GetString("userId")
	if userId == "" {
		userId = "root"
	}
	bucketDefaultAcl := c.GetHeader(acl)
	if bucketDefaultAcl == "" {
		bucketDefaultAcl = session.Private
	}
	var grantee = make(map[string][]string)
	for key, value := range c.Request.Header {
		if key == grantReadAcl {
			grantee[grantReadAcl] = value
		}
		if key == grantWriteAcl {
			grantee[grantWriteAcl] = value
		}
		if key == grantFullControlAcl {
			grantee[grantFullControlAcl] = value
		}
	}
	accessControlList := session.NewAccessControlList(grantee)
	acl := session.NewAcl(userId, bucketDefaultAcl, accessControlList)
	aclByte, err := json.Marshal(acl)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	err = session.CreateBucket(userId, bucketName, string(aclByte))
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func getObject(c *gin.Context) {
	bucketName := c.Param("bucket")
	objectName := c.Param("object")
	userId := c.GetString("userId")
	if userId == "" {
		userId = "root"
	}
	ok, err := session.CouldGet(userId, bucketName, objectName)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}
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

	userId := c.GetString("userId")
	if userId == "" {
		userId = "root"
	}
	ok, err := session.CouldPut(userId, bucketName)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}
	if !ok {
		c.Status(http.StatusForbidden)
		return
	}

	objectDefaultAcl := c.GetHeader(acl)
	if objectDefaultAcl == "" {
		objectDefaultAcl = session.Private
	}
	var grantee = make(map[string][]string)
	for key, value := range c.Request.Header {
		if key == grantReadAcl {
			grantee[grantReadAcl] = value
		}
		if key == grantWriteAcl {
			grantee[grantWriteAcl] = value
		}
		if key == grantFullControlAcl {
			grantee[grantFullControlAcl] = value
		}
	}
	accessControlList := session.NewAccessControlList(grantee)
	acl := session.NewAcl(userId, objectDefaultAcl, accessControlList)
	aclByte, err := json.Marshal(acl)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("%v", err))
		return
	}

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

	err = session.CreateMultipartUpload(objectName, bucketName, string(metadata), string(aclByte))
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
	var hashs []string
	hashs = append(hashs, hash)
	metadata["Content-MD5"] = hashs

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
