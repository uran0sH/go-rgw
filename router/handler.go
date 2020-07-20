package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-rgw/auth"
	"go-rgw/connection"
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

type User struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func login(c *gin.Context) {
	var userRe User
	err := c.BindJSON(&userRe)
	if err != nil {
		c.String(http.StatusInternalServerError, "user parameter error")
		return
	}
	user := connection.MysqlMgr.MySQL.FindUser(userRe.Username)
	if user.Password == userRe.Password {
		tokenString, err := auth.GenToken(user.Username)
		if err != nil {
			c.String(http.StatusInternalServerError, "generate token error")
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "success",
			"data": tokenString,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "failure",
	})
}

func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "header is null",
			})
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"msg":  "header's format error",
			})
			c.Abort()
			return
		}
		mc, err := auth.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"msg":  "invalid token",
			})
			c.Abort()
			return
		}
		// 将当前请求的username信息保存到请求的上下文c上
		c.Set("username", mc.Username)
		c.Next()
	}
}

// register user
// Method: POST
// JSON: username & password
func register(c *gin.Context) {
	var registerUser User
	err := c.BindJSON(&registerUser)
	if err != nil {
		c.String(http.StatusInternalServerError, "user parameter error")
		return
	}
	if u := connection.MysqlMgr.MySQL.FindUser(registerUser.Username).Username; u != "" {
		c.String(http.StatusOK, "user has existed")
		return
	}
	connection.MysqlMgr.MySQL.SaveUser(registerUser.Username, registerUser.Password)
	c.String(http.StatusOK, "success")
	return
}
