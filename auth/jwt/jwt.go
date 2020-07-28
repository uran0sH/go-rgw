package jwt

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go-rgw/allocator"
	"go-rgw/connection"
	"net/http"
	"strings"
	"time"
)

type MyClaims struct {
	UserID string
	jwt.StandardClaims
}

const TokenExpireDuration = 2 * time.Hour

var secret = []byte("Thehardestchoicesrequirethestrongestwills")

func GenToken(userID string) (string, error) {
	claims := MyClaims{
		userID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(TokenExpireDuration).Unix(),
			Issuer:    "go-rgw",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseToken(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	// 校验token
	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

type JWT struct{}

type User struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func (j *JWT) Login(c *gin.Context) {
	var userRe User
	err := c.BindJSON(&userRe)
	if err != nil {
		c.String(http.StatusInternalServerError, "user parameter error")
		return
	}
	user := connection.MysqlMgr.MySQL.FindUser(userRe.Username)
	if user.Password == userRe.Password {
		tokenString, err := GenToken(user.UserID)
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

// create user
// Method: POST
// JSON: username & password
func (j *JWT) CreateUser(c *gin.Context) {
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
	uid := allocator.AllocateUUID()
	connection.MysqlMgr.MySQL.CreateUser(registerUser.Username, registerUser.Password, uid)
	c.String(http.StatusOK, "success")
	return
}

func (j *JWT) Auth() func(c *gin.Context) {
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
		mc, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"msg":  "invalid token",
			})
			c.Abort()
			return
		}
		// 将当前请求的username信息保存到请求的上下文c上
		c.Set("username", mc.UserID)
		c.Next()
	}
}
